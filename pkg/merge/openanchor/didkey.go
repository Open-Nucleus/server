package openanchor

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"strings"

	anchor "github.com/Open-Nucleus/open-anchor/go"
	"github.com/Open-Nucleus/open-anchor/go/backends/didkey"
)

// sharedDIDKeyBackend is the singleton did:key backend from open-anchor.
var sharedDIDKeyBackend = didkey.New()

// DIDKeyFromEd25519 generates a did:key from an Ed25519 public key.
func DIDKeyFromEd25519(pub ed25519.PublicKey) (string, *DIDDocument, error) {
	doc, err := sharedDIDKeyBackend.Create(context.Background(), pub, anchor.DIDOptions{})
	if err != nil {
		return "", nil, err
	}
	return doc.ID, fromExternalDIDDoc(doc), nil
}

// ResolveDIDKey parses a did:key string and returns the DIDDocument.
func ResolveDIDKey(did string) (*DIDDocument, error) {
	if !strings.HasPrefix(did, "did:key:z") {
		return nil, fmt.Errorf("%w: expected did:key:z..., got %s", ErrInvalidDID, did)
	}
	doc, err := sharedDIDKeyBackend.Resolve(context.Background(), did)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDID, err)
	}
	return fromExternalDIDDoc(doc), nil
}

// ExtractPublicKey extracts the Ed25519 public key from a did:key string.
func ExtractPublicKey(did string) (ed25519.PublicKey, error) {
	doc, err := sharedDIDKeyBackend.Resolve(context.Background(), did)
	if err != nil {
		return nil, err
	}
	if len(doc.VerificationMethod) == 0 {
		return nil, fmt.Errorf("%w: no verification method", ErrInvalidDID)
	}
	return anchor.PublicKeyFromMultibase(doc.VerificationMethod[0].PublicKeyMultibase)
}
