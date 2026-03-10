package middleware

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/pkg/consent"
)

// ConsentCheck returns middleware that enforces consent-based access control
// on patient-scoped routes. It runs AFTER JWT auth and extracts the patient_id
// from the URL path and device_id from JWT claims.
//
// Break-glass: if the X-Break-Glass header is "true", an emergency consent is
// auto-created with a 4h expiry and an audit log entry is generated.
func ConsentCheck(mgr *consent.Manager, logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If consent manager is nil, skip (backward compatibility)
			if mgr == nil {
				next.ServeHTTP(w, r)
				return
			}

			patientID := chi.URLParam(r, "id")
			if patientID == "" {
				// Not a patient-scoped route, skip consent check
				next.ServeHTTP(w, r)
				return
			}

			claims := model.ClaimsFromContext(r.Context())
			if claims == nil {
				model.WriteError(w, model.ErrAuthRequired, "No JWT token provided", nil)
				return
			}

			performerID := claims.DeviceID
			if performerID == "" {
				performerID = claims.Subject
			}

			decision, err := mgr.CheckAccess(patientID, performerID, claims.Role)
			if err != nil {
				logger.Error("consent check error", "error", err, "patient_id", patientID, "device_id", performerID)
				model.WriteError(w, model.ErrInternal, "consent check failed", nil)
				return
			}

			if decision.Allowed {
				logger.Debug("consent check passed",
					"patient_id", patientID,
					"device_id", performerID,
					"consent_id", decision.ConsentID,
					"reason", decision.Reason,
				)
				next.ServeHTTP(w, r)
				return
			}

			// Check break-glass header
			if r.Header.Get("X-Break-Glass") == "true" {
				_, _, err := mgr.GrantEmergencyConsent(patientID, performerID)
				if err != nil {
					logger.Error("break-glass consent creation failed", "error", err)
					model.WriteError(w, model.ErrInternal, "break-glass consent creation failed", nil)
					return
				}
				logger.Warn("BREAK-GLASS access granted",
					"patient_id", patientID,
					"device_id", performerID,
					"request_id", model.RequestIDFromContext(r.Context()),
				)
				next.ServeHTTP(w, r)
				return
			}

			logger.Info("consent check denied",
				"patient_id", patientID,
				"device_id", performerID,
				"reason", decision.Reason,
			)
			model.WriteError(w, model.ErrConsentRequired,
				"No active consent grant for this patient",
				map[string]string{
					"patient_id": patientID,
					"device_id":  performerID,
					"reason":     decision.Reason,
				})
		})
	}
}
