package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
)

// CredentialStore manages Git-backed Verifiable Credentials.
type CredentialStore struct {
	git gitstore.Store
}

// NewCredentialStore creates a new CredentialStore.
func NewCredentialStore(git gitstore.Store) *CredentialStore {
	return &CredentialStore{git: git}
}

// Save writes a VC to Git.
func (s *CredentialStore) Save(vc *openanchor.VerifiableCredential) (string, error) {
	data, err := json.MarshalIndent(vc, "", "  ")
	if err != nil {
		return "", err
	}

	// Use a safe filename from the VC ID.
	filename := sanitizeID(vc.ID)
	path := fmt.Sprintf(".nucleus/credentials/%s.json", filename)

	commit, err := s.git.WriteAndCommit(path, data, gitstore.CommitMessage{
		ResourceType: "VerifiableCredential",
		Operation:    "CREATE",
		ResourceID:   vc.ID,
		NodeID:       "anchor-service",
		Author:       "anchor-service",
		SiteID:       "local",
	})
	if err != nil {
		return "", fmt.Errorf("write credential: %w", err)
	}
	return commit, nil
}

// List returns all credentials, optionally filtered by type, with pagination.
func (s *CredentialStore) List(credType string, page, perPage int) ([]openanchor.VerifiableCredential, int, error) {
	var all []openanchor.VerifiableCredential
	err := s.git.TreeWalk(func(path string, data []byte) error {
		if !strings.HasPrefix(path, ".nucleus/credentials/") || !strings.HasSuffix(path, ".json") {
			return nil
		}
		var vc openanchor.VerifiableCredential
		if err := json.Unmarshal(data, &vc); err != nil {
			return nil
		}
		if credType != "" {
			found := false
			for _, t := range vc.Type {
				if t == credType {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}
		all = append(all, vc)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].IssuanceDate > all[j].IssuanceDate
	})

	total := len(all)
	start := (page - 1) * perPage
	if start >= total {
		return nil, total, nil
	}
	end := start + perPage
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func sanitizeID(id string) string {
	// Replace colons and slashes with dashes for safe filenames.
	r := strings.NewReplacer(":", "-", "/", "-")
	return r.Replace(id)
}
