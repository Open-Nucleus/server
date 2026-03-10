package envelope

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrantAccess_RoundTrip(t *testing.T) {
	git := testGitStore(t)
	km, err := NewFileKeyManager(testMasterKey(t), git)
	require.NoError(t, err)

	_, nodePriv, _ := ed25519.GenerateKey(rand.Reader)
	providerPub, providerPriv, _ := ed25519.GenerateKey(rand.Reader)

	patientID := "patient-grant-test"

	_, err = km.GetOrCreateKey(patientID)
	require.NoError(t, err)

	err = km.GrantAccess(patientID, providerPub, nodePriv)
	require.NoError(t, err)

	plaintext := []byte("sensitive patient data for grant test")
	ciphertext, err := km.Encrypt(patientID, plaintext)
	require.NoError(t, err)

	nodePub := nodePriv.Public().(ed25519.PublicKey)
	decrypted, err := km.DecryptFor(patientID, providerPriv, nodePub, ciphertext)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decrypted)
}

func TestRevokeAccess(t *testing.T) {
	git := testGitStore(t)
	km, err := NewFileKeyManager(testMasterKey(t), git)
	require.NoError(t, err)

	_, nodePriv, _ := ed25519.GenerateKey(rand.Reader)
	providerPub, providerPriv, _ := ed25519.GenerateKey(rand.Reader)

	patientID := "patient-revoke-test"

	_, err = km.GetOrCreateKey(patientID)
	require.NoError(t, err)
	err = km.GrantAccess(patientID, providerPub, nodePriv)
	require.NoError(t, err)

	err = km.RevokeAccess(patientID, providerPub)
	require.NoError(t, err)

	ciphertext, _ := km.Encrypt(patientID, []byte("test"))
	nodePub := nodePriv.Public().(ed25519.PublicKey)
	_, err = km.DecryptFor(patientID, providerPriv, nodePub, ciphertext)
	assert.Error(t, err, "should fail after revocation")
}

func TestGrantAccess_DifferentProviders(t *testing.T) {
	git := testGitStore(t)
	km, err := NewFileKeyManager(testMasterKey(t), git)
	require.NoError(t, err)

	_, nodePriv, _ := ed25519.GenerateKey(rand.Reader)
	providerPubA, providerPrivA, _ := ed25519.GenerateKey(rand.Reader)
	providerPubB, providerPrivB, _ := ed25519.GenerateKey(rand.Reader)

	patientID := "patient-multi"

	km.GetOrCreateKey(patientID)
	require.NoError(t, km.GrantAccess(patientID, providerPubA, nodePriv))
	require.NoError(t, km.GrantAccess(patientID, providerPubB, nodePriv))

	plaintext := []byte("shared patient data")
	ciphertext, _ := km.Encrypt(patientID, plaintext)
	nodePub := nodePriv.Public().(ed25519.PublicKey)

	decA, err := km.DecryptFor(patientID, providerPrivA, nodePub, ciphertext)
	require.NoError(t, err)
	decB, err := km.DecryptFor(patientID, providerPrivB, nodePub, ciphertext)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decA)
	assert.Equal(t, plaintext, decB)

	// Revoke A, B should still work
	km.RevokeAccess(patientID, providerPubA)

	_, err = km.DecryptFor(patientID, providerPrivA, nodePub, ciphertext)
	assert.Error(t, err, "A should fail after revocation")

	decB2, err := km.DecryptFor(patientID, providerPrivB, nodePub, ciphertext)
	require.NoError(t, err, "B should still work")
	assert.Equal(t, plaintext, decB2)
}

func TestDeriveGrantKey_Symmetry(t *testing.T) {
	_, privA, _ := ed25519.GenerateKey(rand.Reader)
	pubB, privB, _ := ed25519.GenerateKey(rand.Reader)
	pubA := privA.Public().(ed25519.PublicKey)

	keyAB, err := deriveGrantKey(privA, pubB)
	require.NoError(t, err)

	keyBA, err := deriveGrantKey(privB, pubA)
	require.NoError(t, err)

	assert.Len(t, keyAB, 32)
	assert.Equal(t, keyAB, keyBA, "ECDH shared keys should match")
}

func TestGrantPath(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	path := grantPath("patient-1", pub)

	assert.NotEmpty(t, path)
	assert.Contains(t, path, ".nucleus/grants/patient-1/")
	assert.Contains(t, path, ".grant")
}
