package service

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type syncAdapter struct {
	pool *grpcclient.Pool
}

func NewSyncService(pool *grpcclient.Pool) SyncService {
	return &syncAdapter{pool: pool}
}

func (s *syncAdapter) client() (syncv1.SyncServiceClient, error) {
	conn, err := s.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("sync service unavailable: %w", err)
	}
	return syncv1.NewSyncServiceClient(conn), nil
}

func (s *syncAdapter) GetStatus(ctx context.Context) (*SyncStatusResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetStatus(ctx, &syncv1.GetStatusRequest{})
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	return &SyncStatusResponse{
		State:          resp.State,
		LastSync:       resp.LastSync,
		PendingChanges: int(resp.PendingChanges),
		NodeID:         resp.NodeId,
		SiteID:         resp.SiteId,
	}, nil
}

func (s *syncAdapter) ListPeers(ctx context.Context) (*SyncPeersResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ListPeers(ctx, &syncv1.ListPeersRequest{})
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	peers := make([]PeerInfo, len(resp.Peers))
	for i, p := range resp.Peers {
		peers[i] = PeerInfo{
			NodeID:    p.NodeId,
			SiteID:    p.SiteId,
			LastSeen:  p.LastSeen,
			State:     p.State,
			LatencyMs: int(p.LatencyMs),
		}
	}
	return &SyncPeersResponse{Peers: peers}, nil
}

func (s *syncAdapter) TriggerSync(ctx context.Context, targetNode string) (*SyncTriggerResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.TriggerSync(ctx, &syncv1.TriggerSyncRequest{TargetNode: targetNode})
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	var git *GitMeta
	if resp.Git != nil {
		git = &GitMeta{Commit: resp.Git.Commit, Message: resp.Git.Message}
	}
	return &SyncTriggerResponse{
		SyncID: resp.SyncId,
		State:  resp.State,
		Git:    git,
	}, nil
}

func (s *syncAdapter) GetHistory(ctx context.Context, page, perPage int) (*SyncHistoryResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.GetHistory(ctx, &syncv1.GetSyncHistoryRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	events := make([]SyncEvent, len(resp.Entries))
	for i, e := range resp.Entries {
		events[i] = SyncEvent{
			SyncID:               e.SyncId,
			Timestamp:            e.StartedAt,
			Direction:            e.Direction,
			PeerNode:             e.PeerNode,
			State:                e.State,
			ResourcesTransferred: int(e.ResourcesSent + e.ResourcesReceived),
		}
	}

	total := 0
	totalPages := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &SyncHistoryResponse{
		Events:     events,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *syncAdapter) ExportBundle(ctx context.Context, req *BundleExportRequest) (*BundleExportResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ExportBundle(ctx, &syncv1.ExportBundleRequest{
		ResourceTypes: req.ResourceTypes,
		Since:         req.Since,
	})
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	var git *GitMeta
	if resp.Git != nil {
		git = &GitMeta{Commit: resp.Git.Commit, Message: resp.Git.Message}
	}
	return &BundleExportResponse{
		BundleData:    resp.BundleData,
		Format:        resp.Format,
		ResourceCount: int(resp.ResourceCount),
		Git:           git,
	}, nil
}

func (s *syncAdapter) ImportBundle(ctx context.Context, req *BundleImportRequest) (*BundleImportResponse, error) {
	c, err := s.client()
	if err != nil {
		return nil, err
	}

	resp, err := c.ImportBundle(ctx, &syncv1.ImportBundleRequest{
		BundleData: req.BundleData,
		Format:     req.Format,
		Author:     req.Author,
		NodeId:     req.NodeID,
		SiteId:     req.SiteID,
	})
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	var git *GitMeta
	if resp.Git != nil {
		git = &GitMeta{Commit: resp.Git.Commit, Message: resp.Git.Message}
	}
	return &BundleImportResponse{
		ResourcesImported: int(resp.ResourcesImported),
		ResourcesSkipped:  int(resp.ResourcesSkipped),
		Errors:            resp.Errors,
		Git:               git,
	}, nil
}
