package auth

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Crypto Tests ---

func TestGenerateKeypair(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)
	assert.Len(t, pub, ed25519.PublicKeySize)
	assert.Len(t, priv, ed25519.PrivateKeySize)
}

func TestSignVerify_Roundtrip(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	msg := []byte("hello nucleus")
	sig := Sign(priv, msg)
	assert.True(t, Verify(pub, msg, sig))
}

func TestVerify_WrongKey(t *testing.T) {
	_, priv, err := GenerateKeypair()
	require.NoError(t, err)
	pub2, _, err := GenerateKeypair()
	require.NoError(t, err)

	msg := []byte("hello nucleus")
	sig := Sign(priv, msg)
	assert.False(t, Verify(pub2, msg, sig))
}

func TestEncodeDecodePublicKey(t *testing.T) {
	pub, _, err := GenerateKeypair()
	require.NoError(t, err)

	encoded := EncodePublicKey(pub)
	decoded, err := DecodePublicKey(encoded)
	require.NoError(t, err)
	assert.Equal(t, pub, decoded)
}

func TestDecodePublicKey_InvalidSize(t *testing.T) {
	_, err := DecodePublicKey("dG9vc2hvcnQ") // "tooshort" in base64
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

// --- JWT Tests ---

func TestJWT_SignVerify(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewAccessClaims("practitioner-1", "device-1", "node-1", "site-1", "physician", []string{"patient:read"}, "local", "jti-1", "open-nucleus-auth", time.Hour)
	token, err := SignToken(claims, priv, "key-1")
	require.NoError(t, err)

	parsed, err := VerifyToken(token, pub)
	require.NoError(t, err)
	assert.Equal(t, "practitioner-1", parsed.Subject)
	assert.Equal(t, "device-1", parsed.DeviceID)
	assert.Equal(t, "physician", parsed.Role)
	assert.Equal(t, "access", parsed.TokenType)
}

func TestJWT_Expired(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewAccessClaims("practitioner-1", "device-1", "node-1", "site-1", "physician", nil, "local", "jti-1", "open-nucleus-auth", -time.Hour)
	token, err := SignToken(claims, priv, "key-1")
	require.NoError(t, err)

	_, err = VerifyToken(token, pub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWT_WrongKey(t *testing.T) {
	_, priv, err := GenerateKeypair()
	require.NoError(t, err)
	pub2, _, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewAccessClaims("practitioner-1", "device-1", "node-1", "site-1", "physician", nil, "local", "jti-1", "open-nucleus-auth", time.Hour)
	token, err := SignToken(claims, priv, "key-1")
	require.NoError(t, err)

	_, err = VerifyToken(token, pub2)
	assert.Error(t, err)
}

func TestJWT_RefreshClaims(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewRefreshClaims("practitioner-1", "device-1", "jti-refresh-1", "open-nucleus-auth", 7*24*time.Hour)
	token, err := SignToken(claims, priv, "key-1")
	require.NoError(t, err)

	parsed, err := VerifyToken(token, pub)
	require.NoError(t, err)
	assert.Equal(t, "refresh", parsed.TokenType)
	assert.Equal(t, "device-1", parsed.DeviceID)
}

// --- Nonce Tests ---

func TestNonce_GenerateConsume(t *testing.T) {
	store := NewNonceStore(60 * time.Second)

	nonce, expiresAt, err := store.Generate("device-1")
	require.NoError(t, err)
	assert.Len(t, nonce, 32)
	assert.True(t, expiresAt.After(time.Now()))

	assert.True(t, store.Consume("device-1", nonce))
}

func TestNonce_ReplayRejected(t *testing.T) {
	store := NewNonceStore(60 * time.Second)

	nonce, _, err := store.Generate("device-1")
	require.NoError(t, err)

	assert.True(t, store.Consume("device-1", nonce))
	assert.False(t, store.Consume("device-1", nonce)) // replay
}

func TestNonce_Expired(t *testing.T) {
	store := NewNonceStore(1 * time.Millisecond)

	nonce, _, err := store.Generate("device-1")
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond)
	assert.False(t, store.Consume("device-1", nonce))
}

func TestNonce_WrongDevice(t *testing.T) {
	store := NewNonceStore(60 * time.Second)

	_, _, err := store.Generate("device-1")
	require.NoError(t, err)

	assert.False(t, store.Consume("device-2", []byte("fake")))
}

func TestNonce_Cleanup(t *testing.T) {
	store := NewNonceStore(1 * time.Millisecond)

	_, _, _ = store.Generate("device-1")
	_, _, _ = store.Generate("device-2")

	time.Sleep(5 * time.Millisecond)
	removed := store.Cleanup()
	assert.Equal(t, 2, removed)
}

// --- Role Tests ---

func TestHasPermission_Physician(t *testing.T) {
	assert.True(t, HasPermission(RolePhysician, PermPatientRead))
	assert.True(t, HasPermission(RolePhysician, PermPatientWrite))
	assert.True(t, HasPermission(RolePhysician, PermPatientDelete))
	assert.True(t, HasPermission(RolePhysician, PermConflictResolve))
	assert.False(t, HasPermission(RolePhysician, PermDeviceManage))
}

func TestHasPermission_CHW(t *testing.T) {
	assert.True(t, HasPermission(RoleCHW, PermPatientRead))
	assert.True(t, HasPermission(RoleCHW, PermPatientWrite))
	assert.False(t, HasPermission(RoleCHW, PermPatientDelete))
	assert.False(t, HasPermission(RoleCHW, PermConflictResolve))
	assert.False(t, HasPermission(RoleCHW, PermDeviceManage))
}

func TestHasPermission_RegionalAdmin(t *testing.T) {
	assert.True(t, HasPermission(RoleRegionalAdmin, PermDeviceManage))
	assert.True(t, HasPermission(RoleRegionalAdmin, PermRoleAssign))
	assert.True(t, HasPermission(RoleRegionalAdmin, PermSyncBundle))
}

func TestHasPermission_InvalidRole(t *testing.T) {
	assert.False(t, HasPermission("nonexistent", PermPatientRead))
}

func TestAllRoles(t *testing.T) {
	roles := AllRoles()
	assert.Len(t, roles, 5)
}

func TestGetRole(t *testing.T) {
	role, ok := GetRole(RolePhysician)
	assert.True(t, ok)
	assert.Equal(t, "Physician", role.Display)
	assert.Equal(t, "local", role.SiteScope)

	role, ok = GetRole(RoleRegionalAdmin)
	assert.True(t, ok)
	assert.Equal(t, "regional", role.SiteScope)

	_, ok = GetRole("nonexistent")
	assert.False(t, ok)
}

// --- BruteForce Tests ---

func TestBruteForce_NotBlocked(t *testing.T) {
	guard := NewBruteForceGuard(10, 60*time.Second)
	assert.False(t, guard.IsBlocked("device-1"))
}

func TestBruteForce_Blocked(t *testing.T) {
	guard := NewBruteForceGuard(3, 60*time.Second)

	for i := 0; i < 3; i++ {
		guard.RecordFailure("device-1")
	}
	assert.True(t, guard.IsBlocked("device-1"))
	assert.False(t, guard.IsBlocked("device-2")) // different device
}

func TestBruteForce_Reset(t *testing.T) {
	guard := NewBruteForceGuard(3, 60*time.Second)

	for i := 0; i < 3; i++ {
		guard.RecordFailure("device-1")
	}
	assert.True(t, guard.IsBlocked("device-1"))

	guard.Reset("device-1")
	assert.False(t, guard.IsBlocked("device-1"))
}

func TestBruteForce_WindowExpiry(t *testing.T) {
	guard := NewBruteForceGuard(3, 10*time.Millisecond)

	for i := 0; i < 3; i++ {
		guard.RecordFailure("device-1")
	}
	assert.True(t, guard.IsBlocked("device-1"))

	time.Sleep(20 * time.Millisecond)
	assert.False(t, guard.IsBlocked("device-1"))
}

// --- KeyStore Tests ---

func TestMemoryKeyStore_CRUD(t *testing.T) {
	ks := NewMemoryKeyStore()
	_, priv, err := GenerateKeypair()
	require.NoError(t, err)

	assert.False(t, ks.HasKey("node-1"))

	err = ks.StoreKey("node-1", priv)
	require.NoError(t, err)
	assert.True(t, ks.HasKey("node-1"))

	loaded, err := ks.LoadKey("node-1")
	require.NoError(t, err)
	assert.Equal(t, priv, loaded)

	pub, err := ks.LoadPublicKey("node-1")
	require.NoError(t, err)
	assert.Equal(t, priv.Public(), pub)

	err = ks.DeleteKey("node-1")
	require.NoError(t, err)
	assert.False(t, ks.HasKey("node-1"))
}

func TestFileKeyStore_CRUD(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "keys")
	ks, err := NewFileKeyStore(dir)
	require.NoError(t, err)

	_, priv, err := GenerateKeypair()
	require.NoError(t, err)

	assert.False(t, ks.HasKey("node-1"))

	err = ks.StoreKey("node-1", priv)
	require.NoError(t, err)
	assert.True(t, ks.HasKey("node-1"))

	// Verify file permissions
	info, err := os.Stat(ks.keyPath("node-1"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	loaded, err := ks.LoadKey("node-1")
	require.NoError(t, err)
	assert.Equal(t, priv, loaded)

	err = ks.DeleteKey("node-1")
	require.NoError(t, err)
	assert.False(t, ks.HasKey("node-1"))
}

// --- SMART JWT Tests ---

func TestJWT_SmartAccessClaims(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewSmartAccessClaims(
		"practitioner-1", "device-1", "node-1", "site-1", "physician",
		[]string{"patient:read"}, "local",
		"patient/Patient.r launch", "client-abc", "Practitioner/practitioner-1",
		"patient-123", "encounter-456",
		"jti-smart-1", "open-nucleus-auth", time.Hour,
	)

	assert.Equal(t, "patient/Patient.r launch", claims.Scope)
	assert.Equal(t, "client-abc", claims.ClientID)
	assert.Equal(t, "Practitioner/practitioner-1", claims.FHIRUser)
	assert.Equal(t, "patient-123", claims.LaunchPatient)
	assert.Equal(t, "encounter-456", claims.LaunchEncounter)
	assert.Equal(t, "access", claims.TokenType)

	token, err := SignToken(claims, priv, "key-1")
	require.NoError(t, err)

	parsed, err := VerifyToken(token, pub)
	require.NoError(t, err)
	assert.Equal(t, "patient/Patient.r launch", parsed.Scope)
	assert.Equal(t, "client-abc", parsed.ClientID)
	assert.Equal(t, "Practitioner/practitioner-1", parsed.FHIRUser)
	assert.Equal(t, "patient-123", parsed.LaunchPatient)
	assert.Equal(t, "encounter-456", parsed.LaunchEncounter)
	assert.Equal(t, "physician", parsed.Role)
	assert.Equal(t, "device-1", parsed.DeviceID)
}

func TestJWT_SmartClaimsEmpty_WhenNotSmart(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewAccessClaims("practitioner-1", "device-1", "node-1", "site-1", "physician", []string{"patient:read"}, "local", "jti-1", "open-nucleus-auth", time.Hour)
	token, err := SignToken(claims, priv, "key-1")
	require.NoError(t, err)

	parsed, err := VerifyToken(token, pub)
	require.NoError(t, err)
	assert.Empty(t, parsed.Scope)
	assert.Empty(t, parsed.ClientID)
	assert.Empty(t, parsed.LaunchPatient)
}

func TestJWT_SmartRoundtrip_PreservesFields(t *testing.T) {
	pub, priv, err := GenerateKeypair()
	require.NoError(t, err)

	claims := NewSmartAccessClaims(
		"sub-1", "dev-1", "node-1", "site-1", "nurse",
		[]string{"observation:read"}, "local",
		"user/Observation.rs fhirUser", "my-client", "Practitioner/pract-1",
		"", "", // no patient/encounter context
		"jti-2", "issuer", 30*time.Minute,
	)

	token, err := SignToken(claims, priv, "key-2")
	require.NoError(t, err)

	parsed, err := VerifyToken(token, pub)
	require.NoError(t, err)
	assert.Equal(t, "user/Observation.rs fhirUser", parsed.Scope)
	assert.Equal(t, "my-client", parsed.ClientID)
	assert.Equal(t, "Practitioner/pract-1", parsed.FHIRUser)
	assert.Empty(t, parsed.LaunchPatient)
	assert.Empty(t, parsed.LaunchEncounter)
}
