package model

import "github.com/golang-jwt/jwt/v5"

// NucleusClaims represents the JWT claims per spec section 2.3.
// JSON tags match pkg/auth.NucleusClaims so JWTs from the Auth Service
// deserialize correctly in the gateway middleware.
type NucleusClaims struct {
	jwt.RegisteredClaims
	DeviceID    string   `json:"device_id"`
	Node        string   `json:"node_id"`
	Site        string   `json:"site_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SiteScope   string   `json:"site_scope"`
	TokenType   string   `json:"token_type"`
}

// LoginRequest is the body of POST /auth/login.
type LoginRequest struct {
	DeviceID          string            `json:"device_id"`
	PublicKey         string            `json:"public_key"`
	ChallengeResponse ChallengeResponse `json:"challenge_response"`
	PractitionerID   string            `json:"practitioner_id"`
}

type ChallengeResponse struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
	Timestamp string `json:"timestamp"`
}

// RefreshRequest is the body of POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest is the body of POST /auth/logout.
type LogoutRequest struct {
	Token string `json:"token,omitempty"`
}
