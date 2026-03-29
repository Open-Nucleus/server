package openanchor

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IotaIdentityBridge implements IdentityEngine by calling the Node.js
// identity bridge service (open-anchor/bridge) over HTTP.
type IotaIdentityBridge struct {
	bridgeURL string
	client    *http.Client
}

// NewIotaIdentityBridge returns an IdentityEngine backed by the IOTA identity
// bridge running at bridgeURL (e.g. "http://localhost:3001").
func NewIotaIdentityBridge(bridgeURL string) *IotaIdentityBridge {
	return &IotaIdentityBridge{
		bridgeURL: bridgeURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- request / response DTOs for the bridge API ---

type bridgeCreateDIDReq struct {
	PublicKey string `json:"publicKey"` // base64-encoded Ed25519 public key
}

type bridgeCreateDIDResp struct {
	DID      string      `json:"did"`
	Document DIDDocument `json:"document"`
}

type bridgeResolveDIDResp struct {
	DID      string      `json:"did"`
	Document DIDDocument `json:"document"`
}

type bridgeIssueVCReq struct {
	IssuerDID  string         `json:"issuerDid"`
	IssuerKey  string         `json:"issuerKey"`  // base64-encoded Ed25519 private key
	Claims     bridgeClaims   `json:"claims"`
}

type bridgeClaims struct {
	ID             string         `json:"id"`
	Types          []string       `json:"types"`
	Subject        map[string]any `json:"subject"`
	ExpirationDate string         `json:"expirationDate,omitempty"`
}

type bridgeIssueVCResp struct {
	Credential VerifiableCredential `json:"credential"`
}

type bridgeVerifyVCReq struct {
	Credential *VerifiableCredential `json:"credential"`
}

type bridgeVerifyVCResp struct {
	Valid   bool   `json:"valid"`
	Issuer  string `json:"issuer"`
	Message string `json:"message"`
}

type bridgeErrorResp struct {
	Error string `json:"error"`
}

// --- IdentityEngine implementation ---

// GenerateDID creates a DID via the IOTA identity bridge by POSTing the
// Ed25519 public key to /did/create.
func (b *IotaIdentityBridge) GenerateDID(pub ed25519.PublicKey) (*DIDDocument, error) {
	reqBody := bridgeCreateDIDReq{
		PublicKey: base64.StdEncoding.EncodeToString(pub),
	}

	var resp bridgeCreateDIDResp
	if err := b.post("/did/create", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("bridge GenerateDID: %w", err)
	}
	return &resp.Document, nil
}

// ResolveDID resolves a DID string via GET /did/resolve/{did}.
func (b *IotaIdentityBridge) ResolveDID(did string) (*DIDDocument, error) {
	var resp bridgeResolveDIDResp
	if err := b.get("/did/resolve/"+did, &resp); err != nil {
		return nil, fmt.Errorf("bridge ResolveDID: %w", err)
	}
	return &resp.Document, nil
}

// IssueCredential creates a signed Verifiable Credential via POST /vc/issue.
func (b *IotaIdentityBridge) IssueCredential(claims CredentialClaims, issuerDID string, issuerKey ed25519.PrivateKey) (*VerifiableCredential, error) {
	reqBody := bridgeIssueVCReq{
		IssuerDID: issuerDID,
		IssuerKey: base64.StdEncoding.EncodeToString(issuerKey),
		Claims: bridgeClaims{
			ID:             claims.ID,
			Types:          claims.Types,
			Subject:        claims.Subject,
			ExpirationDate: claims.ExpirationDate,
		},
	}

	var resp bridgeIssueVCResp
	if err := b.post("/vc/issue", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("bridge IssueCredential: %w", err)
	}
	return &resp.Credential, nil
}

// VerifyCredential verifies a VC via POST /vc/verify.
func (b *IotaIdentityBridge) VerifyCredential(vc *VerifiableCredential) (*VerificationResult, error) {
	reqBody := bridgeVerifyVCReq{
		Credential: vc,
	}

	var resp bridgeVerifyVCResp
	if err := b.post("/vc/verify", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("bridge VerifyCredential: %w", err)
	}
	return &VerificationResult{
		Valid:   resp.Valid,
		Issuer:  resp.Issuer,
		Message: resp.Message,
	}, nil
}

// --- HTTP helpers ---

func (b *IotaIdentityBridge) post(path string, body any, dst any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := b.client.Post(b.bridgeURL+path, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("HTTP POST %s: %w", path, err)
	}
	defer resp.Body.Close()

	return b.decodeResponse(resp, dst)
}

func (b *IotaIdentityBridge) get(path string, dst any) error {
	resp, err := b.client.Get(b.bridgeURL + path)
	if err != nil {
		return fmt.Errorf("HTTP GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	return b.decodeResponse(resp, dst)
}

func (b *IotaIdentityBridge) decodeResponse(resp *http.Response, dst any) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp bridgeErrorResp
		if json.Unmarshal(data, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("bridge returned %d: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("bridge returned %d: %s", resp.StatusCode, string(data))
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
