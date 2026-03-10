package local

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/services/anchor/anchorservice"
)

// anchorService implements service.AnchorService by calling the real
// AnchorService directly (no gRPC).
type anchorService struct {
	svc *anchorservice.AnchorService
}

// NewAnchorService creates a local adapter for anchor operations.
func NewAnchorService(svc *anchorservice.AnchorService) service.AnchorService {
	return &anchorService{svc: svc}
}

// --- Anchoring ---

func (a *anchorService) GetStatus(_ context.Context) (*service.AnchorStatusResponse, error) {
	result, err := a.svc.GetStatus()
	if err != nil {
		return nil, err
	}
	return &service.AnchorStatusResponse{
		State:          result.State,
		LastAnchorID:   result.LastAnchorID,
		LastAnchorTime: result.LastAnchorTime,
		MerkleRoot:     result.MerkleRoot,
		NodeDID:        result.NodeDID,
		QueueDepth:     result.QueueDepth,
		Backend:        result.Backend,
	}, nil
}

func (a *anchorService) Verify(_ context.Context, commitHash string) (*service.AnchorVerifyResponse, error) {
	result, err := a.svc.Verify(commitHash)
	if err != nil {
		return nil, err
	}
	return &service.AnchorVerifyResponse{
		Verified:   result.Verified,
		AnchorID:   result.AnchorID,
		MerkleRoot: result.MerkleRoot,
		AnchoredAt: result.AnchoredAt,
		CommitHash: result.CommitHash,
		State:      result.State,
	}, nil
}

func (a *anchorService) GetHistory(_ context.Context, page, perPage int) (*service.AnchorHistoryResponse, error) {
	records, total, err := a.svc.GetHistory(page, perPage)
	if err != nil {
		return nil, err
	}
	dtos := make([]service.AnchorRecord, 0, len(records))
	for _, r := range records {
		dtos = append(dtos, service.AnchorRecord{
			AnchorID:   r.AnchorID,
			MerkleRoot: r.MerkleRoot,
			GitHead:    r.GitHead,
			State:      r.State,
			Timestamp:  r.Timestamp,
			Backend:    r.Backend,
			TxID:       r.TxID,
			NodeDID:    r.NodeDID,
		})
	}
	tp := total / perPage
	if total%perPage != 0 {
		tp++
	}
	return &service.AnchorHistoryResponse{
		Records:    dtos,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: tp,
	}, nil
}

func (a *anchorService) TriggerAnchor(_ context.Context) (*service.AnchorTriggerResponse, error) {
	result, err := a.svc.TriggerAnchor(false)
	if err != nil {
		return nil, err
	}
	return &service.AnchorTriggerResponse{
		AnchorID:   result.AnchorID,
		State:      result.State,
		MerkleRoot: result.MerkleRoot,
		GitHead:    result.GitHead,
		Skipped:    result.Skipped,
		Message:    result.Message,
	}, nil
}

// --- DID ---

func (a *anchorService) GetNodeDID(_ context.Context) (*service.DIDDocumentResponse, error) {
	doc, err := a.svc.GetNodeDID()
	if err != nil {
		return nil, err
	}
	return toDIDDocResp(doc), nil
}

func (a *anchorService) GetDeviceDID(_ context.Context, deviceID string) (*service.DIDDocumentResponse, error) {
	doc, err := a.svc.GetDeviceDID(deviceID)
	if err != nil {
		return nil, err
	}
	return toDIDDocResp(doc), nil
}

func (a *anchorService) ResolveDID(_ context.Context, did string) (*service.DIDDocumentResponse, error) {
	doc, err := a.svc.ResolveDID(did)
	if err != nil {
		return nil, err
	}
	return toDIDDocResp(doc), nil
}

// --- Credentials ---

func (a *anchorService) IssueDataIntegrityCredential(_ context.Context, req *service.IssueCredentialRequest) (*service.CredentialResponse, error) {
	vc, err := a.svc.IssueDataIntegrityCredential(req.AnchorID, req.Types, req.AdditionalClaims)
	if err != nil {
		return nil, err
	}
	return toCredResp(vc), nil
}

func (a *anchorService) VerifyCredential(_ context.Context, credentialJSON string) (*service.CredentialVerificationResponse, error) {
	var vc openanchor.VerifiableCredential
	if err := json.Unmarshal([]byte(credentialJSON), &vc); err != nil {
		return nil, fmt.Errorf("invalid credential JSON: %w", err)
	}

	result, err := a.svc.VerifyCredential(&vc)
	if err != nil {
		return nil, err
	}
	return &service.CredentialVerificationResponse{
		Valid:   result.Valid,
		Issuer:  result.Issuer,
		Message: result.Message,
	}, nil
}

func (a *anchorService) ListCredentials(_ context.Context, credType string, page, perPage int) (*service.CredentialListResponse, error) {
	creds, total, err := a.svc.ListCredentials(credType, page, perPage)
	if err != nil {
		return nil, err
	}
	dtos := make([]service.CredentialResponse, 0, len(creds))
	for i := range creds {
		if c := toCredResp(&creds[i]); c != nil {
			dtos = append(dtos, *c)
		}
	}
	tp := total / perPage
	if total%perPage != 0 {
		tp++
	}
	return &service.CredentialListResponse{
		Credentials: dtos,
		Page:        page,
		PerPage:     perPage,
		Total:       total,
		TotalPages:  tp,
	}, nil
}

// --- Backend ---

func (a *anchorService) ListBackends(_ context.Context) (*service.BackendListResponse, error) {
	backends := a.svc.ListBackends()
	dtos := make([]service.BackendInfoDTO, 0, len(backends))
	for _, b := range backends {
		dtos = append(dtos, service.BackendInfoDTO{
			Name:        b.Name,
			Available:   b.Available,
			Description: b.Description,
		})
	}
	return &service.BackendListResponse{Backends: dtos}, nil
}

func (a *anchorService) GetBackendStatus(_ context.Context, name string) (*service.BackendStatusResponse, error) {
	result, err := a.svc.GetBackendStatus(name)
	if err != nil {
		return nil, err
	}
	return &service.BackendStatusResponse{
		Name:           result.Name,
		Available:      result.Available,
		Description:    result.Description,
		AnchoredCount:  result.AnchoredCount,
		LastAnchorTime: result.LastAnchorTime,
	}, nil
}

func (a *anchorService) GetQueueStatus(_ context.Context) (*service.QueueStatusResponse, error) {
	result, err := a.svc.GetQueueStatus()
	if err != nil {
		return nil, err
	}
	entries := make([]service.QueueEntryDTO, 0, len(result.Entries))
	for _, e := range result.Entries {
		entries = append(entries, service.QueueEntryDTO{
			AnchorID:   e.AnchorID,
			MerkleRoot: e.MerkleRoot,
			GitHead:    e.GitHead,
			EnqueuedAt: e.EnqueuedAt,
			State:      e.State,
		})
	}
	return &service.QueueStatusResponse{
		Pending:        result.Pending,
		TotalProcessed: result.TotalProcessed,
		Entries:        entries,
	}, nil
}

// --- Health ---

func (a *anchorService) Health(_ context.Context) (*service.AnchorHealthResponse, error) {
	return &service.AnchorHealthResponse{
		Status:      "healthy",
		NodeDID:     a.svc.NodeDIDString(),
		Backend:     a.svc.BackendName(),
		AnchorCount: a.svc.AnchorCount(),
		QueueDepth:  a.svc.QueueDepth(),
	}, nil
}

// --- Helpers ---

func toDIDDocResp(doc *openanchor.DIDDocument) *service.DIDDocumentResponse {
	if doc == nil {
		return nil
	}
	vms := make([]service.VerificationMethodDTO, 0, len(doc.VerificationMethod))
	for _, vm := range doc.VerificationMethod {
		vms = append(vms, service.VerificationMethodDTO{
			ID:                 vm.ID,
			Type:               vm.Type,
			Controller:         vm.Controller,
			PublicKeyMultibase: vm.PublicKeyMultibase,
		})
	}
	return &service.DIDDocumentResponse{
		ID:                 doc.ID,
		Context:            doc.Context,
		VerificationMethod: vms,
		Authentication:     doc.Authentication,
		AssertionMethod:    doc.AssertionMethod,
		Created:            doc.Created,
	}
}

func toCredResp(vc *openanchor.VerifiableCredential) *service.CredentialResponse {
	if vc == nil {
		return nil
	}
	subjectJSON, _ := json.Marshal(vc.CredentialSubject)

	resp := &service.CredentialResponse{
		ID:                    vc.ID,
		Context:               vc.Context,
		Type:                  vc.Type,
		Issuer:                vc.Issuer,
		IssuanceDate:          vc.IssuanceDate,
		ExpirationDate:        vc.ExpirationDate,
		CredentialSubjectJSON: string(subjectJSON),
	}
	if vc.Proof != nil {
		resp.Proof = &service.CredentialProofDTO{
			Type:               vc.Proof.Type,
			Created:            vc.Proof.Created,
			VerificationMethod: vc.Proof.VerificationMethod,
			ProofPurpose:       vc.Proof.ProofPurpose,
			ProofValue:         vc.Proof.ProofValue,
		}
	}
	return resp
}
