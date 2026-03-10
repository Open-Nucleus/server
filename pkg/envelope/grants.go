package envelope

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	nucleocrypto "github.com/FibrinLab/open-nucleus/pkg/crypto"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	grantHKDFSalt   = "open-nucleus-grant-v1"
	grantHKDFInfo   = "grant-wrapping"
	grantsDirPrefix = ".nucleus/grants/"
)

// GrantAccess wraps the patient's DEK for a specific provider using ECDH.
func (m *FileKeyManager) GrantAccess(patientID string, providerPubKey ed25519.PublicKey, nodePrivKey ed25519.PrivateKey) error {
	dek, err := m.GetOrCreateKey(patientID)
	if err != nil {
		return fmt.Errorf("grant: get DEK: %w", err)
	}

	wrappingKey, err := deriveGrantKey(nodePrivKey, providerPubKey)
	if err != nil {
		return fmt.Errorf("grant: derive key: %w", err)
	}

	wrappedDEK, err := encryptAESGCM(wrappingKey, dek)
	if err != nil {
		return fmt.Errorf("grant: encrypt DEK: %w", err)
	}

	path := grantPath(patientID, providerPubKey)
	if _, err := m.git.WriteAndCommit(path, wrappedDEK, gitstore.CommitMessage{
		ResourceType: "Grant",
		Operation:    "CREATE",
		ResourceID:   patientID,
	}); err != nil {
		return fmt.Errorf("grant: store: %w", err)
	}

	return nil
}

// RevokeAccess removes a provider's wrapped DEK copy.
func (m *FileKeyManager) RevokeAccess(patientID string, providerPubKey ed25519.PublicKey) error {
	path := grantPath(patientID, providerPubKey)
	if _, err := m.git.WriteAndCommit(path, []byte{}, gitstore.CommitMessage{
		ResourceType: "Grant",
		Operation:    "REVOKE",
		ResourceID:   patientID,
	}); err != nil {
		return fmt.Errorf("revoke grant: %w", err)
	}
	return nil
}

// DecryptFor decrypts ciphertext using a provider's individually-wrapped DEK copy.
func (m *FileKeyManager) DecryptFor(patientID string, providerPrivKey ed25519.PrivateKey, nodePubKey ed25519.PublicKey, ciphertext []byte) ([]byte, error) {
	path := grantPath(patientID, providerPrivKey.Public().(ed25519.PublicKey))
	wrappedDEK, err := m.git.Read(path)
	if err != nil {
		return nil, fmt.Errorf("decrypt-for: read grant: %w", err)
	}
	if len(wrappedDEK) == 0 {
		return nil, fmt.Errorf("decrypt-for: grant revoked or not found for patient %s", patientID)
	}

	wrappingKey, err := deriveGrantKey(providerPrivKey, nodePubKey)
	if err != nil {
		return nil, fmt.Errorf("decrypt-for: derive key: %w", err)
	}

	dek, err := decryptAESGCM(wrappingKey, wrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("decrypt-for: unwrap DEK: %w", err)
	}

	plaintext, err := decryptAESGCM(dek, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt-for: decrypt data: %w", err)
	}

	return plaintext, nil
}

func grantPath(patientID string, pubKey ed25519.PublicKey) string {
	hash := sha256.Sum256(pubKey)
	return fmt.Sprintf("%s%s/%s.grant", grantsDirPrefix, patientID, hex.EncodeToString(hash[:8]))
}

func deriveGrantKey(myPrivate ed25519.PrivateKey, theirPublic ed25519.PublicKey) ([]byte, error) {
	x25519Private := nucleocrypto.EdPrivateToX25519(myPrivate)
	x25519Public, err := nucleocrypto.EdPublicToX25519(theirPublic)
	if err != nil {
		return nil, fmt.Errorf("convert public key: %w", err)
	}

	sharedSecret, err := curve25519.X25519(x25519Private, x25519Public)
	if err != nil {
		return nil, fmt.Errorf("X25519 ECDH: %w", err)
	}

	hkdfReader := hkdf.New(sha256.New, sharedSecret, []byte(grantHKDFSalt), []byte(grantHKDFInfo))
	derivedKey := make([]byte, 32)
	if _, err := io.ReadFull(hkdfReader, derivedKey); err != nil {
		return nil, fmt.Errorf("HKDF expand: %w", err)
	}

	return derivedKey, nil
}
