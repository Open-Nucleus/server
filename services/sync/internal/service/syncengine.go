package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge"
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

	state         SyncState
	currentSyncID string
	currentPeer   string
	lastSync      time.Time
	nodeID        string
	siteID        string
	startedAt     time.Time
}

// NewSyncEngine creates a new sync engine.
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
		state:       StateIdle,
		nodeID:      nodeID,
		siteID:      siteID,
		startedAt:   time.Now(),
	}
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

// ExportBundle creates an encrypted bundle of resources.
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

	// Encrypt with AES-256-GCM
	encrypted, err := encryptAESGCM(bundle)
	if err != nil {
		return nil, 0, fmt.Errorf("encrypt bundle: %w", err)
	}

	return encrypted, len(resources), nil
}

// ImportBundle decrypts and imports a bundle.
func (se *SyncEngine) ImportBundle(bundleData []byte) (int, int, []string, error) {
	// Decrypt
	decrypted, err := decryptAESGCM(bundleData)
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

// encryptAESGCM encrypts data with a random AES-256-GCM key.
// The key is prepended to the ciphertext (32 bytes key + 12 bytes nonce + ciphertext + tag).
func encryptAESGCM(data []byte) ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

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

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	// Prepend key so recipient can decrypt (in production, key exchange happens via handshake)
	result := make([]byte, 32+len(ciphertext))
	copy(result, key)
	copy(result[32:], ciphertext)
	return result, nil
}

// decryptAESGCM decrypts data encrypted by encryptAESGCM.
func decryptAESGCM(data []byte) ([]byte, error) {
	if len(data) < 44 { // 32 key + 12 nonce minimum
		return nil, fmt.Errorf("ciphertext too short")
	}

	key := data[:32]
	ciphertext := data[32:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short for nonce")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
