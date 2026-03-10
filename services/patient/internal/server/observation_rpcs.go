package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListObservations(ctx context.Context, req *patientv1.ListObservationsRequest) (*patientv1.ListObservationsResponse, error) {
	opts := sqliteindex.ObservationListOpts{
		PaginationOpts: paginationFromProto(req.Pagination),
		Code:           req.Code,
		Category:       req.Category,
		DateFrom:       req.DateFrom,
		DateTo:         req.DateTo,
		EncounterID:    req.EncounterId,
	}
	rows, pg, err := s.idx.ListObservations(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListObservationsResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.Observations = append(resp.Observations, toFHIRResource(fhir.ResourceObservation, row.ID, s.readFHIR(fhir.ResourceObservation, req.PatientId, row.ID)))
	}
	return resp, nil
}

func (s *Server) GetObservation(ctx context.Context, req *patientv1.GetObservationRequest) (*patientv1.GetObservationResponse, error) {
	row, err := s.idx.GetObservation(req.PatientId, req.ObservationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "observation %s not found", req.ObservationId)
	}
	return &patientv1.GetObservationResponse{
		Observation: toFHIRResource(fhir.ResourceObservation, row.ID, s.readFHIR(fhir.ResourceObservation, req.PatientId, row.ID)),
	}, nil
}

func (s *Server) CreateObservation(ctx context.Context, req *patientv1.CreateObservationRequest) (*patientv1.CreateObservationResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceObservation, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateObservationResponse{
		Observation: toFHIRResource(fhir.ResourceObservation, result.ResourceID, result.FHIRJson),
		Git:         toGitCommitInfo(result),
	}, nil
}
