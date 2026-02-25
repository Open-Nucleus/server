package handler

import (
	"encoding/json"
	"net/http"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.Logout(r.Context(), req.Token); err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Whoami handles GET /api/v1/auth/whoami
func (h *AuthHandler) Whoami(w http.ResponseWriter, r *http.Request) {
	// First try to return from JWT claims in context (no gRPC call needed)
	claims := model.ClaimsFromContext(r.Context())
	if claims != nil {
		model.Success(w, http.StatusOK, map[string]any{
			"subject": claims.Subject,
			"node_id": claims.Node,
			"site_id": claims.Site,
			"role": map[string]any{
				"code":        claims.Role,
				"permissions": claims.Permissions,
			},
		})
		return
	}

	resp, err := h.svc.Whoami(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}
