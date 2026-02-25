package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requestWithClaims(role string, permissions []string) *http.Request {
	claims := &model.NucleusClaims{
		Role:        role,
		Permissions: permissions,
	}
	ctx := context.WithValue(context.Background(), model.CtxClaims, claims)
	return httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
}

func TestRBAC_AllowsPermittedRole(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequirePermission(model.PermPatientRead)(inner)
	req := requestWithClaims("physician", []string{"patient:read", "patient:write"})
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRBAC_DeniesUnpermittedRole(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequirePermission(model.PermPatientWrite)(inner)
	// CHW doesn't have patient:write
	req := requestWithClaims("community_health_worker", []string{"patient:read", "observation:read"})
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrInsufficientPerms, env.Error.Code)
}

func TestRBAC_DeniesNoClaims(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequirePermission(model.PermPatientRead)(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil) // no claims in context
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRBAC_FallsBackToRoleMatrix(t *testing.T) {
	// Token has no permissions in claims, but role "physician" has patient:read in the matrix
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.RequirePermission(model.PermPatientRead)(inner)
	req := requestWithClaims("physician", nil) // no explicit permissions
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRBAC_PerRolePermissions(t *testing.T) {
	tests := []struct {
		role       string
		permission string
		allowed    bool
	}{
		{"community_health_worker", "patient:read", true},
		{"community_health_worker", "patient:write", false},
		{"community_health_worker", "observation:write", true},
		{"nurse", "encounter:write", true},
		{"nurse", "patient:write", false},
		{"nurse", "medication:read", true},
		{"physician", "patient:write", true},
		{"physician", "conflict:resolve", true},
		{"physician", "sync:trigger", false},
		{"site_administrator", "sync:trigger", true},
		{"site_administrator", "anchor:trigger", true},
		{"regional_administrator", "supply:write", true},
	}

	for _, tt := range tests {
		t.Run(tt.role+"/"+tt.permission, func(t *testing.T) {
			assert.Equal(t, tt.allowed, model.HasPermission(tt.role, tt.permission))
		})
	}
}
