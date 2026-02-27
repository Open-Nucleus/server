package auth

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// KeyStore provides storage for Ed25519 key pairs.
type KeyStore interface {
	StoreKey(id string, privateKey ed25519.PrivateKey) error
	LoadKey(id string) (ed25519.PrivateKey, error)
	LoadPublicKey(id string) (ed25519.PublicKey, error)
	DeleteKey(id string) error
	HasKey(id string) bool
}

// MemoryKeyStore stores keys in memory (for development/testing).
type MemoryKeyStore struct {
	mu   sync.RWMutex
	keys map[string]ed25519.PrivateKey
}

// NewMemoryKeyStore creates a new in-memory key store.
func NewMemoryKeyStore() *MemoryKeyStore {
	return &MemoryKeyStore{
		keys: make(map[string]ed25519.PrivateKey),
	}
}

func (m *MemoryKeyStore) StoreKey(id string, privateKey ed25519.PrivateKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.keys[id] = privateKey
	return nil
}

func (m *MemoryKeyStore) LoadKey(id string) (ed25519.PrivateKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	key, ok := m.keys[id]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", id)
	}
	return key, nil
}

func (m *MemoryKeyStore) LoadPublicKey(id string) (ed25519.PublicKey, error) {
	priv, err := m.LoadKey(id)
	if err != nil {
		return nil, err
	}
	return priv.Public().(ed25519.PublicKey), nil
}

func (m *MemoryKeyStore) DeleteKey(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.keys, id)
	return nil
}

func (m *MemoryKeyStore) HasKey(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.keys[id]
	return ok
}

// FileKeyStore stores keys on the filesystem with 0600 permissions.
type FileKeyStore struct {
	dir string
}

// NewFileKeyStore creates a new file-based key store.
func NewFileKeyStore(dir string) (*FileKeyStore, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create keystore dir: %w", err)
	}
	return &FileKeyStore{dir: dir}, nil
}

func (f *FileKeyStore) keyPath(id string) string {
	return filepath.Join(f.dir, id+".key")
}

func (f *FileKeyStore) StoreKey(id string, privateKey ed25519.PrivateKey) error {
	return os.WriteFile(f.keyPath(id), privateKey, 0o600)
}

func (f *FileKeyStore) LoadKey(id string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(f.keyPath(id))
	if err != nil {
		return nil, fmt.Errorf("load key %s: %w", id, err)
	}
	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size for %s: got %d", id, len(data))
	}
	return ed25519.PrivateKey(data), nil
}

func (f *FileKeyStore) LoadPublicKey(id string) (ed25519.PublicKey, error) {
	priv, err := f.LoadKey(id)
	if err != nil {
		return nil, err
	}
	return priv.Public().(ed25519.PublicKey), nil
}

func (f *FileKeyStore) DeleteKey(id string) error {
	return os.Remove(f.keyPath(id))
}

func (f *FileKeyStore) HasKey(id string) bool {
	_, err := os.Stat(f.keyPath(id))
	return err == nil
}
