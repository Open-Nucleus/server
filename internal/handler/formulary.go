package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	category := r.URL.Query().Get("category")

	resp, err := h.svc.SearchMedications(r.Context(), query, category, page, perPage)
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

// ListMedicationsByCategory handles GET /api/v1/formulary/medications/category/{category}
func (h *FormularyHandler) ListMedicationsByCategory(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListMedicationsByCategory(r.Context(), category, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Medications, pg)
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

// CheckAllergyConflicts handles POST /api/v1/formulary/check-allergies
func (h *FormularyHandler) CheckAllergyConflicts(w http.ResponseWriter, r *http.Request) {
	var req service.CheckAllergyConflictsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.CheckAllergyConflicts(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// ValidateDosing handles POST /api/v1/formulary/dosing/validate
func (h *FormularyHandler) ValidateDosing(w http.ResponseWriter, r *http.Request) {
	var req service.ValidateDosingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.ValidateDosing(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GetDosingOptions handles GET /api/v1/formulary/dosing/options
func (h *FormularyHandler) GetDosingOptions(w http.ResponseWriter, r *http.Request) {
	medicationCode := r.URL.Query().Get("medication_code")
	weightStr := r.URL.Query().Get("patient_weight_kg")
	var weight float64
	if weightStr != "" {
		var err error
		weight, err = strconv.ParseFloat(weightStr, 64)
		if err != nil {
			model.WriteError(w, model.ErrValidation, "Invalid patient_weight_kg", nil)
			return
		}
	}

	resp, err := h.svc.GetDosingOptions(r.Context(), medicationCode, weight)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GenerateSchedule handles POST /api/v1/formulary/dosing/schedule
func (h *FormularyHandler) GenerateSchedule(w http.ResponseWriter, r *http.Request) {
	var req service.GenerateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.GenerateSchedule(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GetStockLevel handles GET /api/v1/formulary/stock/{site_id}/{medication_code}
func (h *FormularyHandler) GetStockLevel(w http.ResponseWriter, r *http.Request) {
	siteID := chi.URLParam(r, "site_id")
	medicationCode := chi.URLParam(r, "medication_code")

	resp, err := h.svc.GetStockLevel(r.Context(), siteID, medicationCode)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// UpdateStockLevel handles PUT /api/v1/formulary/stock/{site_id}/{medication_code}
func (h *FormularyHandler) UpdateStockLevel(w http.ResponseWriter, r *http.Request) {
	siteID := chi.URLParam(r, "site_id")
	medicationCode := chi.URLParam(r, "medication_code")

	var req service.UpdateStockLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}
	req.SiteID = siteID
	req.MedicationCode = medicationCode

	resp, err := h.svc.UpdateStockLevel(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// RecordDelivery handles POST /api/v1/formulary/deliveries
func (h *FormularyHandler) RecordDelivery(w http.ResponseWriter, r *http.Request) {
	var req service.FormularyDeliveryRequest
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

// GetStockPrediction handles GET /api/v1/formulary/stock/{site_id}/{medication_code}/prediction
func (h *FormularyHandler) GetStockPrediction(w http.ResponseWriter, r *http.Request) {
	siteID := chi.URLParam(r, "site_id")
	medicationCode := chi.URLParam(r, "medication_code")

	resp, err := h.svc.GetStockPrediction(r.Context(), siteID, medicationCode)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GetRedistributionSuggestions handles GET /api/v1/formulary/redistribution/{medication_code}
func (h *FormularyHandler) GetRedistributionSuggestions(w http.ResponseWriter, r *http.Request) {
	medicationCode := chi.URLParam(r, "medication_code")

	resp, err := h.svc.GetRedistributionSuggestions(r.Context(), medicationCode)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}

// GetFormularyInfo handles GET /api/v1/formulary/info
func (h *FormularyHandler) GetFormularyInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.GetFormularyInfo(r.Context())
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}
	model.Success(w, http.StatusOK, resp)
}
