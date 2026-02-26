package handler

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

// --- Encounters ---

// ListEncounters handles GET /api/v1/patients/{id}/encounters
func (h *PatientHandler) ListEncounters(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListEncounters(r.Context(), patientID, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Resources, pg)
}

// GetEncounter handles GET /api/v1/patients/{id}/encounters/{eid}
func (h *PatientHandler) GetEncounter(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	encounterID := chi.URLParam(r, "eid")

	resp, err := h.svc.GetEncounter(r.Context(), patientID, encounterID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// CreateEncounter handles POST /api/v1/patients/{id}/encounters
func (h *PatientHandler) CreateEncounter(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.CreateEncounter(r.Context(), patientID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusCreated, resp)
}

// UpdateEncounter handles PUT /api/v1/patients/{id}/encounters/{eid}
func (h *PatientHandler) UpdateEncounter(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	encounterID := chi.URLParam(r, "eid")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.UpdateEncounter(r.Context(), patientID, encounterID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusOK, resp)
}

// --- Observations ---

// ListObservations handles GET /api/v1/patients/{id}/observations
func (h *PatientHandler) ListObservations(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	page, perPage := model.PaginationFromRequest(r)
	q := r.URL.Query()

	filters := service.ObservationFilters{
		Code:        q.Get("code"),
		Category:    q.Get("category"),
		DateFrom:    q.Get("date_from"),
		DateTo:      q.Get("date_to"),
		EncounterID: q.Get("encounter_id"),
	}

	resp, err := h.svc.ListObservations(r.Context(), patientID, filters, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Resources, pg)
}

// GetObservation handles GET /api/v1/patients/{id}/observations/{oid}
func (h *PatientHandler) GetObservation(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	observationID := chi.URLParam(r, "oid")

	resp, err := h.svc.GetObservation(r.Context(), patientID, observationID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// CreateObservation handles POST /api/v1/patients/{id}/observations
func (h *PatientHandler) CreateObservation(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.CreateObservation(r.Context(), patientID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusCreated, resp)
}

// --- Conditions ---

// ListConditions handles GET /api/v1/patients/{id}/conditions
func (h *PatientHandler) ListConditions(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	page, perPage := model.PaginationFromRequest(r)
	q := r.URL.Query()

	filters := service.ConditionFilters{
		ClinicalStatus: q.Get("clinical_status"),
		Category:       q.Get("category"),
		Code:           q.Get("code"),
	}

	resp, err := h.svc.ListConditions(r.Context(), patientID, filters, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Resources, pg)
}

// CreateCondition handles POST /api/v1/patients/{id}/conditions
func (h *PatientHandler) CreateCondition(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.CreateCondition(r.Context(), patientID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusCreated, resp)
}

// UpdateCondition handles PUT /api/v1/patients/{id}/conditions/{cid}
func (h *PatientHandler) UpdateCondition(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	conditionID := chi.URLParam(r, "cid")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.UpdateCondition(r.Context(), patientID, conditionID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusOK, resp)
}

// --- Medication Requests ---

// ListMedicationRequests handles GET /api/v1/patients/{id}/medication-requests
func (h *PatientHandler) ListMedicationRequests(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListMedicationRequests(r.Context(), patientID, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Resources, pg)
}

// CreateMedicationRequest handles POST /api/v1/patients/{id}/medication-requests
func (h *PatientHandler) CreateMedicationRequest(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.CreateMedicationRequest(r.Context(), patientID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusCreated, resp)
}

// UpdateMedicationRequest handles PUT /api/v1/patients/{id}/medication-requests/{mid}
func (h *PatientHandler) UpdateMedicationRequest(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	medicationRequestID := chi.URLParam(r, "mid")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.UpdateMedicationRequest(r.Context(), patientID, medicationRequestID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusOK, resp)
}

// --- Allergy Intolerances ---

// ListAllergyIntolerances handles GET /api/v1/patients/{id}/allergy-intolerances
func (h *PatientHandler) ListAllergyIntolerances(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	page, perPage := model.PaginationFromRequest(r)

	resp, err := h.svc.ListAllergyIntolerances(r.Context(), patientID, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Resources, pg)
}

// CreateAllergyIntolerance handles POST /api/v1/patients/{id}/allergy-intolerances
func (h *PatientHandler) CreateAllergyIntolerance(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.CreateAllergyIntolerance(r.Context(), patientID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusCreated, resp)
}

// UpdateAllergyIntolerance handles PUT /api/v1/patients/{id}/allergy-intolerances/{aid}
func (h *PatientHandler) UpdateAllergyIntolerance(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	allergyIntoleranceID := chi.URLParam(r, "aid")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.UpdateAllergyIntolerance(r.Context(), patientID, allergyIntoleranceID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusOK, resp)
}
