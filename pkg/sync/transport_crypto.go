// Package sync provides transport-layer cryptography for node-to-node sync.
//
// Key exchange uses ECDH over X25519 (converted from Ed25519 identity keys),
// with HKDF-SHA256 key derivation and AES-256-GCM authenticated encryption.
// This replaces the previous broken scheme that prepended the AES key to the
// ciphertext (effectively sending plaintext).
package sync

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	// hkdfSalt is the domain-separation salt for HKDF key derivation.
	hkdfSalt = "open-nucleus-sync-v1"
	// hkdfInfo is the HKDF info string identifying the derived key purpose.
	hkdfInfo = "transport-encryption"
	// derivedKeySize is the size of the AES-256 key derived by HKDF.
	derivedKeySize = 32
)

// DeriveSharedKey computes a shared 32-byte AES-256 key from an Ed25519
// private key and a remote peer's Ed25519 public key.
//
// The process:
//  1. Convert Ed25519 private key to X25519 private key (RFC 7748)
//  2. Convert Ed25519 public key to X25519 public key (RFC 7748)
//  3. Perform X25519 ECDH to produce a raw shared secret
//  4. Derive a 32-byte key via HKDF-SHA256 with fixed salt and info strings
//
// Both peers will derive the same key regardless of which is "my" vs "their".
func DeriveSharedKey(myPrivate ed25519.PrivateKey, theirPublic ed25519.PublicKey) ([]byte, error) {
	if len(myPrivate) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("sync/crypto: invalid private key size: %d", len(myPrivate))
	}
	if len(theirPublic) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("sync/crypto: invalid public key size: %d", len(theirPublic))
	}

	// Step 1: Ed25519 private key → X25519 private key
	// The Ed25519 seed (first 32 bytes of the 64-byte private key) is hashed
	// with SHA-512; the first 32 bytes of that hash (after clamping) form the
	// X25519 scalar.
	x25519Private := edPrivateToX25519(myPrivate)

	// Step 2: Ed25519 public key → X25519 public key (Montgomery form)
	x25519Public, err := edPublicToX25519(theirPublic)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: convert their public key: %w", err)
	}

	// Step 3: X25519 ECDH
	sharedSecret, err := curve25519.X25519(x25519Private, x25519Public)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: X25519 ECDH: %w", err)
	}

	// Step 4: HKDF-SHA256 to derive the final 32-byte encryption key
	hkdfReader := hkdf.New(sha256.New, sharedSecret, []byte(hkdfSalt), []byte(hkdfInfo))
	derivedKey := make([]byte, derivedKeySize)
	if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
		return nil, fmt.Errorf("sync/crypto: HKDF expand: %w", err)
	}

	return derivedKey, nil
}

// EncryptPayload encrypts plaintext using AES-256-GCM with the given shared key.
// Output format: [12-byte nonce][ciphertext + 16-byte GCM tag].
func EncryptPayload(sharedKey, plaintext []byte) ([]byte, error) {
	if len(sharedKey) != derivedKeySize {
		return nil, fmt.Errorf("sync/crypto: key must be %d bytes, got %d", derivedKeySize, len(sharedKey))
	}

	block, err := aes.NewCipher(sharedKey)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: new AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: new GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("sync/crypto: generate nonce: %w", err)
	}

	// Seal appends the ciphertext (with tag) to nonce.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// DecryptPayload decrypts ciphertext produced by EncryptPayload.
// Expects format: [12-byte nonce][ciphertext + 16-byte GCM tag].
func DecryptPayload(sharedKey, ciphertext []byte) ([]byte, error) {
	if len(sharedKey) != derivedKeySize {
		return nil, fmt.Errorf("sync/crypto: key must be %d bytes, got %d", derivedKeySize, len(sharedKey))
	}

	block, err := aes.NewCipher(sharedKey)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: new AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: new GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize+gcm.Overhead() {
		return nil, errors.New("sync/crypto: ciphertext too short")
	}

	nonce := ciphertext[:nonceSize]
	sealed := ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("sync/crypto: GCM open failed (wrong key or tampered data): %w", err)
	}

	return plaintext, nil
}

// edPrivateToX25519 converts an Ed25519 private key to an X25519 private key.
//
// Per RFC 8032 Section 5.1.5, the Ed25519 private scalar is computed by hashing
// the 32-byte seed with SHA-512, then clamping bits in the lower 32 bytes.
// Those clamped 32 bytes are the X25519 private key (scalar).
func edPrivateToX25519(edPriv ed25519.PrivateKey) []byte {
	seed := edPriv.Seed() // 32-byte seed
	h := sha512.Sum512(seed)
	// Clamp per RFC 7748 / RFC 8032
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64
	x25519Key := make([]byte, 32)
	copy(x25519Key, h[:32])
	return x25519Key
}

// edPublicToX25519 converts an Ed25519 public key (Edwards y-coordinate) to an
// X25519 public key (Montgomery u-coordinate).
//
// The conversion formula is:  u = (1 + y) / (1 - y)  mod p
// where p = 2^255 - 19 (the field prime for Curve25519).
func edPublicToX25519(edPub ed25519.PublicKey) ([]byte, error) {
	// The Ed25519 public key is a 32-byte compressed Edwards point.
	// The y-coordinate is stored in the lower 255 bits (little-endian),
	// and the top bit of byte 31 is the sign of x.

	// Extract y (clear the sign bit)
	yBytes := make([]byte, 32)
	copy(yBytes, edPub)
	yBytes[31] &= 0x7f // clear sign bit

	// Convert from little-endian to big.Int
	y := new(big.Int)
	// big.Int expects big-endian, so reverse
	reversed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversed[i] = yBytes[31-i]
	}
	y.SetBytes(reversed)

	// p = 2^255 - 19
	p := new(big.Int).SetBit(new(big.Int), 255, 1)
	p.Sub(p, big.NewInt(19))

	// u = (1 + y) * (1 - y)^(-1) mod p
	one := big.NewInt(1)
	numerator := new(big.Int).Add(one, y)
	numerator.Mod(numerator, p)

	denominator := new(big.Int).Sub(one, y)
	denominator.Mod(denominator, p)

	// Check for degenerate case (y == 1 → denominator == 0)
	if denominator.Sign() == 0 {
		return nil, errors.New("sync/crypto: degenerate public key (y == 1)")
	}

	denominatorInv := new(big.Int).ModInverse(denominator, p)
	if denominatorInv == nil {
		return nil, errors.New("sync/crypto: failed to compute modular inverse")
	}

	u := new(big.Int).Mul(numerator, denominatorInv)
	u.Mod(u, p)

	// Convert u back to 32 bytes little-endian
	uBytes := u.Bytes() // big-endian
	result := make([]byte, 32)
	for i, b := range uBytes {
		result[len(uBytes)-1-i] = b
	}

	return result, nil
}
