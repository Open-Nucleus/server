package store

import (
	"database/sql"
	"time"
)

// QueueEntry represents an item in the anchor queue.
type QueueEntry struct {
	AnchorID   string
	MerkleRoot string
	GitHead    string
	NodeDID    string
	State      string
	EnqueuedAt string
}

// AnchorQueue manages the SQLite-backed anchor queue.
type AnchorQueue struct {
	db *sql.DB
}

// NewAnchorQueue creates a new AnchorQueue.
func NewAnchorQueue(db *sql.DB) *AnchorQueue {
	return &AnchorQueue{db: db}
}

// Enqueue adds an anchor to the queue.
func (q *AnchorQueue) Enqueue(anchorID, merkleRoot, gitHead, nodeDID string) error {
	_, err := q.db.Exec(
		`INSERT INTO anchor_queue (anchor_id, merkle_root, git_head, node_did, state, enqueued_at)
		 VALUES (?, ?, ?, ?, 'pending', ?)`,
		anchorID, merkleRoot, gitHead, nodeDID, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// ListPending returns all pending queue entries.
func (q *AnchorQueue) ListPending() ([]QueueEntry, error) {
	rows, err := q.db.Query(
		`SELECT anchor_id, merkle_root, git_head, node_did, state, enqueued_at
		 FROM anchor_queue WHERE state = 'pending' ORDER BY enqueued_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []QueueEntry
	for rows.Next() {
		var e QueueEntry
		if err := rows.Scan(&e.AnchorID, &e.MerkleRoot, &e.GitHead, &e.NodeDID, &e.State, &e.EnqueuedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// CountPending returns the number of pending entries.
func (q *AnchorQueue) CountPending() (int, error) {
	var count int
	err := q.db.QueryRow(`SELECT COUNT(*) FROM anchor_queue WHERE state = 'pending'`).Scan(&count)
	return count, err
}

// CountTotal returns the total number of processed entries.
func (q *AnchorQueue) CountTotal() (int, error) {
	var count int
	err := q.db.QueryRow(`SELECT COUNT(*) FROM anchor_queue`).Scan(&count)
	return count, err
}
