package fhir

import (
	"encoding/json"
	"testing"
)

func TestNewOperationOutcome_SingleIssue(t *testing.T) {
	issues := []OutcomeIssue{
		{Severity: "error", Code: "not-found", Diagnostics: "Patient/123 not found"},
	}
	data, err := NewOperationOutcome(issues)
	if err != nil {
		t.Fatal(err)
	}

	var oo map[string]any
	if err := json.Unmarshal(data, &oo); err != nil {
		t.Fatal(err)
	}
	if oo["resourceType"] != "OperationOutcome" {
		t.Errorf("resourceType = %v", oo["resourceType"])
	}
	issArr := oo["issue"].([]any)
	if len(issArr) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issArr))
	}
	iss0 := issArr[0].(map[string]any)
	if iss0["severity"] != "error" {
		t.Errorf("severity = %v", iss0["severity"])
	}
	if iss0["code"] != "not-found" {
		t.Errorf("code = %v", iss0["code"])
	}
}

func TestNewOperationOutcome_WithLocation(t *testing.T) {
	issues := []OutcomeIssue{
		{Severity: "error", Code: "required", Diagnostics: "name is required", Location: "Patient.name"},
	}
	data, err := NewOperationOutcome(issues)
	if err != nil {
		t.Fatal(err)
	}

	var oo map[string]any
	json.Unmarshal(data, &oo)
	iss0 := oo["issue"].([]any)[0].(map[string]any)
	expr := iss0["expression"].([]any)
	if len(expr) != 1 || expr[0] != "Patient.name" {
		t.Errorf("expression = %v", expr)
	}
}

func TestFromFieldErrors(t *testing.T) {
	errs := []FieldError{
		{Path: "name", Rule: "required", Message: "Patient.name is required"},
		{Path: "gender", Rule: "value_set", Message: "invalid gender"},
	}
	data, err := FromFieldErrors(errs)
	if err != nil {
		t.Fatal(err)
	}

	var oo map[string]any
	json.Unmarshal(data, &oo)
	issArr := oo["issue"].([]any)
	if len(issArr) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(issArr))
	}
	// Check mapped codes
	iss0 := issArr[0].(map[string]any)
	if iss0["code"] != "required" {
		t.Errorf("expected 'required' code, got %v", iss0["code"])
	}
	iss1 := issArr[1].(map[string]any)
	if iss1["code"] != "code-invalid" {
		t.Errorf("expected 'code-invalid' code, got %v", iss1["code"])
	}
}

func TestFromError(t *testing.T) {
	data, err := FromError("processing", "Internal server error")
	if err != nil {
		t.Fatal(err)
	}

	var oo map[string]any
	json.Unmarshal(data, &oo)
	if oo["resourceType"] != "OperationOutcome" {
		t.Errorf("resourceType = %v", oo["resourceType"])
	}
}

func TestFieldRuleToIssueCode_Mapping(t *testing.T) {
	tests := []struct{ rule, want string }{
		{"required", "required"},
		{"value_set", "code-invalid"},
		{"value", "code-invalid"},
		{"json", "structure"},
		{"type", "structure"},
		{"unknown", "invalid"},
	}
	for _, tt := range tests {
		got := fieldRuleToIssueCode(tt.rule)
		if got != tt.want {
			t.Errorf("fieldRuleToIssueCode(%q) = %q, want %q", tt.rule, got, tt.want)
		}
	}
}
