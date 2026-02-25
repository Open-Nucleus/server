package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"golang.org/x/time/rate"
)

// EndpointCategory determines rate limit tier.
type EndpointCategory int

const (
	CategoryRead EndpointCategory = iota
	CategoryWrite
	CategoryAuth
)

type limiterEntry struct {
	limiter   *rate.Limiter
	lastSeen  time.Time
}

// RateLimiter implements per-device token bucket rate limiting.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]map[EndpointCategory]*limiterEntry
	cfg      config.RateLimitConfig
}

func NewRateLimiter(cfg config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]map[EndpointCategory]*limiterEntry),
		cfg:      cfg,
	}
}

func (rl *RateLimiter) getLimiter(deviceID string, cat EndpointCategory) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if _, ok := rl.limiters[deviceID]; !ok {
		rl.limiters[deviceID] = make(map[EndpointCategory]*limiterEntry)
	}

	entry, ok := rl.limiters[deviceID][cat]
	if !ok {
		var rpm, burst int
		switch cat {
		case CategoryAuth:
			rpm, burst = rl.cfg.AuthRPM, rl.cfg.AuthBurst
		case CategoryWrite:
			rpm, burst = rl.cfg.WriteRPM, rl.cfg.WriteBurst
		default:
			rpm, burst = rl.cfg.ReadRPM, rl.cfg.ReadBurst
		}
		lim := rate.NewLimiter(rate.Every(time.Minute/time.Duration(rpm)), burst)
		entry = &limiterEntry{limiter: lim, lastSeen: time.Now()}
		rl.limiters[deviceID][cat] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

// Middleware returns a rate limiting middleware for a given endpoint category.
func (rl *RateLimiter) Middleware(cat EndpointCategory) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Identify device from JWT sub claim, or use remote addr for unauthenticated
			deviceID := "anonymous"
			if claims := model.ClaimsFromContext(r.Context()); claims != nil {
				deviceID = claims.Subject
			}

			limiter := rl.getLimiter(deviceID, cat)

			// Set rate limit headers
			var rpmLimit int
			switch cat {
			case CategoryAuth:
				rpmLimit = rl.cfg.AuthRPM
			case CategoryWrite:
				rpmLimit = rl.cfg.WriteRPM
			default:
				rpmLimit = rl.cfg.ReadRPM
			}
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rpmLimit))

			if !limiter.Allow() {
				retryAfter := time.Minute / time.Duration(rpmLimit)
				w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
				w.Header().Set("X-RateLimit-Remaining", "0")
				model.WriteError(w, model.ErrRateLimited,
					"Too many requests, retry after indicated duration", nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
