package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type FormularyHandler struct {
	svc service.FormularyService
}

func NewFormularyHandler(svc service.FormularyService) *FormularyHandler {
	return &FormularyHandler{svc: svc}
}

// SearchMedications handles GET /api/v1/formulary/medications
func (h *FormularyHandler) SearchMedications(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)
	query := r.URL.Query().Get("q")

	resp, err := h.svc.SearchMedications(r.Context(), query, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Medications, pg)
}

// GetMedication handles GET /api/v1/formulary/medications/{code}
func (h *FormularyHandler) GetMedication(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	resp, err := h.svc.GetMedication(r.Context(), code)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// CheckInteractions handles POST /api/v1/formulary/check-interactions
func (h *FormularyHandler) CheckInteractions(w http.ResponseWriter, r *http.Request) {
	var req service.CheckInteractionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.CheckInteractions(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GetAvailability handles GET /api/v1/formulary/availability/{site_id}
func (h *FormularyHandler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	siteID := chi.URLParam(r, "site_id")

	resp, err := h.svc.GetAvailability(r.Context(), siteID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// UpdateAvailability handles PUT /api/v1/formulary/availability/{site_id}
func (h *FormularyHandler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	siteID := chi.URLParam(r, "site_id")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.UpdateAvailability(r.Context(), siteID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
