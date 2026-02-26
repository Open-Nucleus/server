package service

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type sentinelAdapter struct {
	pool *grpcclient.Pool
}

func NewSentinelService(pool *grpcclient.Pool) SentinelService {
	return &sentinelAdapter{pool: pool}
}

func (s *sentinelAdapter) ListAlerts(ctx context.Context, page, perPage int) (*AlertListResponse, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("sentinel service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}

func (s *sentinelAdapter) GetAlertSummary(ctx context.Context) (*AlertSummaryResponse, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("sentinel service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}

func (s *sentinelAdapter) GetAlert(ctx context.Context, alertID string) (*AlertDetail, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("sentinel service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}

func (s *sentinelAdapter) AcknowledgeAlert(ctx context.Context, alertID string) (*AlertDetail, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("sentinel service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}

func (s *sentinelAdapter) DismissAlert(ctx context.Context, alertID, reason string) (*AlertDetail, error) {
	_, err := s.pool.Conn("sentinel")
	if err != nil {
		return nil, fmt.Errorf("sentinel service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sentinel service unavailable: backend not connected")
}
