package localnet

import (
	"context"
	"fmt"

	"github.com/FibrinLab/open-nucleus/services/sync/internal/transport"
)

// Adapter implements local network transport using mDNS discovery and gRPC over TCP.
type Adapter struct {
	nodeID      string
	siteID      string
	mdnsService string
	port        int
}

// New creates a new local network adapter.
func New(nodeID, siteID, mdnsService string, port int) *Adapter {
	if mdnsService == "" {
		mdnsService = "_nucleus._tcp"
	}
	if port == 0 {
		port = 50060
	}
	return &Adapter{
		nodeID:      nodeID,
		siteID:      siteID,
		mdnsService: mdnsService,
		port:        port,
	}
}

func (a *Adapter) Name() string { return "local_network" }

func (a *Adapter) Capabilities() transport.Capabilities {
	return transport.Capabilities{
		Discovery: true,
		Streaming: true,
	}
}

func (a *Adapter) Start(_ context.Context) error {
	// In production, this would:
	// 1. Register mDNS service using zeroconf
	// 2. Start gRPC server for node-to-node protocol
	return nil
}

func (a *Adapter) Stop() error {
	return nil
}

func (a *Adapter) Discover(_ context.Context) ([]transport.PeerNode, error) {
	// In production, this would use zeroconf.NewResolver to browse for _nucleus._tcp services
	return nil, nil
}

func (a *Adapter) Connect(_ context.Context, peer transport.PeerNode) (transport.SyncConn, error) {
	return nil, fmt.Errorf("local network sync connection not yet implemented for %s", peer.NodeID)
}

func (a *Adapter) Available() bool {
	return true // local network is always available
}
