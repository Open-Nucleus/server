package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/pkg/smart"
)

// SmartScope returns middleware that enforces SMART v2 scope restrictions on FHIR endpoints.
// If the token has no Scope claim (i.e., it's a regular device token), the middleware passes through.
// If the token has a Scope claim, it must allow the given interaction on the given resource type.
// For patient-context scopes, it also enforces that only the launch patient's data is accessed.
func SmartScope(resourceType, interaction string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := model.ClaimsFromContext(r.Context())
			if claims == nil {
				// No claims at all — JWT middleware should have caught this.
				next.ServeHTTP(w, r)
				return
			}

			// If no SMART scope on this token, pass through (existing RBAC handles it).
			if claims.Scope == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Parse SMART scopes.
			scopes, err := smart.ParseScopes(claims.Scope)
			if err != nil {
				model.WriteError(w, model.ErrInsufficientPerms,
					"Invalid SMART scope in token", nil)
				return
			}

			// Check if any scope allows the interaction on this resource type.
			allowed := false
			hasPatientContext := false
			for _, sc := range scopes {
				if sc.Allows(interaction, resourceType) {
					allowed = true
					if sc.Context == "patient" {
						hasPatientContext = true
					}
				}
			}

			if !allowed {
				model.WriteError(w, model.ErrInsufficientPerms,
					"SMART scope does not permit this operation",
					map[string]string{
						"resource":    resourceType,
						"interaction": interaction,
						"scope":       claims.Scope,
					})
				return
			}

			// For patient-context scopes, enforce that only the launch patient's data is accessed.
			if hasPatientContext && claims.LaunchPatient != "" {
				requestedPatient := extractPatientID(r)
				if requestedPatient != "" && requestedPatient != claims.LaunchPatient {
					model.WriteError(w, model.ErrInsufficientPerms,
						"SMART patient scope restricts access to launch patient only",
						map[string]string{
							"launch_patient":    claims.LaunchPatient,
							"requested_patient": requestedPatient,
						})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractPatientID tries to find the patient ID from the request.
// It checks the URL path parameter {id} (for /fhir/Patient/{id}), the "patient" query param,
// and the "subject" query param.
func extractPatientID(r *http.Request) string {
	// Check if this is a Patient resource endpoint with an ID in the path.
	if id := chi.URLParam(r, "id"); id != "" {
		// Only relevant if the path contains /Patient/
		if strings.Contains(r.URL.Path, "/Patient/") {
			return id
		}
	}

	// Check query parameters.
	if p := r.URL.Query().Get("patient"); p != "" {
		return p
	}
	if s := r.URL.Query().Get("subject"); s != "" {
		// subject might be "Patient/xxx" reference format.
		if strings.HasPrefix(s, "Patient/") {
			return s[len("Patient/"):]
		}
		return s
	}

	return ""
}
