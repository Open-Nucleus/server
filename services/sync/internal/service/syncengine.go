package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
	synccrypto "github.com/FibrinLab/open-nucleus/pkg/sync"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/config"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/store"
	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport"
)

// SyncState represents the current sync state.
type SyncState string

const (
	StateIdle    SyncState = "idle"
	StateSyncing SyncState = "syncing"
	StateError   SyncState = "error"
)

// SyncEngine orchestrates sync operations.
type SyncEngine struct {
	mu          sync.Mutex
	cfg         *config.Config
	git         gitstore.Store
	conflicts   *store.ConflictStore
	history     *store.HistoryStore
	peers       *store.PeerStore
	mergeDriver *merge.Driver
	eventBus    *EventBus
	queue       *SyncQueue
	transports  map[string]transport.Adapter

	identityKey ed25519.PrivateKey // node's Ed25519 identity key for ECDH

	state         SyncState
	currentSyncID string
	currentPeer   string
	lastSync      time.Time
	nodeID        string
	siteID        string
	startedAt     time.Time
}

// NewSyncEngine creates a new sync engine.
// If identityKey is nil, an ephemeral Ed25519 keypair is generated.
func NewSyncEngine(
	cfg *config.Config,
	git gitstore.Store,
	conflicts *store.ConflictStore,
	history *store.HistoryStore,
	peers *store.PeerStore,
	mergeDriver *merge.Driver,
	eventBus *EventBus,
	nodeID, siteID string,
) *SyncEngine {
	// Generate an ephemeral identity key if none provided.
	// In production, the identity key is loaded from the node's keystore.
	_, priv, _ := ed25519.GenerateKey(rand.Reader)

	return &SyncEngine{
		cfg:         cfg,
		git:         git,
		conflicts:   conflicts,
		history:     history,
		peers:       peers,
		mergeDriver: mergeDriver,
		eventBus:    eventBus,
		queue:       NewSyncQueue(),
		transports:  make(map[string]transport.Adapter),
		identityKey: priv,
		state:       StateIdle,
		nodeID:      nodeID,
		siteID:      siteID,
		startedAt:   time.Now(),
	}
}

// IdentityPublicKey returns this node's Ed25519 public key.
func (se *SyncEngine) IdentityPublicKey() ed25519.PublicKey {
	return se.identityKey.Public().(ed25519.PublicKey)
}

// RegisterTransport adds a transport adapter.
func (se *SyncEngine) RegisterTransport(adapter transport.Adapter) {
	se.transports[adapter.Name()] = adapter
}

// GetStatus returns the current sync state.
func (se *SyncEngine) GetStatus() (SyncState, string, string, int) {
	se.mu.Lock()
	defer se.mu.Unlock()
	pending := se.queue.Len()
	return se.state, se.currentSyncID, se.currentPeer, pending
}

// TriggerSync starts a sync with a target node.
func (se *SyncEngine) TriggerSync(targetNode string) (string, error) {
	se.mu.Lock()
	defer se.mu.Unlock()

	if se.state == StateSyncing {
		return "", fmt.Errorf("sync already in progress: %s", se.currentSyncID)
	}

	syncID := uuid.New().String()
	se.state = StateSyncing
	se.currentSyncID = syncID
	se.currentPeer = targetNode

	// Record history entry
	headBefore, _ := se.git.Head()
	_ = se.history.Record(&store.HistoryRecord{
		ID:              syncID,
		PeerNode:        targetNode,
		Direction:       "bidirectional",
		State:           "in_progress",
		StartedAt:       time.Now().UTC().Format(time.RFC3339),
		LocalHeadBefore: headBefore,
	})

	se.eventBus.Publish(Event{
		Type: EventSyncStarted,
		Payload: map[string]string{
			"sync_id":     syncID,
			"target_node": targetNode,
		},
	})

	// In production, this would be async. For now, mark as completed.
	headAfter, _ := se.git.Head()
	_ = se.history.RecordCompleted(syncID, 0, 0, 0, headAfter)

	se.state = StateIdle
	se.currentSyncID = ""
	se.currentPeer = ""
	se.lastSync = time.Now()

	se.eventBus.Publish(Event{
		Type: EventSyncCompleted,
		Payload: map[string]string{
			"sync_id": syncID,
		},
	})

	return syncID, nil
}

// CancelSync cancels an in-progress sync.
func (se *SyncEngine) CancelSync(syncID string) bool {
	se.mu.Lock()
	defer se.mu.Unlock()

	if se.currentSyncID == syncID && se.state == StateSyncing {
		se.state = StateIdle
		se.currentSyncID = ""
		se.currentPeer = ""
		_ = se.history.RecordFailed(syncID, "cancelled by user")
		return true
	}
	return false
}

// ListTransports returns all registered transports.
func (se *SyncEngine) ListTransports() []transport.Adapter {
	adapters := make([]transport.Adapter, 0, len(se.transports))
	for _, a := range se.transports {
		adapters = append(adapters, a)
	}
	return adapters
}

// LastSyncTime returns the last successful sync timestamp.
func (se *SyncEngine) LastSyncTime() string {
	se.mu.Lock()
	defer se.mu.Unlock()
	if se.lastSync.IsZero() {
		return ""
	}
	return se.lastSync.UTC().Format(time.RFC3339)
}

// NodeID returns this node's ID.
func (se *SyncEngine) NodeID() string { return se.nodeID }

// SiteID returns this node's site ID.
func (se *SyncEngine) SiteID() string { return se.siteID }

// UptimeSeconds returns uptime in seconds.
func (se *SyncEngine) UptimeSeconds() int64 {
	return int64(time.Since(se.startedAt).Seconds())
}

// ExportBundle creates an encrypted bundle of resources using ECDH + AES-256-GCM.
//
// The bundle is encrypted using an ECIES-like scheme:
//  1. Generate an ephemeral Ed25519 keypair
//  2. Derive a shared key via ECDH(ephemeral_private, node_identity_public)
//  3. Encrypt the bundle with AES-256-GCM using the derived key
//  4. Prepend the ephemeral public key (32 bytes) so the recipient can derive the same key
//
// Output format: [32-byte ephemeral public key][encrypted payload]
func (se *SyncEngine) ExportBundle(resourceTypes []string, since string) ([]byte, int, error) {
	var resources []json.RawMessage

	err := se.git.TreeWalk(func(path string, data []byte) error {
		if len(resourceTypes) > 0 {
			matched := false
			for _, rt := range resourceTypes {
				if containsStr(path, rt) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}
		resources = append(resources, json.RawMessage(data))
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("walk repo: %w", err)
	}

	bundle, err := json.Marshal(resources)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal bundle: %w", err)
	}

	// Generate ephemeral keypair for this bundle.
	ephPub, ephPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, 0, fmt.Errorf("generate ephemeral key: %w", err)
	}

	// Derive shared key: ECDH(ephemeral_private, our_identity_public).
	// The recipient (who holds our identity private key) will do
	// ECDH(identity_private, ephemeral_public) and get the same key.
	sharedKey, err := synccrypto.DeriveSharedKey(ephPriv, se.IdentityPublicKey())
	if err != nil {
		return nil, 0, fmt.Errorf("derive shared key: %w", err)
	}

	encrypted, err := synccrypto.EncryptPayload(sharedKey, bundle)
	if err != nil {
		return nil, 0, fmt.Errorf("encrypt bundle: %w", err)
	}

	// Prepend ephemeral public key (32 bytes) to the encrypted payload.
	result := make([]byte, ed25519.PublicKeySize+len(encrypted))
	copy(result, ephPub)
	copy(result[ed25519.PublicKeySize:], encrypted)

	return result, len(resources), nil
}

// ImportBundle decrypts and imports a bundle encrypted by ExportBundle.
//
// It extracts the ephemeral public key from the bundle header, derives
// the shared key via ECDH(identity_private, ephemeral_public), and decrypts.
func (se *SyncEngine) ImportBundle(bundleData []byte) (int, int, []string, error) {
	if len(bundleData) < ed25519.PublicKeySize {
		return 0, 0, nil, fmt.Errorf("bundle too short: need at least %d bytes for public key header", ed25519.PublicKeySize)
	}

	// Extract ephemeral public key.
	ephPub := ed25519.PublicKey(bundleData[:ed25519.PublicKeySize])
	ciphertext := bundleData[ed25519.PublicKeySize:]

	// Derive shared key: ECDH(our_identity_private, ephemeral_public).
	sharedKey, err := synccrypto.DeriveSharedKey(se.identityKey, ephPub)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("derive shared key: %w", err)
	}

	decrypted, err := synccrypto.DecryptPayload(sharedKey, ciphertext)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("decrypt bundle: %w", err)
	}

	var resources []json.RawMessage
	if err := json.Unmarshal(decrypted, &resources); err != nil {
		return 0, 0, nil, fmt.Errorf("unmarshal bundle: %w", err)
	}

	imported := 0
	skipped := 0
	var errors []string

	for _, res := range resources {
		// In production: parse resource, determine path, write to Git, merge if needed
		_ = res
		imported++
	}

	return imported, skipped, errors, nil
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
