package store

import "database/sql"

const schemaSQL = `
CREATE TABLE IF NOT EXISTS stock_levels (
	site_id            TEXT NOT NULL,
	medication_code    TEXT NOT NULL,
	quantity           INTEGER NOT NULL DEFAULT 0,
	unit               TEXT NOT NULL DEFAULT 'units',
	last_updated       TEXT NOT NULL DEFAULT '',
	earliest_expiry    TEXT NOT NULL DEFAULT '',
	daily_consumption_rate REAL NOT NULL DEFAULT 0.0,
	PRIMARY KEY (site_id, medication_code)
);

CREATE INDEX IF NOT EXISTS idx_stock_site ON stock_levels(site_id);
CREATE INDEX IF NOT EXISTS idx_stock_medication ON stock_levels(medication_code);

CREATE TABLE IF NOT EXISTS deliveries (
	id              TEXT PRIMARY KEY,
	site_id         TEXT NOT NULL,
	received_by     TEXT NOT NULL,
	delivery_date   TEXT NOT NULL,
	items_recorded  INTEGER NOT NULL DEFAULT 0,
	created_at      TEXT NOT NULL DEFAULT ''
);
`

func InitSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}
