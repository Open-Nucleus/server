package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type formularyAdapter struct {
	pool *grpcclient.Pool
}

func NewFormularyService(pool *grpcclient.Pool) FormularyService {
	return &formularyAdapter{pool: pool}
}

func (f *formularyAdapter) SearchMedications(ctx context.Context, query string, page, perPage int) (*MedicationListResponse, error) {
	_, err := f.pool.Conn("formulary")
	if err != nil {
		return nil, fmt.Errorf("formulary service unavailable: %w", err)
	}
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}

func (f *formularyAdapter) GetMedication(ctx context.Context, code string) (*MedicationDetail, error) {
	_, err := f.pool.Conn("formulary")
	if err != nil {
		return nil, fmt.Errorf("formulary service unavailable: %w", err)
	}
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}

func (f *formularyAdapter) CheckInteractions(ctx context.Context, req *CheckInteractionsRequest) (*CheckInteractionsResponse, error) {
	_, err := f.pool.Conn("formulary")
	if err != nil {
		return nil, fmt.Errorf("formulary service unavailable: %w", err)
	}
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}

func (f *formularyAdapter) GetAvailability(ctx context.Context, siteID string) (*AvailabilityResponse, error) {
	_, err := f.pool.Conn("formulary")
	if err != nil {
		return nil, fmt.Errorf("formulary service unavailable: %w", err)
	}
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}

func (f *formularyAdapter) UpdateAvailability(ctx context.Context, siteID string, body json.RawMessage) (*UpdateAvailabilityResponse, error) {
	_, err := f.pool.Conn("formulary")
	if err != nil {
		return nil, fmt.Errorf("formulary service unavailable: %w", err)
	}
	return nil, fmt.Errorf("formulary service unavailable: backend not connected")
}
