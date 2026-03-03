package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/config"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/server"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/service"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/store"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

func main() {
	cfgPath := os.Getenv("ANCHOR_CONFIG")
	if cfgPath == "" {
		cfgPath = "services/anchor/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("Starting Anchor Service on port %d", cfg.GRPCPort)

	// Open SQLite for queue.
	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Open Git store.
	gs, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	if err != nil {
		log.Fatalf("open gitstore: %v", err)
	}

	// Generate Ed25519 keypair for node identity.
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("generate keypair: %v", err)
	}

	// Create stores.
	queue := store.NewAnchorQueue(db)
	anchorStore := store.NewAnchorStore(gs)
	credStore := store.NewCredentialStore(gs)
	didStore := store.NewDIDStore(gs)

	// Create engines.
	anchorEngine := openanchor.NewStubBackend()
	identityEngine := openanchor.NewLocalIdentityEngine()

	svc := service.New(gs, anchorEngine, identityEngine, queue, anchorStore, credStore, didStore, priv)
	if err := svc.Bootstrap(); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	srv := server.NewServer(svc)
	grpcServer := grpc.NewServer()
	anchorv1.RegisterAnchorServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Anchor Service listening on :%d", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
