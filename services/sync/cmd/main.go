package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"

	syncv1 "github.com/FibrinLab/open-nucleus/gen/proto/sync/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/config"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/server"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/service"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/store"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport/localnet"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

func main() {
	cfgPath := os.Getenv("SYNC_CONFIG")
	if cfgPath == "" {
		cfgPath = "services/sync/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("Starting Sync Service on port %d", cfg.GRPCPort)

	// Generate or load node ID
	nodeID := cfg.Node.ID
	if nodeID == "" {
		nodeID = uuid.New().String()
	}
	siteID := "default-site"

	// Open Git repo (shared with Patient and Auth services)
	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	if err != nil {
		log.Fatalf("init git store: %v", err)
	}

	// Open SQLite for sync state
	db, err := sql.Open("sqlite", cfg.History.DBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Create stores
	conflictStore := store.NewConflictStore(db)
	historyStore := store.NewHistoryStore(db, cfg.History.MaxEntries)
	peerStore := store.NewPeerStore(db)

	// Create merge driver (no formulary checker in this phase)
	mergeDriver := merge.NewDriver(nil)

	// Create event bus
	eventBus := service.NewEventBus(cfg.Events.BufferSize)

	// Create sync engine
	engine := service.NewSyncEngine(cfg, git, conflictStore, historyStore, peerStore, mergeDriver, eventBus, nodeID, siteID)

	// Register transports
	if cfg.Transports.LocalNetwork.Enabled {
		ln := localnet.New(nodeID, siteID, cfg.Transports.LocalNetwork.MDNSService, cfg.Transports.LocalNetwork.Port)
		engine.RegisterTransport(ln)
	}
	if cfg.Transports.WiFiDirect.Enabled {
		engine.RegisterTransport(&transport.StubAdapter{TransportName: "wifi_direct"})
	}
	if cfg.Transports.Bluetooth.Enabled {
		engine.RegisterTransport(&transport.StubAdapter{TransportName: "bluetooth"})
	}
	if cfg.Transports.USB.Enabled {
		engine.RegisterTransport(&transport.StubAdapter{TransportName: "usb"})
	}

	// Create gRPC server
	srv := server.NewServer(cfg, engine, conflictStore, historyStore, peerStore, eventBus)
	grpcServer := grpc.NewServer()
	syncv1.RegisterSyncServiceServer(grpcServer, srv)
	syncv1.RegisterConflictServiceServer(grpcServer, srv)
	syncv1.RegisterNodeSyncServiceServer(grpcServer, srv)

	// Listen
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Sync Service listening on :%d", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
