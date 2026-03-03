package fhir

import "testing"

func TestAllProfileDefs_Count(t *testing.T) {
	defs := AllProfileDefs()
	if len(defs) != 5 {
		t.Errorf("AllProfileDefs() returned %d, want 5", len(defs))
	}
}

func TestProfilesForResource_Patient(t *testing.T) {
	profiles := ProfilesForResource(ResourcePatient)
	if len(profiles) != 1 {
		t.Fatalf("ProfilesForResource(Patient) = %d, want 1", len(profiles))
	}
	if profiles[0].URL != ProfilePatient {
		t.Errorf("URL = %s, want %s", profiles[0].URL, ProfilePatient)
	}
}

func TestProfilesForResource_Observation(t *testing.T) {
	profiles := ProfilesForResource(ResourceObservation)
	if len(profiles) != 1 {
		t.Fatalf("ProfilesForResource(Observation) = %d, want 1", len(profiles))
	}
	if profiles[0].URL != ProfileGrowthObservation {
		t.Errorf("URL = %s, want %s", profiles[0].URL, ProfileGrowthObservation)
	}
}

func TestGetProfileDef_Known(t *testing.T) {
	def := GetProfileDef(ProfileImmunization)
	if def == nil {
		t.Fatal("expected non-nil for ProfileImmunization")
	}
	if def.Name != "OpenNucleus-Immunization" {
		t.Errorf("Name = %s", def.Name)
	}
	if def.BaseResource != ResourceImmunization {
		t.Errorf("BaseResource = %s", def.BaseResource)
	}
}

func TestGetProfileDef_Unknown(t *testing.T) {
	def := GetProfileDef("http://example.com/fake")
	if def != nil {
		t.Errorf("expected nil for unknown profile, got %+v", def)
	}
}

func TestGetMetaProfiles(t *testing.T) {
	resource := map[string]any{
		"meta": map[string]any{
			"profile": []any{
				ProfilePatient,
				ProfileImmunization,
			},
		},
	}
	profiles := GetMetaProfiles(resource)
	if len(profiles) != 2 {
		t.Fatalf("got %d profiles, want 2", len(profiles))
	}
	if profiles[0] != ProfilePatient {
		t.Errorf("profiles[0] = %s", profiles[0])
	}
}

func TestGetMetaProfiles_NoMeta(t *testing.T) {
	resource := map[string]any{
		"resourceType": "Patient",
	}
	profiles := GetMetaProfiles(resource)
	if len(profiles) != 0 {
		t.Errorf("expected 0 profiles, got %d", len(profiles))
	}
}
