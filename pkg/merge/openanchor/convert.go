package openanchor

import (
	anchor "github.com/Open-Nucleus/open-anchor/go"
)

// fromExternalDIDDoc converts an external DIDDocument to the in-repo type.
func fromExternalDIDDoc(ext *anchor.DIDDocument) *DIDDocument {
	doc := &DIDDocument{
		Context:         ext.Context,
		ID:              ext.ID,
		Authentication:  ext.Authentication,
		AssertionMethod: ext.AssertionMethod,
		Created:         ext.Created,
	}
	for _, vm := range ext.VerificationMethod {
		doc.VerificationMethod = append(doc.VerificationMethod, VerificationMethod{
			ID:                 vm.ID,
			Type:               vm.Type,
			Controller:         vm.Controller,
			PublicKeyMultibase: vm.PublicKeyMultibase,
		})
	}
	return doc
}

// fromExternalVC converts an external VerifiableCredential to the in-repo type.
func fromExternalVC(ext *anchor.VerifiableCredential) *VerifiableCredential {
	vc := &VerifiableCredential{
		Context:           ext.Context,
		ID:                ext.ID,
		Type:              ext.Type,
		Issuer:            ext.Issuer,
		IssuanceDate:      ext.IssuanceDate,
		ExpirationDate:    ext.ExpirationDate,
		CredentialSubject: ext.Subject,
		Proof: &CredentialProof{
			Type:               ext.Proof.Type,
			Created:            ext.Proof.Created,
			VerificationMethod: ext.Proof.VerificationMethod,
			ProofPurpose:       ext.Proof.ProofPurpose,
			ProofValue:         ext.Proof.ProofValue,
		},
	}
	return vc
}

// toExternalVC converts an in-repo VerifiableCredential to the external type.
func toExternalVC(vc *VerifiableCredential) *anchor.VerifiableCredential {
	ext := &anchor.VerifiableCredential{
		Context:        vc.Context,
		ID:             vc.ID,
		Type:           vc.Type,
		Issuer:         vc.Issuer,
		IssuanceDate:   vc.IssuanceDate,
		ExpirationDate: vc.ExpirationDate,
		Subject:        vc.CredentialSubject,
	}
	if vc.Proof != nil {
		ext.Proof = anchor.CredentialProof{
			Type:               vc.Proof.Type,
			Created:            vc.Proof.Created,
			VerificationMethod: vc.Proof.VerificationMethod,
			ProofPurpose:       vc.Proof.ProofPurpose,
			ProofValue:         vc.Proof.ProofValue,
		}
	}
	return ext
}

// toExternalClaims converts in-repo CredentialClaims to the external type.
func toExternalClaims(c CredentialClaims) anchor.CredentialClaims {
	types := append([]string{"VerifiableCredential"}, c.Types...)
	return anchor.CredentialClaims{
		ID:             c.ID,
		Type:           types,
		Subject:        c.Subject,
		ExpirationDate: c.ExpirationDate,
	}
}

// fromExternalVerification converts external CredentialVerification to in-repo VerificationResult.
func fromExternalVerification(ext *anchor.CredentialVerification, issuer string) *VerificationResult {
	msg := "valid"
	if !ext.Valid {
		switch {
		case !ext.SignatureValid:
			msg = "signature verification failed"
		case !ext.NotExpired:
			msg = "credential has expired"
		case !ext.NotRevoked:
			msg = "credential has been revoked"
		case !ext.IssuerResolved:
			msg = "issuer DID not resolvable"
		default:
			msg = "verification failed"
		}
	}
	return &VerificationResult{Valid: ext.Valid, Issuer: issuer, Message: msg}
}
