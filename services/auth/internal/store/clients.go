package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/smart"
)

const smartClientsPath = ".nucleus/smart-clients"

// ClientStore manages SMART client registrations in Git + SQLite.
type ClientStore struct {
	git gitstore.Store
	db  *sql.DB
}

// NewClientStore creates a new client store.
func NewClientStore(git gitstore.Store, db *sql.DB) *ClientStore {
	return &ClientStore{git: git, db: db}
}

// InitClientSchema creates the smart_clients SQLite table.
func InitClientSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS smart_clients (
			client_id TEXT PRIMARY KEY,
			client_name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			scope TEXT NOT NULL,
			registered_at TEXT NOT NULL,
			registered_by TEXT NOT NULL
		);
	`)
	return err
}

// Save stores a SMART client in Git and upserts the SQLite index.
func (s *ClientStore) Save(client *smart.Client) (string, error) {
	data, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal client: %w", err)
	}

	path := filepath.Join(smartClientsPath, client.ClientID+".json")
	msg := gitstore.CommitMessage{
		ResourceType: "SmartClient",
		Operation:    "REGISTER",
		ResourceID:   client.ClientID,
		NodeID:       "auth",
		Author:       client.RegisteredBy,
		SiteID:       "local",
		Timestamp:    time.Now().UTC(),
	}

	hash, err := s.git.WriteAndCommit(path, data, msg)
	if err != nil {
		return "", fmt.Errorf("git write client: %w", err)
	}

	// Upsert SQLite index.
	_, err = s.db.Exec(`
		INSERT INTO smart_clients (client_id, client_name, status, scope, registered_at, registered_by)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(client_id) DO UPDATE SET
			client_name = excluded.client_name,
			status = excluded.status,
			scope = excluded.scope
	`, client.ClientID, client.ClientName, client.Status, client.Scope, client.RegisteredAt, client.RegisteredBy)
	if err != nil {
		return hash, fmt.Errorf("sqlite upsert client: %w", err)
	}

	return hash, nil
}

// Get loads a SMART client from Git.
func (s *ClientStore) Get(clientID string) (*smart.Client, error) {
	path := filepath.Join(smartClientsPath, clientID+".json")
	data, err := s.git.Read(path)
	if err != nil {
		return nil, fmt.Errorf("client not found: %s", clientID)
	}

	var client smart.Client
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, fmt.Errorf("unmarshal client: %w", err)
	}
	return &client, nil
}

// List returns all SMART clients from the SQLite index.
func (s *ClientStore) List() ([]*smart.Client, error) {
	rows, err := s.db.Query(`SELECT client_id FROM smart_clients ORDER BY registered_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	defer rows.Close()

	var clients []*smart.Client
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		client, err := s.Get(id)
		if err != nil {
			continue
		}
		clients = append(clients, client)
	}
	return clients, nil
}

// Update modifies a client's status/scope and persists.
func (s *ClientStore) Update(clientID, status, approvedBy, scope string) (*smart.Client, error) {
	client, err := s.Get(clientID)
	if err != nil {
		return nil, err
	}

	if status != "" {
		client.Status = smart.ClientStatus(status)
	}
	if approvedBy != "" {
		client.ApprovedBy = approvedBy
		client.ApprovedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if scope != "" {
		client.Scope = scope
	}

	if _, err := s.Save(client); err != nil {
		return nil, err
	}
	return client, nil
}

// Delete removes a SMART client from Git (by writing empty) and SQLite.
func (s *ClientStore) Delete(clientID string) error {
	// Write empty file to "delete" in Git (no DeleteAndCommit method available).
	path := filepath.Join(smartClientsPath, clientID+".json")
	msg := gitstore.CommitMessage{
		ResourceType: "SmartClient",
		Operation:    "DELETE",
		ResourceID:   clientID,
		NodeID:       "auth",
		Author:       "system",
		SiteID:       "local",
		Timestamp:    time.Now().UTC(),
	}
	// Write a tombstone marker.
	tombstone := []byte(`{"deleted":true}`)
	if _, err := s.git.WriteAndCommit(path, tombstone, msg); err != nil {
		return fmt.Errorf("git delete client: %w", err)
	}

	_, err := s.db.Exec(`DELETE FROM smart_clients WHERE client_id = ?`, clientID)
	return err
}
