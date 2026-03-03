package fhir

import (
	"encoding/json"
	"testing"
)

func TestGenerateStructureDefinition_Fields(t *testing.T) {
	def := GetProfileDef(ProfilePatient)
	if def == nil {
		t.Fatal("ProfilePatient not found")
	}

	data, err := GenerateStructureDefinition(def)
	if err != nil {
		t.Fatal(err)
	}

	var sd map[string]any
	if err := json.Unmarshal(data, &sd); err != nil {
		t.Fatal(err)
	}

	if sd["resourceType"] != "StructureDefinition" {
		t.Errorf("resourceType = %v", sd["resourceType"])
	}
	if sd["id"] != "OpenNucleus-Patient" {
		t.Errorf("id = %v", sd["id"])
	}
	if sd["url"] != ProfilePatient {
		t.Errorf("url = %v", sd["url"])
	}
	if sd["kind"] != "resource" {
		t.Errorf("kind = %v", sd["kind"])
	}
	if sd["type"] != "Patient" {
		t.Errorf("type = %v", sd["type"])
	}
	if sd["derivation"] != "constraint" {
		t.Errorf("derivation = %v", sd["derivation"])
	}
	if sd["fhirVersion"] != "4.0.1" {
		t.Errorf("fhirVersion = %v", sd["fhirVersion"])
	}
	if sd["baseDefinition"] != "http://hl7.org/fhir/StructureDefinition/Patient" {
		t.Errorf("baseDefinition = %v", sd["baseDefinition"])
	}

	diff, ok := sd["differential"].(map[string]any)
	if !ok {
		t.Fatal("differential not found")
	}
	elements, ok := diff["element"].([]any)
	if !ok {
		t.Fatal("differential.element not found")
	}
	// Root + 2 extensions for Patient profile
	if len(elements) != 3 {
		t.Errorf("element count = %d, want 3", len(elements))
	}
}

func TestGenerateAllStructureDefinitions_Count(t *testing.T) {
	result, err := GenerateAllStructureDefinitions()
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 5 {
		t.Errorf("got %d StructureDefinitions, want 5", len(result))
	}
}
