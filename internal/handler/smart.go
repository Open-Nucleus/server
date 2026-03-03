package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/pkg/smart"
)

// SmartHandler handles SMART on FHIR OAuth2 and client management endpoints.
type SmartHandler struct {
	svc     service.SmartService
	baseURL string
}

// NewSmartHandler creates a new SMART handler.
func NewSmartHandler(svc service.SmartService, baseURL string) *SmartHandler {
	return &SmartHandler{svc: svc, baseURL: baseURL}
}

// SmartConfiguration handles GET /.well-known/smart-configuration
func (h *SmartHandler) SmartConfiguration(w http.ResponseWriter, r *http.Request) {
	cfg := smart.GenerateSmartConfiguration(h.baseURL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

// Authorize handles GET /auth/smart/authorize
func (h *SmartHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	claims := model.ClaimsFromContext(r.Context())
	if claims == nil {
		model.WriteError(w, model.ErrAuthRequired, "JWT required for authorize", nil)
		return
	}

	req := &service.AuthorizeRequest{
		ClientID:            q.Get("client_id"),
		RedirectURI:         q.Get("redirect_uri"),
		Scope:               q.Get("scope"),
		State:               q.Get("state"),
		CodeChallenge:       q.Get("code_challenge"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		Launch:              q.Get("launch"),
	}

	resp, err := h.svc.Authorize(r.Context(), req)
	if err != nil {
		model.WriteError(w, model.ErrValidation, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// Token handles POST /auth/smart/token
func (h *SmartHandler) Token(w http.ResponseWriter, r *http.Request) {
	// Token endpoint accepts form-encoded or JSON.
	contentType := r.Header.Get("Content-Type")
	var req service.ExchangeTokenRequest

	if contentType == "application/x-www-form-urlencoded" || contentType == "" {
		if err := r.ParseForm(); err != nil {
			model.WriteError(w, model.ErrValidation, "Invalid form data", nil)
			return
		}
		req = service.ExchangeTokenRequest{
			GrantType:    r.FormValue("grant_type"),
			Code:         r.FormValue("code"),
			RedirectURI:  r.FormValue("redirect_uri"),
			CodeVerifier: r.FormValue("code_verifier"),
			ClientID:     r.FormValue("client_id"),
			ClientSecret: r.FormValue("client_secret"),
		}
	} else {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
		if err != nil {
			model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			model.WriteError(w, model.ErrValidation, "Invalid JSON", nil)
			return
		}
	}

	// Check for client_secret_basic auth in Authorization header.
	if clientID, clientSecret, ok := r.BasicAuth(); ok {
		if req.ClientID == "" {
			req.ClientID = clientID
		}
		if req.ClientSecret == "" {
			req.ClientSecret = clientSecret
		}
	}

	resp, err := h.svc.ExchangeToken(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrValidation, err.Error(), nil)
		return
	}

	// Token responses are raw OAuth2 JSON (not wrapped in envelope).
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Revoke handles POST /auth/smart/revoke
func (h *SmartHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	if err := h.svc.RevokeToken(r.Context(), req.Token); err != nil {
		model.WriteError(w, model.ErrValidation, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Introspect handles POST /auth/smart/introspect
func (h *SmartHandler) Introspect(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.IntrospectToken(r.Context(), req.Token)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Register handles POST /auth/smart/register
func (h *SmartHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.RegisterClient(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrValidation, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusCreated, resp)
}

// Launch handles POST /auth/smart/launch
func (h *SmartHandler) Launch(w http.ResponseWriter, r *http.Request) {
	var req service.CreateLaunchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.CreateLaunch(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrValidation, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusCreated, resp)
}

// ListClients handles GET /api/v1/smart/clients
func (h *SmartHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListClients(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// GetClient handles GET /api/v1/smart/clients/{id}
func (h *SmartHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")

	resp, err := h.svc.GetClient(r.Context(), clientID)
	if err != nil {
		model.WriteError(w, model.ErrResourceNotFound, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// UpdateClient handles PUT /api/v1/smart/clients/{id}
func (h *SmartHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")

	var req service.UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.UpdateClient(r.Context(), clientID, &req)
	if err != nil {
		model.WriteError(w, model.ErrValidation, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// DeleteClient handles DELETE /api/v1/smart/clients/{id}
func (h *SmartHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "id")

	if err := h.svc.DeleteClient(r.Context(), clientID); err != nil {
		model.WriteError(w, model.ErrResourceNotFound, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, map[string]string{"message": "client deleted"})
}
