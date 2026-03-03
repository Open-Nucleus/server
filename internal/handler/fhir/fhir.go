package fhir

import (
	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/service"
	pkgfhir "github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// FHIRHandler provides standards-compliant FHIR R4 REST endpoints.
type FHIRHandler struct {
	patientSvc  service.PatientService
	dispatchers map[string]*ResourceDispatch
}

// NewFHIRHandler creates a new FHIR handler backed by the patient service.
func NewFHIRHandler(patientSvc service.PatientService) *FHIRHandler {
	return &FHIRHandler{
		patientSvc:  patientSvc,
		dispatchers: buildDispatchers(patientSvc),
	}
}

// RegisterRoutes dynamically registers FHIR routes from the resource registry.
func (h *FHIRHandler) RegisterRoutes(r chi.Router, jwtAuth *middleware.JWTAuth, rateLimiter *middleware.RateLimiter) {
	r.Route("/fhir", func(r chi.Router) {
		r.Use(ContentNegotiation)
		r.Use(jwtAuth.Middleware)

		for _, def := range pkgfhir.AllResourceDefs() {
			rt := def.Type
			disp := h.dispatchers[rt]
			if disp == nil {
				continue
			}

			if hasInteraction(def, "search-type") && disp.Search != nil {
				r.With(rateLimiter.Middleware(middleware.CategoryRead), middleware.RequirePermission(disp.ReadPerm), middleware.SmartScope(rt, "s")).
					Get("/"+rt, h.Search(rt))
			}
			if hasInteraction(def, "read") && disp.Read != nil {
				r.With(rateLimiter.Middleware(middleware.CategoryRead), middleware.RequirePermission(disp.ReadPerm), middleware.SmartScope(rt, "r")).
					Get("/"+rt+"/{id}", h.Read(rt))
			}
			if hasInteraction(def, "create") && disp.Create != nil {
				r.With(rateLimiter.Middleware(middleware.CategoryWrite), middleware.RequirePermission(disp.WritePerm), middleware.SmartScope(rt, "c")).
					Post("/"+rt, h.Create(rt))
			}
			if hasInteraction(def, "update") && disp.Update != nil {
				r.With(rateLimiter.Middleware(middleware.CategoryWrite), middleware.RequirePermission(disp.WritePerm), middleware.SmartScope(rt, "u")).
					Put("/"+rt+"/{id}", h.Update(rt))
			}
			if hasInteraction(def, "delete") && disp.Delete != nil {
				r.With(rateLimiter.Middleware(middleware.CategoryWrite), middleware.RequirePermission(disp.WritePerm), middleware.SmartScope(rt, "d")).
					Delete("/"+rt+"/{id}", h.Delete(rt))
			}
		}

		// Patient $everything
		r.With(rateLimiter.Middleware(middleware.CategoryRead), middleware.RequirePermission("patient:read"), middleware.SmartScope("Patient", "r")).
			Get("/Patient/{id}/$everything", h.Everything)
	})
}

func hasInteraction(def *pkgfhir.ResourceDef, interaction string) bool {
	for _, i := range def.Interactions {
		if i == interaction {
			return true
		}
	}
	return false
}
