package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateBatch(ctx context.Context, req *patientv1.CreateBatchRequest) (*patientv1.BatchResponse, error) {
	var items []pipeline.BatchItem
	for _, r := range req.Resources {
		items = append(items, pipeline.BatchItem{
			ResourceType: r.ResourceType,
			FHIRJson:     r.JsonPayload,
		})
	}

	result, err := s.pipeline.WriteBatch(ctx, req.PatientId, items, mutCtxFromProto(req.Context), req.Atomic)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &patientv1.BatchResponse{}
	for _, r := range result.Results {
		resp.Results = append(resp.Results, &patientv1.BatchItemResult{
			ResourceType: r.ResourceType,
			ResourceId:   r.ResourceID,
			Success:      r.Success,
			Error:        r.Error,
		})
	}
	if result.CommitHash != "" {
		resp.Git = &patientv1.GitCommitInfo{
			CommitHash: result.CommitHash,
			Timestamp:  timestamppb.New(result.Timestamp),
		}
	}
	return resp, nil
}
