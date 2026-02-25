package middleware

import (
	"net/http"

	"github.com/FibrinLab/open-nucleus/internal/model"
)

// RequirePermission returns middleware that checks if the authenticated user has the required permission.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := model.ClaimsFromContext(r.Context())
			if claims == nil {
				model.WriteError(w, model.ErrAuthRequired, "No JWT token provided", nil)
				return
			}

			// Check permissions from token claims
			hasPermission := false
			for _, p := range claims.Permissions {
				if p == permission {
					hasPermission = true
					break
				}
			}

			// Fall back to role-based check
			if !hasPermission {
				hasPermission = model.HasPermission(claims.Role, permission)
			}

			if !hasPermission {
				model.WriteError(w, model.ErrInsufficientPerms,
					"Role does not permit this operation", map[string]string{
						"role":       claims.Role,
						"required":   permission,
					})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
