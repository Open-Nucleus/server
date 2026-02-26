package sqliteindex

import "database/sql"

// InitSchema creates all tables, indexes, FTS5, and triggers per spec §5.1.
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(schemaDDL)
	return err
}

// DropAll drops all tables except index_meta.
func DropAll(db *sql.DB) error {
	_, err := db.Exec(dropDDL)
	return err
}

const dropDDL = `
DROP TABLE IF EXISTS patient_summaries;
DROP TABLE IF EXISTS flags;
DROP TABLE IF EXISTS allergy_intolerances;
DROP TABLE IF EXISTS medication_requests;
DROP TABLE IF EXISTS conditions;
DROP TABLE IF EXISTS observations;
DROP TABLE IF EXISTS encounters;
DROP TRIGGER IF EXISTS patients_ai;
DROP TRIGGER IF EXISTS patients_ad;
DROP TRIGGER IF EXISTS patients_au;
DROP TABLE IF EXISTS patients_fts;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS detected_issues;
`

const schemaDDL = `
CREATE TABLE IF NOT EXISTS patients (
    id TEXT PRIMARY KEY,
    family_name TEXT NOT NULL,
    given_names TEXT NOT NULL,
    gender TEXT NOT NULL,
    birth_date TEXT NOT NULL,
    site_id TEXT NOT NULL,
    active INTEGER DEFAULT 1,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_patients_name ON patients(family_name, given_names);
CREATE INDEX IF NOT EXISTS idx_patients_gender ON patients(gender);
CREATE INDEX IF NOT EXISTS idx_patients_birth ON patients(birth_date);
CREATE INDEX IF NOT EXISTS idx_patients_site ON patients(site_id);
CREATE INDEX IF NOT EXISTS idx_patients_updated ON patients(last_updated);

CREATE VIRTUAL TABLE IF NOT EXISTS patients_fts USING fts5(
    id, family_name, given_names,
    content='patients', content_rowid='rowid'
);

CREATE TRIGGER IF NOT EXISTS patients_ai AFTER INSERT ON patients BEGIN
    INSERT INTO patients_fts(rowid, id, family_name, given_names)
    VALUES (new.rowid, new.id, new.family_name, new.given_names);
END;

CREATE TRIGGER IF NOT EXISTS patients_ad AFTER DELETE ON patients BEGIN
    INSERT INTO patients_fts(patients_fts, rowid, id, family_name, given_names)
    VALUES ('delete', old.rowid, old.id, old.family_name, old.given_names);
END;

CREATE TRIGGER IF NOT EXISTS patients_au AFTER UPDATE ON patients BEGIN
    INSERT INTO patients_fts(patients_fts, rowid, id, family_name, given_names)
    VALUES ('delete', old.rowid, old.id, old.family_name, old.given_names);
    INSERT INTO patients_fts(rowid, id, family_name, given_names)
    VALUES (new.rowid, new.id, new.family_name, new.given_names);
END;

CREATE TABLE IF NOT EXISTS encounters (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    class_code TEXT NOT NULL,
    type_code TEXT,
    period_start TEXT NOT NULL,
    period_end TEXT,
    site_id TEXT NOT NULL,
    reason_code TEXT,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_enc_patient ON encounters(patient_id);
CREATE INDEX IF NOT EXISTS idx_enc_status ON encounters(status);
CREATE INDEX IF NOT EXISTS idx_enc_date ON encounters(period_start);
CREATE INDEX IF NOT EXISTS idx_enc_site ON encounters(site_id);

CREATE TABLE IF NOT EXISTS observations (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    encounter_id TEXT REFERENCES encounters(id),
    status TEXT NOT NULL,
    category TEXT,
    code TEXT NOT NULL,
    code_display TEXT,
    effective_datetime TEXT NOT NULL,
    value_quantity_value REAL,
    value_quantity_unit TEXT,
    value_string TEXT,
    value_codeable_concept TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_obs_patient ON observations(patient_id);
CREATE INDEX IF NOT EXISTS idx_obs_encounter ON observations(encounter_id);
CREATE INDEX IF NOT EXISTS idx_obs_code ON observations(code);
CREATE INDEX IF NOT EXISTS idx_obs_category ON observations(category);
CREATE INDEX IF NOT EXISTS idx_obs_date ON observations(effective_datetime);

CREATE TABLE IF NOT EXISTS conditions (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    clinical_status TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    code TEXT NOT NULL,
    code_display TEXT,
    onset_datetime TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cond_patient ON conditions(patient_id);
CREATE INDEX IF NOT EXISTS idx_cond_status ON conditions(clinical_status);
CREATE INDEX IF NOT EXISTS idx_cond_code ON conditions(code);

CREATE TABLE IF NOT EXISTS medication_requests (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    intent TEXT NOT NULL,
    medication_code TEXT NOT NULL,
    medication_display TEXT,
    authored_on TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_medrq_patient ON medication_requests(patient_id);
CREATE INDEX IF NOT EXISTS idx_medrq_status ON medication_requests(status);
CREATE INDEX IF NOT EXISTS idx_medrq_medication ON medication_requests(medication_code);

CREATE TABLE IF NOT EXISTS allergy_intolerances (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    clinical_status TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    type TEXT,
    substance_code TEXT NOT NULL,
    substance_display TEXT,
    criticality TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_allergy_patient ON allergy_intolerances(patient_id);
CREATE INDEX IF NOT EXISTS idx_allergy_substance ON allergy_intolerances(substance_code);
CREATE INDEX IF NOT EXISTS idx_allergy_criticality ON allergy_intolerances(criticality);

CREATE TABLE IF NOT EXISTS flags (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    category TEXT,
    code TEXT,
    period_start TEXT,
    period_end TEXT,
    generated_by TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_flag_patient ON flags(patient_id);
CREATE INDEX IF NOT EXISTS idx_flag_status ON flags(status);
CREATE INDEX IF NOT EXISTS idx_flag_category ON flags(category);

CREATE TABLE IF NOT EXISTS detected_issues (
    id TEXT PRIMARY KEY,
    severity TEXT NOT NULL,
    code TEXT,
    detail TEXT,
    identified_datetime TEXT NOT NULL,
    status TEXT NOT NULL,
    implicated_sites TEXT,
    implicated_patients TEXT,
    generated_by TEXT,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_di_severity ON detected_issues(severity);
CREATE INDEX IF NOT EXISTS idx_di_status ON detected_issues(status);
CREATE INDEX IF NOT EXISTS idx_di_date ON detected_issues(identified_datetime);

CREATE TABLE IF NOT EXISTS patient_summaries (
    patient_id TEXT PRIMARY KEY REFERENCES patients(id),
    encounter_count INTEGER DEFAULT 0,
    active_conditions INTEGER DEFAULT 0,
    active_medications INTEGER DEFAULT 0,
    active_allergies INTEGER DEFAULT 0,
    unresolved_alerts INTEGER DEFAULT 0,
    last_encounter_date TEXT,
    last_updated TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS index_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
`
