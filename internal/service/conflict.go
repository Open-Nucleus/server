package service

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type conflictAdapter struct {
	pool *grpcclient.Pool
}

func NewConflictService(pool *grpcclient.Pool) ConflictService {
	return &conflictAdapter{pool: pool}
}

func (c *conflictAdapter) ListConflicts(ctx context.Context, page, perPage int) (*ConflictListResponse, error) {
	_, err := c.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("conflict service unavailable: %w", err)
	}
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}

func (c *conflictAdapter) GetConflict(ctx context.Context, conflictID string) (*ConflictDetail, error) {
	_, err := c.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("conflict service unavailable: %w", err)
	}
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}

func (c *conflictAdapter) ResolveConflict(ctx context.Context, req *ResolveConflictRequest) (*ResolveConflictResponse, error) {
	_, err := c.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("conflict service unavailable: %w", err)
	}
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}

func (c *conflictAdapter) DeferConflict(ctx context.Context, req *DeferConflictRequest) (*DeferConflictResponse, error) {
	_, err := c.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("conflict service unavailable: %w", err)
	}
	return nil, fmt.Errorf("conflict service unavailable: backend not connected")
}
