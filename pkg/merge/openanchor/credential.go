package openanchor

import (
	"context"
	"crypto/ed25519"

	anchor "github.com/Open-Nucleus/open-anchor/go"
	"github.com/Open-Nucleus/open-anchor/go/backends/didkey"
)

// credentialEngine is a shared IdentityEngine for standalone credential functions.
var credentialEngine = anchor.NewIdentityEngine(didkey.New())

// IssueCredentialLocal creates a Verifiable Credential signed with Ed25519.
func IssueCredentialLocal(claims CredentialClaims, issuerDID string, issuerKey ed25519.PrivateKey) (*VerifiableCredential, error) {
	extClaims := toExternalClaims(claims)
	vc, err := credentialEngine.IssueCredential(context.Background(), issuerDID, issuerKey, extClaims)
	if err != nil {
		return nil, err
	}
	return fromExternalVC(vc), nil
}

// VerifyCredentialLocal verifies a VC's Ed25519 signature.
func VerifyCredentialLocal(vc *VerifiableCredential) (*VerificationResult, error) {
	extVC := toExternalVC(vc)
	result, err := credentialEngine.VerifyCredential(context.Background(), extVC)
	if err != nil {
		return nil, err
	}
	return fromExternalVerification(result, vc.Issuer), nil
}
