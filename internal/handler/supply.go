package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type SupplyHandler struct {
	svc service.SupplyService
}

func NewSupplyHandler(svc service.SupplyService) *SupplyHandler {
	return &SupplyHandler{svc: svc}
}

// Inventory handles GET /api/v1/supply/inventory
func (h *SupplyHandler) Inventory(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.GetInventory(r.Context(), page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Items, pg)
}

// InventoryItem handles GET /api/v1/supply/inventory/{item_code}
func (h *SupplyHandler) InventoryItem(w http.ResponseWriter, r *http.Request) {
	itemCode := chi.URLParam(r, "item_code")

	resp, err := h.svc.GetInventoryItem(r.Context(), itemCode)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// RecordDelivery handles POST /api/v1/supply/deliveries
func (h *SupplyHandler) RecordDelivery(w http.ResponseWriter, r *http.Request) {
	var req service.RecordDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.RecordDelivery(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusCreated, resp)
}

// Predictions handles GET /api/v1/supply/predictions
func (h *SupplyHandler) Predictions(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetPredictions(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// Redistribution handles GET /api/v1/supply/redistribution
func (h *SupplyHandler) Redistribution(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetRedistribution(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
