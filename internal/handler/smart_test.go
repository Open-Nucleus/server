package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

// mockSmartService implements service.SmartService for testing.
type mockSmartService struct {
	authorizeFn     func(ctx context.Context, req *service.AuthorizeRequest) (*service.AuthorizeResponse, error)
	exchangeTokenFn func(ctx context.Context, req *service.ExchangeTokenRequest) (*service.TokenResponse, error)
	introspectFn    func(ctx context.Context, token string) (*service.IntrospectResponse, error)
	revokeTokenFn   func(ctx context.Context, token string) error
	registerFn      func(ctx context.Context, req *service.RegisterClientRequest) (*service.ClientResponse, error)
	listClientsFn   func(ctx context.Context) (*service.ClientListResponse, error)
	getClientFn     func(ctx context.Context, clientID string) (*service.ClientResponse, error)
	updateClientFn  func(ctx context.Context, clientID string, req *service.UpdateClientRequest) (*service.ClientResponse, error)
	deleteClientFn  func(ctx context.Context, clientID string) error
	createLaunchFn  func(ctx context.Context, req *service.CreateLaunchRequest) (*service.CreateLaunchResponse, error)
}

func (m *mockSmartService) Authorize(ctx context.Context, req *service.AuthorizeRequest) (*service.AuthorizeResponse, error) {
	if m.authorizeFn != nil {
		return m.authorizeFn(ctx, req)
	}
	return nil, nil
}

func (m *mockSmartService) ExchangeToken(ctx context.Context, req *service.ExchangeTokenRequest) (*service.TokenResponse, error) {
	if m.exchangeTokenFn != nil {
		return m.exchangeTokenFn(ctx, req)
	}
	return nil, nil
}

func (m *mockSmartService) IntrospectToken(ctx context.Context, token string) (*service.IntrospectResponse, error) {
	if m.introspectFn != nil {
		return m.introspectFn(ctx, token)
	}
	return nil, nil
}

func (m *mockSmartService) RevokeToken(ctx context.Context, token string) error {
	if m.revokeTokenFn != nil {
		return m.revokeTokenFn(ctx, token)
	}
	return nil
}

func (m *mockSmartService) RegisterClient(ctx context.Context, req *service.RegisterClientRequest) (*service.ClientResponse, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, req)
	}
	return nil, nil
}

func (m *mockSmartService) ListClients(ctx context.Context) (*service.ClientListResponse, error) {
	if m.listClientsFn != nil {
		return m.listClientsFn(ctx)
	}
	return nil, nil
}

func (m *mockSmartService) GetClient(ctx context.Context, clientID string) (*service.ClientResponse, error) {
	if m.getClientFn != nil {
		return m.getClientFn(ctx, clientID)
	}
	return nil, nil
}

func (m *mockSmartService) UpdateClient(ctx context.Context, clientID string, req *service.UpdateClientRequest) (*service.ClientResponse, error) {
	if m.updateClientFn != nil {
		return m.updateClientFn(ctx, clientID, req)
	}
	return nil, nil
}

func (m *mockSmartService) DeleteClient(ctx context.Context, clientID string) error {
	if m.deleteClientFn != nil {
		return m.deleteClientFn(ctx, clientID)
	}
	return nil
}

func (m *mockSmartService) CreateLaunch(ctx context.Context, req *service.CreateLaunchRequest) (*service.CreateLaunchResponse, error) {
	if m.createLaunchFn != nil {
		return m.createLaunchFn(ctx, req)
	}
	return nil, nil
}

func TestSmartHandler_SmartConfiguration(t *testing.T) {
	h := handler.NewSmartHandler(&mockSmartService{}, "http://localhost:8080")
	req := httptest.NewRequest(http.MethodGet, "/.well-known/smart-configuration", nil)
	rr := httptest.NewRecorder()

	h.SmartConfiguration(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var cfg map[string]any
	err := json.NewDecoder(rr.Body).Decode(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/auth/smart/authorize", cfg["authorization_endpoint"])
	assert.Equal(t, "http://localhost:8080/auth/smart/token", cfg["token_endpoint"])

	caps, ok := cfg["capabilities"].([]any)
	require.True(t, ok)
	assert.Contains(t, caps, "launch-ehr")
	assert.Contains(t, caps, "permission-v2")
}

func TestSmartHandler_Authorize_Success(t *testing.T) {
	svc := &mockSmartService{
		authorizeFn: func(ctx context.Context, req *service.AuthorizeRequest) (*service.AuthorizeResponse, error) {
			assert.Equal(t, "test-client", req.ClientID)
			assert.Equal(t, "http://localhost:3000/callback", req.RedirectURI)
			assert.Equal(t, "patient/Observation.rs", req.Scope)
			return &service.AuthorizeResponse{
				RedirectURI: "http://localhost:3000/callback?code=abc123&state=xyz",
			}, nil
		},
	}

	h := handler.NewSmartHandler(svc, "http://localhost:8080")

	claims := &model.NucleusClaims{
		Node: "node-01",
		Role: "physician",
	}
	claims.Subject = "dr-test"
	ctx := context.WithValue(context.Background(), model.CtxClaims, claims)

	req := httptest.NewRequest(http.MethodGet, "/auth/smart/authorize?client_id=test-client&redirect_uri=http://localhost:3000/callback&scope=patient/Observation.rs&state=xyz&code_challenge=abc&code_challenge_method=S256", nil)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Authorize(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
}

func TestSmartHandler_Authorize_NoClaims(t *testing.T) {
	h := handler.NewSmartHandler(&mockSmartService{}, "http://localhost:8080")
	req := httptest.NewRequest(http.MethodGet, "/auth/smart/authorize?client_id=test", nil)
	rr := httptest.NewRecorder()

	h.Authorize(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestSmartHandler_Token_FormEncoded(t *testing.T) {
	svc := &mockSmartService{
		exchangeTokenFn: func(ctx context.Context, req *service.ExchangeTokenRequest) (*service.TokenResponse, error) {
			assert.Equal(t, "authorization_code", req.GrantType)
			assert.Equal(t, "code123", req.Code)
			assert.Equal(t, "test-client", req.ClientID)
			return &service.TokenResponse{
				AccessToken: "smart-token-xyz",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
				Scope:       "patient/Observation.rs",
				Patient:     "patient-001",
			}, nil
		},
	}

	h := handler.NewSmartHandler(svc, "http://localhost:8080")
	body := "grant_type=authorization_code&code=code123&redirect_uri=http://localhost:3000/callback&code_verifier=verifier123&client_id=test-client"
	req := httptest.NewRequest(http.MethodPost, "/auth/smart/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	h.Token(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "no-store", rr.Header().Get("Cache-Control"))

	var resp service.TokenResponse
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "smart-token-xyz", resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, "patient-001", resp.Patient)
}

func TestSmartHandler_Token_Error(t *testing.T) {
	svc := &mockSmartService{
		exchangeTokenFn: func(ctx context.Context, req *service.ExchangeTokenRequest) (*service.TokenResponse, error) {
			return nil, errors.New("invalid code")
		},
	}

	h := handler.NewSmartHandler(svc, "http://localhost:8080")
	body := "grant_type=authorization_code&code=bad"
	req := httptest.NewRequest(http.MethodPost, "/auth/smart/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	h.Token(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSmartHandler_Register(t *testing.T) {
	svc := &mockSmartService{
		registerFn: func(ctx context.Context, req *service.RegisterClientRequest) (*service.ClientResponse, error) {
			assert.Equal(t, "Growth Chart App", req.ClientName)
			assert.Equal(t, "patient/Observation.rs", req.Scope)
			return &service.ClientResponse{
				ClientID:   "new-client-id",
				ClientName: "Growth Chart App",
				Status:     "pending",
			}, nil
		},
	}

	h := handler.NewSmartHandler(svc, "http://localhost:8080")
	body := `{"client_name":"Growth Chart App","redirect_uris":["http://localhost:3000/callback"],"scope":"patient/Observation.rs","grant_types":["authorization_code"],"token_endpoint_auth_method":"none","launch_modes":["ehr"]}`
	req := httptest.NewRequest(http.MethodPost, "/auth/smart/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
}

func TestSmartHandler_Launch(t *testing.T) {
	svc := &mockSmartService{
		createLaunchFn: func(ctx context.Context, req *service.CreateLaunchRequest) (*service.CreateLaunchResponse, error) {
			assert.Equal(t, "test-client", req.ClientID)
			assert.Equal(t, "patient-001", req.PatientID)
			return &service.CreateLaunchResponse{
				LaunchToken: "launch-token-abc",
			}, nil
		},
	}

	h := handler.NewSmartHandler(svc, "http://localhost:8080")
	body := `{"client_id":"test-client","patient_id":"patient-001","encounter_id":"enc-001"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/smart/launch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Launch(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestSmartHandler_GetClient(t *testing.T) {
	svc := &mockSmartService{
		getClientFn: func(ctx context.Context, clientID string) (*service.ClientResponse, error) {
			assert.Equal(t, "client-xyz", clientID)
			return &service.ClientResponse{
				ClientID:   "client-xyz",
				ClientName: "Test App",
				Status:     "approved",
			}, nil
		},
	}

	h := handler.NewSmartHandler(svc, "http://localhost:8080")

	r := chi.NewRouter()
	r.Get("/api/v1/smart/clients/{id}", h.GetClient)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/smart/clients/client-xyz", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
}
