// Package crypto provides shared cryptographic utilities for Open Nucleus.
package crypto

import (
	"crypto/ed25519"
	"crypto/sha512"
	"errors"
	"math/big"
)

// EdPrivateToX25519 converts an Ed25519 private key to an X25519 private key.
//
// Per RFC 8032 Section 5.1.5, the Ed25519 private scalar is computed by hashing
// the 32-byte seed with SHA-512, then clamping bits in the lower 32 bytes.
func EdPrivateToX25519(edPriv ed25519.PrivateKey) []byte {
	seed := edPriv.Seed()
	h := sha512.Sum512(seed)
	h[0] &= 248
	h[31] &= 127
	h[31] |= 64
	x25519Key := make([]byte, 32)
	copy(x25519Key, h[:32])
	return x25519Key
}

// EdPublicToX25519 converts an Ed25519 public key (Edwards y-coordinate) to an
// X25519 public key (Montgomery u-coordinate).
//
// The conversion formula is: u = (1 + y) / (1 - y) mod p
// where p = 2^255 - 19.
func EdPublicToX25519(edPub ed25519.PublicKey) ([]byte, error) {
	yBytes := make([]byte, 32)
	copy(yBytes, edPub)
	yBytes[31] &= 0x7f

	reversed := make([]byte, 32)
	for i := 0; i < 32; i++ {
		reversed[i] = yBytes[31-i]
	}
	y := new(big.Int).SetBytes(reversed)

	p := new(big.Int).SetBit(new(big.Int), 255, 1)
	p.Sub(p, big.NewInt(19))

	one := big.NewInt(1)
	numerator := new(big.Int).Add(one, y)
	numerator.Mod(numerator, p)

	denominator := new(big.Int).Sub(one, y)
	denominator.Mod(denominator, p)

	if denominator.Sign() == 0 {
		return nil, errors.New("crypto: degenerate public key (y == 1)")
	}

	denominatorInv := new(big.Int).ModInverse(denominator, p)
	if denominatorInv == nil {
		return nil, errors.New("crypto: failed to compute modular inverse")
	}

	u := new(big.Int).Mul(numerator, denominatorInv)
	u.Mod(u, p)

	uBytes := u.Bytes()
	result := make([]byte, 32)
	for i, b := range uBytes {
		result[len(uBytes)-1-i] = b
	}

	return result, nil
}
