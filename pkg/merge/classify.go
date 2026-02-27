package merge

import "encoding/json"

// FormularyChecker is an optional interface for drug interaction checking.
type FormularyChecker interface {
	HasInteraction(medCodeA, medCodeB string) bool
}

// Classifier determines the ConflictLevel for a merge conflict.
type Classifier struct {
	Formulary FormularyChecker // optional, nil-safe
}

// Classify determines how a conflict between local and remote versions should be handled.
func (c *Classifier) Classify(resourceType string, local, remote, base json.RawMessage) ConflictResult {
	diffs, err := DiffResourcesWithBase(base, local, remote)
	if err != nil {
		return ConflictResult{Level: Review, Reason: "failed to diff resources"}
	}

	if len(diffs) == 0 {
		return ConflictResult{Level: AutoMerge, Reason: "no conflicting changes"}
	}

	changedFields := make([]string, len(diffs))
	for i, d := range diffs {
		changedFields[i] = d.Path
	}

	overlapping := OverlappingFields(diffs)

	// Check Block conditions first
	if reason := c.checkBlockConditions(resourceType, overlapping, local, remote); reason != "" {
		return ConflictResult{
			Level:         Block,
			Reason:        reason,
			ChangedFields: changedFields,
			FieldDiffs:    diffs,
		}
	}

	// If no overlapping changes, auto-merge is safe
	if len(overlapping) == 0 {
		return ConflictResult{
			Level:         AutoMerge,
			Reason:        "non-overlapping changes",
			ChangedFields: changedFields,
			FieldDiffs:    diffs,
		}
	}

	// Check if overlapping changes are only in non-clinical fields
	if c.allNonClinical(resourceType, overlapping) {
		return ConflictResult{
			Level:         Review,
			Reason:        "overlapping non-clinical fields",
			ChangedFields: changedFields,
			FieldDiffs:    diffs,
		}
	}

	return ConflictResult{
		Level:         Review,
		Reason:        "overlapping clinical fields require review",
		ChangedFields: changedFields,
		FieldDiffs:    diffs,
	}
}

// checkBlockConditions returns a reason string if the conflict must be blocked.
func (c *Classifier) checkBlockConditions(resourceType string, overlapping []FieldDiff, local, remote json.RawMessage) string {
	switch resourceType {
	case "AllergyIntolerance":
		for _, d := range overlapping {
			if d.Path == "criticality" {
				return "conflicting allergy criticality changes"
			}
		}
	case "MedicationRequest":
		if c.Formulary != nil {
			if c.hasDrugInteraction(local, remote) {
				return "potential drug interaction conflict"
			}
		}
		for _, d := range overlapping {
			if d.Path == "medicationCodeableConcept" || d.Path == "dosageInstruction" {
				return "conflicting medication changes"
			}
		}
	case "Condition":
		for _, d := range overlapping {
			if d.Path == "code" || d.Path == "clinicalStatus" || d.Path == "verificationStatus" {
				return "conflicting diagnosis changes"
			}
		}
	case "Patient":
		for _, d := range overlapping {
			if d.Path == "identifier" || d.Path == "name" || d.Path == "birthDate" || d.Path == "gender" {
				return "conflicting patient identity fields"
			}
		}
	case "Observation":
		for _, d := range overlapping {
			if d.Path == "valueQuantity" || d.Path == "valueCodeableConcept" || d.Path == "interpretation" {
				return "contradictory vital signs or lab values"
			}
		}
	}
	return ""
}

func (c *Classifier) hasDrugInteraction(local, remote json.RawMessage) bool {
	localCode := extractMedCode(local)
	remoteCode := extractMedCode(remote)
	if localCode == "" || remoteCode == "" || localCode == remoteCode {
		return false
	}
	return c.Formulary.HasInteraction(localCode, remoteCode)
}

func extractMedCode(data json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if med, ok := m["medicationCodeableConcept"].(map[string]any); ok {
		if coding, ok := med["coding"].([]any); ok && len(coding) > 0 {
			if c, ok := coding[0].(map[string]any); ok {
				if code, ok := c["code"].(string); ok {
					return code
				}
			}
		}
	}
	return ""
}

// allNonClinical returns true if all overlapping diffs are in non-clinical fields.
func (c *Classifier) allNonClinical(resourceType string, overlapping []FieldDiff) bool {
	clinical := clinicalFields[resourceType]
	if clinical == nil {
		return false // unknown resource type, assume clinical
	}
	for _, d := range overlapping {
		if clinical[d.Path] {
			return false
		}
	}
	return true
}

// clinicalFields maps resource types to their clinically significant fields.
var clinicalFields = map[string]map[string]bool{
	"Patient": {
		"identifier": true, "name": true, "birthDate": true, "gender": true,
		"deceasedBoolean": true, "deceasedDateTime": true,
	},
	"Encounter": {
		"status": true, "class": true, "type": true, "period": true,
		"diagnosis": true, "hospitalization": true,
	},
	"Observation": {
		"status": true, "code": true, "valueQuantity": true,
		"valueCodeableConcept": true, "interpretation": true,
		"referenceRange": true, "component": true,
	},
	"Condition": {
		"code": true, "clinicalStatus": true, "verificationStatus": true,
		"severity": true, "bodySite": true, "stage": true,
	},
	"MedicationRequest": {
		"status": true, "medicationCodeableConcept": true,
		"dosageInstruction": true, "dispenseRequest": true,
	},
	"AllergyIntolerance": {
		"code": true, "clinicalStatus": true, "verificationStatus": true,
		"criticality": true, "type": true, "reaction": true,
	},
}
