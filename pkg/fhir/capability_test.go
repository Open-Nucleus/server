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
	if len(resources) != 17 {
		t.Errorf("resource count = %d, want 17", len(resources))
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

func TestGenerateCapabilityStatement_SmartEnabled(t *testing.T) {
	cfg := CapabilityConfig{
		ServerName:   "Test",
		ServerURL:    "http://localhost:8080",
		Version:      "0.9.0",
		SmartEnabled: true,
		SmartBaseURL: "http://localhost:8080",
	}

	data, err := GenerateCapabilityStatement(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var cs map[string]any
	if err := json.Unmarshal(data, &cs); err != nil {
		t.Fatal(err)
	}

	rest0 := cs["rest"].([]any)[0].(map[string]any)
	security, ok := rest0["security"].(map[string]any)
	if !ok {
		t.Fatal("security section missing when SmartEnabled=true")
	}

	// Check service coding
	services := security["service"].([]any)
	svc := services[0].(map[string]any)
	coding := svc["coding"].([]any)
	code := coding[0].(map[string]any)
	if code["code"] != "SMART-on-FHIR" {
		t.Errorf("security service code = %v, want SMART-on-FHIR", code["code"])
	}

	// Check oauth-uris extension
	exts := security["extension"].([]any)
	oauthExt := exts[0].(map[string]any)
	if oauthExt["url"] != "http://fhir-registry.smarthealthit.org/StructureDefinition/oauth-uris" {
		t.Errorf("oauth-uris extension URL = %v", oauthExt["url"])
	}

	innerExts := oauthExt["extension"].([]any)
	if len(innerExts) != 4 {
		t.Errorf("oauth-uris inner extensions = %d, want 4", len(innerExts))
	}

	// Verify authorize endpoint
	auth := innerExts[0].(map[string]any)
	if auth["url"] != "authorize" || auth["valueUri"] != "http://localhost:8080/auth/smart/authorize" {
		t.Errorf("authorize extension = %v", auth)
	}

	// Verify token endpoint
	token := innerExts[1].(map[string]any)
	if token["url"] != "token" || token["valueUri"] != "http://localhost:8080/auth/smart/token" {
		t.Errorf("token extension = %v", token)
	}
}

func TestGenerateCapabilityStatement_SmartDisabled(t *testing.T) {
	cfg := CapabilityConfig{
		ServerName:   "Test",
		ServerURL:    "http://localhost:8080",
		Version:      "0.9.0",
		SmartEnabled: false,
	}

	data, err := GenerateCapabilityStatement(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var cs map[string]any
	json.Unmarshal(data, &cs)

	rest0 := cs["rest"].([]any)[0].(map[string]any)
	if _, ok := rest0["security"]; ok {
		t.Error("security section should be absent when SmartEnabled=false")
	}
}
