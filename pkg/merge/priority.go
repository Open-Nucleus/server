package merge

import (
	"encoding/json"
	"strings"
)

// ClassifyResource determines the sync priority tier for a FHIR resource.
func ClassifyResource(gitPath string, fhirJSON json.RawMessage) SyncPriority {
	resourceType := resourceTypeFromPath(gitPath)

	// Tier 1: alerts, revocations, flags
	switch resourceType {
	case "Flag", "DetectedIssue":
		return Tier1Critical
	}

	// Parse resource for status-based classification
	var resource map[string]any
	if err := json.Unmarshal(fhirJSON, &resource); err != nil {
		return Tier3Clinical // default if parse fails
	}

	status, _ := resource["status"].(string)

	switch resourceType {
	case "Patient":
		active, _ := resource["active"].(bool)
		if active || status == "" {
			return Tier2Active
		}
		return Tier4Resolved

	case "Encounter":
		switch status {
		case "in-progress", "planned", "arrived", "triaged":
			return Tier2Active
		case "finished", "cancelled", "entered-in-error":
			return Tier4Resolved
		}
		return Tier2Active

	case "MedicationRequest":
		switch status {
		case "active", "on-hold":
			return Tier2Active
		case "completed", "cancelled", "stopped", "entered-in-error":
			return Tier4Resolved
		}
		return Tier2Active

	case "Observation":
		return Tier3Clinical

	case "Condition":
		cs := extractClinicalStatus(resource)
		switch cs {
		case "active", "recurrence", "relapse":
			return Tier3Clinical
		case "inactive", "remission", "resolved":
			return Tier4Resolved
		}
		return Tier3Clinical

	case "AllergyIntolerance":
		cs := extractClinicalStatus(resource)
		if cs == "active" {
			return Tier2Active
		}
		return Tier4Resolved

	case "SupplyDelivery":
		return Tier3Clinical
	}

	// Default to history tier for unknown resource types
	return Tier5History
}

// resourceTypeFromPath extracts the resource type from a Git path.
// Paths follow the pattern: patients/<patient-id>/<resource-type>/<resource-id>.json
// or: .nucleus/<type>/<id>.json
func resourceTypeFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 && parts[0] == "patients" {
		// patients/<patient-id>/<resource-type>/<id>.json
		typePart := parts[2]
		return singularize(typePart)
	}
	return ""
}

func singularize(s string) string {
	mapping := map[string]string{
		"encounters":           "Encounter",
		"observations":         "Observation",
		"conditions":           "Condition",
		"medication_requests":  "MedicationRequest",
		"allergy_intolerances": "AllergyIntolerance",
		"flags":                "Flag",
		"detected_issues":      "DetectedIssue",
		"supply_deliveries":    "SupplyDelivery",
	}
	if rt, ok := mapping[s]; ok {
		return rt
	}
	// The path segment might be the Patient.json itself
	if strings.HasSuffix(s, ".json") {
		return "Patient"
	}
	return s
}

func extractClinicalStatus(resource map[string]any) string {
	cs, ok := resource["clinicalStatus"].(map[string]any)
	if !ok {
		return ""
	}
	coding, ok := cs["coding"].([]any)
	if !ok || len(coding) == 0 {
		return ""
	}
	first, ok := coding[0].(map[string]any)
	if !ok {
		return ""
	}
	code, _ := first["code"].(string)
	return code
}
