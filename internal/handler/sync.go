package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type SyncHandler struct {
	svc service.SyncService
}

func NewSyncHandler(svc service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

// Status handles GET /api/v1/sync/status
func (h *SyncHandler) Status(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetStatus(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Peers handles GET /api/v1/sync/peers
func (h *SyncHandler) Peers(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListPeers(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Trigger handles POST /api/v1/sync/trigger
func (h *SyncHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TargetNode string `json:"target_node"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.svc.TriggerSync(r.Context(), req.TargetNode)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// History handles GET /api/v1/sync/history
func (h *SyncHandler) History(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.GetHistory(r.Context(), page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Events, pg)
}

// ExportBundle handles POST /api/v1/sync/bundle/export
func (h *SyncHandler) ExportBundle(w http.ResponseWriter, r *http.Request) {
	var req service.BundleExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.ExportBundle(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// ImportBundle handles POST /api/v1/sync/bundle/import
func (h *SyncHandler) ImportBundle(w http.ResponseWriter, r *http.Request) {
	var req service.BundleImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.ImportBundle(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
