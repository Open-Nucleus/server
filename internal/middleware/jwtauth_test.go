package middleware_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return pub, priv
}

func signTestToken(t *testing.T, priv ed25519.PrivateKey, claims model.NucleusClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := token.SignedString(priv)
	require.NoError(t, err)
	return signed
}

func validClaims() model.NucleusClaims {
	return model.NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "dr-adeleye",
			Issuer:    "open-nucleus-auth",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Node:        "node-sheffield-01",
		Site:        "clinic-maiduguri-03",
		Role:        "physician",
		Permissions: []string{"patient:read", "patient:write"},
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	pub, priv := generateTestKey(t)
	auth := middleware.NewJWTAuth(pub, "open-nucleus-auth")

	var capturedClaims *model.NucleusClaims
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedClaims = model.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	claims := validClaims()
	tokenStr := signTestToken(t, priv, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	auth.Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, "dr-adeleye", capturedClaims.Subject)
	assert.Equal(t, "physician", capturedClaims.Role)
	assert.Equal(t, "node-sheffield-01", capturedClaims.Node)
}

func TestJWTAuth_MissingToken(t *testing.T) {
	pub, _ := generateTestKey(t)
	auth := middleware.NewJWTAuth(pub, "open-nucleus-auth")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	auth.Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "error", env.Status)
	assert.Equal(t, model.ErrAuthRequired, env.Error.Code)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	pub, priv := generateTestKey(t)
	auth := middleware.NewJWTAuth(pub, "open-nucleus-auth")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	claims := validClaims()
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))
	tokenStr := signTestToken(t, priv, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	auth.Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrTokenExpired, env.Error.Code)
}

func TestJWTAuth_RevokedToken(t *testing.T) {
	pub, priv := generateTestKey(t)
	auth := middleware.NewJWTAuth(pub, "open-nucleus-auth")

	// Revoke the subject
	auth.RevokeToken("dr-adeleye")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	claims := validClaims()
	tokenStr := signTestToken(t, priv, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	auth.Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, model.ErrTokenRevoked, env.Error.Code)
}

func TestJWTAuth_WrongIssuer(t *testing.T) {
	pub, priv := generateTestKey(t)
	auth := middleware.NewJWTAuth(pub, "open-nucleus-auth")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	claims := validClaims()
	claims.Issuer = "wrong-issuer"
	tokenStr := signTestToken(t, priv, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	auth.Middleware(inner).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
