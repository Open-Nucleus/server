package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type SentinelHandler struct {
	svc service.SentinelService
}

func NewSentinelHandler(svc service.SentinelService) *SentinelHandler {
	return &SentinelHandler{svc: svc}
}

// ListAlerts handles GET /api/v1/alerts
func (h *SentinelHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListAlerts(r.Context(), page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Alerts, pg)
}

// Summary handles GET /api/v1/alerts/summary
func (h *SentinelHandler) Summary(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetAlertSummary(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GetAlert handles GET /api/v1/alerts/{id}
func (h *SentinelHandler) GetAlert(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "id")

	resp, err := h.svc.GetAlert(r.Context(), alertID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Acknowledge handles POST /api/v1/alerts/{id}/acknowledge
func (h *SentinelHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "id")

	resp, err := h.svc.AcknowledgeAlert(r.Context(), alertID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Dismiss handles POST /api/v1/alerts/{id}/dismiss
func (h *SentinelHandler) Dismiss(w http.ResponseWriter, r *http.Request) {
	alertID := chi.URLParam(r, "id")

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.svc.DismissAlert(r.Context(), alertID, req.Reason)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
