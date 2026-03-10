package server

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// validGenericTypes are the resource types supported by the generic CRUD RPCs.
var validGenericTypes = map[string]bool{
	fhir.ResourcePractitioner:  true,
	fhir.ResourceOrganization:  true,
	fhir.ResourceLocation:      true,
	fhir.ResourceMeasureReport: true,
}

func (s *Server) CreateResource(ctx context.Context, req *patientv1.CreateResourceRequest) (*patientv1.CreateResourceResponse, error) {
	if !validGenericTypes[req.ResourceType] {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource type: %s", req.ResourceType)
	}

	result, err := s.pipeline.Write(ctx, fhir.OpCreate, req.ResourceType, "", req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.CreateResourceResponse{
		Resource: toFHIRResource(req.ResourceType, result.ResourceID, result.FHIRJson),
		Git:      toGitCommitInfo(result),
	}, nil
}

func (s *Server) GetResource(ctx context.Context, req *patientv1.GetResourceRequest) (*patientv1.GetResourceResponse, error) {
	if !fhir.IsKnownResource(req.ResourceType) {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource type: %s", req.ResourceType)
	}

	var patientID string
	var resourceID string

	switch req.ResourceType {
	case fhir.ResourcePatient:
		row, err := s.idx.GetPatient(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		resourceID = row.ID
	case fhir.ResourceEncounter:
		row, err := s.idx.GetEncounterByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceObservation:
		row, err := s.idx.GetObservationByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceCondition:
		row, err := s.idx.GetConditionByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceMedicationRequest:
		row, err := s.idx.GetMedicationRequestByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceAllergyIntolerance:
		row, err := s.idx.GetAllergyIntoleranceByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceImmunization:
		row, err := s.idx.GetImmunizationByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceProcedure:
		row, err := s.idx.GetProcedureByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourceFlag:
		row, err := s.idx.GetFlagByID(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		patientID = row.PatientID
		resourceID = row.ID
	case fhir.ResourcePractitioner:
		row, err := s.idx.GetPractitioner(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		resourceID = row.ID
	case fhir.ResourceOrganization:
		row, err := s.idx.GetOrganization(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		resourceID = row.ID
	case fhir.ResourceLocation:
		row, err := s.idx.GetLocation(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		resourceID = row.ID
	case fhir.ResourceMeasureReport:
		row, err := s.idx.GetMeasureReport(req.ResourceId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		if row == nil {
			return nil, status.Errorf(codes.NotFound, "%s %s not found", req.ResourceType, req.ResourceId)
		}
		resourceID = row.ID
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource type: %s", req.ResourceType)
	}

	return &patientv1.GetResourceResponse{
		Resource: toFHIRResource(req.ResourceType, resourceID, s.readFHIR(req.ResourceType, patientID, resourceID)),
	}, nil
}

func (s *Server) ListResources(ctx context.Context, req *patientv1.ListResourcesRequest) (*patientv1.ListResourcesResponse, error) {
	if !validGenericTypes[req.ResourceType] {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource type: %s", req.ResourceType)
	}

	opts := paginationFromProto(req.Pagination)
	var resources []*commonv1.FHIRResource
	var pg *fhir.Pagination

	switch req.ResourceType {
	case fhir.ResourcePractitioner:
		rows, p, err := s.idx.ListPractitioners(opts)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		pg = p
		for _, row := range rows {
			resources = append(resources, toFHIRResource(fhir.ResourcePractitioner, row.ID, s.readFHIR(fhir.ResourcePractitioner, "", row.ID)))
		}
	case fhir.ResourceOrganization:
		rows, p, err := s.idx.ListOrganizations(opts)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		pg = p
		for _, row := range rows {
			resources = append(resources, toFHIRResource(fhir.ResourceOrganization, row.ID, s.readFHIR(fhir.ResourceOrganization, "", row.ID)))
		}
	case fhir.ResourceLocation:
		rows, p, err := s.idx.ListLocations(opts)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		pg = p
		for _, row := range rows {
			resources = append(resources, toFHIRResource(fhir.ResourceLocation, row.ID, s.readFHIR(fhir.ResourceLocation, "", row.ID)))
		}
	case fhir.ResourceMeasureReport:
		rows, p, err := s.idx.ListMeasureReports(opts)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "query failed: %v", err)
		}
		pg = p
		for _, row := range rows {
			resources = append(resources, toFHIRResource(fhir.ResourceMeasureReport, row.ID, s.readFHIR(fhir.ResourceMeasureReport, "", row.ID)))
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported: %s", req.ResourceType)
	}

	return &patientv1.ListResourcesResponse{
		Resources:  resources,
		Pagination: paginationToProto(pg),
	}, nil
}

func (s *Server) UpdateResource(ctx context.Context, req *patientv1.UpdateResourceRequest) (*patientv1.UpdateResourceResponse, error) {
	if !validGenericTypes[req.ResourceType] {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource type: %s", req.ResourceType)
	}

	result, err := s.pipeline.Write(ctx, fhir.OpUpdate, req.ResourceType, "", req.FhirJson, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	return &patientv1.UpdateResourceResponse{
		Resource: toFHIRResource(req.ResourceType, result.ResourceID, result.FHIRJson),
		Git:      toGitCommitInfo(result),
	}, nil
}

func (s *Server) DeleteResource(ctx context.Context, req *patientv1.DeleteResourceRequest) (*patientv1.DeleteResourceResponse, error) {
	if !validGenericTypes[req.ResourceType] {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported resource type: %s", req.ResourceType)
	}

	result, err := s.pipeline.Delete(ctx, req.ResourceType, "", req.ResourceId, mutCtxFromProto(req.Context))
	if err != nil {
		return nil, mapError(err)
	}
	_ = fmt.Sprintf("deleted %s/%s", req.ResourceType, req.ResourceId)
	return &patientv1.DeleteResourceResponse{
		Git: toGitCommitInfo(result),
	}, nil
}
