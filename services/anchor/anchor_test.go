package anchor_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	"github.com/FibrinLab/open-nucleus/services/anchor/anchortest"
)

func setup(t *testing.T) *anchortest.Env {
	t.Helper()
	tmpDir := t.TempDir()
	return anchortest.Start(t, tmpDir)
}

// --- Health ---

func TestHealth(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.Health(context.Background(), &anchorv1.HealthRequest{})
	if err != nil {
		t.Fatalf("Health: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("expected healthy, got %s", resp.Status)
	}
	if resp.NodeDid == "" {
		t.Error("expected node DID in health response")
	}
	if resp.Backend != "none" {
		t.Errorf("expected backend=none, got %s", resp.Backend)
	}
}

// --- Anchor Status ---

func TestGetStatus_Initial(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetStatus(context.Background(), &anchorv1.GetAnchorStatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if resp.State != "idle" {
		t.Errorf("expected state=idle, got %s", resp.State)
	}
	if !strings.HasPrefix(resp.NodeDid, "did:key:z") {
		t.Errorf("expected node DID starting with did:key:z, got %s", resp.NodeDid)
	}
	if resp.QueueDepth != 0 {
		t.Errorf("expected queue_depth=0, got %d", resp.QueueDepth)
	}
	if resp.Backend != "none" {
		t.Errorf("expected backend=none, got %s", resp.Backend)
	}
}

// --- Trigger Anchor ---

func TestTriggerAnchor_First(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.TriggerAnchor(context.Background(), &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("TriggerAnchor: %v", err)
	}
	if resp.Skipped {
		t.Fatalf("expected anchor to not be skipped, got message: %s", resp.Message)
	}
	if resp.AnchorId == "" {
		t.Error("expected anchor ID")
	}
	if resp.MerkleRoot == "" {
		t.Error("expected merkle root")
	}
	if resp.GitHead == "" {
		t.Error("expected git head")
	}
	if resp.State != "queued" {
		t.Errorf("expected state=queued (stub backend), got %s", resp.State)
	}
}

func TestTriggerAnchor_SkipUnchanged(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// First trigger.
	_, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("first TriggerAnchor: %v", err)
	}

	// Second trigger — should be skipped (no changes).
	resp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("second TriggerAnchor: %v", err)
	}
	if !resp.Skipped {
		t.Error("expected second anchor to be skipped (unchanged)")
	}
	if resp.Message == "" {
		t.Error("expected skip message")
	}
}

func TestTriggerAnchor_Manual(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// First trigger.
	_, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("first TriggerAnchor: %v", err)
	}

	// Manual trigger — should NOT be skipped even with no changes.
	resp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{Manual: true})
	if err != nil {
		t.Fatalf("manual TriggerAnchor: %v", err)
	}
	if resp.Skipped {
		t.Error("expected manual trigger to NOT be skipped")
	}
	if resp.AnchorId == "" {
		t.Error("expected anchor ID for manual trigger")
	}
}

// --- Verify ---

func TestVerify_ValidCommit(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Trigger an anchor.
	trigResp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("TriggerAnchor: %v", err)
	}

	// Verify the commit.
	resp, err := env.Client.VerifyAnchor(ctx, &anchorv1.VerifyAnchorRequest{
		CommitHash: trigResp.GitHead,
	})
	if err != nil {
		t.Fatalf("VerifyAnchor: %v", err)
	}
	if !resp.Verified {
		t.Error("expected verified=true for anchored commit")
	}
	if resp.AnchorId != trigResp.AnchorId {
		t.Errorf("expected anchor ID %s, got %s", trigResp.AnchorId, resp.AnchorId)
	}
	if resp.MerkleRoot != trigResp.MerkleRoot {
		t.Errorf("expected merkle root %s, got %s", trigResp.MerkleRoot, resp.MerkleRoot)
	}
}

func TestVerify_UnknownCommit(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.VerifyAnchor(context.Background(), &anchorv1.VerifyAnchorRequest{
		CommitHash: "0000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("VerifyAnchor: %v", err)
	}
	if resp.Verified {
		t.Error("expected verified=false for unknown commit")
	}
}

// --- History ---

func TestGetHistory_Pagination(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Create 3 anchors using manual mode.
	for i := 0; i < 3; i++ {
		_, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{Manual: true})
		if err != nil {
			t.Fatalf("TriggerAnchor %d: %v", i, err)
		}
	}

	// Get first page (2 per page).
	resp, err := env.Client.GetHistory(ctx, &anchorv1.GetAnchorHistoryRequest{
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 2},
	})
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(resp.Records) != 2 {
		t.Errorf("expected 2 records on page 1, got %d", len(resp.Records))
	}
	if resp.Pagination.Total != 3 {
		t.Errorf("expected total=3, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.TotalPages != 2 {
		t.Errorf("expected 2 total pages, got %d", resp.Pagination.TotalPages)
	}

	// Get second page.
	resp2, err := env.Client.GetHistory(ctx, &anchorv1.GetAnchorHistoryRequest{
		Pagination: &commonv1.PaginationRequest{Page: 2, PerPage: 2},
	})
	if err != nil {
		t.Fatalf("GetHistory page 2: %v", err)
	}
	if len(resp2.Records) != 1 {
		t.Errorf("expected 1 record on page 2, got %d", len(resp2.Records))
	}
}

// --- DID ---

func TestGetNodeDID(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetNodeDID(context.Background(), &anchorv1.GetNodeDIDRequest{})
	if err != nil {
		t.Fatalf("GetNodeDID: %v", err)
	}
	if resp.Document == nil {
		t.Fatal("expected DID document")
	}
	if !strings.HasPrefix(resp.Document.Id, "did:key:z") {
		t.Errorf("expected DID starting with did:key:z, got %s", resp.Document.Id)
	}
	if len(resp.Document.VerificationMethod) == 0 {
		t.Error("expected at least 1 verification method")
	}
	if len(resp.Document.Authentication) == 0 {
		t.Error("expected authentication references")
	}
}

func TestGetDeviceDID(t *testing.T) {
	env := setup(t)

	// Bootstrap a device DID through the service directly.
	env.Svc.BootstrapDeviceDID("device-001", env.Svc.NodePublicKey())

	resp, err := env.Client.GetDeviceDID(context.Background(), &anchorv1.GetDeviceDIDRequest{
		DeviceId: "device-001",
	})
	if err != nil {
		t.Fatalf("GetDeviceDID: %v", err)
	}
	if resp.Document == nil {
		t.Fatal("expected device DID document")
	}
	if !strings.HasPrefix(resp.Document.Id, "did:key:z") {
		t.Errorf("expected device DID starting with did:key:z, got %s", resp.Document.Id)
	}
}

func TestResolveDID_Valid(t *testing.T) {
	env := setup(t)

	// Get the node DID first.
	nodeDID, err := env.Client.GetNodeDID(context.Background(), &anchorv1.GetNodeDIDRequest{})
	if err != nil {
		t.Fatalf("GetNodeDID: %v", err)
	}

	// Resolve it.
	resp, err := env.Client.ResolveDID(context.Background(), &anchorv1.ResolveDIDRequest{
		Did: nodeDID.Document.Id,
	})
	if err != nil {
		t.Fatalf("ResolveDID: %v", err)
	}
	if resp.Document == nil {
		t.Fatal("expected resolved DID document")
	}
	if resp.Document.Id != nodeDID.Document.Id {
		t.Errorf("expected resolved ID=%s, got %s", nodeDID.Document.Id, resp.Document.Id)
	}
}

func TestResolveDID_Invalid(t *testing.T) {
	env := setup(t)
	_, err := env.Client.ResolveDID(context.Background(), &anchorv1.ResolveDIDRequest{
		Did: "did:bad:invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid DID")
	}
}

// --- Credentials ---

func TestIssueCredential(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// First trigger an anchor.
	trigResp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("TriggerAnchor: %v", err)
	}

	// Issue a credential for the anchor.
	resp, err := env.Client.IssueDataIntegrityCredential(ctx, &anchorv1.IssueCredentialRequest{
		AnchorId: trigResp.AnchorId,
	})
	if err != nil {
		t.Fatalf("IssueDataIntegrityCredential: %v", err)
	}
	if resp.Credential == nil {
		t.Fatal("expected credential")
	}
	if resp.Credential.Id == "" {
		t.Error("expected credential ID")
	}
	if resp.Credential.Issuer == "" {
		t.Error("expected issuer")
	}
	if resp.Credential.Proof == nil {
		t.Error("expected proof")
	}

	// Check types include DataIntegrityCredential.
	found := false
	for _, typ := range resp.Credential.Type {
		if typ == "DataIntegrityCredential" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DataIntegrityCredential type, got %v", resp.Credential.Type)
	}
}

func TestVerifyCredential_Valid(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Trigger + issue.
	trigResp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("TriggerAnchor: %v", err)
	}
	issueResp, err := env.Client.IssueDataIntegrityCredential(ctx, &anchorv1.IssueCredentialRequest{
		AnchorId: trigResp.AnchorId,
	})
	if err != nil {
		t.Fatalf("IssueDataIntegrityCredential: %v", err)
	}

	// Reconstruct the VC JSON from proto fields.
	vcJSON := protoCredentialToJSON(t, issueResp.Credential)

	// Verify the credential.
	resp, err := env.Client.VerifyCredential(ctx, &anchorv1.VerifyCredentialRequest{
		CredentialJson: vcJSON,
	})
	if err != nil {
		t.Fatalf("VerifyCredential: %v", err)
	}
	if !resp.Valid {
		t.Errorf("expected valid=true, got message: %s", resp.Message)
	}
	if resp.Issuer == "" {
		t.Error("expected issuer in verification result")
	}
}

func TestVerifyCredential_Tampered(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Trigger + issue.
	trigResp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("TriggerAnchor: %v", err)
	}
	issueResp, err := env.Client.IssueDataIntegrityCredential(ctx, &anchorv1.IssueCredentialRequest{
		AnchorId: trigResp.AnchorId,
	})
	if err != nil {
		t.Fatalf("IssueDataIntegrityCredential: %v", err)
	}

	// Tamper with the credential subject.
	vcJSON := protoCredentialToJSON(t, issueResp.Credential)
	vcJSON = strings.Replace(vcJSON, trigResp.MerkleRoot, "tampered_root_0000", 1)

	// Verify — should fail.
	resp, err := env.Client.VerifyCredential(ctx, &anchorv1.VerifyCredentialRequest{
		CredentialJson: vcJSON,
	})
	if err != nil {
		t.Fatalf("VerifyCredential: %v", err)
	}
	if resp.Valid {
		t.Error("expected valid=false for tampered credential")
	}
}

func TestListCredentials(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Create 2 anchors + 2 credentials.
	for i := 0; i < 2; i++ {
		trigResp, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{Manual: true})
		if err != nil {
			t.Fatalf("TriggerAnchor %d: %v", i, err)
		}
		_, err = env.Client.IssueDataIntegrityCredential(ctx, &anchorv1.IssueCredentialRequest{
			AnchorId: trigResp.AnchorId,
		})
		if err != nil {
			t.Fatalf("IssueDataIntegrityCredential %d: %v", i, err)
		}
	}

	// List all credentials.
	resp, err := env.Client.ListCredentials(ctx, &anchorv1.ListCredentialsRequest{
		Pagination: &commonv1.PaginationRequest{Page: 1, PerPage: 25},
	})
	if err != nil {
		t.Fatalf("ListCredentials: %v", err)
	}
	if len(resp.Credentials) != 2 {
		t.Errorf("expected 2 credentials, got %d", len(resp.Credentials))
	}
	if resp.Pagination.Total != 2 {
		t.Errorf("expected total=2, got %d", resp.Pagination.Total)
	}
}

// --- Backend ---

func TestListBackends(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.ListBackends(context.Background(), &anchorv1.ListBackendsRequest{})
	if err != nil {
		t.Fatalf("ListBackends: %v", err)
	}
	if len(resp.Backends) != 1 {
		t.Fatalf("expected 1 backend, got %d", len(resp.Backends))
	}
	if resp.Backends[0].Name != "none" {
		t.Errorf("expected backend name=none, got %s", resp.Backends[0].Name)
	}
	if resp.Backends[0].Available {
		t.Error("expected backend available=false for stub")
	}
}

func TestGetBackendStatus(t *testing.T) {
	env := setup(t)
	resp, err := env.Client.GetBackendStatus(context.Background(), &anchorv1.GetBackendStatusRequest{
		Name: "none",
	})
	if err != nil {
		t.Fatalf("GetBackendStatus: %v", err)
	}
	if resp.Name != "none" {
		t.Errorf("expected name=none, got %s", resp.Name)
	}
	if resp.Available {
		t.Error("expected available=false")
	}
}

// --- Queue ---

func TestGetQueueStatus(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Trigger an anchor — stub backend will enqueue it.
	_, err := env.Client.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		t.Fatalf("TriggerAnchor: %v", err)
	}

	resp, err := env.Client.GetQueueStatus(ctx, &anchorv1.GetQueueStatusRequest{})
	if err != nil {
		t.Fatalf("GetQueueStatus: %v", err)
	}
	if resp.Pending != 1 {
		t.Errorf("expected 1 pending in queue, got %d", resp.Pending)
	}
	if len(resp.Entries) != 1 {
		t.Errorf("expected 1 queue entry, got %d", len(resp.Entries))
	}
}

// --- Helper ---

// protoCredentialToJSON reconstructs a VerifiableCredential JSON from proto fields.
func protoCredentialToJSON(t *testing.T, cred *anchorv1.VerifiableCredential) string {
	t.Helper()

	var subject map[string]any
	if err := json.Unmarshal([]byte(cred.CredentialSubjectJson), &subject); err != nil {
		t.Fatalf("unmarshal subject JSON: %v", err)
	}

	vc := map[string]any{
		"@context":          cred.Context,
		"id":                cred.Id,
		"type":              cred.Type,
		"issuer":            cred.Issuer,
		"issuanceDate":      cred.IssuanceDate,
		"credentialSubject": subject,
	}
	if cred.ExpirationDate != "" {
		vc["expirationDate"] = cred.ExpirationDate
	}
	if cred.Proof != nil {
		vc["proof"] = map[string]any{
			"type":               cred.Proof.Type,
			"created":            cred.Proof.Created,
			"verificationMethod": cred.Proof.VerificationMethod,
			"proofPurpose":       cred.Proof.ProofPurpose,
			"proofValue":         cred.Proof.ProofValue,
		}
	}

	data, err := json.Marshal(vc)
	if err != nil {
		t.Fatalf("marshal VC JSON: %v", err)
	}
	return string(data)
}
