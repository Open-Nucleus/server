package handler

import (
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
