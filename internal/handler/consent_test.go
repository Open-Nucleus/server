package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/FibrinLab/open-nucleus/internal/service"
)

// --- mock ConsentService ---

type mockConsentService struct {
	grantResp *service.ConsentGrantResponse
	grantErr  error
	listResp  *service.ConsentListResponse
	listErr   error
	revokeErr error
	vcResp    *service.ConsentVCResponse
	vcErr     error
}

func (m *mockConsentService) CheckAccess(ctx context.Context, patientID, performerID, role string) (*service.ConsentAccessDecision, error) {
	return &service.ConsentAccessDecision{Allowed: true, Reason: "mock"}, nil
}

func (m *mockConsentService) GrantConsent(ctx context.Context, patientID, performerID, scope string, period *service.ConsentPeriod, category string) (*service.ConsentGrantResponse, error) {
	return m.grantResp, m.grantErr
}

func (m *mockConsentService) RevokeConsent(ctx context.Context, consentID string) error {
	return m.revokeErr
}

func (m *mockConsentService) ListConsentsForPatient(ctx context.Context, patientID string, page, perPage int) (*service.ConsentListResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockConsentService) IssueConsentVC(ctx context.Context, consentID string) (*service.ConsentVCResponse, error) {
	return m.vcResp, m.vcErr
}

func TestListConsents(t *testing.T) {
	svc := &mockConsentService{
		listResp: &service.ConsentListResponse{
			Consents: []service.ConsentSummary{
				{ID: "c1", PatientID: "p1", Status: "active"},
			},
		},
	}

	h := NewConsentHandler(svc)
	r := chi.NewRouter()
	r.Get("/api/v1/patients/{id}/consents", h.ListConsents)

	req := httptest.NewRequest("GET", "/api/v1/patients/p1/consents", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGrantConsent_Success(t *testing.T) {
	svc := &mockConsentService{
		grantResp: &service.ConsentGrantResponse{
			ConsentID:  "consent-new",
			CommitHash: "abc123",
			Status:     "active",
		},
	}

	h := NewConsentHandler(svc)
	r := chi.NewRouter()
	r.Post("/api/v1/patients/{id}/consents", h.GrantConsent)

	body, _ := json.Marshal(map[string]string{
		"performer_id": "device-1",
		"scope":        "patient-privacy",
	})

	req := httptest.NewRequest("POST", "/api/v1/patients/p1/consents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGrantConsent_MissingPerformer(t *testing.T) {
	svc := &mockConsentService{}
	h := NewConsentHandler(svc)
	r := chi.NewRouter()
	r.Post("/api/v1/patients/{id}/consents", h.GrantConsent)

	body, _ := json.Marshal(map[string]string{"scope": "patient-privacy"})
	req := httptest.NewRequest("POST", "/api/v1/patients/p1/consents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGrantConsent_DefaultScope(t *testing.T) {
	var capturedScope string
	svc := &mockConsentService{
		grantResp: &service.ConsentGrantResponse{ConsentID: "c1", Status: "active"},
	}

	// Wrap to capture scope
	h := NewConsentHandler(&scopeCapture{svc: svc, scope: &capturedScope})
	r := chi.NewRouter()
	r.Post("/api/v1/patients/{id}/consents", h.GrantConsent)

	body, _ := json.Marshal(map[string]string{"performer_id": "device-1"})
	req := httptest.NewRequest("POST", "/api/v1/patients/p1/consents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if capturedScope != "patient-privacy" {
		t.Errorf("default scope = %q, want 'patient-privacy'", capturedScope)
	}
}

type scopeCapture struct {
	svc   *mockConsentService
	scope *string
}

func (s *scopeCapture) CheckAccess(ctx context.Context, patientID, performerID, role string) (*service.ConsentAccessDecision, error) {
	return s.svc.CheckAccess(ctx, patientID, performerID, role)
}
func (s *scopeCapture) GrantConsent(ctx context.Context, patientID, performerID, scope string, period *service.ConsentPeriod, category string) (*service.ConsentGrantResponse, error) {
	*s.scope = scope
	return s.svc.GrantConsent(ctx, patientID, performerID, scope, period, category)
}
func (s *scopeCapture) RevokeConsent(ctx context.Context, consentID string) error {
	return s.svc.RevokeConsent(ctx, consentID)
}
func (s *scopeCapture) ListConsentsForPatient(ctx context.Context, patientID string, page, perPage int) (*service.ConsentListResponse, error) {
	return s.svc.ListConsentsForPatient(ctx, patientID, page, perPage)
}
func (s *scopeCapture) IssueConsentVC(ctx context.Context, consentID string) (*service.ConsentVCResponse, error) {
	return s.svc.IssueConsentVC(ctx, consentID)
}

func TestRevokeConsent_Success(t *testing.T) {
	svc := &mockConsentService{}
	h := NewConsentHandler(svc)
	r := chi.NewRouter()
	r.Delete("/api/v1/consents/{consentId}", h.RevokeConsent)

	req := httptest.NewRequest("DELETE", "/api/v1/consents/c1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestIssueVC_Success(t *testing.T) {
	svc := &mockConsentService{
		vcResp: &service.ConsentVCResponse{
			VerifiableCredential: map[string]string{"type": "ConsentGrant"},
		},
	}
	h := NewConsentHandler(svc)
	r := chi.NewRouter()
	r.Post("/api/v1/consents/{consentId}/vc", h.IssueVC)

	req := httptest.NewRequest("POST", "/api/v1/consents/c1/vc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
