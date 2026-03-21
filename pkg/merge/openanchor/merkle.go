package openanchor

import (
	"crypto/sha256"
	"fmt"
	"sort"

	anchor "github.com/Open-Nucleus/open-anchor/go"
)

// SHA256Merkle implements MerkleTree using the external open-anchor library.
type SHA256Merkle struct{}

// NewMerkleTree returns a new SHA-256 Merkle tree implementation.
func NewMerkleTree() MerkleTree {
	return &SHA256Merkle{}
}

// ComputeRoot computes the SHA-256 Merkle root from sorted file entries.
func (m *SHA256Merkle) ComputeRoot(entries []FileEntry) ([]byte, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("no entries to hash")
	}

	// Sort entries by path for deterministic ordering.
	sorted := make([]FileEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Path < sorted[j].Path
	})

	// Convert to external MerkleLeaf format.
	// The external library hashes leaves differently (prefix-based), so we
	// pre-compute leaf hashes matching our format: H(path || fileHash).
	leaves := make([]anchor.MerkleLeaf, len(sorted))
	for i, entry := range sorted {
		h := sha256.New()
		h.Write([]byte(entry.Path))
		h.Write(entry.Hash)
		leaves[i] = anchor.MerkleLeaf{
			Path: entry.Path,
			Hash: h.Sum(nil),
		}
	}

	tree, err := anchor.NewMerkleTree(leaves)
	if err != nil {
		return nil, err
	}
	return tree.GetRoot(), nil
}
