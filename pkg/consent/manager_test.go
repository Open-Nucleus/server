package consent

import (
	"log/slog"
	"testing"
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
)

// --- mock git store ---

type mockGit struct {
	files map[string][]byte
}

func newMockGit() *mockGit {
	return &mockGit{files: make(map[string][]byte)}
}

func (s *mockGit) WriteAndCommit(path string, data []byte, msg gitstore.CommitMessage) (string, error) {
	s.files[path] = append([]byte(nil), data...)
	return "test-hash", nil
}

func (s *mockGit) Read(path string) ([]byte, error) {
	data, ok := s.files[path]
	if !ok {
		return nil, nil
	}
	return data, nil
}

func (s *mockGit) LogPath(string, int) ([]gitstore.CommitInfo, error) { return nil, nil }
func (s *mockGit) Head() (string, error)                              { return "head", nil }
func (s *mockGit) TreeWalk(func(string, []byte) error) error          { return nil }
func (s *mockGit) Rollback() error                                    { return nil }

// --- mock index (embeds a real SQLite index for consent ops) ---

func setupTestManager(t *testing.T) (*Manager, *mockGit) {
	t.Helper()
	dir := t.TempDir()
	idx, err := sqliteindex.NewIndex(dir + "/test.db")
	if err != nil {
		t.Fatalf("sqliteindex.New: %v", err)
	}
	t.Cleanup(func() { idx.Close() })

	git := newMockGit()
	mgr := NewManager(idx, git, slog.Default())
	return mgr, git
}

func TestCheckAccess_AdminBypass(t *testing.T) {
	mgr, _ := setupTestManager(t)

	decision, err := mgr.CheckAccess("patient-1", "device-1", "site_administrator")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("admin should bypass consent")
	}
}

func TestCheckAccess_RegionalAdminBypass(t *testing.T) {
	mgr, _ := setupTestManager(t)

	decision, err := mgr.CheckAccess("patient-1", "device-1", "regional_administrator")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("regional admin should bypass consent")
	}
}

func TestCheckAccess_NoConsent(t *testing.T) {
	mgr, _ := setupTestManager(t)

	decision, err := mgr.CheckAccess("patient-1", "device-1", "physician")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if decision.Allowed {
		t.Fatal("should be denied without consent")
	}
}

func TestGrantAndCheckConsent(t *testing.T) {
	mgr, _ := setupTestManager(t)

	// Grant consent
	row, hash, err := mgr.GrantConsent("patient-1", "device-1", fhir.ConsentScopePatientPrivacy, nil, "")
	if err != nil {
		t.Fatalf("GrantConsent: %v", err)
	}
	if row == nil {
		t.Fatal("expected non-nil row")
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if row.Status != fhir.ConsentStatusActive {
		t.Errorf("status = %q, want %q", row.Status, fhir.ConsentStatusActive)
	}

	// Check access
	decision, err := mgr.CheckAccess("patient-1", "device-1", "physician")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !decision.Allowed {
		t.Fatal("should be allowed after granting consent")
	}
	if decision.ConsentID == "" {
		t.Fatal("expected consent ID in decision")
	}
}

func TestGrantConsent_WithPeriod(t *testing.T) {
	mgr, _ := setupTestManager(t)

	period := &Period{
		Start: time.Now().UTC().Add(-1 * time.Hour),
		End:   time.Now().UTC().Add(1 * time.Hour),
	}

	row, _, err := mgr.GrantConsent("patient-2", "device-2", fhir.ConsentScopePatientPrivacy, period, "npp")
	if err != nil {
		t.Fatalf("GrantConsent: %v", err)
	}
	if row.PeriodStart == nil {
		t.Fatal("expected non-nil PeriodStart")
	}
	if row.PeriodEnd == nil {
		t.Fatal("expected non-nil PeriodEnd")
	}
	if row.Category == nil || *row.Category != "npp" {
		t.Fatal("expected category 'npp'")
	}
}

func TestGrantEmergencyConsent(t *testing.T) {
	mgr, _ := setupTestManager(t)

	row, _, err := mgr.GrantEmergencyConsent("patient-3", "device-3")
	if err != nil {
		t.Fatalf("GrantEmergencyConsent: %v", err)
	}

	if row.Category == nil || *row.Category != fhir.ConsentCategoryEmrgOnly {
		t.Fatalf("category = %v, want %q", row.Category, fhir.ConsentCategoryEmrgOnly)
	}
	if row.PeriodEnd == nil {
		t.Fatal("emergency consent should have expiry")
	}
}

func TestRevokeConsent(t *testing.T) {
	mgr, _ := setupTestManager(t)

	// Grant
	row, _, err := mgr.GrantConsent("patient-4", "device-4", fhir.ConsentScopePatientPrivacy, nil, "")
	if err != nil {
		t.Fatalf("GrantConsent: %v", err)
	}

	// Revoke
	err = mgr.RevokeConsent(row.ID)
	if err != nil {
		t.Fatalf("RevokeConsent: %v", err)
	}

	// Check access should now be denied
	decision, err := mgr.CheckAccess("patient-4", "device-4", "physician")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if decision.Allowed {
		t.Fatal("should be denied after revoking consent")
	}
}

func TestRevokeConsent_NotFound(t *testing.T) {
	mgr, _ := setupTestManager(t)

	err := mgr.RevokeConsent("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent consent")
	}
}

func TestListConsentsForPatient(t *testing.T) {
	mgr, _ := setupTestManager(t)

	// Grant two consents for same patient
	mgr.GrantConsent("patient-5", "device-a", fhir.ConsentScopePatientPrivacy, nil, "")
	mgr.GrantConsent("patient-5", "device-b", fhir.ConsentScopeTreatment, nil, "")

	rows, pg, err := mgr.ListConsentsForPatient("patient-5", fhir.PaginationOpts{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("ListConsentsForPatient: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 consents, got %d", len(rows))
	}
	if pg == nil {
		t.Fatal("expected non-nil pagination")
	}
}

func TestListConsentsForPatient_Empty(t *testing.T) {
	mgr, _ := setupTestManager(t)

	rows, _, err := mgr.ListConsentsForPatient("no-consents-patient", fhir.PaginationOpts{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("ListConsentsForPatient: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("expected 0 consents, got %d", len(rows))
	}
}

func TestCheckAccess_DifferentPerformers(t *testing.T) {
	mgr, _ := setupTestManager(t)

	// Grant only to device-a
	mgr.GrantConsent("patient-6", "device-a", fhir.ConsentScopePatientPrivacy, nil, "")

	// device-a should have access
	d1, err := mgr.CheckAccess("patient-6", "device-a", "physician")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !d1.Allowed {
		t.Fatal("device-a should be allowed")
	}

	// device-b should not
	d2, err := mgr.CheckAccess("patient-6", "device-b", "physician")
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if d2.Allowed {
		t.Fatal("device-b should be denied")
	}
}
