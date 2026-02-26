package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) ListMedicationRequests(ctx context.Context, req *patientv1.ListMedicationRequestsRequest) (*patientv1.ListMedicationRequestsResponse, error) {
	opts := paginationFromProto(req.Pagination)
	rows, pg, err := s.idx.ListMedicationRequests(req.PatientId, opts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}

	resp := &patientv1.ListMedicationRequestsResponse{Pagination: paginationToProto(pg)}
	for _, row := range rows {
		resp.MedicationRequests = append(resp.MedicationRequests, toFHIRResource(fhir.ResourceMedicationRequest, row.ID, rowFHIRBytes(row.FHIRJson)))
	}
	return resp, nil
}

func (s *Server) GetMedicationRequest(ctx context.Context, req *patientv1.GetMedicationRequestRequest) (*patientv1.GetMedicationRequestResponse, error) {
	row, err := s.idx.GetMedicationRequest(req.PatientId, req.MedicationRequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "query failed: %v", err)
	}
	if row == nil {
		return nil, status.Errorf(codes.NotFound, "medication request %s not found", req.MedicationRequestId)
	}
	return &patientv1.GetMedicationRequestResponse{
		MedicationRequest: toFHIRResource(fhir.ResourceMedicationRequest, row.ID, rowFHIRBytes(row.FHIRJson)),
	}, nil
}

func (s *Server) CreateMedicationRequest(ctx context.Context, req *patientv1.CreateMedicationRequestRequest) (*patientv1.CreateMedicationRequestResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpCreate, fhir.ResourceMedicationRequest, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateMedicationRequestResponse{
		MedicationRequest: toFHIRResource(fhir.ResourceMedicationRequest, result.ResourceID, result.FHIRJson),
		Git:               toGitCommitInfo(result),
	}, nil
}

func (s *Server) UpdateMedicationRequest(ctx context.Context, req *patientv1.UpdateMedicationRequestRequest) (*patientv1.UpdateMedicationRequestResponse, error) {
	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, fhir.ResourceMedicationRequest, req.PatientId, req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdateMedicationRequestResponse{
		MedicationRequest: toFHIRResource(fhir.ResourceMedicationRequest, result.ResourceID, result.FHIRJson),
		Git:               toGitCommitInfo(result),
	}, nil
}
