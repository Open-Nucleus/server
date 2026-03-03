package fhir

import "testing"

func TestExtractExtension_Found(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
		"extension": []any{
			map[string]any{
				"url":         extBaseURL + "national-health-id",
				"valueString": "NIN-12345",
			},
		},
	}
	val, found := ExtractExtension(resource, extBaseURL+"national-health-id")
	if !found {
		t.Fatal("expected extension to be found")
	}
	if val != "NIN-12345" {
		t.Errorf("got %v, want NIN-12345", val)
	}
}

func TestExtractExtension_NotFound(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
		"extension":    []any{},
	}
	_, found := ExtractExtension(resource, extBaseURL+"national-health-id")
	if found {
		t.Error("expected extension not to be found")
	}
}

func TestHasExtension(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
		"extension": []any{
			map[string]any{
				"url":         extBaseURL + "ethnic-group",
				"valueCoding": map[string]any{"code": "hausa"},
			},
		},
	}
	if !HasExtension(resource, extBaseURL+"ethnic-group") {
		t.Error("expected HasExtension to return true")
	}
	if HasExtension(resource, extBaseURL+"nonexistent") {
		t.Error("expected HasExtension to return false for unknown URL")
	}
}

func TestValidateExtensions_RequiredMissing(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
	}
	defs := []ExtensionDef{
		{URL: extBaseURL + "required-ext", ValueType: "valueString", Required: true, Short: "Required"},
	}
	errs := ValidateExtensions(resource, defs)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Rule != "required_extension" {
		t.Errorf("rule = %s, want required_extension", errs[0].Rule)
	}
}

func TestValidateExtensions_TypeMismatch(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
		"extension": []any{
			map[string]any{
				"url":          extBaseURL + "who-zscore",
				"valueInteger": 42, // should be valueDecimal
			},
		},
	}
	defs := []ExtensionDef{
		{URL: extBaseURL + "who-zscore", ValueType: "valueDecimal", Short: "WHO z-score"},
	}
	errs := ValidateExtensions(resource, defs)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %+v", len(errs), errs)
	}
	if errs[0].Rule != "extension_type" {
		t.Errorf("rule = %s, want extension_type", errs[0].Rule)
	}
}

func TestValidateExtensions_UnknownPassthrough(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
		"extension": []any{
			map[string]any{
				"url":         "http://example.com/custom",
				"valueString": "custom-value",
			},
		},
	}
	errs := ValidateExtensions(resource, nil)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors for unknown extensions, got %d", len(errs))
	}
}
