package transport

import (
	"context"
	"io"
)

// PeerNode represents a discovered peer.
type PeerNode struct {
	NodeID    string
	SiteID    string
	Address   string
	Port      int
	Transport string
}

// Capabilities describes what a transport can do.
type Capabilities struct {
	Discovery   bool
	Streaming   bool
	Constrained bool // bandwidth-limited (e.g., Bluetooth)
}

// SyncConn represents a connection to a peer for syncing.
type SyncConn interface {
	io.ReadWriteCloser
	RemoteNode() PeerNode
}

// Adapter is the interface for pluggable transport mechanisms.
type Adapter interface {
	// Name returns the transport name (e.g., "local_network", "bluetooth").
	Name() string

	// Capabilities returns what this transport supports.
	Capabilities() Capabilities

	// Start begins listening and discovery.
	Start(ctx context.Context) error

	// Stop shuts down the transport.
	Stop() error

	// Discover returns currently visible peers.
	Discover(ctx context.Context) ([]PeerNode, error)

	// Connect initiates a connection to a peer.
	Connect(ctx context.Context, peer PeerNode) (SyncConn, error)

	// Available returns true if the transport hardware/service is available.
	Available() bool
}
