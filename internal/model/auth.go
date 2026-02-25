package model

import "github.com/golang-jwt/jwt/v5"

// NucleusClaims represents the JWT claims per spec section 2.3.
type NucleusClaims struct {
	jwt.RegisteredClaims
	Node        string   `json:"node"`
	Site        string   `json:"site"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
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
