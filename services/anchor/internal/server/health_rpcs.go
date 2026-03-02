package server

import (
	"context"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
)

func (s *Server) Health(_ context.Context, _ *anchorv1.HealthRequest) (*anchorv1.HealthResponse, error) {
	return &anchorv1.HealthResponse{
		Status:      "healthy",
		NodeDid:     s.svc.NodeDIDString(),
		Backend:     s.svc.BackendName(),
		AnchorCount: int32(s.svc.AnchorCount()),
		QueueDepth:  int32(s.svc.QueueDepth()),
	}, nil
}
