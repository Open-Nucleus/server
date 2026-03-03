package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

func newSmartContext(scope, launchPatient string) context.Context {
	claims := &model.NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "practitioner-1"},
		Role:             "physician",
		Scope:            scope,
		LaunchPatient:    launchPatient,
	}
	return context.WithValue(context.Background(), model.CtxClaims, claims)
}

func TestSmartScope_NoScopePassthrough(t *testing.T) {
	// Token without SMART scope should pass through.
	claims := &model.NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: "practitioner-1"},
		Role:             "physician",
	}
	ctx := context.WithValue(context.Background(), model.CtxClaims, claims)

	handler := SmartScope("Patient", "r")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/fhir/Patient/123", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestSmartScope_MatchingScope(t *testing.T) {
	ctx := newSmartContext("patient/Patient.r", "")

	handler := SmartScope("Patient", "r")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/fhir/Patient/123", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestSmartScope_WrongResourceDenied(t *testing.T) {
	ctx := newSmartContext("patient/Observation.r", "")

	handler := SmartScope("Patient", "r")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/fhir/Patient/123", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestSmartScope_WildcardResource(t *testing.T) {
	ctx := newSmartContext("user/*.cruds", "")

	handler := SmartScope("Observation", "c")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/fhir/Observation", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestSmartScope_PatientContextEnforcement(t *testing.T) {
	ctx := newSmartContext("patient/Patient.r", "patient-123")

	handler := SmartScope("Patient", "r")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request for wrong patient should be denied.
	req := httptest.NewRequest("GET", "/fhir/Patient/patient-999?patient=patient-999", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for wrong patient, got %d", rr.Code)
	}

	// Request for correct patient should pass.
	req2 := httptest.NewRequest("GET", "/fhir/Patient/patient-123?patient=patient-123", nil).WithContext(ctx)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200 for matching patient, got %d", rr2.Code)
	}
}

func TestSmartScope_MultipleScopes(t *testing.T) {
	ctx := newSmartContext("patient/Patient.r patient/Observation.rs", "")

	// Patient read should work.
	handler := SmartScope("Patient", "r")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/fhir/Patient/123", nil).WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for Patient.r, got %d", rr.Code)
	}

	// Observation search should work.
	handler2 := SmartScope("Observation", "s")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req2 := httptest.NewRequest("GET", "/fhir/Observation", nil).WithContext(ctx)
	rr2 := httptest.NewRecorder()
	handler2.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200 for Observation.s, got %d", rr2.Code)
	}

	// Condition read should be denied (not in scopes).
	handler3 := SmartScope("Condition", "r")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req3 := httptest.NewRequest("GET", "/fhir/Condition/123", nil).WithContext(ctx)
	rr3 := httptest.NewRecorder()
	handler3.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusForbidden {
		t.Errorf("expected 403 for Condition, got %d", rr3.Code)
	}
}
