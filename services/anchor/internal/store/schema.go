package store

import "database/sql"

// InitSchema creates the anchor queue table.
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS anchor_queue (
			anchor_id    TEXT PRIMARY KEY,
			merkle_root  TEXT NOT NULL,
			git_head     TEXT NOT NULL,
			node_did     TEXT NOT NULL,
			state        TEXT NOT NULL DEFAULT 'pending',
			enqueued_at  TEXT NOT NULL,
			processed_at TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_queue_state ON anchor_queue(state);
		CREATE INDEX IF NOT EXISTS idx_queue_enqueued ON anchor_queue(enqueued_at);
	`)
	return err
}
