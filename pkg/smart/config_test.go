package smart

import (
	"testing"
)

func TestGenerateSmartConfiguration(t *testing.T) {
	cfg := GenerateSmartConfiguration("http://localhost:8080")

	if cfg.Issuer != "http://localhost:8080" {
		t.Errorf("Issuer = %q, want %q", cfg.Issuer, "http://localhost:8080")
	}
	if cfg.AuthorizationEndpoint != "http://localhost:8080/auth/smart/authorize" {
		t.Errorf("AuthorizationEndpoint = %q", cfg.AuthorizationEndpoint)
	}
	if cfg.TokenEndpoint != "http://localhost:8080/auth/smart/token" {
		t.Errorf("TokenEndpoint = %q", cfg.TokenEndpoint)
	}
	if len(cfg.ScopesSupported) == 0 {
		t.Error("ScopesSupported is empty")
	}
	if len(cfg.ResponseTypesSupported) != 1 || cfg.ResponseTypesSupported[0] != "code" {
		t.Errorf("ResponseTypesSupported = %v", cfg.ResponseTypesSupported)
	}
	if len(cfg.CodeChallengeMethodsSupported) != 1 || cfg.CodeChallengeMethodsSupported[0] != "S256" {
		t.Errorf("CodeChallengeMethodsSupported = %v", cfg.CodeChallengeMethodsSupported)
	}
}

func TestGenerateSmartConfiguration_Capabilities(t *testing.T) {
	cfg := GenerateSmartConfiguration("http://localhost:8080")

	expected := []string{
		"launch-ehr",
		"launch-standalone",
		"client-public",
		"client-confidential-symmetric",
		"permission-v2",
		"context-ehr-patient",
		"context-ehr-encounter",
		"context-standalone-patient",
	}

	if len(cfg.Capabilities) != len(expected) {
		t.Fatalf("Capabilities length = %d, want %d", len(cfg.Capabilities), len(expected))
	}

	capSet := make(map[string]bool)
	for _, c := range cfg.Capabilities {
		capSet[c] = true
	}

	for _, e := range expected {
		if !capSet[e] {
			t.Errorf("missing capability %q", e)
		}
	}
}
