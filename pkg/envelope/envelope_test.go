package envelope

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testMasterKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	require.NoError(t, err)
	return key
}

func testGitStore(t *testing.T) gitstore.Store {
	t.Helper()
	dir := t.TempDir()
	gs, err := gitstore.NewStore(filepath.Join(dir, "repo"), "test", "test@test.com")
	require.NoError(t, err)
	return gs
}

func TestNewFileKeyManager(t *testing.T) {
	git := testGitStore(t)

	t.Run("valid 32-byte key", func(t *testing.T) {
		km, err := NewFileKeyManager(testMasterKey(t), git)
		require.NoError(t, err)
		assert.NotNil(t, km)
	})

	t.Run("rejects short key", func(t *testing.T) {
		_, err := NewFileKeyManager(make([]byte, 16), git)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "32 bytes")
	})

	t.Run("rejects long key", func(t *testing.T) {
		_, err := NewFileKeyManager(make([]byte, 64), git)
		assert.Error(t, err)
	})
}

func TestGetOrCreateKey(t *testing.T) {
	git := testGitStore(t)
	mk := testMasterKey(t)
	km, err := NewFileKeyManager(mk, git)
	require.NoError(t, err)

	t.Run("creates new key", func(t *testing.T) {
		dek, err := km.GetOrCreateKey("patient-001")
		require.NoError(t, err)
		assert.Len(t, dek, 32)
	})

	t.Run("returns same key on second call", func(t *testing.T) {
		dek1, err := km.GetOrCreateKey("patient-002")
		require.NoError(t, err)
		dek2, err := km.GetOrCreateKey("patient-002")
		require.NoError(t, err)
		assert.Equal(t, dek1, dek2)
	})

	t.Run("different patients get different keys", func(t *testing.T) {
		dek1, err := km.GetOrCreateKey("patient-A")
		require.NoError(t, err)
		dek2, err := km.GetOrCreateKey("patient-B")
		require.NoError(t, err)
		assert.NotEqual(t, dek1, dek2)
	})

	t.Run("rejects empty patient ID", func(t *testing.T) {
		_, err := km.GetOrCreateKey("")
		assert.Error(t, err)
	})

	t.Run("key survives new manager instance", func(t *testing.T) {
		dek1, err := km.GetOrCreateKey("patient-persist")
		require.NoError(t, err)

		// New manager with same master key and git store
		km2, err := NewFileKeyManager(mk, git)
		require.NoError(t, err)

		dek2, err := km2.GetOrCreateKey("patient-persist")
		require.NoError(t, err)
		assert.Equal(t, dek1, dek2, "key should be loaded from Git")
	})

	t.Run("wrong master key cannot unwrap", func(t *testing.T) {
		_, err := km.GetOrCreateKey("patient-wrongkey")
		require.NoError(t, err)

		// New manager with different master key
		differentMK := testMasterKey(t)
		km2, err := NewFileKeyManager(differentMK, git)
		require.NoError(t, err)

		_, err = km2.GetOrCreateKey("patient-wrongkey")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unwrap")
	})
}

func TestEncryptDecrypt(t *testing.T) {
	git := testGitStore(t)
	km, err := NewFileKeyManager(testMasterKey(t), git)
	require.NoError(t, err)

	plaintext := []byte(`{"resourceType":"Patient","id":"123","name":[{"family":"Smith"}]}`)

	t.Run("round trip", func(t *testing.T) {
		ct, err := km.Encrypt("patient-enc", plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, ct, "ciphertext should differ from plaintext")

		pt, err := km.Decrypt("patient-enc", ct)
		require.NoError(t, err)
		assert.Equal(t, plaintext, pt)
	})

	t.Run("different nonces produce different ciphertexts", func(t *testing.T) {
		ct1, err := km.Encrypt("patient-nonce", plaintext)
		require.NoError(t, err)
		ct2, err := km.Encrypt("patient-nonce", plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, ct1, ct2, "should use random nonce each time")
	})

	t.Run("wrong patient key cannot decrypt", func(t *testing.T) {
		ct, err := km.Encrypt("patient-X", plaintext)
		require.NoError(t, err)

		_, err = km.Decrypt("patient-Y", ct)
		assert.Error(t, err, "different patient key should fail")
	})

	t.Run("empty plaintext", func(t *testing.T) {
		ct, err := km.Encrypt("patient-empty", []byte{})
		require.NoError(t, err)
		pt, err := km.Decrypt("patient-empty", ct)
		require.NoError(t, err)
		assert.Empty(t, pt)
	})

	t.Run("large plaintext", func(t *testing.T) {
		big := make([]byte, 1<<20) // 1 MB
		_, err := io.ReadFull(rand.Reader, big)
		require.NoError(t, err)

		ct, err := km.Encrypt("patient-big", big)
		require.NoError(t, err)
		pt, err := km.Decrypt("patient-big", ct)
		require.NoError(t, err)
		assert.Equal(t, big, pt)
	})
}

func TestDestroyKey(t *testing.T) {
	git := testGitStore(t)
	km, err := NewFileKeyManager(testMasterKey(t), git)
	require.NoError(t, err)

	plaintext := []byte(`{"resourceType":"Patient","id":"destroy-me"}`)

	t.Run("destroy makes data unreadable", func(t *testing.T) {
		ct, err := km.Encrypt("patient-destroy", plaintext)
		require.NoError(t, err)

		// Verify it works before destroy
		pt, err := km.Decrypt("patient-destroy", ct)
		require.NoError(t, err)
		assert.Equal(t, plaintext, pt)

		// Destroy
		err = km.DestroyKey("patient-destroy")
		require.NoError(t, err)

		// Cannot decrypt
		_, err = km.Decrypt("patient-destroy", ct)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destroyed")

		// Cannot get key
		_, err = km.GetOrCreateKey("patient-destroy")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destroyed")
	})

	t.Run("IsKeyDestroyed", func(t *testing.T) {
		_, err := km.Encrypt("patient-check", plaintext)
		require.NoError(t, err)

		assert.False(t, km.IsKeyDestroyed("patient-check"))
		require.NoError(t, km.DestroyKey("patient-check"))
		assert.True(t, km.IsKeyDestroyed("patient-check"))
	})

	t.Run("destroy non-existent key", func(t *testing.T) {
		err := km.DestroyKey("patient-never-existed")
		require.NoError(t, err, "destroying non-existent key should not error")
	})

	t.Run("rejects empty patient ID", func(t *testing.T) {
		err := km.DestroyKey("")
		assert.Error(t, err)
	})
}

func TestSystemKey(t *testing.T) {
	git := testGitStore(t)
	km, err := NewFileKeyManager(testMasterKey(t), git)
	require.NoError(t, err)

	plaintext := []byte(`{"resourceType":"Practitioner","id":"prac-001"}`)

	ct, err := km.Encrypt(SystemKeyID, plaintext)
	require.NoError(t, err)

	pt, err := km.Decrypt(SystemKeyID, ct)
	require.NoError(t, err)
	assert.Equal(t, plaintext, pt)
}

func TestMasterKeyFromEnv(t *testing.T) {
	t.Run("valid hex key", func(t *testing.T) {
		key := testMasterKey(t)
		t.Setenv(masterKeyEnvVar, hex.EncodeToString(key))
		got, err := MasterKeyFromEnv()
		require.NoError(t, err)
		assert.Equal(t, key, got)
	})

	t.Run("missing env var", func(t *testing.T) {
		t.Setenv(masterKeyEnvVar, "")
		_, err := MasterKeyFromEnv()
		assert.Error(t, err)
	})

	t.Run("invalid hex", func(t *testing.T) {
		t.Setenv(masterKeyEnvVar, "not-hex")
		_, err := MasterKeyFromEnv()
		assert.Error(t, err)
	})

	t.Run("wrong length", func(t *testing.T) {
		t.Setenv(masterKeyEnvVar, hex.EncodeToString(make([]byte, 16)))
		_, err := MasterKeyFromEnv()
		assert.Error(t, err)
	})
}

func TestMasterKeyFromFile(t *testing.T) {
	t.Run("hex file", func(t *testing.T) {
		key := testMasterKey(t)
		path := filepath.Join(t.TempDir(), "master.key")
		require.NoError(t, os.WriteFile(path, []byte(hex.EncodeToString(key)), 0600))

		got, err := MasterKeyFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, key, got)
	})

	t.Run("hex file with trailing newline", func(t *testing.T) {
		key := testMasterKey(t)
		path := filepath.Join(t.TempDir(), "master.key")
		require.NoError(t, os.WriteFile(path, []byte(hex.EncodeToString(key)+"\n"), 0600))

		got, err := MasterKeyFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, key, got)
	})

	t.Run("raw bytes file", func(t *testing.T) {
		key := testMasterKey(t)
		path := filepath.Join(t.TempDir(), "master.key")
		require.NoError(t, os.WriteFile(path, key, 0600))

		got, err := MasterKeyFromFile(path)
		require.NoError(t, err)
		assert.Equal(t, key, got)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := MasterKeyFromFile("/nonexistent/path")
		assert.Error(t, err)
	})

	t.Run("wrong size", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "bad.key")
		require.NoError(t, os.WriteFile(path, []byte("too short"), 0600))
		_, err := MasterKeyFromFile(path)
		assert.Error(t, err)
	})
}

func TestGenerateMasterKey(t *testing.T) {
	key1, err := GenerateMasterKey()
	require.NoError(t, err)
	assert.Len(t, key1, 32)

	key2, err := GenerateMasterKey()
	require.NoError(t, err)
	assert.NotEqual(t, key1, key2, "should be random")
}
