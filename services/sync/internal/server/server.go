package server

import (
	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/config"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the SyncService and ConflictService gRPC servers.
type Server struct {
	syncv1.UnimplementedSyncServiceServer
	syncv1.UnimplementedConflictServiceServer
	syncv1.UnimplementedNodeSyncServiceServer
	cfg       *config.Config
	engine    *service.SyncEngine
	conflicts *store.ConflictStore
	history   *store.HistoryStore
	peers     *store.PeerStore
	eventBus  *service.EventBus
}

// NewServer creates a new sync gRPC server.
func NewServer(
	cfg *config.Config,
	engine *service.SyncEngine,
	conflicts *store.ConflictStore,
	history *store.HistoryStore,
	peers *store.PeerStore,
	eventBus *service.EventBus,
) *Server {
	return &Server{
		cfg:       cfg,
		engine:    engine,
		conflicts: conflicts,
		history:   history,
		peers:     peers,
		eventBus:  eventBus,
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "not found"):
		return status.Error(codes.NotFound, msg)
	case contains(msg, "already in progress"):
		return status.Error(codes.FailedPrecondition, msg)
	case contains(msg, "cancelled"):
		return status.Error(codes.Aborted, msg)
	default:
		return status.Error(codes.Internal, msg)
	}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}
