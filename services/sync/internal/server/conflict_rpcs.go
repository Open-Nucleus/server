package server

import (
	"context"
	"encoding/json"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/store"
)

func (s *Server) ListConflicts(_ context.Context, req *syncv1.ListConflictsRequest) (*syncv1.ListConflictsResponse, error) {
	page := int(req.GetPagination().GetPage())
	perPage := int(req.GetPagination().GetPerPage())
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	conflicts, total, err := s.conflicts.List(req.Status, req.Level, perPage, offset)
	if err != nil {
		return nil, mapError(err)
	}

	protoConflicts := make([]*syncv1.Conflict, len(conflicts))
	for i, c := range conflicts {
		protoConflicts[i] = conflictToProto(c)
	}

	totalPages := (total + perPage - 1) / perPage
	return &syncv1.ListConflictsResponse{
		Conflicts: protoConflicts,
		Pagination: &commonv1.PaginationResponse{
			Page:       int32(page),
			PerPage:    int32(perPage),
			Total:      int32(total),
			TotalPages: int32(totalPages),
		},
	}, nil
}

func (s *Server) GetConflict(_ context.Context, req *syncv1.GetConflictRequest) (*syncv1.GetConflictResponse, error) {
	c, err := s.conflicts.Get(req.ConflictId)
	if err != nil {
		return nil, mapError(err)
	}
	return &syncv1.GetConflictResponse{Conflict: conflictToProto(c)}, nil
}

func (s *Server) ResolveConflict(_ context.Context, req *syncv1.ResolveConflictRequest) (*syncv1.ResolveConflictResponse, error) {
	err := s.conflicts.Resolve(req.ConflictId, req.Resolution, req.Author, req.MergedResource)
	if err != nil {
		return nil, mapError(err)
	}

	s.eventBus.Publish(service.Event{
		Type: service.EventConflictResolved,
		Payload: map[string]string{
			"conflict_id": req.ConflictId,
			"resolution":  req.Resolution,
		},
	})

	return &syncv1.ResolveConflictResponse{}, nil
}

func (s *Server) DeferConflict(_ context.Context, req *syncv1.DeferConflictRequest) (*syncv1.DeferConflictResponse, error) {
	err := s.conflicts.Defer(req.ConflictId, req.Reason)
	if err != nil {
		return nil, mapError(err)
	}
	return &syncv1.DeferConflictResponse{Status: "deferred"}, nil
}

func conflictToProto(c *store.ConflictRecord) *syncv1.Conflict {
	proto := &syncv1.Conflict{
		Id:            c.ID,
		ResourceType:  c.ResourceType,
		ResourceId:    c.ResourceID,
		Status:        c.Status,
		Level:         c.Level,
		DetectedAt:    c.DetectedAt,
		LocalNode:     c.LocalNode,
		RemoteNode:    c.RemoteNode,
		PatientId:     c.PatientID,
		ChangedFields: c.ChangedFields,
		PeerSiteId:    c.PeerSiteID,
		MergedVersion: c.MergedVersion,
		Reason:        c.Reason,
	}

	if c.LocalVersion != nil {
		var lv map[string]any
		if json.Unmarshal(c.LocalVersion, &lv) == nil {
			rt, _ := lv["resourceType"].(string)
			id, _ := lv["id"].(string)
			proto.LocalVersion = &commonv1.FHIRResource{
				ResourceType: rt,
				Id:           id,
				JsonPayload:  c.LocalVersion,
			}
		}
	}
	if c.RemoteVersion != nil {
		var rv map[string]any
		if json.Unmarshal(c.RemoteVersion, &rv) == nil {
			rt, _ := rv["resourceType"].(string)
			id, _ := rv["id"].(string)
			proto.RemoteVersion = &commonv1.FHIRResource{
				ResourceType: rt,
				Id:           id,
				JsonPayload:  c.RemoteVersion,
			}
		}
	}

	return proto
}
