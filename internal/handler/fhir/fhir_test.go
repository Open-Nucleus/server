package fhir

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/FibrinLab/open-nucleus/internal/service"
	pkgfhir "github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// --- Mock PatientService ---

type mockPatientService struct {
	service.PatientService // embed to satisfy interface (panics on unimplemented methods)
	getResourceFn          func(ctx context.Context, resourceType, resourceID string) (any, error)
	listPatientsFn         func(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error)
	createPatientFn        func(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error)
	updatePatientFn        func(ctx context.Context, id string, body json.RawMessage) (*service.WriteResponse, error)
	deletePatientFn        func(ctx context.Context, id string) (*service.WriteResponse, error)
	getPatientFn           func(ctx context.Context, id string) (*service.PatientBundle, error)
}

func (m *mockPatientService) GetResource(ctx context.Context, resourceType, resourceID string) (any, error) {
	if m.getResourceFn != nil {
		return m.getResourceFn(ctx, resourceType, resourceID)
	}
	return nil, nil
}

func (m *mockPatientService) ListPatients(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
	if m.listPatientsFn != nil {
		return m.listPatientsFn(ctx, req)
	}
	return &service.ListPatientsResponse{Page: 1, PerPage: 25, Total: 0}, nil
}

func (m *mockPatientService) CreatePatient(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error) {
	if m.createPatientFn != nil {
		return m.createPatientFn(ctx, body)
	}
	return nil, nil
}

func (m *mockPatientService) UpdatePatient(ctx context.Context, id string, body json.RawMessage) (*service.WriteResponse, error) {
	if m.updatePatientFn != nil {
		return m.updatePatientFn(ctx, id, body)
	}
	return nil, nil
}

func (m *mockPatientService) DeletePatient(ctx context.Context, id string) (*service.WriteResponse, error) {
	if m.deletePatientFn != nil {
		return m.deletePatientFn(ctx, id)
	}
	return &service.WriteResponse{}, nil
}

func (m *mockPatientService) GetPatient(ctx context.Context, id string) (*service.PatientBundle, error) {
	if m.getPatientFn != nil {
		return m.getPatientFn(ctx, id)
	}
	return &service.PatientBundle{}, nil
}

// --- Content Negotiation Tests ---

func TestContentNegotiation_AcceptJSON(t *testing.T) {
	handler := ContentNegotiation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name   string
		accept string
		status int
	}{
		{"fhir+json", "application/fhir+json", http.StatusOK},
		{"json", "application/json", http.StatusOK},
		{"wildcard", "*/*", http.StatusOK},
		{"empty", "", http.StatusOK},
		{"xml rejected", "application/fhir+xml", http.StatusNotAcceptable},
		{"xml with json fallback", "application/fhir+json, application/xml", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.status, rr.Code)
		})
	}
}

func TestContentNegotiation_XMLReturnsOperationOutcome(t *testing.T) {
	handler := ContentNegotiation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/fhir+xml")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotAcceptable, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/fhir+json")

	var outcome map[string]any
	err := json.NewDecoder(rr.Body).Decode(&outcome)
	require.NoError(t, err)
	assert.Equal(t, "OperationOutcome", outcome["resourceType"])
}

// --- Search Param Parsing Tests ---

func TestParseSearchParams(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
		params := ParseFHIRSearchParams(req)
		assert.Equal(t, 25, params.Count)
		assert.Equal(t, 0, params.Offset)
		assert.Equal(t, 1, params.Page)
		assert.Empty(t, params.Patient)
	})

	t.Run("custom count and offset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?_count=10&_offset=20", nil)
		params := ParseFHIRSearchParams(req)
		assert.Equal(t, 10, params.Count)
		assert.Equal(t, 20, params.Offset)
		assert.Equal(t, 3, params.Page) // (20/10)+1
	})

	t.Run("patient reference with prefix", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Encounter?patient=Patient/abc-123", nil)
		params := ParseFHIRSearchParams(req)
		assert.Equal(t, "abc-123", params.Patient)
	})

	t.Run("subject reference", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?subject=Patient/xyz", nil)
		params := ParseFHIRSearchParams(req)
		assert.Equal(t, "xyz", params.Patient)
	})

	t.Run("max count clamped", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=500", nil)
		params := ParseFHIRSearchParams(req)
		assert.Equal(t, 100, params.Count)
	})

	t.Run("filters collected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/fhir/Observation?code=8310-5&status=final", nil)
		params := ParseFHIRSearchParams(req)
		assert.Equal(t, "8310-5", params.Filters["code"])
		assert.Equal(t, "final", params.Filters["status"])
	})
}

// --- Read Handler Tests ---

func TestReadHandler_Success(t *testing.T) {
	svc := &mockPatientService{
		getResourceFn: func(ctx context.Context, rt, id string) (any, error) {
			return map[string]any{
				"resourceType": "Patient",
				"id":           id,
				"meta": map[string]any{
					"versionId":   "1",
					"lastUpdated": "2025-01-01T00:00:00Z",
				},
			}, nil
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Read("Patient")

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/test-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/fhir+json")
	assert.Equal(t, `W/"1"`, rr.Header().Get("ETag"))
	assert.Equal(t, "2025-01-01T00:00:00Z", rr.Header().Get("Last-Modified"))

	var resource map[string]any
	err := json.NewDecoder(rr.Body).Decode(&resource)
	require.NoError(t, err)
	assert.Equal(t, "Patient", resource["resourceType"])
	assert.Equal(t, "test-123", resource["id"])
}

func TestReadHandler_NotFound(t *testing.T) {
	svc := &mockPatientService{
		getResourceFn: func(ctx context.Context, rt, id string) (any, error) {
			return nil, fmt.Errorf("not found")
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Read("Patient")

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/missing", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "missing")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	var outcome map[string]any
	err := json.NewDecoder(rr.Body).Decode(&outcome)
	require.NoError(t, err)
	assert.Equal(t, "OperationOutcome", outcome["resourceType"])
}

func TestReadHandler_Conditional304(t *testing.T) {
	svc := &mockPatientService{
		getResourceFn: func(ctx context.Context, rt, id string) (any, error) {
			return map[string]any{
				"resourceType": "Patient",
				"id":           id,
				"meta": map[string]any{
					"versionId": "v42",
				},
			}, nil
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Read("Patient")

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/test-123", nil)
	req.Header.Set("If-None-Match", `W/"v42"`)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotModified, rr.Code)
}

// --- Search Handler Tests ---

func TestSearchHandler_BundleStructure(t *testing.T) {
	svc := &mockPatientService{
		listPatientsFn: func(ctx context.Context, req *service.ListPatientsRequest) (*service.ListPatientsResponse, error) {
			return &service.ListPatientsResponse{
				Patients: []any{
					map[string]any{"resourceType": "Patient", "id": "p1"},
					map[string]any{"resourceType": "Patient", "id": "p2"},
				},
				Page:       1,
				PerPage:    25,
				Total:      2,
				TotalPages: 1,
			}, nil
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Search("Patient")

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var bundle map[string]any
	err := json.NewDecoder(rr.Body).Decode(&bundle)
	require.NoError(t, err)
	assert.Equal(t, "Bundle", bundle["resourceType"])
	assert.Equal(t, "searchset", bundle["type"])
	assert.Equal(t, float64(2), bundle["total"])

	entries := bundle["entry"].([]any)
	assert.Len(t, entries, 2)

	// Check first entry structure
	entry0 := entries[0].(map[string]any)
	assert.Equal(t, "/fhir/Patient/p1", entry0["fullUrl"])
	assert.Equal(t, map[string]any{"mode": "match"}, entry0["search"])
}

func TestSearchHandler_PatientScopedRequiresPatient(t *testing.T) {
	svc := &mockPatientService{}
	h := NewFHIRHandler(svc)

	// Encounter is patient-scoped — requires patient param
	handler := h.Search("Encounter")

	req := httptest.NewRequest(http.MethodGet, "/fhir/Encounter", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var outcome map[string]any
	err := json.NewDecoder(rr.Body).Decode(&outcome)
	require.NoError(t, err)
	assert.Equal(t, "OperationOutcome", outcome["resourceType"])
}

// --- Create Handler Tests ---

func TestCreateHandler_Success(t *testing.T) {
	svc := &mockPatientService{
		createPatientFn: func(ctx context.Context, body json.RawMessage) (*service.WriteResponse, error) {
			return &service.WriteResponse{
				Resource: map[string]any{
					"resourceType": "Patient",
					"id":           "new-123",
				},
				Git: &service.GitMeta{Commit: "abc", Message: "CREATE Patient/new-123"},
			}, nil
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Create("Patient")

	body := `{"resourceType":"Patient","name":[{"family":"Test"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	assert.Equal(t, "/fhir/Patient/new-123", rr.Header().Get("Location"))
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/fhir+json")
}

func TestCreateHandler_ResourceTypeMismatch(t *testing.T) {
	svc := &mockPatientService{}
	h := NewFHIRHandler(svc)
	handler := h.Create("Patient")

	body := `{"resourceType":"Encounter","status":"planned"}`
	req := httptest.NewRequest(http.MethodPost, "/fhir/Patient", strings.NewReader(body))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	var outcome map[string]any
	err := json.NewDecoder(rr.Body).Decode(&outcome)
	require.NoError(t, err)
	assert.Equal(t, "OperationOutcome", outcome["resourceType"])
}

// --- Update Handler Tests ---

func TestUpdateHandler_Success(t *testing.T) {
	svc := &mockPatientService{
		updatePatientFn: func(ctx context.Context, id string, body json.RawMessage) (*service.WriteResponse, error) {
			return &service.WriteResponse{
				Resource: map[string]any{
					"resourceType": "Patient",
					"id":           id,
				},
			}, nil
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Update("Patient")

	body := `{"resourceType":"Patient","id":"test-456","name":[{"family":"Updated"}]}`
	req := httptest.NewRequest(http.MethodPut, "/fhir/Patient/test-456", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-456")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateHandler_IDMismatch(t *testing.T) {
	svc := &mockPatientService{}
	h := NewFHIRHandler(svc)
	handler := h.Update("Patient")

	body := `{"resourceType":"Patient","id":"different-id"}`
	req := httptest.NewRequest(http.MethodPut, "/fhir/Patient/test-456", strings.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-456")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// --- Delete Handler Tests ---

func TestDeleteHandler_Success(t *testing.T) {
	svc := &mockPatientService{
		deletePatientFn: func(ctx context.Context, id string) (*service.WriteResponse, error) {
			return &service.WriteResponse{}, nil
		},
	}

	h := NewFHIRHandler(svc)
	handler := h.Delete("Patient")

	req := httptest.NewRequest(http.MethodDelete, "/fhir/Patient/test-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

// --- Everything Handler Tests ---

func TestEverythingHandler_BundleStructure(t *testing.T) {
	svc := &mockPatientService{
		getPatientFn: func(ctx context.Context, id string) (*service.PatientBundle, error) {
			return &service.PatientBundle{
				Patient: map[string]any{"resourceType": "Patient", "id": id},
				Encounters: []any{
					map[string]any{"resourceType": "Encounter", "id": "enc-1"},
				},
				Observations: []any{
					map[string]any{"resourceType": "Observation", "id": "obs-1"},
				},
			}, nil
		},
	}

	h := NewFHIRHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1/$everything", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "p1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Everything(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var bundle map[string]any
	err := json.NewDecoder(rr.Body).Decode(&bundle)
	require.NoError(t, err)
	assert.Equal(t, "Bundle", bundle["resourceType"])
	assert.Equal(t, "searchset", bundle["type"])
	assert.Equal(t, float64(3), bundle["total"]) // patient + encounter + observation

	entries := bundle["entry"].([]any)
	assert.Len(t, entries, 3)

	// Patient entry should have search mode "match"
	patientEntry := entries[0].(map[string]any)
	assert.Equal(t, "/fhir/Patient/p1", patientEntry["fullUrl"])
	assert.Equal(t, map[string]any{"mode": "match"}, patientEntry["search"])

	// Related entries should have search mode "include"
	encEntry := entries[1].(map[string]any)
	assert.Equal(t, map[string]any{"mode": "include"}, encEntry["search"])
}

// --- Permission Mapping Tests ---

func TestDispatcherPermissions(t *testing.T) {
	svc := &mockPatientService{}
	h := NewFHIRHandler(svc)

	tests := []struct {
		resourceType string
		readPerm     string
		writePerm    string
	}{
		{"Patient", "patient:read", "patient:write"},
		{"Encounter", "encounter:read", "encounter:write"},
		{"Observation", "observation:read", "observation:write"},
		{"Condition", "condition:read", "condition:write"},
		{"MedicationRequest", "medication:read", "medication:write"},
		{"AllergyIntolerance", "allergy:read", "allergy:write"},
		{"Immunization", "encounter:read", "encounter:write"},
		{"Procedure", "encounter:read", "encounter:write"},
		{"Flag", "alert:read", "alert:write"},
		{"Practitioner", "patient:read", "patient:write"},
		{"Organization", "patient:read", "patient:write"},
		{"Location", "patient:read", "patient:write"},
		{"Provenance", "patient:read", ""},
		{"DetectedIssue", "alert:read", ""},
		{"SupplyDelivery", "supply:read", ""},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			disp := h.dispatchers[tt.resourceType]
			require.NotNil(t, disp, "no dispatcher for %s", tt.resourceType)
			assert.Equal(t, tt.readPerm, disp.ReadPerm)
			assert.Equal(t, tt.writePerm, disp.WritePerm)
		})
	}
}

// --- Response Helper Tests ---

func TestWriteFHIRError_OperationOutcome(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteFHIRError(rr, http.StatusNotFound, "not-found", "Patient xyz not found")

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/fhir+json")

	var outcome map[string]any
	err := json.NewDecoder(rr.Body).Decode(&outcome)
	require.NoError(t, err)
	assert.Equal(t, "OperationOutcome", outcome["resourceType"])

	issues := outcome["issue"].([]any)
	require.Len(t, issues, 1)
	issue := issues[0].(map[string]any)
	assert.Equal(t, "error", issue["severity"])
	assert.Equal(t, "not-found", issue["code"])
	assert.Equal(t, "Patient xyz not found", issue["diagnostics"])
}

func TestExtractMetaForHeaders(t *testing.T) {
	fhirJSON := json.RawMessage(`{"resourceType":"Patient","id":"p1","meta":{"versionId":"3","lastUpdated":"2025-06-15T12:00:00Z"}}`)
	vid, lu := extractMetaForHeaders(fhirJSON)
	assert.Equal(t, "3", vid)
	assert.Equal(t, "2025-06-15T12:00:00Z", lu)
}

func TestExtractMetaForHeaders_NoMeta(t *testing.T) {
	fhirJSON := json.RawMessage(`{"resourceType":"Patient","id":"p1"}`)
	vid, lu := extractMetaForHeaders(fhirJSON)
	assert.Equal(t, "", vid)
	assert.Equal(t, "", lu)
}

// --- ExtractPatientReference Tests ---

func TestExtractPatientReference(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected string
	}{
		{
			name:     "subject.reference",
			json:     `{"resourceType":"Encounter","subject":{"reference":"Patient/abc-123"}}`,
			expected: "abc-123",
		},
		{
			name:     "patient.reference",
			json:     `{"resourceType":"AllergyIntolerance","patient":{"reference":"Patient/xyz"}}`,
			expected: "xyz",
		},
		{
			name:     "no reference",
			json:     `{"resourceType":"Patient","id":"p1"}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pkgfhir.ExtractPatientReference([]byte(tt.json))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- Validate Resource Type Tests ---

func TestValidateResourceType(t *testing.T) {
	t.Run("matches", func(t *testing.T) {
		err := validateResourceType([]byte(`{"resourceType":"Patient"}`), "Patient")
		assert.NoError(t, err)
	})

	t.Run("mismatch", func(t *testing.T) {
		err := validateResourceType([]byte(`{"resourceType":"Encounter"}`), "Patient")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mismatch")
	})

	t.Run("missing", func(t *testing.T) {
		err := validateResourceType([]byte(`{"id":"p1"}`), "Patient")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestValidateBodyID(t *testing.T) {
	t.Run("matches", func(t *testing.T) {
		err := validateBodyID([]byte(`{"id":"abc"}`), "abc")
		assert.NoError(t, err)
	})

	t.Run("empty body id ok", func(t *testing.T) {
		err := validateBodyID([]byte(`{"resourceType":"Patient"}`), "abc")
		assert.NoError(t, err)
	})

	t.Run("mismatch", func(t *testing.T) {
		err := validateBodyID([]byte(`{"id":"xyz"}`), "abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mismatch")
	})
}
