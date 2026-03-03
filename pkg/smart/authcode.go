package smart

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AuthCode represents a SMART authorization code with associated context.
type AuthCode struct {
	Code           string
	ClientID       string
	RedirectURI    string
	Scope          string // granted scopes
	CodeChallenge  string // S256 PKCE
	DeviceID       string // authenticated user
	PractitionerID string
	SiteID         string
	Role           string
	PatientID      string // launch context (optional)
	EncounterID    string // launch context (optional)
	ExpiresAt      time.Time
}

// AuthCodeParams holds the parameters for generating an auth code.
type AuthCodeParams struct {
	ClientID       string
	RedirectURI    string
	Scope          string
	CodeChallenge  string
	DeviceID       string
	PractitionerID string
	SiteID         string
	Role           string
	PatientID      string
	EncounterID    string
}

// AuthCodeStore is an in-memory store for authorization codes with TTL expiry.
type AuthCodeStore struct {
	mu    sync.Mutex
	codes map[string]*AuthCode
	ttl   time.Duration
	done  chan struct{}
}

// NewAuthCodeStore creates a new auth code store with the given TTL.
// Default TTL is 60 seconds if ttl <= 0.
func NewAuthCodeStore(ttl time.Duration) *AuthCodeStore {
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	s := &AuthCodeStore{
		codes: make(map[string]*AuthCode),
		ttl:   ttl,
		done:  make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Generate creates a new authorization code with the given parameters.
func (s *AuthCodeStore) Generate(params AuthCodeParams) (string, error) {
	if params.ClientID == "" {
		return "", fmt.Errorf("client_id is required")
	}
	if params.RedirectURI == "" {
		return "", fmt.Errorf("redirect_uri is required")
	}

	code, err := randomHex(32)
	if err != nil {
		return "", fmt.Errorf("generate auth code: %w", err)
	}

	ac := &AuthCode{
		Code:           code,
		ClientID:       params.ClientID,
		RedirectURI:    params.RedirectURI,
		Scope:          params.Scope,
		CodeChallenge:  params.CodeChallenge,
		DeviceID:       params.DeviceID,
		PractitionerID: params.PractitionerID,
		SiteID:         params.SiteID,
		Role:           params.Role,
		PatientID:      params.PatientID,
		EncounterID:    params.EncounterID,
		ExpiresAt:      time.Now().Add(s.ttl),
	}

	s.mu.Lock()
	s.codes[code] = ac
	s.mu.Unlock()

	return code, nil
}

// Exchange validates and consumes an authorization code (one-shot).
func (s *AuthCodeStore) Exchange(code, clientID, codeVerifier, redirectURI string) (*AuthCode, error) {
	s.mu.Lock()
	ac, ok := s.codes[code]
	if ok {
		delete(s.codes, code) // one-shot: consume immediately
	}
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("invalid or expired authorization code")
	}

	if time.Now().After(ac.ExpiresAt) {
		return nil, fmt.Errorf("authorization code has expired")
	}

	if ac.ClientID != clientID {
		return nil, fmt.Errorf("client_id mismatch")
	}

	if ac.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect_uri mismatch")
	}

	// PKCE verification.
	if ac.CodeChallenge != "" {
		if codeVerifier == "" {
			return nil, fmt.Errorf("code_verifier is required for PKCE")
		}
		if !ValidatePKCE(codeVerifier, ac.CodeChallenge) {
			return nil, fmt.Errorf("PKCE verification failed")
		}
	}

	return ac, nil
}

// Close stops the background cleanup goroutine.
func (s *AuthCodeStore) Close() {
	close(s.done)
}

func (s *AuthCodeStore) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			s.mu.Lock()
			for code, ac := range s.codes {
				if now.After(ac.ExpiresAt) {
					delete(s.codes, code)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			return
		}
	}
}

// ValidatePKCE checks that the code verifier matches the S256 code challenge.
// challenge = BASE64URL(SHA256(verifier))
func ValidatePKCE(codeVerifier, codeChallenge string) bool {
	h := sha256.Sum256([]byte(codeVerifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return computed == codeChallenge
}

// GeneratePKCEChallenge computes the S256 challenge for a given verifier.
func GeneratePKCEChallenge(codeVerifier string) string {
	h := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
