package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClinicalService embeds mockPatientService and overrides clinical methods.
type mockClinicalService struct {
	mockPatientService
	listEncountersFn    func(ctx context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error)
	createEncounterFn   func(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error)
	listObservationsFn  func(ctx context.Context, patientID string, filters service.ObservationFilters, page, perPage int) (*service.ClinicalListResponse, error)
	createObservationFn func(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error)
}

func (m *mockClinicalService) ListEncounters(ctx context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
	if m.listEncountersFn != nil {
		return m.listEncountersFn(ctx, patientID, page, perPage)
	}
	return nil, nil
}

func (m *mockClinicalService) CreateEncounter(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	if m.createEncounterFn != nil {
		return m.createEncounterFn(ctx, patientID, body)
	}
	return nil, nil
}

func (m *mockClinicalService) ListObservations(ctx context.Context, patientID string, filters service.ObservationFilters, page, perPage int) (*service.ClinicalListResponse, error) {
	if m.listObservationsFn != nil {
		return m.listObservationsFn(ctx, patientID, filters, page, perPage)
	}
	return nil, nil
}

func (m *mockClinicalService) CreateObservation(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
	if m.createObservationFn != nil {
		return m.createObservationFn(ctx, patientID, body)
	}
	return nil, nil
}

func TestClinicalHandler_ListEncounters_Success(t *testing.T) {
	svc := &mockClinicalService{
		listEncountersFn: func(ctx context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
			assert.Equal(t, "patient-001", patientID)
			return &service.ClinicalListResponse{
				Resources:  []any{map[string]string{"resourceType": "Encounter", "id": "enc-001"}},
				Page:       1,
				PerPage:    25,
				Total:      1,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}/encounters", h.ListEncounters)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001/encounters", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
	require.NotNil(t, env.Pagination)
	assert.Equal(t, 1, env.Pagination.Total)
}

func TestClinicalHandler_ListEncounters_ServiceError(t *testing.T) {
	svc := &mockClinicalService{
		listEncountersFn: func(ctx context.Context, patientID string, page, perPage int) (*service.ClinicalListResponse, error) {
			return nil, fmt.Errorf("patient service unavailable: backend not connected")
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}/encounters", h.ListEncounters)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001/encounters", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestClinicalHandler_CreateEncounter_Success(t *testing.T) {
	svc := &mockClinicalService{
		createEncounterFn: func(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
			assert.Equal(t, "patient-001", patientID)
			return &service.WriteResponse{
				Resource: map[string]string{"resourceType": "Encounter", "id": "enc-new"},
				Git:      &service.GitMeta{Commit: "abc123", Message: "Created encounter"},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/patients/{id}/encounters", h.CreateEncounter)

	body := `{"resourceType":"Encounter","status":"in-progress","class":{"code":"AMB"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients/patient-001/encounters", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
	require.NotNil(t, env.Git)
	assert.Equal(t, "abc123", env.Git.Commit)
}

func TestClinicalHandler_ListObservations_WithFilters(t *testing.T) {
	svc := &mockClinicalService{
		listObservationsFn: func(ctx context.Context, patientID string, filters service.ObservationFilters, page, perPage int) (*service.ClinicalListResponse, error) {
			assert.Equal(t, "patient-001", patientID)
			assert.Equal(t, "vital-signs", filters.Category)
			assert.Equal(t, "85354-9", filters.Code)
			return &service.ClinicalListResponse{
				Resources:  []any{},
				Page:       1,
				PerPage:    25,
				Total:      0,
				TotalPages: 0,
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}/observations", h.ListObservations)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001/observations?category=vital-signs&code=85354-9", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestClinicalHandler_CreateObservation_Success(t *testing.T) {
	svc := &mockClinicalService{
		createObservationFn: func(ctx context.Context, patientID string, body json.RawMessage) (*service.WriteResponse, error) {
			assert.Equal(t, "patient-001", patientID)
			return &service.WriteResponse{
				Resource: map[string]string{"resourceType": "Observation", "id": "obs-new"},
				Git:      &service.GitMeta{Commit: "xyz789", Message: "Created observation"},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/v1/patients/{id}/observations", h.CreateObservation)

	body := `{"resourceType":"Observation","status":"final","code":{"coding":[{"code":"85354-9"}]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients/patient-001/observations", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
}
