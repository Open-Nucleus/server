package store

import "database/sql"

// InitSchema creates the auth service SQLite tables.
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS deny_list (
			jti TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			added_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE INDEX IF NOT EXISTS idx_deny_list_device ON deny_list(device_id);

		CREATE TABLE IF NOT EXISTS revocations (
			device_id TEXT PRIMARY KEY,
			public_key TEXT NOT NULL,
			revoked_at TEXT NOT NULL DEFAULT (datetime('now')),
			revoked_by TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS node_info (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	return err
}
