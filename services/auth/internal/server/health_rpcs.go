package server

import (
	"context"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
)

func (s *Server) Health(_ context.Context, _ *authv1.HealthRequest) (*authv1.HealthResponse, error) {
	return &authv1.HealthResponse{
		Status:        "healthy",
		Version:       "0.4.0",
		UptimeSeconds: s.svc.UptimeSeconds(),
	}, nil
}
