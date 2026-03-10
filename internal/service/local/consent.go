package local

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/FibrinLab/open-nucleus/internal/service"
	"github.com/FibrinLab/open-nucleus/pkg/consent"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

// localConsentService wraps ConsentManager for the service interface.
type localConsentService struct {
	mgr        *consent.Manager
	issuerDID  string
	issuerKey  ed25519.PrivateKey
}

// NewLocalConsentService creates a ConsentService backed by the local ConsentManager.
func NewLocalConsentService(mgr *consent.Manager, issuerDID string, issuerKey ed25519.PrivateKey) service.ConsentService {
	return &localConsentService{mgr: mgr, issuerDID: issuerDID, issuerKey: issuerKey}
}

func (s *localConsentService) CheckAccess(_ context.Context, patientID, performerID, role string) (*service.ConsentAccessDecision, error) {
	decision, err := s.mgr.CheckAccess(patientID, performerID, role)
	if err != nil {
		return nil, err
	}
	return &service.ConsentAccessDecision{
		Allowed:   decision.Allowed,
		ConsentID: decision.ConsentID,
		Reason:    decision.Reason,
	}, nil
}

func (s *localConsentService) GrantConsent(_ context.Context, patientID, performerID, scope string, period *service.ConsentPeriod, category string) (*service.ConsentGrantResponse, error) {
	var p *consent.Period
	if period != nil {
		start, err := time.Parse(time.RFC3339, period.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid period start: %w", err)
		}
		end, err := time.Parse(time.RFC3339, period.End)
		if err != nil {
			return nil, fmt.Errorf("invalid period end: %w", err)
		}
		p = &consent.Period{Start: start, End: end}
	}

	row, hash, err := s.mgr.GrantConsent(patientID, performerID, scope, p, category)
	if err != nil {
		return nil, err
	}
	return &service.ConsentGrantResponse{
		ConsentID:  row.ID,
		CommitHash: hash,
		Status:     row.Status,
	}, nil
}

func (s *localConsentService) RevokeConsent(_ context.Context, consentID string) error {
	return s.mgr.RevokeConsent(consentID)
}

func (s *localConsentService) ListConsentsForPatient(_ context.Context, patientID string, page, perPage int) (*service.ConsentListResponse, error) {
	rows, pg, err := s.mgr.ListConsentsForPatient(patientID, fhir.PaginationOpts{Page: page, PerPage: perPage})
	if err != nil {
		return nil, err
	}

	consents := make([]service.ConsentSummary, len(rows))
	for i, r := range rows {
		consents[i] = service.ConsentSummary{
			ID:            r.ID,
			PatientID:     r.PatientID,
			Status:        r.Status,
			ScopeCode:     r.ScopeCode,
			PerformerID:   r.PerformerID,
			ProvisionType: r.ProvisionType,
			PeriodStart:   r.PeriodStart,
			PeriodEnd:     r.PeriodEnd,
			Category:      r.Category,
			LastUpdated:   r.LastUpdated,
		}
	}

	resp := &service.ConsentListResponse{
		Consents: consents,
	}
	if pg != nil {
		resp.Pagination = &service.PaginationMeta{
			Page:       pg.Page,
			PerPage:    pg.PerPage,
			Total:      pg.Total,
			TotalPages: pg.TotalPages,
		}
	}
	return resp, nil
}

func (s *localConsentService) IssueConsentVC(_ context.Context, consentID string) (*service.ConsentVCResponse, error) {
	if s.issuerKey == nil {
		return nil, fmt.Errorf("consent VC issuance requires an issuer key")
	}

	vc, err := s.mgr.IssueConsentVC(consentID, s.issuerDID, s.issuerKey)
	if err != nil {
		return nil, err
	}
	return &service.ConsentVCResponse{
		VerifiableCredential: vc,
	}, nil
}
