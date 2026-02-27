package service

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type conflictAdapter struct {
	pool *grpcclient.Pool
}

func NewConflictService(pool *grpcclient.Pool) ConflictService {
	return &conflictAdapter{pool: pool}
}

func (c *conflictAdapter) client() (syncv1.ConflictServiceClient, error) {
	conn, err := c.pool.Conn("sync")
	if err != nil {
		return nil, fmt.Errorf("conflict service unavailable: %w", err)
	}
	return syncv1.NewConflictServiceClient(conn), nil
}

func (c *conflictAdapter) ListConflicts(ctx context.Context, page, perPage int) (*ConflictListResponse, error) {
	cl, err := c.client()
	if err != nil {
		return nil, err
	}

	resp, err := cl.ListConflicts(ctx, &syncv1.ListConflictsRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}

	conflicts := make([]ConflictDetail, len(resp.Conflicts))
	for i, cf := range resp.Conflicts {
		conflicts[i] = conflictFromProto(cf)
	}

	total := 0
	totalPages := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.Total)
		totalPages = int(resp.Pagination.TotalPages)
	}

	return &ConflictListResponse{
		Conflicts:  conflicts,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (c *conflictAdapter) GetConflict(ctx context.Context, conflictID string) (*ConflictDetail, error) {
	cl, err := c.client()
	if err != nil {
		return nil, err
	}

	resp, err := cl.GetConflict(ctx, &syncv1.GetConflictRequest{ConflictId: conflictID})
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}

	detail := conflictFromProto(resp.Conflict)
	return &detail, nil
}

func (c *conflictAdapter) ResolveConflict(ctx context.Context, req *ResolveConflictRequest) (*ResolveConflictResponse, error) {
	cl, err := c.client()
	if err != nil {
		return nil, err
	}

	resp, err := cl.ResolveConflict(ctx, &syncv1.ResolveConflictRequest{
		ConflictId:     req.ConflictID,
		Resolution:     req.Resolution,
		MergedResource: req.MergedResource,
		Author:         req.Author,
	})
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}

	var git *GitMeta
	if resp.Git != nil {
		git = &GitMeta{Commit: resp.Git.Commit, Message: resp.Git.Message}
	}
	return &ResolveConflictResponse{Git: git}, nil
}

func (c *conflictAdapter) DeferConflict(ctx context.Context, req *DeferConflictRequest) (*DeferConflictResponse, error) {
	cl, err := c.client()
	if err != nil {
		return nil, err
	}

	resp, err := cl.DeferConflict(ctx, &syncv1.DeferConflictRequest{
		ConflictId: req.ConflictID,
		Reason:     req.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("conflict: %w", err)
	}

	return &DeferConflictResponse{Status: resp.Status}, nil
}

func conflictFromProto(cf *syncv1.Conflict) ConflictDetail {
	return ConflictDetail{
		ID:           cf.Id,
		ResourceType: cf.ResourceType,
		ResourceID:   cf.ResourceId,
		Status:       cf.Status,
		DetectedAt:   cf.DetectedAt,
		LocalNode:    cf.LocalNode,
		RemoteNode:   cf.RemoteNode,
	}
}
