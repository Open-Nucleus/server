package server

import (
	"context"
	"fmt"
	"time"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) RebuildIndex(ctx context.Context, req *patientv1.RebuildIndexRequest) (*patientv1.RebuildIndexResponse, error) {
	start := time.Now()
	count, head, err := s.pipeline.RebuildIndex()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "rebuild failed: %v", err)
	}
	duration := time.Since(start).Milliseconds()

	return &patientv1.RebuildIndexResponse{
		ResourcesIndexed: int32(count),
		GitHead:          head,
		DurationMs:       duration,
	}, nil
}

func (s *Server) CheckIndexHealth(ctx context.Context, req *patientv1.CheckIndexHealthRequest) (*patientv1.IndexHealthResponse, error) {
	indexHead, _ := s.idx.GetMeta("git_head")
	gitHead, err := s.git.Head()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get HEAD: %v", err)
	}

	indexCount, _ := s.idx.ResourceCount()

	healthy := indexHead == gitHead
	msg := "healthy"
	if !healthy {
		msg = fmt.Sprintf("index stale: index at %s, git at %s", indexHead, gitHead)
	}

	return &patientv1.IndexHealthResponse{
		Healthy:    healthy,
		IndexHead:  indexHead,
		GitHead:    gitHead,
		IndexCount: int32(indexCount),
		Message:    msg,
	}, nil
}

func (s *Server) ReindexResources(ctx context.Context, req *patientv1.ReindexRequest) (*patientv1.ReindexResponse, error) {
	var indexed, failed int
	var errors []string

	for _, path := range req.ResourcePaths {
		data, err := s.git.Read(path)
		if err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		rt, err := fhir.GetResourceType(data)
		if err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}

		patientID := extractPatientIDFromPath(path)
		head, _ := s.git.Head()

		if err := s.upsertFromJSON(rt, patientID, head, data); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		indexed++
	}

	return &patientv1.ReindexResponse{
		Indexed: int32(indexed),
		Failed:  int32(failed),
		Errors:  errors,
	}, nil
}

func (s *Server) upsertFromJSON(resourceType, patientID, commitHash string, fhirJSON []byte) error {
	switch resourceType {
	case fhir.ResourcePatient:
		row, err := fhir.ExtractPatientFields(fhirJSON, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertPatient(row)
	case fhir.ResourceEncounter:
		row, err := fhir.ExtractEncounterFields(fhirJSON, patientID, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertEncounter(row)
	case fhir.ResourceObservation:
		row, err := fhir.ExtractObservationFields(fhirJSON, patientID, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertObservation(row)
	case fhir.ResourceCondition:
		row, err := fhir.ExtractConditionFields(fhirJSON, patientID, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertCondition(row)
	case fhir.ResourceMedicationRequest:
		row, err := fhir.ExtractMedicationRequestFields(fhirJSON, patientID, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertMedicationRequest(row)
	case fhir.ResourceAllergyIntolerance:
		row, err := fhir.ExtractAllergyIntoleranceFields(fhirJSON, patientID, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertAllergyIntolerance(row)
	case fhir.ResourceFlag:
		row, err := fhir.ExtractFlagFields(fhirJSON, patientID, "", commitHash)
		if err != nil {
			return err
		}
		return s.idx.UpsertFlag(row)
	}
	return nil
}

func extractPatientIDFromPath(path string) string {
	if len(path) < 10 || path[:9] != "patients/" {
		return ""
	}
	rest := path[9:]
	for i, c := range rest {
		if c == '/' {
			return rest[:i]
		}
	}
	return ""
}
