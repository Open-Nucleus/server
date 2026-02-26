package service

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type anchorAdapter struct {
	pool *grpcclient.Pool
}

func NewAnchorService(pool *grpcclient.Pool) AnchorService {
	return &anchorAdapter{pool: pool}
}

func (a *anchorAdapter) GetStatus(ctx context.Context) (*AnchorStatusResponse, error) {
	_, err := a.pool.Conn("anchor")
	if err != nil {
		return nil, fmt.Errorf("anchor service unavailable: %w", err)
	}
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}

func (a *anchorAdapter) Verify(ctx context.Context, commitHash string) (*AnchorVerifyResponse, error) {
	_, err := a.pool.Conn("anchor")
	if err != nil {
		return nil, fmt.Errorf("anchor service unavailable: %w", err)
	}
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}

func (a *anchorAdapter) GetHistory(ctx context.Context, page, perPage int) (*AnchorHistoryResponse, error) {
	_, err := a.pool.Conn("anchor")
	if err != nil {
		return nil, fmt.Errorf("anchor service unavailable: %w", err)
	}
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}

func (a *anchorAdapter) TriggerAnchor(ctx context.Context) (*AnchorTriggerResponse, error) {
	_, err := a.pool.Conn("anchor")
	if err != nil {
		return nil, fmt.Errorf("anchor service unavailable: %w", err)
	}
	return nil, fmt.Errorf("anchor service unavailable: backend not connected")
}
