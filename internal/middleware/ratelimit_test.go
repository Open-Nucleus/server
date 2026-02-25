package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	cfg := config.RateLimitConfig{
		ReadRPM:   200,
		ReadBurst: 50,
		AuthRPM:   10,
		AuthBurst: 5,
	}
	rl := middleware.NewRateLimiter(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := rl.Middleware(middleware.CategoryRead)(inner)

	// Should allow the first request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimiter_BlocksWhenExceeded(t *testing.T) {
	cfg := config.RateLimitConfig{
		AuthRPM:   10,
		AuthBurst: 1, // burst of 1 = very tight
	}
	rl := middleware.NewRateLimiter(cfg)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := rl.Middleware(middleware.CategoryAuth)(inner)

	// First request — should succeed (uses the 1 burst token)
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr1 := httptest.NewRecorder()
	mw.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second request — should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rr2 := httptest.NewRecorder()
	mw.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
}
