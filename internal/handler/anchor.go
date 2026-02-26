package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type AnchorHandler struct {
	svc service.AnchorService
}

func NewAnchorHandler(svc service.AnchorService) *AnchorHandler {
	return &AnchorHandler{svc: svc}
}

// Status handles GET /api/v1/anchor/status
func (h *AnchorHandler) Status(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetStatus(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Verify handles POST /api/v1/anchor/verify
func (h *AnchorHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CommitHash string `json:"commit_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.Verify(r.Context(), req.CommitHash)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// History handles GET /api/v1/anchor/history
func (h *AnchorHandler) History(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.GetHistory(r.Context(), page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Events, pg)
}

// Trigger handles POST /api/v1/anchor/trigger
func (h *AnchorHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.TriggerAnchor(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
