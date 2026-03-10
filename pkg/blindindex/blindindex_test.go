package blindindex

import (
	"testing"
)

func testKey() []byte {
	return []byte("0123456789abcdef0123456789abcdef") // 32 bytes
}

func TestNewIndexer(t *testing.T) {
	idx, err := NewIndexer(testKey())
	if err != nil {
		t.Fatalf("NewIndexer: %v", err)
	}
	if idx == nil {
		t.Fatal("expected non-nil indexer")
	}
}

func TestNewIndexer_EmptyKey(t *testing.T) {
	_, err := NewIndexer([]byte{})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestBlindExact_Deterministic(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	h1 := idx.BlindExact("Smith")
	h2 := idx.BlindExact("Smith")
	if h1 != h2 {
		t.Fatalf("expected same hash, got %s vs %s", h1, h2)
	}
}

func TestBlindExact_CaseInsensitive(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	h1 := idx.BlindExact("Smith")
	h2 := idx.BlindExact("smith")
	h3 := idx.BlindExact("SMITH")
	if h1 != h2 || h2 != h3 {
		t.Fatalf("expected case-insensitive match: %s %s %s", h1, h2, h3)
	}
}

func TestBlindExact_TrimWhitespace(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	h1 := idx.BlindExact("Smith")
	h2 := idx.BlindExact("  Smith  ")
	if h1 != h2 {
		t.Fatalf("expected whitespace-trimmed match: %s vs %s", h1, h2)
	}
}

func TestBlindExact_DifferentValues(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	h1 := idx.BlindExact("Smith")
	h2 := idx.BlindExact("Jones")
	if h1 == h2 {
		t.Fatal("different values should produce different hashes")
	}
}

func TestBlindNgram(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	ngrams := idx.BlindNgram("Smith", 3)
	// "smith" has trigrams: "smi", "mit", "ith" → 3 unique n-grams
	if len(ngrams) != 3 {
		t.Fatalf("expected 3 n-grams for 'Smith', got %d", len(ngrams))
	}
}

func TestBlindNgram_ShortValue(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	ngrams := idx.BlindNgram("ab", 3)
	// Value shorter than n-gram size → hash whole value
	if len(ngrams) != 1 {
		t.Fatalf("expected 1 n-gram for short value, got %d", len(ngrams))
	}
}

func TestBlindNgram_Overlap(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	// "smith" trigrams include "smi"
	ngrams := idx.BlindNgram("Smith", 3)
	query := idx.BlindNgram("Smi", 3)

	// The query "Smi" should produce 1 trigram that matches one of the indexed trigrams
	if len(query) != 1 {
		t.Fatalf("expected 1 n-gram for query 'Smi', got %d", len(query))
	}

	found := false
	for _, qh := range query {
		for _, nh := range ngrams {
			if qh == nh {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("query 'Smi' should match at least one n-gram from 'Smith'")
	}
}

func TestBlindDatePrefix(t *testing.T) {
	idx, _ := NewIndexer(testKey())
	h1 := idx.BlindDatePrefix("1990-03-15")
	h2 := idx.BlindDatePrefix("1990-03-20")
	// Same YYYY-MM prefix → same hash
	if h1 != h2 {
		t.Fatalf("expected same month prefix hash: %s vs %s", h1, h2)
	}

	h3 := idx.BlindDatePrefix("1990-04-15")
	if h1 == h3 {
		t.Fatal("different months should produce different hashes")
	}
}

func TestBlindExact_DifferentKeys(t *testing.T) {
	idx1, _ := NewIndexer([]byte("aaaabbbbccccddddaaaabbbbccccdddd"))
	idx2, _ := NewIndexer([]byte("eeeeffff00001111eeeeffff00001111"))
	h1 := idx1.BlindExact("Smith")
	h2 := idx2.BlindExact("Smith")
	if h1 == h2 {
		t.Fatal("different keys should produce different hashes")
	}
}
