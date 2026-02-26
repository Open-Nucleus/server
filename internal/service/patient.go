package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

// patientAdapter adapts the Patient gRPC client to the PatientService interface.
type patientAdapter struct {
	pool *grpcclient.Pool
}

func NewPatientService(pool *grpcclient.Pool) PatientService {
	return &patientAdapter{pool: pool}
}

func (p *patientAdapter) ListPatients(ctx context.Context, req *ListPatientsRequest) (*ListPatientsResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) GetPatient(ctx context.Context, patientID string) (*PatientBundle, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) SearchPatients(ctx context.Context, query string, page, perPage int) (*ListPatientsResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) CreatePatient(ctx context.Context, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) UpdatePatient(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) DeletePatient(ctx context.Context, patientID string) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) MatchPatients(ctx context.Context, req *MatchPatientsRequest) (*MatchPatientsResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) GetPatientHistory(ctx context.Context, patientID string) (*PatientHistoryResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) GetPatientTimeline(ctx context.Context, patientID string) (*PatientTimelineResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

// --- Encounters ---

func (p *patientAdapter) ListEncounters(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) GetEncounter(ctx context.Context, patientID, encounterID string) (any, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) CreateEncounter(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) UpdateEncounter(ctx context.Context, patientID, encounterID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

// --- Observations ---

func (p *patientAdapter) ListObservations(ctx context.Context, patientID string, filters ObservationFilters, page, perPage int) (*ClinicalListResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) GetObservation(ctx context.Context, patientID, observationID string) (any, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) CreateObservation(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

// --- Conditions ---

func (p *patientAdapter) ListConditions(ctx context.Context, patientID string, filters ConditionFilters, page, perPage int) (*ClinicalListResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) CreateCondition(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) UpdateCondition(ctx context.Context, patientID, conditionID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

// --- Medication Requests ---

func (p *patientAdapter) ListMedicationRequests(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) CreateMedicationRequest(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) UpdateMedicationRequest(ctx context.Context, patientID, medicationRequestID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

// --- Allergy Intolerances ---

func (p *patientAdapter) ListAllergyIntolerances(ctx context.Context, patientID string, page, perPage int) (*ClinicalListResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) CreateAllergyIntolerance(ctx context.Context, patientID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}

func (p *patientAdapter) UpdateAllergyIntolerance(ctx context.Context, patientID, allergyIntoleranceID string, body json.RawMessage) (*WriteResponse, error) {
	_, err := p.pool.Conn("patient")
	if err != nil {
		return nil, fmt.Errorf("patient service unavailable: %w", err)
	}
	return nil, fmt.Errorf("patient service unavailable: backend not connected")
}
