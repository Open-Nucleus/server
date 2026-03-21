// Command seed populates the Open Nucleus data directory with demo data.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/auth/authservice"
	"github.com/FibrinLab/open-nucleus/services/patient/pipeline"

	_ "modernc.org/sqlite"
)

var ctx = context.Background()

func main() {
	repoPath := flag.String("repo", "data/repo", "Git repository path")
	dbPath := flag.String("db", "data/nucleus.db", "SQLite database path")
	bootstrap := flag.String("bootstrap-secret", "", "Bootstrap secret (or NUCLEUS_BOOTSTRAP_SECRET env)")
	flag.Parse()

	secret := *bootstrap
	if secret == "" {
		secret = os.Getenv("NUCLEUS_BOOTSTRAP_SECRET")
	}
	if secret == "" {
		secret = "demo"
	}

	git, err := gitstore.NewStore(*repoPath, "nucleus-seed", "seed@open-nucleus.local")
	if err != nil {
		log.Fatalf("git init: %v", err)
	}

	dsn := *dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_cache_size=-20000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("sqlite open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := sqliteindex.InitUnifiedSchema(db); err != nil {
		log.Fatalf("schema init: %v", err)
	}
	idx := sqliteindex.NewIndexFromDB(db)

	// --- Auth: register demo device ---
	ks := auth.NewMemoryKeyStore()
	denyList := authservice.NewDenyList(db)

	authCfg := &authservice.Config{
		JWT: authservice.JWTConfig{
			Issuer:          "open-nucleus-auth",
			AccessLifetime:  24 * time.Hour,
			RefreshLifetime: 48 * time.Hour,
			ClockSkew:       2 * time.Hour,
		},
		Git: authservice.GitConfig{
			RepoPath:    *repoPath,
			AuthorName:  "nucleus-seed",
			AuthorEmail: "seed@open-nucleus.local",
		},
		Devices: authservice.DevicesConfig{Path: ".nucleus/devices"},
		Security: authservice.SecurityConfig{
			NonceTTL:        60 * time.Second,
			MaxFailures:     10,
			FailureWindow:   60 * time.Second,
			BootstrapSecret: secret,
		},
		KeyStore: authservice.KeyStoreConfig{Type: "memory"},
		SQLite:   authservice.SQLiteConfig{DBPath: *dbPath},
	}

	authImpl, err := authservice.NewAuthService(authCfg, git, ks, denyList)
	if err != nil {
		log.Fatalf("auth init: %v", err)
	}

	pub, _, err := auth.GenerateKeypair()
	if err != nil {
		log.Fatalf("keygen: %v", err)
	}

	device, err := authImpl.RegisterDevice(
		auth.EncodePublicKey(pub), "demo-clinician", "site-alpha", "demo-tablet", "physician", secret,
	)
	if err != nil {
		log.Fatalf("register device: %v", err)
	}

	fmt.Printf("Device registered: %s\n", device.DeviceID)

	// --- Patient writer ---
	pw := pipeline.NewWriter(git, idx, 10*time.Second)
	mutCtx := pipeline.MutationContext{
		PractitionerID: "demo-clinician",
		NodeID:         "node-demo",
		SiteID:         "site-alpha",
		Timestamp:      time.Now().UTC(),
	}

	// --- Seed patients ---
	patients := []struct {
		given, family, gender, birthDate string
	}{
		{"Amina", "Okafor", "female", "1988-03-15"},
		{"Kwame", "Mensah", "male", "1975-11-02"},
		{"Fatou", "Diallo", "female", "1995-07-22"},
		{"Emmanuel", "Mutombo", "male", "1962-01-08"},
		{"Ngozi", "Eze", "female", "2001-09-30"},
		{"Ibrahim", "Toure", "male", "1990-05-17"},
	}

	patientIDs := make([]string, 0, len(patients))

	for _, p := range patients {
		body := map[string]any{
			"resourceType": "Patient",
			"name": []map[string]any{{
				"use":    "official",
				"family": p.family,
				"given":  []string{p.given},
			}},
			"gender":    p.gender,
			"birthDate": p.birthDate,
			"active":    true,
		}
		data, _ := json.Marshal(body)
		resp, err := pw.Write(ctx, "CREATE", "Patient", "", data, mutCtx)
		if err != nil {
			log.Fatalf("create patient %s %s: %v", p.given, p.family, err)
		}
		patientIDs = append(patientIDs, resp.ResourceID)
		fmt.Printf("  Patient: %s %s (%s)\n", p.given, p.family, resp.ResourceID)
	}

	// --- Seed clinical data for each patient ---
	vitalCodes := []struct {
		code, display, unit string
		value               float64
	}{
		{"8310-5", "Body temperature", "Cel", 37.2},
		{"8867-4", "Heart rate", "/min", 72},
		{"8480-6", "Systolic blood pressure", "mm[Hg]", 120},
		{"8462-4", "Diastolic blood pressure", "mm[Hg]", 80},
		{"29463-7", "Body weight", "kg", 68.5},
	}

	conditionCodes := []struct {
		code, display string
	}{
		{"386661006", "Fever"},
		{"49727002", "Cough"},
		{"38341003", "Hypertension"},
	}

	for i, pid := range patientIDs {
		// 2 encounters per patient
		for j := 0; j < 2; j++ {
			classes := []struct{ code, display string }{
				{"AMB", "ambulatory"},
				{"EMER", "emergency"},
			}
			ec := classes[j%2]
			enc := map[string]any{
				"resourceType": "Encounter",
				"status":       "finished",
				"class": map[string]any{
					"system":  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
					"code":    ec.code,
					"display": ec.display,
				},
				"subject": map[string]any{"reference": "Patient/" + pid},
				"period": map[string]any{
					"start": time.Now().AddDate(0, 0, -(i*7+j)).Format("2006-01-02"),
				},
			}
			data, _ := json.Marshal(enc)
			pw.Write(ctx, "CREATE", "Encounter", pid, data, mutCtx)
		}

		// 3 vitals per patient
		for j := 0; j < 3; j++ {
			v := vitalCodes[(i+j)%len(vitalCodes)]
			obs := map[string]any{
				"resourceType": "Observation",
				"status":       "final",
				"code": map[string]any{
					"coding": []map[string]any{{
						"system":  "http://loinc.org",
						"code":    v.code,
						"display": v.display,
					}},
				},
				"subject": map[string]any{"reference": "Patient/" + pid},
				"valueQuantity": map[string]any{
					"value":  v.value + float64(i)*0.3,
					"unit":   v.unit,
					"system": "http://unitsofmeasure.org",
					"code":   v.unit,
				},
				"effectiveDateTime": time.Now().AddDate(0, 0, -i).Format("2006-01-02T15:04:05Z"),
			}
			data, _ := json.Marshal(obs)
			pw.Write(ctx, "CREATE", "Observation", pid, data, mutCtx)
		}

		// 1 condition
		cc := conditionCodes[i%len(conditionCodes)]
		cond := map[string]any{
			"resourceType": "Condition",
			"clinicalStatus": map[string]any{
				"coding": []map[string]any{{
					"system": "http://terminology.hl7.org/CodeSystem/condition-clinical",
					"code":   "active",
				}},
			},
			"code": map[string]any{
				"coding": []map[string]any{{
					"system":  "http://snomed.info/sct",
					"code":    cc.code,
					"display": cc.display,
				}},
			},
			"subject":     map[string]any{"reference": "Patient/" + pid},
			"recordedDate": time.Now().AddDate(0, 0, -i*3).Format("2006-01-02"),
		}
		data, _ := json.Marshal(cond)
		pw.Write(ctx, "CREATE", "Condition", pid, data, mutCtx)

		// 1 medication request
		medReq := map[string]any{
			"resourceType": "MedicationRequest",
			"status":       "active",
			"intent":       "order",
			"medicationCodeableConcept": map[string]any{
				"coding": []map[string]any{{
					"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
					"code":    "161",
					"display": "Acetaminophen 500mg",
				}},
			},
			"subject":    map[string]any{"reference": "Patient/" + pid},
			"authoredOn": time.Now().AddDate(0, 0, -i).Format("2006-01-02"),
			"dosageInstruction": []map[string]any{{
				"text":   "500mg twice daily",
				"timing": map[string]any{"code": map[string]any{"text": "BD"}},
				"doseAndRate": []map[string]any{{
					"doseQuantity": map[string]any{"value": 500, "unit": "mg"},
				}},
			}},
		}
		data, _ = json.Marshal(medReq)
		pw.Write(ctx, "CREATE", "MedicationRequest", pid, data, mutCtx)
	}

	fmt.Println()
	fmt.Println("=== Demo Data Seeded ===")
	fmt.Printf("  Patients:     %d\n", len(patientIDs))
	fmt.Printf("  Encounters:   %d\n", len(patientIDs)*2)
	fmt.Printf("  Observations: %d\n", len(patientIDs)*3)
	fmt.Printf("  Conditions:   %d\n", len(patientIDs))
	fmt.Printf("  MedRequests:  %d\n", len(patientIDs))
	fmt.Println()
	fmt.Printf("  Device ID:    %s\n", device.DeviceID)
	fmt.Printf("  Bootstrap:    %s\n", secret)
	fmt.Println()
	fmt.Println("To start the server:")
	fmt.Printf("  NUCLEUS_BOOTSTRAP_SECRET=%s go run ./cmd/nucleus\n", secret)
}
