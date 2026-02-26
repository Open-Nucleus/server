package fhir

import "fmt"

// GitPath returns the Git repository file path for a FHIR resource per spec §3.3.
func GitPath(resourceType, patientID, resourceID string) string {
	switch resourceType {
	case ResourcePatient:
		return fmt.Sprintf("patients/%s/Patient.json", resourceID)
	case ResourceEncounter:
		return fmt.Sprintf("patients/%s/encounters/%s.json", patientID, resourceID)
	case ResourceObservation:
		return fmt.Sprintf("patients/%s/observations/%s.json", patientID, resourceID)
	case ResourceCondition:
		return fmt.Sprintf("patients/%s/conditions/%s.json", patientID, resourceID)
	case ResourceMedicationRequest:
		return fmt.Sprintf("patients/%s/medication-requests/%s.json", patientID, resourceID)
	case ResourceAllergyIntolerance:
		return fmt.Sprintf("patients/%s/allergy-intolerances/%s.json", patientID, resourceID)
	case ResourceFlag:
		return fmt.Sprintf("patients/%s/flags/%s.json", patientID, resourceID)
	case ResourceDetectedIssue:
		return fmt.Sprintf("alerts/%s.json", resourceID)
	case ResourceSupplyDelivery:
		return fmt.Sprintf("supply/deliveries/%s.json", resourceID)
	default:
		return fmt.Sprintf("unknown/%s/%s.json", resourceType, resourceID)
	}
}

// PatientDirPath returns the Git directory path for a patient's folder.
func PatientDirPath(patientID string) string {
	return fmt.Sprintf("patients/%s/", patientID)
}
