package server

import (
	"context"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
)

func (s *Server) ListBackends(_ context.Context, _ *anchorv1.ListBackendsRequest) (*anchorv1.ListBackendsResponse, error) {
	backends := s.svc.ListBackends()
	protoBackends := make([]*anchorv1.BackendInfo, 0, len(backends))
	for _, b := range backends {
		protoBackends = append(protoBackends, &anchorv1.BackendInfo{
			Name:        b.Name,
			Available:   b.Available,
			Description: b.Description,
		})
	}
	return &anchorv1.ListBackendsResponse{
		Backends: protoBackends,
	}, nil
}

func (s *Server) GetBackendStatus(_ context.Context, req *anchorv1.GetBackendStatusRequest) (*anchorv1.GetBackendStatusResponse, error) {
	result, err := s.svc.GetBackendStatus(req.Name)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.GetBackendStatusResponse{
		Name:           result.Name,
		Available:      result.Available,
		Description:    result.Description,
		AnchoredCount:  int32(result.AnchoredCount),
		LastAnchorTime: result.LastAnchorTime,
	}, nil
}

func (s *Server) GetQueueStatus(_ context.Context, _ *anchorv1.GetQueueStatusRequest) (*anchorv1.GetQueueStatusResponse, error) {
	result, err := s.svc.GetQueueStatus()
	if err != nil {
		return nil, mapError(err)
	}

	entries := make([]*anchorv1.QueueEntry, 0, len(result.Entries))
	for _, e := range result.Entries {
		entries = append(entries, &anchorv1.QueueEntry{
			AnchorId:   e.AnchorID,
			MerkleRoot: e.MerkleRoot,
			GitHead:    e.GitHead,
			EnqueuedAt: e.EnqueuedAt,
			State:      e.State,
		})
	}

	return &anchorv1.GetQueueStatusResponse{
		Pending:        int32(result.Pending),
		TotalProcessed: int32(result.TotalProcessed),
		Entries:        entries,
	}, nil
}
