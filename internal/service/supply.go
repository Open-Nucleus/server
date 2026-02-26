package service

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type supplyAdapter struct {
	pool *grpcclient.Pool
}

func NewSupplyService(pool *grpcclient.Pool) SupplyService {
	return &supplyAdapter{pool: pool}
}

func (s *supplyAdapter) GetInventory(ctx context.Context, page, perPage int) (*InventoryListResponse, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("supply service unavailable: %w", err)
	}
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}

func (s *supplyAdapter) GetInventoryItem(ctx context.Context, itemCode string) (*InventoryItemDetail, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("supply service unavailable: %w", err)
	}
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}

func (s *supplyAdapter) RecordDelivery(ctx context.Context, req *RecordDeliveryRequest) (*RecordDeliveryResponse, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("supply service unavailable: %w", err)
	}
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}

func (s *supplyAdapter) GetPredictions(ctx context.Context) (*PredictionsResponse, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("supply service unavailable: %w", err)
	}
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}

func (s *supplyAdapter) GetRedistribution(ctx context.Context) (*RedistributionResponse, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("supply service unavailable: %w", err)
	}
	return nil, fmt.Errorf("supply service unavailable: backend not connected")
}
