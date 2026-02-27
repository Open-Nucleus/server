package store

import (
	"database/sql"
	"time"
)

// PeerRecord represents a known peer node.
type PeerRecord struct {
	NodeID    string
	SiteID    string
	PublicKey []byte
	Trusted   bool
	LastSeen  string
	TheirHead string
	Transport string
	Revoked   bool
}

// PeerStore manages peer state in SQLite.
type PeerStore struct {
	db *sql.DB
}

// NewPeerStore creates a new peer store.
func NewPeerStore(db *sql.DB) *PeerStore {
	return &PeerStore{db: db}
}

// Upsert creates or updates a peer record.
func (ps *PeerStore) Upsert(p *PeerRecord) error {
	_, err := ps.db.Exec(
		`INSERT INTO peer_state (node_id, site_id, public_key, trusted, last_seen, their_head, transport, revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(node_id) DO UPDATE SET
			site_id = excluded.site_id,
			public_key = excluded.public_key,
			last_seen = excluded.last_seen,
			their_head = excluded.their_head,
			transport = excluded.transport`,
		p.NodeID, p.SiteID, p.PublicKey, p.Trusted,
		time.Now().UTC().Format(time.RFC3339), p.TheirHead, p.Transport, p.Revoked,
	)
	return err
}

// Get returns a peer by node ID.
func (ps *PeerStore) Get(nodeID string) (*PeerRecord, error) {
	var p PeerRecord
	var trusted, revoked int
	err := ps.db.QueryRow(
		`SELECT node_id, site_id, public_key, trusted, last_seen, their_head, transport, revoked
		FROM peer_state WHERE node_id = ?`, nodeID,
	).Scan(&p.NodeID, &p.SiteID, &p.PublicKey, &trusted, &p.LastSeen, &p.TheirHead, &p.Transport, &revoked)
	if err != nil {
		return nil, err
	}
	p.Trusted = trusted != 0
	p.Revoked = revoked != 0
	return &p, nil
}

// List returns all known peers.
func (ps *PeerStore) List() ([]*PeerRecord, error) {
	rows, err := ps.db.Query(
		`SELECT node_id, site_id, public_key, trusted, last_seen, their_head, transport, revoked
		FROM peer_state ORDER BY last_seen DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*PeerRecord
	for rows.Next() {
		var p PeerRecord
		var trusted, revoked int
		if err := rows.Scan(&p.NodeID, &p.SiteID, &p.PublicKey, &trusted, &p.LastSeen, &p.TheirHead, &p.Transport, &revoked); err != nil {
			return nil, err
		}
		p.Trusted = trusted != 0
		p.Revoked = revoked != 0
		peers = append(peers, &p)
	}
	return peers, rows.Err()
}

// Trust marks a peer as trusted.
func (ps *PeerStore) Trust(nodeID string) error {
	_, err := ps.db.Exec("UPDATE peer_state SET trusted = 1 WHERE node_id = ?", nodeID)
	return err
}

// Untrust marks a peer as untrusted.
func (ps *PeerStore) Untrust(nodeID string) error {
	_, err := ps.db.Exec("UPDATE peer_state SET trusted = 0 WHERE node_id = ?", nodeID)
	return err
}

// MarkRevoked marks a peer as revoked.
func (ps *PeerStore) MarkRevoked(nodeID string) error {
	_, err := ps.db.Exec("UPDATE peer_state SET revoked = 1 WHERE node_id = ?", nodeID)
	return err
}
