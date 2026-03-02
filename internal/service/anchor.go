package service

import (
	"context"
	"fmt"

	commonv1 "github.com/FibrinLab/open-nucleus/gen/proto/common/v1"
	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
	"github.com/FibrinLab/open-nucleus/internal/grpcclient"
)

type anchorAdapter struct {
	pool *grpcclient.Pool
}

func NewAnchorService(pool *grpcclient.Pool) AnchorService {
	return &anchorAdapter{pool: pool}
}

func (a *anchorAdapter) client() (anchorv1.AnchorServiceClient, error) {
	conn, err := a.pool.Conn("anchor")
	if err != nil {
		return nil, fmt.Errorf("anchor service unavailable: %w", err)
	}
	return anchorv1.NewAnchorServiceClient(conn), nil
}

func (a *anchorAdapter) GetStatus(ctx context.Context) (*AnchorStatusResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetStatus(ctx, &anchorv1.GetAnchorStatusRequest{})
	if err != nil {
		return nil, err
	}
	return &AnchorStatusResponse{
		State:          resp.State,
		LastAnchorID:   resp.LastAnchorId,
		LastAnchorTime: resp.LastAnchorTime,
		MerkleRoot:     resp.MerkleRoot,
		NodeDID:        resp.NodeDid,
		QueueDepth:     int(resp.QueueDepth),
		Backend:        resp.Backend,
		PendingCommits: int(resp.PendingCommits),
	}, nil
}

func (a *anchorAdapter) Verify(ctx context.Context, commitHash string) (*AnchorVerifyResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.VerifyAnchor(ctx, &anchorv1.VerifyAnchorRequest{CommitHash: commitHash})
	if err != nil {
		return nil, err
	}
	return &AnchorVerifyResponse{
		Verified:   resp.Verified,
		AnchorID:   resp.AnchorId,
		MerkleRoot: resp.MerkleRoot,
		AnchoredAt: resp.AnchoredAt,
		CommitHash: resp.CommitHash,
		State:      resp.State,
	}, nil
}

func (a *anchorAdapter) GetHistory(ctx context.Context, page, perPage int) (*AnchorHistoryResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetHistory(ctx, &anchorv1.GetAnchorHistoryRequest{
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, err
	}
	records := make([]AnchorRecord, 0, len(resp.Records))
	for _, r := range resp.Records {
		records = append(records, AnchorRecord{
			AnchorID:   r.AnchorId,
			MerkleRoot: r.MerkleRoot,
			GitHead:    r.GitHead,
			State:      r.State,
			Timestamp:  r.Timestamp,
			Backend:    r.Backend,
			TxID:       r.TxId,
			NodeDID:    r.NodeDid,
		})
	}
	return &AnchorHistoryResponse{
		Records:    records,
		Page:       anchorPgPage(resp.Pagination),
		PerPage:    anchorPgPerPage(resp.Pagination),
		Total:      anchorPgTotal(resp.Pagination),
		TotalPages: anchorPgTotalPages(resp.Pagination),
	}, nil
}

func (a *anchorAdapter) TriggerAnchor(ctx context.Context) (*AnchorTriggerResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.TriggerAnchor(ctx, &anchorv1.TriggerAnchorRequest{})
	if err != nil {
		return nil, err
	}
	return &AnchorTriggerResponse{
		AnchorID:   resp.AnchorId,
		State:      resp.State,
		MerkleRoot: resp.MerkleRoot,
		GitHead:    resp.GitHead,
		Skipped:    resp.Skipped,
		Message:    resp.Message,
	}, nil
}

func (a *anchorAdapter) GetNodeDID(ctx context.Context) (*DIDDocumentResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetNodeDID(ctx, &anchorv1.GetNodeDIDRequest{})
	if err != nil {
		return nil, err
	}
	return toDIDDocumentResponse(resp.Document), nil
}

func (a *anchorAdapter) GetDeviceDID(ctx context.Context, deviceID string) (*DIDDocumentResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetDeviceDID(ctx, &anchorv1.GetDeviceDIDRequest{DeviceId: deviceID})
	if err != nil {
		return nil, err
	}
	return toDIDDocumentResponse(resp.Document), nil
}

func (a *anchorAdapter) ResolveDID(ctx context.Context, did string) (*DIDDocumentResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ResolveDID(ctx, &anchorv1.ResolveDIDRequest{Did: did})
	if err != nil {
		return nil, err
	}
	return toDIDDocumentResponse(resp.Document), nil
}

func (a *anchorAdapter) IssueDataIntegrityCredential(ctx context.Context, req *IssueCredentialRequest) (*CredentialResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.IssueDataIntegrityCredential(ctx, &anchorv1.IssueCredentialRequest{
		AnchorId:         req.AnchorID,
		Types:            req.Types,
		AdditionalClaims: req.AdditionalClaims,
	})
	if err != nil {
		return nil, err
	}
	return toCredentialResponse(resp.Credential), nil
}

func (a *anchorAdapter) VerifyCredential(ctx context.Context, credentialJSON string) (*CredentialVerificationResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.VerifyCredential(ctx, &anchorv1.VerifyCredentialRequest{CredentialJson: credentialJSON})
	if err != nil {
		return nil, err
	}
	return &CredentialVerificationResponse{
		Valid:   resp.Valid,
		Issuer:  resp.Issuer,
		Message: resp.Message,
	}, nil
}

func (a *anchorAdapter) ListCredentials(ctx context.Context, credType string, page, perPage int) (*CredentialListResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ListCredentials(ctx, &anchorv1.ListCredentialsRequest{
		Type: credType,
		Pagination: &commonv1.PaginationRequest{
			Page:    int32(page),
			PerPage: int32(perPage),
		},
	})
	if err != nil {
		return nil, err
	}
	creds := make([]CredentialResponse, 0, len(resp.Credentials))
	for _, cr := range resp.Credentials {
		if c := toCredentialResponse(cr); c != nil {
			creds = append(creds, *c)
		}
	}
	return &CredentialListResponse{
		Credentials: creds,
		Page:        anchorPgPage(resp.Pagination),
		PerPage:     anchorPgPerPage(resp.Pagination),
		Total:       anchorPgTotal(resp.Pagination),
		TotalPages:  anchorPgTotalPages(resp.Pagination),
	}, nil
}

func (a *anchorAdapter) ListBackends(ctx context.Context) (*BackendListResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.ListBackends(ctx, &anchorv1.ListBackendsRequest{})
	if err != nil {
		return nil, err
	}
	backends := make([]BackendInfoDTO, 0, len(resp.Backends))
	for _, b := range resp.Backends {
		backends = append(backends, BackendInfoDTO{
			Name:        b.Name,
			Available:   b.Available,
			Description: b.Description,
		})
	}
	return &BackendListResponse{Backends: backends}, nil
}

func (a *anchorAdapter) GetBackendStatus(ctx context.Context, name string) (*BackendStatusResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetBackendStatus(ctx, &anchorv1.GetBackendStatusRequest{Name: name})
	if err != nil {
		return nil, err
	}
	return &BackendStatusResponse{
		Name:           resp.Name,
		Available:      resp.Available,
		Description:    resp.Description,
		AnchoredCount:  int(resp.AnchoredCount),
		LastAnchorTime: resp.LastAnchorTime,
	}, nil
}

func (a *anchorAdapter) GetQueueStatus(ctx context.Context) (*QueueStatusResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetQueueStatus(ctx, &anchorv1.GetQueueStatusRequest{})
	if err != nil {
		return nil, err
	}
	entries := make([]QueueEntryDTO, 0, len(resp.Entries))
	for _, e := range resp.Entries {
		entries = append(entries, QueueEntryDTO{
			AnchorID:   e.AnchorId,
			MerkleRoot: e.MerkleRoot,
			GitHead:    e.GitHead,
			EnqueuedAt: e.EnqueuedAt,
			State:      e.State,
		})
	}
	return &QueueStatusResponse{
		Pending:        int(resp.Pending),
		TotalProcessed: int(resp.TotalProcessed),
		Entries:        entries,
	}, nil
}

func (a *anchorAdapter) Health(ctx context.Context) (*AnchorHealthResponse, error) {
	c, err := a.client()
	if err != nil {
		return nil, err
	}
	resp, err := c.Health(ctx, &anchorv1.HealthRequest{})
	if err != nil {
		return nil, err
	}
	return &AnchorHealthResponse{
		Status:      resp.Status,
		NodeDID:     resp.NodeDid,
		Backend:     resp.Backend,
		AnchorCount: int(resp.AnchorCount),
		QueueDepth:  int(resp.QueueDepth),
	}, nil
}

// --- Proto → DTO converters ---

func toDIDDocumentResponse(doc *anchorv1.DIDDocument) *DIDDocumentResponse {
	if doc == nil {
		return nil
	}
	vms := make([]VerificationMethodDTO, 0, len(doc.VerificationMethod))
	for _, vm := range doc.VerificationMethod {
		vms = append(vms, VerificationMethodDTO{
			ID:                 vm.Id,
			Type:               vm.Type,
			Controller:         vm.Controller,
			PublicKeyMultibase: vm.PublicKeyMultibase,
		})
	}
	return &DIDDocumentResponse{
		ID:                 doc.Id,
		Context:            doc.Context,
		VerificationMethod: vms,
		Authentication:     doc.Authentication,
		AssertionMethod:    doc.AssertionMethod,
		Created:            doc.Created,
	}
}

func toCredentialResponse(c *anchorv1.VerifiableCredential) *CredentialResponse {
	if c == nil {
		return nil
	}
	resp := &CredentialResponse{
		ID:                    c.Id,
		Context:               c.Context,
		Type:                  c.Type,
		Issuer:                c.Issuer,
		IssuanceDate:          c.IssuanceDate,
		ExpirationDate:        c.ExpirationDate,
		CredentialSubjectJSON: c.CredentialSubjectJson,
	}
	if c.Proof != nil {
		resp.Proof = &CredentialProofDTO{
			Type:               c.Proof.Type,
			Created:            c.Proof.Created,
			VerificationMethod: c.Proof.VerificationMethod,
			ProofPurpose:       c.Proof.ProofPurpose,
			ProofValue:         c.Proof.ProofValue,
		}
	}
	return resp
}

func anchorPgPage(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 1
	}
	return int(pg.Page)
}

func anchorPgPerPage(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 25
	}
	return int(pg.PerPage)
}

func anchorPgTotal(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 0
	}
	return int(pg.Total)
}

func anchorPgTotalPages(pg *commonv1.PaginationResponse) int {
	if pg == nil {
		return 0
	}
	return int(pg.TotalPages)
}
