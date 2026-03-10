package fhir

// Resource type constants matching FHIR R4 resourceType values.
const (
	ResourcePatient              = "Patient"
	ResourceEncounter            = "Encounter"
	ResourceObservation          = "Observation"
	ResourceCondition            = "Condition"
	ResourceMedicationRequest    = "MedicationRequest"
	ResourceAllergyIntolerance   = "AllergyIntolerance"
	ResourceFlag                 = "Flag"
	ResourceDetectedIssue        = "DetectedIssue"
	ResourceSupplyDelivery       = "SupplyDelivery"
	ResourceImmunization         = "Immunization"
	ResourceProcedure            = "Procedure"
	ResourcePractitioner         = "Practitioner"
	ResourceOrganization         = "Organization"
	ResourceLocation             = "Location"
	ResourceProvenance           = "Provenance"
	ResourceMeasureReport        = "MeasureReport"
	ResourceStructureDefinition  = "StructureDefinition"
)

// Operation constants for commit messages.
const (
	OpCreate = "CREATE"
	OpUpdate = "UPDATE"
	OpDelete = "DELETE"
)

// FieldError represents a single FHIR validation failure.
type FieldError struct {
	Path    string `json:"path"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// PatientRow holds indexed fields extracted from a Patient FHIR resource.
type PatientRow struct {
	ID          string `json:"id"`
	FamilyName  string `json:"family_name"`
	GivenNames  string `json:"given_names"` // JSON array as string
	Gender      string `json:"gender"`
	BirthDate   string `json:"birth_date"`
	SiteID      string `json:"site_id"`
	Active      bool   `json:"active"`
	LastUpdated string `json:"last_updated"`
	GitBlobHash string `json:"git_blob_hash"`
}

// EncounterRow holds indexed fields extracted from an Encounter FHIR resource.
type EncounterRow struct {
	ID          string  `json:"id"`
	PatientID   string  `json:"patient_id"`
	Status      string  `json:"status"`
	ClassCode   string  `json:"class_code"`
	TypeCode    *string `json:"type_code"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   *string `json:"period_end"`
	SiteID      string  `json:"site_id"`
	ReasonCode  *string `json:"reason_code"`
	LastUpdated string  `json:"last_updated"`
	GitBlobHash string  `json:"git_blob_hash"`
}

// ObservationRow holds indexed fields extracted from an Observation FHIR resource.
type ObservationRow struct {
	ID                    string   `json:"id"`
	PatientID             string   `json:"patient_id"`
	EncounterID           *string  `json:"encounter_id"`
	Status                string   `json:"status"`
	Category              *string  `json:"category"`
	Code                  string   `json:"code"`
	CodeDisplay           *string  `json:"code_display"`
	EffectiveDatetime     string   `json:"effective_datetime"`
	ValueQuantityValue    *float64 `json:"value_quantity_value"`
	ValueQuantityUnit     *string  `json:"value_quantity_unit"`
	ValueString           *string  `json:"value_string"`
	ValueCodeableConcept  *string  `json:"value_codeable_concept"` // JSON
	SiteID                string   `json:"site_id"`
	LastUpdated           string   `json:"last_updated"`
	GitBlobHash           string   `json:"git_blob_hash"`
}

// ConditionRow holds indexed fields extracted from a Condition FHIR resource.
type ConditionRow struct {
	ID                 string  `json:"id"`
	PatientID          string  `json:"patient_id"`
	ClinicalStatus     string  `json:"clinical_status"`
	VerificationStatus string  `json:"verification_status"`
	Code               string  `json:"code"`
	CodeDisplay        *string `json:"code_display"`
	OnsetDatetime      *string `json:"onset_datetime"`
	SiteID             string  `json:"site_id"`
	LastUpdated        string  `json:"last_updated"`
	GitBlobHash        string  `json:"git_blob_hash"`
}

// MedicationRequestRow holds indexed fields extracted from a MedicationRequest.
type MedicationRequestRow struct {
	ID                string  `json:"id"`
	PatientID         string  `json:"patient_id"`
	Status            string  `json:"status"`
	Intent            string  `json:"intent"`
	MedicationCode    string  `json:"medication_code"`
	MedicationDisplay *string `json:"medication_display"`
	AuthoredOn        *string `json:"authored_on"`
	SiteID            string  `json:"site_id"`
	LastUpdated       string  `json:"last_updated"`
	GitBlobHash       string  `json:"git_blob_hash"`
}

// AllergyIntoleranceRow holds indexed fields extracted from an AllergyIntolerance.
type AllergyIntoleranceRow struct {
	ID                 string  `json:"id"`
	PatientID          string  `json:"patient_id"`
	ClinicalStatus     string  `json:"clinical_status"`
	VerificationStatus string  `json:"verification_status"`
	Type               *string `json:"type"`
	SubstanceCode      string  `json:"substance_code"`
	SubstanceDisplay   *string `json:"substance_display"`
	Criticality        *string `json:"criticality"`
	SiteID             string  `json:"site_id"`
	LastUpdated        string  `json:"last_updated"`
	GitBlobHash        string  `json:"git_blob_hash"`
}

// FlagRow holds indexed fields extracted from a Flag FHIR resource.
type FlagRow struct {
	ID          string  `json:"id"`
	PatientID   string  `json:"patient_id"`
	Status      string  `json:"status"`
	Category    *string `json:"category"`
	Code        *string `json:"code"`
	PeriodStart *string `json:"period_start"`
	PeriodEnd   *string `json:"period_end"`
	GeneratedBy *string `json:"generated_by"`
	SiteID      string  `json:"site_id"`
	LastUpdated string  `json:"last_updated"`
	GitBlobHash string  `json:"git_blob_hash"`
}

// TimelineEvent represents a single event in a patient timeline.
type TimelineEvent struct {
	EventType  string `json:"event_type"`
	ResourceID string `json:"resource_id"`
	Date       string `json:"date"`
}

// Pagination holds pagination metadata for list responses.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginationOpts holds pagination options for list queries.
type PaginationOpts struct {
	Page    int
	PerPage int
	Sort    string
}

// ImmunizationRow holds indexed fields extracted from an Immunization FHIR resource.
type ImmunizationRow struct {
	ID                  string  `json:"id"`
	PatientID           string  `json:"patient_id"`
	Status              string  `json:"status"`
	VaccineCode         string  `json:"vaccine_code"`
	VaccineDisplay      *string `json:"vaccine_display"`
	OccurrenceDatetime  string  `json:"occurrence_datetime"`
	SiteID              string  `json:"site_id"`
	LastUpdated         string  `json:"last_updated"`
	GitBlobHash         string  `json:"git_blob_hash"`
}

// ProcedureRow holds indexed fields extracted from a Procedure FHIR resource.
type ProcedureRow struct {
	ID                string  `json:"id"`
	PatientID         string  `json:"patient_id"`
	Status            string  `json:"status"`
	Code              string  `json:"code"`
	CodeDisplay       *string `json:"code_display"`
	PerformedDatetime *string `json:"performed_datetime"`
	SiteID            string  `json:"site_id"`
	LastUpdated       string  `json:"last_updated"`
	GitBlobHash       string  `json:"git_blob_hash"`
}

// PractitionerRow holds indexed fields extracted from a Practitioner FHIR resource.
type PractitionerRow struct {
	ID          string `json:"id"`
	FamilyName  string `json:"family_name"`
	GivenNames  string `json:"given_names"` // JSON array as string
	Active      bool   `json:"active"`
	SiteID      string `json:"site_id"`
	LastUpdated string `json:"last_updated"`
	GitBlobHash string `json:"git_blob_hash"`
}

// OrganizationRow holds indexed fields extracted from an Organization FHIR resource.
type OrganizationRow struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        *string `json:"type"`
	Active      bool    `json:"active"`
	SiteID      string  `json:"site_id"`
	LastUpdated string  `json:"last_updated"`
	GitBlobHash string  `json:"git_blob_hash"`
}

// LocationRow holds indexed fields extracted from a Location FHIR resource.
type LocationRow struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        *string `json:"type"`
	Status      string  `json:"status"`
	SiteID      string  `json:"site_id"`
	LastUpdated string  `json:"last_updated"`
	GitBlobHash string  `json:"git_blob_hash"`
}

// MeasureReportRow holds indexed fields extracted from a MeasureReport FHIR resource.
type MeasureReportRow struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   *string `json:"period_end"`
	Reporter    *string `json:"reporter"`
	SiteID      string  `json:"site_id"`
	LastUpdated string  `json:"last_updated"`
	GitBlobHash string  `json:"git_blob_hash"`
}
