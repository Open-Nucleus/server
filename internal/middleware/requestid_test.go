package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestID_GeneratesUUID(t *testing.T) {
	var capturedID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = model.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequestID(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.NotEmpty(t, capturedID)
	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
	assert.Equal(t, capturedID, rr.Header().Get("X-Request-ID"))
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	var capturedID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = model.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequestID(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "custom-id-123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, "custom-id-123", capturedID)
	assert.Equal(t, "custom-id-123", rr.Header().Get("X-Request-ID"))
}
