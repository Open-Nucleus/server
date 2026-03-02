package openanchor

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// SHA256Merkle implements MerkleTree using SHA-256.
type SHA256Merkle struct{}

// NewMerkleTree returns a new SHA-256 Merkle tree implementation.
func NewMerkleTree() MerkleTree {
	return &SHA256Merkle{}
}

// ComputeRoot computes the SHA-256 Merkle root from sorted file entries.
// Each leaf is H(path || fileHash). The tree is built bottom-up, duplicating
// the last node when the level has an odd count.
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

	// Build leaf nodes: H(path || fileHash)
	leaves := make([][]byte, len(sorted))
	for i, entry := range sorted {
		h := sha256.New()
		h.Write([]byte(entry.Path))
		h.Write(entry.Hash)
		leaves[i] = h.Sum(nil)
	}

	// Build tree bottom-up.
	level := leaves
	for len(level) > 1 {
		var next [][]byte
		for i := 0; i < len(level); i += 2 {
			left := level[i]
			var right []byte
			if i+1 < len(level) {
				right = level[i+1]
			} else {
				right = left // duplicate last node
			}
			h := sha256.New()
			h.Write(left)
			h.Write(right)
			next = append(next, h.Sum(nil))
		}
		level = next
	}

	return level[0], nil
}
