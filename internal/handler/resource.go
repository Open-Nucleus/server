package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// ResourceHandler handles CRUD for top-level resources (Practitioner, Organization, Location).
type ResourceHandler struct {
	svc service.PatientService
}

// NewResourceHandler creates a new handler for top-level resource CRUD.
func NewResourceHandler(svc service.PatientService) *ResourceHandler {
	return &ResourceHandler{svc: svc}
}

// ListFactory returns a handler for listing resources of a given type.
func (h *ResourceHandler) ListFactory(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page, perPage := model.PaginationFromRequest(r)
		resp, err := h.svc.ListResources(r.Context(), resourceType, page, perPage)
		if err != nil {
			model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
			return
		}
		pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
		model.SuccessWithPagination(w, resp.Resources, pg)
	}
}

// GetFactory returns a handler for getting a single resource by ID.
func (h *ResourceHandler) GetFactory(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourceID := chi.URLParam(r, "rid")
		resp, err := h.svc.GetResource(r.Context(), resourceType, resourceID)
		if err != nil {
			model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
			return
		}
		model.Success(w, http.StatusOK, resp)
	}
}

// CreateFactory returns a handler for creating a resource.
func (h *ResourceHandler) CreateFactory(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
			return
		}
		resp, err := h.svc.CreateResource(r.Context(), resourceType, body)
		if err != nil {
			model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
			return
		}
		writeResponseWithGit(w, http.StatusCreated, resp)
	}
}

// UpdateFactory returns a handler for updating a resource.
func (h *ResourceHandler) UpdateFactory(resourceType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resourceID := chi.URLParam(r, "rid")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
			return
		}
		_ = resourceID // ID is in the FHIR JSON body for UPDATE
		resp, err := h.svc.UpdateResource(r.Context(), resourceType, resourceID, body)
		if err != nil {
			model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
			return
		}
		writeResponseWithGit(w, http.StatusOK, resp)
	}
}

// CapabilityStatementHandler returns the FHIR CapabilityStatement.
func CapabilityStatementHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cs, err := fhir.GenerateCapabilityStatement(fhir.CapabilityConfig{
			ServerName:   "Open Nucleus",
			ServerURL:    "http://localhost:8080",
			Version:      "0.9.0",
			SmartEnabled: true,
			SmartBaseURL: "http://localhost:8080",
		})
		if err != nil {
			model.WriteError(w, model.ErrInternal, "Failed to generate CapabilityStatement", nil)
			return
		}
		w.Header().Set("Content-Type", "application/fhir+json")
		w.WriteHeader(http.StatusOK)
		w.Write(cs)
	}
}
