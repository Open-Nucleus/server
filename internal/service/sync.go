package service

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type syncAdapter struct {
	pool *grpcclient.Pool
}

func NewSyncService(pool *grpcclient.Pool) SyncService {
	return &syncAdapter{pool: pool}
}

func (s *syncAdapter) GetStatus(ctx context.Context) (*SyncStatusResponse, error) {
	_, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}

func (s *syncAdapter) ListPeers(ctx context.Context) (*SyncPeersResponse, error) {
	_, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}

func (s *syncAdapter) TriggerSync(ctx context.Context, targetNode string) (*SyncTriggerResponse, error) {
	_, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}

func (s *syncAdapter) GetHistory(ctx context.Context, page, perPage int) (*SyncHistoryResponse, error) {
	_, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}

func (s *syncAdapter) ExportBundle(ctx context.Context, req *BundleExportRequest) (*BundleExportResponse, error) {
	_, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}

func (s *syncAdapter) ImportBundle(ctx context.Context, req *BundleImportRequest) (*BundleImportResponse, error) {
	_, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return nil, fmt.Errorf("sync service unavailable: backend not connected")
}
