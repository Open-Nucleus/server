package server

import (
	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the FormularyService gRPC server.
type Server struct {
	formularyv1.UnimplementedFormularyServiceServer
	svc *service.FormularyService
}

// NewServer creates a new gRPC server for the formulary service.
func NewServer(svc *service.FormularyService) *Server {
	return &Server{svc: svc}
}

// mapError converts service errors to gRPC status codes.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if contains(msg, "not found") {
		return status.Errorf(codes.NotFound, "%s", msg)
	}
	if contains(msg, "not configured") {
		return status.Errorf(codes.Unimplemented, "%s", msg)
	}
	return status.Errorf(codes.Internal, "%s", msg)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
