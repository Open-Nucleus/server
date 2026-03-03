package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	fhirhandler "github.com/FibrinLab/open-nucleus/internal/handler/fhir"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/model"
)

// Config holds all dependencies needed by the router.
type Config struct {
	AuthHandler      *handler.AuthHandler
	PatientHandler   *handler.PatientHandler
	ResourceHandler  *handler.ResourceHandler
	SyncHandler      *handler.SyncHandler
	ConflictHandler  *handler.ConflictHandler
	SentinelHandler  *handler.SentinelHandler
	FormularyHandler *handler.FormularyHandler
	AnchorHandler    *handler.AnchorHandler
	SupplyHandler    *handler.SupplyHandler
	SmartHandler     *handler.SmartHandler
	FHIRHandler      *fhirhandler.FHIRHandler
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

	// SMART on FHIR discovery (public, no auth)
	if cfg.SmartHandler != nil {
		r.Get("/.well-known/smart-configuration", cfg.SmartHandler.SmartConfiguration)
	}

	// FHIR metadata endpoint (no auth — public discovery)
	r.Get("/fhir/metadata", handler.CapabilityStatementHandler())

	// FHIR R4 REST API (/fhir/{Type}) — authenticated, raw FHIR JSON
	if cfg.FHIRHandler != nil {
		cfg.FHIRHandler.RegisterRoutes(r, cfg.JWTAuth, cfg.RateLimiter)
	}

	// SMART OAuth2 endpoints (outside /api/v1 — standard OAuth2 paths)
	if cfg.SmartHandler != nil {
		r.Route("/auth/smart", func(r chi.Router) {
			r.Use(cfg.RateLimiter.Middleware(middleware.CategoryAuth))
			r.With(cfg.JWTAuth.Middleware).Get("/authorize", cfg.SmartHandler.Authorize)
			r.Post("/token", cfg.SmartHandler.Token)
			r.With(cfg.JWTAuth.Middleware).Post("/revoke", cfg.SmartHandler.Revoke)
			r.With(cfg.JWTAuth.Middleware).Post("/introspect", cfg.SmartHandler.Introspect)
			r.With(cfg.JWTAuth.Middleware, middleware.RequirePermission(model.PermSmartRegister)).Post("/register", cfg.SmartHandler.Register)
			r.With(cfg.JWTAuth.Middleware, middleware.RequirePermission(model.PermSmartLaunch)).Post("/launch", cfg.SmartHandler.Launch)
		})
	}

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

					// Immunizations
					r.Route("/immunizations", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermEncounterRead),
						).Get("/", cfg.PatientHandler.ListImmunizations)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermEncounterWrite),
							validatorMiddleware(cfg.SchemaValidator, "immunization"),
						).Post("/", cfg.PatientHandler.CreateImmunization)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermEncounterRead),
						).Get("/{iid}", cfg.PatientHandler.GetImmunization)
					})

					// Procedures
					r.Route("/procedures", func(r chi.Router) {
						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermEncounterRead),
						).Get("/", cfg.PatientHandler.ListProcedures)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryWrite),
							middleware.RequirePermission(model.PermEncounterWrite),
							validatorMiddleware(cfg.SchemaValidator, "procedure"),
						).Post("/", cfg.PatientHandler.CreateProcedure)

						r.With(
							cfg.RateLimiter.Middleware(middleware.CategoryRead),
							middleware.RequirePermission(model.PermEncounterRead),
						).Get("/{pid}", cfg.PatientHandler.GetProcedure)
					})
				})
			})

			// Practitioner endpoints (top-level)
			r.Route("/practitioners", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/", cfg.ResourceHandler.ListFactory("Practitioner"))

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermPatientWrite),
				).Post("/", cfg.ResourceHandler.CreateFactory("Practitioner"))

				r.Route("/{rid}", func(r chi.Router) {
					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryRead),
						middleware.RequirePermission(model.PermPatientRead),
					).Get("/", cfg.ResourceHandler.GetFactory("Practitioner"))

					r.With(
						cfg.RateLimiter.Middleware(middleware.CategoryWrite),
						middleware.RequirePermission(model.PermPatientWrite),
					).Put("/", cfg.ResourceHandler.UpdateFactory("Practitioner"))
				})
			})

			// Organization endpoints (top-level)
			r.Route("/organizations", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/", cfg.ResourceHandler.ListFactory("Organization"))

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermPatientWrite),
				).Post("/", cfg.ResourceHandler.CreateFactory("Organization"))

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/{rid}", cfg.ResourceHandler.GetFactory("Organization"))
			})

			// Location endpoints (top-level)
			r.Route("/locations", func(r chi.Router) {
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/", cfg.ResourceHandler.ListFactory("Location"))

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermPatientWrite),
				).Post("/", cfg.ResourceHandler.CreateFactory("Location"))

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermPatientRead),
				).Get("/{rid}", cfg.ResourceHandler.GetFactory("Location"))
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
				// Drug lookup
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications", cfg.FormularyHandler.SearchMedications)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications/category/{category}", cfg.FormularyHandler.ListMedicationsByCategory)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/medications/{code}", cfg.FormularyHandler.GetMedication)

				// Safety checks
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyRead),
				).Post("/check-interactions", cfg.FormularyHandler.CheckInteractions)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyRead),
				).Post("/check-allergies", cfg.FormularyHandler.CheckAllergyConflicts)

				// Dosing (stub)
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyRead),
				).Post("/dosing/validate", cfg.FormularyHandler.ValidateDosing)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/dosing/options", cfg.FormularyHandler.GetDosingOptions)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyRead),
				).Post("/dosing/schedule", cfg.FormularyHandler.GenerateSchedule)

				// Stock management
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/stock/{site_id}/{medication_code}", cfg.FormularyHandler.GetStockLevel)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyWrite),
				).Put("/stock/{site_id}/{medication_code}", cfg.FormularyHandler.UpdateStockLevel)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/stock/{site_id}/{medication_code}/prediction", cfg.FormularyHandler.GetStockPrediction)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermFormularyWrite),
				).Post("/deliveries", cfg.FormularyHandler.RecordDelivery)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/redistribution/{medication_code}", cfg.FormularyHandler.GetRedistributionSuggestions)

				// Formulary metadata
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermFormularyRead),
				).Get("/info", cfg.FormularyHandler.GetFormularyInfo)
			})

			// Anchor/IOTA endpoints
			r.Route("/anchor", func(r chi.Router) {
				// Anchoring
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

				// DID
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/did/node", cfg.AnchorHandler.NodeDID)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/did/device/{device_id}", cfg.AnchorHandler.DeviceDID)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/did/resolve", cfg.AnchorHandler.ResolveDID)

				// Credentials
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermAnchorTrigger),
				).Post("/credentials/issue", cfg.AnchorHandler.IssueCredential)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryWrite),
					middleware.RequirePermission(model.PermAnchorRead),
				).Post("/credentials/verify", cfg.AnchorHandler.VerifyCredentialHandler)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/credentials", cfg.AnchorHandler.ListCredentials)

				// Backend
				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/backends", cfg.AnchorHandler.ListBackends)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/backends/{name}", cfg.AnchorHandler.BackendStatus)

				r.With(
					cfg.RateLimiter.Middleware(middleware.CategoryRead),
					middleware.RequirePermission(model.PermAnchorRead),
				).Get("/queue", cfg.AnchorHandler.QueueStatus)
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

			// SMART client management (admin only)
			if cfg.SmartHandler != nil {
				r.Route("/smart/clients", func(r chi.Router) {
					r.With(middleware.RequirePermission(model.PermDeviceManage)).Get("/", cfg.SmartHandler.ListClients)
					r.With(middleware.RequirePermission(model.PermDeviceManage)).Get("/{id}", cfg.SmartHandler.GetClient)
					r.With(middleware.RequirePermission(model.PermDeviceManage)).Put("/{id}", cfg.SmartHandler.UpdateClient)
					r.With(middleware.RequirePermission(model.PermDeviceManage)).Delete("/{id}", cfg.SmartHandler.DeleteClient)
				})
			}

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
