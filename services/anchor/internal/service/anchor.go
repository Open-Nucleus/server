package service

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/services/anchor/internal/store"
)

// AnchorService contains the core business logic for anchoring.
type AnchorService struct {
	gitStore       gitstore.Store
	anchorEngine   openanchor.AnchorEngine
	identityEngine openanchor.IdentityEngine
	merkle         openanchor.MerkleTree
	queue          *store.AnchorQueue
	anchorStore    *store.AnchorStore
	credStore      *store.CredentialStore
	didStore       *store.DIDStore
	nodePrivKey    ed25519.PrivateKey
	nodeDID        string
	lastAnchorRoot string
}

// New creates a new AnchorService.
func New(
	gitStore gitstore.Store,
	anchorEngine openanchor.AnchorEngine,
	identityEngine openanchor.IdentityEngine,
	queue *store.AnchorQueue,
	anchorStore *store.AnchorStore,
	credStore *store.CredentialStore,
	didStore *store.DIDStore,
	nodePrivKey ed25519.PrivateKey,
) *AnchorService {
	return &AnchorService{
		gitStore:       gitStore,
		anchorEngine:   anchorEngine,
		identityEngine: identityEngine,
		merkle:         openanchor.NewMerkleTree(),
		queue:          queue,
		anchorStore:    anchorStore,
		credStore:      credStore,
		didStore:       didStore,
		nodePrivKey:    nodePrivKey,
	}
}

// Bootstrap generates and stores the node DID from the Ed25519 private key.
func (s *AnchorService) Bootstrap() error {
	pub := s.nodePrivKey.Public().(ed25519.PublicKey)
	did, doc, err := openanchor.DIDKeyFromEd25519(pub)
	if err != nil {
		return fmt.Errorf("generate node DID: %w", err)
	}

	if _, err := s.didStore.SaveNodeDID(doc); err != nil {
		return fmt.Errorf("save node DID: %w", err)
	}

	s.nodeDID = did
	return nil
}

// GetStatus returns current anchor status.
func (s *AnchorService) GetStatus() (*StatusResult, error) {
	queueDepth, err := s.queue.CountPending()
	if err != nil {
		return nil, err
	}

	// Get latest anchor.
	records, _, err := s.anchorStore.List(1, 1)
	if err != nil {
		return nil, err
	}

	result := &StatusResult{
		State:      "idle",
		NodeDID:    s.nodeDID,
		QueueDepth: queueDepth,
		Backend:    s.anchorEngine.Name(),
	}

	if len(records) > 0 {
		result.LastAnchorID = records[0].AnchorID
		result.LastAnchorTime = records[0].Timestamp
		result.MerkleRoot = records[0].MerkleRoot
	}

	return result, nil
}

// TriggerAnchor computes the Merkle tree and creates an anchor record.
func (s *AnchorService) TriggerAnchor(manual bool) (*TriggerResult, error) {
	// Walk Git tree to build file entries (exclude internal .nucleus/ metadata).
	var entries []openanchor.FileEntry
	err := s.gitStore.TreeWalk(func(path string, data []byte) error {
		if strings.HasPrefix(path, ".nucleus/") {
			return nil
		}
		hash := sha256.Sum256(data)
		entries = append(entries, openanchor.FileEntry{Path: path, Hash: hash[:]})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("tree walk: %w", err)
	}

	if len(entries) == 0 {
		return &TriggerResult{
			Skipped: true,
			Message: "no files in repository",
		}, nil
	}

	root, err := s.merkle.ComputeRoot(entries)
	if err != nil {
		return nil, fmt.Errorf("compute merkle root: %w", err)
	}

	rootHex := hex.EncodeToString(root)

	// Skip if root unchanged (unless manual).
	if !manual && rootHex == s.lastAnchorRoot {
		return &TriggerResult{
			Skipped: true,
			Message: "merkle root unchanged",
		}, nil
	}

	gitHead, err := s.gitStore.Head()
	if err != nil {
		return nil, fmt.Errorf("get HEAD: %w", err)
	}

	anchorID := uuid.New().String()
	now := time.Now().UTC().Format(time.RFC3339)

	// Attempt to anchor via engine.
	state := "queued"
	var txID string

	_, anchorErr := s.anchorEngine.Anchor(root, openanchor.AnchorMetadata{
		GitHead:   gitHead,
		NodeDID:   s.nodeDID,
		Timestamp: time.Now(),
	})
	if anchorErr != nil {
		// Backend not available — enqueue.
		if err := s.queue.Enqueue(anchorID, rootHex, gitHead, s.nodeDID); err != nil {
			return nil, fmt.Errorf("enqueue: %w", err)
		}
	} else {
		state = "confirmed"
	}

	// Store anchor record in Git.
	rec := &store.AnchorRecordData{
		AnchorID:   anchorID,
		MerkleRoot: rootHex,
		GitHead:    gitHead,
		State:      state,
		Timestamp:  now,
		Backend:    s.anchorEngine.Name(),
		TxID:       txID,
		NodeDID:    s.nodeDID,
	}
	if _, err := s.anchorStore.Save(rec); err != nil {
		return nil, fmt.Errorf("save anchor record: %w", err)
	}

	s.lastAnchorRoot = rootHex

	return &TriggerResult{
		AnchorID:   anchorID,
		State:      state,
		MerkleRoot: rootHex,
		GitHead:    gitHead,
	}, nil
}

// Verify checks if a commit hash has been anchored.
func (s *AnchorService) Verify(commitHash string) (*VerifyResult, error) {
	rec, err := s.anchorStore.FindByGitHead(commitHash)
	if err != nil {
		return &VerifyResult{
			Verified:   false,
			CommitHash: commitHash,
		}, nil
	}

	return &VerifyResult{
		Verified:   true,
		AnchorID:   rec.AnchorID,
		MerkleRoot: rec.MerkleRoot,
		AnchoredAt: rec.Timestamp,
		CommitHash: commitHash,
		State:      rec.State,
	}, nil
}

// GetHistory returns paginated anchor records.
func (s *AnchorService) GetHistory(page, perPage int) ([]store.AnchorRecordData, int, error) {
	return s.anchorStore.List(page, perPage)
}

// GetNodeDID returns the node's DID document.
func (s *AnchorService) GetNodeDID() (*openanchor.DIDDocument, error) {
	return s.didStore.GetNodeDID()
}

// GetDeviceDID returns a device's DID document.
func (s *AnchorService) GetDeviceDID(deviceID string) (*openanchor.DIDDocument, error) {
	return s.didStore.GetDeviceDID(deviceID)
}

// BootstrapDeviceDID creates and stores a DID for a device.
func (s *AnchorService) BootstrapDeviceDID(deviceID string, pub ed25519.PublicKey) (*openanchor.DIDDocument, error) {
	doc, err := s.identityEngine.GenerateDID(pub)
	if err != nil {
		return nil, err
	}
	if _, err := s.didStore.SaveDeviceDID(deviceID, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// ResolveDID resolves a DID string to a DID document.
func (s *AnchorService) ResolveDID(did string) (*openanchor.DIDDocument, error) {
	return s.identityEngine.ResolveDID(did)
}

// IssueDataIntegrityCredential issues a VC for an anchor record.
func (s *AnchorService) IssueDataIntegrityCredential(anchorID string, extraTypes []string, additionalClaims map[string]string) (*openanchor.VerifiableCredential, error) {
	rec, err := s.anchorStore.Get(anchorID)
	if err != nil {
		return nil, fmt.Errorf("anchor record not found: %w", err)
	}

	subject := map[string]any{
		"anchorId":   rec.AnchorID,
		"merkleRoot": rec.MerkleRoot,
		"gitHead":    rec.GitHead,
		"state":      rec.State,
		"timestamp":  rec.Timestamp,
		"backend":    rec.Backend,
	}
	for k, v := range additionalClaims {
		subject[k] = v
	}

	types := []string{"DataIntegrityCredential"}
	types = append(types, extraTypes...)

	claims := openanchor.CredentialClaims{
		ID:      "urn:uuid:" + uuid.New().String(),
		Types:   types,
		Subject: subject,
	}

	vc, err := s.identityEngine.IssueCredential(claims, s.nodeDID, s.nodePrivKey)
	if err != nil {
		return nil, fmt.Errorf("issue credential: %w", err)
	}

	if _, err := s.credStore.Save(vc); err != nil {
		return nil, fmt.Errorf("save credential: %w", err)
	}

	return vc, nil
}

// VerifyCredential verifies a Verifiable Credential.
func (s *AnchorService) VerifyCredential(vc *openanchor.VerifiableCredential) (*openanchor.VerificationResult, error) {
	return s.identityEngine.VerifyCredential(vc)
}

// ListCredentials returns paginated credentials, optionally filtered by type.
func (s *AnchorService) ListCredentials(credType string, page, perPage int) ([]openanchor.VerifiableCredential, int, error) {
	return s.credStore.List(credType, page, perPage)
}

// ListBackends returns information about available anchor backends.
func (s *AnchorService) ListBackends() []BackendInfo {
	return []BackendInfo{
		{
			Name:        s.anchorEngine.Name(),
			Available:   s.anchorEngine.Available(),
			Description: "Stub backend — anchors are queued but never submitted",
		},
	}
}

// GetBackendStatus returns detailed status for a named backend.
func (s *AnchorService) GetBackendStatus(name string) (*BackendStatusResult, error) {
	if name != s.anchorEngine.Name() {
		return nil, fmt.Errorf("backend %q not found", name)
	}

	_, total, err := s.anchorStore.List(1, 1)
	if err != nil {
		return nil, err
	}

	records, _, err := s.anchorStore.List(1, 1)
	if err != nil {
		return nil, err
	}

	var lastTime string
	if len(records) > 0 {
		lastTime = records[0].Timestamp
	}

	return &BackendStatusResult{
		Name:           name,
		Available:      s.anchorEngine.Available(),
		Description:    "Stub backend — anchors are queued but never submitted",
		AnchoredCount:  total,
		LastAnchorTime: lastTime,
	}, nil
}

// GetQueueStatus returns the anchor queue status.
func (s *AnchorService) GetQueueStatus() (*QueueStatusResult, error) {
	pending, err := s.queue.CountPending()
	if err != nil {
		return nil, err
	}

	total, err := s.queue.CountTotal()
	if err != nil {
		return nil, err
	}

	entries, err := s.queue.ListPending()
	if err != nil {
		return nil, err
	}

	return &QueueStatusResult{
		Pending:        pending,
		TotalProcessed: total,
		Entries:        entries,
	}, nil
}

// AnchorCount returns the number of anchor records.
func (s *AnchorService) AnchorCount() int {
	_, total, _ := s.anchorStore.List(1, 0)
	return total
}

// NodeDIDString returns the node DID string.
func (s *AnchorService) NodeDIDString() string {
	return s.nodeDID
}

// BackendName returns the anchor engine name.
func (s *AnchorService) BackendName() string {
	return s.anchorEngine.Name()
}

// QueueDepth returns the number of pending queue entries.
func (s *AnchorService) QueueDepth() int {
	n, _ := s.queue.CountPending()
	return n
}

// NodePublicKey returns the node's Ed25519 public key.
func (s *AnchorService) NodePublicKey() ed25519.PublicKey {
	return s.nodePrivKey.Public().(ed25519.PublicKey)
}

// --- Result types ---

type StatusResult struct {
	State          string
	LastAnchorID   string
	LastAnchorTime string
	MerkleRoot     string
	NodeDID        string
	QueueDepth     int
	Backend        string
}

type TriggerResult struct {
	AnchorID   string
	State      string
	MerkleRoot string
	GitHead    string
	Skipped    bool
	Message    string
}

type VerifyResult struct {
	Verified   bool
	AnchorID   string
	MerkleRoot string
	AnchoredAt string
	CommitHash string
	State      string
}

type BackendInfo struct {
	Name        string
	Available   bool
	Description string
}

type BackendStatusResult struct {
	Name           string
	Available      bool
	Description    string
	AnchoredCount  int
	LastAnchorTime string
}

type QueueStatusResult struct {
	Pending        int
	TotalProcessed int
	Entries        []store.QueueEntry
}
