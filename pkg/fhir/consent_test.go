package fhir

import (
	"encoding/json"
	"testing"
)

func TestExtractConsentFields(t *testing.T) {
	consent := map[string]any{
		"resourceType": "Consent",
		"id":           "consent-123",
		"status":       "active",
		"scope": map[string]any{
			"coding": []any{
				map[string]any{
					"system": "http://terminology.hl7.org/CodeSystem/consentscope",
					"code":   "patient-privacy",
				},
			},
		},
		"patient": map[string]any{
			"reference": "Patient/pat-001",
		},
		"performer": []any{
			map[string]any{
				"reference": "device-abc",
			},
		},
		"provision": map[string]any{
			"type": "permit",
			"period": map[string]any{
				"start": "2024-01-01T00:00:00Z",
				"end":   "2025-01-01T00:00:00Z",
			},
		},
		"category": []any{
			map[string]any{
				"coding": []any{
					map[string]any{
						"system": "http://terminology.hl7.org/CodeSystem/consentcategorycodes",
						"code":   "npp",
					},
				},
			},
		},
		"meta": map[string]any{
			"lastUpdated": "2024-06-15T10:30:00Z",
		},
	}

	data, err := json.Marshal(consent)
	if err != nil {
		t.Fatal(err)
	}

	row, err := ExtractConsentFields(data, "abc123hash")
	if err != nil {
		t.Fatalf("ExtractConsentFields: %v", err)
	}

	if row.ID != "consent-123" {
		t.Errorf("ID = %q, want %q", row.ID, "consent-123")
	}
	if row.Status != "active" {
		t.Errorf("Status = %q, want %q", row.Status, "active")
	}
	if row.ScopeCode != "patient-privacy" {
		t.Errorf("ScopeCode = %q, want %q", row.ScopeCode, "patient-privacy")
	}
	if row.PatientID != "pat-001" {
		t.Errorf("PatientID = %q, want %q", row.PatientID, "pat-001")
	}
	if row.PerformerID != "device-abc" {
		t.Errorf("PerformerID = %q, want %q", row.PerformerID, "device-abc")
	}
	if row.ProvisionType != "permit" {
		t.Errorf("ProvisionType = %q, want %q", row.ProvisionType, "permit")
	}
	if row.PeriodStart == nil || *row.PeriodStart != "2024-01-01T00:00:00Z" {
		t.Errorf("PeriodStart = %v, want %q", row.PeriodStart, "2024-01-01T00:00:00Z")
	}
	if row.PeriodEnd == nil || *row.PeriodEnd != "2025-01-01T00:00:00Z" {
		t.Errorf("PeriodEnd = %v, want %q", row.PeriodEnd, "2025-01-01T00:00:00Z")
	}
	if row.Category == nil || *row.Category != "npp" {
		t.Errorf("Category = %v, want %q", row.Category, "npp")
	}
	if row.LastUpdated != "2024-06-15T10:30:00Z" {
		t.Errorf("LastUpdated = %q, want %q", row.LastUpdated, "2024-06-15T10:30:00Z")
	}
	if row.GitBlobHash != "abc123hash" {
		t.Errorf("GitBlobHash = %q, want %q", row.GitBlobHash, "abc123hash")
	}
}

func TestExtractConsentFields_Minimal(t *testing.T) {
	consent := map[string]any{
		"resourceType": "Consent",
		"id":           "c-min",
		"status":       "draft",
		"patient": map[string]any{
			"reference": "Patient/p1",
		},
		"provision": map[string]any{
			"type": "deny",
		},
	}

	data, _ := json.Marshal(consent)
	row, err := ExtractConsentFields(data, "hash")
	if err != nil {
		t.Fatalf("ExtractConsentFields: %v", err)
	}

	if row.PeriodStart != nil {
		t.Error("expected nil PeriodStart for minimal consent")
	}
	if row.PeriodEnd != nil {
		t.Error("expected nil PeriodEnd for minimal consent")
	}
	if row.Category != nil {
		t.Error("expected nil Category for minimal consent")
	}
}

func TestExtractConsentFields_InvalidJSON(t *testing.T) {
	_, err := ExtractConsentFields([]byte("not json"), "hash")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateConsent_Valid(t *testing.T) {
	r := map[string]any{
		"status": "active",
		"scope": map[string]any{
			"coding": []any{
				map[string]any{"code": "patient-privacy"},
			},
		},
		"patient": map[string]any{
			"reference": "Patient/p1",
		},
		"performer": []any{
			map[string]any{"reference": "device-1"},
		},
		"provision": map[string]any{
			"type": "permit",
		},
	}

	errs := validateConsent(r)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateConsent_MissingFields(t *testing.T) {
	r := map[string]any{}
	errs := validateConsent(r)

	// Should have errors for: status, scope, patient, performer, provision
	if len(errs) < 4 {
		t.Fatalf("expected at least 4 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateConsent_InvalidStatus(t *testing.T) {
	r := map[string]any{
		"status": "invalid-status",
		"scope": map[string]any{
			"coding": []any{
				map[string]any{"code": "patient-privacy"},
			},
		},
		"patient": map[string]any{
			"reference": "Patient/p1",
		},
		"performer": []any{
			map[string]any{"reference": "device-1"},
		},
		"provision": map[string]any{
			"type": "permit",
		},
	}

	errs := validateConsent(r)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Path != "status" {
		t.Errorf("expected error on 'status', got %q", errs[0].Path)
	}
}

func TestValidateConsent_InvalidProvisionType(t *testing.T) {
	r := map[string]any{
		"status": "active",
		"scope": map[string]any{
			"coding": []any{
				map[string]any{"code": "treatment"},
			},
		},
		"patient": map[string]any{
			"reference": "Patient/p1",
		},
		"performer": []any{
			map[string]any{"reference": "device-1"},
		},
		"provision": map[string]any{
			"type": "invalid",
		},
	}

	errs := validateConsent(r)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Path != "provision.type" {
		t.Errorf("expected error on 'provision.type', got %q", errs[0].Path)
	}
}

func TestIsConsentActive(t *testing.T) {
	if !IsConsentActive("active") {
		t.Error("expected active to be true")
	}
	if IsConsentActive("draft") {
		t.Error("expected draft to be false")
	}
	if IsConsentActive("inactive") {
		t.Error("expected inactive to be false")
	}
}
