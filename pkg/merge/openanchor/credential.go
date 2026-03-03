package openanchor

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// IssueCredentialLocal creates a Verifiable Credential signed with Ed25519.
func IssueCredentialLocal(claims CredentialClaims, issuerDID string, issuerKey ed25519.PrivateKey) (*VerifiableCredential, error) {
	types := append([]string{"VerifiableCredential"}, claims.Types...)

	vc := &VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID:                claims.ID,
		Type:              types,
		Issuer:            issuerDID,
		IssuanceDate:      time.Now().UTC().Format(time.RFC3339),
		ExpirationDate:    claims.ExpirationDate,
		CredentialSubject: claims.Subject,
	}

	// Canonicalize and sign the VC without proof.
	payload, err := canonicalize(vc)
	if err != nil {
		return nil, fmt.Errorf("canonicalize: %w", err)
	}

	hash := sha256.Sum256(payload)
	sig := ed25519.Sign(issuerKey, hash[:])

	// Resolve the verification method ID from the issuer DID.
	doc, err := ResolveDIDKey(issuerDID)
	if err != nil {
		return nil, fmt.Errorf("resolve issuer DID: %w", err)
	}
	if len(doc.VerificationMethod) == 0 {
		return nil, fmt.Errorf("issuer DID has no verification methods")
	}

	vc.Proof = &CredentialProof{
		Type:               "Ed25519Signature2020",
		Created:            time.Now().UTC().Format(time.RFC3339),
		VerificationMethod: doc.VerificationMethod[0].ID,
		ProofPurpose:       "assertionMethod",
		ProofValue:         Base58Encode(sig),
	}

	return vc, nil
}

// VerifyCredentialLocal verifies a VC's Ed25519 signature.
func VerifyCredentialLocal(vc *VerifiableCredential) (*VerificationResult, error) {
	if vc.Proof == nil {
		return &VerificationResult{Valid: false, Message: "no proof attached"}, nil
	}

	// Check expiry.
	if vc.ExpirationDate != "" {
		exp, err := time.Parse(time.RFC3339, vc.ExpirationDate)
		if err == nil && time.Now().After(exp) {
			return &VerificationResult{Valid: false, Issuer: vc.Issuer, Message: "credential has expired"}, nil
		}
	}

	// Extract public key from issuer DID.
	pub, err := ExtractPublicKey(vc.Issuer)
	if err != nil {
		return &VerificationResult{Valid: false, Message: fmt.Sprintf("cannot resolve issuer: %v", err)}, nil
	}

	// Decode signature.
	sig, err := Base58Decode(vc.Proof.ProofValue)
	if err != nil {
		return &VerificationResult{Valid: false, Message: fmt.Sprintf("invalid proof value: %v", err)}, nil
	}

	// Recreate the payload (VC without proof) and verify.
	vcCopy := *vc
	vcCopy.Proof = nil
	payload, err := canonicalize(&vcCopy)
	if err != nil {
		return &VerificationResult{Valid: false, Message: fmt.Sprintf("canonicalize: %v", err)}, nil
	}

	hash := sha256.Sum256(payload)
	valid := ed25519.Verify(pub, hash[:], sig)

	if !valid {
		return &VerificationResult{Valid: false, Issuer: vc.Issuer, Message: "signature verification failed"}, nil
	}

	return &VerificationResult{Valid: true, Issuer: vc.Issuer, Message: "valid"}, nil
}

// canonicalize produces a deterministic JSON representation for signing.
// Keys are sorted, no proof field included.
func canonicalize(vc *VerifiableCredential) ([]byte, error) {
	// Marshal to map for deterministic key ordering.
	data, err := json.Marshal(vc)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	// Remove proof if present.
	delete(m, "proof")

	// Remove empty optional fields.
	if exp, ok := m["expirationDate"]; ok {
		if s, ok := exp.(string); ok && s == "" {
			delete(m, "expirationDate")
		}
	}

	return marshalSorted(m)
}

// marshalSorted marshals a map with sorted keys recursively.
func marshalSorted(v any) ([]byte, error) {
	switch val := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		out := []byte("{")
		for i, k := range keys {
			if i > 0 {
				out = append(out, ',')
			}
			keyJSON, _ := json.Marshal(k)
			out = append(out, keyJSON...)
			out = append(out, ':')
			valJSON, err := marshalSorted(val[k])
			if err != nil {
				return nil, err
			}
			out = append(out, valJSON...)
		}
		out = append(out, '}')
		return out, nil

	case []any:
		out := []byte("[")
		for i, item := range val {
			if i > 0 {
				out = append(out, ',')
			}
			itemJSON, err := marshalSorted(item)
			if err != nil {
				return nil, err
			}
			out = append(out, itemJSON...)
		}
		out = append(out, ']')
		return out, nil

	default:
		return json.Marshal(v)
	}
}
