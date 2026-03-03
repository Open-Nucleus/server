package openanchor

// StubBackend implements AnchorEngine with no actual blockchain backend.
// Anchor() returns ErrBackendNotConfigured, signaling that anchors should be queued.
type StubBackend struct{}

// NewStubBackend returns a new StubBackend.
func NewStubBackend() *StubBackend {
	return &StubBackend{}
}

func (s *StubBackend) Anchor(_ []byte, _ AnchorMetadata) (*AnchorReceipt, error) {
	return nil, ErrBackendNotConfigured
}

func (s *StubBackend) Available() bool {
	return false
}

func (s *StubBackend) Name() string {
	return "none"
}
