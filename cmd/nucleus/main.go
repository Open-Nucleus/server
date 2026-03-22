package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/config"
	"github.com/FibrinLab/open-nucleus/internal/handler"
	fhirhandler "github.com/FibrinLab/open-nucleus/internal/handler/fhir"
	"github.com/FibrinLab/open-nucleus/internal/middleware"
	"github.com/FibrinLab/open-nucleus/internal/router"
	"github.com/FibrinLab/open-nucleus/internal/server"
	"github.com/FibrinLab/open-nucleus/internal/service/local"
	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/consent"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
	nucleustls "github.com/FibrinLab/open-nucleus/pkg/tls"
	"github.com/FibrinLab/open-nucleus/services/anchor/anchorservice"
	"github.com/FibrinLab/open-nucleus/services/auth/authservice"
	"github.com/FibrinLab/open-nucleus/services/formulary/formularyservice"
	"github.com/FibrinLab/open-nucleus/services/patient/pipeline"
	"github.com/FibrinLab/open-nucleus/services/sync/syncservice"

	_ "modernc.org/sqlite"
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

	// --- Data layer: shared Git + SQLite ---

	git, err := gitstore.NewStore(cfg.Data.RepoPath, cfg.Data.AuthorName, cfg.Data.AuthorEmail)
	if err != nil {
		logger.Error("failed to init git store", "error", err)
		os.Exit(1)
	}

	db, err := openDB(cfg.Data.DBPath)
	if err != nil {
		logger.Error("failed to open sqlite", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := sqliteindex.InitUnifiedSchema(db); err != nil {
		logger.Error("failed to init unified schema", "error", err)
		os.Exit(1)
	}

	idx := sqliteindex.NewIndexFromDB(db)

	// --- Patient service ---

	pw := pipeline.NewWriter(git, idx, 10*time.Second)
	patientSvc := local.NewPatientService(pw, idx, git)

	// --- Auth service ---

	ks := auth.NewMemoryKeyStore()
	denyList := authservice.NewDenyList(db)
	clientStore := authservice.NewClientStore(git, db)

	authCfg := &authservice.Config{
		JWT: authservice.JWTConfig{
			Issuer:          cfg.Auth.JWTIssuer,
			AccessLifetime:  cfg.Auth.TokenLifetime,
			RefreshLifetime: cfg.Auth.RefreshWindow,
			ClockSkew:       2 * time.Hour,
		},
		Git: authservice.GitConfig{
			RepoPath:    cfg.Data.RepoPath,
			AuthorName:  cfg.Data.AuthorName,
			AuthorEmail: cfg.Data.AuthorEmail,
		},
		Devices: authservice.DevicesConfig{
			Path: ".nucleus/devices",
		},
		Security: authservice.SecurityConfig{
			NonceTTL:        60 * time.Second,
			MaxFailures:     10,
			FailureWindow:   60 * time.Second,
			BootstrapSecret: os.Getenv("NUCLEUS_BOOTSTRAP_SECRET"),
		},
		KeyStore: authservice.KeyStoreConfig{
			Type: "memory",
		},
		SQLite: authservice.SQLiteConfig{
			DBPath: cfg.Data.DBPath,
		},
	}

	authImpl, err := authservice.NewAuthService(authCfg, git, ks, denyList)
	if err != nil {
		logger.Error("failed to init auth service", "error", err)
		os.Exit(1)
	}

	authSvc := local.NewAuthService(authImpl)

	smartImpl := authservice.NewSmartService(authImpl, clientStore)
	smartSvc := local.NewSmartService(smartImpl)

	// --- Sync service ---

	mergeDriver := merge.NewDriver(nil)
	eventBus := syncservice.NewEventBus(100)
	conflictStore := syncservice.NewConflictStore(db)
	historyStore := syncservice.NewHistoryStore(db, 10000)
	peerStore := syncservice.NewPeerStore(db)

	syncCfg := &syncservice.Config{
		Git: syncservice.GitConfig{
			RepoPath:    cfg.Data.RepoPath,
			AuthorName:  cfg.Data.AuthorName,
			AuthorEmail: cfg.Data.AuthorEmail,
		},
	}
	// Workaround: type alias doesn't carry the unexported GitConfig, so
	// we use the re-exported config type directly.
	_ = syncCfg

	syncEngine := syncservice.NewSyncEngine(
		syncCfg, git, conflictStore, historyStore, peerStore,
		mergeDriver, eventBus, "node-local", "site-local",
	)

	syncSvc := local.NewSyncService(syncEngine, historyStore, peerStore)
	conflictSvc := local.NewConflictService(conflictStore, eventBus)

	// --- Formulary service ---

	drugDB := formularyservice.NewDrugDB()
	interactions := formularyservice.NewInteractionIndex()
	stockStore := formularyservice.NewStockStore(db)
	dosingEngine := formularyservice.DosingEngine(formularyservice.NewPharmDosingEngine())

	formularyImpl := formularyservice.New(drugDB, interactions, stockStore, dosingEngine)
	formularySvc := local.NewFormularyService(formularyImpl)

	// --- Anchor service ---

	anchorQueue := anchorservice.NewAnchorQueue(db)
	anchorStore := anchorservice.NewAnchorStore(git)
	credStore := anchorservice.NewCredentialStore(git)
	didStore := anchorservice.NewDIDStore(git)

	identityEngine := openanchor.NewLocalIdentityEngine()

	// Use the auth service's node key for anchoring
	nodePrivKey := authImpl.NodePrivateKey()

	// Select anchor backend based on config.
	var anchorBackend openanchor.AnchorEngine
	if cfg.Anchor.Backend == "hedera" {
		operatorKey := cfg.Anchor.OperatorKey
		if operatorKey == "" {
			operatorKey = os.Getenv("NUCLEUS_HEDERA_KEY")
		}
		hb, err := openanchor.NewHederaBackendFromConfig(
			cfg.Anchor.Network,
			cfg.Anchor.OperatorID,
			operatorKey,
			cfg.Anchor.TopicID,
			cfg.Anchor.DIDTopicID,
			cfg.Anchor.MirrorURL,
			nodePrivKey,
		)
		if err != nil {
			logger.Error("failed to init Hedera anchor backend", "error", err)
			os.Exit(1)
		}
		anchorBackend = hb
		logger.Info("anchor backend: hedera", "network", cfg.Anchor.Network, "topic", cfg.Anchor.TopicID)
	} else if cfg.Anchor.Backend == "iota" {
		ib, err := openanchor.NewIotaBackendFromConfig(
			cfg.Anchor.Network,
			cfg.Anchor.RPCURL,
			cfg.Anchor.AnchorPackageID,
			cfg.Anchor.IdentityPackageID,
			nodePrivKey,
		)
		if err != nil {
			logger.Error("failed to init IOTA anchor backend", "error", err)
			os.Exit(1)
		}
		anchorBackend = ib
		logger.Info("anchor backend: iota", "network", cfg.Anchor.Network, "rpc", cfg.Anchor.RPCURL)
	} else {
		anchorBackend = openanchor.NewStubBackend()
		logger.Info("anchor backend: stub (no blockchain)")
	}

	anchorImpl := anchorservice.New(
		git, anchorBackend, identityEngine,
		anchorQueue, anchorStore, credStore, didStore,
		nodePrivKey,
	)
	anchorSvc := local.NewAnchorService(anchorImpl)

	// --- Consent service ---
	consentMgr := consent.NewManager(idx, git, logger)
	consentSvc := local.NewLocalConsentService(consentMgr, anchorImpl.NodeDIDString(), nodePrivKey)
	consentHandler := handler.NewConsentHandler(consentSvc)
	consentMiddleware := middleware.ConsentCheck(consentMgr, logger)

	// --- Sentinel + Supply remain gRPC stubs (Python process) ---
	// For now, use nil-safe stubs. When Sentinel is running, connect via gRPC.
	sentinelSvc := local.NewStubSentinelService()
	supplySvc := local.NewStubSupplyService()

	// --- Handlers (identical to gateway) ---

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

	// Middleware components — use auth service's public key for JWT verification
	pubKey := authImpl.NodePublicKey()
	jwtAuth := middleware.NewJWTAuth(pubKey, cfg.Auth.JWTIssuer)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit)

	// Audit logger
	auditLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Router (identical to gateway)
	mux := router.New(router.Config{
		AuthHandler:       authHandler,
		PatientHandler:    patientHandler,
		ResourceHandler:   resourceHandler,
		SyncHandler:       syncHandler,
		ConflictHandler:   conflictHandler,
		SentinelHandler:   sentinelHandler,
		FormularyHandler:  formularyHandler,
		AnchorHandler:     anchorHandler,
		SupplyHandler:     supplyHandler,
		ConsentHandler:    consentHandler,
		FHIRHandler:       fhirHandler,
		SmartHandler:      smartHandler,
		SchemaValidator:   sv,
		JWTAuth:           jwtAuth,
		RateLimiter:       rateLimiter,
		ConsentMiddleware: consentMiddleware,
		CORSOrigins:       cfg.CORS.AllowedOrigins,
		AuditLogger:       auditLogger,
	})

	// TLS
	tlsCfg, err := nucleustls.LoadOrGenerate(nucleustls.Config{
		Mode:     cfg.TLS.Mode,
		CertFile: cfg.TLS.CertFile,
		KeyFile:  cfg.TLS.KeyFile,
		CertDir:  cfg.TLS.CertDir,
	})
	if err != nil {
		logger.Error("failed to configure TLS", "error", err)
		os.Exit(1)
	}

	// Server
	srv := server.New(cfg, mux, logger).WithTLS(tlsCfg)
	logger.Info("Open Nucleus starting",
		"port", cfg.Server.Port,
		"tls", cfg.TLS.Mode,
		"version", "1.0.0-monolith",
	)

	if err := srv.Run(); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func openDB(path string) (*sql.DB, error) {
	dsn := path + "?_journal_mode=WAL&_busy_timeout=5000&_cache_size=-20000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes
	return db, nil
}
