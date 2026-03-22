package openanchor

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	anchor "github.com/Open-Nucleus/open-anchor/go"
	hederabackend "github.com/Open-Nucleus/open-anchor/go/backends/hedera"
)

// HederaBackend wraps the external open-anchor Hedera HCS backend and
// adapts it to the nucleus AnchorEngine interface.
type HederaBackend struct {
	backend *hederabackend.AnchorBackend
}

// NewHederaBackend creates a new Hedera HCS anchor backend.
func NewHederaBackend(config hederabackend.Config, signingKey ed25519.PrivateKey) (*HederaBackend, error) {
	b, err := hederabackend.NewAnchorBackend(config, signingKey)
	if err != nil {
		return nil, err
	}
	return &HederaBackend{backend: b}, nil
}

// NewHederaBackendFromConfig creates a Hedera backend from individual config strings.
// This avoids requiring callers to import the hedera backend package directly.
func NewHederaBackendFromConfig(network, operatorID, operatorKey, topicID, didTopicID, mirrorURL string, signingKey ed25519.PrivateKey) (*HederaBackend, error) {
	if operatorKey == "" {
		return nil, fmt.Errorf("hedera: operator key is required (set anchor.operator_key or NUCLEUS_HEDERA_KEY)")
	}

	cfg := hederabackend.Config{
		Network:     network,
		OperatorID:  operatorID,
		OperatorKey: operatorKey,
		TopicID:     topicID,
		DIDTopicID:  didTopicID,
		MirrorURL:   mirrorURL,
	}

	return NewHederaBackend(cfg, signingKey)
}

// Anchor submits a Merkle root to Hedera Consensus Service.
func (h *HederaBackend) Anchor(root []byte, metadata AnchorMetadata) (*AnchorReceipt, error) {
	proof := anchor.AnchorProof{
		MerkleRoot:  root,
		Description: "nucleus git anchor: " + metadata.GitHead,
		SourceID:    metadata.NodeDID,
		ComputedAt:  metadata.Timestamp,
	}

	receipt, err := h.backend.Anchor(context.Background(), proof)
	if err != nil {
		return nil, err
	}

	return &AnchorReceipt{
		BackendName: "hedera",
		TxID:        receipt.TransactionID,
		Timestamp:   receipt.AnchoredAt,
	}, nil
}

// Available checks if the Hedera network is reachable.
func (h *HederaBackend) Available() bool {
	status, err := h.backend.Status(context.Background())
	if err != nil {
		return false
	}
	return status.Connected
}

// Name returns "hedera".
func (h *HederaBackend) Name() string { return "hedera" }

// VerifyAnchor checks an anchor receipt against the Hedera Mirror Node.
func (h *HederaBackend) VerifyAnchor(root []byte, txID string, blockRef string) (bool, error) {
	receipt := anchor.AnchorReceipt{
		Backend:       "hedera",
		MerkleRoot:    root,
		TransactionID: txID,
		AnchoredAt:    time.Now(),
		BlockRef:      blockRef,
	}

	result, err := h.backend.Verify(context.Background(), receipt)
	if err != nil {
		return false, err
	}
	return result.Valid, nil
}

// Close releases Hedera client resources.
func (h *HederaBackend) Close() error {
	return h.backend.Close()
}
