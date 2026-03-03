package openanchor

import (
	"crypto/ed25519"
	"fmt"
	"strings"
)

// LocalIdentityEngine implements IdentityEngine using local Ed25519 crypto.
// No network calls — fully offline.
type LocalIdentityEngine struct{}

// NewLocalIdentityEngine returns a new LocalIdentityEngine.
func NewLocalIdentityEngine() *LocalIdentityEngine {
	return &LocalIdentityEngine{}
}

func (e *LocalIdentityEngine) GenerateDID(pub ed25519.PublicKey) (*DIDDocument, error) {
	_, doc, err := DIDKeyFromEd25519(pub)
	return doc, err
}

func (e *LocalIdentityEngine) ResolveDID(did string) (*DIDDocument, error) {
	if !strings.HasPrefix(did, "did:key:") {
		return nil, fmt.Errorf("%w: only did:key is supported, got %s", ErrUnsupportedDIDMethod, did)
	}
	return ResolveDIDKey(did)
}

func (e *LocalIdentityEngine) IssueCredential(claims CredentialClaims, issuerDID string, issuerKey ed25519.PrivateKey) (*VerifiableCredential, error) {
	return IssueCredentialLocal(claims, issuerDID, issuerKey)
}

func (e *LocalIdentityEngine) VerifyCredential(vc *VerifiableCredential) (*VerificationResult, error) {
	return VerifyCredentialLocal(vc)
}
