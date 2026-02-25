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
	AuthHandler    *handler.AuthHandler
	PatientHandler *handler.PatientHandler
	JWTAuth        *middleware.JWTAuth
	RateLimiter    *middleware.RateLimiter
	CORSOrigins    []string
	AuditLogger    *slog.Logger
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
				).Post("/", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/search", cfg.PatientHandler.Search)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermPatientRead),
				).Post("/match", handler.StubHandler())

				r.Route("/{id}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/", cfg.PatientHandler.GetByID)

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermPatientWrite),
					).Put("/", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermPatientWrite),
					).Delete("/", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/history", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/timeline", handler.StubHandler())

					// Encounters
					r.Route("/encounters", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermEncounterRead),
						).Get("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermEncounterWrite),
						).Post("/", handler.StubHandler())

						r.Route("/{eid}", func(r chi.Router) {
							r.With(
								cfg.RateLimiter.Middleware(middleware.CategoryRead),
								middleware.RequirePermission(model.PermEncounterRead),
							).Get("/", handler.StubHandler())

							r.With(
								cfg.RateLimiter.Middleware(middleware.CategoryWrite),
								middleware.RequirePermission(model.PermEncounterWrite),
							).Put("/", handler.StubHandler())
						})
					})

					// Observations
					r.Route("/observations", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermObservationRead),
						).Get("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermObservationWrite),
						).Post("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermObservationRead),
						).Get("/{oid}", handler.StubHandler())
					})

					// Conditions
					r.Route("/conditions", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermConditionRead),
						).Get("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermConditionWrite),
						).Post("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermConditionWrite),
						).Put("/{cid}", handler.StubHandler())
					})

					// Medication Requests
					r.Route("/medication-requests", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermMedicationRead),
						).Get("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermMedicationWrite),
						).Post("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermMedicationWrite),
						).Put("/{mid}", handler.StubHandler())
					})

					// Allergy Intolerances
					r.Route("/allergy-intolerances", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermAllergyRead),
						).Get("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermAllergyWrite),
						).Post("/", handler.StubHandler())

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermAllergyWrite),
						).Put("/{aid}", handler.StubHandler())
					})
				})
			})

			// Sync endpoints
			r.Route("/sync", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSyncRead),
				).Get("/status", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSyncRead),
				).Get("/peers", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermSyncTrigger),
				).Post("/trigger", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSyncRead),
				).Get("/history", handler.StubHandler())

				r.Route("/bundle", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermSyncTrigger),
					).Post("/export", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermSyncTrigger),
					).Post("/import", handler.StubHandler())
				})
			})

			// Conflict endpoints
			r.Route("/conflicts", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermConflictRead),
				).Get("/", handler.StubHandler())

				r.Route("/{id}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermConflictRead),
					).Get("/", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermConflictResolve),
					).Post("/resolve", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermConflictResolve),
					).Post("/defer", handler.StubHandler())
				})
			})

			// Alert endpoints
			r.Route("/alerts", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAlertRead),
				).Get("/", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAlertRead),
				).Get("/summary", handler.StubHandler())

				r.Route("/{id}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermAlertRead),
					).Get("/", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermAlertWrite),
					).Post("/acknowledge", handler.StubHandler())

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermAlertWrite),
					).Post("/dismiss", handler.StubHandler())
				})
			})

			// Formulary endpoints
			r.Route("/formulary", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications/{code}", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyRead),
				).Post("/check-interactions", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/availability/{site_id}", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyWrite),
				).Put("/availability/{site_id}", handler.StubHandler())
			})

			// Anchor/IOTA endpoints
			r.Route("/anchor", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/status", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermAnchorRead),
				).Post("/verify", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/history", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermAnchorTrigger),
				).Post("/trigger", handler.StubHandler())
			})

			// Supply chain endpoints
			r.Route("/supply", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/inventory", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/inventory/{item_code}", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermSupplyWrite),
				).Post("/deliveries", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/predictions", handler.StubHandler())

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermSupplyRead),
				).Get("/redistribution", handler.StubHandler())
			})

			// WebSocket endpoint (stub for now)
			r.Get("/ws", handler.StubHandler())
		})
	})

	return r
}
