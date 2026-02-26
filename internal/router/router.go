package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/model"
)

// Config holds all dependencies needed by the router.
type Config struct {
	AuthHandler      *handler.AuthHandler
	PatientHandler   *handler.PatientHandler
	SyncHandler      *handler.SyncHandler
	ConflictHandler  *handler.ConflictHandler
	SentinelHandler  *handler.SentinelHandler
	FormularyHandler *handler.FormularyHandler
	AnchorHandler    *handler.AnchorHandler
	SupplyHandler    *handler.SupplyHandler
	SchemaValidator  *middleware.SchemaValidator
	JWTAuth          *middleware.JWTAuth
	RateLimiter      *middleware.RateLimiter
	CORSOrigins      []string
	AuditLogger      *slog.Logger
}

// New creates the full route tree with middleware scoping.
func New(cfg Config) http.Handler {
	r := chi.NewRouter()

	// Global middleware (all routes)
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.RequestID)
	r.Use(middleware.AuditLog(cfg.AuditLogger))

	// Health check (no auth needed)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		model.Success(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes — rate limiter + request ID only (NO JWT/RBAC)
		r.Route("/auth", func(r chi.Router) {
			r.Use(cfg.RateLimiter.Middleware(middleware.CategoryAuth))
			r.Post("/login", cfg.AuthHandler.Login)
			r.Post("/refresh", cfg.AuthHandler.Refresh)
			r.Post("/logout", cfg.AuthHandler.Logout)
			r.Get("/whoami", cfg.AuthHandler.Whoami)
		})

		// All other routes — full pipeline (JWT + RBAC + rate limiting)
		r.Group(func(r chi.Router) {
			r.Use(cfg.JWTAuth.Middleware)

			// Patient endpoints
			r.Route("/patients", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/", cfg.PatientHandler.List)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermPatientWrite),
					validatorMiddleware(cfg.SchemaValidator, "patient"),
				).Post("/", cfg.PatientHandler.Create)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/search", cfg.PatientHandler.Search)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermPatientRead),
				).Post("/match", cfg.PatientHandler.Match)

				r.Route("/{id}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/", cfg.PatientHandler.GetByID)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermPatientWrite),
						validatorMiddleware(cfg.SchemaValidator, "patient"),
					).Put("/", cfg.PatientHandler.Update)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermPatientWrite),
					).Delete("/", cfg.PatientHandler.Delete)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/history", cfg.PatientHandler.History)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/timeline", cfg.PatientHandler.Timeline)

					// Encounters
					r.Route("/encounters", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermEncounterRead),
						).Get("/", cfg.PatientHandler.ListEncounters)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermEncounterWrite),
							validatorMiddleware(cfg.SchemaValidator, "encounter"),
						).Post("/", cfg.PatientHandler.CreateEncounter)

						r.Route("/{eid}", func(r chi.Router) {
							r.With(
								cfg.RateLimiter.Middleware(middleware.CategoryRead),
								middleware.RequirePermission(model.PermEncounterRead),
							).Get("/", cfg.PatientHandler.GetEncounter)

							r.With(
								cfg.RateLimiter.Middleware(middleware.CategoryWrite),
								middleware.RequirePermission(model.PermEncounterWrite),
								validatorMiddleware(cfg.SchemaValidator, "encounter"),
							).Put("/", cfg.PatientHandler.UpdateEncounter)
						})
					})

					// Observations
					r.Route("/observations", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermObservationRead),
						).Get("/", cfg.PatientHandler.ListObservations)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermObservationWrite),
							validatorMiddleware(cfg.SchemaValidator, "observation"),
						).Post("/", cfg.PatientHandler.CreateObservation)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermObservationRead),
						).Get("/{oid}", cfg.PatientHandler.GetObservation)
					})

					// Conditions
					r.Route("/conditions", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermConditionRead),
						).Get("/", cfg.PatientHandler.ListConditions)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermConditionWrite),
							validatorMiddleware(cfg.SchemaValidator, "condition"),
						).Post("/", cfg.PatientHandler.CreateCondition)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermConditionWrite),
							validatorMiddleware(cfg.SchemaValidator, "condition"),
						).Put("/{cid}", cfg.PatientHandler.UpdateCondition)
					})

					// Medication Requests
					r.Route("/medication-requests", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermMedicationRead),
						).Get("/", cfg.PatientHandler.ListMedicationRequests)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermMedicationWrite),
							validatorMiddleware(cfg.SchemaValidator, "medication_request"),
						).Post("/", cfg.PatientHandler.CreateMedicationRequest)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermMedicationWrite),
							validatorMiddleware(cfg.SchemaValidator, "medication_request"),
						).Put("/{mid}", cfg.PatientHandler.UpdateMedicationRequest)
					})

					// Allergy Intolerances
					r.Route("/allergy-intolerances", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermAllergyRead),
						).Get("/", cfg.PatientHandler.ListAllergyIntolerances)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermAllergyWrite),
							validatorMiddleware(cfg.SchemaValidator, "allergy_intolerance"),
						).Post("/", cfg.PatientHandler.CreateAllergyIntolerance)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermAllergyWrite),
							validatorMiddleware(cfg.SchemaValidator, "allergy_intolerance"),
						).Put("/{aid}", cfg.PatientHandler.UpdateAllergyIntolerance)
					})
				})
			})

			// Sync endpoints
			r.Route("/sync", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSyncRead),
				).Get("/status", cfg.SyncHandler.Status)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSyncRead),
				).Get("/peers", cfg.SyncHandler.Peers)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermSyncTrigger),
				).Post("/trigger", cfg.SyncHandler.Trigger)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSyncRead),
				).Get("/history", cfg.SyncHandler.History)

				r.Route("/bundle", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermSyncTrigger),
					).Post("/export", cfg.SyncHandler.ExportBundle)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermSyncTrigger),
					).Post("/import", cfg.SyncHandler.ImportBundle)
				})
			})

			// Conflict endpoints
			r.Route("/conflicts", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermConflictRead),
				).Get("/", cfg.ConflictHandler.List)

				r.Route("/{id}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermConflictRead),
					).Get("/", cfg.ConflictHandler.GetByID)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermConflictResolve),
					).Post("/resolve", cfg.ConflictHandler.Resolve)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermConflictResolve),
					).Post("/defer", cfg.ConflictHandler.Defer)
				})
			})

			// Alert endpoints
			r.Route("/alerts", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAlertRead),
				).Get("/", cfg.SentinelHandler.ListAlerts)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAlertRead),
				).Get("/summary", cfg.SentinelHandler.Summary)

				r.Route("/{id}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermAlertRead),
					).Get("/", cfg.SentinelHandler.GetAlert)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermAlertWrite),
					).Post("/acknowledge", cfg.SentinelHandler.Acknowledge)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermAlertWrite),
					).Post("/dismiss", cfg.SentinelHandler.Dismiss)
				})
			})

			// Formulary endpoints
			r.Route("/formulary", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications", cfg.FormularyHandler.SearchMedications)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications/{code}", cfg.FormularyHandler.GetMedication)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyRead),
				).Post("/check-interactions", cfg.FormularyHandler.CheckInteractions)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/availability/{site_id}", cfg.FormularyHandler.GetAvailability)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyWrite),
				).Put("/availability/{site_id}", cfg.FormularyHandler.UpdateAvailability)
			})

			// Anchor/IOTA endpoints
			r.Route("/anchor", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/status", cfg.AnchorHandler.Status)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermAnchorRead),
				).Post("/verify", cfg.AnchorHandler.Verify)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/history", cfg.AnchorHandler.History)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermAnchorTrigger),
				).Post("/trigger", cfg.AnchorHandler.Trigger)
			})

			// Supply chain endpoints
			r.Route("/supply", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/inventory", cfg.SupplyHandler.Inventory)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/inventory/{item_code}", cfg.SupplyHandler.InventoryItem)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermSupplyWrite),
				).Post("/deliveries", cfg.SupplyHandler.RecordDelivery)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/predictions", cfg.SupplyHandler.Predictions)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/redistribution", cfg.SupplyHandler.Redistribution)
			})

			// WebSocket endpoint (stub for now — Phase 5)
			r.Get("/ws", handler.StubHandler())
		})
	})

	return r
}

// validatorMiddleware returns a no-op middleware if SchemaValidator is nil,
// otherwise applies the schema validation for the given pattern.
func validatorMiddleware(sv *middleware.SchemaValidator, pattern string) func(http.Handler) http.Handler {
	if sv == nil {
		return func(next http.Handler) http.Handler { return next }
	}
	return sv.Middleware(pattern)
}
