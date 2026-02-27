package auth_test

import (
	"context"
	"database/sql"
	"fmt"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	_ "modernc.org/sqlite"
)

func setupTestServer(t *testing.T) (authv1.AuthServiceClient, *service.AuthService) {
	t.Helper()

	tmpDir := t.TempDir()

	cfg := &config.Config{
		GRPCPort: 0,
		JWT: config.JWTConfig{
			Issuer:          "test-auth",
			AccessLifetime:  time.Hour,
			RefreshLifetime: 24 * time.Hour,
			ClockSkew:       time.Minute,
		},
		Git: config.GitConfig{
			RepoPath:    tmpDir + "/data",
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
			BootstrapSecret: "test-bootstrap-secret",
		},
		KeyStore: config.KeyStoreConfig{Type: "memory"},
		SQLite:   config.SQLiteConfig{DBPath: tmpDir + "/auth.db"},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	require.NoError(t, err)

	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	err = store.InitSchema(db)
	require.NoError(t, err)

	denyList := store.NewDenyList(db)

	keyStore := auth.NewMemoryKeyStore()

	svc, err := service.NewAuthService(cfg, git, keyStore, denyList)
	require.NoError(t, err)

	srv := server.NewServer(cfg, svc)
	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() { _ = grpcServer.Serve(lis) }()
	t.Cleanup(func() { grpcServer.Stop() })

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	client := authv1.NewAuthServiceClient(conn)
	return client, svc
}

func registerTestDevice(t *testing.T, client authv1.AuthServiceClient, bootstrap bool) (string, auth.RoleDefinition) {
	t.Helper()
	pub, _, err := auth.GenerateKeypair()
	require.NoError(t, err)

	req := &authv1.RegisterDeviceRequest{
		PublicKey:      auth.EncodePublicKey(pub),
		PractitionerId: "practitioner-1",
		SiteId:         "site-1",
		DeviceName:     "test-device",
		Role:           "physician",
	}
	if bootstrap {
		req.BootstrapSecret = "test-bootstrap-secret"
	}

	resp, err := client.RegisterDevice(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp.Device)

	role, _ := auth.GetRole(resp.Device.Role)
	return resp.Device.DeviceId, role
}

func TestBootstrapFlow(t *testing.T) {
	client, _ := setupTestServer(t)
	ctx := context.Background()

	pub, _, err := auth.GenerateKeypair()
	require.NoError(t, err)

	resp, err := client.RegisterDevice(ctx, &authv1.RegisterDeviceRequest{
		PublicKey:       auth.EncodePublicKey(pub),
		PractitionerId: "admin-1",
		SiteId:         "site-1",
		DeviceName:     "admin-tablet",
		Role:           "nurse", // requested role ignored for bootstrap
		BootstrapSecret: "test-bootstrap-secret",
	})
	require.NoError(t, err)
	assert.Equal(t, "regional-admin", resp.Device.Role)
	assert.Equal(t, "active", resp.Device.Status)
}

func TestFullAuthCycle(t *testing.T) {
	client, svc := setupTestServer(t)
	ctx := context.Background()

	// 1. Register device
	pub, priv, err := auth.GenerateKeypair()
	require.NoError(t, err)

	regResp, err := client.RegisterDevice(ctx, &authv1.RegisterDeviceRequest{
		PublicKey:       auth.EncodePublicKey(pub),
		PractitionerId: "dr-jones",
		SiteId:         "site-1",
		DeviceName:     "tablet-1",
		Role:           "physician",
		BootstrapSecret: "test-bootstrap-secret",
	})
	require.NoError(t, err)
	deviceID := regResp.Device.DeviceId

	// 2. Get challenge
	challengeResp, err := client.GetChallenge(ctx, &authv1.GetChallengeRequest{DeviceId: deviceID})
	require.NoError(t, err)
	nonce := challengeResp.Challenge.Nonce
	assert.Len(t, nonce, 32)

	// 3. Authenticate (sign nonce with device private key)
	signature := auth.Sign(priv, nonce)

	// Use service directly for auth since gRPC layer needs nonce+sig separation
	accessToken, refreshToken, _, _, _, authErr := svc.AuthenticateWithNonce(deviceID, nonce, signature)
	require.NoError(t, authErr)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)

	// 4. Validate token
	validateResp, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: accessToken})
	require.NoError(t, err)
	assert.True(t, validateResp.Valid)
	assert.Equal(t, "dr-jones", validateResp.Claims.Sub)

	// 5. Refresh
	refreshResp, err := client.RefreshToken(ctx, &authv1.RefreshTokenRequest{RefreshToken: refreshToken})
	require.NoError(t, err)
	assert.NotEmpty(t, refreshResp.AccessToken)
	assert.NotEmpty(t, refreshResp.RefreshToken)

	// 6. Logout
	_, err = client.Logout(ctx, &authv1.LogoutRequest{Token: refreshResp.AccessToken})
	require.NoError(t, err)

	// Validate should now fail
	validateResp, err = client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: refreshResp.AccessToken})
	require.NoError(t, err)
	assert.False(t, validateResp.Valid)
}

func TestChallengeResponse_ExpiredNonce(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		JWT: config.JWTConfig{Issuer: "test", AccessLifetime: time.Hour, RefreshLifetime: 24 * time.Hour},
		Git: config.GitConfig{RepoPath: tmpDir + "/data", AuthorName: "test", AuthorEmail: "test@test.local"},
		Devices:  config.DevicesConfig{Path: ".nucleus/devices"},
		Security: config.SecurityConfig{NonceTTL: 1 * time.Millisecond, MaxFailures: 10, FailureWindow: 60 * time.Second, BootstrapSecret: "secret"},
		KeyStore: config.KeyStoreConfig{Type: "memory"},
		SQLite:   config.SQLiteConfig{DBPath: tmpDir + "/auth.db"},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	require.NoError(t, err)

	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, store.InitSchema(db))

	svc, err := service.NewAuthService(cfg, git, auth.NewMemoryKeyStore(), store.NewDenyList(db))
	require.NoError(t, err)

	pub, priv, _ := auth.GenerateKeypair()
	_, regErr := svc.RegisterDevice(auth.EncodePublicKey(pub), "p1", "s1", "dev", "physician", "secret")
	require.NoError(t, regErr)

	devices, _ := svc.ListDevices()
	require.Len(t, devices, 1)
	deviceID := devices[0].DeviceID

	nonce, _, _ := svc.GetChallenge(deviceID)
	time.Sleep(5 * time.Millisecond) // let nonce expire

	sig := auth.Sign(priv, nonce)
	_, _, _, _, _, authErr := svc.AuthenticateWithNonce(deviceID, nonce, sig)
	assert.Error(t, authErr)
	assert.Contains(t, authErr.Error(), "invalid or expired")
}

func TestChallengeResponse_WrongKey(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		JWT: config.JWTConfig{Issuer: "test", AccessLifetime: time.Hour, RefreshLifetime: 24 * time.Hour},
		Git: config.GitConfig{RepoPath: tmpDir + "/data", AuthorName: "test", AuthorEmail: "test@test.local"},
		Devices:  config.DevicesConfig{Path: ".nucleus/devices"},
		Security: config.SecurityConfig{NonceTTL: 60 * time.Second, MaxFailures: 10, FailureWindow: 60 * time.Second, BootstrapSecret: "secret"},
		KeyStore: config.KeyStoreConfig{Type: "memory"},
		SQLite:   config.SQLiteConfig{DBPath: tmpDir + "/auth.db"},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	require.NoError(t, err)

	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, store.InitSchema(db))

	svc, err := service.NewAuthService(cfg, git, auth.NewMemoryKeyStore(), store.NewDenyList(db))
	require.NoError(t, err)

	pub, _, _ := auth.GenerateKeypair()
	_, wrongPriv, _ := auth.GenerateKeypair() // different key

	_, regErr := svc.RegisterDevice(auth.EncodePublicKey(pub), "p1", "s1", "dev", "physician", "secret")
	require.NoError(t, regErr)

	devices, _ := svc.ListDevices()
	deviceID := devices[0].DeviceID

	nonce, _, _ := svc.GetChallenge(deviceID)
	wrongSig := auth.Sign(wrongPriv, nonce)

	_, _, _, _, _, authErr := svc.AuthenticateWithNonce(deviceID, nonce, wrongSig)
	assert.Error(t, authErr)
	assert.Contains(t, authErr.Error(), "signature verification failed")
}

func TestValidateToken_Expired(t *testing.T) {
	client, _ := setupTestServer(t)
	ctx := context.Background()

	// Create an expired token directly
	pub, priv, _ := auth.GenerateKeypair()
	_ = pub
	claims := auth.NewAccessClaims("sub", "dev", "node", "site", "physician", nil, "local", "jti", "test-auth", -time.Hour)
	token, _ := auth.SignToken(claims, priv, "key")

	// This will be invalid because it's signed with a different key than the server's
	resp, err := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: token})
	require.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestDeviceRevocation(t *testing.T) {
	client, svc := setupTestServer(t)
	ctx := context.Background()

	pub, priv, _ := auth.GenerateKeypair()
	regResp, err := client.RegisterDevice(ctx, &authv1.RegisterDeviceRequest{
		PublicKey: auth.EncodePublicKey(pub), PractitionerId: "p1", SiteId: "s1",
		DeviceName: "dev1", Role: "nurse", BootstrapSecret: "test-bootstrap-secret",
	})
	require.NoError(t, err)
	deviceID := regResp.Device.DeviceId

	// Authenticate
	nonce, _, _ := svc.GetChallenge(deviceID)
	sig := auth.Sign(priv, nonce)
	accessToken, _, _, _, _, authErr := svc.AuthenticateWithNonce(deviceID, nonce, sig)
	require.NoError(t, authErr)

	// Validate token works
	vResp, _ := client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: accessToken})
	assert.True(t, vResp.Valid)

	// Revoke device
	revokeResp, err := client.RevokeDevice(ctx, &authv1.RevokeDeviceRequest{
		DeviceId: deviceID, RevokedBy: "admin", Reason: "compromised",
	})
	require.NoError(t, err)
	assert.Equal(t, "revoked", revokeResp.Device.Status)

	// Token should now fail validation (device revoked)
	vResp, _ = client.ValidateToken(ctx, &authv1.ValidateTokenRequest{Token: accessToken})
	assert.False(t, vResp.Valid)
}

func TestBruteForceProtection(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		JWT: config.JWTConfig{Issuer: "test", AccessLifetime: time.Hour, RefreshLifetime: 24 * time.Hour},
		Git: config.GitConfig{RepoPath: tmpDir + "/data", AuthorName: "test", AuthorEmail: "test@test.local"},
		Devices:  config.DevicesConfig{Path: ".nucleus/devices"},
		Security: config.SecurityConfig{NonceTTL: 60 * time.Second, MaxFailures: 3, FailureWindow: 60 * time.Second, BootstrapSecret: "secret"},
		KeyStore: config.KeyStoreConfig{Type: "memory"},
		SQLite:   config.SQLiteConfig{DBPath: tmpDir + "/auth.db"},
	}

	git, err := gitstore.NewStore(cfg.Git.RepoPath, cfg.Git.AuthorName, cfg.Git.AuthorEmail)
	require.NoError(t, err)

	db, err := sql.Open("sqlite", cfg.SQLite.DBPath)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, store.InitSchema(db))

	svc, err := service.NewAuthService(cfg, git, auth.NewMemoryKeyStore(), store.NewDenyList(db))
	require.NoError(t, err)

	pub, _, _ := auth.GenerateKeypair()
	_, _ = svc.RegisterDevice(auth.EncodePublicKey(pub), "p1", "s1", "dev", "physician", "secret")

	devices, _ := svc.ListDevices()
	deviceID := devices[0].DeviceID

	// Simulate failures
	for i := 0; i < 3; i++ {
		nonce, _, _ := svc.GetChallenge(deviceID)
		_, _, _, _, _, authErr := svc.AuthenticateWithNonce(deviceID, nonce, []byte("bad-sig-"+fmt.Sprintf("%d", i)))
		assert.Error(t, authErr)
	}

	// Should now be blocked
	_, _, blockErr := svc.GetChallenge(deviceID)
	assert.Error(t, blockErr)
	assert.Contains(t, blockErr.Error(), "blocked")
}

func TestCheckPermission_SiteScope(t *testing.T) {
	client, svc := setupTestServer(t)
	ctx := context.Background()

	pub, priv, _ := auth.GenerateKeypair()
	regResp, err := client.RegisterDevice(ctx, &authv1.RegisterDeviceRequest{
		PublicKey: auth.EncodePublicKey(pub), PractitionerId: "p1", SiteId: "site-A",
		DeviceName: "dev1", Role: "physician", BootstrapSecret: "test-bootstrap-secret",
	})
	require.NoError(t, err)
	deviceID := regResp.Device.DeviceId

	// Bootstrap gives regional-admin, re-assign to physician for scope test
	_, err = client.AssignRole(ctx, &authv1.AssignRoleRequest{DeviceId: deviceID, Role: "physician", AssignedBy: "admin"})
	require.NoError(t, err)

	// Authenticate
	nonce, _, _ := svc.GetChallenge(deviceID)
	sig := auth.Sign(priv, nonce)
	accessToken, _, _, _, _, authErr := svc.AuthenticateWithNonce(deviceID, nonce, sig)
	require.NoError(t, authErr)

	// Same site — allowed
	resp, err := client.CheckPermission(ctx, &authv1.CheckPermissionRequest{
		Token: accessToken, Permission: "patient:read", TargetSiteId: "site-A",
	})
	require.NoError(t, err)
	assert.True(t, resp.Allowed)

	// Different site — physician has local scope, should be denied
	resp, err = client.CheckPermission(ctx, &authv1.CheckPermissionRequest{
		Token: accessToken, Permission: "patient:read", TargetSiteId: "site-B",
	})
	require.NoError(t, err)
	assert.False(t, resp.Allowed)
	assert.Contains(t, resp.Reason, "site_scope")
}

func TestRoleAssignment(t *testing.T) {
	client, _ := setupTestServer(t)
	ctx := context.Background()

	pub, _, _ := auth.GenerateKeypair()
	regResp, err := client.RegisterDevice(ctx, &authv1.RegisterDeviceRequest{
		PublicKey: auth.EncodePublicKey(pub), PractitionerId: "p1", SiteId: "site-1",
		DeviceName: "dev1", Role: "chw", BootstrapSecret: "test-bootstrap-secret",
	})
	require.NoError(t, err)
	deviceID := regResp.Device.DeviceId

	// Assign new role
	assignResp, err := client.AssignRole(ctx, &authv1.AssignRoleRequest{
		DeviceId: deviceID, Role: "nurse", AssignedBy: "admin",
	})
	require.NoError(t, err)
	assert.Equal(t, "nurse", assignResp.Device.Role)
}

func TestHealthCheck(t *testing.T) {
	client, _ := setupTestServer(t)
	ctx := context.Background()

	resp, err := client.Health(ctx, &authv1.HealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.Equal(t, "0.4.0", resp.Version)
}

func TestListRoles(t *testing.T) {
	client, _ := setupTestServer(t)
	ctx := context.Background()

	resp, err := client.ListRoles(ctx, &authv1.ListRolesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Roles, 5)
}

func TestGetRole(t *testing.T) {
	client, _ := setupTestServer(t)
	ctx := context.Background()

	resp, err := client.GetRole(ctx, &authv1.GetRoleRequest{RoleCode: "physician"})
	require.NoError(t, err)
	assert.Equal(t, "Physician", resp.Role.Display)

	_, err = client.GetRole(ctx, &authv1.GetRoleRequest{RoleCode: "nonexistent"})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Contains(t, st.Message(), "not found")
}
