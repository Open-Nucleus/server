package sqliteindex

import (
	"database/sql"
	"fmt"

	"github.com/FibrinLab/open-nucleus/pkg/fhir"
	_ "modernc.org/sqlite"
)

// Index provides SQLite query index operations.
type Index interface {
	UpsertPatient(row *fhir.PatientRow) error
	UpsertEncounter(row *fhir.EncounterRow) error
	UpsertObservation(row *fhir.ObservationRow) error
	UpsertCondition(row *fhir.ConditionRow) error
	UpsertMedicationRequest(row *fhir.MedicationRequestRow) error
	UpsertAllergyIntolerance(row *fhir.AllergyIntoleranceRow) error
	UpsertFlag(row *fhir.FlagRow) error
	UpsertImmunization(row *fhir.ImmunizationRow) error
	UpsertProcedure(row *fhir.ProcedureRow) error
	UpsertPractitioner(row *fhir.PractitionerRow) error
	UpsertOrganization(row *fhir.OrganizationRow) error
	UpsertLocation(row *fhir.LocationRow) error
	UpsertMeasureReport(row *fhir.MeasureReportRow) error

	GetPatient(id string) (*fhir.PatientRow, error)
	ListPatients(opts PatientListOpts) ([]*fhir.PatientRow, *fhir.Pagination, error)
	GetEncounter(patientID, id string) (*fhir.EncounterRow, error)
	ListEncounters(patientID string, opts fhir.PaginationOpts) ([]*fhir.EncounterRow, *fhir.Pagination, error)
	GetObservation(patientID, id string) (*fhir.ObservationRow, error)
	ListObservations(patientID string, opts ObservationListOpts) ([]*fhir.ObservationRow, *fhir.Pagination, error)
	GetCondition(patientID, id string) (*fhir.ConditionRow, error)
	ListConditions(patientID string, opts ConditionListOpts) ([]*fhir.ConditionRow, *fhir.Pagination, error)
	GetMedicationRequest(patientID, id string) (*fhir.MedicationRequestRow, error)
	ListMedicationRequests(patientID string, opts fhir.PaginationOpts) ([]*fhir.MedicationRequestRow, *fhir.Pagination, error)
	GetAllergyIntolerance(patientID, id string) (*fhir.AllergyIntoleranceRow, error)
	ListAllergyIntolerances(patientID string, opts fhir.PaginationOpts) ([]*fhir.AllergyIntoleranceRow, *fhir.Pagination, error)
	ListFlags(patientID string, opts fhir.PaginationOpts) ([]*fhir.FlagRow, *fhir.Pagination, error)

	GetImmunization(patientID, id string) (*fhir.ImmunizationRow, error)
	ListImmunizations(patientID string, opts fhir.PaginationOpts) ([]*fhir.ImmunizationRow, *fhir.Pagination, error)
	GetProcedure(patientID, id string) (*fhir.ProcedureRow, error)
	ListProcedures(patientID string, opts fhir.PaginationOpts) ([]*fhir.ProcedureRow, *fhir.Pagination, error)
	GetPractitioner(id string) (*fhir.PractitionerRow, error)
	ListPractitioners(opts fhir.PaginationOpts) ([]*fhir.PractitionerRow, *fhir.Pagination, error)
	GetOrganization(id string) (*fhir.OrganizationRow, error)
	ListOrganizations(opts fhir.PaginationOpts) ([]*fhir.OrganizationRow, *fhir.Pagination, error)
	GetLocation(id string) (*fhir.LocationRow, error)
	ListLocations(opts fhir.PaginationOpts) ([]*fhir.LocationRow, *fhir.Pagination, error)
	GetMeasureReport(id string) (*fhir.MeasureReportRow, error)
	ListMeasureReports(opts fhir.PaginationOpts) ([]*fhir.MeasureReportRow, *fhir.Pagination, error)

	// ID-only lookups (for FHIR REST API — no patient ID needed)
	GetEncounterByID(id string) (*fhir.EncounterRow, error)
	GetObservationByID(id string) (*fhir.ObservationRow, error)
	GetConditionByID(id string) (*fhir.ConditionRow, error)
	GetMedicationRequestByID(id string) (*fhir.MedicationRequestRow, error)
	GetAllergyIntoleranceByID(id string) (*fhir.AllergyIntoleranceRow, error)
	GetImmunizationByID(id string) (*fhir.ImmunizationRow, error)
	GetProcedureByID(id string) (*fhir.ProcedureRow, error)
	GetFlagByID(id string) (*fhir.FlagRow, error)

	// Blind index operations
	UpsertPatientNgrams(patientID string, field string, ngramHashes []string) error
	SearchPatientsByNgrams(ngramHashes []string, opts fhir.PaginationOpts) ([]*fhir.PatientRow, *fhir.Pagination, error)

	// Consent operations
	UpsertConsent(row *fhir.ConsentRow) error
	GetConsent(id string) (*fhir.ConsentRow, error)
	ListConsentsForPatient(patientID string, opts fhir.PaginationOpts) ([]*fhir.ConsentRow, *fhir.Pagination, error)
	GetActiveConsent(patientID, performerID, scopeCode string) (*fhir.ConsentRow, error)
	DeleteConsent(id string) error

	GetPatientBundle(patientID string) (*BundleResult, error)
	SearchPatients(query string, opts fhir.PaginationOpts) ([]*fhir.PatientRow, *fhir.Pagination, error)
	GetTimeline(patientID string, opts fhir.PaginationOpts) ([]fhir.TimelineEvent, *fhir.Pagination, error)
	GetMatchCandidates(familyName, birthYear string) ([]*fhir.PatientRow, error)
	UpdateSummary(patientID string) error

	DeletePatientData(patientID string) error

	GetMeta(key string) (string, error)
	SetMeta(key, value string) error
	ResourceCount() (int, error)

	Close() error
}

// PatientListOpts extends PaginationOpts with patient-specific filters.
type PatientListOpts struct {
	fhir.PaginationOpts
	Gender        string
	BirthDateFrom string
	BirthDateTo   string
	SiteID        string
	ActiveOnly    bool
}

// ObservationListOpts extends PaginationOpts with observation-specific filters.
type ObservationListOpts struct {
	fhir.PaginationOpts
	Code        string
	Category    string
	DateFrom    string
	DateTo      string
	EncounterID string
}

// ConditionListOpts extends PaginationOpts with condition-specific filters.
type ConditionListOpts struct {
	fhir.PaginationOpts
	ClinicalStatus string
	Category       string
	Code           string
}

// BundleResult holds all resources for a patient bundle.
type BundleResult struct {
	Patient             *fhir.PatientRow
	Encounters          []*fhir.EncounterRow
	Observations        []*fhir.ObservationRow
	Conditions          []*fhir.ConditionRow
	MedicationRequests  []*fhir.MedicationRequestRow
	AllergyIntolerances []*fhir.AllergyIntoleranceRow
	Flags               []*fhir.FlagRow
}

type sqliteIndex struct {
	db *sql.DB
}

// NewIndex opens a SQLite database and initialises the schema.
func NewIndex(dbPath string) (Index, error) {
	dsn := dbPath
	if dbPath == ":memory:" {
		dsn = "file::memory:?cache=shared"
	} else {
		dsn = dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_cache_size=-20000"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := InitSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &sqliteIndex{db: db}, nil
}

// NewIndexFromDB wraps an existing *sql.DB as an Index.
// The caller is responsible for schema initialization and closing the DB.
func NewIndexFromDB(db *sql.DB) Index {
	return &sqliteIndex{db: db}
}

func (idx *sqliteIndex) Close() error {
	return idx.db.Close()
}

// --- Upsert methods ---

func (idx *sqliteIndex) UpsertPatient(row *fhir.PatientRow) error {
	active := 1
	if !row.Active {
		active = 0
	}
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO patients (id, family_name, given_names, gender, birth_date, site_id, active, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.FamilyName, row.GivenNames, row.Gender, row.BirthDate, row.SiteID, active, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertEncounter(row *fhir.EncounterRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO encounters (id, patient_id, status, class_code, type_code, period_start, period_end, site_id, reason_code, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.Status, row.ClassCode, row.TypeCode, row.PeriodStart, row.PeriodEnd, row.SiteID, row.ReasonCode, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertObservation(row *fhir.ObservationRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO observations (id, patient_id, encounter_id, status, category, code, code_display, effective_datetime, value_quantity_value, value_quantity_unit, value_string, value_codeable_concept, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.EncounterID, row.Status, row.Category, row.Code, row.CodeDisplay, row.EffectiveDatetime, row.ValueQuantityValue, row.ValueQuantityUnit, row.ValueString, row.ValueCodeableConcept, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertCondition(row *fhir.ConditionRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO conditions (id, patient_id, clinical_status, verification_status, code, code_display, onset_datetime, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.ClinicalStatus, row.VerificationStatus, row.Code, row.CodeDisplay, row.OnsetDatetime, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertMedicationRequest(row *fhir.MedicationRequestRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO medication_requests (id, patient_id, status, intent, medication_code, medication_display, authored_on, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.Status, row.Intent, row.MedicationCode, row.MedicationDisplay, row.AuthoredOn, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertAllergyIntolerance(row *fhir.AllergyIntoleranceRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO allergy_intolerances (id, patient_id, clinical_status, verification_status, type, substance_code, substance_display, criticality, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.ClinicalStatus, row.VerificationStatus, row.Type, row.SubstanceCode, row.SubstanceDisplay, row.Criticality, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertFlag(row *fhir.FlagRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO flags (id, patient_id, status, category, code, period_start, period_end, generated_by, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.Status, row.Category, row.Code, row.PeriodStart, row.PeriodEnd, row.GeneratedBy, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertImmunization(row *fhir.ImmunizationRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO immunizations (id, patient_id, status, vaccine_code, vaccine_display, occurrence_datetime, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.Status, row.VaccineCode, row.VaccineDisplay, row.OccurrenceDatetime, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertProcedure(row *fhir.ProcedureRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO procedures (id, patient_id, status, code, code_display, performed_datetime, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.Status, row.Code, row.CodeDisplay, row.PerformedDatetime, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertPractitioner(row *fhir.PractitionerRow) error {
	active := 1
	if !row.Active {
		active = 0
	}
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO practitioners (id, family_name, given_names, active, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.FamilyName, row.GivenNames, active, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertOrganization(row *fhir.OrganizationRow) error {
	active := 1
	if !row.Active {
		active = 0
	}
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO organizations (id, name, type, active, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.Name, row.Type, active, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertLocation(row *fhir.LocationRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO locations (id, name, type, status, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.Name, row.Type, row.Status, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) UpsertMeasureReport(row *fhir.MeasureReportRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO measure_reports (id, status, type, period_start, period_end, reporter, site_id, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.Status, row.Type, row.PeriodStart, row.PeriodEnd, row.Reporter, row.SiteID, row.LastUpdated, row.GitBlobHash)
	return err
}

// --- Get methods ---

func (idx *sqliteIndex) GetPatient(id string) (*fhir.PatientRow, error) {
	row := idx.db.QueryRow(`SELECT id, family_name, given_names, gender, birth_date, site_id, active, last_updated, git_blob_hash FROM patients WHERE id = ?`, id)
	p := &fhir.PatientRow{}
	var active int
	err := row.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &p.Gender, &p.BirthDate, &p.SiteID, &active, &p.LastUpdated, &p.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	return p, nil
}

func (idx *sqliteIndex) GetEncounter(patientID, id string) (*fhir.EncounterRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, class_code, type_code, period_start, period_end, site_id, reason_code, last_updated, git_blob_hash FROM encounters WHERE id = ? AND patient_id = ?`, id, patientID)
	e := &fhir.EncounterRow{}
	err := row.Scan(&e.ID, &e.PatientID, &e.Status, &e.ClassCode, &e.TypeCode, &e.PeriodStart, &e.PeriodEnd, &e.SiteID, &e.ReasonCode, &e.LastUpdated, &e.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (idx *sqliteIndex) GetObservation(patientID, id string) (*fhir.ObservationRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, encounter_id, status, category, code, code_display, effective_datetime, value_quantity_value, value_quantity_unit, value_string, value_codeable_concept, site_id, last_updated, git_blob_hash FROM observations WHERE id = ? AND patient_id = ?`, id, patientID)
	o := &fhir.ObservationRow{}
	err := row.Scan(&o.ID, &o.PatientID, &o.EncounterID, &o.Status, &o.Category, &o.Code, &o.CodeDisplay, &o.EffectiveDatetime, &o.ValueQuantityValue, &o.ValueQuantityUnit, &o.ValueString, &o.ValueCodeableConcept, &o.SiteID, &o.LastUpdated, &o.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (idx *sqliteIndex) GetCondition(patientID, id string) (*fhir.ConditionRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, clinical_status, verification_status, code, code_display, onset_datetime, site_id, last_updated, git_blob_hash FROM conditions WHERE id = ? AND patient_id = ?`, id, patientID)
	c := &fhir.ConditionRow{}
	err := row.Scan(&c.ID, &c.PatientID, &c.ClinicalStatus, &c.VerificationStatus, &c.Code, &c.CodeDisplay, &c.OnsetDatetime, &c.SiteID, &c.LastUpdated, &c.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (idx *sqliteIndex) GetMedicationRequest(patientID, id string) (*fhir.MedicationRequestRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, intent, medication_code, medication_display, authored_on, site_id, last_updated, git_blob_hash FROM medication_requests WHERE id = ? AND patient_id = ?`, id, patientID)
	m := &fhir.MedicationRequestRow{}
	err := row.Scan(&m.ID, &m.PatientID, &m.Status, &m.Intent, &m.MedicationCode, &m.MedicationDisplay, &m.AuthoredOn, &m.SiteID, &m.LastUpdated, &m.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (idx *sqliteIndex) GetAllergyIntolerance(patientID, id string) (*fhir.AllergyIntoleranceRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, clinical_status, verification_status, type, substance_code, substance_display, criticality, site_id, last_updated, git_blob_hash FROM allergy_intolerances WHERE id = ? AND patient_id = ?`, id, patientID)
	a := &fhir.AllergyIntoleranceRow{}
	err := row.Scan(&a.ID, &a.PatientID, &a.ClinicalStatus, &a.VerificationStatus, &a.Type, &a.SubstanceCode, &a.SubstanceDisplay, &a.Criticality, &a.SiteID, &a.LastUpdated, &a.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (idx *sqliteIndex) GetImmunization(patientID, id string) (*fhir.ImmunizationRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, vaccine_code, vaccine_display, occurrence_datetime, site_id, last_updated, git_blob_hash FROM immunizations WHERE id = ? AND patient_id = ?`, id, patientID)
	i := &fhir.ImmunizationRow{}
	err := row.Scan(&i.ID, &i.PatientID, &i.Status, &i.VaccineCode, &i.VaccineDisplay, &i.OccurrenceDatetime, &i.SiteID, &i.LastUpdated, &i.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (idx *sqliteIndex) GetProcedure(patientID, id string) (*fhir.ProcedureRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, code, code_display, performed_datetime, site_id, last_updated, git_blob_hash FROM procedures WHERE id = ? AND patient_id = ?`, id, patientID)
	p := &fhir.ProcedureRow{}
	err := row.Scan(&p.ID, &p.PatientID, &p.Status, &p.Code, &p.CodeDisplay, &p.PerformedDatetime, &p.SiteID, &p.LastUpdated, &p.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (idx *sqliteIndex) GetPractitioner(id string) (*fhir.PractitionerRow, error) {
	row := idx.db.QueryRow(`SELECT id, family_name, given_names, active, site_id, last_updated, git_blob_hash FROM practitioners WHERE id = ?`, id)
	p := &fhir.PractitionerRow{}
	var active int
	err := row.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &active, &p.SiteID, &p.LastUpdated, &p.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Active = active == 1
	return p, nil
}

func (idx *sqliteIndex) GetOrganization(id string) (*fhir.OrganizationRow, error) {
	row := idx.db.QueryRow(`SELECT id, name, type, active, site_id, last_updated, git_blob_hash FROM organizations WHERE id = ?`, id)
	o := &fhir.OrganizationRow{}
	var active int
	err := row.Scan(&o.ID, &o.Name, &o.Type, &active, &o.SiteID, &o.LastUpdated, &o.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	o.Active = active == 1
	return o, nil
}

func (idx *sqliteIndex) GetLocation(id string) (*fhir.LocationRow, error) {
	row := idx.db.QueryRow(`SELECT id, name, type, status, site_id, last_updated, git_blob_hash FROM locations WHERE id = ?`, id)
	l := &fhir.LocationRow{}
	err := row.Scan(&l.ID, &l.Name, &l.Type, &l.Status, &l.SiteID, &l.LastUpdated, &l.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (idx *sqliteIndex) GetMeasureReport(id string) (*fhir.MeasureReportRow, error) {
	row := idx.db.QueryRow(`SELECT id, status, type, period_start, period_end, reporter, site_id, last_updated, git_blob_hash FROM measure_reports WHERE id = ?`, id)
	m := &fhir.MeasureReportRow{}
	err := row.Scan(&m.ID, &m.Status, &m.Type, &m.PeriodStart, &m.PeriodEnd, &m.Reporter, &m.SiteID, &m.LastUpdated, &m.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

// --- ID-only Get methods (for FHIR REST API) ---

func (idx *sqliteIndex) GetEncounterByID(id string) (*fhir.EncounterRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, class_code, type_code, period_start, period_end, site_id, reason_code, last_updated, git_blob_hash FROM encounters WHERE id = ?`, id)
	e := &fhir.EncounterRow{}
	err := row.Scan(&e.ID, &e.PatientID, &e.Status, &e.ClassCode, &e.TypeCode, &e.PeriodStart, &e.PeriodEnd, &e.SiteID, &e.ReasonCode, &e.LastUpdated, &e.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (idx *sqliteIndex) GetObservationByID(id string) (*fhir.ObservationRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, encounter_id, status, category, code, code_display, effective_datetime, value_quantity_value, value_quantity_unit, value_string, value_codeable_concept, site_id, last_updated, git_blob_hash FROM observations WHERE id = ?`, id)
	o := &fhir.ObservationRow{}
	err := row.Scan(&o.ID, &o.PatientID, &o.EncounterID, &o.Status, &o.Category, &o.Code, &o.CodeDisplay, &o.EffectiveDatetime, &o.ValueQuantityValue, &o.ValueQuantityUnit, &o.ValueString, &o.ValueCodeableConcept, &o.SiteID, &o.LastUpdated, &o.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (idx *sqliteIndex) GetConditionByID(id string) (*fhir.ConditionRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, clinical_status, verification_status, code, code_display, onset_datetime, site_id, last_updated, git_blob_hash FROM conditions WHERE id = ?`, id)
	c := &fhir.ConditionRow{}
	err := row.Scan(&c.ID, &c.PatientID, &c.ClinicalStatus, &c.VerificationStatus, &c.Code, &c.CodeDisplay, &c.OnsetDatetime, &c.SiteID, &c.LastUpdated, &c.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (idx *sqliteIndex) GetMedicationRequestByID(id string) (*fhir.MedicationRequestRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, intent, medication_code, medication_display, authored_on, site_id, last_updated, git_blob_hash FROM medication_requests WHERE id = ?`, id)
	m := &fhir.MedicationRequestRow{}
	err := row.Scan(&m.ID, &m.PatientID, &m.Status, &m.Intent, &m.MedicationCode, &m.MedicationDisplay, &m.AuthoredOn, &m.SiteID, &m.LastUpdated, &m.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (idx *sqliteIndex) GetAllergyIntoleranceByID(id string) (*fhir.AllergyIntoleranceRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, clinical_status, verification_status, type, substance_code, substance_display, criticality, site_id, last_updated, git_blob_hash FROM allergy_intolerances WHERE id = ?`, id)
	a := &fhir.AllergyIntoleranceRow{}
	err := row.Scan(&a.ID, &a.PatientID, &a.ClinicalStatus, &a.VerificationStatus, &a.Type, &a.SubstanceCode, &a.SubstanceDisplay, &a.Criticality, &a.SiteID, &a.LastUpdated, &a.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (idx *sqliteIndex) GetImmunizationByID(id string) (*fhir.ImmunizationRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, vaccine_code, vaccine_display, occurrence_datetime, site_id, last_updated, git_blob_hash FROM immunizations WHERE id = ?`, id)
	i := &fhir.ImmunizationRow{}
	err := row.Scan(&i.ID, &i.PatientID, &i.Status, &i.VaccineCode, &i.VaccineDisplay, &i.OccurrenceDatetime, &i.SiteID, &i.LastUpdated, &i.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (idx *sqliteIndex) GetProcedureByID(id string) (*fhir.ProcedureRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, code, code_display, performed_datetime, site_id, last_updated, git_blob_hash FROM procedures WHERE id = ?`, id)
	p := &fhir.ProcedureRow{}
	err := row.Scan(&p.ID, &p.PatientID, &p.Status, &p.Code, &p.CodeDisplay, &p.PerformedDatetime, &p.SiteID, &p.LastUpdated, &p.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (idx *sqliteIndex) GetFlagByID(id string) (*fhir.FlagRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, category, code, period_start, period_end, generated_by, site_id, last_updated, git_blob_hash FROM flags WHERE id = ?`, id)
	f := &fhir.FlagRow{}
	err := row.Scan(&f.ID, &f.PatientID, &f.Status, &f.Category, &f.Code, &f.PeriodStart, &f.PeriodEnd, &f.GeneratedBy, &f.SiteID, &f.LastUpdated, &f.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// --- List methods ---

func (idx *sqliteIndex) ListPatients(opts PatientListOpts) ([]*fhir.PatientRow, *fhir.Pagination, error) {
	query := "SELECT id, family_name, given_names, gender, birth_date, site_id, active, last_updated, git_blob_hash FROM patients WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM patients WHERE 1=1"
	var args []any

	if opts.ActiveOnly {
		query += " AND active = 1"
		countQuery += " AND active = 1"
	}
	if opts.Gender != "" {
		query += " AND gender = ?"
		countQuery += " AND gender = ?"
		args = append(args, opts.Gender)
	}
	if opts.BirthDateFrom != "" {
		query += " AND birth_date >= ?"
		countQuery += " AND birth_date >= ?"
		args = append(args, opts.BirthDateFrom)
	}
	if opts.BirthDateTo != "" {
		query += " AND birth_date <= ?"
		countQuery += " AND birth_date <= ?"
		args = append(args, opts.BirthDateTo)
	}
	if opts.SiteID != "" {
		query += " AND site_id = ?"
		countQuery += " AND site_id = ?"
		args = append(args, opts.SiteID)
	}

	// Count
	var total int
	err := idx.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, nil, err
	}

	pg := paginate(opts.PaginationOpts, total)

	query += " ORDER BY last_updated DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pg.PerPage, (pg.Page-1)*pg.PerPage)

	rows, err := idx.db.Query(query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.PatientRow
	for rows.Next() {
		p := &fhir.PatientRow{}
		var active int
		if err := rows.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &p.Gender, &p.BirthDate, &p.SiteID, &active, &p.LastUpdated, &p.GitBlobHash); err != nil {
			return nil, nil, err
		}
		p.Active = active == 1
		results = append(results, p)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListEncounters(patientID string, opts fhir.PaginationOpts) ([]*fhir.EncounterRow, *fhir.Pagination, error) {
	var total int
	err := idx.db.QueryRow("SELECT COUNT(*) FROM encounters WHERE patient_id = ?", patientID).Scan(&total)
	if err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, patient_id, status, class_code, type_code, period_start, period_end, site_id, reason_code, last_updated, git_blob_hash FROM encounters WHERE patient_id = ? ORDER BY period_start DESC LIMIT ? OFFSET ?",
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.EncounterRow
	for rows.Next() {
		e := &fhir.EncounterRow{}
		if err := rows.Scan(&e.ID, &e.PatientID, &e.Status, &e.ClassCode, &e.TypeCode, &e.PeriodStart, &e.PeriodEnd, &e.SiteID, &e.ReasonCode, &e.LastUpdated, &e.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, e)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListObservations(patientID string, opts ObservationListOpts) ([]*fhir.ObservationRow, *fhir.Pagination, error) {
	query := "SELECT id, patient_id, encounter_id, status, category, code, code_display, effective_datetime, value_quantity_value, value_quantity_unit, value_string, value_codeable_concept, site_id, last_updated, git_blob_hash FROM observations WHERE patient_id = ?"
	countQuery := "SELECT COUNT(*) FROM observations WHERE patient_id = ?"
	args := []any{patientID}

	if opts.Code != "" {
		query += " AND code = ?"
		countQuery += " AND code = ?"
		args = append(args, opts.Code)
	}
	if opts.Category != "" {
		query += " AND category = ?"
		countQuery += " AND category = ?"
		args = append(args, opts.Category)
	}
	if opts.DateFrom != "" {
		query += " AND effective_datetime >= ?"
		countQuery += " AND effective_datetime >= ?"
		args = append(args, opts.DateFrom)
	}
	if opts.DateTo != "" {
		query += " AND effective_datetime <= ?"
		countQuery += " AND effective_datetime <= ?"
		args = append(args, opts.DateTo)
	}
	if opts.EncounterID != "" {
		query += " AND encounter_id = ?"
		countQuery += " AND encounter_id = ?"
		args = append(args, opts.EncounterID)
	}

	var total int
	if err := idx.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts.PaginationOpts, total)

	query += " ORDER BY effective_datetime DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pg.PerPage, (pg.Page-1)*pg.PerPage)

	rows, err := idx.db.Query(query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.ObservationRow
	for rows.Next() {
		o := &fhir.ObservationRow{}
		if err := rows.Scan(&o.ID, &o.PatientID, &o.EncounterID, &o.Status, &o.Category, &o.Code, &o.CodeDisplay, &o.EffectiveDatetime, &o.ValueQuantityValue, &o.ValueQuantityUnit, &o.ValueString, &o.ValueCodeableConcept, &o.SiteID, &o.LastUpdated, &o.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, o)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListConditions(patientID string, opts ConditionListOpts) ([]*fhir.ConditionRow, *fhir.Pagination, error) {
	query := "SELECT id, patient_id, clinical_status, verification_status, code, code_display, onset_datetime, site_id, last_updated, git_blob_hash FROM conditions WHERE patient_id = ?"
	countQuery := "SELECT COUNT(*) FROM conditions WHERE patient_id = ?"
	args := []any{patientID}

	if opts.ClinicalStatus != "" {
		query += " AND clinical_status = ?"
		countQuery += " AND clinical_status = ?"
		args = append(args, opts.ClinicalStatus)
	}
	if opts.Code != "" {
		query += " AND code = ?"
		countQuery += " AND code = ?"
		args = append(args, opts.Code)
	}

	var total int
	if err := idx.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts.PaginationOpts, total)

	query += " ORDER BY last_updated DESC LIMIT ? OFFSET ?"
	queryArgs := append(args, pg.PerPage, (pg.Page-1)*pg.PerPage)

	rows, err := idx.db.Query(query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.ConditionRow
	for rows.Next() {
		c := &fhir.ConditionRow{}
		if err := rows.Scan(&c.ID, &c.PatientID, &c.ClinicalStatus, &c.VerificationStatus, &c.Code, &c.CodeDisplay, &c.OnsetDatetime, &c.SiteID, &c.LastUpdated, &c.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, c)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListMedicationRequests(patientID string, opts fhir.PaginationOpts) ([]*fhir.MedicationRequestRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM medication_requests WHERE patient_id = ?", patientID).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, patient_id, status, intent, medication_code, medication_display, authored_on, site_id, last_updated, git_blob_hash FROM medication_requests WHERE patient_id = ? ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.MedicationRequestRow
	for rows.Next() {
		m := &fhir.MedicationRequestRow{}
		if err := rows.Scan(&m.ID, &m.PatientID, &m.Status, &m.Intent, &m.MedicationCode, &m.MedicationDisplay, &m.AuthoredOn, &m.SiteID, &m.LastUpdated, &m.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, m)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListAllergyIntolerances(patientID string, opts fhir.PaginationOpts) ([]*fhir.AllergyIntoleranceRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM allergy_intolerances WHERE patient_id = ?", patientID).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, patient_id, clinical_status, verification_status, type, substance_code, substance_display, criticality, site_id, last_updated, git_blob_hash FROM allergy_intolerances WHERE patient_id = ? ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.AllergyIntoleranceRow
	for rows.Next() {
		a := &fhir.AllergyIntoleranceRow{}
		if err := rows.Scan(&a.ID, &a.PatientID, &a.ClinicalStatus, &a.VerificationStatus, &a.Type, &a.SubstanceCode, &a.SubstanceDisplay, &a.Criticality, &a.SiteID, &a.LastUpdated, &a.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, a)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListFlags(patientID string, opts fhir.PaginationOpts) ([]*fhir.FlagRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM flags WHERE patient_id = ?", patientID).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, patient_id, status, category, code, period_start, period_end, generated_by, site_id, last_updated, git_blob_hash FROM flags WHERE patient_id = ? ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.FlagRow
	for rows.Next() {
		f := &fhir.FlagRow{}
		if err := rows.Scan(&f.ID, &f.PatientID, &f.Status, &f.Category, &f.Code, &f.PeriodStart, &f.PeriodEnd, &f.GeneratedBy, &f.SiteID, &f.LastUpdated, &f.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, f)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListImmunizations(patientID string, opts fhir.PaginationOpts) ([]*fhir.ImmunizationRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM immunizations WHERE patient_id = ?", patientID).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, patient_id, status, vaccine_code, vaccine_display, occurrence_datetime, site_id, last_updated, git_blob_hash FROM immunizations WHERE patient_id = ? ORDER BY occurrence_datetime DESC LIMIT ? OFFSET ?",
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.ImmunizationRow
	for rows.Next() {
		i := &fhir.ImmunizationRow{}
		if err := rows.Scan(&i.ID, &i.PatientID, &i.Status, &i.VaccineCode, &i.VaccineDisplay, &i.OccurrenceDatetime, &i.SiteID, &i.LastUpdated, &i.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, i)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListProcedures(patientID string, opts fhir.PaginationOpts) ([]*fhir.ProcedureRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM procedures WHERE patient_id = ?", patientID).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, patient_id, status, code, code_display, performed_datetime, site_id, last_updated, git_blob_hash FROM procedures WHERE patient_id = ? ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.ProcedureRow
	for rows.Next() {
		p := &fhir.ProcedureRow{}
		if err := rows.Scan(&p.ID, &p.PatientID, &p.Status, &p.Code, &p.CodeDisplay, &p.PerformedDatetime, &p.SiteID, &p.LastUpdated, &p.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, p)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListPractitioners(opts fhir.PaginationOpts) ([]*fhir.PractitionerRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM practitioners").Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, family_name, given_names, active, site_id, last_updated, git_blob_hash FROM practitioners ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.PractitionerRow
	for rows.Next() {
		p := &fhir.PractitionerRow{}
		var active int
		if err := rows.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &active, &p.SiteID, &p.LastUpdated, &p.GitBlobHash); err != nil {
			return nil, nil, err
		}
		p.Active = active == 1
		results = append(results, p)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListOrganizations(opts fhir.PaginationOpts) ([]*fhir.OrganizationRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM organizations").Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, name, type, active, site_id, last_updated, git_blob_hash FROM organizations ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.OrganizationRow
	for rows.Next() {
		o := &fhir.OrganizationRow{}
		var active int
		if err := rows.Scan(&o.ID, &o.Name, &o.Type, &active, &o.SiteID, &o.LastUpdated, &o.GitBlobHash); err != nil {
			return nil, nil, err
		}
		o.Active = active == 1
		results = append(results, o)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListLocations(opts fhir.PaginationOpts) ([]*fhir.LocationRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM locations").Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, name, type, status, site_id, last_updated, git_blob_hash FROM locations ORDER BY last_updated DESC LIMIT ? OFFSET ?",
		pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.LocationRow
	for rows.Next() {
		l := &fhir.LocationRow{}
		if err := rows.Scan(&l.ID, &l.Name, &l.Type, &l.Status, &l.SiteID, &l.LastUpdated, &l.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, l)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) ListMeasureReports(opts fhir.PaginationOpts) ([]*fhir.MeasureReportRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM measure_reports").Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query("SELECT id, status, type, period_start, period_end, reporter, site_id, last_updated, git_blob_hash FROM measure_reports ORDER BY period_start DESC LIMIT ? OFFSET ?",
		pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.MeasureReportRow
	for rows.Next() {
		m := &fhir.MeasureReportRow{}
		if err := rows.Scan(&m.ID, &m.Status, &m.Type, &m.PeriodStart, &m.PeriodEnd, &m.Reporter, &m.SiteID, &m.LastUpdated, &m.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, m)
	}
	return results, pg, rows.Err()
}

// --- Blind index methods ---

func (idx *sqliteIndex) UpsertPatientNgrams(patientID string, field string, ngramHashes []string) error {
	tx, err := idx.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Delete existing n-grams for this patient+field
	if _, err := tx.Exec("DELETE FROM patients_ngrams WHERE patient_id = ? AND field = ?", patientID, field); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete old ngrams: %w", err)
	}

	// Insert new n-grams
	stmt, err := tx.Prepare("INSERT OR IGNORE INTO patients_ngrams (patient_id, ngram_hash, field) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, hash := range ngramHashes {
		if _, err := stmt.Exec(patientID, hash, field); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert ngram: %w", err)
		}
	}

	return tx.Commit()
}

func (idx *sqliteIndex) SearchPatientsByNgrams(ngramHashes []string, opts fhir.PaginationOpts) ([]*fhir.PatientRow, *fhir.Pagination, error) {
	if len(ngramHashes) == 0 {
		return nil, &fhir.Pagination{Page: 1, PerPage: 25, Total: 0, TotalPages: 1}, nil
	}

	// Build IN clause
	placeholders := ""
	args := make([]any, len(ngramHashes))
	for i, h := range ngramHashes {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = h
	}

	// Find patients matching ALL n-grams (intersection via COUNT)
	countQuery := fmt.Sprintf(`SELECT COUNT(DISTINCT p.id) FROM patients p
		INNER JOIN patients_ngrams ng ON p.id = ng.patient_id
		WHERE ng.ngram_hash IN (%s) AND p.active = 1
		GROUP BY p.id
		HAVING COUNT(DISTINCT ng.ngram_hash) = ?`, placeholders)

	countArgs := append(args, len(ngramHashes))

	// Count matching patients
	var total int
	rows, err := idx.db.Query(countQuery, countArgs...)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		total++
	}
	rows.Close()

	pg := paginate(opts, total)

	// Fetch matching patient IDs with pagination
	query := fmt.Sprintf(`SELECT p.id, p.family_name, p.given_names, p.gender, p.birth_date, p.site_id, p.active, p.last_updated, p.git_blob_hash
		FROM patients p
		INNER JOIN patients_ngrams ng ON p.id = ng.patient_id
		WHERE ng.ngram_hash IN (%s) AND p.active = 1
		GROUP BY p.id
		HAVING COUNT(DISTINCT ng.ngram_hash) = ?
		ORDER BY p.last_updated DESC
		LIMIT ? OFFSET ?`, placeholders)

	queryArgs := append(args, len(ngramHashes), pg.PerPage, (pg.Page-1)*pg.PerPage)

	pRows, err := idx.db.Query(query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer pRows.Close()

	var results []*fhir.PatientRow
	for pRows.Next() {
		p := &fhir.PatientRow{}
		var active int
		if err := pRows.Scan(&p.ID, &p.FamilyName, &p.GivenNames, &p.Gender, &p.BirthDate, &p.SiteID, &active, &p.LastUpdated, &p.GitBlobHash); err != nil {
			return nil, nil, err
		}
		p.Active = active == 1
		results = append(results, p)
	}
	return results, pg, pRows.Err()
}

// --- Consent methods ---

func (idx *sqliteIndex) UpsertConsent(row *fhir.ConsentRow) error {
	_, err := idx.db.Exec(`INSERT OR REPLACE INTO consents (id, patient_id, status, scope_code, performer_id, provision_type, period_start, period_end, category, last_updated, git_blob_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.PatientID, row.Status, row.ScopeCode, row.PerformerID, row.ProvisionType, row.PeriodStart, row.PeriodEnd, row.Category, row.LastUpdated, row.GitBlobHash)
	return err
}

func (idx *sqliteIndex) GetConsent(id string) (*fhir.ConsentRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, scope_code, performer_id, provision_type, period_start, period_end, category, last_updated, git_blob_hash FROM consents WHERE id = ?`, id)
	c := &fhir.ConsentRow{}
	err := row.Scan(&c.ID, &c.PatientID, &c.Status, &c.ScopeCode, &c.PerformerID, &c.ProvisionType, &c.PeriodStart, &c.PeriodEnd, &c.Category, &c.LastUpdated, &c.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (idx *sqliteIndex) ListConsentsForPatient(patientID string, opts fhir.PaginationOpts) ([]*fhir.ConsentRow, *fhir.Pagination, error) {
	var total int
	if err := idx.db.QueryRow("SELECT COUNT(*) FROM consents WHERE patient_id = ?", patientID).Scan(&total); err != nil {
		return nil, nil, err
	}
	pg := paginate(opts, total)

	rows, err := idx.db.Query(`SELECT id, patient_id, status, scope_code, performer_id, provision_type, period_start, period_end, category, last_updated, git_blob_hash FROM consents WHERE patient_id = ? ORDER BY last_updated DESC LIMIT ? OFFSET ?`,
		patientID, pg.PerPage, (pg.Page-1)*pg.PerPage)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var results []*fhir.ConsentRow
	for rows.Next() {
		c := &fhir.ConsentRow{}
		if err := rows.Scan(&c.ID, &c.PatientID, &c.Status, &c.ScopeCode, &c.PerformerID, &c.ProvisionType, &c.PeriodStart, &c.PeriodEnd, &c.Category, &c.LastUpdated, &c.GitBlobHash); err != nil {
			return nil, nil, err
		}
		results = append(results, c)
	}
	return results, pg, rows.Err()
}

func (idx *sqliteIndex) GetActiveConsent(patientID, performerID, scopeCode string) (*fhir.ConsentRow, error) {
	row := idx.db.QueryRow(`SELECT id, patient_id, status, scope_code, performer_id, provision_type, period_start, period_end, category, last_updated, git_blob_hash
		FROM consents
		WHERE patient_id = ? AND performer_id = ? AND scope_code = ? AND status = 'active' AND provision_type = 'permit'
		AND (period_start IS NULL OR period_start <= datetime('now'))
		AND (period_end IS NULL OR period_end >= datetime('now'))
		ORDER BY last_updated DESC LIMIT 1`, patientID, performerID, scopeCode)
	c := &fhir.ConsentRow{}
	err := row.Scan(&c.ID, &c.PatientID, &c.Status, &c.ScopeCode, &c.PerformerID, &c.ProvisionType, &c.PeriodStart, &c.PeriodEnd, &c.Category, &c.LastUpdated, &c.GitBlobHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (idx *sqliteIndex) DeleteConsent(id string) error {
	_, err := idx.db.Exec("DELETE FROM consents WHERE id = ?", id)
	return err
}

// --- Meta ---

func (idx *sqliteIndex) GetMeta(key string) (string, error) {
	var val string
	err := idx.db.QueryRow("SELECT value FROM index_meta WHERE key = ?", key).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return val, err
}

func (idx *sqliteIndex) SetMeta(key, value string) error {
	_, err := idx.db.Exec("INSERT OR REPLACE INTO index_meta (key, value) VALUES (?, ?)", key, value)
	return err
}

func (idx *sqliteIndex) ResourceCount() (int, error) {
	var total int
	for _, table := range []string{"patients", "encounters", "observations", "conditions", "medication_requests", "allergy_intolerances", "flags", "immunizations", "procedures", "practitioners", "organizations", "locations", "measure_reports"} {
		var count int
		if err := idx.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

// paginate normalises pagination options and computes pagination metadata.
func paginate(opts fhir.PaginationOpts, total int) *fhir.Pagination {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	perPage := opts.PerPage
	if perPage < 1 {
		perPage = 25
	}
	if perPage > 100 {
		perPage = 100
	}
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	return &fhir.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
