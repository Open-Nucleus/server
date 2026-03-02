package server

import (
	"context"
	"encoding/json"
	"fmt"

	anchorv1 "github.com/FibrinLab/open-nucleus/gen/proto/anchor/v1"
	"github.com/FibrinLab/open-nucleus/pkg/openanchor"
)

func (s *Server) IssueDataIntegrityCredential(_ context.Context, req *anchorv1.IssueCredentialRequest) (*anchorv1.IssueCredentialResponse, error) {
	vc, err := s.svc.IssueDataIntegrityCredential(req.AnchorId, req.Types, req.AdditionalClaims)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.IssueCredentialResponse{
		Credential: vcToProto(vc),
	}, nil
}

func (s *Server) VerifyCredential(_ context.Context, req *anchorv1.VerifyCredentialRequest) (*anchorv1.VerifyCredentialResponse, error) {
	var vc openanchor.VerifiableCredential
	if err := json.Unmarshal([]byte(req.CredentialJson), &vc); err != nil {
		return nil, fmt.Errorf("invalid credential JSON: %w", err)
	}

	result, err := s.svc.VerifyCredential(&vc)
	if err != nil {
		return nil, mapError(err)
	}
	return &anchorv1.VerifyCredentialResponse{
		Valid:   result.Valid,
		Issuer:  result.Issuer,
		Message: result.Message,
	}, nil
}

func (s *Server) ListCredentials(_ context.Context, req *anchorv1.ListCredentialsRequest) (*anchorv1.ListCredentialsResponse, error) {
	page, perPage := paginationFromProto(req.Pagination)
	creds, total, err := s.svc.ListCredentials(req.Type, page, perPage)
	if err != nil {
		return nil, mapError(err)
	}

	protoCreds := make([]*anchorv1.VerifiableCredential, 0, len(creds))
	for i := range creds {
		protoCreds = append(protoCreds, vcToProto(&creds[i]))
	}

	return &anchorv1.ListCredentialsResponse{
		Credentials: protoCreds,
		Pagination:  paginationToProto(page, perPage, total),
	}, nil
}

func vcToProto(vc *openanchor.VerifiableCredential) *anchorv1.VerifiableCredential {
	if vc == nil {
		return nil
	}

	subjectJSON, _ := json.Marshal(vc.CredentialSubject)

	proto := &anchorv1.VerifiableCredential{
		Id:                    vc.ID,
		Context:               vc.Context,
		Type:                  vc.Type,
		Issuer:                vc.Issuer,
		IssuanceDate:          vc.IssuanceDate,
		ExpirationDate:        vc.ExpirationDate,
		CredentialSubjectJson: string(subjectJSON),
	}
	if vc.Proof != nil {
		proto.Proof = &anchorv1.CredentialProof{
			Type:               vc.Proof.Type,
			Created:            vc.Proof.Created,
			VerificationMethod: vc.Proof.VerificationMethod,
			ProofPurpose:       vc.Proof.ProofPurpose,
			ProofValue:         vc.Proof.ProofValue,
		}
	}
	return proto
}
