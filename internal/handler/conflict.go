package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type ConflictHandler struct {
	svc service.ConflictService
}

func NewConflictHandler(svc service.ConflictService) *ConflictHandler {
	return &ConflictHandler{svc: svc}
}

// List handles GET /api/v1/conflicts
func (h *ConflictHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListConflicts(r.Context(), page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Conflicts, pg)
}

// GetByID handles GET /api/v1/conflicts/{id}
func (h *ConflictHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	conflictID := chi.URLParam(r, "id")

	resp, err := h.svc.GetConflict(r.Context(), conflictID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Resolve handles POST /api/v1/conflicts/{id}/resolve
func (h *ConflictHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	conflictID := chi.URLParam(r, "id")

	var req service.ResolveConflictRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}
	req.ConflictID = conflictID

	resp, err := h.svc.ResolveConflict(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Defer handles POST /api/v1/conflicts/{id}/defer
func (h *ConflictHandler) Defer(w http.ResponseWriter, r *http.Request) {
	conflictID := chi.URLParam(r, "id")

	var req service.DeferConflictRequest
	json.NewDecoder(r.Body).Decode(&req)
	req.ConflictID = conflictID

	resp, err := h.svc.DeferConflict(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
