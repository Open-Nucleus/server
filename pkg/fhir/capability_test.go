package fhir

import (
	"encoding/json"
	"testing"
	"time"
)

func TestGenerateCapabilityStatement_Structure(t *testing.T) {
	cfg := CapabilityConfig{
		ServerName:  "Open Nucleus",
		ServerURL:   "http://localhost:8080",
		Version:     "0.6.0",
		PublishedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := GenerateCapabilityStatement(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var cs map[string]any
	if err := json.Unmarshal(data, &cs); err != nil {
		t.Fatal(err)
	}

	if cs["resourceType"] != "CapabilityStatement" {
		t.Errorf("resourceType = %v", cs["resourceType"])
	}
	if cs["fhirVersion"] != "4.0.1" {
		t.Errorf("fhirVersion = %v", cs["fhirVersion"])
	}
	if cs["kind"] != "instance" {
		t.Errorf("kind = %v", cs["kind"])
	}
	if cs["status"] != "active" {
		t.Errorf("status = %v", cs["status"])
	}
}

func TestGenerateCapabilityStatement_RestResources(t *testing.T) {
	cfg := CapabilityConfig{
		ServerName: "Test",
		ServerURL:  "http://localhost:8080",
		Version:    "0.1.0",
	}

	data, _ := GenerateCapabilityStatement(cfg)
	var cs map[string]any
	json.Unmarshal(data, &cs)

	rest := cs["rest"].([]any)
	if len(rest) != 1 {
		t.Fatalf("rest length = %d, want 1", len(rest))
	}

	rest0 := rest[0].(map[string]any)
	if rest0["mode"] != "server" {
		t.Errorf("rest mode = %v", rest0["mode"])
	}

	resources := rest0["resource"].([]any)
	if len(resources) != 15 {
		t.Errorf("resource count = %d, want 15", len(resources))
	}
}

func TestGenerateCapabilityStatement_SearchParams(t *testing.T) {
	cfg := CapabilityConfig{
		ServerName: "Test",
		ServerURL:  "http://localhost:8080",
		Version:    "0.1.0",
	}

	data, _ := GenerateCapabilityStatement(cfg)
	var cs map[string]any
	json.Unmarshal(data, &cs)

	rest0 := cs["rest"].([]any)[0].(map[string]any)
	resources := rest0["resource"].([]any)

	// Find Patient resource and check it has search params
	for _, r := range resources {
		res := r.(map[string]any)
		if res["type"] == "Patient" {
			params, ok := res["searchParam"].([]any)
			if !ok || len(params) == 0 {
				t.Error("Patient should have search params")
			}
			return
		}
	}
	t.Error("Patient resource not found in CapabilityStatement")
}

func TestGenerateCapabilityStatement_Sorted(t *testing.T) {
	cfg := CapabilityConfig{ServerName: "Test", ServerURL: "http://localhost:8080", Version: "0.1.0"}
	data, _ := GenerateCapabilityStatement(cfg)
	var cs map[string]any
	json.Unmarshal(data, &cs)

	rest0 := cs["rest"].([]any)[0].(map[string]any)
	resources := rest0["resource"].([]any)

	// Verify sorted alphabetically
	prev := ""
	for _, r := range resources {
		res := r.(map[string]any)
		cur := res["type"].(string)
		if cur < prev {
			t.Errorf("resources not sorted: %q came after %q", cur, prev)
		}
		prev = cur
	}
}
