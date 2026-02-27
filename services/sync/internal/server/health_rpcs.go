package server

import (
	"context"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
)

func (s *Server) Health(_ context.Context, _ *syncv1.HealthRequest) (*syncv1.HealthResponse, error) {
	return &syncv1.HealthResponse{
		Status:        "healthy",
		Version:       "0.4.0",
		UptimeSeconds: s.engine.UptimeSeconds(),
	}, nil
}
