package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCORS_AllowsLocalhost(t *testing.T) {
	origins := []string{"http://localhost:*", "http://127.0.0.1:*"}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.CORS(origins)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.Equal(t, "http://localhost:3000", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_BlocksExternalOrigin(t *testing.T) {
	origins := []string{"http://localhost:*", "http://127.0.0.1:*"}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.CORS(origins)(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_HandlesPreflightOptions(t *testing.T) {
	origins := []string{"http://localhost:*"}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := middleware.CORS(origins)(inner)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:8080")
	rr := httptest.NewRecorder()

	mw.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "http://localhost:8080", rr.Header().Get("Access-Control-Allow-Origin"))
}
