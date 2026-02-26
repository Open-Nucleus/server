package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

func (s *Server) CreateFlag(ctx context.Context, req *patientv1.CreateFlagRequest) (*patientv1.CreateFlagResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceFlag, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateFlagResponse{
		Flag: toFHIRResource(fhir.ResourceFlag, result.ResourceID, result.FHIRJson),
		Git:  toGitCommitInfo(result),
	}, nil
}

func (s *Server) UpdateFlag(ctx context.Context, req *patientv1.UpdateFlagRequest) (*patientv1.UpdateFlagResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, fhir.ResourceFlag, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdateFlagResponse{
		Flag: toFHIRResource(fhir.ResourceFlag, result.ResourceID, result.FHIRJson),
		Git:  toGitCommitInfo(result),
	}, nil
}
