package openanchor

import (
	"fmt"
	"math/big"
)

// Base58btc alphabet (Bitcoin).
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Base58Encode encodes bytes to base58btc string.
func Base58Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Count leading zeros.
	var leadingZeros int
	for _, b := range data {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	// Convert to big.Int and encode.
	n := new(big.Int).SetBytes(data)
	zero := big.NewInt(0)
	base := big.NewInt(58)
	mod := new(big.Int)

	var encoded []byte
	for n.Cmp(zero) > 0 {
		n.DivMod(n, base, mod)
		encoded = append(encoded, base58Alphabet[mod.Int64()])
	}

	// Add leading '1's for each leading zero byte.
	for i := 0; i < leadingZeros; i++ {
		encoded = append(encoded, base58Alphabet[0])
	}

	// Reverse.
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}

	return string(encoded)
}

// Base58Decode decodes a base58btc string to bytes.
func Base58Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return nil, nil
	}

	n := new(big.Int)
	base := big.NewInt(58)

	for _, c := range s {
		idx := -1
		for i, a := range base58Alphabet {
			if a == c {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, fmt.Errorf("invalid base58 character: %c", c)
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(idx)))
	}

	decoded := n.Bytes()

	// Count leading '1's → leading zero bytes.
	var leadingOnes int
	for _, c := range s {
		if c != rune(base58Alphabet[0]) {
			break
		}
		leadingOnes++
	}

	result := make([]byte, leadingOnes+len(decoded))
	copy(result[leadingOnes:], decoded)
	return result, nil
}
