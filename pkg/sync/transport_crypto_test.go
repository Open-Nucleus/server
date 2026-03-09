package sync_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	synccrypto "github.com/FibrinLab/open-nucleus/pkg/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateEd25519Pair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return pub, priv
}

func TestDeriveSharedKey(t *testing.T) {
	// Both sides should derive the same shared key.
	pubA, privA := generateEd25519Pair(t)
	pubB, privB := generateEd25519Pair(t)

	// Alice derives using her private key and Bob's public key.
	keyAB, err := synccrypto.DeriveSharedKey(privA, pubB)
	require.NoError(t, err)
	assert.Len(t, keyAB, 32)

	// Bob derives using his private key and Alice's public key.
	keyBA, err := synccrypto.DeriveSharedKey(privB, pubA)
	require.NoError(t, err)
	assert.Len(t, keyBA, 32)

	// Both must be identical.
	assert.Equal(t, keyAB, keyBA, "shared keys must match regardless of which side derives")
}

func TestDeriveSharedKeyDeterministic(t *testing.T) {
	// Same inputs must always produce the same output.
	pubA, privA := generateEd25519Pair(t)
	_, privB := generateEd25519Pair(t)

	key1, err := synccrypto.DeriveSharedKey(privB, pubA)
	require.NoError(t, err)

	key2, err := synccrypto.DeriveSharedKey(privB, pubA)
	require.NoError(t, err)

	assert.Equal(t, key1, key2, "same inputs must produce same derived key")

	// Different peer → different key
	pubC, _ := generateEd25519Pair(t)
	keyDiff, err := synccrypto.DeriveSharedKey(privA, pubC)
	require.NoError(t, err)
	assert.NotEqual(t, key1, keyDiff, "different peer must produce different key")
}

func TestEncryptDecrypt(t *testing.T) {
	pubA, privA := generateEd25519Pair(t)
	pubB, privB := generateEd25519Pair(t)

	sharedKey, err := synccrypto.DeriveSharedKey(privA, pubB)
	require.NoError(t, err)

	plaintext := []byte(`{"resourceType":"Patient","id":"p1","name":[{"family":"Smith"}]}`)

	// Encrypt
	ciphertext, err := synccrypto.EncryptPayload(sharedKey, plaintext)
	require.NoError(t, err)
	assert.NotEqual(t, plaintext, ciphertext)
	assert.Greater(t, len(ciphertext), len(plaintext), "ciphertext should be larger (nonce + tag)")

	// Decrypt with the same shared key derived from the other side
	sharedKey2, err := synccrypto.DeriveSharedKey(privB, pubA)
	require.NoError(t, err)

	decrypted, err := synccrypto.DecryptPayload(sharedKey2, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptEmptyPayload(t *testing.T) {
	_, privA := generateEd25519Pair(t)
	pubB, _ := generateEd25519Pair(t)

	sharedKey, err := synccrypto.DeriveSharedKey(privA, pubB)
	require.NoError(t, err)

	ciphertext, err := synccrypto.EncryptPayload(sharedKey, []byte{})
	require.NoError(t, err)

	decrypted, err := synccrypto.DecryptPayload(sharedKey, ciphertext)
	require.NoError(t, err)
	assert.Empty(t, decrypted)
}

func TestEncryptDecryptLargePayload(t *testing.T) {
	_, privA := generateEd25519Pair(t)
	pubB, _ := generateEd25519Pair(t)

	sharedKey, err := synccrypto.DeriveSharedKey(privA, pubB)
	require.NoError(t, err)

	// 1 MB payload
	plaintext := make([]byte, 1<<20)
	_, err = rand.Read(plaintext)
	require.NoError(t, err)

	ciphertext, err := synccrypto.EncryptPayload(sharedKey, plaintext)
	require.NoError(t, err)

	decrypted, err := synccrypto.DecryptPayload(sharedKey, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptWrongKey(t *testing.T) {
	_, privA := generateEd25519Pair(t)
	pubB, _ := generateEd25519Pair(t)
	pubC, _ := generateEd25519Pair(t)

	// Key for A↔B
	keyAB, err := synccrypto.DeriveSharedKey(privA, pubB)
	require.NoError(t, err)

	// Key for A↔C (different)
	keyAC, err := synccrypto.DeriveSharedKey(privA, pubC)
	require.NoError(t, err)

	assert.NotEqual(t, keyAB, keyAC)

	plaintext := []byte("clinical data that must remain confidential")
	ciphertext, err := synccrypto.EncryptPayload(keyAB, plaintext)
	require.NoError(t, err)

	// Decrypting with the wrong key must fail.
	_, err = synccrypto.DecryptPayload(keyAC, ciphertext)
	assert.Error(t, err, "decryption with wrong key must fail")
	assert.Contains(t, err.Error(), "GCM open failed")
}

func TestDeriveSharedKeyInvalidInputs(t *testing.T) {
	pub, priv := generateEd25519Pair(t)

	// Short private key
	_, err := synccrypto.DeriveSharedKey(priv[:16], pub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key size")

	// Short public key
	_, err = synccrypto.DeriveSharedKey(priv, pub[:16])
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid public key size")
}

func TestEncryptPayloadInvalidKey(t *testing.T) {
	_, err := synccrypto.EncryptPayload([]byte("short"), []byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key must be 32 bytes")
}

func TestDecryptPayloadInvalidKey(t *testing.T) {
	_, err := synccrypto.DecryptPayload([]byte("short"), []byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key must be 32 bytes")
}

func TestDecryptPayloadTooShort(t *testing.T) {
	key := make([]byte, 32)
	_, err := synccrypto.DecryptPayload(key, []byte("tiny"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ciphertext too short")
}

func TestEncryptDecryptNonceUniqueness(t *testing.T) {
	_, privA := generateEd25519Pair(t)
	pubB, _ := generateEd25519Pair(t)

	sharedKey, err := synccrypto.DeriveSharedKey(privA, pubB)
	require.NoError(t, err)

	plaintext := []byte("same data encrypted twice")

	ct1, err := synccrypto.EncryptPayload(sharedKey, plaintext)
	require.NoError(t, err)

	ct2, err := synccrypto.EncryptPayload(sharedKey, plaintext)
	require.NoError(t, err)

	// Ciphertexts must differ (different random nonces).
	assert.NotEqual(t, ct1, ct2, "two encryptions of same data must produce different ciphertexts")

	// But both must decrypt to the same plaintext.
	dec1, err := synccrypto.DecryptPayload(sharedKey, ct1)
	require.NoError(t, err)
	dec2, err := synccrypto.DecryptPayload(sharedKey, ct2)
	require.NoError(t, err)
	assert.Equal(t, dec1, dec2)
}
