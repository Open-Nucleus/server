package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListProcedures(ctx context.Context, req *patientv1.ListProceduresRequest) (*patientv1.ListProceduresResponse, error) {
	opts := paginationFromProto(req.Pagination)
	rows, pg, err := s.idx.ListProcedures(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListProceduresResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.Procedures = append(resp.Procedures, toFHIRResource(fhir.ResourceProcedure, row.ID, rowFHIRBytes(row.FHIRJson)))
	}
	return resp, nil
}

func (s *Server) GetProcedure(ctx context.Context, req *patientv1.GetProcedureRequest) (*patientv1.GetProcedureResponse, error) {
	row, err := s.idx.GetProcedure(req.PatientId, req.ProcedureId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "procedure %s not found", req.ProcedureId)
	}
	return &patientv1.GetProcedureResponse{
		Procedure: toFHIRResource(fhir.ResourceProcedure, row.ID, rowFHIRBytes(row.FHIRJson)),
	}, nil
}

func (s *Server) CreateProcedure(ctx context.Context, req *patientv1.CreateProcedureRequest) (*patientv1.CreateProcedureResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceProcedure, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateProcedureResponse{
		Procedure: toFHIRResource(fhir.ResourceProcedure, result.ResourceID, result.FHIRJson),
		Git:       toGitCommitInfo(result),
	}, nil
}
