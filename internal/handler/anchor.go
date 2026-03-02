package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

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
	model.SuccessWithPagination(w, resp.Records, pg)
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

// NodeDID handles GET /api/v1/anchor/did/node
func (h *AnchorHandler) NodeDID(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetNodeDID(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// DeviceDID handles GET /api/v1/anchor/did/device/{device_id}
func (h *AnchorHandler) DeviceDID(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "device_id")
	if deviceID == "" {
		model.WriteError(w, model.ErrValidation, "device_id is required", nil)
		return
	}

	resp, err := h.svc.GetDeviceDID(r.Context(), deviceID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// ResolveDID handles GET /api/v1/anchor/did/resolve?did=...
func (h *AnchorHandler) ResolveDID(w http.ResponseWriter, r *http.Request) {
	did := r.URL.Query().Get("did")
	if did == "" {
		model.WriteError(w, model.ErrValidation, "did query parameter is required", nil)
		return
	}

	resp, err := h.svc.ResolveDID(r.Context(), did)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// IssueCredential handles POST /api/v1/anchor/credentials/issue
func (h *AnchorHandler) IssueCredential(w http.ResponseWriter, r *http.Request) {
	var req service.IssueCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.IssueDataIntegrityCredential(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// VerifyCredentialHandler handles POST /api/v1/anchor/credentials/verify
func (h *AnchorHandler) VerifyCredentialHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CredentialJSON string `json:"credential_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.VerifyCredential(r.Context(), req.CredentialJSON)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// ListCredentials handles GET /api/v1/anchor/credentials?type=...
func (h *AnchorHandler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	credType := r.URL.Query().Get("type")
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListCredentials(r.Context(), credType, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Credentials, pg)
}

// ListBackends handles GET /api/v1/anchor/backends
func (h *AnchorHandler) ListBackends(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListBackends(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// BackendStatus handles GET /api/v1/anchor/backends/{name}
func (h *AnchorHandler) BackendStatus(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		model.WriteError(w, model.ErrValidation, "backend name is required", nil)
		return
	}

	resp, err := h.svc.GetBackendStatus(r.Context(), name)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// QueueStatus handles GET /api/v1/anchor/queue
func (h *AnchorHandler) QueueStatus(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetQueueStatus(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
