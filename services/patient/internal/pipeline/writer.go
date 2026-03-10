package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/FibrinLab/open-nucleus/pkg/envelope"
	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	"github.com/FibrinLab/open-nucleus/pkg/gitstore"
	"github.com/FibrinLab/open-nucleus/pkg/sqliteindex"
)

// WriteResult holds the result of a write operation.
type WriteResult struct {
	ResourceType string
	ResourceID   string
	PatientID    string
	FHIRJson     []byte
	CommitHash   string
	CommitMsg    string
	Timestamp    time.Time
}

// BatchItem represents a single resource in a batch operation.
type BatchItem struct {
	ResourceType string
	FHIRJson     []byte
}

// BatchResult holds the result of a batch write.
type BatchResult struct {
	Results   []BatchItemResult
	CommitHash string
	CommitMsg  string
	Timestamp  time.Time
}

// BatchItemResult holds the result for one item in a batch.
type BatchItemResult struct {
	ResourceType string
	ResourceID   string
	Success      bool
	Error        string
}

// MutationContext holds metadata about who/where/when a mutation is happening.
type MutationContext struct {
	PractitionerID string
	NodeID         string
	SiteID         string
	Timestamp      time.Time
}

// Writer implements the validate → git commit → sqlite upsert pipeline per spec §3.1.
type Writer struct {
	mu          sync.Mutex
	git         gitstore.Store
	idx         sqliteindex.Index
	keys        envelope.KeyManager // optional; nil = no encryption
	lockTimeout time.Duration
}

// NewWriter creates a new write pipeline.
func NewWriter(git gitstore.Store, idx sqliteindex.Index, lockTimeout time.Duration) *Writer {
	if lockTimeout == 0 {
		lockTimeout = 5 * time.Second
	}
	return &Writer{
		git:         git,
		idx:         idx,
		lockTimeout: lockTimeout,
	}
}

// WithEncryption attaches a KeyManager for per-patient encryption at rest.
// When set, FHIR JSON is encrypted before writing to Git.
func (w *Writer) WithEncryption(km envelope.KeyManager) *Writer {
	w.keys = km
	return w
}

// encryptForGit encrypts data if a KeyManager is configured.
// Patient resources use the patient's key; non-patient resources use the system key.
func (w *Writer) encryptForGit(patientID string, data []byte) ([]byte, error) {
	if w.keys == nil {
		return data, nil
	}
	keyID := patientID
	if keyID == "" {
		keyID = envelope.SystemKeyID
	}
	return w.keys.Encrypt(keyID, data)
}

// DecryptFromGit decrypts data if a KeyManager is configured.
func (w *Writer) DecryptFromGit(patientID string, data []byte) ([]byte, error) {
	if w.keys == nil {
		return data, nil
	}
	keyID := patientID
	if keyID == "" {
		keyID = envelope.SystemKeyID
	}
	return w.keys.Decrypt(keyID, data)
}

// DestroyPatientKey destroys the encryption key for a patient, making
// their data permanently unreadable (crypto-erasure). No-op if encryption is not configured.
func (w *Writer) DestroyPatientKey(patientID string) error {
	if w.keys == nil {
		return nil
	}
	return w.keys.DestroyKey(patientID)
}

// Write performs a single resource write operation.
func (w *Writer) Write(ctx context.Context, op, resourceType, patientID string, fhirJSON []byte, mutCtx MutationContext) (*WriteResult, error) {
	// 1. Validate (with profile support)
	errs := fhir.ValidateWithProfile(resourceType, fhirJSON)
	if len(errs) > 0 {
		errJSON, _ := json.Marshal(errs)
		return nil, &ValidationError{FieldErrors: errs, Message: string(errJSON)}
	}

	// 2. Assign UUID if CREATE
	var resourceID string
	var err error
	if op == fhir.OpCreate {
		fhirJSON, resourceID, err = fhir.AssignID(fhirJSON)
		if err != nil {
			return nil, fmt.Errorf("assign ID: %w", err)
		}
	} else {
		resourceID, err = fhir.GetID(fhirJSON)
		if err != nil {
			return nil, fmt.Errorf("get ID: %w", err)
		}
	}

	// For Patient CREATE, the patientID is the resourceID
	if resourceType == fhir.ResourcePatient {
		patientID = resourceID
	}

	// 3. Set meta fields
	now := mutCtx.Timestamp
	if now.IsZero() {
		now = time.Now().UTC()
	}
	shortHash := resourceID[:8]
	if len(resourceID) < 8 {
		shortHash = resourceID
	}
	fhirJSON, err = fhir.SetMeta(fhirJSON, now, shortHash, mutCtx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("set meta: %w", err)
	}

	// 4. Extract search fields from cleartext BEFORE encryption
	//    (SQLite index gets extracted fields only, never the full JSON)

	// 5. Encrypt for Git storage (if encryption is configured)
	gitData, err := w.encryptForGit(patientID, fhirJSON)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	// 6. Acquire write lock with timeout
	if err := w.acquireLock(ctx); err != nil {
		return nil, err
	}
	defer w.mu.Unlock()

	// 7. Compute git path
	gitPath := fhir.GitPath(resourceType, patientID, resourceID)

	// 8. Write to git (encrypted if KeyManager is set)
	commitMsg := gitstore.CommitMessage{
		ResourceType: resourceType,
		Operation:    op,
		ResourceID:   resourceID,
		NodeID:       mutCtx.NodeID,
		Author:       mutCtx.PractitionerID,
		SiteID:       mutCtx.SiteID,
		Timestamp:    now,
	}

	commitHash, err := w.git.WriteAndCommit(gitPath, gitData, commitMsg)
	if err != nil {
		// Rollback on git failure
		w.git.Rollback()
		return nil, fmt.Errorf("git write failed: %w", err)
	}

	// 9. Extract fields from cleartext + upsert SQLite (fhir_json not stored)
	if sqErr := w.upsertIndex(resourceType, patientID, mutCtx.SiteID, commitHash, fhirJSON); sqErr != nil {
		// Git succeeded but SQLite failed — return success with warning
		// Data is safe in Git per spec §11.3
		fmt.Printf("WARNING: SQLite upsert failed after git commit %s: %v\n", commitHash, sqErr)
	}

	// 7b. Auto-generate Provenance (skip for Provenance itself to prevent recursion)
	if resourceType != fhir.ResourceProvenance {
		provJSON, provID, provErr := fhir.GenerateProvenance(fhir.ProvenanceContext{
			TargetResourceType: resourceType,
			TargetResourceID:   resourceID,
			Activity:           op,
			PractitionerID:     mutCtx.PractitionerID,
			DeviceID:           mutCtx.NodeID,
			SiteID:             mutCtx.SiteID,
			Recorded:           now,
		})
		if provErr != nil {
			fmt.Printf("WARNING: provenance generation failed: %v\n", provErr)
		} else {
			provPath := fhir.GitPath(fhir.ResourceProvenance, patientID, provID)
			provMsg := gitstore.CommitMessage{
				ResourceType: fhir.ResourceProvenance,
				Operation:    fhir.OpCreate,
				ResourceID:   provID,
				NodeID:       mutCtx.NodeID,
				Author:       mutCtx.PractitionerID,
				SiteID:       mutCtx.SiteID,
				Timestamp:    now,
			}
			if _, provWriteErr := w.git.WriteAndCommit(provPath, provJSON, provMsg); provWriteErr != nil {
				fmt.Printf("WARNING: provenance write failed: %v\n", provWriteErr)
			}
		}
	}

	// 8. Update patient_summaries
	if patientID != "" {
		if err := w.idx.UpdateSummary(patientID); err != nil {
			fmt.Printf("WARNING: summary update failed for patient %s: %v\n", patientID, err)
		}
	}

	return &WriteResult{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		PatientID:    patientID,
		FHIRJson:     fhirJSON,
		CommitHash:   commitHash,
		CommitMsg:    commitMsg.Format(),
		Timestamp:    now,
	}, nil
}

// Delete performs a soft-delete per spec §3.4.
func (w *Writer) Delete(ctx context.Context, resourceType, patientID, resourceID string, mutCtx MutationContext) (*WriteResult, error) {
	// Read existing resource from git (may be encrypted)
	gitPath := fhir.GitPath(resourceType, patientID, resourceID)
	raw, err := w.git.Read(gitPath)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Decrypt if needed
	existing, err := w.DecryptFromGit(patientID, raw)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	// Apply soft delete mutations
	deleted, err := fhir.ApplySoftDelete(resourceType, existing)
	if err != nil {
		return nil, fmt.Errorf("soft delete: %w", err)
	}

	// Write as an UPDATE with the soft-deleted content
	return w.Write(ctx, fhir.OpDelete, resourceType, patientID, deleted, mutCtx)
}

// WriteBatch writes multiple resources in a single git commit per spec §8.1.
func (w *Writer) WriteBatch(ctx context.Context, patientID string, items []BatchItem, mutCtx MutationContext, atomic bool) (*BatchResult, error) {
	now := mutCtx.Timestamp
	if now.IsZero() {
		now = time.Now().UTC()
	}

	type prepared struct {
		resourceType string
		resourceID   string
		fhirJSON     []byte
		gitPath      string
	}

	// Validate all first if atomic
	var preparedItems []prepared
	var results []BatchItemResult

	for _, item := range items {
		errs := fhir.ValidateWithProfile(item.ResourceType, item.FHIRJson)
		if len(errs) > 0 {
			if atomic {
				errJSON, _ := json.Marshal(errs)
				return nil, &ValidationError{FieldErrors: errs, Message: string(errJSON)}
			}
			results = append(results, BatchItemResult{
				ResourceType: item.ResourceType,
				Success:      false,
				Error:        fmt.Sprintf("validation failed: %d errors", len(errs)),
			})
			continue
		}

		jsonData, resID, err := fhir.AssignID(item.FHIRJson)
		if err != nil {
			if atomic {
				return nil, fmt.Errorf("assign ID: %w", err)
			}
			results = append(results, BatchItemResult{
				ResourceType: item.ResourceType,
				Success:      false,
				Error:        err.Error(),
			})
			continue
		}

		shortHash := resID[:8]
		if len(resID) < 8 {
			shortHash = resID
		}
		jsonData, err = fhir.SetMeta(jsonData, now, shortHash, mutCtx.SiteID)
		if err != nil {
			if atomic {
				return nil, fmt.Errorf("set meta: %w", err)
			}
			results = append(results, BatchItemResult{
				ResourceType: item.ResourceType,
				ResourceID:   resID,
				Success:      false,
				Error:        err.Error(),
			})
			continue
		}

		pid := patientID
		if item.ResourceType == fhir.ResourcePatient {
			pid = resID
		}
		gitPath := fhir.GitPath(item.ResourceType, pid, resID)

		preparedItems = append(preparedItems, prepared{
			resourceType: item.ResourceType,
			resourceID:   resID,
			fhirJSON:     jsonData,
			gitPath:      gitPath,
		})
	}

	if len(preparedItems) == 0 {
		return &BatchResult{Results: results}, nil
	}

	// Acquire lock
	if err := w.acquireLock(ctx); err != nil {
		return nil, err
	}
	defer w.mu.Unlock()

	// Write all files and create a single commit
	var lastCommitHash string
	for _, p := range preparedItems {
		// Encrypt for Git storage
		pid := patientID
		if p.resourceType == fhir.ResourcePatient {
			pid = p.resourceID
		}
		gitData, encErr := w.encryptForGit(pid, p.fhirJSON)
		if encErr != nil {
			return nil, fmt.Errorf("encrypt batch item: %w", encErr)
		}

		commitMsg := gitstore.CommitMessage{
			ResourceType: p.resourceType,
			Operation:    fhir.OpCreate,
			ResourceID:   p.resourceID,
			NodeID:       mutCtx.NodeID,
			Author:       mutCtx.PractitionerID,
			SiteID:       mutCtx.SiteID,
			Timestamp:    now,
		}

		hash, err := w.git.WriteAndCommit(p.gitPath, gitData, commitMsg)
		if err != nil {
			w.git.Rollback()
			return nil, fmt.Errorf("git write failed: %w", err)
		}
		lastCommitHash = hash

		if sqErr := w.upsertIndex(p.resourceType, pid, mutCtx.SiteID, hash, p.fhirJSON); sqErr != nil {
			fmt.Printf("WARNING: batch SQLite upsert failed: %v\n", sqErr)
		}

		results = append(results, BatchItemResult{
			ResourceType: p.resourceType,
			ResourceID:   p.resourceID,
			Success:      true,
		})
	}

	// Update summary
	if patientID != "" {
		w.idx.UpdateSummary(patientID)
	}

	return &BatchResult{
		Results:    results,
		CommitHash: lastCommitHash,
		Timestamp:  now,
	}, nil
}

// RebuildIndex drops and rebuilds the SQLite index from Git per spec §9.1.
func (w *Writer) RebuildIndex() (int, string, error) {
	if err := w.acquireLockDirect(); err != nil {
		return 0, "", err
	}
	defer w.mu.Unlock()

	// Get current head
	head, err := w.git.Head()
	if err != nil {
		return 0, "", fmt.Errorf("get HEAD: %w", err)
	}

	count := 0
	err = w.git.TreeWalk(func(path string, data []byte) error {
		// Decrypt if encryption is configured
		patientID := extractPatientIDFromPath(path)
		cleartext, decErr := w.DecryptFromGit(patientID, data)
		if decErr != nil {
			// Key destroyed (crypto-erasure) or corrupted — skip silently
			return nil
		}

		rt, err := fhir.GetResourceType(cleartext)
		if err != nil {
			return nil // skip non-FHIR files
		}

		siteID := "" // Will be extracted from meta if available
		if sqErr := w.upsertIndex(rt, patientID, siteID, head, cleartext); sqErr != nil {
			return fmt.Errorf("upsert %s: %w", path, sqErr)
		}
		count++
		return nil
	})

	if err != nil {
		return 0, "", err
	}

	w.idx.SetMeta("git_head", head)
	w.idx.SetMeta("resource_count", fmt.Sprintf("%d", count))

	return count, head, nil
}

func (w *Writer) upsertIndex(resourceType, patientID, siteID, commitHash string, fhirJSON []byte) error {
	switch resourceType {
	case fhir.ResourcePatient:
		row, err := fhir.ExtractPatientFields(fhirJSON, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertPatient(row)
	case fhir.ResourceEncounter:
		row, err := fhir.ExtractEncounterFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertEncounter(row)
	case fhir.ResourceObservation:
		row, err := fhir.ExtractObservationFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertObservation(row)
	case fhir.ResourceCondition:
		row, err := fhir.ExtractConditionFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertCondition(row)
	case fhir.ResourceMedicationRequest:
		row, err := fhir.ExtractMedicationRequestFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertMedicationRequest(row)
	case fhir.ResourceAllergyIntolerance:
		row, err := fhir.ExtractAllergyIntoleranceFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertAllergyIntolerance(row)
	case fhir.ResourceFlag:
		row, err := fhir.ExtractFlagFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertFlag(row)
	case fhir.ResourceImmunization:
		row, err := fhir.ExtractImmunizationFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertImmunization(row)
	case fhir.ResourceProcedure:
		row, err := fhir.ExtractProcedureFields(fhirJSON, patientID, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertProcedure(row)
	case fhir.ResourcePractitioner:
		row, err := fhir.ExtractPractitionerFields(fhirJSON, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertPractitioner(row)
	case fhir.ResourceOrganization:
		row, err := fhir.ExtractOrganizationFields(fhirJSON, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertOrganization(row)
	case fhir.ResourceLocation:
		row, err := fhir.ExtractLocationFields(fhirJSON, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertLocation(row)
	case fhir.ResourceMeasureReport:
		row, err := fhir.ExtractMeasureReportFields(fhirJSON, siteID, commitHash)
		if err != nil {
			return err
		}
		return w.idx.UpsertMeasureReport(row)
	default:
		return nil
	}
}

func (w *Writer) acquireLock(ctx context.Context) error {
	lockCtx, cancel := context.WithTimeout(ctx, w.lockTimeout)
	defer cancel()

	ch := make(chan struct{}, 1)
	go func() {
		w.mu.Lock()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
		return nil
	case <-lockCtx.Done():
		return fmt.Errorf("write lock timeout: %w", lockCtx.Err())
	}
}

func (w *Writer) acquireLockDirect() error {
	w.mu.Lock()
	return nil
}

// extractPatientIDFromPath extracts the patient ID from a git file path.
func extractPatientIDFromPath(path string) string {
	// paths are like: patients/{patient-id}/...
	if len(path) < 10 || path[:9] != "patients/" {
		return ""
	}
	rest := path[9:]
	for i, c := range rest {
		if c == '/' {
			return rest[:i]
		}
	}
	return ""
}

// ValidationError represents a FHIR validation failure.
type ValidationError struct {
	FieldErrors []fhir.FieldError
	Message     string
}

func (e *ValidationError) Error() string {
	return e.Message
}
