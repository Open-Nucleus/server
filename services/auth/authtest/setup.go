// Package authtest exports helpers for spinning up an in-process Auth Service
// for integration/E2E tests. This package is allowed to import the internal
// packages because it lives under services/auth/.
package authtest

import (
	"crypto/ed25519"
	"database/sql"
	"net"
	"testing"
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

// Env holds the running Auth Service test environment.
type Env struct {
	Addr      string
	PublicKey ed25519.PublicKey
	Client    authv1.AuthServiceClient
	svc       *service.AuthService
}

// Start boots an in-process Auth Service on a dynamic port.
func Start(t *testing.T, tmpDir, bootstrapSecret string) *Env {
	t.Helper()

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
			AuthorName:  "test",
			AuthorEmail: "test@test.local",
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
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := store.InitSchema(db); err != nil {
		t.Fatal(err)
	}

	denyList := store.NewDenyList(db)
	keyStore := auth.NewMemoryKeyStore()

	svc, err := service.NewAuthService(cfg, git, keyStore, denyList)
	if err != nil {
		t.Fatal(err)
	}

	srv := server.NewServer(cfg, svc)
	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })

	return &Env{
		Addr:      lis.Addr().String(),
		PublicKey: svc.NodePublicKey(),
		svc:       svc,
	}
}

// GetChallenge generates a challenge nonce for the given device.
func (e *Env) GetChallenge(deviceID string) ([]byte, time.Time, error) {
	return e.svc.GetChallenge(deviceID)
}

// AuthenticateWithNonce completes challenge-response auth.
func (e *Env) AuthenticateWithNonce(deviceID string, nonce, sig []byte) (accessToken, refreshToken string, err error) {
	access, refresh, _, _, _, authErr := e.svc.AuthenticateWithNonce(deviceID, nonce, sig)
	return access, refresh, authErr
}
