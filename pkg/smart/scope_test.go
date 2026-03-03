package smart

import (
	"strings"
	"testing"
)

func TestParseScope_ValidResource(t *testing.T) {
	tests := []struct {
		input   string
		ctx     string
		res     string
		inters  string
	}{
		{"patient/Patient.r", "patient", "Patient", "r"},
		{"patient/Observation.rs", "patient", "Observation", "rs"},
		{"user/Condition.cruds", "user", "Condition", "cruds"},
		{"system/MedicationRequest.r", "system", "MedicationRequest", "r"},
		{"patient/*.r", "patient", "*", "r"},
		{"user/*.*", "user", "*", "*"},
	}
	for _, tc := range tests {
		sc, err := ParseScope(tc.input)
		if err != nil {
			t.Fatalf("ParseScope(%q) unexpected error: %v", tc.input, err)
		}
		if sc.Context != tc.ctx {
			t.Errorf("ParseScope(%q).Context = %q, want %q", tc.input, sc.Context, tc.ctx)
		}
		if sc.Resource != tc.res {
			t.Errorf("ParseScope(%q).Resource = %q, want %q", tc.input, sc.Resource, tc.res)
		}
		if sc.Interactions != tc.inters {
			t.Errorf("ParseScope(%q).Interactions = %q, want %q", tc.input, sc.Interactions, tc.inters)
		}
		if sc.Raw != tc.input {
			t.Errorf("ParseScope(%q).Raw = %q, want %q", tc.input, sc.Raw, tc.input)
		}
	}
}

func TestParseScope_SpecialScopes(t *testing.T) {
	specials := []string{"launch", "launch/patient", "launch/encounter", "fhirUser", "offline_access", "openid"}
	for _, s := range specials {
		sc, err := ParseScope(s)
		if err != nil {
			t.Fatalf("ParseScope(%q) unexpected error: %v", s, err)
		}
		if !sc.IsSpecial() {
			t.Errorf("ParseScope(%q).IsSpecial() = false, want true", s)
		}
		if sc.Raw != s {
			t.Errorf("ParseScope(%q).Raw = %q, want %q", s, sc.Raw, s)
		}
	}
}

func TestParseScope_Invalid(t *testing.T) {
	invalids := []string{
		"",                        // empty
		"patient",                 // no slash
		"bad/Patient.r",           // invalid context
		"patient/Unknown.r",       // unknown resource
		"patient/Patient",         // no dot
		"patient/Patient.",        // empty interactions
		"patient/.r",              // empty resource
		"patient/Patient.xyz",     // invalid interaction char
	}
	for _, s := range invalids {
		_, err := ParseScope(s)
		if err == nil {
			t.Errorf("ParseScope(%q) expected error, got nil", s)
		}
	}
}

func TestParseScopes(t *testing.T) {
	scopes, err := ParseScopes("patient/Patient.r launch user/Observation.rs")
	if err != nil {
		t.Fatalf("ParseScopes() unexpected error: %v", err)
	}
	if len(scopes) != 3 {
		t.Fatalf("ParseScopes() returned %d scopes, want 3", len(scopes))
	}
	// First should be patient/Patient.r
	if scopes[0].Context != "patient" || scopes[0].Resource != "Patient" {
		t.Errorf("scopes[0] = %+v, want patient/Patient.r", scopes[0])
	}
	// Second should be special launch
	if !scopes[1].IsSpecial() || scopes[1].Raw != "launch" {
		t.Errorf("scopes[1] = %+v, want special launch", scopes[1])
	}
}

func TestParseScopes_Empty(t *testing.T) {
	scopes, err := ParseScopes("")
	if err != nil {
		t.Fatalf("ParseScopes(\"\") unexpected error: %v", err)
	}
	if scopes != nil {
		t.Errorf("ParseScopes(\"\") = %v, want nil", scopes)
	}
}

func TestScope_Allows(t *testing.T) {
	sc, _ := ParseScope("patient/Observation.rs")

	if !sc.Allows("r", "Observation") {
		t.Error("expected Allows(r, Observation) = true")
	}
	if !sc.Allows("s", "Observation") {
		t.Error("expected Allows(s, Observation) = true")
	}
	if sc.Allows("c", "Observation") {
		t.Error("expected Allows(c, Observation) = false")
	}
	if sc.Allows("r", "Patient") {
		t.Error("expected Allows(r, Patient) = false for non-matching resource")
	}

	// Wildcard resource
	wild, _ := ParseScope("user/*.*")
	if !wild.Allows("c", "Patient") {
		t.Error("expected wildcard Allows(c, Patient) = true")
	}
	if !wild.Allows("d", "Observation") {
		t.Error("expected wildcard Allows(d, Observation) = true")
	}

	// Special scope should never allow
	special, _ := ParseScope("launch")
	if special.Allows("r", "Patient") {
		t.Error("expected special scope Allows() = false")
	}
}

func TestFilterByResource(t *testing.T) {
	scopes, _ := ParseScopes("patient/Patient.r patient/Observation.rs user/*.cruds launch")
	filtered := FilterByResource(scopes, "Patient")
	if len(filtered) != 2 {
		t.Fatalf("FilterByResource(Patient) returned %d scopes, want 2", len(filtered))
	}
	// Should include patient/Patient.r and user/*.cruds
	if filtered[0].Resource != "Patient" {
		t.Errorf("filtered[0].Resource = %q, want Patient", filtered[0].Resource)
	}
	if filtered[1].Resource != "*" {
		t.Errorf("filtered[1].Resource = %q, want *", filtered[1].Resource)
	}
}

func TestScopesToPermissions(t *testing.T) {
	scopes, _ := ParseScopes("patient/Observation.rs patient/Patient.cruds launch")
	perms := ScopesToPermissions(scopes)

	// observation:read from .rs, patient:read + patient:write from .cruds
	permSet := map[string]bool{}
	for _, p := range perms {
		permSet[p] = true
	}

	if !permSet["observation:read"] {
		t.Error("expected observation:read in permissions")
	}
	if !permSet["patient:read"] {
		t.Error("expected patient:read in permissions")
	}
	if !permSet["patient:write"] {
		t.Error("expected patient:write in permissions")
	}
}

func TestAllSupportedScopes(t *testing.T) {
	scopes := AllSupportedScopes()
	if len(scopes) < 10 {
		t.Fatalf("AllSupportedScopes() returned %d scopes, expected more", len(scopes))
	}

	// Check that special scopes are included.
	found := false
	for _, s := range scopes {
		if s == "launch" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AllSupportedScopes() does not include 'launch'")
	}

	// Check that a resource scope is included.
	found = false
	for _, s := range scopes {
		if strings.HasPrefix(s, "patient/Patient.") {
			found = true
			break
		}
	}
	if !found {
		t.Error("AllSupportedScopes() does not include patient/Patient scopes")
	}
}

func TestScope_String(t *testing.T) {
	sc, _ := ParseScope("patient/Observation.rs")
	if sc.String() != "patient/Observation.rs" {
		t.Errorf("String() = %q, want %q", sc.String(), "patient/Observation.rs")
	}

	special, _ := ParseScope("launch")
	if special.String() != "launch" {
		t.Errorf("String() = %q, want %q", special.String(), "launch")
	}
}
