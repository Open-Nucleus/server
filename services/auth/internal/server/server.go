package server

import (
	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/config"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the AuthService gRPC server.
type Server struct {
	authv1.UnimplementedAuthServiceServer
	cfg *config.Config
	svc *service.AuthService
}

// NewServer creates a new auth gRPC server.
func NewServer(cfg *config.Config, svc *service.AuthService) *Server {
	return &Server{cfg: cfg, svc: svc}
}

// mapError converts service errors to gRPC status errors.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "invalid", "expired", "verification failed", "not a refresh"):
		return status.Error(codes.Unauthenticated, msg)
	case contains(msg, "revoked", "denied"):
		return status.Error(codes.Unauthenticated, msg)
	case contains(msg, "insufficient", "scope_violation"):
		return status.Error(codes.PermissionDenied, msg)
	case contains(msg, "not found", "unknown device", "unknown role"):
		return status.Error(codes.NotFound, msg)
	case contains(msg, "bootstrap already", "invalid role", "invalid public key"):
		return status.Error(codes.FailedPrecondition, msg)
	case contains(msg, "blocked", "too many"):
		return status.Error(codes.ResourceExhausted, msg)
	default:
		return status.Error(codes.Internal, msg)
	}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
