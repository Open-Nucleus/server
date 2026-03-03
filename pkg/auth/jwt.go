package auth

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NucleusClaims holds the JWT claims for Open Nucleus auth tokens.
type NucleusClaims struct {
	jwt.RegisteredClaims
	DeviceID    string   `json:"device_id"`
	NodeID      string   `json:"node_id"`
	SiteID      string   `json:"site_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SiteScope   string   `json:"site_scope"` // "local" or "regional"
	TokenType   string   `json:"token_type"` // "access" or "refresh"

	// SMART on FHIR fields (present only on OAuth2 tokens).
	Scope           string `json:"scope,omitempty"`            // space-delimited SMART scopes
	ClientID        string `json:"client_id,omitempty"`        // OAuth2 client ID
	FHIRUser        string `json:"fhirUser,omitempty"`         // "Practitioner/{id}"
	LaunchPatient   string `json:"patient,omitempty"`          // patient launch context
	LaunchEncounter string `json:"encounter,omitempty"`        // encounter launch context
}

// SignToken creates a signed JWT using Ed25519 (EdDSA).
func SignToken(claims NucleusClaims, privateKey ed25519.PrivateKey, keyID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = keyID
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// VerifyToken parses and validates a JWT using Ed25519 (EdDSA).
// Returns the parsed claims or an error.
func VerifyToken(tokenString string, publicKey ed25519.PublicKey) (*NucleusClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &NucleusClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*NucleusClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// NewAccessClaims creates access token claims.
func NewAccessClaims(subject, deviceID, nodeID, siteID, role string, permissions []string, siteScope, jti, issuer string, lifetime time.Duration) NucleusClaims {
	now := time.Now()
	return NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
		},
		DeviceID:    deviceID,
		NodeID:      nodeID,
		SiteID:      siteID,
		Role:        role,
		Permissions: permissions,
		SiteScope:   siteScope,
		TokenType:   "access",
	}
}

// NewRefreshClaims creates refresh token claims.
func NewRefreshClaims(subject, deviceID, jti, issuer string, lifetime time.Duration) NucleusClaims {
	now := time.Now()
	return NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
		},
		DeviceID:  deviceID,
		TokenType: "refresh",
	}
}

// NewSmartAccessClaims creates access token claims with SMART on FHIR fields.
func NewSmartAccessClaims(
	subject, deviceID, nodeID, siteID, role string,
	permissions []string, siteScope string,
	scope, clientID, fhirUser, patientID, encounterID string,
	jti, issuer string, lifetime time.Duration,
) NucleusClaims {
	now := time.Now()
	return NucleusClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   subject,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
		},
		DeviceID:        deviceID,
		NodeID:          nodeID,
		SiteID:          siteID,
		Role:            role,
		Permissions:     permissions,
		SiteScope:       siteScope,
		TokenType:       "access",
		Scope:           scope,
		ClientID:        clientID,
		FHIRUser:        fhirUser,
		LaunchPatient:   patientID,
		LaunchEncounter: encounterID,
	}
}
