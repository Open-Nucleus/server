package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"flag"
	"log/slog"
	"os"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
	"github.com/FibrinLab/open-nucleus/internal/handler"
	fhirhandler "github.com/FibrinLab/open-nucleus/internal/handler/fhir"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/router"
	"github.com/FibrinLab/open-nucleus/internal/server"
	"github.com/FibrinLab/open-nucleus/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Ed25519 keypair — in production, loaded from config/secrets.
	// For Phase 1, generate an ephemeral keypair for development.
	pubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		logger.Error("failed to generate Ed25519 key", "error", err)
		os.Exit(1)
	}

	// gRPC connection pool — non-blocking, services may not be up yet
	pool, err := grpcclient.NewPool(cfg.GRPC)
	if err != nil {
		logger.Warn("gRPC pool init had errors (services may not be running)", "error", err)
	}
	defer pool.Close()

	// Services
	authSvc := service.NewAuthService(pool)
	patientSvc := service.NewPatientService(pool)
	syncSvc := service.NewSyncService(pool)
	conflictSvc := service.NewConflictService(pool)
	sentinelSvc := service.NewSentinelService(pool)
	formularySvc := service.NewFormularyService(pool)
	anchorSvc := service.NewAnchorService(pool)
	supplySvc := service.NewSupplyService(pool)
	smartSvc := service.NewSmartService(pool)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	patientHandler := handler.NewPatientHandler(patientSvc)
	syncHandler := handler.NewSyncHandler(syncSvc)
	conflictHandler := handler.NewConflictHandler(conflictSvc)
	sentinelHandler := handler.NewSentinelHandler(sentinelSvc)
	formularyHandler := handler.NewFormularyHandler(formularySvc)
	anchorHandler := handler.NewAnchorHandler(anchorSvc)
	supplyHandler := handler.NewSupplyHandler(supplySvc)
	resourceHandler := handler.NewResourceHandler(patientSvc)
	fhirHandler := fhirhandler.NewFHIRHandler(patientSvc)
	smartHandler := handler.NewSmartHandler(smartSvc, cfg.Smart.BaseURL)

	// Schema validator
	sv := middleware.NewSchemaValidator()
	schemas := map[string]string{
		"patient":             "schemas/patient.json",
		"encounter":           "schemas/encounter.json",
		"observation":         "schemas/observation.json",
		"condition":           "schemas/condition.json",
		"medication_request":  "schemas/medication_request.json",
		"allergy_intolerance": "schemas/allergy_intolerance.json",
		"immunization":        "schemas/immunization.json",
		"procedure":           "schemas/procedure.json",
	}
	for pattern, path := range schemas {
		data, err := os.ReadFile(path)
		if err != nil {
			logger.Warn("failed to load JSON schema", "pattern", pattern, "path", path, "error", err)
			continue
		}
		if err := sv.RegisterSchema(pattern, string(data)); err != nil {
			logger.Warn("failed to compile JSON schema", "pattern", pattern, "error", err)
		}
	}

	// Middleware components
	jwtAuth := middleware.NewJWTAuth(pubKey, cfg.Auth.JWTIssuer)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)

	// Audit logger
	auditLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Router
	mux := router.New(router.Config{
		AuthHandler:      authHandler,
		PatientHandler:   patientHandler,
		ResourceHandler:  resourceHandler,
		SyncHandler:      syncHandler,
		ConflictHandler:  conflictHandler,
		SentinelHandler:  sentinelHandler,
		FormularyHandler: formularyHandler,
		AnchorHandler:    anchorHandler,
		SupplyHandler:    supplyHandler,
		FHIRHandler:      fhirHandler,
		SmartHandler:     smartHandler,
		SchemaValidator:  sv,
		JWTAuth:          jwtAuth,
		RateLimiter:      rateLimiter,
		CORSOrigins:      cfg.CORS.AllowedOrigins,
		AuditLogger:      auditLogger,
	})

	// Server
	srv := server.New(cfg, mux, logger)
	logger.Info("Open Nucleus API Gateway starting",
		"port", cfg.Server.Port,
		"version", "0.2.0-phase2",
	)

	if err := srv.Run(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
