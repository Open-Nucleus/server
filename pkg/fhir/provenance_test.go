package fhir

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGenerateProvenance_Create(t *testing.T) {
	ctx := ProvenanceContext{
		TargetResourceType: "Patient",
		TargetResourceID:   "abc-123",
		Activity:           OpCreate,
		PractitionerID:     "dr-smith",
		DeviceID:           "node-01",
		SiteID:             "site-alpha",
		Recorded:           time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	data, id, err := GenerateProvenance(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Error("expected non-empty provenance ID")
	}

	var prov map[string]any
	if err := json.Unmarshal(data, &prov); err != nil {
		t.Fatal(err)
	}
	if prov["resourceType"] != "Provenance" {
		t.Errorf("resourceType = %v", prov["resourceType"])
	}

	// Check target reference
	targets := prov["target"].([]any)
	target0 := targets[0].(map[string]any)
	if target0["reference"] != "Patient/abc-123" {
		t.Errorf("target reference = %v", target0["reference"])
	}

	// Check activity code
	activity := prov["activity"].(map[string]any)
	codings := activity["coding"].([]any)
	coding0 := codings[0].(map[string]any)
	if coding0["code"] != "CREATE" {
		t.Errorf("activity code = %v", coding0["code"])
	}

	// Check agents (should have 2: author + custodian)
	agents := prov["agent"].([]any)
	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestGenerateProvenance_WithoutDevice(t *testing.T) {
	ctx := ProvenanceContext{
		TargetResourceType: "Observation",
		TargetResourceID:   "obs-456",
		Activity:           OpUpdate,
		PractitionerID:     "dr-jones",
		SiteID:             "site-beta",
	}

	data, _, err := GenerateProvenance(ctx)
	if err != nil {
		t.Fatal(err)
	}

	var prov map[string]any
	json.Unmarshal(data, &prov)

	// Only author agent (no custodian)
	agents := prov["agent"].([]any)
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
}

func TestGenerateProvenance_UniqueIDs(t *testing.T) {
	ctx := ProvenanceContext{
		TargetResourceType: "Patient",
		TargetResourceID:   "p1",
		Activity:           OpCreate,
		PractitionerID:     "dr-x",
	}

	_, id1, _ := GenerateProvenance(ctx)
	_, id2, _ := GenerateProvenance(ctx)
	if id1 == id2 {
		t.Error("provenance IDs should be unique")
	}
}
