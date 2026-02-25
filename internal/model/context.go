package model

import "context"

type contextKey string

const (
	CtxRequestID contextKey = "request_id"
	CtxClaims    contextKey = "claims"
	CtxStartTime contextKey = "start_time"
)

// RequestIDFromContext extracts the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(CtxRequestID).(string)
	return v
}

// ClaimsFromContext extracts JWT claims from context.
func ClaimsFromContext(ctx context.Context) *NucleusClaims {
	v, _ := ctx.Value(CtxClaims).(*NucleusClaims)
	return v
}
