package fhir

import (
	"encoding/json"
	"fmt"
)

// ApplySoftDelete modifies a FHIR resource JSON to mark it as soft-deleted per spec §3.4.
func ApplySoftDelete(resourceType string, fhirJSON []byte) ([]byte, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	switch resourceType {
	case ResourcePatient:
		r["active"] = false

	case ResourceEncounter:
		r["status"] = "entered-in-error"

	case ResourceCondition:
		r["clinicalStatus"] = codeableConcept("inactive")
		r["verificationStatus"] = codeableConcept("entered-in-error")

	case ResourceMedicationRequest:
		r["status"] = "entered-in-error"

	case ResourceAllergyIntolerance:
		r["verificationStatus"] = codeableConcept("entered-in-error")

	case ResourceFlag:
		r["status"] = "entered-in-error"

	default:
		return nil, fmt.Errorf("unsupported resource type for soft delete: %s", resourceType)
	}

	out, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return out, nil
}

func codeableConcept(code string) map[string]any {
	return map[string]any{
		"coding": []any{
			map[string]any{
				"code": code,
			},
		},
	}
}
