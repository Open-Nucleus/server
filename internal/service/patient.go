package service

import (
	"context"
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
