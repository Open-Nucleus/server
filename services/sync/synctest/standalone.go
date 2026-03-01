package synctest

import (
	"database/sql"
	"fmt"
	"net"
	"time"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/config"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/server"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/store"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

// StartStandalone boots an in-process Sync Service without requiring *testing.T.
// Returns the environment and a cleanup function.
func StartStandalone(tmpDir string) (*Env, func(), error) {
	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	cfg := &config.Config{
		GRPCPort: 0,
		Git: config.GitConfig{
			RepoPath:    tmpDir + "/sync-data",
			AuthorName:  "smoke",
			AuthorEmail: "smoke@test.local",
		},
		Sync: config.SyncConfig{
			MaxConcurrent:    1,
			Cooldown:         time.Second,
			HandshakeTimeout: 5 * time.Second,
			TransferTimeout:  30 * time.Second,
			ChunkSize:        65536,
		},
		History: config.HistoryConfig{
			DBPath:     tmpDir + "/sync.db",
			MaxEntries: 1000,
		},
		Events: config.EventsConfig{BufferSize: 100},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	if err != nil {
		return nil, cleanup, fmt.Errorf("sync gitstore: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.History.DBPath)
	if err != nil {
		return nil, cleanup, fmt.Errorf("sync sqlite: %w", err)
	}
	cleanups = append(cleanups, func() { db.Close() })

	if err := store.InitSchema(db); err != nil {
		return nil, cleanup, fmt.Errorf("sync schema: %w", err)
	}

	conflictStore := store.NewConflictStore(db)
	historyStore := store.NewHistoryStore(db, cfg.History.MaxEntries)
	peerStore := store.NewPeerStore(db)
	mergeDriver := merge.NewDriver(nil)
	eventBus := service.NewEventBus(cfg.Events.BufferSize)

	engine := service.NewSyncEngine(cfg, git, conflictStore, historyStore, peerStore, mergeDriver, eventBus, "smoke-node", "smoke-site")
	engine.RegisterTransport(&transport.StubAdapter{TransportName: "local_network"})

	srv := server.NewServer(cfg, engine, conflictStore, historyStore, peerStore, eventBus)
	grpcServer := grpc.NewServer()
	syncv1.RegisterSyncServiceServer(grpcServer, srv)
	syncv1.RegisterConflictServiceServer(grpcServer, srv)
	syncv1.RegisterNodeSyncServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, cleanup, fmt.Errorf("sync listen: %w", err)
	}

	go func() { _ = grpcServer.Serve(lis) }()
	cleanups = append(cleanups, func() { grpcServer.Stop() })

	return &Env{Addr: lis.Addr().String()}, cleanup, nil
}
