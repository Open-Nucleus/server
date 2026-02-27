package auth

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// NonceEntry stores a nonce with its expiration time.
type NonceEntry struct {
	Nonce     []byte
	ExpiresAt time.Time
}

// NonceStore manages challenge nonces with TTL-based expiry.
type NonceStore struct {
	mu    sync.Mutex
	store map[string]*NonceEntry // keyed by device ID
	ttl   time.Duration
}

// NewNonceStore creates a new NonceStore with the given TTL.
func NewNonceStore(ttl time.Duration) *NonceStore {
	return &NonceStore{
		store: make(map[string]*NonceEntry),
		ttl:   ttl,
	}
}

// Generate creates a 32-byte cryptographic nonce for the given device ID.
// Overwrites any existing nonce for that device.
func (ns *NonceStore) Generate(deviceID string) ([]byte, time.Time, error) {
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, time.Time{}, fmt.Errorf("generate nonce: %w", err)
	}

	expiresAt := time.Now().Add(ns.ttl)

	ns.mu.Lock()
	ns.store[deviceID] = &NonceEntry{
		Nonce:     nonce,
		ExpiresAt: expiresAt,
	}
	ns.mu.Unlock()

	return nonce, expiresAt, nil
}

// Consume validates and removes a nonce for the given device ID.
// Returns true if the nonce was valid and not expired, false otherwise.
func (ns *NonceStore) Consume(deviceID string, nonce []byte) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	entry, ok := ns.store[deviceID]
	if !ok {
		return false
	}

	// Remove regardless of validity (one-shot)
	delete(ns.store, deviceID)

	// Check expiry
	if time.Now().After(entry.ExpiresAt) {
		return false
	}

	// Constant-time comparison to prevent timing attacks
	if len(nonce) != len(entry.Nonce) {
		return false
	}
	var diff byte
	for i := range nonce {
		diff |= nonce[i] ^ entry.Nonce[i]
	}
	return diff == 0
}

// Cleanup removes expired nonces.
func (ns *NonceStore) Cleanup() int {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	now := time.Now()
	removed := 0
	for id, entry := range ns.store {
		if now.After(entry.ExpiresAt) {
			delete(ns.store, id)
			removed++
		}
	}
	return removed
}
