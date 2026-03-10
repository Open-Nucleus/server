// Package envelope implements per-patient envelope encryption for FHIR resources.
//
// Key hierarchy: Master Key (AES-256, from env/file) wraps per-patient DEKs (AES-256-GCM).
// Wrapped keys are stored in Git at .nucleus/keys/{patient_id}.key and sync with data.
// Non-patient resources use a system key at .nucleus/keys/_system.key.
//
// Crypto-erasure: destroying a patient's wrapped key renders all their data permanently
// unreadable, even though ciphertext remains in Git history.
package envelope

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
)

const (
	// keyVersion is prepended to wrapped keys for future key rotation.
	keyVersion byte = 0x01

	// dekSize is the size of per-patient data encryption keys.
	dekSize = 32 // AES-256

	// nonceSize is the GCM nonce size.
	nonceSize = 12

	// SystemKeyID is used for non-patient resources (Practitioner, Organization, Location).
	SystemKeyID = "_system"

	// masterKeyEnvVar is the environment variable for the hex-encoded master key.
	masterKeyEnvVar = "NUCLEUS_MASTER_KEY"

	// keysDirPrefix is the Git path prefix for wrapped keys.
	keysDirPrefix = ".nucleus/keys/"
)

// KeyManager provides per-patient envelope encryption with crypto-erasure support.
type KeyManager interface {
	// GetOrCreateKey returns the unwrapped DEK for a patient, creating one if needed.
	GetOrCreateKey(patientID string) ([]byte, error)

	// DestroyKey permanently deletes the wrapped DEK, making all patient data unreadable.
	DestroyKey(patientID string) error

	// Encrypt encrypts plaintext using the patient's DEK.
	Encrypt(patientID string, plaintext []byte) ([]byte, error)

	// Decrypt decrypts ciphertext using the patient's DEK.
	Decrypt(patientID string, ciphertext []byte) ([]byte, error)

	// IsKeyDestroyed returns true if the patient's key has been destroyed.
	IsKeyDestroyed(patientID string) bool
}

// FileKeyManager implements KeyManager using Git-stored wrapped keys.
type FileKeyManager struct {
	mu        sync.RWMutex
	masterKey []byte            // 32-byte AES-256 master key
	cache     map[string][]byte // patientID → unwrapped DEK
	destroyed map[string]bool   // patientID → destroyed flag
	git       gitstore.Store
}

// NewFileKeyManager creates a KeyManager backed by Git-stored wrapped keys.
// masterKey must be exactly 32 bytes (AES-256).
func NewFileKeyManager(masterKey []byte, git gitstore.Store) (*FileKeyManager, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("envelope: master key must be 32 bytes, got %d", len(masterKey))
	}
	return &FileKeyManager{
		masterKey: append([]byte(nil), masterKey...), // defensive copy
		cache:     make(map[string][]byte),
		destroyed: make(map[string]bool),
		git:       git,
	}, nil
}

// MasterKeyFromEnv reads the master key from NUCLEUS_MASTER_KEY env var (hex-encoded).
func MasterKeyFromEnv() ([]byte, error) {
	hexKey := os.Getenv(masterKeyEnvVar)
	if hexKey == "" {
		return nil, fmt.Errorf("envelope: %s not set", masterKeyEnvVar)
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("envelope: invalid hex in %s: %w", masterKeyEnvVar, err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("envelope: %s must be 64 hex chars (32 bytes), got %d bytes", masterKeyEnvVar, len(key))
	}
	return key, nil
}

// MasterKeyFromFile reads the master key from a file (raw 32 bytes or 64 hex chars).
func MasterKeyFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("envelope: reading master key file: %w", err)
	}
	// Try hex first (64 chars = 32 bytes)
	if len(data) == 64 || (len(data) == 65 && data[64] == '\n') {
		trimmed := data
		if len(trimmed) == 65 {
			trimmed = trimmed[:64]
		}
		key, err := hex.DecodeString(string(trimmed))
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}
	// Raw bytes
	if len(data) == 32 {
		return data, nil
	}
	return nil, fmt.Errorf("envelope: master key file must be 32 raw bytes or 64 hex chars, got %d bytes", len(data))
}

// GenerateMasterKey creates a random 32-byte master key.
func GenerateMasterKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("envelope: generating master key: %w", err)
	}
	return key, nil
}

func (m *FileKeyManager) GetOrCreateKey(patientID string) ([]byte, error) {
	if patientID == "" {
		return nil, errors.New("envelope: empty patient ID")
	}

	// Check cache (read lock)
	m.mu.RLock()
	if m.destroyed[patientID] {
		m.mu.RUnlock()
		return nil, fmt.Errorf("envelope: key destroyed for patient %s", patientID)
	}
	if dek, ok := m.cache[patientID]; ok {
		m.mu.RUnlock()
		return dek, nil
	}
	m.mu.RUnlock()

	// Upgrade to write lock
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after lock upgrade
	if m.destroyed[patientID] {
		return nil, fmt.Errorf("envelope: key destroyed for patient %s", patientID)
	}
	if dek, ok := m.cache[patientID]; ok {
		return dek, nil
	}

	// Try loading from Git
	keyPath := keysDirPrefix + patientID + ".key"
	wrappedData, err := m.git.Read(keyPath)
	if err == nil && len(wrappedData) > 0 {
		dek, err := m.unwrapKey(wrappedData)
		if err != nil {
			return nil, fmt.Errorf("envelope: unwrapping key for %s: %w", patientID, err)
		}
		m.cache[patientID] = dek
		return dek, nil
	}

	// Generate new DEK
	dek := make([]byte, dekSize)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, fmt.Errorf("envelope: generating DEK for %s: %w", patientID, err)
	}

	// Wrap and store
	wrapped, err := m.wrapKey(dek)
	if err != nil {
		return nil, fmt.Errorf("envelope: wrapping key for %s: %w", patientID, err)
	}

	if _, err := m.git.WriteAndCommit(keyPath, wrapped, gitstore.CommitMessage{
		ResourceType: "Key",
		Operation:    "CREATE",
		ResourceID:   patientID,
	}); err != nil {
		return nil, fmt.Errorf("envelope: storing wrapped key for %s: %w", patientID, err)
	}

	m.cache[patientID] = dek
	return dek, nil
}

func (m *FileKeyManager) DestroyKey(patientID string) error {
	if patientID == "" {
		return errors.New("envelope: empty patient ID")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete from Git by writing an empty tombstone
	keyPath := keysDirPrefix + patientID + ".key"
	if _, err := m.git.WriteAndCommit(keyPath, []byte{}, gitstore.CommitMessage{
		ResourceType: "Key",
		Operation:    "DESTROY",
		ResourceID:   patientID,
	}); err != nil {
		return fmt.Errorf("envelope: destroying key for %s: %w", patientID, err)
	}

	// Remove from cache and mark destroyed
	delete(m.cache, patientID)
	m.destroyed[patientID] = true
	return nil
}

func (m *FileKeyManager) Encrypt(patientID string, plaintext []byte) ([]byte, error) {
	dek, err := m.GetOrCreateKey(patientID)
	if err != nil {
		return nil, err
	}
	return encryptAESGCM(dek, plaintext)
}

func (m *FileKeyManager) Decrypt(patientID string, ciphertext []byte) ([]byte, error) {
	dek, err := m.GetOrCreateKey(patientID)
	if err != nil {
		return nil, err
	}
	return decryptAESGCM(dek, ciphertext)
}

func (m *FileKeyManager) IsKeyDestroyed(patientID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.destroyed[patientID]
}

// wrapKey encrypts a DEK with the master key using AES-256-GCM.
// Format: [1-byte version][12-byte nonce][sealed DEK (32 + 16 tag)]
func (m *FileKeyManager) wrapKey(dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, dek, nil)

	result := make([]byte, 1+len(nonce)+len(sealed))
	result[0] = keyVersion
	copy(result[1:], nonce)
	copy(result[1+len(nonce):], sealed)
	return result, nil
}

// unwrapKey decrypts a wrapped DEK using the master key.
func (m *FileKeyManager) unwrapKey(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, errors.New("envelope: wrapped key too short")
	}
	if data[0] != keyVersion {
		return nil, fmt.Errorf("envelope: unsupported key version %d", data[0])
	}
	data = data[1:] // strip version byte

	block, err := aes.NewCipher(m.masterKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("envelope: wrapped key data too short for nonce")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	dek, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("envelope: unwrap failed (wrong master key?): %w", err)
	}
	if len(dek) != dekSize {
		return nil, fmt.Errorf("envelope: unwrapped key is %d bytes, expected %d", len(dek), dekSize)
	}
	return dek, nil
}

// encryptAESGCM encrypts plaintext with AES-256-GCM.
// Format: [12-byte nonce][ciphertext + 16-byte tag]
func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, plaintext, nil)

	result := make([]byte, len(nonce)+len(sealed))
	copy(result, nonce)
	copy(result[len(nonce):], sealed)
	return result, nil
}

// decryptAESGCM decrypts AES-256-GCM ciphertext.
func decryptAESGCM(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, errors.New("envelope: ciphertext too short")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
