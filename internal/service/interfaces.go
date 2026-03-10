package service

import (
	"context"
	"encoding/json"
)

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, error)
	Logout(ctx context.Context, token string) error
	Whoami(ctx context.Context) (*WhoamiResponse, error)
}

// PatientService defines the interface for patient operations.
type PatientService interface {
	// Reads (existing)
	ListPatients(ctx context.Context, req *ListPatientsRequest) (*ListPatientsResponse, error)
	GetPatient(ctx context.Context, patientID string) (*PatientBundle, error)
	SearchPatients(ctx context.Context, query string, page, perPage int) (*ListPatientsResponse, error)

	// Writes
	CreatePatient(ctx context.Context, body json.RawMessage) (*WriteResponse, error)
	UpdatePatient(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)
	DeletePatient(ctx context.Context, patientID string) (*WriteResponse, error)
	MatchPatients(ctx context.Context, req *MatchPatientsRequest) (*MatchPatientsResponse, error)
	GetPatientHistory(ctx context.Context, patientID string) (*PatientHistoryResponse, error)
	GetPatientTimeline(ctx context.Context, patientID string) (*PatientTimelineResponse, error)

	// Encounters
	ListEncounters(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error)
	GetEncounter(ctx context.Context, patientID, encounterID string) (any, error)
	CreateEncounter(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)
	UpdateEncounter(ctx context.Context, patientID, encounterID string, body json.RawMessage) (*WriteResponse, error)

	// Observations
	ListObservations(ctx context.Context, patientID string, filters ObservationFilters, page, perPage int) (*ClinicalListResponse, error)
	GetObservation(ctx context.Context, patientID, observationID string) (any, error)
	CreateObservation(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)

	// Conditions
	ListConditions(ctx context.Context, patientID string, filters ConditionFilters, page, perPage int) (*ClinicalListResponse, error)
	CreateCondition(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)
	UpdateCondition(ctx context.Context, patientID, conditionID string, body json.RawMessage) (*WriteResponse, error)

	// Medication Requests
	ListMedicationRequests(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error)
	CreateMedicationRequest(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)
	UpdateMedicationRequest(ctx context.Context, patientID, medicationRequestID string, body json.RawMessage) (*WriteResponse, error)

	// Allergy Intolerances
	ListAllergyIntolerances(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error)
	CreateAllergyIntolerance(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)
	UpdateAllergyIntolerance(ctx context.Context, patientID, allergyIntoleranceID string, body json.RawMessage) (*WriteResponse, error)

	// Immunizations
	ListImmunizations(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error)
	GetImmunization(ctx context.Context, patientID, immunizationID string) (any, error)
	CreateImmunization(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)

	// Procedures
	ListProcedures(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error)
	GetProcedure(ctx context.Context, patientID, procedureID string) (any, error)
	CreateProcedure(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error)

	// Generic top-level resources (Practitioner, Organization, Location)
	ListResources(ctx context.Context, resourceType string, page, perPage int) (*ClinicalListResponse, error)
	GetResource(ctx context.Context, resourceType, resourceID string) (any, error)
	CreateResource(ctx context.Context, resourceType string, body json.RawMessage) (*WriteResponse, error)
	UpdateResource(ctx context.Context, resourceType, resourceID string, body json.RawMessage) (*WriteResponse, error)

	// Crypto-erasure (GDPR Art 17, POPIA, Kenya DPA, Nigeria NDPA)
	ErasePatient(ctx context.Context, patientID string) (*EraseResponse, error)
}

// SyncService defines the interface for sync operations.
type SyncService interface {
	GetStatus(ctx context.Context) (*SyncStatusResponse, error)
	ListPeers(ctx context.Context) (*SyncPeersResponse, error)
	TriggerSync(ctx context.Context, targetNode string) (*SyncTriggerResponse, error)
	GetHistory(ctx context.Context, page, perPage int) (*SyncHistoryResponse, error)
	ExportBundle(ctx context.Context, req *BundleExportRequest) (*BundleExportResponse, error)
	ImportBundle(ctx context.Context, req *BundleImportRequest) (*BundleImportResponse, error)
}

// ConflictService defines the interface for conflict resolution.
type ConflictService interface {
	ListConflicts(ctx context.Context, page, perPage int) (*ConflictListResponse, error)
	GetConflict(ctx context.Context, conflictID string) (*ConflictDetail, error)
	ResolveConflict(ctx context.Context, req *ResolveConflictRequest) (*ResolveConflictResponse, error)
	DeferConflict(ctx context.Context, req *DeferConflictRequest) (*DeferConflictResponse, error)
}

// SentinelService defines the interface for alert operations.
type SentinelService interface {
	ListAlerts(ctx context.Context, page, perPage int) (*AlertListResponse, error)
	GetAlertSummary(ctx context.Context) (*AlertSummaryResponse, error)
	GetAlert(ctx context.Context, alertID string) (*AlertDetail, error)
	AcknowledgeAlert(ctx context.Context, alertID string) (*AlertDetail, error)
	DismissAlert(ctx context.Context, alertID, reason string) (*AlertDetail, error)
}

// FormularyService defines the interface for formulary operations.
type FormularyService interface {
	// Drug lookup
	SearchMedications(ctx context.Context, query, category string, page, perPage int) (*MedicationListResponse, error)
	GetMedication(ctx context.Context, code string) (*MedicationDetail, error)
	ListMedicationsByCategory(ctx context.Context, category string, page, perPage int) (*MedicationListResponse, error)

	// Safety checks
	CheckInteractions(ctx context.Context, req *CheckInteractionsRequest) (*CheckInteractionsResponse, error)
	CheckAllergyConflicts(ctx context.Context, req *CheckAllergyConflictsRequest) (*CheckAllergyConflictsResponse, error)

	// Dosing (stub — depends on open-pharm-dosing)
	ValidateDosing(ctx context.Context, req *ValidateDosingRequest) (*ValidateDosingResponse, error)
	GetDosingOptions(ctx context.Context, medicationCode string, patientWeightKg float64) (*GetDosingOptionsResponse, error)
	GenerateSchedule(ctx context.Context, req *GenerateScheduleRequest) (*GenerateScheduleResponse, error)

	// Stock management
	GetStockLevel(ctx context.Context, siteID, medicationCode string) (*StockLevelResponse, error)
	UpdateStockLevel(ctx context.Context, req *UpdateStockLevelRequest) (*UpdateStockLevelResponse, error)
	RecordDelivery(ctx context.Context, req *FormularyDeliveryRequest) (*FormularyDeliveryResponse, error)
	GetStockPrediction(ctx context.Context, siteID, medicationCode string) (*StockPredictionResponse, error)
	GetRedistributionSuggestions(ctx context.Context, medicationCode string) (*FormularyRedistributionResponse, error)

	// Formulary metadata
	GetFormularyInfo(ctx context.Context) (*FormularyInfoResponse, error)
}

// AnchorService defines the interface for anchor operations.
type AnchorService interface {
	// Anchoring
	GetStatus(ctx context.Context) (*AnchorStatusResponse, error)
	Verify(ctx context.Context, commitHash string) (*AnchorVerifyResponse, error)
	GetHistory(ctx context.Context, page, perPage int) (*AnchorHistoryResponse, error)
	TriggerAnchor(ctx context.Context) (*AnchorTriggerResponse, error)

	// DID
	GetNodeDID(ctx context.Context) (*DIDDocumentResponse, error)
	GetDeviceDID(ctx context.Context, deviceID string) (*DIDDocumentResponse, error)
	ResolveDID(ctx context.Context, did string) (*DIDDocumentResponse, error)

	// Credentials
	IssueDataIntegrityCredential(ctx context.Context, req *IssueCredentialRequest) (*CredentialResponse, error)
	VerifyCredential(ctx context.Context, credentialJSON string) (*CredentialVerificationResponse, error)
	ListCredentials(ctx context.Context, credType string, page, perPage int) (*CredentialListResponse, error)

	// Backend
	ListBackends(ctx context.Context) (*BackendListResponse, error)
	GetBackendStatus(ctx context.Context, name string) (*BackendStatusResponse, error)
	GetQueueStatus(ctx context.Context) (*QueueStatusResponse, error)

	// Health
	Health(ctx context.Context) (*AnchorHealthResponse, error)
}

// SupplyService defines the interface for supply chain operations.
type SupplyService interface {
	GetInventory(ctx context.Context, page, perPage int) (*InventoryListResponse, error)
	GetInventoryItem(ctx context.Context, itemCode string) (*InventoryItemDetail, error)
	RecordDelivery(ctx context.Context, req *RecordDeliveryRequest) (*RecordDeliveryResponse, error)
	GetPredictions(ctx context.Context) (*PredictionsResponse, error)
	GetRedistribution(ctx context.Context) (*RedistributionResponse, error)
}

// --- Auth DTOs ---

type LoginRequest struct {
	DeviceID          string            `json:"device_id"`
	PublicKey         string            `json:"public_key"`
	ChallengeResponse ChallengeResponseDTO `json:"challenge_response"`
	PractitionerID   string            `json:"practitioner_id"`
}

type ChallengeResponseDTO struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	Timestamp string `json:"timestamp"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	ExpiresAt    string `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	Role         RoleDTO `json:"role"`
	SiteID       string `json:"site_id"`
	NodeID       string `json:"node_id"`
}

type RoleDTO struct {
	Code        string   `json:"code"`
	Display     string   `json:"display"`
	Permissions []string `json:"permissions"`
}

type RefreshResponse struct {
	Token        string `json:"token"`
	ExpiresAt    string `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

type WhoamiResponse struct {
	Subject string  `json:"subject"`
	NodeID  string  `json:"node_id"`
	SiteID  string  `json:"site_id"`
	Role    RoleDTO `json:"role"`
}

// --- Patient DTOs ---

type ListPatientsRequest struct {
	Page          int
	PerPage       int
	Sort          string
	Gender        string
	BirthDateFrom string
	BirthDateTo   string
	SiteID        string
	Status        string
	HasAlerts     bool
}

type ListPatientsResponse struct {
	Patients   []any
	Page       int
	PerPage    int
	Total      int
	TotalPages int
}

type PatientBundle struct {
	Patient              any   `json:"patient"`
	Encounters           []any `json:"encounters"`
	Observations         []any `json:"observations"`
	Conditions           []any `json:"conditions"`
	MedicationRequests   []any `json:"medication_requests"`
	AllergyIntolerances  []any `json:"allergy_intolerances"`
	Flags                []any `json:"flags"`
}

// --- Shared Write DTOs ---

type WriteResponse struct {
	Resource any      `json:"resource,omitempty"`
	Git      *GitMeta `json:"git"`
}

type GitMeta struct {
	Commit  string `json:"commit"`
	Message string `json:"message"`
}

// EraseResponse is returned by the crypto-erasure endpoint.
type EraseResponse struct {
	Erased    bool   `json:"erased"`
	PatientID string `json:"patient_id"`
}

// --- Patient-specific DTOs ---

type MatchPatientsRequest struct {
	FamilyName     string   `json:"family_name"`
	GivenNames     []string `json:"given_names"`
	Gender         string   `json:"gender"`
	BirthDateApprox string  `json:"birth_date_approx"`
	District       string   `json:"district"`
	Threshold      float64  `json:"threshold"`
}

type MatchPatientsResponse struct {
	Matches []PatientMatch `json:"matches"`
}

type PatientMatch struct {
	PatientID    string   `json:"patient_id"`
	Confidence   float64  `json:"confidence"`
	MatchFactors []string `json:"match_factors"`
}

type PatientHistoryResponse struct {
	Entries []HistoryEntry `json:"entries"`
}

type HistoryEntry struct {
	CommitHash   string `json:"commit_hash"`
	Timestamp    string `json:"timestamp"`
	Author       string `json:"author"`
	Node         string `json:"node"`
	Site         string `json:"site"`
	Operation    string `json:"operation"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Message      string `json:"message"`
}

type PatientTimelineResponse struct {
	Events []any `json:"events"`
}

// --- Clinical DTOs ---

type ClinicalListResponse struct {
	Resources  []any `json:"resources"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int   `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ObservationFilters struct {
	Code        string
	Category    string
	DateFrom    string
	DateTo      string
	EncounterID string
}

type ConditionFilters struct {
	ClinicalStatus string
	Category       string
	Code           string
}

// --- Sync DTOs ---

type SyncStatusResponse struct {
	State          string `json:"state"`
	LastSync       string `json:"last_sync"`
	PendingChanges int    `json:"pending_changes"`
	NodeID         string `json:"node_id"`
	SiteID         string `json:"site_id"`
}

type SyncPeersResponse struct {
	Peers []PeerInfo `json:"peers"`
}

type PeerInfo struct {
	NodeID    string `json:"node_id"`
	SiteID    string `json:"site_id"`
	LastSeen  string `json:"last_seen"`
	State     string `json:"state"`
	LatencyMs int    `json:"latency_ms"`
}

type SyncTriggerResponse struct {
	SyncID string   `json:"sync_id"`
	State  string   `json:"state"`
	Git    *GitMeta `json:"git"`
}

type SyncHistoryResponse struct {
	Events     []SyncEvent `json:"events"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

type SyncEvent struct {
	SyncID               string `json:"sync_id"`
	Timestamp            string `json:"timestamp"`
	Direction            string `json:"direction"`
	PeerNode             string `json:"peer_node"`
	State                string `json:"state"`
	ResourcesTransferred int    `json:"resources_transferred"`
}

type BundleExportRequest struct {
	ResourceTypes []string `json:"resource_types"`
	Since         string   `json:"since"`
}

type BundleExportResponse struct {
	BundleData    []byte   `json:"bundle_data"`
	Format        string   `json:"format"`
	ResourceCount int      `json:"resource_count"`
	Git           *GitMeta `json:"git"`
}

type BundleImportRequest struct {
	BundleData []byte `json:"bundle_data"`
	Format     string `json:"format"`
	Author     string `json:"author"`
	NodeID     string `json:"node_id"`
	SiteID     string `json:"site_id"`
}

type BundleImportResponse struct {
	ResourcesImported int      `json:"resources_imported"`
	ResourcesSkipped  int      `json:"resources_skipped"`
	Errors            []string `json:"errors"`
	Git               *GitMeta `json:"git"`
}

// --- Conflict DTOs ---

type ConflictListResponse struct {
	Conflicts  []ConflictDetail `json:"conflicts"`
	Page       int              `json:"page"`
	PerPage    int              `json:"per_page"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
}

type ConflictDetail struct {
	ID             string `json:"id"`
	ResourceType   string `json:"resource_type"`
	ResourceID     string `json:"resource_id"`
	Status         string `json:"status"`
	DetectedAt     string `json:"detected_at"`
	LocalVersion   any    `json:"local_version"`
	RemoteVersion  any    `json:"remote_version"`
	LocalNode      string `json:"local_node"`
	RemoteNode     string `json:"remote_node"`
}

type ResolveConflictRequest struct {
	ConflictID     string          `json:"conflict_id"`
	Resolution     string          `json:"resolution"`
	MergedResource json.RawMessage `json:"merged_resource"`
	Author         string          `json:"author"`
}

type ResolveConflictResponse struct {
	Git *GitMeta `json:"git"`
}

type DeferConflictRequest struct {
	ConflictID string `json:"conflict_id"`
	Reason     string `json:"reason"`
}

type DeferConflictResponse struct {
	Status string `json:"status"`
}

// --- Sentinel / Alert DTOs ---

type AlertListResponse struct {
	Alerts     []AlertDetail `json:"alerts"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
}

type AlertSummaryResponse struct {
	Total          int `json:"total"`
	Critical       int `json:"critical"`
	Warning        int `json:"warning"`
	Info           int `json:"info"`
	Unacknowledged int `json:"unacknowledged"`
}

type AlertDetail struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Status         string `json:"status"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	PatientID      string `json:"patient_id"`
	CreatedAt      string `json:"created_at"`
	AcknowledgedAt string `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string `json:"acknowledged_by,omitempty"`
}

// --- Formulary DTOs ---

type MedicationListResponse struct {
	Medications []MedicationDetail `json:"medications"`
	Page        int                `json:"page"`
	PerPage     int                `json:"per_page"`
	Total       int                `json:"total"`
	TotalPages  int                `json:"total_pages"`
}

type MedicationDetail struct {
	Code              string   `json:"code"`
	Display           string   `json:"display"`
	Form              string   `json:"form"`
	Route             string   `json:"route"`
	Category          string   `json:"category"`
	Available         bool     `json:"available"`
	WHOEssential      bool     `json:"who_essential"`
	TherapeuticClass  string   `json:"therapeutic_class"`
	CommonFrequencies []string `json:"common_frequencies,omitempty"`
	Strength          string   `json:"strength,omitempty"`
	Unit              string   `json:"unit,omitempty"`
}

type CheckInteractionsRequest struct {
	MedicationCodes []string `json:"medication_codes"`
	PatientID       string   `json:"patient_id"`
	AllergyCodes    []string `json:"allergy_codes,omitempty"`
	SiteID          string   `json:"site_id,omitempty"`
}

type CheckInteractionsResponse struct {
	Interactions   []InteractionDetail `json:"interactions"`
	AllergyAlerts  []AllergyAlertDTO   `json:"allergy_alerts,omitempty"`
	DosingWarnings []DosingWarningDTO  `json:"dosing_warnings,omitempty"`
	StockSummary   *StockSummaryDTO    `json:"stock_summary,omitempty"`
	OverallRisk    string              `json:"overall_risk"`
}

type InteractionDetail struct {
	Severity       string `json:"severity"`
	Type           string `json:"type"`
	Description    string `json:"description"`
	MedicationA    string `json:"medication_a"`
	MedicationB    string `json:"medication_b"`
	Source         string `json:"source"`
	ClinicalEffect  string `json:"clinical_effect,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}

type AllergyAlertDTO struct {
	Severity             string `json:"severity"`
	AllergyCode          string `json:"allergy_code"`
	MedicationCode       string `json:"medication_code"`
	Description          string `json:"description"`
	CrossReactivityClass string `json:"cross_reactivity_class,omitempty"`
}

type DosingWarningDTO struct {
	MedicationCode string `json:"medication_code"`
	Warning        string `json:"warning"`
	Severity       string `json:"severity"`
}

type StockSummaryDTO struct {
	Items []StockItemDTO `json:"items"`
}

type StockItemDTO struct {
	MedicationCode string `json:"medication_code"`
	Available      bool   `json:"available"`
	Quantity       int    `json:"quantity"`
	Unit           string `json:"unit"`
}

type CheckAllergyConflictsRequest struct {
	MedicationCodes []string `json:"medication_codes"`
	AllergyCodes    []string `json:"allergy_codes"`
}

type CheckAllergyConflictsResponse struct {
	Alerts []AllergyAlertDTO `json:"alerts"`
	Safe   bool              `json:"safe"`
}

type ValidateDosingRequest struct {
	MedicationCode  string  `json:"medication_code"`
	DoseValue       float64 `json:"dose_value"`
	DoseUnit        string  `json:"dose_unit"`
	Frequency       string  `json:"frequency"`
	Route           string  `json:"route"`
	PatientWeightKg float64 `json:"patient_weight_kg"`
}

type ValidateDosingResponse struct {
	Valid      bool   `json:"valid"`
	Message    string `json:"message"`
	Configured bool   `json:"configured"`
}

type GetDosingOptionsResponse struct {
	Options    []DosingOptionDTO `json:"options,omitempty"`
	Configured bool              `json:"configured"`
	Message    string            `json:"message"`
}

type DosingOptionDTO struct {
	DoseValue  float64 `json:"dose_value"`
	DoseUnit   string  `json:"dose_unit"`
	Frequency  string  `json:"frequency"`
	Route      string  `json:"route"`
	Indication string  `json:"indication"`
}

type GenerateScheduleRequest struct {
	MedicationCode string  `json:"medication_code"`
	DoseValue      float64 `json:"dose_value"`
	DoseUnit       string  `json:"dose_unit"`
	Frequency      string  `json:"frequency"`
	StartTime      string  `json:"start_time"`
	DurationDays   int     `json:"duration_days"`
}

type GenerateScheduleResponse struct {
	Entries    []ScheduleEntryDTO `json:"entries,omitempty"`
	Configured bool               `json:"configured"`
	Message    string             `json:"message"`
}

type ScheduleEntryDTO struct {
	Time      string  `json:"time"`
	DoseValue float64 `json:"dose_value"`
	DoseUnit  string  `json:"dose_unit"`
	Note      string  `json:"note,omitempty"`
}

type StockLevelResponse struct {
	SiteID               string  `json:"site_id"`
	MedicationCode       string  `json:"medication_code"`
	Quantity             int     `json:"quantity"`
	Unit                 string  `json:"unit"`
	LastUpdated          string  `json:"last_updated"`
	EarliestExpiry       string  `json:"earliest_expiry,omitempty"`
	DailyConsumptionRate float64 `json:"daily_consumption_rate"`
}

type UpdateStockLevelRequest struct {
	SiteID         string `json:"site_id"`
	MedicationCode string `json:"medication_code"`
	Quantity       int    `json:"quantity"`
	Unit           string `json:"unit"`
	Reason         string `json:"reason"`
	UpdatedBy      string `json:"updated_by"`
}

type UpdateStockLevelResponse struct {
	Success     bool   `json:"success"`
	LastUpdated string `json:"last_updated"`
}

type FormularyDeliveryRequest struct {
	SiteID       string                `json:"site_id"`
	Items        []FormularyDeliveryItem `json:"items"`
	ReceivedBy   string                `json:"received_by"`
	DeliveryDate string                `json:"delivery_date"`
}

type FormularyDeliveryItem struct {
	MedicationCode string `json:"medication_code"`
	Quantity       int    `json:"quantity"`
	Unit           string `json:"unit"`
	BatchNumber    string `json:"batch_number"`
	ExpiryDate     string `json:"expiry_date"`
}

type FormularyDeliveryResponse struct {
	DeliveryID    string `json:"delivery_id"`
	ItemsRecorded int    `json:"items_recorded"`
}

type StockPredictionResponse struct {
	DaysRemaining     int    `json:"days_remaining"`
	RiskLevel         string `json:"risk_level"`
	EarliestExpiry    string `json:"earliest_expiry,omitempty"`
	ExpiringQuantity  int    `json:"expiring_quantity"`
	RecommendedAction string `json:"recommended_action"`
}

type FormularyRedistributionResponse struct {
	Suggestions []FormularyRedistributionSuggestion `json:"suggestions"`
}

type FormularyRedistributionSuggestion struct {
	FromSite          string `json:"from_site"`
	ToSite            string `json:"to_site"`
	SuggestedQuantity int    `json:"suggested_quantity"`
	Rationale         string `json:"rationale"`
	FromSiteQuantity  int    `json:"from_site_quantity"`
	ToSiteQuantity    int    `json:"to_site_quantity"`
}

type FormularyInfoResponse struct {
	Version              string   `json:"version"`
	TotalMedications     int      `json:"total_medications"`
	TotalInteractions    int      `json:"total_interactions"`
	LastUpdated          string   `json:"last_updated"`
	Categories           []string `json:"categories"`
	DosingEngineAvailable bool    `json:"dosing_engine_available"`
}

// --- Anchor DTOs ---

type AnchorStatusResponse struct {
	State          string `json:"state"`
	LastAnchorID   string `json:"last_anchor_id"`
	LastAnchorTime string `json:"last_anchor_time"`
	MerkleRoot     string `json:"merkle_root"`
	NodeDID        string `json:"node_did"`
	QueueDepth     int    `json:"queue_depth"`
	Backend        string `json:"backend"`
	PendingCommits int    `json:"pending_commits"`
}

type AnchorVerifyResponse struct {
	Verified   bool   `json:"verified"`
	AnchorID   string `json:"anchor_id"`
	MerkleRoot string `json:"merkle_root"`
	AnchoredAt string `json:"anchored_at"`
	CommitHash string `json:"commit_hash"`
	State      string `json:"state"`
}

type AnchorHistoryResponse struct {
	Records    []AnchorRecord `json:"records"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	Total      int            `json:"total"`
	TotalPages int            `json:"total_pages"`
}

type AnchorRecord struct {
	AnchorID   string `json:"anchor_id"`
	MerkleRoot string `json:"merkle_root"`
	GitHead    string `json:"git_head"`
	State      string `json:"state"`
	Timestamp  string `json:"timestamp"`
	Backend    string `json:"backend"`
	TxID       string `json:"tx_id"`
	NodeDID    string `json:"node_did"`
}

type AnchorTriggerResponse struct {
	AnchorID   string   `json:"anchor_id"`
	State      string   `json:"state"`
	MerkleRoot string   `json:"merkle_root"`
	GitHead    string   `json:"git_head"`
	Skipped    bool     `json:"skipped"`
	Message    string   `json:"message"`
	Git        *GitMeta `json:"git"`
}

type DIDDocumentResponse struct {
	ID                 string                     `json:"id"`
	Context            []string                   `json:"@context"`
	VerificationMethod []VerificationMethodDTO    `json:"verificationMethod"`
	Authentication     []string                   `json:"authentication"`
	AssertionMethod    []string                   `json:"assertionMethod"`
	Created            string                     `json:"created,omitempty"`
}

type VerificationMethodDTO struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase"`
}

type IssueCredentialRequest struct {
	AnchorID         string            `json:"anchor_id"`
	Types            []string          `json:"types,omitempty"`
	AdditionalClaims map[string]string `json:"additional_claims,omitempty"`
}

type CredentialResponse struct {
	ID                    string                 `json:"id"`
	Context               []string               `json:"@context"`
	Type                  []string               `json:"type"`
	Issuer                string                 `json:"issuer"`
	IssuanceDate          string                 `json:"issuanceDate"`
	ExpirationDate        string                 `json:"expirationDate,omitempty"`
	CredentialSubjectJSON string                 `json:"credentialSubjectJson"`
	Proof                 *CredentialProofDTO    `json:"proof"`
}

type CredentialProofDTO struct {
	Type               string `json:"type"`
	Created            string `json:"created"`
	VerificationMethod string `json:"verificationMethod"`
	ProofPurpose       string `json:"proofPurpose"`
	ProofValue         string `json:"proofValue"`
}

type CredentialVerificationResponse struct {
	Valid   bool   `json:"valid"`
	Issuer  string `json:"issuer"`
	Message string `json:"message"`
}

type CredentialListResponse struct {
	Credentials []CredentialResponse `json:"credentials"`
	Page        int                  `json:"page"`
	PerPage     int                  `json:"per_page"`
	Total       int                  `json:"total"`
	TotalPages  int                  `json:"total_pages"`
}

type BackendListResponse struct {
	Backends []BackendInfoDTO `json:"backends"`
}

type BackendInfoDTO struct {
	Name        string `json:"name"`
	Available   bool   `json:"available"`
	Description string `json:"description"`
}

type BackendStatusResponse struct {
	Name           string `json:"name"`
	Available      bool   `json:"available"`
	Description    string `json:"description"`
	AnchoredCount  int    `json:"anchored_count"`
	LastAnchorTime string `json:"last_anchor_time"`
}

type QueueStatusResponse struct {
	Pending        int             `json:"pending"`
	TotalProcessed int             `json:"total_processed"`
	Entries        []QueueEntryDTO `json:"entries"`
}

type QueueEntryDTO struct {
	AnchorID   string `json:"anchor_id"`
	MerkleRoot string `json:"merkle_root"`
	GitHead    string `json:"git_head"`
	EnqueuedAt string `json:"enqueued_at"`
	State      string `json:"state"`
}

type AnchorHealthResponse struct {
	Status      string `json:"status"`
	NodeDID     string `json:"node_did"`
	Backend     string `json:"backend"`
	AnchorCount int    `json:"anchor_count"`
	QueueDepth  int    `json:"queue_depth"`
}

// --- Supply Chain DTOs ---

type InventoryListResponse struct {
	Items      []InventoryItemDetail `json:"items"`
	Page       int                   `json:"page"`
	PerPage    int                   `json:"per_page"`
	Total      int                   `json:"total"`
	TotalPages int                   `json:"total_pages"`
}

type InventoryItemDetail struct {
	ItemCode     string `json:"item_code"`
	Display      string `json:"display"`
	Quantity     int    `json:"quantity"`
	Unit         string `json:"unit"`
	SiteID       string `json:"site_id"`
	LastUpdated  string `json:"last_updated"`
	ReorderLevel int    `json:"reorder_level"`
}

type RecordDeliveryRequest struct {
	SiteID       string         `json:"site_id"`
	Items        []DeliveryItem `json:"items"`
	ReceivedBy   string         `json:"received_by"`
	DeliveryDate string         `json:"delivery_date"`
}

type DeliveryItem struct {
	ItemCode    string `json:"item_code"`
	Quantity    int    `json:"quantity"`
	Unit        string `json:"unit"`
	BatchNumber string `json:"batch_number"`
	ExpiryDate  string `json:"expiry_date"`
}

type RecordDeliveryResponse struct {
	DeliveryID    string `json:"delivery_id"`
	ItemsRecorded int    `json:"items_recorded"`
}

type PredictionsResponse struct {
	Predictions []SupplyPrediction `json:"predictions"`
}

type SupplyPrediction struct {
	ItemCode              string `json:"item_code"`
	Display               string `json:"display"`
	CurrentQuantity       int    `json:"current_quantity"`
	PredictedDaysRemaining int   `json:"predicted_days_remaining"`
	RiskLevel             string `json:"risk_level"`
	RecommendedAction     string `json:"recommended_action"`
}

type RedistributionResponse struct {
	Suggestions []RedistributionSuggestion `json:"suggestions"`
}

type RedistributionSuggestion struct {
	ItemCode          string `json:"item_code"`
	FromSite          string `json:"from_site"`
	ToSite            string `json:"to_site"`
	SuggestedQuantity int    `json:"suggested_quantity"`
	Rationale         string `json:"rationale"`
}

// SmartService defines the interface for SMART on FHIR operations.
type SmartService interface {
	Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error)
	ExchangeToken(ctx context.Context, req *ExchangeTokenRequest) (*TokenResponse, error)
	IntrospectToken(ctx context.Context, token string) (*IntrospectResponse, error)
	RevokeToken(ctx context.Context, token string) error
	RegisterClient(ctx context.Context, req *RegisterClientRequest) (*ClientResponse, error)
	ListClients(ctx context.Context) (*ClientListResponse, error)
	GetClient(ctx context.Context, clientID string) (*ClientResponse, error)
	UpdateClient(ctx context.Context, clientID string, req *UpdateClientRequest) (*ClientResponse, error)
	DeleteClient(ctx context.Context, clientID string) error
	CreateLaunch(ctx context.Context, req *CreateLaunchRequest) (*CreateLaunchResponse, error)
}

type AuthorizeRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Launch              string `json:"launch"`
}

type AuthorizeResponse struct {
	RedirectURI string `json:"redirect_uri"`
}

type ExchangeTokenRequest struct {
	GrantType    string `json:"grant_type"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int32  `json:"expires_in"`
	Scope        string `json:"scope"`
	Patient      string `json:"patient,omitempty"`
	Encounter    string `json:"encounter,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type IntrospectResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Sub       string `json:"sub,omitempty"`
	Patient   string `json:"patient,omitempty"`
	Encounter string `json:"encounter,omitempty"`
	FHIRUser  string `json:"fhirUser,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
}

type RegisterClientRequest struct {
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	Scope                   string   `json:"scope"`
	GrantTypes              []string `json:"grant_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	LaunchModes             []string `json:"launch_modes"`
}

type UpdateClientRequest struct {
	Status string `json:"status"`
	Scope  string `json:"scope"`
}

type ClientResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientName              string   `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	Scope                   string   `json:"scope"`
	GrantTypes              []string `json:"grant_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	LaunchModes             []string `json:"launch_modes"`
	Status                  string   `json:"status"`
	RegisteredAt            string   `json:"registered_at"`
	RegisteredBy            string   `json:"registered_by"`
	ApprovedBy              string   `json:"approved_by,omitempty"`
	ApprovedAt              string   `json:"approved_at,omitempty"`
}

type ClientListResponse struct {
	Clients []ClientResponse `json:"clients"`
}

type CreateLaunchRequest struct {
	ClientID    string `json:"client_id"`
	PatientID   string `json:"patient_id"`
	EncounterID string `json:"encounter_id"`
}

type CreateLaunchResponse struct {
	LaunchToken string `json:"launch_token"`
}

// ConsentService defines the interface for consent management operations.
type ConsentService interface {
	CheckAccess(ctx context.Context, patientID, performerID, role string) (*ConsentAccessDecision, error)
	GrantConsent(ctx context.Context, patientID, performerID, scope string, period *ConsentPeriod, category string) (*ConsentGrantResponse, error)
	RevokeConsent(ctx context.Context, consentID string) error
	ListConsentsForPatient(ctx context.Context, patientID string, page, perPage int) (*ConsentListResponse, error)
	IssueConsentVC(ctx context.Context, consentID string) (*ConsentVCResponse, error)
}

type ConsentAccessDecision struct {
	Allowed   bool   `json:"allowed"`
	ConsentID string `json:"consent_id,omitempty"`
	Reason    string `json:"reason"`
}

type ConsentPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ConsentGrantResponse struct {
	ConsentID  string `json:"consent_id"`
	CommitHash string `json:"commit_hash"`
	Status     string `json:"status"`
}

type ConsentListResponse struct {
	Consents   []ConsentSummary `json:"consents"`
	Pagination *PaginationMeta  `json:"pagination,omitempty"`
}

type ConsentSummary struct {
	ID            string  `json:"id"`
	PatientID     string  `json:"patient_id"`
	Status        string  `json:"status"`
	ScopeCode     string  `json:"scope_code"`
	PerformerID   string  `json:"performer_id"`
	ProvisionType string  `json:"provision_type"`
	PeriodStart   *string `json:"period_start,omitempty"`
	PeriodEnd     *string `json:"period_end,omitempty"`
	Category      *string `json:"category,omitempty"`
	LastUpdated   string  `json:"last_updated"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ConsentVCResponse struct {
	VerifiableCredential any `json:"verifiable_credential"`
}
