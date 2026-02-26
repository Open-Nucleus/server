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
	SearchMedications(ctx context.Context, query string, page, perPage int) (*MedicationListResponse, error)
	GetMedication(ctx context.Context, code string) (*MedicationDetail, error)
	CheckInteractions(ctx context.Context, req *CheckInteractionsRequest) (*CheckInteractionsResponse, error)
	GetAvailability(ctx context.Context, siteID string) (*AvailabilityResponse, error)
	UpdateAvailability(ctx context.Context, siteID string, body json.RawMessage) (*UpdateAvailabilityResponse, error)
}

// AnchorService defines the interface for IOTA anchor operations.
type AnchorService interface {
	GetStatus(ctx context.Context) (*AnchorStatusResponse, error)
	Verify(ctx context.Context, commitHash string) (*AnchorVerifyResponse, error)
	GetHistory(ctx context.Context, page, perPage int) (*AnchorHistoryResponse, error)
	TriggerAnchor(ctx context.Context) (*AnchorTriggerResponse, error)
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
	Code      string `json:"code"`
	Display   string `json:"display"`
	Form      string `json:"form"`
	Route     string `json:"route"`
	Category  string `json:"category"`
	Available bool   `json:"available"`
}

type CheckInteractionsRequest struct {
	MedicationCodes []string `json:"medication_codes"`
	PatientID       string   `json:"patient_id"`
}

type CheckInteractionsResponse struct {
	Interactions []InteractionDetail `json:"interactions"`
}

type InteractionDetail struct {
	Severity    string `json:"severity"`
	Type        string `json:"type"`
	Description string `json:"description"`
	MedicationA string `json:"medication_a"`
	MedicationB string `json:"medication_b"`
	Source      string `json:"source"`
}

type AvailabilityResponse struct {
	Items  []AvailabilityItem `json:"items"`
	SiteID string             `json:"site_id"`
}

type AvailabilityItem struct {
	MedicationCode string `json:"medication_code"`
	Display        string `json:"display"`
	Quantity       int    `json:"quantity"`
	Unit           string `json:"unit"`
	LastUpdated    string `json:"last_updated"`
}

type UpdateAvailabilityResponse struct {
	UpdatedCount int `json:"updated_count"`
}

// --- Anchor DTOs ---

type AnchorStatusResponse struct {
	State          string `json:"state"`
	LastAnchor     string `json:"last_anchor"`
	TangleNode     string `json:"tangle_node"`
	PendingCommits int    `json:"pending_commits"`
}

type AnchorVerifyResponse struct {
	Verified        bool   `json:"verified"`
	AnchorID        string `json:"anchor_id"`
	TangleMessageID string `json:"tangle_message_id"`
	AnchoredAt      string `json:"anchored_at"`
	CommitHash      string `json:"commit_hash"`
}

type AnchorHistoryResponse struct {
	Events     []AnchorEvent `json:"events"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	Total      int           `json:"total"`
	TotalPages int           `json:"total_pages"`
}

type AnchorEvent struct {
	AnchorID        string `json:"anchor_id"`
	CommitHash      string `json:"commit_hash"`
	TangleMessageID string `json:"tangle_message_id"`
	Timestamp       string `json:"timestamp"`
	State           string `json:"state"`
}

type AnchorTriggerResponse struct {
	AnchorID string   `json:"anchor_id"`
	State    string   `json:"state"`
	Git      *GitMeta `json:"git"`
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
