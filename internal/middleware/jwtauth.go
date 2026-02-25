package middleware

import (
	"context"
	"crypto/ed25519"
	"net/http"
	"strings"
	"sync"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth validates Ed25519-signed JWTs.
type JWTAuth struct {
	publicKey ed25519.PublicKey
	issuer    string
	mu        sync.RWMutex
	denyList  map[string]struct{} // token JTI or subject deny list
}

func NewJWTAuth(publicKey ed25519.PublicKey, issuer string) *JWTAuth {
	return &JWTAuth{
		publicKey: publicKey,
		issuer:    issuer,
		denyList:  make(map[string]struct{}),
	}
}

// RevokeToken adds a token subject to the deny list.
func (j *JWTAuth) RevokeToken(subject string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.denyList[subject] = struct{}{}
}

// IsRevoked checks if a subject is in the deny list.
func (j *JWTAuth) IsRevoked(subject string) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	_, ok := j.denyList[subject]
	return ok
}

// Middleware returns the JWT validation middleware.
func (j *JWTAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			model.WriteError(w, model.ErrAuthRequired, "No JWT token provided", nil)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			model.WriteError(w, model.ErrAuthRequired, "Invalid authorization header format", nil)
			return
		}
		tokenStr := parts[1]

		claims := &model.NucleusClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return j.publicKey, nil
		}, jwt.WithIssuer(j.issuer))

		if err != nil {
			if strings.Contains(err.Error(), "token is expired") {
				model.WriteError(w, model.ErrTokenExpired, "JWT has expired, use /auth/refresh", nil)
				return
			}
			model.WriteError(w, model.ErrAuthRequired, "Invalid token: "+err.Error(), nil)
			return
		}

		if !token.Valid {
			model.WriteError(w, model.ErrAuthRequired, "Invalid token", nil)
			return
		}

		// Check deny list
		if j.IsRevoked(claims.Subject) {
			model.WriteError(w, model.ErrTokenRevoked, "Device has been decommissioned", nil)
			return
		}

		ctx := context.WithValue(r.Context(), model.CtxClaims, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
