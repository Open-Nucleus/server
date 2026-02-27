package store

import (
	"database/sql"
	"fmt"
	"time"
)

// HistoryRecord represents a sync history entry.
type HistoryRecord struct {
	ID                string
	PeerNode          string
	Transport         string
	Direction         string
	State             string
	StartedAt         string
	CompletedAt       string
	ResourcesSent     int
	ResourcesReceived int
	ConflictsDetected int
	LocalHeadBefore   string
	LocalHeadAfter    string
	ErrorMessage      string
}

// HistoryStore manages sync history records.
type HistoryStore struct {
	db         *sql.DB
	maxEntries int
}

// NewHistoryStore creates a new history store.
func NewHistoryStore(db *sql.DB, maxEntries int) *HistoryStore {
	return &HistoryStore{db: db, maxEntries: maxEntries}
}

// Record inserts a new history entry.
func (hs *HistoryStore) Record(h *HistoryRecord) error {
	_, err := hs.db.Exec(
		`INSERT INTO sync_history (id, peer_node, transport, direction, state,
			started_at, completed_at, resources_sent, resources_received,
			conflicts_detected, local_head_before, local_head_after, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		h.ID, h.PeerNode, h.Transport, h.Direction, h.State,
		h.StartedAt, h.CompletedAt, h.ResourcesSent, h.ResourcesReceived,
		h.ConflictsDetected, h.LocalHeadBefore, h.LocalHeadAfter, h.ErrorMessage,
	)
	if err != nil {
		return err
	}

	// Prune old entries
	if hs.maxEntries > 0 {
		_, _ = hs.db.Exec(
			`DELETE FROM sync_history WHERE id NOT IN
				(SELECT id FROM sync_history ORDER BY started_at DESC LIMIT ?)`,
			hs.maxEntries,
		)
	}
	return nil
}

// List returns sync history entries with pagination.
func (hs *HistoryStore) List(limit, offset int) ([]*HistoryRecord, int, error) {
	var total int
	if err := hs.db.QueryRow("SELECT COUNT(*) FROM sync_history").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := hs.db.Query(
		`SELECT id, peer_node, transport, direction, state, started_at, completed_at,
			resources_sent, resources_received, conflicts_detected,
			local_head_before, local_head_after, error_message
		FROM sync_history ORDER BY started_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []*HistoryRecord
	for rows.Next() {
		var h HistoryRecord
		var completedAt sql.NullString
		if err := rows.Scan(&h.ID, &h.PeerNode, &h.Transport, &h.Direction, &h.State,
			&h.StartedAt, &completedAt, &h.ResourcesSent, &h.ResourcesReceived,
			&h.ConflictsDetected, &h.LocalHeadBefore, &h.LocalHeadAfter, &h.ErrorMessage); err != nil {
			return nil, 0, err
		}
		if completedAt.Valid {
			h.CompletedAt = completedAt.String
		}
		entries = append(entries, &h)
	}
	return entries, total, rows.Err()
}

// Get returns a single history entry by ID.
func (hs *HistoryStore) Get(id string) (*HistoryRecord, error) {
	var h HistoryRecord
	var completedAt sql.NullString
	err := hs.db.QueryRow(
		`SELECT id, peer_node, transport, direction, state, started_at, completed_at,
			resources_sent, resources_received, conflicts_detected,
			local_head_before, local_head_after, error_message
		FROM sync_history WHERE id = ?`, id,
	).Scan(&h.ID, &h.PeerNode, &h.Transport, &h.Direction, &h.State,
		&h.StartedAt, &completedAt, &h.ResourcesSent, &h.ResourcesReceived,
		&h.ConflictsDetected, &h.LocalHeadBefore, &h.LocalHeadAfter, &h.ErrorMessage)
	if err != nil {
		return nil, fmt.Errorf("history entry not found: %s", id)
	}
	if completedAt.Valid {
		h.CompletedAt = completedAt.String
	}
	return &h, nil
}

// RecordCompleted updates a sync entry as completed.
func (hs *HistoryStore) RecordCompleted(id string, sent, received, conflicts int, headAfter string) error {
	_, err := hs.db.Exec(
		`UPDATE sync_history SET state = 'completed', completed_at = ?,
			resources_sent = ?, resources_received = ?, conflicts_detected = ?,
			local_head_after = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), sent, received, conflicts, headAfter, id,
	)
	return err
}

// RecordFailed updates a sync entry as failed.
func (hs *HistoryStore) RecordFailed(id, errMsg string) error {
	_, err := hs.db.Exec(
		`UPDATE sync_history SET state = 'failed', completed_at = ?, error_message = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), errMsg, id,
	)
	return err
}
