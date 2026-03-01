package server

import (
	"context"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
)

func (s *Server) Health(_ context.Context, _ *formularyv1.HealthRequest) (*formularyv1.HealthResponse, error) {
	info := s.svc.GetFormularyInfo()
	return &formularyv1.HealthResponse{
		Status:             "healthy",
		Version:            info.Version,
		MedicationsLoaded:  int32(info.TotalMedications),
		InteractionsLoaded: int32(info.TotalInteractions),
	}, nil
}
