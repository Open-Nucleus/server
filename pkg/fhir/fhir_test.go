package fhir

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGitPath_AllResourceTypes(t *testing.T) {
	tests := []struct {
		resType   string
		patientID string
		resID     string
		want      string
	}{
		{ResourcePatient, "", "pat-123", "patients/pat-123/Patient.json"},
		{ResourceEncounter, "pat-123", "enc-456", "patients/pat-123/encounters/enc-456.json"},
		{ResourceObservation, "pat-123", "obs-789", "patients/pat-123/observations/obs-789.json"},
		{ResourceCondition, "pat-123", "cond-1", "patients/pat-123/conditions/cond-1.json"},
		{ResourceMedicationRequest, "pat-123", "medrq-1", "patients/pat-123/medication-requests/medrq-1.json"},
		{ResourceAllergyIntolerance, "pat-123", "allergy-1", "patients/pat-123/allergy-intolerances/allergy-1.json"},
		{ResourceFlag, "pat-123", "flag-1", "patients/pat-123/flags/flag-1.json"},
		{ResourceDetectedIssue, "", "di-1", "alerts/di-1.json"},
		{ResourceSupplyDelivery, "", "sd-1", "supply/deliveries/sd-1.json"},
	}
	for _, tt := range tests {
		got := GitPath(tt.resType, tt.patientID, tt.resID)
		if got != tt.want {
			t.Errorf("GitPath(%q, %q, %q) = %q, want %q", tt.resType, tt.patientID, tt.resID, got, tt.want)
		}
	}
}

func TestSetMeta(t *testing.T) {
	input := []byte(`{"resourceType":"Patient","id":"p1"}`)
	now := time.Date(2026, 3, 15, 9, 42, 0, 0, time.UTC)

	out, err := SetMeta(input, now, "abc123", "clinic-01")
	if err != nil {
		t.Fatal(err)
	}

	var r map[string]any
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatal(err)
	}

	meta := r["meta"].(map[string]any)
	if meta["lastUpdated"] != "2026-03-15T09:42:00Z" {
		t.Errorf("lastUpdated = %v", meta["lastUpdated"])
	}
	if meta["versionId"] != "abc123" {
		t.Errorf("versionId = %v", meta["versionId"])
	}
	if meta["source"] != "clinic-01" {
		t.Errorf("source = %v", meta["source"])
	}
}

func TestAssignID(t *testing.T) {
	t.Run("assigns UUID when id missing", func(t *testing.T) {
		input := []byte(`{"resourceType":"Patient"}`)
		out, id, err := AssignID(input)
		if err != nil {
			t.Fatal(err)
		}
		if id == "" {
			t.Error("expected non-empty ID")
		}
		var r map[string]any
		json.Unmarshal(out, &r)
		if r["id"] != id {
			t.Errorf("JSON id = %v, returned id = %v", r["id"], id)
		}
	})

	t.Run("preserves existing id", func(t *testing.T) {
		input := []byte(`{"resourceType":"Patient","id":"existing-id"}`)
		_, id, err := AssignID(input)
		if err != nil {
			t.Fatal(err)
		}
		if id != "existing-id" {
			t.Errorf("expected existing-id, got %s", id)
		}
	})
}

func TestGetResourceType(t *testing.T) {
	input := []byte(`{"resourceType":"Encounter","id":"e1"}`)
	rt, err := GetResourceType(input)
	if err != nil {
		t.Fatal(err)
	}
	if rt != "Encounter" {
		t.Errorf("got %s, want Encounter", rt)
	}
}

func TestGetResourceType_Missing(t *testing.T) {
	input := []byte(`{"id":"e1"}`)
	_, err := GetResourceType(input)
	if err == nil {
		t.Error("expected error for missing resourceType")
	}
}

func TestValidatePatient_AllRequired(t *testing.T) {
	input := []byte(`{
		"resourceType": "Patient",
		"name": [{"family": "Ibrahim", "given": ["Fatima"]}],
		"gender": "female",
		"birthDate": "1990-01-15"
	}`)
	errs := Validate(ResourcePatient, input)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %+v", len(errs), errs)
	}
}

func TestValidatePatient_MissingName(t *testing.T) {
	input := []byte(`{
		"resourceType": "Patient",
		"gender": "female",
		"birthDate": "1990-01-15"
	}`)
	errs := Validate(ResourcePatient, input)
	if len(errs) == 0 {
		t.Error("expected errors for missing name")
	}
	found := false
	for _, e := range errs {
		if e.Path == "name" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for path 'name'")
	}
}

func TestValidatePatient_InvalidGender(t *testing.T) {
	input := []byte(`{
		"resourceType": "Patient",
		"name": [{"family": "Test", "given": ["User"]}],
		"gender": "invalid",
		"birthDate": "1990"
	}`)
	errs := Validate(ResourcePatient, input)
	found := false
	for _, e := range errs {
		if e.Path == "gender" && e.Rule == "value_set" {
			found = true
		}
	}
	if !found {
		t.Error("expected value_set error for gender")
	}
}

func TestValidateEncounter_MissingStatus(t *testing.T) {
	input := []byte(`{
		"resourceType": "Encounter",
		"class": {"code": "AMB"},
		"subject": {"reference": "Patient/p1"},
		"period": {"start": "2026-01-01"}
	}`)
	errs := Validate(ResourceEncounter, input)
	found := false
	for _, e := range errs {
		if e.Path == "status" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing status")
	}
}

func TestValidateObservation_MissingCode(t *testing.T) {
	input := []byte(`{
		"resourceType": "Observation",
		"status": "final",
		"subject": {"reference": "Patient/p1"},
		"effectiveDateTime": "2026-01-01"
	}`)
	errs := Validate(ResourceObservation, input)
	found := false
	for _, e := range errs {
		if e.Path == "code" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing code")
	}
}

func TestExtractPatientFields(t *testing.T) {
	input := []byte(`{
		"resourceType": "Patient",
		"id": "p1",
		"name": [{"family": "Ibrahim", "given": ["Fatima", "Aisha"]}],
		"gender": "female",
		"birthDate": "1990-01-15",
		"active": true,
		"meta": {"lastUpdated": "2026-03-15T09:42:00Z"}
	}`)

	row, err := ExtractPatientFields(input, "site-1", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if row.ID != "p1" {
		t.Errorf("ID = %s", row.ID)
	}
	if row.FamilyName != "Ibrahim" {
		t.Errorf("FamilyName = %s", row.FamilyName)
	}
	if row.Gender != "female" {
		t.Errorf("Gender = %s", row.Gender)
	}
	if !row.Active {
		t.Error("expected Active = true")
	}
}

func TestExtractEncounterFields(t *testing.T) {
	input := []byte(`{
		"resourceType": "Encounter",
		"id": "e1",
		"status": "finished",
		"class": {"code": "AMB"},
		"subject": {"reference": "Patient/p1"},
		"period": {"start": "2026-01-01T09:00:00Z", "end": "2026-01-01T10:00:00Z"},
		"meta": {"lastUpdated": "2026-03-15T09:42:00Z"}
	}`)

	row, err := ExtractEncounterFields(input, "p1", "site-1", "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if row.ID != "e1" {
		t.Errorf("ID = %s", row.ID)
	}
	if row.Status != "finished" {
		t.Errorf("Status = %s", row.Status)
	}
	if row.ClassCode != "AMB" {
		t.Errorf("ClassCode = %s", row.ClassCode)
	}
}

func TestApplySoftDelete_Patient(t *testing.T) {
	input := []byte(`{"resourceType":"Patient","id":"p1","active":true}`)
	out, err := ApplySoftDelete(ResourcePatient, input)
	if err != nil {
		t.Fatal(err)
	}
	var r map[string]any
	json.Unmarshal(out, &r)
	if r["active"] != false {
		t.Errorf("expected active=false, got %v", r["active"])
	}
}

func TestApplySoftDelete_Encounter(t *testing.T) {
	input := []byte(`{"resourceType":"Encounter","id":"e1","status":"finished"}`)
	out, err := ApplySoftDelete(ResourceEncounter, input)
	if err != nil {
		t.Fatal(err)
	}
	var r map[string]any
	json.Unmarshal(out, &r)
	if r["status"] != "entered-in-error" {
		t.Errorf("expected status=entered-in-error, got %v", r["status"])
	}
}
