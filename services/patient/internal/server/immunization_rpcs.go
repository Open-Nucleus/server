package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListImmunizations(ctx context.Context, req *patientv1.ListImmunizationsRequest) (*patientv1.ListImmunizationsResponse, error) {
	opts := paginationFromProto(req.Pagination)
	rows, pg, err := s.idx.ListImmunizations(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListImmunizationsResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.Immunizations = append(resp.Immunizations, toFHIRResource(fhir.ResourceImmunization, row.ID, s.readFHIR(fhir.ResourceImmunization, req.PatientId, row.ID)))
	}
	return resp, nil
}

func (s *Server) GetImmunization(ctx context.Context, req *patientv1.GetImmunizationRequest) (*patientv1.GetImmunizationResponse, error) {
	row, err := s.idx.GetImmunization(req.PatientId, req.ImmunizationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "immunization %s not found", req.ImmunizationId)
	}
	return &patientv1.GetImmunizationResponse{
		Immunization: toFHIRResource(fhir.ResourceImmunization, row.ID, s.readFHIR(fhir.ResourceImmunization, req.PatientId, row.ID)),
	}, nil
}

func (s *Server) CreateImmunization(ctx context.Context, req *patientv1.CreateImmunizationRequest) (*patientv1.CreateImmunizationResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceImmunization, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateImmunizationResponse{
		Immunization: toFHIRResource(fhir.ResourceImmunization, result.ResourceID, result.FHIRJson),
		Git:          toGitCommitInfo(result),
	}, nil
}
