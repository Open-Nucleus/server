package openanchor

import (
	"crypto/ed25519"
	"fmt"
	"strings"
	"time"
)

// Multicodec prefix for Ed25519 public keys.
var ed25519MulticodecPrefix = []byte{0xed, 0x01}

// DIDKeyFromEd25519 generates a did:key from an Ed25519 public key.
// Format: did:key:z<base58btc(multicodec_prefix + pubkey)>
func DIDKeyFromEd25519(pub ed25519.PublicKey) (string, *DIDDocument, error) {
	if len(pub) != ed25519.PublicKeySize {
		return "", nil, fmt.Errorf("invalid Ed25519 public key size: %d", len(pub))
	}

	// Multicodec encode: 0xed01 prefix + raw public key bytes.
	multicodecBytes := make([]byte, 0, len(ed25519MulticodecPrefix)+len(pub))
	multicodecBytes = append(multicodecBytes, ed25519MulticodecPrefix...)
	multicodecBytes = append(multicodecBytes, pub...)

	// Base58btc encode with 'z' multibase prefix.
	encoded := "z" + Base58Encode(multicodecBytes)
	did := "did:key:" + encoded

	doc := buildDIDDocument(did, encoded, pub)
	return did, doc, nil
}

// ResolveDIDKey parses a did:key string and returns the DIDDocument.
// Only supports did:key with Ed25519 keys.
func ResolveDIDKey(did string) (*DIDDocument, error) {
	if !strings.HasPrefix(did, "did:key:z") {
		return nil, fmt.Errorf("%w: expected did:key:z..., got %s", ErrInvalidDID, did)
	}

	// Extract multibase value (after "did:key:z").
	multibaseValue := did[len("did:key:z"):]
	if multibaseValue == "" {
		return nil, fmt.Errorf("%w: empty key value", ErrInvalidDID)
	}

	decoded, err := Base58Decode(multibaseValue)
	if err != nil {
		return nil, fmt.Errorf("%w: base58 decode: %v", ErrInvalidDID, err)
	}

	// Verify multicodec prefix.
	if len(decoded) < 2 || decoded[0] != 0xed || decoded[1] != 0x01 {
		return nil, fmt.Errorf("%w: not an Ed25519 multicodec key", ErrUnsupportedDIDMethod)
	}

	pubBytes := decoded[2:]
	if len(pubBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("%w: invalid key length %d", ErrInvalidDID, len(pubBytes))
	}

	pub := ed25519.PublicKey(pubBytes)
	encoded := "z" + multibaseValue
	doc := buildDIDDocument(did, encoded, pub)
	return doc, nil
}

// ExtractPublicKey extracts the Ed25519 public key from a did:key string.
func ExtractPublicKey(did string) (ed25519.PublicKey, error) {
	doc, err := ResolveDIDKey(did)
	if err != nil {
		return nil, err
	}
	if len(doc.VerificationMethod) == 0 {
		return nil, fmt.Errorf("%w: no verification method", ErrInvalidDID)
	}

	mb := doc.VerificationMethod[0].PublicKeyMultibase
	if !strings.HasPrefix(mb, "z") {
		return nil, fmt.Errorf("%w: invalid multibase prefix", ErrInvalidDID)
	}

	decoded, err := Base58Decode(mb[1:])
	if err != nil {
		return nil, fmt.Errorf("%w: base58 decode: %v", ErrInvalidDID, err)
	}

	if len(decoded) < 2 || decoded[0] != 0xed || decoded[1] != 0x01 {
		return nil, fmt.Errorf("%w: not Ed25519", ErrUnsupportedDIDMethod)
	}

	return ed25519.PublicKey(decoded[2:]), nil
}

func buildDIDDocument(did, multibaseKey string, pub ed25519.PublicKey) *DIDDocument {
	vmID := did + "#" + multibaseKey
	return &DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID: did,
		VerificationMethod: []VerificationMethod{
			{
				ID:                 vmID,
				Type:               "Ed25519VerificationKey2020",
				Controller:         did,
				PublicKeyMultibase: multibaseKey,
			},
		},
		Authentication:  []string{vmID},
		AssertionMethod: []string{vmID},
		Created:         time.Now().UTC().Format(time.RFC3339),
	}
}
