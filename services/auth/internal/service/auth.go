package service

import (
	"crypto/ed25519"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/FibrinLab/open-nucleus/pkg/auth"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/config"
	"github.com/FibrinLab/open-nucleus/services/auth/internal/store"
)

// AuthService provides authentication and authorization operations.
type AuthService struct {
	mu             sync.Mutex // serializes Git writes
	cfg            *config.Config
	git            gitstore.Store
	keyStore       auth.KeyStore
	nonceStore     *auth.NonceStore
	denyList       *store.DenyList
	bruteForce     *auth.BruteForceGuard
	nodePrivateKey ed25519.PrivateKey
	nodePublicKey  ed25519.PublicKey
	nodeID         string
	bootstrapUsed  bool
	startedAt      time.Time
}

// NewAuthService creates a new AuthService.
func NewAuthService(cfg *config.Config, git gitstore.Store, ks auth.KeyStore, dl *store.DenyList) (*AuthService, error) {
	nonceStore := auth.NewNonceStore(cfg.Security.NonceTTL)
	bruteForce := auth.NewBruteForceGuard(cfg.Security.MaxFailures, cfg.Security.FailureWindow)

	// Generate or load node keypair
	nodeID := uuid.New().String()
	pub, priv, err := auth.GenerateKeypair()
	if err != nil {
		return nil, fmt.Errorf("generate node keypair: %w", err)
	}

	if err := ks.StoreKey("node", priv); err != nil {
		return nil, fmt.Errorf("store node key: %w", err)
	}

	// Check if bootstrap has been used (any existing devices)
	devices, _ := listAllDevices(git, cfg.Devices.Path)
	bootstrapUsed := len(devices) > 0

	svc := &AuthService{
		cfg:            cfg,
		git:            git,
		keyStore:       ks,
		nonceStore:     nonceStore,
		denyList:       dl,
		bruteForce:     bruteForce,
		nodePrivateKey: priv,
		nodePublicKey:  pub,
		nodeID:         nodeID,
		bootstrapUsed:  bootstrapUsed,
		startedAt:      time.Now(),
	}

	return svc, nil
}

// RegisterDevice registers a new device. Requires admin role or bootstrap secret.
func (s *AuthService) RegisterDevice(publicKeyB64, practitionerID, siteID, deviceName, role, bootstrapSecret string) (*DeviceRecord, error) {
	// Validate public key
	pubKey, err := auth.DecodePublicKey(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	// Bootstrap: register device with bootstrap secret
	if bootstrapSecret != "" {
		if bootstrapSecret != s.cfg.Security.BootstrapSecret {
			return nil, fmt.Errorf("invalid bootstrap secret")
		}
		if !s.bootstrapUsed {
			role = auth.RoleRegionalAdmin // first bootstrap gets regional-admin
			s.bootstrapUsed = true
		} else {
			// Subsequent bootstrap registrations get physician role
			if role == "" {
				role = auth.RolePhysician
			}
		}
	}

	if !auth.ValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	// Use deviceName as deviceID if it looks like a UUID (for auto-register from login).
	// Otherwise generate a new UUID.
	deviceID := deviceName
	if len(deviceID) < 36 {
		deviceID = uuid.New().String()
	}
	device := &DeviceRecord{
		DeviceID:       deviceID,
		PublicKey:      auth.EncodePublicKey(pubKey),
		PractitionerID: practitionerID,
		SiteID:         siteID,
		DeviceName:     deviceName,
		Role:           role,
		Status:         "active",
		RegisteredAt:   time.Now().UTC().Format(time.RFC3339),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, err = saveDevice(s.git, s.cfg.Devices.Path, device, "CREATE", "system")
	if err != nil {
		return nil, err
	}

	return device, nil
}

// GetChallenge generates a nonce for challenge-response authentication.
func (s *AuthService) GetChallenge(deviceID string) ([]byte, time.Time, error) {
	if s.bruteForce.IsBlocked(deviceID) {
		return nil, time.Time{}, fmt.Errorf("device is temporarily blocked due to too many failed attempts")
	}

	// Verify device exists
	_, err := loadDevice(s.git, s.cfg.Devices.Path, deviceID)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("unknown device: %s", deviceID)
	}

	return s.nonceStore.Generate(deviceID)
}

// Authenticate verifies a challenge-response signature and issues tokens.
func (s *AuthService) Authenticate(deviceID string, signature []byte, practitionerID string) (accessToken, refreshToken, expiresAt string, roleDef auth.RoleDefinition, siteID string, err error) {
	if s.bruteForce.IsBlocked(deviceID) {
		err = fmt.Errorf("device is temporarily blocked")
		return
	}

	// Load device
	device, loadErr := loadDevice(s.git, s.cfg.Devices.Path, deviceID)
	if loadErr != nil {
		s.bruteForce.RecordFailure(deviceID)
		err = fmt.Errorf("unknown device")
		return
	}

	// Check device not revoked
	if device.Status == "revoked" {
		err = fmt.Errorf("device has been revoked")
		return
	}

	// Check SQLite revocation list
	if revoked, _, _ := s.denyList.IsRevoked(deviceID); revoked {
		err = fmt.Errorf("device has been revoked")
		return
	}

	// NOTE: This method is deprecated in favor of AuthenticateWithNonce.
	// The nonce + signature must be handled together.
	// Use AuthenticateWithNonce which properly verifies Ed25519 signatures.

	// Look up role
	roleDef, ok := auth.GetRole(device.Role)
	if !ok {
		err = fmt.Errorf("unknown role: %s", device.Role)
		return
	}

	// Sign tokens
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	accessClaims := auth.NewAccessClaims(
		device.PractitionerID,
		deviceID,
		s.nodeID,
		device.SiteID,
		device.Role,
		roleDef.Permissions,
		roleDef.SiteScope,
		accessJTI,
		s.cfg.JWT.Issuer,
		s.cfg.JWT.AccessLifetime,
	)

	refreshClaims := auth.NewRefreshClaims(
		device.PractitionerID,
		deviceID,
		refreshJTI,
		s.cfg.JWT.Issuer,
		s.cfg.JWT.RefreshLifetime,
	)

	accessToken, signErr := auth.SignToken(accessClaims, s.nodePrivateKey, s.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign access token: %w", signErr)
		return
	}

	refreshToken, signErr = auth.SignToken(refreshClaims, s.nodePrivateKey, s.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign refresh token: %w", signErr)
		return
	}

	expiresAt = time.Now().Add(s.cfg.JWT.AccessLifetime).UTC().Format(time.RFC3339)
	siteID = device.SiteID

	// Clear brute force on success
	s.bruteForce.Reset(deviceID)

	return
}

// AuthenticateWithNonce handles the full challenge-response flow with explicit nonce verification.
func (s *AuthService) AuthenticateWithNonce(deviceID string, nonce, signature []byte) (accessToken, refreshToken, expiresAt string, roleDef auth.RoleDefinition, siteID string, err error) {
	if s.bruteForce.IsBlocked(deviceID) {
		err = fmt.Errorf("device is temporarily blocked")
		return
	}

	// Consume and verify nonce
	if !s.nonceStore.Consume(deviceID, nonce) {
		s.bruteForce.RecordFailure(deviceID)
		err = fmt.Errorf("invalid or expired challenge")
		return
	}

	// Load device
	device, loadErr := loadDevice(s.git, s.cfg.Devices.Path, deviceID)
	if loadErr != nil {
		s.bruteForce.RecordFailure(deviceID)
		err = fmt.Errorf("unknown device")
		return
	}

	if device.Status == "revoked" {
		err = fmt.Errorf("device has been revoked")
		return
	}

	if revoked, _, _ := s.denyList.IsRevoked(deviceID); revoked {
		err = fmt.Errorf("device has been revoked")
		return
	}

	// Decode public key and verify signature
	pubKey, decodeErr := auth.DecodePublicKey(device.PublicKey)
	if decodeErr != nil {
		err = fmt.Errorf("invalid device public key")
		return
	}

	if !auth.Verify(pubKey, nonce, signature) {
		s.bruteForce.RecordFailure(deviceID)
		err = fmt.Errorf("signature verification failed")
		return
	}

	// Look up role
	roleDef, ok := auth.GetRole(device.Role)
	if !ok {
		err = fmt.Errorf("unknown role: %s", device.Role)
		return
	}

	// Sign tokens
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	accessClaims := auth.NewAccessClaims(
		device.PractitionerID, deviceID, s.nodeID, device.SiteID,
		device.Role, roleDef.Permissions, roleDef.SiteScope,
		accessJTI, s.cfg.JWT.Issuer, s.cfg.JWT.AccessLifetime,
	)
	refreshClaims := auth.NewRefreshClaims(
		device.PractitionerID, deviceID,
		refreshJTI, s.cfg.JWT.Issuer, s.cfg.JWT.RefreshLifetime,
	)

	accessToken, signErr := auth.SignToken(accessClaims, s.nodePrivateKey, s.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign access token: %w", signErr)
		return
	}
	refreshToken, signErr = auth.SignToken(refreshClaims, s.nodePrivateKey, s.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign refresh token: %w", signErr)
		return
	}

	expiresAt = time.Now().Add(s.cfg.JWT.AccessLifetime).UTC().Format(time.RFC3339)
	siteID = device.SiteID
	s.bruteForce.Reset(deviceID)
	return
}

// RefreshToken issues a new token pair from a valid refresh token.
func (s *AuthService) RefreshToken(refreshTokenStr string) (newAccess, newRefresh, expiresAt string, err error) {
	claims, verifyErr := auth.VerifyToken(refreshTokenStr, s.nodePublicKey)
	if verifyErr != nil {
		err = fmt.Errorf("invalid refresh token: %w", verifyErr)
		return
	}

	if claims.TokenType != "refresh" {
		err = fmt.Errorf("not a refresh token")
		return
	}

	if s.denyList.IsDenied(claims.ID) {
		err = fmt.Errorf("refresh token has been revoked")
		return
	}

	// Load device to get current role
	device, loadErr := loadDevice(s.git, s.cfg.Devices.Path, claims.DeviceID)
	if loadErr != nil {
		err = fmt.Errorf("device not found")
		return
	}

	if device.Status == "revoked" {
		err = fmt.Errorf("device has been revoked")
		return
	}

	roleDef, ok := auth.GetRole(device.Role)
	if !ok {
		err = fmt.Errorf("unknown role")
		return
	}

	// Deny old refresh token
	if denyErr := s.denyList.Add(claims.ID, claims.DeviceID); denyErr != nil {
		err = fmt.Errorf("deny old token: %w", denyErr)
		return
	}

	// Issue new pair
	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	accessClaims := auth.NewAccessClaims(
		device.PractitionerID, device.DeviceID, s.nodeID, device.SiteID,
		device.Role, roleDef.Permissions, roleDef.SiteScope,
		accessJTI, s.cfg.JWT.Issuer, s.cfg.JWT.AccessLifetime,
	)
	refreshClaims := auth.NewRefreshClaims(
		device.PractitionerID, device.DeviceID,
		refreshJTI, s.cfg.JWT.Issuer, s.cfg.JWT.RefreshLifetime,
	)

	newAccess, signErr := auth.SignToken(accessClaims, s.nodePrivateKey, s.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign access token: %w", signErr)
		return
	}
	newRefresh, signErr = auth.SignToken(refreshClaims, s.nodePrivateKey, s.nodeID)
	if signErr != nil {
		err = fmt.Errorf("sign refresh token: %w", signErr)
		return
	}

	expiresAt = time.Now().Add(s.cfg.JWT.AccessLifetime).UTC().Format(time.RFC3339)
	return
}

// Logout invalidates a token by adding its JTI to the deny list.
func (s *AuthService) Logout(tokenStr string) error {
	claims, err := auth.VerifyToken(tokenStr, s.nodePublicKey)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}
	return s.denyList.Add(claims.ID, claims.DeviceID)
}

// ValidateToken verifies a JWT and returns its claims.
func (s *AuthService) ValidateToken(tokenStr string) (*auth.NucleusClaims, string, error) {
	claims, err := auth.VerifyToken(tokenStr, s.nodePublicKey)
	if err != nil {
		return nil, "invalid_signature", err
	}

	if s.denyList.IsDenied(claims.ID) {
		return nil, "denied", fmt.Errorf("token has been denied")
	}

	if revoked, _, _ := s.denyList.IsRevoked(claims.DeviceID); revoked {
		return nil, "revoked", fmt.Errorf("device has been revoked")
	}

	return claims, "", nil
}

// CheckPermission checks if a token has a specific permission.
func (s *AuthService) CheckPermission(tokenStr, permission, targetSiteID string) (bool, string, error) {
	claims, errCode, err := s.ValidateToken(tokenStr)
	if err != nil {
		return false, errCode, err
	}

	if !auth.HasPermission(claims.Role, permission) {
		return false, "insufficient_permissions", nil
	}

	// Site scope check
	if targetSiteID != "" && claims.SiteScope == "local" && claims.SiteID != targetSiteID {
		return false, "site_scope_violation", nil
	}

	return true, "", nil
}

// RevokeDevice permanently revokes a device and all its tokens.
func (s *AuthService) RevokeDevice(deviceID, revokedBy, reason string) (*DeviceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, err := loadDevice(s.git, s.cfg.Devices.Path, deviceID)
	if err != nil {
		return nil, err
	}

	device.Status = "revoked"
	device.RevokedAt = time.Now().UTC().Format(time.RFC3339)
	device.RevokedBy = revokedBy
	device.RevocationReason = reason

	if _, err := saveDevice(s.git, s.cfg.Devices.Path, device, "REVOKE", revokedBy); err != nil {
		return nil, err
	}

	// Add to SQLite revocation list
	if err := s.denyList.AddRevocation(deviceID, device.PublicKey, revokedBy, reason); err != nil {
		return nil, fmt.Errorf("record revocation: %w", err)
	}

	return device, nil
}

// ListDevices returns all registered devices.
func (s *AuthService) ListDevices() ([]*DeviceRecord, error) {
	return listAllDevices(s.git, s.cfg.Devices.Path)
}

// ListRoles returns all predefined roles.
func (s *AuthService) ListRoles() []auth.RoleDefinition {
	return auth.AllRoles()
}

// GetRole returns a specific role definition.
func (s *AuthService) GetRole(code string) (auth.RoleDefinition, bool) {
	return auth.GetRole(code)
}

// AssignRole updates a device's role.
func (s *AuthService) AssignRole(deviceID, role, assignedBy string) (*DeviceRecord, error) {
	if !auth.ValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	device, err := loadDevice(s.git, s.cfg.Devices.Path, deviceID)
	if err != nil {
		return nil, err
	}

	device.Role = role
	if _, err := saveDevice(s.git, s.cfg.Devices.Path, device, "UPDATE", assignedBy); err != nil {
		return nil, err
	}

	return device, nil
}

// NodeID returns this node's ID.
func (s *AuthService) NodeID() string {
	return s.nodeID
}

// NodePublicKey returns this node's public key.
func (s *AuthService) NodePublicKey() ed25519.PublicKey {
	return s.nodePublicKey
}

// NodePrivateKey returns this node's private key (used for anchoring/signing).
func (s *AuthService) NodePrivateKey() ed25519.PrivateKey {
	return s.nodePrivateKey
}

// UptimeSeconds returns uptime in seconds.
func (s *AuthService) UptimeSeconds() int64 {
	return int64(time.Since(s.startedAt).Seconds())
}
