package store

import (
	"encoding/json"
	"fmt"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
)

// DIDStore manages Git-backed DID documents.
type DIDStore struct {
	git gitstore.Store
}

// NewDIDStore creates a new DIDStore.
func NewDIDStore(git gitstore.Store) *DIDStore {
	return &DIDStore{git: git}
}

// SaveNodeDID writes the node DID document to Git.
func (s *DIDStore) SaveNodeDID(doc *openanchor.DIDDocument) (string, error) {
	return s.save("node-did", doc)
}

// SaveDeviceDID writes a device DID document to Git.
func (s *DIDStore) SaveDeviceDID(deviceID string, doc *openanchor.DIDDocument) (string, error) {
	return s.save("device-"+deviceID, doc)
}

// GetNodeDID reads the node DID document from Git.
func (s *DIDStore) GetNodeDID() (*openanchor.DIDDocument, error) {
	return s.get("node-did")
}

// GetDeviceDID reads a device DID document from Git.
func (s *DIDStore) GetDeviceDID(deviceID string) (*openanchor.DIDDocument, error) {
	return s.get("device-" + deviceID)
}

func (s *DIDStore) save(name string, doc *openanchor.DIDDocument) (string, error) {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf(".nucleus/dids/%s.json", name)
	commit, err := s.git.WriteAndCommit(path, data, gitstore.CommitMessage{
		ResourceType: "DIDDocument",
		Operation:    "CREATE",
		ResourceID:   doc.ID,
		NodeID:       "anchor-service",
		Author:       "anchor-service",
		SiteID:       "local",
	})
	if err != nil {
		return "", fmt.Errorf("write DID: %w", err)
	}
	return commit, nil
}

func (s *DIDStore) get(name string) (*openanchor.DIDDocument, error) {
	path := fmt.Sprintf(".nucleus/dids/%s.json", name)
	data, err := s.git.Read(path)
	if err != nil {
		return nil, fmt.Errorf("DID %s not found: %w", name, err)
	}

	var doc openanchor.DIDDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse DID: %w", err)
	}
	return &doc, nil
}
