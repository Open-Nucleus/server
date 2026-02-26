package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPatientService struct {
	service.PatientService // embed to satisfy interface with zero methods
	listFn                 func(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error)
	getFn                  func(ctx context.Context, id string) (*service.PatientBundle, error)
	searchFn               func(ctx context.Context, query string, page, perPage int) (*service.ListPatientsResponse, error)
	createFn               func(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error)
	updateFn               func(ctx context.Context, id string, body json.RawMessage) (*service.WriteResponse, error)
	deleteFn               func(ctx context.Context, id string) (*service.WriteResponse, error)
	matchFn                func(ctx context.Context, req *service.MatchPatientsRequest) (*service.MatchPatientsResponse, error)
	historyFn              func(ctx context.Context, id string) (*service.PatientHistoryResponse, error)
	timelineFn             func(ctx context.Context, id string) (*service.PatientTimelineResponse, error)
}

func (m *mockPatientService) ListPatients(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
	if m.listFn != nil {
		return m.listFn(ctx, req)
	}
	return nil, nil
}

func (m *mockPatientService) GetPatient(ctx context.Context, id string) (*service.PatientBundle, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPatientService) SearchPatients(ctx context.Context, query string, page, perPage int) (*service.ListPatientsResponse, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, query, page, perPage)
	}
	return nil, nil
}

func (m *mockPatientService) CreatePatient(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error) {
	if m.createFn != nil {
		return m.createFn(ctx, body)
	}
	return nil, nil
}

func (m *mockPatientService) UpdatePatient(ctx context.Context, id string, body json.RawMessage) (*service.WriteResponse, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, body)
	}
	return nil, nil
}

func (m *mockPatientService) DeletePatient(ctx context.Context, id string) (*service.WriteResponse, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPatientService) MatchPatients(ctx context.Context, req *service.MatchPatientsRequest) (*service.MatchPatientsResponse, error) {
	if m.matchFn != nil {
		return m.matchFn(ctx, req)
	}
	return nil, nil
}

func (m *mockPatientService) GetPatientHistory(ctx context.Context, id string) (*service.PatientHistoryResponse, error) {
	if m.historyFn != nil {
		return m.historyFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPatientService) GetPatientTimeline(ctx context.Context, id string) (*service.PatientTimelineResponse, error) {
	if m.timelineFn != nil {
		return m.timelineFn(ctx, id)
	}
	return nil, nil
}

func TestPatientHandler_List_Success(t *testing.T) {
	svc := &mockPatientService{
		listFn: func(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
			assert.Equal(t, 1, req.Page)
			assert.Equal(t, 25, req.PerPage)
			return &service.ListPatientsResponse{
				Patients:   []any{map[string]string{"id": "patient-001"}},
				Page:       1,
				PerPage:    25,
				Total:      1,
				TotalPages: 1,
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients?page=1&per_page=25", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
	require.NotNil(t, env.Pagination)
	assert.Equal(t, 1, env.Pagination.Total)
}

func TestPatientHandler_List_ServiceError(t *testing.T) {
	svc := &mockPatientService{
		listFn: func(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
			return nil, fmt.Errorf("patient service unavailable")
		},
	}

	h := handler.NewPatientHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestPatientHandler_GetByID_Success(t *testing.T) {
	svc := &mockPatientService{
		getFn: func(ctx context.Context, id string) (*service.PatientBundle, error) {
			assert.Equal(t, "patient-001", id)
			return &service.PatientBundle{
				Patient:    map[string]string{"id": "patient-001", "resourceType": "Patient"},
				Encounters: []any{},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	// Use chi's URL params
	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
}
