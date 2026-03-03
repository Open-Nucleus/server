package smart

import (
	"testing"
	"time"
)

func TestAuthCodeStore_GenerateAndExchange(t *testing.T) {
	store := NewAuthCodeStore(5 * time.Second)
	defer store.Close()

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GeneratePKCEChallenge(verifier)

	code, err := store.Generate(AuthCodeParams{
		ClientID:       "test-client",
		RedirectURI:    "https://example.com/cb",
		Scope:          "patient/Patient.r",
		CodeChallenge:  challenge,
		DeviceID:       "dev-1",
		PractitionerID: "pract-1",
		SiteID:         "site-1",
		Role:           "physician",
		PatientID:      "patient-123",
	})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if code == "" {
		t.Fatal("Generate() returned empty code")
	}

	ac, err := store.Exchange(code, "test-client", verifier, "https://example.com/cb")
	if err != nil {
		t.Fatalf("Exchange() error: %v", err)
	}
	if ac.Scope != "patient/Patient.r" {
		t.Errorf("Exchange().Scope = %q, want %q", ac.Scope, "patient/Patient.r")
	}
	if ac.PatientID != "patient-123" {
		t.Errorf("Exchange().PatientID = %q, want %q", ac.PatientID, "patient-123")
	}
}

func TestAuthCodeStore_Expired(t *testing.T) {
	store := NewAuthCodeStore(1 * time.Millisecond)
	defer store.Close()

	code, err := store.Generate(AuthCodeParams{
		ClientID:    "test-client",
		RedirectURI: "https://example.com/cb",
		Scope:       "launch",
	})
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	time.Sleep(10 * time.Millisecond) // let it expire

	_, err = store.Exchange(code, "test-client", "", "https://example.com/cb")
	if err == nil {
		t.Fatal("Exchange() expected error for expired code")
	}
}

func TestAuthCodeStore_WrongClient(t *testing.T) {
	store := NewAuthCodeStore(5 * time.Second)
	defer store.Close()

	code, _ := store.Generate(AuthCodeParams{
		ClientID:    "test-client",
		RedirectURI: "https://example.com/cb",
		Scope:       "launch",
	})

	_, err := store.Exchange(code, "wrong-client", "", "https://example.com/cb")
	if err == nil {
		t.Fatal("Exchange() expected error for wrong client_id")
	}
}

func TestAuthCodeStore_WrongRedirectURI(t *testing.T) {
	store := NewAuthCodeStore(5 * time.Second)
	defer store.Close()

	code, _ := store.Generate(AuthCodeParams{
		ClientID:    "test-client",
		RedirectURI: "https://example.com/cb",
		Scope:       "launch",
	})

	_, err := store.Exchange(code, "test-client", "", "https://evil.com/cb")
	if err == nil {
		t.Fatal("Exchange() expected error for wrong redirect_uri")
	}
}

func TestAuthCodeStore_WrongPKCE(t *testing.T) {
	store := NewAuthCodeStore(5 * time.Second)
	defer store.Close()

	challenge := GeneratePKCEChallenge("correct-verifier")

	code, _ := store.Generate(AuthCodeParams{
		ClientID:      "test-client",
		RedirectURI:   "https://example.com/cb",
		Scope:         "launch",
		CodeChallenge: challenge,
	})

	_, err := store.Exchange(code, "test-client", "wrong-verifier", "https://example.com/cb")
	if err == nil {
		t.Fatal("Exchange() expected error for wrong PKCE verifier")
	}
}

func TestAuthCodeStore_OneShot(t *testing.T) {
	store := NewAuthCodeStore(5 * time.Second)
	defer store.Close()

	code, _ := store.Generate(AuthCodeParams{
		ClientID:    "test-client",
		RedirectURI: "https://example.com/cb",
		Scope:       "launch",
	})

	// First exchange should succeed.
	_, err := store.Exchange(code, "test-client", "", "https://example.com/cb")
	if err != nil {
		t.Fatalf("first Exchange() error: %v", err)
	}

	// Second exchange should fail (consumed).
	_, err = store.Exchange(code, "test-client", "", "https://example.com/cb")
	if err == nil {
		t.Fatal("second Exchange() expected error for consumed code")
	}
}

func TestValidatePKCE_S256(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GeneratePKCEChallenge(verifier)

	if !ValidatePKCE(verifier, challenge) {
		t.Error("ValidatePKCE(correct) = false, want true")
	}
	if ValidatePKCE("wrong-verifier", challenge) {
		t.Error("ValidatePKCE(wrong) = true, want false")
	}
}
