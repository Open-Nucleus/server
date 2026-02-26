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
	default:
		return nil
	}
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
