package store

import "database/sql"

// InitSchema creates the sync service SQLite tables.
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS conflicts (
			id TEXT PRIMARY KEY,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			patient_id TEXT NOT NULL DEFAULT '',
			level TEXT NOT NULL DEFAULT 'review',
			status TEXT NOT NULL DEFAULT 'pending',
			detected_at TEXT NOT NULL DEFAULT (datetime('now')),
			local_version BLOB,
			remote_version BLOB,
			merged_version BLOB,
			changed_fields TEXT NOT NULL DEFAULT '[]',
			reason TEXT NOT NULL DEFAULT '',
			local_node TEXT NOT NULL DEFAULT '',
			remote_node TEXT NOT NULL DEFAULT '',
			peer_site_id TEXT NOT NULL DEFAULT '',
			resolved_at TEXT,
			resolved_by TEXT,
			resolution TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_conflicts_status ON conflicts(status);
		CREATE INDEX IF NOT EXISTS idx_conflicts_level ON conflicts(level);
		CREATE INDEX IF NOT EXISTS idx_conflicts_patient ON conflicts(patient_id);

		CREATE TABLE IF NOT EXISTS sync_history (
			id TEXT PRIMARY KEY,
			peer_node TEXT NOT NULL,
			transport TEXT NOT NULL DEFAULT '',
			direction TEXT NOT NULL DEFAULT 'bidirectional',
			state TEXT NOT NULL DEFAULT 'completed',
			started_at TEXT NOT NULL DEFAULT (datetime('now')),
			completed_at TEXT,
			resources_sent INTEGER NOT NULL DEFAULT 0,
			resources_received INTEGER NOT NULL DEFAULT 0,
			conflicts_detected INTEGER NOT NULL DEFAULT 0,
			local_head_before TEXT NOT NULL DEFAULT '',
			local_head_after TEXT NOT NULL DEFAULT '',
			error_message TEXT NOT NULL DEFAULT ''
		);

		CREATE INDEX IF NOT EXISTS idx_sync_history_peer ON sync_history(peer_node);
		CREATE INDEX IF NOT EXISTS idx_sync_history_state ON sync_history(state);

		CREATE TABLE IF NOT EXISTS peer_state (
			node_id TEXT PRIMARY KEY,
			site_id TEXT NOT NULL DEFAULT '',
			public_key BLOB,
			trusted INTEGER NOT NULL DEFAULT 0,
			last_seen TEXT NOT NULL DEFAULT (datetime('now')),
			their_head TEXT NOT NULL DEFAULT '',
			transport TEXT NOT NULL DEFAULT '',
			revoked INTEGER NOT NULL DEFAULT 0
		);
	`)
	return err
}
