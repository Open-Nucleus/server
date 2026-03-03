package smart

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// ClientStatus represents the lifecycle state of a registered SMART client.
type ClientStatus string

const (
	ClientPending  ClientStatus = "pending"
	ClientApproved ClientStatus = "approved"
	ClientRevoked  ClientStatus = "revoked"
)

// Client represents a registered SMART on FHIR client application.
type Client struct {
	ClientID                string       `json:"client_id"`
	ClientSecret            string       `json:"client_secret,omitempty"` // empty for public clients
	ClientName              string       `json:"client_name"`
	RedirectURIs            []string     `json:"redirect_uris"`
	Scope                   string       `json:"scope"`                       // max allowed scopes
	GrantTypes              []string     `json:"grant_types"`                 // authorization_code, refresh_token
	TokenEndpointAuthMethod string       `json:"token_endpoint_auth_method"`  // none, client_secret_basic
	LaunchModes             []string     `json:"launch_modes"`                // ehr, standalone
	Status                  ClientStatus `json:"status"`
	RegisteredAt            string       `json:"registered_at"`
	RegisteredBy            string       `json:"registered_by"`               // device_id or "dynamic"
	ApprovedBy              string       `json:"approved_by,omitempty"`
	ApprovedAt              string       `json:"approved_at,omitempty"`
}

// GenerateClientID returns a new random UUID client identifier.
func GenerateClientID() string {
	return uuid.New().String()
}

// GenerateClientSecret returns a cryptographically random 32-byte base64url-encoded secret.
func GenerateClientSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ValidateRedirectURI checks that a redirect URI is acceptable.
// Allowed: https://, http://localhost, http://127.0.0.1, custom scheme (e.g. myapp://).
func ValidateRedirectURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("redirect_uri is required")
	}

	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid redirect_uri %q: %w", uri, err)
	}

	switch u.Scheme {
	case "https":
		return nil
	case "http":
		host := strings.Split(u.Host, ":")[0] // strip port
		if host == "localhost" || host == "127.0.0.1" {
			return nil
		}
		return fmt.Errorf("http redirect_uri only allowed for localhost, got %q", u.Host)
	case "":
		return fmt.Errorf("redirect_uri must have a scheme")
	default:
		// Custom scheme (e.g. myapp://callback) — allowed for native apps.
		return nil
	}
}

// ValidateClient validates a Client struct for required fields and scope correctness.
func ValidateClient(c *Client) error {
	if c.ClientName == "" {
		return fmt.Errorf("client_name is required")
	}
	if len(c.RedirectURIs) == 0 {
		return fmt.Errorf("at least one redirect_uri is required")
	}
	for _, uri := range c.RedirectURIs {
		if err := ValidateRedirectURI(uri); err != nil {
			return err
		}
	}
	if c.Scope == "" {
		return fmt.Errorf("scope is required")
	}

	// Validate each scope in the requested scope string.
	_, err := ParseScopes(c.Scope)
	if err != nil {
		return fmt.Errorf("invalid scope: %w", err)
	}

	// Validate grant types.
	if len(c.GrantTypes) == 0 {
		return fmt.Errorf("at least one grant_type is required")
	}
	for _, gt := range c.GrantTypes {
		if gt != "authorization_code" && gt != "refresh_token" {
			return fmt.Errorf("unsupported grant_type %q", gt)
		}
	}

	// Validate auth method.
	if c.TokenEndpointAuthMethod == "" {
		c.TokenEndpointAuthMethod = "none"
	}
	if c.TokenEndpointAuthMethod != "none" && c.TokenEndpointAuthMethod != "client_secret_basic" {
		return fmt.Errorf("unsupported token_endpoint_auth_method %q", c.TokenEndpointAuthMethod)
	}

	// Validate launch modes.
	for _, lm := range c.LaunchModes {
		if lm != "ehr" && lm != "standalone" {
			return fmt.Errorf("unsupported launch_mode %q", lm)
		}
	}

	return nil
}
