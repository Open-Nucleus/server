package authtest

import (
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"net"
	"time"

	authv1 "github.com/FibrinLab/open-nucleus/gen/proto/auth/v1"
	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/config"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/server"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/service"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/store"
	"google.golang.org/grpc"

	_ "modernc.org/sqlite"
)

// StandaloneEnv is like Env but includes the AuthService for direct method calls.
type StandaloneEnv struct {
	Addr      string
	PublicKey ed25519.PublicKey
	Client    authv1.AuthServiceClient
	Svc       *service.AuthService
}

// GetChallenge generates a challenge nonce for the given device.
func (e *StandaloneEnv) GetChallenge(deviceID string) ([]byte, time.Time, error) {
	return e.Svc.GetChallenge(deviceID)
}

// AuthenticateWithNonce completes challenge-response auth.
func (e *StandaloneEnv) AuthenticateWithNonce(deviceID string, nonce, sig []byte) (accessToken, refreshToken string, err error) {
	access, refresh, _, _, _, authErr := e.Svc.AuthenticateWithNonce(deviceID, nonce, sig)
	return access, refresh, authErr
}

// StartStandalone boots an in-process Auth Service without requiring *testing.T.
// Returns the environment and a cleanup function.
func StartStandalone(tmpDir, bootstrapSecret string) (*StandaloneEnv, func(), error) {
	var cleanups []func()
	cleanup := func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}

	cfg := &config.Config{
		GRPCPort: 0,
		JWT: config.JWTConfig{
			Issuer:          "open-nucleus-auth",
			AccessLifetime:  time.Hour,
			RefreshLifetime: 24 * time.Hour,
			ClockSkew:       time.Minute,
		},
		Git: config.GitConfig{
			RepoPath:    tmpDir + "/auth-data",
			AuthorName:  "smoke",
			AuthorEmail: "smoke@test.local",
		},
		Devices: config.DevicesConfig{
			Path: ".nucleus/devices",
		},
		Security: config.SecurityConfig{
			NonceTTL:        60 * time.Second,
			MaxFailures:     10,
			FailureWindow:   60 * time.Second,
			BootstrapSecret: bootstrapSecret,
		},
		KeyStore: config.KeyStoreConfig{Type: "memory"},
		SQLite:   config.SQLiteConfig{DBPath: tmpDir + "/auth.db"},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	if err != nil {
		return nil, cleanup, fmt.Errorf("auth gitstore: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	if err != nil {
		return nil, cleanup, fmt.Errorf("auth sqlite: %w", err)
	}
	cleanups = append(cleanups, func() { db.Close() })

	if err := store.InitSchema(db); err != nil {
		return nil, cleanup, fmt.Errorf("auth schema: %w", err)
	}

	denyList := store.NewDenyList(db)
	keyStore := auth.NewMemoryKeyStore()

	svc, err := service.NewAuthService(cfg, git, keyStore, denyList)
	if err != nil {
		return nil, cleanup, fmt.Errorf("auth service: %w", err)
	}

	srv := server.NewServer(cfg, svc)
	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, cleanup, fmt.Errorf("auth listen: %w", err)
	}

	go func() { _ = grpcServer.Serve(lis) }()
	cleanups = append(cleanups, func() { grpcServer.Stop() })

	return &StandaloneEnv{
		Addr:      lis.Addr().String(),
		PublicKey: svc.NodePublicKey(),
		Svc:       svc,
	}, cleanup, nil
}
