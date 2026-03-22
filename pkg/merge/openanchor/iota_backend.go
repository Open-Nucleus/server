package openanchor

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	anchor "github.com/Open-Nucleus/open-anchor/go"
	iotabackend "github.com/Open-Nucleus/open-anchor/go/backends/iota"
)

// IotaBackend wraps the external open-anchor IOTA Rebased backend and
// adapts it to the nucleus AnchorEngine interface.
type IotaBackend struct {
	backend *iotabackend.AnchorBackend
}

// NewIotaBackend creates a new IOTA Rebased anchor backend.
func NewIotaBackend(config iotabackend.Config, signingKey ed25519.PrivateKey) (*IotaBackend, error) {
	b, err := iotabackend.NewAnchorBackend(config, signingKey)
	if err != nil {
		return nil, err
	}
	return &IotaBackend{backend: b}, nil
}

// NewIotaBackendFromConfig creates an IOTA backend from individual config strings.
func NewIotaBackendFromConfig(network, rpcURL, anchorPackageID, identityPackageID string, signingKey ed25519.PrivateKey) (*IotaBackend, error) {
	if rpcURL == "" {
		switch network {
		case "mainnet":
			rpcURL = iotabackend.MainnetRPCURL
		case "devnet":
			rpcURL = iotabackend.DevnetRPCURL
		default:
			rpcURL = iotabackend.TestnetRPCURL
		}
	}
	if anchorPackageID == "" {
		return nil, fmt.Errorf("iota: anchor_package_id is required")
	}

	cfg := iotabackend.Config{
		RPCURL:            rpcURL,
		NetworkID:         network,
		AnchorPackageID:   anchorPackageID,
		IdentityPackageID: identityPackageID,
	}

	return NewIotaBackend(cfg, signingKey)
}

// Anchor submits a Merkle root to the IOTA Rebased network via Move contract.
func (b *IotaBackend) Anchor(root []byte, metadata AnchorMetadata) (*AnchorReceipt, error) {
	proof := anchor.AnchorProof{
		MerkleRoot:  root,
		Description: "nucleus git anchor: " + metadata.GitHead,
		SourceID:    metadata.NodeDID,
		ComputedAt:  metadata.Timestamp,
	}

	receipt, err := b.backend.Anchor(context.Background(), proof)
	if err != nil {
		return nil, err
	}

	return &AnchorReceipt{
		BackendName: "iota",
		TxID:        receipt.TransactionID,
		Timestamp:   receipt.AnchoredAt,
	}, nil
}

// Available checks if the IOTA network is reachable.
func (b *IotaBackend) Available() bool {
	status, err := b.backend.Status(context.Background())
	if err != nil {
		return false
	}
	return status.Connected
}

// Name returns "iota".
func (b *IotaBackend) Name() string { return "iota" }

// VerifyAnchor checks an anchor receipt against the IOTA network.
func (b *IotaBackend) VerifyAnchor(root []byte, txID string, blockRef string) (bool, error) {
	receipt := anchor.AnchorReceipt{
		Backend:       "iota",
		MerkleRoot:    root,
		TransactionID: txID,
		AnchoredAt:    time.Now(),
		BlockRef:      blockRef,
	}

	result, err := b.backend.Verify(context.Background(), receipt)
	if err != nil {
		return false, err
	}
	return result.Valid, nil
}
