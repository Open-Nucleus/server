package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthService implements service.AuthService for testing.
type mockAuthService struct {
	loginFn   func(ctx context.Context, req *service.LoginRequest) (*service.LoginResponse, error)
	refreshFn func(ctx context.Context, token string) (*service.RefreshResponse, error)
	logoutFn  func(ctx context.Context, token string) error
	whoamiFn  func(ctx context.Context) (*service.WhoamiResponse, error)
}

func (m *mockAuthService) Login(ctx context.Context, req *service.LoginRequest) (*service.LoginResponse, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthService) Refresh(ctx context.Context, token string) (*service.RefreshResponse, error) {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, token)
	}
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, token)
	}
	return nil
}

func (m *mockAuthService) Whoami(ctx context.Context) (*service.WhoamiResponse, error) {
	if m.whoamiFn != nil {
		return m.whoamiFn(ctx)
	}
	return nil, nil
}

func TestAuthHandler_Login_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(ctx context.Context, req *service.LoginRequest) (*service.LoginResponse, error) {
			assert.Equal(t, "node-sheffield-01", req.DeviceID)
			return &service.LoginResponse{
				Token:        "test-token",
				ExpiresAt:    "2026-02-26T09:00:00Z",
				RefreshToken: "test-refresh",
				Role:         service.RoleDTO{Code: "physician", Display: "Physician"},
				SiteID:       "clinic-maiduguri-03",
				NodeID:       "node-sheffield-01",
			}, nil
		},
	}

	h := handler.NewAuthHandler(svc)
	body := `{"device_id":"node-sheffield-01","public_key":"test","challenge_response":{"nonce":"abc","signature":"def","timestamp":"2026-02-25T09:00:00Z"},"practitioner_id":"dr-adeleye"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
}

func TestAuthHandler_Login_InvalidBody(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthService{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrValidation, env.Error.Code)
}

func TestAuthHandler_Whoami_FromClaims(t *testing.T) {
	h := handler.NewAuthHandler(&mockAuthService{})

	claims := &model.NucleusClaims{
		Node:        "node-sheffield-01",
		Site:        "clinic-maiduguri-03",
		Role:        "physician",
		Permissions: []string{"patient:read"},
	}
	claims.Subject = "dr-adeleye"

	ctx := context.WithValue(context.Background(), model.CtxClaims, claims)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/whoami", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	h.Whoami(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)

	data, ok := env.Data.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "dr-adeleye", data["subject"])
	assert.Equal(t, "node-sheffield-01", data["node_id"])
}
