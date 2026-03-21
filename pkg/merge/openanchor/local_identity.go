package openanchor

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"strings"

	anchor "github.com/Open-Nucleus/open-anchor/go"
	"github.com/Open-Nucleus/open-anchor/go/backends/didkey"
)

// LocalIdentityEngine implements IdentityEngine using the external open-anchor library.
type LocalIdentityEngine struct {
	engine *anchor.IdentityEngine
}

// NewLocalIdentityEngine returns a new LocalIdentityEngine backed by open-anchor.
func NewLocalIdentityEngine() *LocalIdentityEngine {
	return &LocalIdentityEngine{
		engine: anchor.NewIdentityEngine(didkey.New()),
	}
}

func (e *LocalIdentityEngine) GenerateDID(pub ed25519.PublicKey) (*DIDDocument, error) {
	doc, err := e.engine.CreateDID(context.Background(), "key", pub, anchor.DIDOptions{})
	if err != nil {
		return nil, err
	}
	return fromExternalDIDDoc(doc), nil
}

func (e *LocalIdentityEngine) ResolveDID(did string) (*DIDDocument, error) {
	if !strings.HasPrefix(did, "did:key:") {
		return nil, fmt.Errorf("%w: only did:key is supported, got %s", ErrUnsupportedDIDMethod, did)
	}
	doc, err := e.engine.ResolveDID(context.Background(), did)
	if err != nil {
		return nil, err
	}
	return fromExternalDIDDoc(doc), nil
}

func (e *LocalIdentityEngine) IssueCredential(claims CredentialClaims, issuerDID string, issuerKey ed25519.PrivateKey) (*VerifiableCredential, error) {
	extClaims := toExternalClaims(claims)
	vc, err := e.engine.IssueCredential(context.Background(), issuerDID, issuerKey, extClaims)
	if err != nil {
		return nil, err
	}
	return fromExternalVC(vc), nil
}

func (e *LocalIdentityEngine) VerifyCredential(vc *VerifiableCredential) (*VerificationResult, error) {
	extVC := toExternalVC(vc)
	result, err := e.engine.VerifyCredential(context.Background(), extVC)
	if err != nil {
		return nil, err
	}
	return fromExternalVerification(result, vc.Issuer), nil
}
