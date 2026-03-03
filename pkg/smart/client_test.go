package smart

import (
	"strings"
	"testing"
)

func TestGenerateClientID(t *testing.T) {
	id1 := GenerateClientID()
	id2 := GenerateClientID()
	if id1 == "" {
		t.Fatal("GenerateClientID() returned empty string")
	}
	if id1 == id2 {
		t.Fatal("GenerateClientID() returned duplicate IDs")
	}
	// Should be UUID format (36 chars with dashes).
	if len(id1) != 36 {
		t.Errorf("GenerateClientID() length = %d, want 36", len(id1))
	}
}

func TestGenerateClientSecret(t *testing.T) {
	s1 := GenerateClientSecret()
	s2 := GenerateClientSecret()
	if s1 == "" {
		t.Fatal("GenerateClientSecret() returned empty string")
	}
	if s1 == s2 {
		t.Fatal("GenerateClientSecret() returned duplicate secrets")
	}
	// 32 bytes → 43 chars base64url (no padding).
	if len(s1) != 43 {
		t.Errorf("GenerateClientSecret() length = %d, want 43", len(s1))
	}
}

func TestValidateRedirectURI(t *testing.T) {
	tests := []struct {
		uri     string
		wantErr bool
	}{
		{"https://example.com/callback", false},
		{"http://localhost:3000/callback", false},
		{"http://127.0.0.1:8080/cb", false},
		{"myapp://callback", false},                // custom scheme OK
		{"http://example.com/callback", true},       // non-localhost http
		{"", true},                                   // empty
	}
	for _, tc := range tests {
		err := ValidateRedirectURI(tc.uri)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateRedirectURI(%q) error=%v, wantErr=%v", tc.uri, err, tc.wantErr)
		}
	}
}

func TestValidateClient_Valid(t *testing.T) {
	c := &Client{
		ClientName:              "Test App",
		RedirectURIs:            []string{"https://example.com/cb"},
		Scope:                   "launch patient/Patient.r",
		GrantTypes:              []string{"authorization_code"},
		TokenEndpointAuthMethod: "none",
		LaunchModes:             []string{"standalone"},
	}
	if err := ValidateClient(c); err != nil {
		t.Fatalf("ValidateClient() unexpected error: %v", err)
	}
}

func TestValidateClient_MissingFields(t *testing.T) {
	tests := []struct {
		name    string
		client  Client
		errText string
	}{
		{
			name:    "missing name",
			client:  Client{RedirectURIs: []string{"https://x.com/cb"}, Scope: "launch", GrantTypes: []string{"authorization_code"}},
			errText: "client_name",
		},
		{
			name:    "empty redirects",
			client:  Client{ClientName: "X", Scope: "launch", GrantTypes: []string{"authorization_code"}},
			errText: "redirect_uri",
		},
		{
			name:    "missing scope",
			client:  Client{ClientName: "X", RedirectURIs: []string{"https://x.com/cb"}, GrantTypes: []string{"authorization_code"}},
			errText: "scope",
		},
		{
			name:    "bad grant type",
			client:  Client{ClientName: "X", RedirectURIs: []string{"https://x.com/cb"}, Scope: "launch", GrantTypes: []string{"implicit"}},
			errText: "grant_type",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateClient(&tc.client)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.errText) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.errText)
			}
		})
	}
}
