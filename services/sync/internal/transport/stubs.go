package transport

import (
	"context"
	"fmt"
)

// StubAdapter is a placeholder for unimplemented transports.
type StubAdapter struct {
	TransportName string
}

func (s *StubAdapter) Name() string { return s.TransportName }

func (s *StubAdapter) Capabilities() Capabilities {
	return Capabilities{}
}

func (s *StubAdapter) Start(_ context.Context) error {
	return nil
}

func (s *StubAdapter) Stop() error {
	return nil
}

func (s *StubAdapter) Discover(_ context.Context) ([]PeerNode, error) {
	return nil, nil
}

func (s *StubAdapter) Connect(_ context.Context, _ PeerNode) (SyncConn, error) {
	return nil, fmt.Errorf("%s transport not implemented", s.TransportName)
}

func (s *StubAdapter) Available() bool {
	return false
}
