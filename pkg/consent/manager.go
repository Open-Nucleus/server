// Package consent implements FHIR Consent-based access control.
//
// The ConsentManager handles consent lifecycle (grant, revoke, check) and
// integrates with the envelope KeyManager for per-provider key wrapping
// and the openanchor VC system for offline-verifiable consent proofs.
package consent

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/merge/openanchor"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
)

// AccessDecision represents the result of a consent check.
type AccessDecision struct {
	Allowed   bool   `json:"allowed"`
	ConsentID string `json:"consent_id,omitempty"`
	Reason    string `json:"reason"`
}

// Period represents a time range for consent validity.
type Period struct {
	Start time.Time
	End   time.Time
}

// Manager provides consent management operations.
type Manager struct {
	idx    sqliteindex.Index
	git    gitstore.Store
	logger *slog.Logger
}

// NewManager creates a ConsentManager.
func NewManager(idx sqliteindex.Index, git gitstore.Store, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}
	return &Manager{idx: idx, git: git, logger: logger}
}

// CheckAccess verifies whether a device/performer has an active consent grant
// for a patient. Admin roles bypass consent checks.
func (m *Manager) CheckAccess(patientID, performerID, role string) (*AccessDecision, error) {
	// Admin roles bypass consent
	if role == "site_administrator" || role == "regional_administrator" {
		return &AccessDecision{
			Allowed: true,
			Reason:  "admin role bypasses consent",
		}, nil
	}

	consent, err := m.idx.GetActiveConsent(patientID, performerID, fhir.ConsentScopePatientPrivacy)
	if err != nil {
		return nil, fmt.Errorf("consent: check access: %w", err)
	}

	if consent == nil {
		return &AccessDecision{
			Allowed: false,
			Reason:  "no active consent grant found",
		}, nil
	}

	return &AccessDecision{
		Allowed:   true,
		ConsentID: consent.ID,
		Reason:    "active consent grant",
	}, nil
}

// GrantConsent creates a FHIR Consent resource granting access.
func (m *Manager) GrantConsent(patientID, performerID, scopeCode string, period *Period, category string) (*fhir.ConsentRow, string, error) {
	now := time.Now().UTC()
	consentID := fmt.Sprintf("consent-%s-%s-%d", patientID[:min(8, len(patientID))], performerID[:min(8, len(performerID))], now.UnixMilli())

	consent := buildConsentResource(consentID, patientID, performerID, scopeCode, fhir.ConsentStatusActive, fhir.ConsentProvisionPermit, period, category, now)

	fhirJSON, err := json.Marshal(consent)
	if err != nil {
		return nil, "", fmt.Errorf("consent: marshal: %w", err)
	}

	// Write to Git
	gitPath := fhir.GitPath(fhir.ResourceConsent, patientID, consentID)
	commitHash, err := m.git.WriteAndCommit(gitPath, fhirJSON, gitstore.CommitMessage{
		ResourceType: fhir.ResourceConsent,
		Operation:    fhir.OpCreate,
		ResourceID:   consentID,
		Timestamp:    now,
	})
	if err != nil {
		return nil, "", fmt.Errorf("consent: git write: %w", err)
	}

	// Extract and index
	row, err := fhir.ExtractConsentFields(fhirJSON, commitHash)
	if err != nil {
		return nil, "", fmt.Errorf("consent: extract: %w", err)
	}

	if err := m.idx.UpsertConsent(row); err != nil {
		return nil, "", fmt.Errorf("consent: index: %w", err)
	}

	m.logger.Info("consent granted",
		"consent_id", consentID,
		"patient_id", patientID,
		"performer_id", performerID,
		"scope", scopeCode,
		"category", category,
	)

	return row, commitHash, nil
}

// GrantEmergencyConsent creates a time-limited (4h) emergency consent with category=emrgonly.
func (m *Manager) GrantEmergencyConsent(patientID, performerID string) (*fhir.ConsentRow, string, error) {
	period := &Period{
		Start: time.Now().UTC(),
		End:   time.Now().UTC().Add(4 * time.Hour),
	}

	row, hash, err := m.GrantConsent(patientID, performerID, fhir.ConsentScopePatientPrivacy, period, fhir.ConsentCategoryEmrgOnly)
	if err != nil {
		return nil, "", err
	}

	m.logger.Warn("BREAK-GLASS: emergency consent granted",
		"consent_id", row.ID,
		"patient_id", patientID,
		"performer_id", performerID,
		"expires", period.End.Format(time.RFC3339),
	)

	return row, hash, nil
}

// RevokeConsent marks a consent as inactive.
func (m *Manager) RevokeConsent(consentID string) error {
	existing, err := m.idx.GetConsent(consentID)
	if err != nil {
		return fmt.Errorf("consent: get for revoke: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("consent: not found: %s", consentID)
	}

	// Read from Git, update status, write back
	gitPath := fhir.GitPath(fhir.ResourceConsent, existing.PatientID, consentID)
	data, err := m.git.Read(gitPath)
	if err != nil {
		return fmt.Errorf("consent: git read: %w", err)
	}

	var resource map[string]any
	if err := json.Unmarshal(data, &resource); err != nil {
		return fmt.Errorf("consent: unmarshal: %w", err)
	}

	resource["status"] = fhir.ConsentStatusInactive
	now := time.Now().UTC()
	if meta, ok := resource["meta"].(map[string]any); ok {
		meta["lastUpdated"] = now.Format(time.RFC3339)
	}

	updatedJSON, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("consent: marshal update: %w", err)
	}

	commitHash, err := m.git.WriteAndCommit(gitPath, updatedJSON, gitstore.CommitMessage{
		ResourceType: fhir.ResourceConsent,
		Operation:    fhir.OpUpdate,
		ResourceID:   consentID,
		Timestamp:    now,
	})
	if err != nil {
		return fmt.Errorf("consent: git write revoke: %w", err)
	}

	// Update index
	existing.Status = fhir.ConsentStatusInactive
	existing.LastUpdated = now.Format(time.RFC3339)
	existing.GitBlobHash = commitHash
	if err := m.idx.UpsertConsent(existing); err != nil {
		return fmt.Errorf("consent: index revoke: %w", err)
	}

	m.logger.Info("consent revoked",
		"consent_id", consentID,
		"patient_id", existing.PatientID,
		"performer_id", existing.PerformerID,
	)

	return nil
}

// ListConsentsForPatient returns all consents for a patient.
func (m *Manager) ListConsentsForPatient(patientID string, opts fhir.PaginationOpts) ([]*fhir.ConsentRow, *fhir.Pagination, error) {
	return m.idx.ListConsentsForPatient(patientID, opts)
}

// IssueConsentVC creates an offline-verifiable credential for a consent grant.
func (m *Manager) IssueConsentVC(consentID string, issuerDID string, issuerKey ed25519.PrivateKey) (*openanchor.VerifiableCredential, error) {
	consent, err := m.idx.GetConsent(consentID)
	if err != nil {
		return nil, fmt.Errorf("consent: get for VC: %w", err)
	}
	if consent == nil {
		return nil, fmt.Errorf("consent: not found: %s", consentID)
	}

	validFrom := consent.LastUpdated
	validUntil := ""
	if consent.PeriodEnd != nil {
		validUntil = *consent.PeriodEnd
	}

	expirationDate := ""
	if consent.PeriodEnd != nil {
		expirationDate = *consent.PeriodEnd
	} else {
		expirationDate = time.Now().UTC().Add(365 * 24 * time.Hour).Format(time.RFC3339)
	}

	claims := openanchor.CredentialClaims{
		ID:    "urn:uuid:" + consent.ID,
		Types: []string{"VerifiableCredential", "ConsentGrant"},
		Subject: map[string]any{
			"type":        "ConsentGrant",
			"consentId":   consent.ID,
			"patientId":   consent.PatientID,
			"performerId": consent.PerformerID,
			"scope":       consent.ScopeCode,
			"validFrom":   validFrom,
			"validUntil":  validUntil,
		},
		ExpirationDate: expirationDate,
	}

	vc, err := openanchor.IssueCredentialLocal(claims, issuerDID, issuerKey)
	if err != nil {
		return nil, fmt.Errorf("consent: issue VC: %w", err)
	}

	return vc, nil
}

// VerifyConsentVC verifies a consent credential.
func (m *Manager) VerifyConsentVC(vc *openanchor.VerifiableCredential) (*openanchor.VerificationResult, error) {
	return openanchor.VerifyCredentialLocal(vc)
}

// buildConsentResource creates a FHIR Consent resource as a map.
func buildConsentResource(id, patientID, performerID, scopeCode, status, provisionType string, period *Period, category string, now time.Time) map[string]any {
	consent := map[string]any{
		"resourceType": fhir.ResourceConsent,
		"id":           id,
		"status":       status,
		"scope": map[string]any{
			"coding": []any{
				map[string]any{
					"system": "http://terminology.hl7.org/CodeSystem/consentscope",
					"code":   scopeCode,
				},
			},
		},
		"patient": map[string]any{
			"reference": "Patient/" + patientID,
		},
		"performer": []any{
			map[string]any{
				"reference": performerID,
			},
		},
		"provision": map[string]any{
			"type": provisionType,
		},
		"meta": map[string]any{
			"lastUpdated": now.Format(time.RFC3339),
		},
	}

	if period != nil {
		provision := consent["provision"].(map[string]any)
		provision["period"] = map[string]any{
			"start": period.Start.Format(time.RFC3339),
			"end":   period.End.Format(time.RFC3339),
		}
	}

	if category != "" {
		consent["category"] = []any{
			map[string]any{
				"coding": []any{
					map[string]any{
						"system": "http://terminology.hl7.org/CodeSystem/consentcategorycodes",
						"code":   category,
					},
				},
			},
		}
	}

	return consent
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
