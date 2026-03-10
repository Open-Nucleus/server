// Package blindindex provides HMAC-based blind indexing for PII fields.
//
// Instead of storing plaintext names or dates in SQLite, we store HMAC-SHA256
// digests. Searches work by computing the same HMAC on the query and matching
// against stored values. This allows equality and n-gram substring searches
// without exposing PII if the database is exfiltrated.
package blindindex

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"unicode"

	"golang.org/x/crypto/hkdf"
	"golang.org/x/text/unicode/norm"
)

const (
	// hkdfSalt is the domain-separation salt for deriving the HMAC key.
	hkdfSalt = "open-nucleus-blind-v1"
	// hkdfInfo is the HKDF info string.
	hkdfInfo = "hmac-key"
	// hmacKeySize is the derived HMAC key size.
	hmacKeySize = 32
)

// Indexer computes blind indexes from plaintext values.
type Indexer struct {
	hmacKey []byte
}

// NewIndexer derives an HMAC key from the master key using HKDF-SHA256.
func NewIndexer(masterKey []byte) (*Indexer, error) {
	if len(masterKey) == 0 {
		return nil, fmt.Errorf("blindindex: empty master key")
	}
	hkdfReader := hkdf.New(sha256.New, masterKey, []byte(hkdfSalt), []byte(hkdfInfo))
	hmacKey := make([]byte, hmacKeySize)
	if _, err := io.ReadFull(hkdfReader, hmacKey); err != nil {
		return nil, fmt.Errorf("blindindex: HKDF derive: %w", err)
	}
	return &Indexer{hmacKey: hmacKey}, nil
}

// BlindExact computes HMAC-SHA256(key, normalize(value)) and returns a hex string.
// Used for exact-match lookups (e.g., family name equality).
func (idx *Indexer) BlindExact(value string) string {
	normalized := normalize(value)
	mac := hmac.New(sha256.New, idx.hmacKey)
	mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil))
}

// BlindNgram computes blind indexes for each n-gram (sliding window) of the
// normalized value. Returns a sorted, deduplicated set of hex HMAC strings.
// Used for substring/prefix search via the patients_ngrams table.
func (idx *Indexer) BlindNgram(value string, n int) []string {
	normalized := normalize(value)
	if len(normalized) < n {
		// Value shorter than n-gram size: hash the whole value as single n-gram
		return []string{idx.hmacNgram(normalized)}
	}

	seen := make(map[string]struct{})
	var result []string
	runes := []rune(normalized)
	for i := 0; i <= len(runes)-n; i++ {
		gram := string(runes[i : i+n])
		hash := idx.hmacNgram(gram)
		if _, ok := seen[hash]; !ok {
			seen[hash] = struct{}{}
			result = append(result, hash)
		}
	}
	return result
}

// BlindDatePrefix blinds the YYYY-MM prefix of a date for range-style queries.
// Input should be an ISO date like "1990-03-15"; output is HMAC("1990-03").
func (idx *Indexer) BlindDatePrefix(date string) string {
	prefix := date
	if len(date) >= 7 {
		prefix = date[:7] // "YYYY-MM"
	}
	return idx.BlindExact(prefix)
}

// hmacNgram computes HMAC for a single n-gram with a domain separator.
func (idx *Indexer) hmacNgram(gram string) string {
	mac := hmac.New(sha256.New, idx.hmacKey)
	mac.Write([]byte("ngram:"))
	mac.Write([]byte(gram))
	return hex.EncodeToString(mac.Sum(nil))
}

// normalize lowercases, trims whitespace, and applies NFC unicode normalization.
func normalize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Map(unicode.ToLower, s)
	s = norm.NFC.String(s)
	return s
}
