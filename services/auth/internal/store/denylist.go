package store

import (
	"database/sql"
	"sync"
	"time"
)

// DenyList tracks revoked JWT IDs using an in-memory set backed by SQLite.
type DenyList struct {
	mu    sync.RWMutex
	items map[string]bool
	db    *sql.DB
}

// NewDenyList creates a new deny list backed by the given database.
func NewDenyList(db *sql.DB) *DenyList {
	return &DenyList{
		items: make(map[string]bool),
		db:    db,
	}
}

// LoadFromDB populates the in-memory set from SQLite.
func (dl *DenyList) LoadFromDB() error {
	rows, err := dl.db.Query("SELECT jti FROM deny_list")
	if err != nil {
		return err
	}
	defer rows.Close()

	dl.mu.Lock()
	defer dl.mu.Unlock()

	for rows.Next() {
		var jti string
		if err := rows.Scan(&jti); err != nil {
			return err
		}
		dl.items[jti] = true
	}
	return rows.Err()
}

// Add adds a JTI to the deny list (both in-memory and SQLite).
func (dl *DenyList) Add(jti, deviceID string) error {
	dl.mu.Lock()
	dl.items[jti] = true
	dl.mu.Unlock()

	_, err := dl.db.Exec(
		"INSERT OR IGNORE INTO deny_list (jti, device_id, added_at) VALUES (?, ?, ?)",
		jti, deviceID, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// IsDenied returns true if the JTI is in the deny list. O(1) in-memory lookup.
func (dl *DenyList) IsDenied(jti string) bool {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return dl.items[jti]
}

// AddAllForDevice adds deny entries for all JTIs associated with a device.
func (dl *DenyList) AddAllForDevice(deviceID string) error {
	rows, err := dl.db.Query("SELECT jti FROM deny_list WHERE device_id = ?", deviceID)
	if err != nil {
		return err
	}
	defer rows.Close()

	dl.mu.Lock()
	defer dl.mu.Unlock()

	for rows.Next() {
		var jti string
		if err := rows.Scan(&jti); err != nil {
			return err
		}
		dl.items[jti] = true
	}
	return rows.Err()
}

// AddRevocation records a device revocation.
func (dl *DenyList) AddRevocation(deviceID, publicKey, revokedBy, reason string) error {
	_, err := dl.db.Exec(
		"INSERT OR REPLACE INTO revocations (device_id, public_key, revoked_at, revoked_by, reason) VALUES (?, ?, ?, ?, ?)",
		deviceID, publicKey, time.Now().UTC().Format(time.RFC3339), revokedBy, reason,
	)
	return err
}

// IsRevoked checks if a device has been revoked.
func (dl *DenyList) IsRevoked(deviceID string) (bool, string, string) {
	var revokedAt, reason string
	err := dl.db.QueryRow("SELECT revoked_at, reason FROM revocations WHERE device_id = ?", deviceID).Scan(&revokedAt, &reason)
	if err != nil {
		return false, "", ""
	}
	return true, revokedAt, reason
}
