package service

import "context"

// AuthService defines the interface for authentication operations.
type AuthService interface {
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*RefreshResponse, error)
	Logout(ctx context.Context, token string) error
	Whoami(ctx context.Context) (*WhoamiResponse, error)
}

// PatientService defines the interface for patient operations.
type PatientService interface {
	ListPatients(ctx context.Context, req *ListPatientsRequest) (*ListPatientsResponse, error)
	GetPatient(ctx context.Context, patientID string) (*PatientBundle, error)
	SearchPatients(ctx context.Context, query string, page, perPage int) (*ListPatientsResponse, error)
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
