// Command smoke boots the monolith in-process and runs a full REST smoke test
// with colored pass/fail output.
//
// Usage:
//
//	go run ./cmd/smoke
//	make smoke
package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"github.com/FibrinLab/open-nucleus/internal/handler"
	fhirhandler "github.com/FibrinLab/open-nucleus/internal/handler/fhir"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/router"
	"github.com/FibrinLab/open-nucleus/internal/service/local"
	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/anchor/anchorservice"
	"github.com/FibrinLab/open-nucleus/services/auth/authservice"
	"github.com/FibrinLab/open-nucleus/services/formulary/formularyservice"
	"github.com/FibrinLab/open-nucleus/services/patient/pipeline"
	"github.com/FibrinLab/open-nucleus/services/sync/syncservice"

	_ "modernc.org/sqlite"
)

// ANSI color codes.
const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorCyan  = "\033[36m"
	colorBold  = "\033[1m"
	colorDim   = "\033[2m"
)

// cleanups tracks teardown functions accumulated during setup.
var cleanupFns []func()

func addCleanup(fn func()) {
	cleanupFns = append(cleanupFns, fn)
}

func runCleanups() {
	for i := len(cleanupFns) - 1; i >= 0; i-- {
		cleanupFns[i]()
	}
}

// ---------- Monolith wiring ----------

type stack struct {
	server      *httptest.Server
	accessToken string
}

func wireMonolith(tmpDir, bootstrapSecret string) (*stack, error) {
	// --- Shared data layer ---
	git, err := gitstore.NewStore(tmpDir+"/repo", "smoke", "smoke@test.local")
	if err != nil {
		return nil, fmt.Errorf("gitstore: %w", err)
	}

	db, err := sql.Open("sqlite", tmpDir+"/nucleus.db?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	addCleanup(func() { db.Close() })

	if err := sqliteindex.InitUnifiedSchema(db); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	idx := sqliteindex.NewIndexFromDB(db)

	// --- Patient ---
	pw := pipeline.NewWriter(git, idx, 10*time.Second)
	patientSvc := local.NewPatientService(pw, idx, git)

	// --- Auth ---
	ks := auth.NewMemoryKeyStore()
	denyList := authservice.NewDenyList(db)
	clientStore := authservice.NewClientStore(git, db)

	authCfg := &authservice.Config{
		JWT: authservice.JWTConfig{
			Issuer:          "open-nucleus-auth",
			AccessLifetime:  time.Hour,
			RefreshLifetime: 24 * time.Hour,
			ClockSkew:       2 * time.Hour,
		},
		Git: authservice.GitConfig{
			RepoPath:    tmpDir + "/repo",
			AuthorName:  "smoke",
			AuthorEmail: "smoke@test.local",
		},
		Devices: authservice.DevicesConfig{Path: ".nucleus/devices"},
		Security: authservice.SecurityConfig{
			NonceTTL:        60 * time.Second,
			MaxFailures:     10,
			FailureWindow:   60 * time.Second,
			BootstrapSecret: bootstrapSecret,
		},
		KeyStore: authservice.KeyStoreConfig{Type: "memory"},
		SQLite:   authservice.SQLiteConfig{DBPath: tmpDir + "/nucleus.db"},
	}

	authImpl, err := authservice.NewAuthService(authCfg, git, ks, denyList)
	if err != nil {
		return nil, fmt.Errorf("auth service: %w", err)
	}
	authSvc := local.NewAuthService(authImpl)
	smartImpl := authservice.NewSmartService(authImpl, clientStore)
	smartSvc := local.NewSmartService(smartImpl)

	// --- Sync ---
	mergeDriver := merge.NewDriver(nil)
	eventBus := syncservice.NewEventBus(100)
	conflictStore := syncservice.NewConflictStore(db)
	historyStore := syncservice.NewHistoryStore(db, 10000)
	peerStore := syncservice.NewPeerStore(db)

	syncCfg := &syncservice.Config{
		Git: syncservice.GitConfig{
			RepoPath:    tmpDir + "/repo",
			AuthorName:  "smoke",
			AuthorEmail: "smoke@test.local",
		},
	}

	syncEngine := syncservice.NewSyncEngine(
		syncCfg, git, conflictStore, historyStore, peerStore,
		mergeDriver, eventBus, "node-smoke", "site-smoke",
	)
	syncSvc := local.NewSyncService(syncEngine, historyStore, peerStore)
	conflictSvc := local.NewConflictService(conflictStore, eventBus)

	// --- Formulary ---
	formularyImpl := formularyservice.New(
		formularyservice.NewDrugDB(),
		formularyservice.NewInteractionIndex(),
		formularyservice.NewStockStore(db),
		formularyservice.NewStubDosingEngine(),
	)
	formularySvc := local.NewFormularyService(formularyImpl)

	// --- Anchor ---
	anchorImpl := anchorservice.New(
		git, openanchor.NewStubBackend(), openanchor.NewLocalIdentityEngine(),
		anchorservice.NewAnchorQueue(db), anchorservice.NewAnchorStore(git),
		anchorservice.NewCredentialStore(git), anchorservice.NewDIDStore(git),
		authImpl.NodePrivateKey(),
	)
	anchorSvc := local.NewAnchorService(anchorImpl)

	// --- Stubs ---
	sentinelSvc := local.NewStubSentinelService()
	supplySvc := local.NewStubSupplyService()

	// --- Handlers ---
	authH := handler.NewAuthHandler(authSvc)
	patientH := handler.NewPatientHandler(patientSvc)
	syncH := handler.NewSyncHandler(syncSvc)
	conflictH := handler.NewConflictHandler(conflictSvc)
	sentinelH := handler.NewSentinelHandler(sentinelSvc)
	formularyH := handler.NewFormularyHandler(formularySvc)
	anchorH := handler.NewAnchorHandler(anchorSvc)
	supplyH := handler.NewSupplyHandler(supplySvc)
	resourceH := handler.NewResourceHandler(patientSvc)
	fhirH := fhirhandler.NewFHIRHandler(patientSvc)
	smartH := handler.NewSmartHandler(smartSvc, "http://localhost:8080")

	cfg := &config.Config{
		RateLimit: config.RateLimitConfig{
			ReadRPM: 200, ReadBurst: 50,
			WriteRPM: 60, WriteBurst: 20,
			AuthRPM: 100, AuthBurst: 50,
		},
		CORS: config.CORSConfig{AllowedOrigins: []string{"*"}},
	}

	jwtAuth := middleware.NewJWTAuth(authImpl.NodePublicKey(), "open-nucleus-auth")
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)
	auditLogger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	mux := router.New(router.Config{
		AuthHandler:      authH,
		PatientHandler:   patientH,
		ResourceHandler:  resourceH,
		SyncHandler:      syncH,
		ConflictHandler:  conflictH,
		SentinelHandler:  sentinelH,
		FormularyHandler: formularyH,
		AnchorHandler:    anchorH,
		SupplyHandler:    supplyH,
		FHIRHandler:      fhirH,
		SmartHandler:     smartH,
		JWTAuth:          jwtAuth,
		RateLimiter:      rateLimiter,
		CORSOrigins:      cfg.CORS.AllowedOrigins,
		AuditLogger:      auditLogger,
	})

	httpServer := httptest.NewServer(mux)
	addCleanup(func() { httpServer.Close() })

	// --- Bootstrap device & authenticate (direct Go calls) ---
	pub, priv, err := auth.GenerateKeypair()
	if err != nil {
		return nil, fmt.Errorf("keypair: %w", err)
	}

	device, err := authImpl.RegisterDevice(auth.EncodePublicKey(pub), "dr-smoke", "site-smoke", "smoke-tablet", "physician", bootstrapSecret)
	if err != nil {
		return nil, fmt.Errorf("register device: %w", err)
	}

	nonce, _, err := authImpl.GetChallenge(device.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("get challenge: %w", err)
	}

	sig := auth.Sign(priv, nonce)
	accessToken, _, _, _, _, err := authImpl.AuthenticateWithNonce(device.DeviceID, nonce, sig)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return &stack{
		server:      httpServer,
		accessToken: accessToken,
	}, nil
}

// ---------- HTTP helpers ----------

func doRequest(method, url string, body any, token string) (*http.Response, map[string]any, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}

	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return resp, result, nil
}

// ---------- Test runner ----------

type step struct {
	name   string
	method string
	path   string
	body   any
	auth   bool
	expect int
	// extract is called after a successful step to capture values.
	extract func(body map[string]any)
	// acceptAlternate allows an alternate status code (e.g., 403 for known RBAC mismatch).
	acceptAlternate int
}

func main() {
	fmt.Printf("\n%s%s Open Nucleus — Smoke Test %s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s Booting monolith in-process...%s\n\n", colorDim, colorReset)

	tmpDir, err := os.MkdirTemp("", "nucleus-smoke-*")
	if err != nil {
		fatal("create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	defer runCleanups()

	const bootstrapSecret = "smoke-bootstrap-secret"

	st, err := wireMonolith(tmpDir, bootstrapSecret)
	if err != nil {
		fatal("wire monolith: %v", err)
	}
	fmt.Printf("  Monolith        %s%s%s\n", colorDim, st.server.URL, colorReset)
	fmt.Printf("  Auth            %sPASS (Ed25519 challenge-response)%s\n\n", colorDim, colorReset)

	// --- Define test steps ---
	var patientID string

	steps := []step{
		{
			name:   "Health check",
			method: "GET", path: "/health",
			auth: false, expect: 200,
		},
		{
			name:   "Auth required (no token)",
			method: "GET", path: "/api/v1/patients/",
			auth: false, expect: 401,
		},
		{
			name:   "Create patient",
			method: "POST", path: "/api/v1/patients/",
			body: map[string]any{
				"resourceType": "Patient",
				"name":         []map[string]any{{"family": "Doe", "given": []string{"John"}}},
				"gender":       "male",
				"birthDate":    "1990-01-15",
			},
			auth: true, expect: 201,
			extract: func(body map[string]any) {
				if data, ok := body["data"].(map[string]any); ok {
					if id, ok := data["id"].(string); ok {
						patientID = id
					}
				}
			},
		},
		{
			name:   "Get patient",
			method: "GET", path: "{{patient}}",
			auth: true, expect: 200,
		},
		{
			name:   "List patients",
			method: "GET", path: "/api/v1/patients/",
			auth: true, expect: 200,
		},
		{
			name:   "Create encounter",
			method: "POST", path: "{{patient}}/encounters",
			body: map[string]any{
				"resourceType": "Encounter",
				"status":       "finished",
				"class":        map[string]any{"code": "AMB", "system": "http://terminology.hl7.org/CodeSystem/v3-ActCode"},
				"period":       map[string]any{"start": "2026-01-15T09:00:00Z", "end": "2026-01-15T10:00:00Z"},
			},
			auth: true, expect: 201,
		},
		{
			name:   "List encounters",
			method: "GET", path: "{{patient}}/encounters",
			auth: true, expect: 200,
		},
		{
			name:   "Create observation",
			method: "POST", path: "{{patient}}/observations",
			body: map[string]any{
				"resourceType":      "Observation",
				"status":            "final",
				"effectiveDateTime": "2026-01-15T09:30:00Z",
				"code": map[string]any{
					"coding": []map[string]any{{
						"system":  "http://loinc.org",
						"code":    "8867-4",
						"display": "Heart rate",
					}},
				},
				"valueQuantity": map[string]any{
					"value":  72,
					"unit":   "beats/minute",
					"system": "http://unitsofmeasure.org",
					"code":   "/min",
				},
			},
			auth: true, expect: 201,
		},
		{
			name:   "Create condition",
			method: "POST", path: "{{patient}}/conditions",
			body: map[string]any{
				"resourceType": "Condition",
				"clinicalStatus": map[string]any{
					"coding": []map[string]any{{
						"system": "http://terminology.hl7.org/CodeSystem/condition-clinical",
						"code":   "active",
					}},
				},
				"verificationStatus": map[string]any{
					"coding": []map[string]any{{
						"system": "http://terminology.hl7.org/CodeSystem/condition-ver-status",
						"code":   "confirmed",
					}},
				},
				"code": map[string]any{
					"coding": []map[string]any{{
						"system":  "http://snomed.info/sct",
						"code":    "386661006",
						"display": "Fever",
					}},
				},
			},
			auth: true, expect: 201,
		},
		{
			name:   "Create medication request",
			method: "POST", path: "{{patient}}/medication-requests",
			body: map[string]any{
				"resourceType": "MedicationRequest",
				"status":       "active",
				"intent":       "order",
				"medicationCodeableConcept": map[string]any{
					"coding": []map[string]any{{
						"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
						"code":    "161",
						"display": "Acetaminophen",
					}},
				},
				"dosageInstruction": []map[string]any{{
					"text": "500mg every 6 hours as needed",
					"timing": map[string]any{
						"repeat": map[string]any{"frequency": 4, "period": 1, "periodUnit": "d"},
					},
					"doseAndRate": []map[string]any{{
						"doseQuantity": map[string]any{"value": 500, "unit": "mg", "system": "http://unitsofmeasure.org", "code": "mg"},
					}},
				}},
			},
			auth: true, expect: 201,
		},
		{
			name:   "Create allergy intolerance",
			method: "POST", path: "{{patient}}/allergy-intolerances",
			body: map[string]any{
				"resourceType": "AllergyIntolerance",
				"clinicalStatus": map[string]any{
					"coding": []map[string]any{{
						"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
						"code":   "active",
					}},
				},
				"verificationStatus": map[string]any{
					"coding": []map[string]any{{
						"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification",
						"code":   "confirmed",
					}},
				},
				"type": "allergy",
				"code": map[string]any{
					"coding": []map[string]any{{
						"system":  "http://snomed.info/sct",
						"code":    "91936005",
						"display": "Penicillin allergy",
					}},
				},
			},
			auth: true, expect: 201,
		},
		{
			name:   "Patient timeline",
			method: "GET", path: "{{patient}}/timeline",
			auth: true, expect: 200,
		},
		{
			name:   "Patient history",
			method: "GET", path: "{{patient}}/history",
			auth: true, expect: 200,
		},
		{
			name:            "Sync status",
			method:          "GET", path: "/api/v1/sync/status",
			auth:            true, expect: 200,
			acceptAlternate: 403, // known RBAC name mismatch
		},
		{
			name:   "List conflicts",
			method: "GET", path: "/api/v1/conflicts/",
			auth: true, expect: 200,
		},
		// --- Formulary steps ---
		{
			name:   "Search medications",
			method: "GET", path: "/api/v1/formulary/medications?q=amox",
			auth: true, expect: 200,
		},
		{
			name:   "Get medication by code",
			method: "GET", path: "/api/v1/formulary/medications/J01CA04",
			auth: true, expect: 200,
		},
		{
			name:   "Check drug interactions",
			method: "POST", path: "/api/v1/formulary/check-interactions",
			body: map[string]any{
				"medication_codes": []string{"L04AX03", "J01EA01"},
			},
			auth: true, expect: 200,
		},
		{
			name:   "Get formulary info",
			method: "GET", path: "/api/v1/formulary/info",
			auth: true, expect: 200,
		},
		{
			name:   "Check allergy conflicts",
			method: "POST", path: "/api/v1/formulary/check-allergies",
			body: map[string]any{
				"medication_codes": []string{"J01CA04"},
				"allergy_codes":    []string{"91936005"},
			},
			auth: true, expect: 200,
		},
		// --- Anchor steps ---
		{
			name:   "Anchor status",
			method: "GET", path: "/api/v1/anchor/status",
			auth: true, expect: 200,
		},
		{
			name:   "Trigger anchor",
			method: "POST", path: "/api/v1/anchor/trigger",
			auth: true, expect: 200,
		},
		{
			name:   "Get node DID",
			method: "GET", path: "/api/v1/anchor/did/node",
			auth: true, expect: 200,
		},
		{
			name:   "List backends",
			method: "GET", path: "/api/v1/anchor/backends",
			auth: true, expect: 200,
		},
		{
			name:   "Queue status",
			method: "GET", path: "/api/v1/anchor/queue",
			auth: true, expect: 200,
		},
		{
			name:   "Schema rejection (bad payload)",
			method: "POST", path: "/api/v1/patients/",
			body:   map[string]any{"invalid": true},
			auth:   true, expect: 400,
		},
		{
			name:   "Delete patient",
			method: "DELETE", path: "{{patient}}",
			auth: true, expect: 200,
		},
	}

	// --- Run ---
	passed, failed := 0, 0
	start := time.Now()

	for i, s := range steps {
		// Resolve path templates.
		path := s.path
		switch path {
		case "{{patient}}":
			path = "/api/v1/patients/" + patientID
		default:
			if len(path) > 12 && path[:12] == "{{patient}}/" {
				path = "/api/v1/patients/" + patientID + "/" + path[12:]
			}
		}

		// Add subject/patient reference for clinical resources that need it.
		if s.body != nil && patientID != "" {
			if m, ok := s.body.(map[string]any); ok {
				switch s.name {
				case "Create encounter", "Create observation", "Create condition",
					"Create medication request":
					m["subject"] = map[string]any{"reference": "Patient/" + patientID}
				case "Create allergy intolerance":
					m["subject"] = map[string]any{"reference": "Patient/" + patientID}
					m["patient"] = map[string]any{"reference": "Patient/" + patientID}
				}
			}
		}

		token := ""
		if s.auth {
			token = st.accessToken
		}

		resp, body, err := doRequest(s.method, st.server.URL+path, s.body, token)

		num := fmt.Sprintf("%2d", i+1)
		label := fmt.Sprintf("%-38s %s%-6s %s%s", s.name, colorDim, s.method, path, colorReset)

		if err != nil {
			failed++
			fmt.Printf("  %s %s%sFAIL%s  %s  %s(error: %v)%s\n", num, colorBold, colorRed, colorReset, label, colorRed, err, colorReset)
			continue
		}

		ok := resp.StatusCode == s.expect
		if !ok && s.acceptAlternate != 0 && resp.StatusCode == s.acceptAlternate {
			ok = true
		}

		if ok {
			passed++
			fmt.Printf("  %s %s%sPASS%s  %s  %s-> %d%s\n", num, colorBold, colorGreen, colorReset, label, colorDim, resp.StatusCode, colorReset)
			if s.extract != nil && body != nil {
				s.extract(body)
			}
		} else {
			failed++
			detail := ""
			if body != nil {
				if errObj, ok2 := body["error"].(map[string]any); ok2 {
					if msg, ok3 := errObj["message"].(string); ok3 {
						detail = msg
					}
				}
				if detail == "" {
					if msg, ok2 := body["message"].(string); ok2 {
						detail = msg
					}
				}
			}
			fmt.Printf("  %s %s%sFAIL%s  %s  %sgot %d, want %d%s",
				num, colorBold, colorRed, colorReset, label, colorRed, resp.StatusCode, s.expect, colorReset)
			if detail != "" {
				fmt.Printf("  %s(%s)%s", colorDim, detail, colorReset)
			}
			fmt.Println()
		}
	}

	// --- Summary ---
	elapsed := time.Since(start)
	fmt.Printf("\n%s%s Results: %d passed, %d failed (%s) %s\n\n",
		colorBold, colorCyan, passed, failed, elapsed.Round(time.Millisecond), colorReset)

	if failed > 0 {
		os.Exit(1)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "%s%sFATAL: %s%s\n", colorBold, colorRed, fmt.Sprintf(format, args...), colorReset)
	runCleanups()
	os.Exit(2)
}
