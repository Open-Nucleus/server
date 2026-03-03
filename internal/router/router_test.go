package router_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/router"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testAuthSvc struct{}

func (t *testAuthSvc) Login(_ context.Context, req *service.LoginRequest) (*service.LoginResponse, error) {
	return &service.LoginResponse{
		Token:        "mock-token",
		ExpiresAt:    "2026-02-26T09:00:00Z",
		RefreshToken: "mock-refresh",
		Role:         service.RoleDTO{Code: "physician", Display: "Physician"},
		SiteID:       "clinic-maiduguri-03",
		NodeID:       "node-sheffield-01",
	}, nil
}
func (t *testAuthSvc) Refresh(_ context.Context, _ string) (*service.RefreshResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
func (t *testAuthSvc) Logout(_ context.Context, _ string) error { return nil }
func (t *testAuthSvc) Whoami(_ context.Context) (*service.WhoamiResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

type testPatientSvc struct{ service.PatientService }

func (t *testPatientSvc) ListPatients(_ context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
	return &service.ListPatientsResponse{
		Patients:   []any{map[string]string{"id": "patient-001", "resourceType": "Patient"}},
		Page:       req.Page,
		PerPage:    req.PerPage,
		Total:      1,
		TotalPages: 1,
	}, nil
}

func (t *testPatientSvc) GetPatient(_ context.Context, id string) (*service.PatientBundle, error) {
	return &service.PatientBundle{
		Patient: map[string]string{"id": id, "resourceType": "Patient"},
	}, nil
}

func (t *testPatientSvc) SearchPatients(_ context.Context, _ string, _, _ int) (*service.ListPatientsResponse, error) {
	return &service.ListPatientsResponse{Patients: []any{}, Total: 0, TotalPages: 0}, nil
}

func (t *testPatientSvc) CreatePatient(_ context.Context, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) UpdatePatient(_ context.Context, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) DeletePatient(_ context.Context, _ string) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) MatchPatients(_ context.Context, _ *service.MatchPatientsRequest) (*service.MatchPatientsResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) GetPatientHistory(_ context.Context, _ string) (*service.PatientHistoryResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) GetPatientTimeline(_ context.Context, _ string) (*service.PatientTimelineResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) ListEncounters(_ context.Context, _ string, _, _ int) (*service.ClinicalListResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) GetEncounter(_ context.Context, _, _ string) (any, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) CreateEncounter(_ context.Context, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) UpdateEncounter(_ context.Context, _, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) ListObservations(_ context.Context, _ string, _ service.ObservationFilters, _, _ int) (*service.ClinicalListResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) GetObservation(_ context.Context, _, _ string) (any, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) CreateObservation(_ context.Context, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) ListConditions(_ context.Context, _ string, _ service.ConditionFilters, _, _ int) (*service.ClinicalListResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) CreateCondition(_ context.Context, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) UpdateCondition(_ context.Context, _, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) ListMedicationRequests(_ context.Context, _ string, _, _ int) (*service.ClinicalListResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) CreateMedicationRequest(_ context.Context, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) UpdateMedicationRequest(_ context.Context, _, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) ListAllergyIntolerances(_ context.Context, _ string, _, _ int) (*service.ClinicalListResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) CreateAllergyIntolerance(_ context.Context, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
func (t *testPatientSvc) UpdateAllergyIntolerance(_ context.Context, _, _ string, _ json.RawMessage) (*service.WriteResponse, error) {
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

// Stub implementations for other services

type testSyncSvc struct{}

func (t *testSyncSvc) GetStatus(_ context.Context) (*service.SyncStatusResponse, error) {
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}
func (t *testSyncSvc) ListPeers(_ context.Context) (*service.SyncPeersResponse, error) {
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}
func (t *testSyncSvc) TriggerSync(_ context.Context, _ string) (*service.SyncTriggerResponse, error) {
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}
func (t *testSyncSvc) GetHistory(_ context.Context, _, _ int) (*service.SyncHistoryResponse, error) {
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}
func (t *testSyncSvc) ExportBundle(_ context.Context, _ *service.BundleExportRequest) (*service.BundleExportResponse, error) {
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}
func (t *testSyncSvc) ImportBundle(_ context.Context, _ *service.BundleImportRequest) (*service.BundleImportResponse, error) {
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}

type testConflictSvc struct{}

func (t *testConflictSvc) ListConflicts(_ context.Context, _, _ int) (*service.ConflictListResponse, error) {
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}
func (t *testConflictSvc) GetConflict(_ context.Context, _ string) (*service.ConflictDetail, error) {
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}
func (t *testConflictSvc) ResolveConflict(_ context.Context, _ *service.ResolveConflictRequest) (*service.ResolveConflictResponse, error) {
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}
func (t *testConflictSvc) DeferConflict(_ context.Context, _ *service.DeferConflictRequest) (*service.DeferConflictResponse, error) {
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}

type testSentinelSvc struct{}

func (t *testSentinelSvc) ListAlerts(_ context.Context, _, _ int) (*service.AlertListResponse, error) {
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}
func (t *testSentinelSvc) GetAlertSummary(_ context.Context) (*service.AlertSummaryResponse, error) {
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}
func (t *testSentinelSvc) GetAlert(_ context.Context, _ string) (*service.AlertDetail, error) {
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}
func (t *testSentinelSvc) AcknowledgeAlert(_ context.Context, _ string) (*service.AlertDetail, error) {
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}
func (t *testSentinelSvc) DismissAlert(_ context.Context, _, _ string) (*service.AlertDetail, error) {
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}

type testFormularySvc struct{}

func (t *testFormularySvc) SearchMedications(_ context.Context, _, _ string, _, _ int) (*service.MedicationListResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GetMedication(_ context.Context, _ string) (*service.MedicationDetail, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) ListMedicationsByCategory(_ context.Context, _ string, _, _ int) (*service.MedicationListResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) CheckInteractions(_ context.Context, _ *service.CheckInteractionsRequest) (*service.CheckInteractionsResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) CheckAllergyConflicts(_ context.Context, _ *service.CheckAllergyConflictsRequest) (*service.CheckAllergyConflictsResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) ValidateDosing(_ context.Context, _ *service.ValidateDosingRequest) (*service.ValidateDosingResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GetDosingOptions(_ context.Context, _ string, _ float64) (*service.GetDosingOptionsResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GenerateSchedule(_ context.Context, _ *service.GenerateScheduleRequest) (*service.GenerateScheduleResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GetStockLevel(_ context.Context, _, _ string) (*service.StockLevelResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) UpdateStockLevel(_ context.Context, _ *service.UpdateStockLevelRequest) (*service.UpdateStockLevelResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) RecordDelivery(_ context.Context, _ *service.FormularyDeliveryRequest) (*service.FormularyDeliveryResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GetStockPrediction(_ context.Context, _, _ string) (*service.StockPredictionResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GetRedistributionSuggestions(_ context.Context, _ string) (*service.FormularyRedistributionResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
func (t *testFormularySvc) GetFormularyInfo(_ context.Context) (*service.FormularyInfoResponse, error) {
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}

type testAnchorSvc struct{}

func (t *testAnchorSvc) GetStatus(_ context.Context) (*service.AnchorStatusResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) Verify(_ context.Context, _ string) (*service.AnchorVerifyResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) GetHistory(_ context.Context, _, _ int) (*service.AnchorHistoryResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) TriggerAnchor(_ context.Context) (*service.AnchorTriggerResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) GetNodeDID(_ context.Context) (*service.DIDDocumentResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) GetDeviceDID(_ context.Context, _ string) (*service.DIDDocumentResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) ResolveDID(_ context.Context, _ string) (*service.DIDDocumentResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) IssueDataIntegrityCredential(_ context.Context, _ *service.IssueCredentialRequest) (*service.CredentialResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) VerifyCredential(_ context.Context, _ string) (*service.CredentialVerificationResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) ListCredentials(_ context.Context, _ string, _, _ int) (*service.CredentialListResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) ListBackends(_ context.Context) (*service.BackendListResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) GetBackendStatus(_ context.Context, _ string) (*service.BackendStatusResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) GetQueueStatus(_ context.Context) (*service.QueueStatusResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
func (t *testAnchorSvc) Health(_ context.Context) (*service.AnchorHealthResponse, error) {
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}

type testSupplySvc struct{}

func (t *testSupplySvc) GetInventory(_ context.Context, _, _ int) (*service.InventoryListResponse, error) {
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}
func (t *testSupplySvc) GetInventoryItem(_ context.Context, _ string) (*service.InventoryItemDetail, error) {
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}
func (t *testSupplySvc) RecordDelivery(_ context.Context, _ *service.RecordDeliveryRequest) (*service.RecordDeliveryResponse, error) {
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}
func (t *testSupplySvc) GetPredictions(_ context.Context) (*service.PredictionsResponse, error) {
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}
func (t *testSupplySvc) GetRedistribution(_ context.Context) (*service.RedistributionResponse, error) {
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}

func setupTestRouter(t *testing.T) (http.Handler, ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	jwtAuth := middleware.NewJWTAuth(pub, "open-nucleus-auth")
	rl := middleware.NewRateLimiter(config.RateLimitConfig{
		ReadRPM: 200, ReadBurst: 50,
		WriteRPM: 60, WriteBurst: 20,
		AuthRPM: 10, AuthBurst: 5,
	})

	auditLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := router.New(router.Config{
		AuthHandler:      handler.NewAuthHandler(&testAuthSvc{}),
		PatientHandler:   handler.NewPatientHandler(&testPatientSvc{}),
		ResourceHandler:  handler.NewResourceHandler(&testPatientSvc{}),
		SyncHandler:      handler.NewSyncHandler(&testSyncSvc{}),
		ConflictHandler:  handler.NewConflictHandler(&testConflictSvc{}),
		SentinelHandler:  handler.NewSentinelHandler(&testSentinelSvc{}),
		FormularyHandler: handler.NewFormularyHandler(&testFormularySvc{}),
		AnchorHandler:    handler.NewAnchorHandler(&testAnchorSvc{}),
		SupplyHandler:    handler.NewSupplyHandler(&testSupplySvc{}),
		JWTAuth:          jwtAuth,
		RateLimiter:      rl,
		CORSOrigins:      []string{"http://localhost:*"},
		AuditLogger:      auditLogger,
	})

	return mux, pub, priv
}

func makeToken(t *testing.T, priv ed25519.PrivateKey) string {
	t.Helper()
	claims := model.NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "dr-adeleye",
			Issuer:    "open-nucleus-auth",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Node:        "node-sheffield-01",
		Site:        "clinic-maiduguri-03",
		Role:        "physician",
		Permissions: []string{"patient:read", "patient:write", "encounter:read", "encounter:write", "observation:read", "observation:write", "condition:read", "condition:write", "medication:read", "medication:write", "allergy:read", "allergy:write", "conflict:read", "conflict:resolve", "alert:read", "alert:write", "sync:read", "formulary:read", "anchor:read", "supply:read"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := token.SignedString(priv)
	require.NoError(t, err)
	return signed
}

func TestIntegration_HealthCheck(t *testing.T) {
	mux, _, _ := setupTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestIntegration_LoginAndListPatients(t *testing.T) {
	mux, _, priv := setupTestRouter(t)

	// Step 1: Login
	loginBody := `{"device_id":"node-sheffield-01","public_key":"test","challenge_response":{"nonce":"abc","signature":"def","timestamp":"2026-02-25T09:00:00Z"},"practitioner_id":"dr-adeleye"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	mux.ServeHTTP(loginRR, loginReq)

	assert.Equal(t, http.StatusOK, loginRR.Code)

	var loginEnv model.Envelope
	err := json.NewDecoder(loginRR.Body).Decode(&loginEnv)
	require.NoError(t, err)
	assert.Equal(t, "success", loginEnv.Status)

	// Step 2: List patients with JWT
	token := makeToken(t, priv)
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRR := httptest.NewRecorder()
	mux.ServeHTTP(listRR, listReq)

	assert.Equal(t, http.StatusOK, listRR.Code)

	var listEnv model.Envelope
	err = json.NewDecoder(listRR.Body).Decode(&listEnv)
	require.NoError(t, err)
	assert.Equal(t, "success", listEnv.Status)
	require.NotNil(t, listEnv.Pagination)
}

func TestIntegration_PatientsWithoutJWT_Returns401(t *testing.T) {
	mux, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrAuthRequired, env.Error.Code)
}

func TestIntegration_RequestID_IsSet(t *testing.T) {
	mux, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}

func TestIntegration_PostPatient_Returns503(t *testing.T) {
	mux, _, priv := setupTestRouter(t)
	token := makeToken(t, priv)

	body := `{"resourceType":"Patient","name":[{"family":"Okafor"}],"gender":"male"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrServiceUnavailable, env.Error.Code)
}

func TestIntegration_EncountersList_Returns503(t *testing.T) {
	mux, _, priv := setupTestRouter(t)
	token := makeToken(t, priv)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001/encounters", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrServiceUnavailable, env.Error.Code)
}

func TestIntegration_NoMore501s(t *testing.T) {
	mux, _, priv := setupTestRouter(t)
	token := makeToken(t, priv)

	// Test a selection of formerly-stubbed routes
	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPut, "/api/v1/patients/p1", `{"resourceType":"Patient","name":[{"family":"Test"}],"gender":"male"}`},
		{http.MethodDelete, "/api/v1/patients/p1", ""},
		{http.MethodGet, "/api/v1/patients/p1/history", ""},
		{http.MethodGet, "/api/v1/patients/p1/timeline", ""},
		{http.MethodGet, "/api/v1/patients/p1/observations", ""},
		{http.MethodGet, "/api/v1/patients/p1/conditions", ""},
		{http.MethodGet, "/api/v1/patients/p1/medication-requests", ""},
		{http.MethodGet, "/api/v1/patients/p1/allergy-intolerances", ""},
		{http.MethodGet, "/api/v1/sync/status", ""},
		{http.MethodGet, "/api/v1/sync/peers", ""},
		{http.MethodGet, "/api/v1/alerts", ""},
		{http.MethodGet, "/api/v1/alerts/summary", ""},
		{http.MethodGet, "/api/v1/formulary/medications", ""},
		{http.MethodGet, "/api/v1/anchor/status", ""},
		{http.MethodGet, "/api/v1/supply/inventory", ""},
		{http.MethodGet, "/api/v1/supply/predictions", ""},
		{http.MethodGet, "/api/v1/supply/redistribution", ""},
	}

	for _, tc := range routes {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			} else {
				bodyReader = strings.NewReader("")
			}

			req := httptest.NewRequest(tc.method, tc.path, bodyReader)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			// Should be 503 (service unavailable), NOT 501 (not implemented)
			assert.NotEqual(t, http.StatusNotImplemented, rr.Code,
				"Expected non-501 for %s %s, got %d", tc.method, tc.path, rr.Code)
		})
	}
}
