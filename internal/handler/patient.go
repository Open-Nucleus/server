package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

type PatientHandler struct {
	svc service.PatientService
}

func NewPatientHandler(svc service.PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

// List handles GET /api/v1/patients
func (h *PatientHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)
	q := r.URL.Query()

	req := &service.ListPatientsRequest{
		Page:          page,
		PerPage:       perPage,
		Sort:          q.Get("sort"),
		Gender:        q.Get("gender"),
		BirthDateFrom: q.Get("birth_date_from"),
		BirthDateTo:   q.Get("birth_date_to"),
		SiteID:        q.Get("site_id"),
		Status:        q.Get("status"),
		HasAlerts:     q.Get("has_alerts") == "true",
	}

	resp, err := h.svc.ListPatients(r.Context(), req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Patients, pg)
}

// GetByID handles GET /api/v1/patients/{id}
func (h *PatientHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "Patient ID is required", nil)
		return
	}

	bundle, err := h.svc.GetPatient(r.Context(), patientID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, bundle)
}

// Search handles GET /api/v1/patients/search
func (h *PatientHandler) Search(w http.ResponseWriter, r *http.Request) {
	page, perPage := model.PaginationFromRequest(r)
	query := r.URL.Query().Get("q")

	resp, err := h.svc.SearchPatients(r.Context(), query, page, perPage)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	pg := model.NewPagination(resp.Page, resp.PerPage, resp.Total)
	model.SuccessWithPagination(w, resp.Patients, pg)
}

// Create handles POST /api/v1/patients
func (h *PatientHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.CreatePatient(r.Context(), body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusCreated, resp)
}

// Update handles PUT /api/v1/patients/{id}
func (h *PatientHandler) Update(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "Patient ID is required", nil)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
		return
	}

	resp, err := h.svc.UpdatePatient(r.Context(), patientID, body)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/patients/{id}
func (h *PatientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "Patient ID is required", nil)
		return
	}

	resp, err := h.svc.DeletePatient(r.Context(), patientID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	writeResponseWithGit(w, http.StatusOK, resp)
}

// History handles GET /api/v1/patients/{id}/history
func (h *PatientHandler) History(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "Patient ID is required", nil)
		return
	}

	resp, err := h.svc.GetPatientHistory(r.Context(), patientID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// Timeline handles GET /api/v1/patients/{id}/timeline
func (h *PatientHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "Patient ID is required", nil)
		return
	}

	resp, err := h.svc.GetPatientTimeline(r.Context(), patientID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// Match handles POST /api/v1/patients/match
func (h *PatientHandler) Match(w http.ResponseWriter, r *http.Request) {
	var req service.MatchPatientsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		model.WriteError(w, model.ErrValidation, "Invalid request body", nil)
		return
	}

	resp, err := h.svc.MatchPatients(r.Context(), &req)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// Erase handles DELETE /api/v1/patients/{id}/erase — crypto-erasure for privacy compliance.
func (h *PatientHandler) Erase(w http.ResponseWriter, r *http.Request) {
	patientID := chi.URLParam(r, "id")
	if patientID == "" {
		model.WriteError(w, model.ErrValidation, "Patient ID is required", nil)
		return
	}

	resp, err := h.svc.ErasePatient(r.Context(), patientID)
	if err != nil {
		model.WriteError(w, model.ErrServiceUnavailable, err.Error(), nil)
		return
	}

	model.Success(w, http.StatusOK, resp)
}

// writeResponseWithGit writes a write response including git metadata in the envelope.
func writeResponseWithGit(w http.ResponseWriter, status int, resp *service.WriteResponse) {
	env := model.Envelope{Status: "success", Data: resp.Resource}
	if resp.Git != nil {
		env.Git = &model.GitInfo{
			Commit:  resp.Git.Commit,
			Message: resp.Git.Message,
		}
	}
	model.JSON(w, status, env)
}
