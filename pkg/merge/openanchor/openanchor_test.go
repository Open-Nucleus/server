package openanchor

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

// --- Merkle tests ---

func TestMerkle_SingleFile(t *testing.T) {
	m := NewMerkleTree()
	hash := sha256.Sum256([]byte("hello"))
	root, err := m.ComputeRoot([]FileEntry{{Path: "a.json", Hash: hash[:]}})
	if err != nil {
		t.Fatalf("ComputeRoot: %v", err)
	}
	if len(root) != 32 {
		t.Errorf("expected 32-byte root, got %d", len(root))
	}
}

func TestMerkle_Empty(t *testing.T) {
	m := NewMerkleTree()
	_, err := m.ComputeRoot(nil)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestMerkle_Deterministic(t *testing.T) {
	m := NewMerkleTree()
	entries := make([]FileEntry, 10)
	for i := range entries {
		h := sha256.Sum256([]byte{byte(i)})
		entries[i] = FileEntry{Path: string(rune('a'+i)) + ".json", Hash: h[:]}
	}

	root1, err := m.ComputeRoot(entries)
	if err != nil {
		t.Fatalf("root1: %v", err)
	}

	// Shuffle entries — should produce same root due to sorting.
	shuffled := make([]FileEntry, len(entries))
	copy(shuffled, entries)
	shuffled[0], shuffled[9] = shuffled[9], shuffled[0]
	shuffled[3], shuffled[7] = shuffled[7], shuffled[3]

	root2, err := m.ComputeRoot(shuffled)
	if err != nil {
		t.Fatalf("root2: %v", err)
	}

	if !bytes.Equal(root1, root2) {
		t.Errorf("roots differ after shuffle:\n  %s\n  %s", hex.EncodeToString(root1), hex.EncodeToString(root2))
	}
}

func TestMerkle_TenFiles(t *testing.T) {
	m := NewMerkleTree()
	entries := make([]FileEntry, 10)
	for i := range entries {
		h := sha256.Sum256([]byte{byte(i)})
		entries[i] = FileEntry{Path: string(rune('a'+i)) + ".json", Hash: h[:]}
	}
	root, err := m.ComputeRoot(entries)
	if err != nil {
		t.Fatalf("ComputeRoot: %v", err)
	}
	if len(root) != 32 {
		t.Errorf("expected 32-byte root, got %d", len(root))
	}
}

// --- Base58 tests ---

func TestBase58_Roundtrip(t *testing.T) {
	data := []byte{0xed, 0x01, 0xaa, 0xbb, 0xcc, 0xdd}
	encoded := Base58Encode(data)
	decoded, err := Base58Decode(encoded)
	if err != nil {
		t.Fatalf("Base58Decode: %v", err)
	}
	if !bytes.Equal(data, decoded) {
		t.Errorf("roundtrip failed: got %x, want %x", decoded, data)
	}
}

func TestBase58_LeadingZeros(t *testing.T) {
	data := []byte{0x00, 0x00, 0x01}
	encoded := Base58Encode(data)
	if !strings.HasPrefix(encoded, "11") {
		t.Errorf("expected leading '11' for two zero bytes, got %s", encoded)
	}
	decoded, err := Base58Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !bytes.Equal(data, decoded) {
		t.Errorf("roundtrip: got %x, want %x", decoded, data)
	}
}

func TestBase58_InvalidChar(t *testing.T) {
	_, err := Base58Decode("0OIl") // 0, O, I, l are not in base58
	if err == nil {
		t.Fatal("expected error for invalid characters")
	}
}

// --- DID:key tests ---

func TestDIDKey_GenerateAndResolve(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	did, doc, err := DIDKeyFromEd25519(pub)
	if err != nil {
		t.Fatalf("DIDKeyFromEd25519: %v", err)
	}

	if !strings.HasPrefix(did, "did:key:z") {
		t.Errorf("expected did:key:z..., got %s", did)
	}
	if doc.ID != did {
		t.Errorf("doc.ID != did: %s != %s", doc.ID, did)
	}
	if len(doc.VerificationMethod) != 1 {
		t.Fatalf("expected 1 verification method, got %d", len(doc.VerificationMethod))
	}
	if doc.VerificationMethod[0].Type != "Ed25519VerificationKey2020" {
		t.Errorf("unexpected type: %s", doc.VerificationMethod[0].Type)
	}

	// Resolve back.
	resolved, err := ResolveDIDKey(did)
	if err != nil {
		t.Fatalf("ResolveDIDKey: %v", err)
	}
	if resolved.ID != did {
		t.Errorf("resolved ID mismatch: %s != %s", resolved.ID, did)
	}

	// Extract key and compare.
	extractedPub, err := ExtractPublicKey(did)
	if err != nil {
		t.Fatalf("ExtractPublicKey: %v", err)
	}
	if !bytes.Equal(pub, extractedPub) {
		t.Error("extracted public key does not match original")
	}
}

func TestDIDKey_InvalidFormat(t *testing.T) {
	_, err := ResolveDIDKey("did:web:example.com")
	if err == nil {
		t.Fatal("expected error for non did:key")
	}
}

// --- Verifiable Credential tests ---

func TestVC_IssueAndVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}

	did, _, err := DIDKeyFromEd25519(pub)
	if err != nil {
		t.Fatalf("DID: %v", err)
	}

	claims := CredentialClaims{
		ID:    "urn:uuid:test-credential-001",
		Types: []string{"DataIntegrityCredential"},
		Subject: map[string]any{
			"merkleRoot": "abc123",
			"gitHead":    "def456",
		},
	}

	vc, err := IssueCredentialLocal(claims, did, priv)
	if err != nil {
		t.Fatalf("IssueCredentialLocal: %v", err)
	}

	if vc.Issuer != did {
		t.Errorf("issuer mismatch: %s", vc.Issuer)
	}
	if vc.Proof == nil {
		t.Fatal("expected proof")
	}
	if vc.Proof.Type != "Ed25519Signature2020" {
		t.Errorf("unexpected proof type: %s", vc.Proof.Type)
	}

	// Verify.
	result, err := VerifyCredentialLocal(vc)
	if err != nil {
		t.Fatalf("VerifyCredentialLocal: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid, got: %s", result.Message)
	}
}

func TestVC_TamperDetection(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	did, _, _ := DIDKeyFromEd25519(pub)

	claims := CredentialClaims{
		ID:    "urn:uuid:tamper-test",
		Types: []string{"DataIntegrityCredential"},
		Subject: map[string]any{
			"merkleRoot": "original",
		},
	}

	vc, err := IssueCredentialLocal(claims, did, priv)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	// Tamper with the subject.
	vc.CredentialSubject["merkleRoot"] = "tampered"

	result, err := VerifyCredentialLocal(vc)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if result.Valid {
		t.Error("expected invalid after tampering")
	}
}

// --- StubBackend tests ---

func TestStubBackend(t *testing.T) {
	be := NewStubBackend()
	if be.Available() {
		t.Error("expected Available()=false")
	}
	if be.Name() != "none" {
		t.Errorf("expected name 'none', got %s", be.Name())
	}
	_, err := be.Anchor(nil, AnchorMetadata{})
	if err != ErrBackendNotConfigured {
		t.Errorf("expected ErrBackendNotConfigured, got %v", err)
	}
}

// --- LocalIdentityEngine tests ---

func TestLocalIdentityEngine_RoundTrip(t *testing.T) {
	eng := NewLocalIdentityEngine()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)

	doc, err := eng.GenerateDID(pub)
	if err != nil {
		t.Fatalf("GenerateDID: %v", err)
	}
	if !strings.HasPrefix(doc.ID, "did:key:z") {
		t.Errorf("unexpected DID: %s", doc.ID)
	}

	resolved, err := eng.ResolveDID(doc.ID)
	if err != nil {
		t.Fatalf("ResolveDID: %v", err)
	}
	if resolved.ID != doc.ID {
		t.Error("resolved DID mismatch")
	}

	claims := CredentialClaims{
		ID:      "urn:uuid:engine-test",
		Types:   []string{"TestCredential"},
		Subject: map[string]any{"test": true},
	}

	vc, err := eng.IssueCredential(claims, doc.ID, priv)
	if err != nil {
		t.Fatalf("IssueCredential: %v", err)
	}

	result, err := eng.VerifyCredential(vc)
	if err != nil {
		t.Fatalf("VerifyCredential: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid: %s", result.Message)
	}
}
