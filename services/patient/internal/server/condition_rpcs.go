package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListConditions(ctx context.Context, req *patientv1.ListConditionsRequest) (*patientv1.ListConditionsResponse, error) {
	opts := sqliteindex.ConditionListOpts{
		PaginationOpts: paginationFromProto(req.Pagination),
		ClinicalStatus: req.ClinicalStatus,
		Category:       req.Category,
		Code:           req.Code,
	}
	rows, pg, err := s.idx.ListConditions(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListConditionsResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.Conditions = append(resp.Conditions, toFHIRResource(fhir.ResourceCondition, row.ID, s.readFHIR(fhir.ResourceCondition, req.PatientId, row.ID)))
	}
	return resp, nil
}

func (s *Server) GetCondition(ctx context.Context, req *patientv1.GetConditionRequest) (*patientv1.GetConditionResponse, error) {
	row, err := s.idx.GetCondition(req.PatientId, req.ConditionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "condition %s not found", req.ConditionId)
	}
	return &patientv1.GetConditionResponse{
		Condition: toFHIRResource(fhir.ResourceCondition, row.ID, s.readFHIR(fhir.ResourceCondition, req.PatientId, row.ID)),
	}, nil
}

func (s *Server) CreateCondition(ctx context.Context, req *patientv1.CreateConditionRequest) (*patientv1.CreateConditionResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceCondition, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateConditionResponse{
		Condition: toFHIRResource(fhir.ResourceCondition, result.ResourceID, result.FHIRJson),
		Git:       toGitCommitInfo(result),
	}, nil
}

func (s *Server) UpdateCondition(ctx context.Context, req *patientv1.UpdateConditionRequest) (*patientv1.UpdateConditionResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, fhir.ResourceCondition, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdateConditionResponse{
		Condition: toFHIRResource(fhir.ResourceCondition, result.ResourceID, result.FHIRJson),
		Git:       toGitCommitInfo(result),
	}, nil
}
