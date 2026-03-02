package openanchor

import (
	"crypto/ed25519"
	"errors"
	"time"
)

// Sentinel errors.
var (
	ErrBackendNotConfigured = errors.New("anchor backend not configured")
	ErrInvalidDID           = errors.New("invalid DID format")
	ErrUnsupportedDIDMethod = errors.New("unsupported DID method")
	ErrInvalidSignature     = errors.New("invalid credential signature")
	ErrCredentialExpired    = errors.New("credential has expired")
	ErrIssuerNotResolvable  = errors.New("issuer DID not resolvable")
)

// AnchorEngine submits Merkle roots to an external anchoring backend.
type AnchorEngine interface {
	// Anchor submits a Merkle root for anchoring. Returns receipt or error.
	Anchor(root []byte, metadata AnchorMetadata) (*AnchorReceipt, error)
	// Available reports whether the backend is reachable.
	Available() bool
	// Name returns the backend identifier (e.g. "iota", "none").
	Name() string
}

// AnchorMetadata accompanies an anchor submission.
type AnchorMetadata struct {
	GitHead   string
	NodeDID   string
	Timestamp time.Time
}

// AnchorReceipt is returned by a successful anchor submission.
type AnchorReceipt struct {
	BackendName string
	TxID        string
	Timestamp   time.Time
}

// IdentityEngine handles DID and Verifiable Credential operations.
type IdentityEngine interface {
	// GenerateDID creates a did:key from an Ed25519 public key.
	GenerateDID(pub ed25519.PublicKey) (*DIDDocument, error)
	// ResolveDID resolves a DID string to a DIDDocument.
	ResolveDID(did string) (*DIDDocument, error)
	// IssueCredential creates a signed Verifiable Credential.
	IssueCredential(claims CredentialClaims, issuerDID string, issuerKey ed25519.PrivateKey) (*VerifiableCredential, error)
	// VerifyCredential checks a VC's signature and expiry.
	VerifyCredential(vc *VerifiableCredential) (*VerificationResult, error)
}

// MerkleTree computes Merkle roots from file data.
type MerkleTree interface {
	// ComputeRoot returns the SHA-256 Merkle root for the given file entries.
	ComputeRoot(entries []FileEntry) ([]byte, error)
}

// FileEntry represents a file in the Merkle tree.
type FileEntry struct {
	Path string
	Hash []byte // SHA-256 of file contents
}

// DIDDocument represents a W3C DID Document.
type DIDDocument struct {
	Context            []string             `json:"@context"`
	ID                 string               `json:"id"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	Authentication     []string             `json:"authentication"`
	AssertionMethod    []string             `json:"assertionMethod"`
	Created            string               `json:"created,omitempty"`
}

// VerificationMethod is a DID verification method entry.
type VerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase"`
}

// VerifiableCredential represents a W3C Verifiable Credential.
type VerifiableCredential struct {
	Context           []string        `json:"@context"`
	ID                string          `json:"id"`
	Type              []string        `json:"type"`
	Issuer            string          `json:"issuer"`
	IssuanceDate      string          `json:"issuanceDate"`
	ExpirationDate    string          `json:"expirationDate,omitempty"`
	CredentialSubject map[string]any  `json:"credentialSubject"`
	Proof             *CredentialProof `json:"proof"`
}

// CredentialProof contains the cryptographic proof for a VC.
type CredentialProof struct {
	Type               string `json:"type"`
	Created            string `json:"created"`
	VerificationMethod string `json:"verificationMethod"`
	ProofPurpose       string `json:"proofPurpose"`
	ProofValue         string `json:"proofValue"` // base58btc-encoded signature
}

// CredentialClaims are the claims to include in a VC.
type CredentialClaims struct {
	ID             string         // Credential ID (e.g. "urn:uuid:...")
	Types          []string       // Additional VC types beyond "VerifiableCredential"
	Subject        map[string]any // credentialSubject fields
	ExpirationDate string         // Optional RFC3339 expiry
}

// VerificationResult is returned by VerifyCredential.
type VerificationResult struct {
	Valid   bool   `json:"valid"`
	Issuer  string `json:"issuer"`
	Message string `json:"message"`
}

// AnchorResult is the outcome of a Merkle anchoring attempt.
type AnchorResult struct {
	AnchorID   string
	MerkleRoot string // hex-encoded
	GitHead    string
	State      string // "queued", "confirmed", "failed"
	Receipt    *AnchorReceipt
	Timestamp  time.Time
}
