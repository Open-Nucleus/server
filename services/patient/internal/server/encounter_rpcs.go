package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListEncounters(ctx context.Context, req *patientv1.ListEncountersRequest) (*patientv1.ListEncountersResponse, error) {
	opts := paginationFromProto(req.Pagination)
	rows, pg, err := s.idx.ListEncounters(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListEncountersResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.Encounters = append(resp.Encounters, toFHIRResource(fhir.ResourceEncounter, row.ID, s.readFHIR(fhir.ResourceEncounter, req.PatientId, row.ID)))
	}
	return resp, nil
}

func (s *Server) GetEncounter(ctx context.Context, req *patientv1.GetEncounterRequest) (*patientv1.GetEncounterResponse, error) {
	row, err := s.idx.GetEncounter(req.PatientId, req.EncounterId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "encounter %s not found", req.EncounterId)
	}
	return &patientv1.GetEncounterResponse{
		Encounter: toFHIRResource(fhir.ResourceEncounter, row.ID, s.readFHIR(fhir.ResourceEncounter, req.PatientId, row.ID)),
	}, nil
}

func (s *Server) CreateEncounter(ctx context.Context, req *patientv1.CreateEncounterRequest) (*patientv1.CreateEncounterResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceEncounter, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateEncounterResponse{
		Encounter: toFHIRResource(fhir.ResourceEncounter, result.ResourceID, result.FHIRJson),
		Git:       toGitCommitInfo(result),
	}, nil
}

func (s *Server) UpdateEncounter(ctx context.Context, req *patientv1.UpdateEncounterRequest) (*patientv1.UpdateEncounterResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, fhir.ResourceEncounter, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdateEncounterResponse{
		Encounter: toFHIRResource(fhir.ResourceEncounter, result.ResourceID, result.FHIRJson),
		Git:       toGitCommitInfo(result),
	}, nil
}
