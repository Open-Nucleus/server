package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

// ConsentHandler handles consent HTTP endpoints.
type ConsentHandler struct {
	svc service.ConsentService
}

// NewConsentHandler creates a new ConsentHandler.
func NewConsentHandler(svc service.ConsentService) *ConsentHandler {
	return &ConsentHandler{svc: svc}
}

// ListConsents handles GET /api/v1/patients/{id}/consents
func (h *ConsentHandler) ListConsents(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "patient ID required", nil)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	resp, err := h.svc.ListConsentsForPatient(r.Context(), patientID, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrInternal, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

type grantConsentRequest struct {
	PerformerID string                 `json:"performer_id"`
	Scope       string                 `json:"scope"`
	Period      *service.ConsentPeriod `json:"period,omitempty"`
	Category    string                 `json:"category,omitempty"`
}

// GrantConsent handles POST /api/v1/patients/{id}/consents
func (h *ConsentHandler) GrantConsent(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "patient ID required", nil)
		return
	}

	var req grantConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "invalid request body", nil)
		return
	}

	if req.PerformerID == "" {
		model.WriteError(w, model.ErrValidation, "performer_id required", nil)
		return
	}
	if req.Scope == "" {
		req.Scope = "patient-privacy"
	}

	resp, err := h.svc.GrantConsent(r.Context(), patientID, req.PerformerID, req.Scope, req.Period, req.Category)
	if err != nil {
		model.WriteError(w, model.ErrInternal, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusCreated, resp)
}

// RevokeConsent handles DELETE /api/v1/consents/{consentId}
func (h *ConsentHandler) RevokeConsent(w http.ResponseWriter, r *http.Request) {
	consentID := chi.URLParam(r, "consentId")
	if consentID == "" {
		model.WriteError(w, model.ErrValidation, "consent ID required", nil)
		return
	}

	if err := h.svc.RevokeConsent(r.Context(), consentID); err != nil {
		model.WriteError(w, model.ErrInternal, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, map[string]string{"status": "revoked", "consent_id": consentID})
}

// IssueVC handles POST /api/v1/consents/{consentId}/vc
func (h *ConsentHandler) IssueVC(w http.ResponseWriter, r *http.Request) {
	consentID := chi.URLParam(r, "consentId")
	if consentID == "" {
		model.WriteError(w, model.ErrValidation, "consent ID required", nil)
		return
	}

	resp, err := h.svc.IssueConsentVC(r.Context(), consentID)
	if err != nil {
		model.WriteError(w, model.ErrInternal, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}
