package middleware

import (
	"context"
	"net/http"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/google/uuid"
)

// RequestID generates a UUID v4 request ID and stores it in context + response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), model.CtxRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
