package fhir

import (
	"encoding/json"
	"fmt"
)

// ExtractPatientFields extracts SQLite indexed columns from a Patient FHIR resource.
func ExtractPatientFields(fhirJSON []byte, siteID, gitBlobHash string) (*PatientRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &PatientRow{
		ID:          getStr(r, "id"),
		Gender:      getStr(r, "gender"),
		BirthDate:   getStr(r, "birthDate"),
		SiteID:      siteID,
		Active:      true,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	// active field
	if active, ok := r["active"].(bool); ok {
		row.Active = active
	}

	// Extract name
	if names, ok := getArray(r, "name"); ok && len(names) > 0 {
		if name0, ok := names[0].(map[string]any); ok {
			row.FamilyName = getStr(name0, "family")
			if given, ok := getArray(name0, "given"); ok {
				givenBytes, _ := json.Marshal(given)
				row.GivenNames = string(givenBytes)
			}
		}
	}

	// last updated from meta
	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractEncounterFields extracts SQLite indexed columns from an Encounter.
func ExtractEncounterFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*EncounterRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &EncounterRow{
		ID:          getStr(r, "id"),
		PatientID:   patientID,
		Status:      getStr(r, "status"),
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if classObj, ok := r["class"].(map[string]any); ok {
		row.ClassCode = getStr(classObj, "code")
	}

	if period, ok := r["period"].(map[string]any); ok {
		row.PeriodStart = getStr(period, "start")
		if end := getStr(period, "end"); end != "" {
			row.PeriodEnd = &end
		}
	}

	if types, ok := getArray(r, "type"); ok && len(types) > 0 {
		if t0, ok := types[0].(map[string]any); ok {
			if codings, ok := getArray(t0, "coding"); ok && len(codings) > 0 {
				if c0, ok := codings[0].(map[string]any); ok {
					tc := getStr(c0, "code")
					row.TypeCode = &tc
				}
			}
		}
	}

	if reason, ok := getArray(r, "reasonCode"); ok && len(reason) > 0 {
		if r0, ok := reason[0].(map[string]any); ok {
			if codings, ok := getArray(r0, "coding"); ok && len(codings) > 0 {
				if c0, ok := codings[0].(map[string]any); ok {
					rc := getStr(c0, "code")
					row.ReasonCode = &rc
				}
			}
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractObservationFields extracts SQLite indexed columns from an Observation.
func ExtractObservationFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*ObservationRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &ObservationRow{
		ID:                getStr(r, "id"),
		PatientID:         patientID,
		Status:            getStr(r, "status"),
		EffectiveDatetime: getStr(r, "effectiveDateTime"),
		SiteID:            siteID,
		GitBlobHash:       gitBlobHash,
		FHIRJson:          string(fhirJSON),
	}

	// encounter reference
	if enc, ok := r["encounter"].(map[string]any); ok {
		ref := getStr(enc, "reference")
		if ref != "" {
			row.EncounterID = &ref
		}
	}

	// category
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

	// code
	if code, ok := r["code"].(map[string]any); ok {
		if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.Code = getStr(c0, "code")
				disp := getStr(c0, "display")
				if disp != "" {
					row.CodeDisplay = &disp
				}
			}
		}
	}

	// valueQuantity
	if vq, ok := r["valueQuantity"].(map[string]any); ok {
		if val, ok := vq["value"].(float64); ok {
			row.ValueQuantityValue = &val
		}
		unit := getStr(vq, "unit")
		if unit != "" {
			row.ValueQuantityUnit = &unit
		}
	}

	// valueString
	if vs := getStr(r, "valueString"); vs != "" {
		row.ValueString = &vs
	}

	// valueCodeableConcept
	if vcc, ok := r["valueCodeableConcept"].(map[string]any); ok {
		vccBytes, _ := json.Marshal(vcc)
		vccStr := string(vccBytes)
		row.ValueCodeableConcept = &vccStr
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractConditionFields extracts SQLite indexed columns from a Condition.
func ExtractConditionFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*ConditionRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &ConditionRow{
		ID:        getStr(r, "id"),
		PatientID: patientID,
		SiteID:    siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	row.ClinicalStatus = extractCodeFromCodeableConcept(r, "clinicalStatus")
	row.VerificationStatus = extractCodeFromCodeableConcept(r, "verificationStatus")

	if code, ok := r["code"].(map[string]any); ok {
		if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.Code = getStr(c0, "code")
				disp := getStr(c0, "display")
				if disp != "" {
					row.CodeDisplay = &disp
				}
			}
		}
	}

	if onset := getStr(r, "onsetDateTime"); onset != "" {
		row.OnsetDatetime = &onset
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractMedicationRequestFields extracts SQLite indexed columns from a MedicationRequest.
func ExtractMedicationRequestFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*MedicationRequestRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &MedicationRequestRow{
		ID:        getStr(r, "id"),
		PatientID: patientID,
		Status:    getStr(r, "status"),
		Intent:    getStr(r, "intent"),
		SiteID:    siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if medCC, ok := r["medicationCodeableConcept"].(map[string]any); ok {
		if codings, ok := getArray(medCC, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.MedicationCode = getStr(c0, "code")
				disp := getStr(c0, "display")
				if disp != "" {
					row.MedicationDisplay = &disp
				}
			}
		}
	}

	if ao := getStr(r, "authoredOn"); ao != "" {
		row.AuthoredOn = &ao
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractAllergyIntoleranceFields extracts SQLite indexed columns from an AllergyIntolerance.
func ExtractAllergyIntoleranceFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*AllergyIntoleranceRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &AllergyIntoleranceRow{
		ID:        getStr(r, "id"),
		PatientID: patientID,
		SiteID:    siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	row.ClinicalStatus = extractCodeFromCodeableConcept(r, "clinicalStatus")
	row.VerificationStatus = extractCodeFromCodeableConcept(r, "verificationStatus")

	if t := getStr(r, "type"); t != "" {
		row.Type = &t
	}

	if code, ok := r["code"].(map[string]any); ok {
		if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.SubstanceCode = getStr(c0, "code")
				disp := getStr(c0, "display")
				if disp != "" {
					row.SubstanceDisplay = &disp
				}
			}
		}
	}

	if crit := getStr(r, "criticality"); crit != "" {
		row.Criticality = &crit
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractFlagFields extracts SQLite indexed columns from a Flag.
func ExtractFlagFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*FlagRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &FlagRow{
		ID:          getStr(r, "id"),
		PatientID:   patientID,
		Status:      getStr(r, "status"),
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

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

	if code, ok := r["code"].(map[string]any); ok {
		if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				c := getStr(c0, "code")
				row.Code = &c
			}
		}
	}

	if period, ok := r["period"].(map[string]any); ok {
		if start := getStr(period, "start"); start != "" {
			row.PeriodStart = &start
		}
		if end := getStr(period, "end"); end != "" {
			row.PeriodEnd = &end
		}
	}

	// Sentinel-generated flags may have extension with generated_by
	if exts, ok := getArray(r, "extension"); ok {
		for _, ext := range exts {
			if extMap, ok := ext.(map[string]any); ok {
				if getStr(extMap, "url") == "http://open-nucleus.org/fhir/StructureDefinition/generated-by" {
					gen := getStr(extMap, "valueString")
					row.GeneratedBy = &gen
				}
			}
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractImmunizationFields extracts SQLite indexed columns from an Immunization.
func ExtractImmunizationFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*ImmunizationRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &ImmunizationRow{
		ID:                 getStr(r, "id"),
		PatientID:          patientID,
		Status:             getStr(r, "status"),
		OccurrenceDatetime: getStr(r, "occurrenceDateTime"),
		SiteID:             siteID,
		GitBlobHash:        gitBlobHash,
		FHIRJson:           string(fhirJSON),
	}

	if vc, ok := r["vaccineCode"].(map[string]any); ok {
		if codings, ok := getArray(vc, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.VaccineCode = getStr(c0, "code")
				disp := getStr(c0, "display")
				if disp != "" {
					row.VaccineDisplay = &disp
				}
			}
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractProcedureFields extracts SQLite indexed columns from a Procedure.
func ExtractProcedureFields(fhirJSON []byte, patientID, siteID, gitBlobHash string) (*ProcedureRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &ProcedureRow{
		ID:          getStr(r, "id"),
		PatientID:   patientID,
		Status:      getStr(r, "status"),
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if code, ok := r["code"].(map[string]any); ok {
		if codings, ok := getArray(code, "coding"); ok && len(codings) > 0 {
			if c0, ok := codings[0].(map[string]any); ok {
				row.Code = getStr(c0, "code")
				disp := getStr(c0, "display")
				if disp != "" {
					row.CodeDisplay = &disp
				}
			}
		}
	}

	if pd := getStr(r, "performedDateTime"); pd != "" {
		row.PerformedDatetime = &pd
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractPractitionerFields extracts SQLite indexed columns from a Practitioner.
func ExtractPractitionerFields(fhirJSON []byte, siteID, gitBlobHash string) (*PractitionerRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &PractitionerRow{
		ID:          getStr(r, "id"),
		Active:      true,
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if active, ok := r["active"].(bool); ok {
		row.Active = active
	}

	if names, ok := getArray(r, "name"); ok && len(names) > 0 {
		if name0, ok := names[0].(map[string]any); ok {
			row.FamilyName = getStr(name0, "family")
			if given, ok := getArray(name0, "given"); ok {
				givenBytes, _ := json.Marshal(given)
				row.GivenNames = string(givenBytes)
			}
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractOrganizationFields extracts SQLite indexed columns from an Organization.
func ExtractOrganizationFields(fhirJSON []byte, siteID, gitBlobHash string) (*OrganizationRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &OrganizationRow{
		ID:          getStr(r, "id"),
		Name:        getStr(r, "name"),
		Active:      true,
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if active, ok := r["active"].(bool); ok {
		row.Active = active
	}

	if types, ok := getArray(r, "type"); ok && len(types) > 0 {
		if t0, ok := types[0].(map[string]any); ok {
			if codings, ok := getArray(t0, "coding"); ok && len(codings) > 0 {
				if c0, ok := codings[0].(map[string]any); ok {
					tc := getStr(c0, "code")
					row.Type = &tc
				}
			}
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractLocationFields extracts SQLite indexed columns from a Location.
func ExtractLocationFields(fhirJSON []byte, siteID, gitBlobHash string) (*LocationRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &LocationRow{
		ID:          getStr(r, "id"),
		Name:        getStr(r, "name"),
		Status:      getStr(r, "status"),
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if types, ok := getArray(r, "type"); ok && len(types) > 0 {
		if t0, ok := types[0].(map[string]any); ok {
			if codings, ok := getArray(t0, "coding"); ok && len(codings) > 0 {
				if c0, ok := codings[0].(map[string]any); ok {
					tc := getStr(c0, "code")
					row.Type = &tc
				}
			}
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractMeasureReportFields extracts SQLite indexed columns from a MeasureReport.
func ExtractMeasureReportFields(fhirJSON []byte, siteID, gitBlobHash string) (*MeasureReportRow, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	row := &MeasureReportRow{
		ID:          getStr(r, "id"),
		Status:      getStr(r, "status"),
		Type:        getStr(r, "type"),
		SiteID:      siteID,
		GitBlobHash: gitBlobHash,
		FHIRJson:    string(fhirJSON),
	}

	if period, ok := r["period"].(map[string]any); ok {
		row.PeriodStart = getStr(period, "start")
		if end := getStr(period, "end"); end != "" {
			row.PeriodEnd = &end
		}
	}

	if reporter, ok := r["reporter"].(map[string]any); ok {
		if ref := getStr(reporter, "reference"); ref != "" {
			row.Reporter = &ref
		}
	}

	if meta, ok := r["meta"].(map[string]any); ok {
		row.LastUpdated = getStr(meta, "lastUpdated")
	}

	return row, nil
}

// ExtractPatientReference extracts the patient ID from a FHIR resource's
// subject.reference or patient.reference field. Returns ("", nil) if no
// patient reference is found.
func ExtractPatientReference(fhirJSON []byte) (string, error) {
	var r map[string]any
	if err := json.Unmarshal(fhirJSON, &r); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// Try subject.reference (Encounter, Observation, Condition, etc.)
	if subj, ok := r["subject"].(map[string]any); ok {
		if ref := getStr(subj, "reference"); ref != "" {
			return stripPatientPrefix(ref), nil
		}
	}

	// Try patient.reference (AllergyIntolerance, Flag, etc.)
	if pat, ok := r["patient"].(map[string]any); ok {
		if ref := getStr(pat, "reference"); ref != "" {
			return stripPatientPrefix(ref), nil
		}
	}

	return "", nil
}

// stripPatientPrefix removes "Patient/" prefix from a reference string.
func stripPatientPrefix(ref string) string {
	const prefix = "Patient/"
	if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
		return ref[len(prefix):]
	}
	return ref
}

func extractCodeFromCodeableConcept(r map[string]any, field string) string {
	cc, ok := r[field].(map[string]any)
	if !ok {
		return ""
	}
	if codings, ok := getArray(cc, "coding"); ok && len(codings) > 0 {
		if c0, ok := codings[0].(map[string]any); ok {
			return getStr(c0, "code")
		}
	}
	return getStr(cc, "text")
}
