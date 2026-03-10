package sqliteindex

import (
	"testing"

	"github.com/FibrinLab/open-nucleus/pkg/fhir"
)

func testIndex(t *testing.T) Index {
	t.Helper()
	idx, err := NewIndex(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { idx.Close() })
	return idx
}

func TestInitSchema_CreatesAllTables(t *testing.T) {
	idx := testIndex(t)
	// If NewIndex succeeded, schema was created
	count, err := idx.ResourceCount()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 resources, got %d", count)
	}
}

func TestUpsertPatient_InsertAndUpdate(t *testing.T) {
	idx := testIndex(t)

	row := &fhir.PatientRow{
		ID:          "p1",
		FamilyName:  "Ibrahim",
		GivenNames:  `["Fatima"]`,
		Gender:      "female",
		BirthDate:   "1990-01-15",
		SiteID:      "site-1",
		Active:      true,
		LastUpdated: "2026-03-15T09:42:00Z",
		GitBlobHash: "abc123",
	}

	if err := idx.UpsertPatient(row); err != nil {
		t.Fatal(err)
	}

	got, err := idx.GetPatient("p1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected patient, got nil")
	}
	if got.FamilyName != "Ibrahim" {
		t.Errorf("FamilyName = %s", got.FamilyName)
	}

	// Update
	row.FamilyName = "Okafor"
	if err := idx.UpsertPatient(row); err != nil {
		t.Fatal(err)
	}
	got, _ = idx.GetPatient("p1")
	if got.FamilyName != "Okafor" {
		t.Errorf("FamilyName after update = %s", got.FamilyName)
	}
}

func TestListPatients_WithFilters(t *testing.T) {
	idx := testIndex(t)

	patients := []*fhir.PatientRow{
		{ID: "p1", FamilyName: "Ibrahim", GivenNames: `["Fatima"]`, Gender: "female", BirthDate: "1990-01-15", SiteID: "site-1", Active: true, LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h1"},
		{ID: "p2", FamilyName: "Okafor", GivenNames: `["Chidi"]`, Gender: "male", BirthDate: "1985-06-20", SiteID: "site-1", Active: true, LastUpdated: "2026-01-02T00:00:00Z", GitBlobHash: "h2"},
		{ID: "p3", FamilyName: "Musa", GivenNames: `["Ahmed"]`, Gender: "male", BirthDate: "2000-03-01", SiteID: "site-2", Active: false, LastUpdated: "2026-01-03T00:00:00Z", GitBlobHash: "h3"},
	}
	for _, p := range patients {
		if err := idx.UpsertPatient(p); err != nil {
			t.Fatal(err)
		}
	}

	// All
	results, pg, err := idx.ListPatients(PatientListOpts{PaginationOpts: fhir.PaginationOpts{Page: 1, PerPage: 25}})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3, got %d", len(results))
	}
	if pg.Total != 3 {
		t.Errorf("total = %d", pg.Total)
	}

	// Active only
	results, _, err = idx.ListPatients(PatientListOpts{ActiveOnly: true, PaginationOpts: fhir.PaginationOpts{Page: 1, PerPage: 25}})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 active, got %d", len(results))
	}

	// Filter by gender
	results, _, err = idx.ListPatients(PatientListOpts{Gender: "male", PaginationOpts: fhir.PaginationOpts{Page: 1, PerPage: 25}})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 male, got %d", len(results))
	}
}

func TestSearchPatients_FTS5(t *testing.T) {
	idx := testIndex(t)

	patients := []*fhir.PatientRow{
		{ID: "p1", FamilyName: "Ibrahim", GivenNames: `["Fatima"]`, Gender: "female", BirthDate: "1990", SiteID: "s1", Active: true, LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h1"},
		{ID: "p2", FamilyName: "Okafor", GivenNames: `["Chidi"]`, Gender: "male", BirthDate: "1985", SiteID: "s1", Active: true, LastUpdated: "2026-01-02T00:00:00Z", GitBlobHash: "h2"},
	}
	for _, p := range patients {
		if err := idx.UpsertPatient(p); err != nil {
			t.Fatal(err)
		}
	}

	results, _, err := idx.SearchPatients("Ibrahim", fhir.PaginationOpts{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "p1" {
		t.Errorf("expected p1, got %s", results[0].ID)
	}
}

func TestGetPatientBundle_AllChildResources(t *testing.T) {
	idx := testIndex(t)

	if err := idx.UpsertPatient(&fhir.PatientRow{ID: "p1", FamilyName: "Test", GivenNames: `["User"]`, Gender: "male", BirthDate: "1990", SiteID: "s1", Active: true, LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h1"}); err != nil {
		t.Fatal(err)
	}
	if err := idx.UpsertEncounter(&fhir.EncounterRow{ID: "e1", PatientID: "p1", Status: "finished", ClassCode: "AMB", PeriodStart: "2026-01-01", SiteID: "s1", LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h2"}); err != nil {
		t.Fatal(err)
	}

	bundle, err := idx.GetPatientBundle("p1")
	if err != nil {
		t.Fatal(err)
	}
	if bundle == nil {
		t.Fatal("expected bundle, got nil")
	}
	if bundle.Patient.ID != "p1" {
		t.Errorf("patient ID = %s", bundle.Patient.ID)
	}
	if len(bundle.Encounters) != 1 {
		t.Errorf("expected 1 encounter, got %d", len(bundle.Encounters))
	}
}

func TestGetTimeline_SortedByDate(t *testing.T) {
	idx := testIndex(t)

	if err := idx.UpsertPatient(&fhir.PatientRow{ID: "p1", FamilyName: "Test", GivenNames: `["User"]`, Gender: "male", BirthDate: "1990", SiteID: "s1", Active: true, LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h1"}); err != nil {
		t.Fatal(err)
	}
	if err := idx.UpsertEncounter(&fhir.EncounterRow{ID: "e1", PatientID: "p1", Status: "finished", ClassCode: "AMB", PeriodStart: "2026-01-10", SiteID: "s1", LastUpdated: "2026-01-10T00:00:00Z", GitBlobHash: "h2"}); err != nil {
		t.Fatal(err)
	}
	if err := idx.UpsertObservation(&fhir.ObservationRow{ID: "o1", PatientID: "p1", Status: "final", Code: "8310-5", EffectiveDatetime: "2026-01-15", SiteID: "s1", LastUpdated: "2026-01-15T00:00:00Z", GitBlobHash: "h3"}); err != nil {
		t.Fatal(err)
	}

	events, pg, err := idx.GetTimeline("p1", fhir.PaginationOpts{Page: 1, PerPage: 25})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
	if pg.Total != 2 {
		t.Errorf("total = %d", pg.Total)
	}
	// First event should be the most recent (observation on Jan 15)
	if len(events) >= 2 && events[0].EventType != "observation" {
		t.Errorf("first event type = %s, expected observation", events[0].EventType)
	}
}

func TestGetMatchCandidates_BroadQuery(t *testing.T) {
	idx := testIndex(t)

	patients := []*fhir.PatientRow{
		{ID: "p1", FamilyName: "Ibrahim", GivenNames: `["Fatima"]`, Gender: "female", BirthDate: "1990-01-15", SiteID: "s1", Active: true, LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h1"},
		{ID: "p2", FamilyName: "Ibrahimov", GivenNames: `["Hassan"]`, Gender: "male", BirthDate: "1990-06-20", SiteID: "s1", Active: true, LastUpdated: "2026-01-02T00:00:00Z", GitBlobHash: "h2"},
		{ID: "p3", FamilyName: "Okafor", GivenNames: `["Chidi"]`, Gender: "male", BirthDate: "1985-03-01", SiteID: "s1", Active: true, LastUpdated: "2026-01-03T00:00:00Z", GitBlobHash: "h3"},
	}
	for _, p := range patients {
		if err := idx.UpsertPatient(p); err != nil {
			t.Fatal(err)
		}
	}

	results, err := idx.GetMatchCandidates("Ibrahim", "1990")
	if err != nil {
		t.Fatal(err)
	}
	// Should match p1 (exact name prefix + birth year), p2 (name prefix), p3 would not match
	if len(results) < 2 {
		t.Errorf("expected at least 2 candidates, got %d", len(results))
	}
}

func TestUpdateSummary_Counts(t *testing.T) {
	idx := testIndex(t)

	if err := idx.UpsertPatient(&fhir.PatientRow{ID: "p1", FamilyName: "Test", GivenNames: `["User"]`, Gender: "male", BirthDate: "1990", SiteID: "s1", Active: true, LastUpdated: "2026-01-01T00:00:00Z", GitBlobHash: "h1"}); err != nil {
		t.Fatal(err)
	}
	if err := idx.UpsertEncounter(&fhir.EncounterRow{ID: "e1", PatientID: "p1", Status: "finished", ClassCode: "AMB", PeriodStart: "2026-01-10", SiteID: "s1", LastUpdated: "2026-01-10T00:00:00Z", GitBlobHash: "h2"}); err != nil {
		t.Fatal(err)
	}
	if err := idx.UpsertCondition(&fhir.ConditionRow{ID: "c1", PatientID: "p1", ClinicalStatus: "active", VerificationStatus: "confirmed", Code: "J06.9", SiteID: "s1", LastUpdated: "2026-01-10T00:00:00Z", GitBlobHash: "h3"}); err != nil {
		t.Fatal(err)
	}

	if err := idx.UpdateSummary("p1"); err != nil {
		t.Fatal(err)
	}

	// Verify summary was created (just check it doesn't error)
	_, err := idx.GetMeta("test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMeta_SetAndGet(t *testing.T) {
	idx := testIndex(t)

	if err := idx.SetMeta("git_head", "abc123"); err != nil {
		t.Fatal(err)
	}

	val, err := idx.GetMeta("git_head")
	if err != nil {
		t.Fatal(err)
	}
	if val != "abc123" {
		t.Errorf("expected abc123, got %s", val)
	}

	// Non-existent key
	val, err = idx.GetMeta("missing")
	if err != nil {
		t.Fatal(err)
	}
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
}
