package server

import (
	"context"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Health(ctx context.Context, req *patientv1.HealthRequest) (*patientv1.HealthResponse, error) {
	gitHead, err := s.git.Head()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "git HEAD: %v", err)
	}

	indexHead, _ := s.idx.GetMeta("git_head")
	healthy := indexHead == gitHead || gitHead == "" // empty repo is healthy

	patientCount, _ := s.idx.ResourceCount()

	return &patientv1.HealthResponse{
		Status:       "ok",
		GitHead:      gitHead,
		PatientCount: int32(patientCount),
		IndexHealthy: healthy,
	}, nil
}
