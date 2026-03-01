package server

import (
	"context"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
)

func (s *Server) GetFormularyInfo(_ context.Context, _ *formularyv1.GetFormularyInfoRequest) (*formularyv1.GetFormularyInfoResponse, error) {
	info := s.svc.GetFormularyInfo()
	return &formularyv1.GetFormularyInfoResponse{
		Version:               info.Version,
		TotalMedications:      int32(info.TotalMedications),
		TotalInteractions:     int32(info.TotalInteractions),
		LastUpdated:           info.LastUpdated,
		Categories:            info.Categories,
		DosingEngineAvailable: info.DosingEngineAvailable,
	}, nil
}
