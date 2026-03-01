package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	formularyv1 "github.com/FibrinLab/open-nucleus/gen/proto/formulary/v1"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/config"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/dosing"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/server"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/service"
	"github.com/FibrinLab/open-nucleus/services/formulary/internal/store"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

func main() {
	cfgPath := os.Getenv("FORMULARY_CONFIG")
	if cfgPath == "" {
		cfgPath = "services/formulary/config.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("Starting Formulary Service on port %d", cfg.GRPCPort)

	// Open SQLite for stock levels
	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}

	// Load drug database
	drugDB := store.NewDrugDB()
	interactions := store.NewInteractionIndex()

	// TODO: load from configured paths or embedded data
	// For now, the service starts empty and relies on the test helpers
	// to load seed data

	stockStore := store.NewStockStore(db)
	dosingEngine := dosing.NewStubEngine()

	svc := service.New(drugDB, interactions, stockStore, dosingEngine)
	srv := server.NewServer(svc)
	grpcServer := grpc.NewServer()
	formularyv1.RegisterFormularyServiceServer(grpcServer, srv)

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

	log.Printf("Formulary Service listening on :%d", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
