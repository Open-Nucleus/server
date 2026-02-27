package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ConflictRecord represents a merge conflict stored in SQLite.
type ConflictRecord struct {
	ID             string
	ResourceType   string
	ResourceID     string
	PatientID      string
	Level          string
	Status         string
	DetectedAt     string
	LocalVersion   []byte
	RemoteVersion  []byte
	MergedVersion  []byte
	ChangedFields  []string
	Reason         string
	LocalNode      string
	RemoteNode     string
	PeerSiteID     string
	ResolvedAt     string
	ResolvedBy     string
	Resolution     string
}

// ConflictStore manages conflict records in SQLite.
type ConflictStore struct {
	db *sql.DB
}

// NewConflictStore creates a new conflict store.
func NewConflictStore(db *sql.DB) *ConflictStore {
	return &ConflictStore{db: db}
}

// Create inserts a new conflict record.
func (cs *ConflictStore) Create(c *ConflictRecord) error {
	fieldsJSON, _ := json.Marshal(c.ChangedFields)
	_, err := cs.db.Exec(
		`INSERT INTO conflicts (id, resource_type, resource_id, patient_id, level, status,
			detected_at, local_version, remote_version, merged_version, changed_fields, reason,
			local_node, remote_node, peer_site_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.ResourceType, c.ResourceID, c.PatientID, c.Level, c.Status,
		time.Now().UTC().Format(time.RFC3339), c.LocalVersion, c.RemoteVersion, c.MergedVersion,
		string(fieldsJSON), c.Reason, c.LocalNode, c.RemoteNode, c.PeerSiteID,
	)
	return err
}

// Get returns a single conflict by ID.
func (cs *ConflictStore) Get(id string) (*ConflictRecord, error) {
	var c ConflictRecord
	var fieldsJSON string
	var resolvedAt, resolvedBy, resolution sql.NullString

	err := cs.db.QueryRow(
		`SELECT id, resource_type, resource_id, patient_id, level, status, detected_at,
			local_version, remote_version, merged_version, changed_fields, reason,
			local_node, remote_node, peer_site_id, resolved_at, resolved_by, resolution
		FROM conflicts WHERE id = ?`, id,
	).Scan(
		&c.ID, &c.ResourceType, &c.ResourceID, &c.PatientID, &c.Level, &c.Status, &c.DetectedAt,
		&c.LocalVersion, &c.RemoteVersion, &c.MergedVersion, &fieldsJSON, &c.Reason,
		&c.LocalNode, &c.RemoteNode, &c.PeerSiteID, &resolvedAt, &resolvedBy, &resolution,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conflict not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(fieldsJSON), &c.ChangedFields)
	if resolvedAt.Valid {
		c.ResolvedAt = resolvedAt.String
	}
	if resolvedBy.Valid {
		c.ResolvedBy = resolvedBy.String
	}
	if resolution.Valid {
		c.Resolution = resolution.String
	}

	return &c, nil
}

// List returns conflicts with optional filtering.
func (cs *ConflictStore) List(statusFilter, levelFilter string, limit, offset int) ([]*ConflictRecord, int, error) {
	query := "SELECT id, resource_type, resource_id, patient_id, level, status, detected_at, local_version, remote_version, merged_version, changed_fields, reason, local_node, remote_node, peer_site_id FROM conflicts WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM conflicts WHERE 1=1"
	var args []any

	if statusFilter != "" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, statusFilter)
	}
	if levelFilter != "" {
		query += " AND level = ?"
		countQuery += " AND level = ?"
		args = append(args, levelFilter)
	}

	var total int
	if err := cs.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY detected_at DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, limit, offset)

	rows, err := cs.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var conflicts []*ConflictRecord
	for rows.Next() {
		var c ConflictRecord
		var fieldsJSON string
		if err := rows.Scan(&c.ID, &c.ResourceType, &c.ResourceID, &c.PatientID, &c.Level, &c.Status, &c.DetectedAt, &c.LocalVersion, &c.RemoteVersion, &c.MergedVersion, &fieldsJSON, &c.Reason, &c.LocalNode, &c.RemoteNode, &c.PeerSiteID); err != nil {
			return nil, 0, err
		}
		json.Unmarshal([]byte(fieldsJSON), &c.ChangedFields)
		conflicts = append(conflicts, &c)
	}
	return conflicts, total, rows.Err()
}

// Resolve marks a conflict as resolved.
func (cs *ConflictStore) Resolve(id, resolution, resolvedBy string, mergedResource []byte) error {
	result, err := cs.db.Exec(
		`UPDATE conflicts SET status = 'resolved', resolution = ?, resolved_by = ?,
			resolved_at = ?, merged_version = ? WHERE id = ?`,
		resolution, resolvedBy, time.Now().UTC().Format(time.RFC3339), mergedResource, id,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("conflict not found: %s", id)
	}
	return nil
}

// Defer marks a conflict as deferred.
func (cs *ConflictStore) Defer(id, reason string) error {
	result, err := cs.db.Exec(
		"UPDATE conflicts SET status = 'deferred', reason = ? WHERE id = ?",
		reason, id,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("conflict not found: %s", id)
	}
	return nil
}
