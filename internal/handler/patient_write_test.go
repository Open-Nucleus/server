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

func TestPatientHandler_Create_Success(t *testing.T) {
	svc := &mockPatientService{
		createFn: func(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error) {
			return &service.WriteResponse{
				Resource: map[string]string{"id": "patient-new", "resourceType": "Patient"},
				Git:      &service.GitMeta{Commit: "abc123", Message: "Created patient"},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)
	body := `{"resourceType":"Patient","name":[{"family":"Okafor"}],"gender":"male"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
	require.NotNil(t, env.Git)
	assert.Equal(t, "abc123", env.Git.Commit)
}

func TestPatientHandler_Create_ServiceError(t *testing.T) {
	svc := &mockPatientService{
		createFn: func(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error) {
			return nil, fmt.Errorf("patient service unavailable: backend not connected")
		},
	}

	h := handler.NewPatientHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}

func TestPatientHandler_Update_Success(t *testing.T) {
	svc := &mockPatientService{
		updateFn: func(ctx context.Context, id string, body json.RawMessage) (*service.WriteResponse, error) {
			assert.Equal(t, "patient-001", id)
			return &service.WriteResponse{
				Resource: map[string]string{"id": "patient-001", "resourceType": "Patient"},
				Git:      &service.GitMeta{Commit: "def456", Message: "Updated patient"},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Put("/api/v1/patients/{id}", h.Update)

	body := `{"resourceType":"Patient","name":[{"family":"Updated"}],"gender":"female"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/patients/patient-001", strings.NewReader(body))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
	require.NotNil(t, env.Git)
	assert.Equal(t, "def456", env.Git.Commit)
}

func TestPatientHandler_Delete_Success(t *testing.T) {
	svc := &mockPatientService{
		deleteFn: func(ctx context.Context, id string) (*service.WriteResponse, error) {
			assert.Equal(t, "patient-001", id)
			return &service.WriteResponse{
				Git: &service.GitMeta{Commit: "ghi789", Message: "Deleted patient"},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Delete("/api/v1/patients/{id}", h.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/patients/patient-001", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "success", env.Status)
	require.NotNil(t, env.Git)
}

func TestPatientHandler_Match_Success(t *testing.T) {
	svc := &mockPatientService{
		matchFn: func(ctx context.Context, req *service.MatchPatientsRequest) (*service.MatchPatientsResponse, error) {
			assert.Equal(t, "Okafor", req.FamilyName)
			return &service.MatchPatientsResponse{
				Matches: []service.PatientMatch{
					{PatientID: "patient-001", Confidence: 0.95, MatchFactors: []string{"name", "gender"}},
				},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)
	body := `{"family_name":"Okafor","given_names":["Chukwu"],"gender":"male"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/patients/match", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.Match(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPatientHandler_History_Success(t *testing.T) {
	svc := &mockPatientService{
		historyFn: func(ctx context.Context, id string) (*service.PatientHistoryResponse, error) {
			assert.Equal(t, "patient-001", id)
			return &service.PatientHistoryResponse{
				Entries: []service.HistoryEntry{
					{CommitHash: "abc123", Operation: "create", ResourceType: "Patient"},
				},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}/history", h.History)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001/history", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestPatientHandler_Timeline_Success(t *testing.T) {
	svc := &mockPatientService{
		timelineFn: func(ctx context.Context, id string) (*service.PatientTimelineResponse, error) {
			assert.Equal(t, "patient-001", id)
			return &service.PatientTimelineResponse{
				Events: []any{map[string]string{"type": "encounter"}},
			}, nil
		},
	}

	h := handler.NewPatientHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}/timeline", h.Timeline)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/patients/patient-001/timeline", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
