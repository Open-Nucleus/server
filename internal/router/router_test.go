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

type testPatientSvc struct{}

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
		AuthHandler:    handler.NewAuthHandler(&testAuthSvc{}),
		PatientHandler: handler.NewPatientHandler(&testPatientSvc{}),
		JWTAuth:        jwtAuth,
		RateLimiter:    rl,
		CORSOrigins:    []string{"http://localhost:*"},
		AuditLogger:    auditLogger,
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
		Permissions: []string{"patient:read", "patient:write", "encounter:read", "encounter:write"},
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

func TestIntegration_UnimplementedEndpoint_Returns501(t *testing.T) {
	mux, _, priv := setupTestRouter(t)
	token := makeToken(t, priv)

	// POST /patients (stub)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotImplemented, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrNotImplemented, env.Error.Code)
}

func TestIntegration_RequestID_IsSet(t *testing.T) {
	mux, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}
