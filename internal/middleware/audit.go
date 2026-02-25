package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/model"
)

// statusWriter wraps ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// AuditLog returns middleware that logs every request to the audit logger.
func AuditLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)

			duration := time.Since(start)
			requestID := model.RequestIDFromContext(r.Context())
			user := "anonymous"
			if claims := model.ClaimsFromContext(r.Context()); claims != nil {
				user = claims.Subject
			}

			logger.Info("request",
				"request_id", requestID,
				"timestamp", start.UTC().Format(time.RFC3339),
				"user", user,
				"method", r.Method,
				"endpoint", r.URL.Path,
				"status_code", sw.status,
				"duration_ms", duration.Milliseconds(),
				"ip", r.RemoteAddr,
			)
		})
	}
}
