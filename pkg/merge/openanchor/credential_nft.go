package openanchor

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log/slog"

	hederabackend "github.com/Open-Nucleus/open-anchor/go/backends/hedera"
)

// HederaConfigForTokens creates a Hedera config suitable for the token service.
func HederaConfigForTokens(network, operatorID, operatorKey string) hederabackend.Config {
	return hederabackend.Config{
		Network:     network,
		OperatorID:  operatorID,
		OperatorKey: operatorKey,
	}
}

// CredentialNFTService wraps the Hedera Token Service for practitioner
// credential NFT minting. Creates a collection on init and mints
// individual NFTs per practitioner registration.
type CredentialNFTService struct {
	tokenService *hederabackend.TokenService
	collectionID string // HTS Token ID for the credential collection
	logger       *slog.Logger
}

// NewCredentialNFTService creates the HTS token service and optionally
// creates the NFT collection if collectionID is empty.
func NewCredentialNFTService(config hederabackend.Config, signingKey ed25519.PrivateKey, collectionID string, logger *slog.Logger) (*CredentialNFTService, error) {
	ts, err := hederabackend.NewTokenService(config, signingKey)
	if err != nil {
		return nil, fmt.Errorf("credential NFT service: %w", err)
	}

	svc := &CredentialNFTService{
		tokenService: ts,
		collectionID: collectionID,
		logger:       logger,
	}

	// Auto-create collection if not provided
	if collectionID == "" {
		id, err := ts.CreateNFTCollection(context.Background(), "OpenNucleusCredentials", "ONC")
		if err != nil {
			logger.Warn("failed to create NFT collection — credential minting disabled", "error", err)
			return svc, nil // non-fatal: service works without minting
		}
		svc.collectionID = id
		logger.Info("created credential NFT collection", "token_id", id)
	} else {
		logger.Info("using existing credential NFT collection", "token_id", collectionID)
	}

	return svc, nil
}

// MintPractitionerCredential mints an NFT for a practitioner.
// Returns the serial number and token ID, or an error.
// Errors are non-fatal — the caller should log and continue.
func (s *CredentialNFTService) MintPractitionerCredential(practitionerID, role, siteID, issuerDID string) (tokenID string, serial int64, err error) {
	if s.collectionID == "" {
		return "", 0, fmt.Errorf("no credential collection configured")
	}

	metadata := hederabackend.CredentialMetadata{
		Type:           "PractitionerCredential",
		PractitionerID: practitionerID,
		Role:           role,
		SiteID:         siteID,
		IssuerDID:      issuerDID,
	}

	serial, err = s.tokenService.MintCredentialNFT(context.Background(), s.collectionID, metadata)
	if err != nil {
		return "", 0, fmt.Errorf("mint credential: %w", err)
	}

	s.logger.Info("minted practitioner credential NFT",
		"practitioner", practitionerID,
		"role", role,
		"token_id", s.collectionID,
		"serial", serial,
	)

	return s.collectionID, serial, nil
}

// CollectionID returns the HTS Token ID for the credential collection.
func (s *CredentialNFTService) CollectionID() string {
	return s.collectionID
}

// Close releases resources.
func (s *CredentialNFTService) Close() error {
	if s.tokenService != nil {
		return s.tokenService.Close()
	}
	return nil
}
