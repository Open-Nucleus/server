package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListAllergyIntolerances(ctx context.Context, req *patientv1.ListAllergyIntolerancesRequest) (*patientv1.ListAllergyIntolerancesResponse, error) {
	opts := paginationFromProto(req.Pagination)
	rows, pg, err := s.idx.ListAllergyIntolerances(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListAllergyIntolerancesResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.AllergyIntolerances = append(resp.AllergyIntolerances, toFHIRResource(fhir.ResourceAllergyIntolerance, row.ID, s.readFHIR(fhir.ResourceAllergyIntolerance, req.PatientId, row.ID)))
	}
	return resp, nil
}

func (s *Server) GetAllergyIntolerance(ctx context.Context, req *patientv1.GetAllergyIntoleranceRequest) (*patientv1.GetAllergyIntoleranceResponse, error) {
	row, err := s.idx.GetAllergyIntolerance(req.PatientId, req.AllergyIntoleranceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "allergy intolerance %s not found", req.AllergyIntoleranceId)
	}
	return &patientv1.GetAllergyIntoleranceResponse{
		AllergyIntolerance: toFHIRResource(fhir.ResourceAllergyIntolerance, row.ID, s.readFHIR(fhir.ResourceAllergyIntolerance, req.PatientId, row.ID)),
	}, nil
}

func (s *Server) CreateAllergyIntolerance(ctx context.Context, req *patientv1.CreateAllergyIntoleranceRequest) (*patientv1.CreateAllergyIntoleranceResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceAllergyIntolerance, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateAllergyIntoleranceResponse{
		AllergyIntolerance: toFHIRResource(fhir.ResourceAllergyIntolerance, result.ResourceID, result.FHIRJson),
		Git:                toGitCommitInfo(result),
	}, nil
}

func (s *Server) UpdateAllergyIntolerance(ctx context.Context, req *patientv1.UpdateAllergyIntoleranceRequest) (*patientv1.UpdateAllergyIntoleranceResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, fhir.ResourceAllergyIntolerance, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdateAllergyIntoleranceResponse{
		AllergyIntolerance: toFHIRResource(fhir.ResourceAllergyIntolerance, result.ResourceID, result.FHIRJson),
		Git:                toGitCommitInfo(result),
	}, nil
}
