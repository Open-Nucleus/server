package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
)

// AnchorRecordData is stored as JSON in Git.
type AnchorRecordData struct {
	AnchorID   string `json:"anchor_id"`
	MerkleRoot string `json:"merkle_root"`
	GitHead    string `json:"git_head"`
	State      string `json:"state"`
	Timestamp  string `json:"timestamp"`
	Backend    string `json:"backend"`
	TxID       string `json:"tx_id"`
	NodeDID    string `json:"node_did"`
}

// AnchorStore manages Git-backed anchor records.
type AnchorStore struct {
	git gitstore.Store
}

// NewAnchorStore creates a new AnchorStore.
func NewAnchorStore(git gitstore.Store) *AnchorStore {
	return &AnchorStore{git: git}
}

// Save writes an anchor record to Git.
func (s *AnchorStore) Save(rec *AnchorRecordData) (string, error) {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf(".nucleus/anchors/%s.json", rec.AnchorID)
	commit, err := s.git.WriteAndCommit(path, data, gitstore.CommitMessage{
		ResourceType: "AnchorRecord",
		Operation:    "CREATE",
		ResourceID:   rec.AnchorID,
		NodeID:       "anchor-service",
		Author:       "anchor-service",
		SiteID:       "local",
	})
	if err != nil {
		return "", fmt.Errorf("write anchor record: %w", err)
	}
	return commit, nil
}

// Get reads an anchor record from Git.
func (s *AnchorStore) Get(anchorID string) (*AnchorRecordData, error) {
	path := fmt.Sprintf(".nucleus/anchors/%s.json", anchorID)
	data, err := s.git.Read(path)
	if err != nil {
		return nil, fmt.Errorf("anchor %s not found: %w", anchorID, err)
	}

	var rec AnchorRecordData
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("parse anchor record: %w", err)
	}
	return &rec, nil
}

// FindByGitHead searches for an anchor record matching the given git head.
func (s *AnchorStore) FindByGitHead(gitHead string) (*AnchorRecordData, error) {
	var found *AnchorRecordData
	err := s.git.TreeWalk(func(path string, data []byte) error {
		if !strings.HasPrefix(path, ".nucleus/anchors/") || !strings.HasSuffix(path, ".json") {
			return nil
		}
		var rec AnchorRecordData
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil
		}
		if rec.GitHead == gitHead {
			found = &rec
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("anchor for commit %s not found", gitHead)
	}
	return found, nil
}

// List returns all anchor records with pagination.
func (s *AnchorStore) List(page, perPage int) ([]AnchorRecordData, int, error) {
	var all []AnchorRecordData
	err := s.git.TreeWalk(func(path string, data []byte) error {
		if !strings.HasPrefix(path, ".nucleus/anchors/") || !strings.HasSuffix(path, ".json") {
			return nil
		}
		var rec AnchorRecordData
		if err := json.Unmarshal(data, &rec); err != nil {
			return nil
		}
		all = append(all, rec)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Sort by timestamp descending.
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp > all[j].Timestamp
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
