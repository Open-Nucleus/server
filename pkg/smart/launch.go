package smart

import (
	"fmt"
	"sync"
	"time"
)

// LaunchToken represents an EHR launch context token.
type LaunchToken struct {
	Token       string
	ClientID    string
	PatientID   string
	EncounterID string
	CreatedBy   string // device_id
	ExpiresAt   time.Time
}

// LaunchStore is an in-memory store for EHR launch tokens with TTL expiry.
type LaunchStore struct {
	mu     sync.Mutex
	tokens map[string]*LaunchToken
	ttl    time.Duration
	done   chan struct{}
}

// NewLaunchStore creates a new launch store with the given TTL.
// Default TTL is 5 minutes if ttl <= 0.
func NewLaunchStore(ttl time.Duration) *LaunchStore {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	s := &LaunchStore{
		tokens: make(map[string]*LaunchToken),
		ttl:    ttl,
		done:   make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Create generates a new launch token with the given context.
func (s *LaunchStore) Create(clientID, patientID, encounterID, createdBy string) (string, error) {
	if clientID == "" {
		return "", fmt.Errorf("client_id is required")
	}

	token, err := randomHex(32)
	if err != nil {
		return "", fmt.Errorf("generate launch token: %w", err)
	}

	lt := &LaunchToken{
		Token:       token,
		ClientID:    clientID,
		PatientID:   patientID,
		EncounterID: encounterID,
		CreatedBy:   createdBy,
		ExpiresAt:   time.Now().Add(s.ttl),
	}

	s.mu.Lock()
	s.tokens[token] = lt
	s.mu.Unlock()

	return token, nil
}

// Consume retrieves and removes a launch token (one-shot).
func (s *LaunchStore) Consume(token string) (*LaunchToken, error) {
	s.mu.Lock()
	lt, ok := s.tokens[token]
	if ok {
		delete(s.tokens, token)
	}
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("invalid or expired launch token")
	}

	if time.Now().After(lt.ExpiresAt) {
		return nil, fmt.Errorf("launch token has expired")
	}

	return lt, nil
}

// Close stops the background cleanup goroutine.
func (s *LaunchStore) Close() {
	close(s.done)
}

func (s *LaunchStore) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for tok, lt := range s.tokens {
				if now.After(lt.ExpiresAt) {
					delete(s.tokens, tok)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}
