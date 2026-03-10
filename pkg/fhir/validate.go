package fhir

import (
	"encoding/json"
	"fmt"
)

// Validate performs Layer 1 structural validation on a FHIR resource per spec §4.1.
// Returns nil if valid, or a slice of field errors.
func Validate(resourceType string, fhirJSON []byte) []FieldError {
	var resource map[string]any
	if err := json.Unmarshal(fhirJSON, &resource); err != nil {
		return []FieldError{{Path: "", Rule: "json", Message: "Invalid JSON: " + err.Error()}}
	}

	rt, _ := resource["resourceType"].(string)
	if rt == "" {
		return []FieldError{{Path: "resourceType", Rule: "required", Message: "resourceType field is required"}}
	}
	if rt != resourceType {
		return []FieldError{{Path: "resourceType", Rule: "value", Message: fmt.Sprintf("Expected resourceType %q, got %q", resourceType, rt)}}
	}

	switch resourceType {
	case ResourcePatient:
		return validatePatient(resource)
	case ResourceEncounter:
		return validateEncounter(resource)
	case ResourceObservation:
		return validateObservation(resource)
	case ResourceCondition:
		return validateCondition(resource)
	case ResourceMedicationRequest:
		return validateMedicationRequest(resource)
	case ResourceAllergyIntolerance:
		return validateAllergyIntolerance(resource)
	case ResourceFlag:
		return validateFlag(resource)
	case ResourceImmunization:
		return validateImmunization(resource)
	case ResourceProcedure:
		return validateProcedure(resource)
	case ResourcePractitioner:
		return validatePractitioner(resource)
	case ResourceOrganization:
		return validateOrganization(resource)
	case ResourceLocation:
		return validateLocation(resource)
	case ResourceMeasureReport:
		return validateMeasureReport(resource)
	case ResourceConsent:
		return validateConsent(resource)
	default:
		return nil
	}
}

// ValidateWithProfile performs base validation then profile-specific validation
// for any profiles declared in meta.profile.
func ValidateWithProfile(resourceType string, fhirJSON []byte) []FieldError {
	errs := Validate(resourceType, fhirJSON)
	if len(errs) > 0 {
		return errs
	}

	var resource map[string]any
	if err := json.Unmarshal(fhirJSON, &resource); err != nil {
		return errs // base validation already passed, shouldn't happen
	}

	for _, profileURL := range GetMetaProfiles(resource) {
		def := GetProfileDef(profileURL)
		if def == nil {
			continue
		}
		if def.BaseResource != resourceType {
			errs = append(errs, FieldError{
				Path:    "meta.profile",
				Rule:    "profile_mismatch",
				Message: fmt.Sprintf("Profile %s targets %s but resource is %s", profileURL, def.BaseResource, resourceType),
			})
			continue
		}
		errs = append(errs, ValidateExtensions(resource, def.Extensions)...)
		if def.ValidateFunc != nil {
			errs = append(errs, def.ValidateFunc(resource)...)
		}
	}
	return errs
}

func validatePatient(r map[string]any) []FieldError {
	var errs []FieldError

	names, ok := getArray(r, "name")
	if !ok || len(names) == 0 {
		errs = append(errs, FieldError{Path: "name", Rule: "required", Message: "Patient.name is required (at least one name)"})
	} else {
		name0, ok := names[0].(map[string]any)
		if !ok {
			errs = append(errs, FieldError{Path: "name[0]", Rule: "type", Message: "name entry must be an object"})
		} else {
			if getStr(name0, "family") == "" {
				errs = append(errs, FieldError{Path: "name[0].family", Rule: "required", Message: "Patient.name[0].family is required"})
			}
			given, _ := getArray(name0, "given")
			if len(given) == 0 {
				errs = append(errs, FieldError{Path: "name[0].given", Rule: "required", Message: "Patient.name[0].given is required (at least one given name)"})
			}
		}
	}

	gender := getStr(r, "gender")
	if gender == "" {
		errs = append(errs, FieldError{Path: "gender", Rule: "required", Message: "Patient.gender is required"})
	} else if !isValidGender(gender) {
		errs = append(errs, FieldError{Path: "gender", Rule: "value_set", Message: "Must be one of: male, female, other, unknown"})
	}

	if getStr(r, "birthDate") == "" {
		errs = append(errs, FieldError{Path: "birthDate", Rule: "required", Message: "Patient.birthDate is required"})
	}

	return errs
}

func validateEncounter(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "Encounter.status is required"})
	} else if !isValidEncounterStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: planned, arrived, triaged, in-progress, onleave, finished, cancelled, entered-in-error"})
	}

	classObj, ok := r["class"].(map[string]any)
	if !ok {
		errs = append(errs, FieldError{Path: "class", Rule: "required", Message: "Encounter.class is required"})
	} else if getStr(classObj, "code") == "" {
		errs = append(errs, FieldError{Path: "class.code", Rule: "required", Message: "Encounter.class.code is required"})
	}

	if !hasReference(r, "subject") {
		errs = append(errs, FieldError{Path: "subject", Rule: "required", Message: "Encounter.subject reference is required"})
	}

	period, ok := r["period"].(map[string]any)
	if !ok {
		errs = append(errs, FieldError{Path: "period", Rule: "required", Message: "Encounter.period is required"})
	} else if getStr(period, "start") == "" {
		errs = append(errs, FieldError{Path: "period.start", Rule: "required", Message: "Encounter.period.start is required"})
	}

	return errs
}

func validateObservation(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "Observation.status is required"})
	} else if !isValidObservationStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: registered, preliminary, final, amended, corrected, cancelled, entered-in-error"})
	}

	if !hasCodeableConcept(r, "code") {
		errs = append(errs, FieldError{Path: "code", Rule: "required", Message: "Observation.code is required"})
	}

	if !hasReference(r, "subject") {
		errs = append(errs, FieldError{Path: "subject", Rule: "required", Message: "Observation.subject reference is required"})
	}

	if getStr(r, "effectiveDateTime") == "" {
		errs = append(errs, FieldError{Path: "effectiveDateTime", Rule: "required", Message: "Observation.effectiveDateTime is required"})
	}

	return errs
}

func validateCondition(r map[string]any) []FieldError {
	var errs []FieldError

	if !hasCodeableConcept(r, "clinicalStatus") {
		errs = append(errs, FieldError{Path: "clinicalStatus", Rule: "required", Message: "Condition.clinicalStatus is required"})
	}

	if !hasCodeableConcept(r, "verificationStatus") {
		errs = append(errs, FieldError{Path: "verificationStatus", Rule: "required", Message: "Condition.verificationStatus is required"})
	}

	if !hasCodeableConcept(r, "code") {
		errs = append(errs, FieldError{Path: "code", Rule: "required", Message: "Condition.code is required"})
	}

	if !hasReference(r, "subject") {
		errs = append(errs, FieldError{Path: "subject", Rule: "required", Message: "Condition.subject reference is required"})
	}

	return errs
}

func validateMedicationRequest(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "MedicationRequest.status is required"})
	}

	intent := getStr(r, "intent")
	if intent == "" {
		errs = append(errs, FieldError{Path: "intent", Rule: "required", Message: "MedicationRequest.intent is required"})
	}

	if !hasCodeableConcept(r, "medicationCodeableConcept") {
		errs = append(errs, FieldError{Path: "medicationCodeableConcept", Rule: "required", Message: "MedicationRequest.medicationCodeableConcept is required"})
	}

	if !hasReference(r, "subject") {
		errs = append(errs, FieldError{Path: "subject", Rule: "required", Message: "MedicationRequest.subject reference is required"})
	}

	dosage, ok := getArray(r, "dosageInstruction")
	if !ok || len(dosage) == 0 {
		errs = append(errs, FieldError{Path: "dosageInstruction", Rule: "required", Message: "MedicationRequest.dosageInstruction is required (at least one)"})
	}

	return errs
}

func validateAllergyIntolerance(r map[string]any) []FieldError {
	var errs []FieldError

	if !hasCodeableConcept(r, "clinicalStatus") {
		errs = append(errs, FieldError{Path: "clinicalStatus", Rule: "required", Message: "AllergyIntolerance.clinicalStatus is required"})
	}

	if !hasCodeableConcept(r, "verificationStatus") {
		errs = append(errs, FieldError{Path: "verificationStatus", Rule: "required", Message: "AllergyIntolerance.verificationStatus is required"})
	}

	if !hasCodeableConcept(r, "code") {
		errs = append(errs, FieldError{Path: "code", Rule: "required", Message: "AllergyIntolerance.code is required"})
	}

	if !hasReference(r, "patient") {
		errs = append(errs, FieldError{Path: "patient", Rule: "required", Message: "AllergyIntolerance.patient reference is required"})
	}

	return errs
}

func validateFlag(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "Flag.status is required"})
	} else if !isValidFlagStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: active, inactive, entered-in-error"})
	}

	if !hasReference(r, "subject") {
		errs = append(errs, FieldError{Path: "subject", Rule: "required", Message: "Flag.subject reference is required"})
	}

	return errs
}

// helpers

func getStr(m map[string]any, key string) string {
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return v
}

func getArray(m map[string]any, key string) ([]any, bool) {
	v, ok := m[key].([]any)
	return v, ok
}

func hasReference(m map[string]any, key string) bool {
	ref, ok := m[key].(map[string]any)
	if !ok {
		return false
	}
	return getStr(ref, "reference") != ""
}

func hasCodeableConcept(m map[string]any, key string) bool {
	cc, ok := m[key].(map[string]any)
	if !ok {
		return false
	}
	// A CodeableConcept has coding[] or text
	if codings, ok := getArray(cc, "coding"); ok && len(codings) > 0 {
		return true
	}
	return getStr(cc, "text") != ""
}

func isValidGender(s string) bool {
	switch s {
	case "male", "female", "other", "unknown":
		return true
	}
	return false
}

func isValidEncounterStatus(s string) bool {
	switch s {
	case "planned", "arrived", "triaged", "in-progress", "onleave", "finished", "cancelled", "entered-in-error":
		return true
	}
	return false
}

func isValidObservationStatus(s string) bool {
	switch s {
	case "registered", "preliminary", "final", "amended", "corrected", "cancelled", "entered-in-error":
		return true
	}
	return false
}

func isValidFlagStatus(s string) bool {
	switch s {
	case "active", "inactive", "entered-in-error":
		return true
	}
	return false
}

func validateImmunization(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "Immunization.status is required"})
	} else if !isValidImmunizationStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: completed, entered-in-error, not-done"})
	}

	if !hasCodeableConcept(r, "vaccineCode") {
		errs = append(errs, FieldError{Path: "vaccineCode", Rule: "required", Message: "Immunization.vaccineCode is required"})
	}

	if !hasReference(r, "patient") {
		errs = append(errs, FieldError{Path: "patient", Rule: "required", Message: "Immunization.patient reference is required"})
	}

	if getStr(r, "occurrenceDateTime") == "" {
		errs = append(errs, FieldError{Path: "occurrenceDateTime", Rule: "required", Message: "Immunization.occurrenceDateTime is required"})
	}

	return errs
}

func validateProcedure(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "Procedure.status is required"})
	} else if !isValidProcedureStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: preparation, in-progress, not-done, on-hold, stopped, completed, entered-in-error, unknown"})
	}

	if !hasCodeableConcept(r, "code") {
		errs = append(errs, FieldError{Path: "code", Rule: "required", Message: "Procedure.code is required"})
	}

	if !hasReference(r, "subject") {
		errs = append(errs, FieldError{Path: "subject", Rule: "required", Message: "Procedure.subject reference is required"})
	}

	return errs
}

func validatePractitioner(r map[string]any) []FieldError {
	var errs []FieldError

	names, ok := getArray(r, "name")
	if !ok || len(names) == 0 {
		errs = append(errs, FieldError{Path: "name", Rule: "required", Message: "Practitioner.name is required (at least one name)"})
	} else {
		name0, ok := names[0].(map[string]any)
		if !ok {
			errs = append(errs, FieldError{Path: "name[0]", Rule: "type", Message: "name entry must be an object"})
		} else if getStr(name0, "family") == "" {
			errs = append(errs, FieldError{Path: "name[0].family", Rule: "required", Message: "Practitioner.name[0].family is required"})
		}
	}

	return errs
}

func validateOrganization(r map[string]any) []FieldError {
	var errs []FieldError

	if getStr(r, "name") == "" {
		errs = append(errs, FieldError{Path: "name", Rule: "required", Message: "Organization.name is required"})
	}

	return errs
}

func validateLocation(r map[string]any) []FieldError {
	var errs []FieldError

	if getStr(r, "name") == "" {
		errs = append(errs, FieldError{Path: "name", Rule: "required", Message: "Location.name is required"})
	}

	status := getStr(r, "status")
	if status != "" && !isValidLocationStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: active, suspended, inactive"})
	}

	return errs
}

func isValidImmunizationStatus(s string) bool {
	switch s {
	case "completed", "entered-in-error", "not-done":
		return true
	}
	return false
}

func isValidProcedureStatus(s string) bool {
	switch s {
	case "preparation", "in-progress", "not-done", "on-hold", "stopped", "completed", "entered-in-error", "unknown":
		return true
	}
	return false
}

func isValidLocationStatus(s string) bool {
	switch s {
	case "active", "suspended", "inactive":
		return true
	}
	return false
}

func validateMeasureReport(r map[string]any) []FieldError {
	var errs []FieldError

	status := getStr(r, "status")
	if status == "" {
		errs = append(errs, FieldError{Path: "status", Rule: "required", Message: "MeasureReport.status is required"})
	} else if !isValidMeasureReportStatus(status) {
		errs = append(errs, FieldError{Path: "status", Rule: "value_set", Message: "Must be one of: complete, pending, error"})
	}

	mrType := getStr(r, "type")
	if mrType == "" {
		errs = append(errs, FieldError{Path: "type", Rule: "required", Message: "MeasureReport.type is required"})
	} else if !isValidMeasureReportType(mrType) {
		errs = append(errs, FieldError{Path: "type", Rule: "value_set", Message: "Must be one of: individual, subject-list, summary, data-collection"})
	}

	period, ok := r["period"].(map[string]any)
	if !ok {
		errs = append(errs, FieldError{Path: "period", Rule: "required", Message: "MeasureReport.period is required"})
	} else if getStr(period, "start") == "" {
		errs = append(errs, FieldError{Path: "period.start", Rule: "required", Message: "MeasureReport.period.start is required"})
	}

	return errs
}

func isValidMeasureReportStatus(s string) bool {
	switch s {
	case "complete", "pending", "error":
		return true
	}
	return false
}

func isValidMeasureReportType(s string) bool {
	switch s {
	case "individual", "subject-list", "summary", "data-collection":
		return true
	}
	return false
}
