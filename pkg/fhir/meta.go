package fhir

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SetMeta updates the meta.lastUpdated, meta.versionId, and meta.source fields on a FHIR resource.
func SetMeta(fhirJSON []byte, lastUpdated time.Time, versionID, source string) ([]byte, error) {
	var resource map[string]any
	if err := json.Unmarshal(fhirJSON, &resource); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	meta, ok := resource["meta"].(map[string]any)
	if !ok {
		meta = make(map[string]any)
	}

	meta["lastUpdated"] = lastUpdated.UTC().Format(time.RFC3339)
	meta["versionId"] = versionID
	if source != "" {
		meta["source"] = source
	}
	resource["meta"] = meta

	out, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return out, nil
}

// AssignID assigns a UUID to the resource if the id field is absent.
// Returns the updated JSON and the resource ID (existing or newly assigned).
func AssignID(fhirJSON []byte) ([]byte, string, error) {
	var resource map[string]any
	if err := json.Unmarshal(fhirJSON, &resource); err != nil {
		return nil, "", fmt.Errorf("invalid JSON: %w", err)
	}

	idVal, ok := resource["id"]
	if ok {
		if idStr, isStr := idVal.(string); isStr && idStr != "" {
			return fhirJSON, idStr, nil
		}
	}

	newID := uuid.New().String()
	resource["id"] = newID

	out, err := json.Marshal(resource)
	if err != nil {
		return nil, "", fmt.Errorf("marshal: %w", err)
	}
	return out, newID, nil
}

// GetResourceType reads the resourceType field from a FHIR JSON resource.
func GetResourceType(fhirJSON []byte) (string, error) {
	var partial struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(fhirJSON, &partial); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	if partial.ResourceType == "" {
		return "", fmt.Errorf("missing resourceType field")
	}
	return partial.ResourceType, nil
}

// GetID reads the id field from a FHIR JSON resource.
func GetID(fhirJSON []byte) (string, error) {
	var partial struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(fhirJSON, &partial); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	if partial.ID == "" {
		return "", fmt.Errorf("missing id field")
	}
	return partial.ID, nil
}
