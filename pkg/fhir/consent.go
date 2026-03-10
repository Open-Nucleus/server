package fhir

import (
	"encoding/json"
	"fmt"
)

// Consent status values per FHIR R4.
const (
	ConsentStatusDraft    = "draft"
	ConsentStatusProposed = "proposed"
	ConsentStatusActive   = "active"
	ConsentStatusRejected = "rejected"
	ConsentStatusInactive = "inactive"
)

// Consent scope codes.
const (
	ConsentScopePatientPrivacy = "patient-privacy"
	ConsentScopeTreatment      = "treatment"
	ConsentScopeResearch       = "research"
)

// Consent provision types.
const (
	ConsentProvisionPermit = "permit"
	ConsentProvisionDeny   = "deny"
)

// Consent category codes.
const (
	ConsentCategoryACD      = "acd"
	ConsentCategoryDNR      = "dnr"
	ConsentCategoryEmrgOnly = "emrgonly"
	ConsentCategoryHCD      = "hcd"
	ConsentCategoryNPP      = "npp"
	ConsentCategoryPOLST    = "polst"
	ConsentCategoryResearch = "research"
	ConsentCategoryRSRID    = "rsdid"
	ConsentCategoryRSREID   = "rsreid"
)

// ExtractConsentFields extracts SQLite indexed columns from a Consent FHIR resource.
func ExtractConsentFields(fhirJSON []byte, gitBlobHash string) (*ConsentRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &ConsentRow{
		ID:          getStr(r, "id"),
		GitBlobHash: gitBlobHash,
	}

	// status
	row.Status = getStr(r, "status")

	// scope.coding[0].code
	if scope, ok := r["scope"].(map[string]any); ok {
		if codings, ok := getArray(scope, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.ScopeCode = getStr(c0, "code")
			}
		}
	}

	// patient.reference → patient ID
	if pat, ok := r["patient"].(map[string]any); ok {
		ref := getStr(pat, "reference")
		row.PatientID = stripPatientPrefix(ref)
	}

	// performer[0].reference → performer ID (device/practitioner granted access)
	if perfs, ok := getArray(r, "performer"); ok && len(perfs) > 0 {
		if perf0, ok := perfs[0].(map[string]any); ok {
			row.PerformerID = getStr(perf0, "reference")
		}
	}

	// provision.type
	if prov, ok := r["provision"].(map[string]any); ok {
		row.ProvisionType = getStr(prov, "type")

		// provision.period
		if period, ok := prov["period"].(map[string]any); ok {
			if start := getStr(period, "start"); start != "" {
				row.PeriodStart = &start
			}
			if end := getStr(period, "end"); end != "" {
				row.PeriodEnd = &end
			}
		}
	}

	// category[0].coding[0].code
	if cats, ok := getArray(r, "category"); ok && len(cats) > 0 {
		if cat0, ok := cats[0].(map[string]any); ok {
			if codings, ok := getArray(cat0, "coding"); ok && len(codings) > 0 {
				if c0, ok := codings[0].(map[string]any); ok {
					cat := getStr(c0, "code")
					row.Category = &cat
				}
			}
		}
	}

	// meta.lastUpdated
	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

func validateConsent(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "Consent.status is required"})
	} else if !isValidConsentStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: draft, proposed, active, rejected, inactive"})
	}

	if !hasCodeableConcept(r, "scope") {
		errs = append(errs, FieldError{Path: "scope", Rule: "required", Message: "Consent.scope is required"})
	}

	if !hasReference(r, "patient") {
		errs = append(errs, FieldError{Path: "patient", Rule: "required", Message: "Consent.patient reference is required"})
	}

	perfs, ok := getArray(r, "performer")
	if !ok || len(perfs) == 0 {
		errs = append(errs, FieldError{Path: "performer", Rule: "required", Message: "Consent.performer is required (at least one)"})
	}

	prov, ok := r["provision"].(map[string]any)
	if !ok {
		errs = append(errs, FieldError{Path: "provision", Rule: "required", Message: "Consent.provision is required"})
	} else {
		pt := getStr(prov, "type")
		if pt == "" {
			errs = append(errs, FieldError{Path: "provision.type", Rule: "required", Message: "Consent.provision.type is required"})
		} else if pt != ConsentProvisionPermit && pt != ConsentProvisionDeny {
			errs = append(errs, FieldError{Path: "provision.type", Rule: "value_set", Message: "Must be one of: permit, deny"})
		}
	}

	return errs
}

func isValidConsentStatus(s string) bool {
	switch s {
	case ConsentStatusDraft, ConsentStatusProposed, ConsentStatusActive, ConsentStatusRejected, ConsentStatusInactive:
		return true
	}
	return false
}

// IsConsentActive returns true if the consent status is "active".
func IsConsentActive(status string) bool {
	return status == ConsentStatusActive
}
