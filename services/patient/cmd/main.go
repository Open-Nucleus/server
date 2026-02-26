package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	patientv1 "github.com/FibrinLab/open-nucleus/gen/proto/patient/v1"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/config"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/pipeline"
	"github.com/FibrinLab/open-nucleus/services/patient/internal/server"
	"google.golang.org/grpc"
)

func main() {
	// Load config
	cfgPath := os.Getenv("PATIENT_CONFIG")
	if cfgPath == "" {
		cfgPath = "services/patient/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("Starting Patient Service on port %d", cfg.GRPCPort)
	log.Printf("Git repo: %s", cfg.Git.RepoPath)
	log.Printf("SQLite DB: %s", cfg.SQLite.DBPath)

	// Open/init Git repo
	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	if err != nil {
		log.Fatalf("init git store: %v", err)
	}

	// Open SQLite database + init schema
	idx, err := sqliteindex.NewIndex(cfg.SQLite.DBPath)
	if err != nil {
		log.Fatalf("init sqlite index: %v", err)
	}
	defer idx.Close()

	// Create write pipeline
	pw := pipeline.NewWriter(git, idx, cfg.WriteLock.Timeout)

	// Health check on startup
	if cfg.Index.HealthCheckOnStartup {
		indexHead, _ := idx.GetMeta("git_head")
		gitHead, _ := git.Head()
		if indexHead != gitHead && gitHead != "" {
			log.Printf("Index drift detected (index: %s, git: %s)", indexHead, gitHead)
			if cfg.Index.AutoRebuildOnDrift {
				log.Println("Auto-rebuilding index...")
				count, head, err := pw.RebuildIndex()
				if err != nil {
					log.Fatalf("index rebuild failed: %v", err)
				}
				log.Printf("Index rebuilt: %d resources, HEAD=%s", count, head)
			}
		} else {
			log.Println("Index health check: OK")
		}
	}

	// Create gRPC server
	srv := server.NewServer(cfg, pw, idx, git)

	grpcServer := grpc.NewServer()
	patientv1.RegisterPatientServiceServer(grpcServer, srv)

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

	log.Printf("Patient Service listening on :%d", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
