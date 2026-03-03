package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	smartv1 "github.com/FibrinLab/open-nucleus/gen/proto/smart/v1"
	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/config"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/server"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/service"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/store"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

func main() {
	cfgPath := os.Getenv("AUTH_CONFIG")
	if cfgPath == "" {
		cfgPath = "services/auth/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Override bootstrap secret from env
	if secret := os.Getenv("NUCLEUS_BOOTSTRAP_SECRET"); secret != "" {
		cfg.Security.BootstrapSecret = secret
	}

	log.Printf("Starting Auth Service on port %d", cfg.GRPCPort)

	// Open Git repo (shared with Patient Service)
	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	if err != nil {
		log.Fatalf("init git store: %v", err)
	}

	// Open SQLite for deny list
	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Create deny list
	denyList := store.NewDenyList(db)
	if err := denyList.LoadFromDB(); err != nil {
		log.Fatalf("load deny list: %v", err)
	}

	// Create key store
	var keyStore auth.KeyStore
	switch cfg.KeyStore.Type {
	case "file":
		ks, err := auth.NewFileKeyStore(cfg.Node.KeyPath)
		if err != nil {
			log.Fatalf("init file keystore: %v", err)
		}
		keyStore = ks
	default:
		keyStore = auth.NewMemoryKeyStore()
	}

	// Create auth service
	svc, err := service.NewAuthService(cfg, git, keyStore, denyList)
	if err != nil {
		log.Fatalf("init auth service: %v", err)
	}

	// Init SMART client store
	if err := store.InitClientSchema(db); err != nil {
		log.Fatalf("init smart client schema: %v", err)
	}
	clientStore := store.NewClientStore(git, db)
	smartSvc := service.NewSmartService(svc, clientStore)

	// Create gRPC server
	srv := server.NewServer(cfg, svc)
	smartSrv := server.NewSmartServer(smartSvc)
	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, srv)
	smartv1.RegisterSmartServiceServer(grpcServer, smartSrv)

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

	log.Printf("Auth Service listening on :%d", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
