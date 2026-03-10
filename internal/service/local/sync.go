package local

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/services/sync/syncservice"
)

// --------------------------------------------------------------------------
// SyncService local adapter
// --------------------------------------------------------------------------

// syncSvc implements service.SyncService by calling the SyncEngine, HistoryStore,
// and PeerStore directly (no gRPC).
type syncSvc struct {
	engine  *syncservice.SyncEngine
	history *syncservice.HistoryStore
	peers   *syncservice.PeerStore
}

// NewSyncService creates a local adapter for sync operations.
func NewSyncService(
	engine *syncservice.SyncEngine,
	history *syncservice.HistoryStore,
	peers *syncservice.PeerStore,
) service.SyncService {
	return &syncSvc{engine: engine, history: history, peers: peers}
}

func (s *syncSvc) GetStatus(_ context.Context) (*service.SyncStatusResponse, error) {
	state, _, _, pending := s.engine.GetStatus()
	return &service.SyncStatusResponse{
		State:          string(state),
		LastSync:       s.engine.LastSyncTime(),
		PendingChanges: pending,
		NodeID:         s.engine.NodeID(),
		SiteID:         s.engine.SiteID(),
	}, nil
}

func (s *syncSvc) ListPeers(_ context.Context) (*service.SyncPeersResponse, error) {
	records, err := s.peers.List()
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	peers := make([]service.PeerInfo, len(records))
	for i, p := range records {
		peers[i] = service.PeerInfo{
			NodeID:   p.NodeID,
			SiteID:   p.SiteID,
			LastSeen: p.LastSeen,
			State:    "offline",
		}
	}
	return &service.SyncPeersResponse{Peers: peers}, nil
}

func (s *syncSvc) TriggerSync(_ context.Context, targetNode string) (*service.SyncTriggerResponse, error) {
	syncID, err := s.engine.TriggerSync(targetNode)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}
	return &service.SyncTriggerResponse{
		SyncID: syncID,
		State:  "completed",
	}, nil
}

func (s *syncSvc) GetHistory(_ context.Context, page, perPage int) (*service.SyncHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	entries, total, err := s.history.List(perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}

	events := make([]service.SyncEvent, len(entries))
	for i, e := range entries {
		events[i] = service.SyncEvent{
			SyncID:               e.ID,
			Timestamp:            e.StartedAt,
			Direction:            e.Direction,
			PeerNode:             e.PeerNode,
			State:                e.State,
			ResourcesTransferred: e.ResourcesSent + e.ResourcesReceived,
		}
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return &service.SyncHistoryResponse{
		Events:     events,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (s *syncSvc) ExportBundle(_ context.Context, req *service.BundleExportRequest) (*service.BundleExportResponse, error) {
	data, count, err := s.engine.ExportBundle(req.ResourceTypes, req.Since)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}
	return &service.BundleExportResponse{
		BundleData:    data,
		Format:        "nucleus-bundle-v1",
		ResourceCount: count,
	}, nil
}

func (s *syncSvc) ImportBundle(_ context.Context, req *service.BundleImportRequest) (*service.BundleImportResponse, error) {
	imported, skipped, errors, err := s.engine.ImportBundle(req.BundleData)
	if err != nil {
		return nil, fmt.Errorf("sync: %w", err)
	}
	return &service.BundleImportResponse{
		ResourcesImported: imported,
		ResourcesSkipped:  skipped,
		Errors:            errors,
	}, nil
}

// --------------------------------------------------------------------------
// ConflictService local adapter
// --------------------------------------------------------------------------

// conflictSvc implements service.ConflictService by calling the ConflictStore
// directly (no gRPC).
type conflictSvc struct {
	conflicts *syncservice.ConflictStore
	eventBus  *syncservice.EventBus
}

// NewConflictService creates a local adapter for conflict resolution.
func NewConflictService(
	conflicts *syncservice.ConflictStore,
	eventBus *syncservice.EventBus,
) service.ConflictService {
	return &conflictSvc{conflicts: conflicts, eventBus: eventBus}
}

func (c *conflictSvc) ListConflicts(_ context.Context, page, perPage int) (*service.ConflictListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	records, total, err := c.conflicts.List("", "", perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}

	conflicts := make([]service.ConflictDetail, len(records))
	for i, cr := range records {
		conflicts[i] = conflictRecordToDTO(cr)
	}

	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}

	return &service.ConflictListResponse{
		Conflicts:  conflicts,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (c *conflictSvc) GetConflict(_ context.Context, conflictID string) (*service.ConflictDetail, error) {
	record, err := c.conflicts.Get(conflictID)
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}
	detail := conflictRecordToDTO(record)
	return &detail, nil
}

func (c *conflictSvc) ResolveConflict(_ context.Context, req *service.ResolveConflictRequest) (*service.ResolveConflictResponse, error) {
	err := c.conflicts.Resolve(req.ConflictID, req.Resolution, req.Author, req.MergedResource)
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}

	if c.eventBus != nil {
		c.eventBus.Publish(syncservice.Event{
			Type: syncservice.EventConflictResolved,
			Payload: map[string]string{
				"conflict_id": req.ConflictID,
				"resolution":  req.Resolution,
			},
		})
	}

	return &service.ResolveConflictResponse{}, nil
}

func (c *conflictSvc) DeferConflict(_ context.Context, req *service.DeferConflictRequest) (*service.DeferConflictResponse, error) {
	err := c.conflicts.Defer(req.ConflictID, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}
	return &service.DeferConflictResponse{Status: "deferred"}, nil
}

// conflictRecordToDTO converts a store.ConflictRecord to a service.ConflictDetail.
// This mirrors the field mapping performed by the gRPC adapter's conflictFromProto.
func conflictRecordToDTO(cr *syncservice.ConflictRecord) service.ConflictDetail {
	detail := service.ConflictDetail{
		ID:           cr.ID,
		ResourceType: cr.ResourceType,
		ResourceID:   cr.ResourceID,
		Status:       cr.Status,
		DetectedAt:   cr.DetectedAt,
		LocalNode:    cr.LocalNode,
		RemoteNode:   cr.RemoteNode,
	}

	// Parse local/remote versions into typed maps, matching the gRPC adapter
	// which exposes them as proto FHIRResource (then decoded by the gateway).
	if cr.LocalVersion != nil {
		var m map[string]any
		if json.Unmarshal(cr.LocalVersion, &m) == nil {
			detail.LocalVersion = m
		}
	}
	if cr.RemoteVersion != nil {
		var m map[string]any
		if json.Unmarshal(cr.RemoteVersion, &m) == nil {
			detail.RemoteVersion = m
		}
	}

	return detail
}
