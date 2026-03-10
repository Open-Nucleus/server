package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/curve25519"
)

func TestEdPrivateToX25519(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	x25519Key := EdPrivateToX25519(priv)
	if len(x25519Key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(x25519Key))
	}

	// Verify clamping: bit 0 of byte 0 should be 0, bits 6-7 of byte 31 should be 01
	if x25519Key[0]&7 != 0 {
		t.Fatal("lowest 3 bits should be cleared")
	}
	if x25519Key[31]&128 != 0 {
		t.Fatal("highest bit should be cleared")
	}
	if x25519Key[31]&64 == 0 {
		t.Fatal("second highest bit should be set")
	}
}

func TestEdPublicToX25519(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	x25519Pub, err := EdPublicToX25519(pub)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if len(x25519Pub) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(x25519Pub))
	}
}

func TestECDH_SharedSecret(t *testing.T) {
	// Generate two Ed25519 keypairs
	_, privA, _ := ed25519.GenerateKey(rand.Reader)
	pubB, privB, _ := ed25519.GenerateKey(rand.Reader)

	// Convert to X25519
	x25519PrivA := EdPrivateToX25519(privA)
	x25519PubB, err := EdPublicToX25519(pubB)
	if err != nil {
		t.Fatalf("convert B public: %v", err)
	}

	x25519PrivB := EdPrivateToX25519(privB)
	x25519PubA, err := EdPublicToX25519(privA.Public().(ed25519.PublicKey))
	if err != nil {
		t.Fatalf("convert A public: %v", err)
	}

	// ECDH: A→B and B→A should produce same shared secret
	sharedAB, err := curve25519.X25519(x25519PrivA, x25519PubB)
	if err != nil {
		t.Fatalf("ECDH A→B: %v", err)
	}

	sharedBA, err := curve25519.X25519(x25519PrivB, x25519PubA)
	if err != nil {
		t.Fatalf("ECDH B→A: %v", err)
	}

	if len(sharedAB) != 32 || len(sharedBA) != 32 {
		t.Fatal("shared secrets should be 32 bytes")
	}

	for i := range sharedAB {
		if sharedAB[i] != sharedBA[i] {
			t.Fatal("shared secrets should match")
		}
	}
}
