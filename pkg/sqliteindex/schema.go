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

// InitUnifiedSchema creates all tables for the monolith: patient index +
// auth deny list + sync state + formulary stock + anchor queue.
// Table names across services don't collide — verified during Phase 1 design.
func InitUnifiedSchema(db *sql.DB) error {
	if err := InitSchema(db); err != nil {
		return err
	}
	_, err := db.Exec(unifiedExtraDDL)
	return err
}

// unifiedExtraDDL contains tables from auth, sync, formulary, and anchor services.
const unifiedExtraDDL = `
-- Auth: deny list & revocations
CREATE TABLE IF NOT EXISTS deny_list (
    jti TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    added_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_deny_list_device ON deny_list(device_id);

CREATE TABLE IF NOT EXISTS revocations (
    device_id TEXT PRIMARY KEY,
    public_key TEXT NOT NULL,
    revoked_at TEXT NOT NULL DEFAULT (datetime('now')),
    revoked_by TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS node_info (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Auth: SMART clients
CREATE TABLE IF NOT EXISTS smart_clients (
    client_id TEXT PRIMARY KEY,
    client_name TEXT NOT NULL DEFAULT '',
    redirect_uris TEXT NOT NULL DEFAULT '[]',
    scope TEXT NOT NULL DEFAULT '',
    launch_modes TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Sync: conflicts, history, peer state
CREATE TABLE IF NOT EXISTS conflicts (
    id TEXT PRIMARY KEY,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    patient_id TEXT NOT NULL DEFAULT '',
    level TEXT NOT NULL DEFAULT 'review',
    status TEXT NOT NULL DEFAULT 'pending',
    detected_at TEXT NOT NULL DEFAULT (datetime('now')),
    local_version BLOB,
    remote_version BLOB,
    merged_version BLOB,
    changed_fields TEXT NOT NULL DEFAULT '[]',
    reason TEXT NOT NULL DEFAULT '',
    local_node TEXT NOT NULL DEFAULT '',
    remote_node TEXT NOT NULL DEFAULT '',
    peer_site_id TEXT NOT NULL DEFAULT '',
    resolved_at TEXT,
    resolved_by TEXT,
    resolution TEXT
);
CREATE INDEX IF NOT EXISTS idx_conflicts_status ON conflicts(status);
CREATE INDEX IF NOT EXISTS idx_conflicts_level ON conflicts(level);
CREATE INDEX IF NOT EXISTS idx_conflicts_patient ON conflicts(patient_id);

CREATE TABLE IF NOT EXISTS sync_history (
    id TEXT PRIMARY KEY,
    peer_node TEXT NOT NULL,
    transport TEXT NOT NULL DEFAULT '',
    direction TEXT NOT NULL DEFAULT 'bidirectional',
    state TEXT NOT NULL DEFAULT 'completed',
    started_at TEXT NOT NULL DEFAULT (datetime('now')),
    completed_at TEXT,
    resources_sent INTEGER NOT NULL DEFAULT 0,
    resources_received INTEGER NOT NULL DEFAULT 0,
    conflicts_detected INTEGER NOT NULL DEFAULT 0,
    local_head_before TEXT NOT NULL DEFAULT '',
    local_head_after TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sync_history_peer ON sync_history(peer_node);
CREATE INDEX IF NOT EXISTS idx_sync_history_state ON sync_history(state);

CREATE TABLE IF NOT EXISTS peer_state (
    node_id TEXT PRIMARY KEY,
    site_id TEXT NOT NULL DEFAULT '',
    public_key BLOB,
    trusted INTEGER NOT NULL DEFAULT 0,
    last_seen TEXT NOT NULL DEFAULT (datetime('now')),
    their_head TEXT NOT NULL DEFAULT '',
    transport TEXT NOT NULL DEFAULT '',
    revoked INTEGER NOT NULL DEFAULT 0
);

-- Formulary: stock levels & deliveries
CREATE TABLE IF NOT EXISTS stock_levels (
    site_id            TEXT NOT NULL,
    medication_code    TEXT NOT NULL,
    quantity           INTEGER NOT NULL DEFAULT 0,
    unit               TEXT NOT NULL DEFAULT 'units',
    last_updated       TEXT NOT NULL DEFAULT '',
    earliest_expiry    TEXT NOT NULL DEFAULT '',
    daily_consumption_rate REAL NOT NULL DEFAULT 0.0,
    PRIMARY KEY (site_id, medication_code)
);
CREATE INDEX IF NOT EXISTS idx_stock_site ON stock_levels(site_id);
CREATE INDEX IF NOT EXISTS idx_stock_medication ON stock_levels(medication_code);

CREATE TABLE IF NOT EXISTS deliveries (
    id              TEXT PRIMARY KEY,
    site_id         TEXT NOT NULL,
    received_by     TEXT NOT NULL,
    delivery_date   TEXT NOT NULL,
    items_recorded  INTEGER NOT NULL DEFAULT 0,
    created_at      TEXT NOT NULL DEFAULT ''
);

-- Anchor: queue
CREATE TABLE IF NOT EXISTS anchor_queue (
    anchor_id    TEXT PRIMARY KEY,
    merkle_root  TEXT NOT NULL,
    git_head     TEXT NOT NULL,
    node_did     TEXT NOT NULL,
    state        TEXT NOT NULL DEFAULT 'pending',
    enqueued_at  TEXT NOT NULL,
    processed_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_queue_state ON anchor_queue(state);
CREATE INDEX IF NOT EXISTS idx_queue_enqueued ON anchor_queue(enqueued_at);
`

const dropDDL = `
DROP TABLE IF EXISTS patient_summaries;
DROP TABLE IF EXISTS consents;
DROP TABLE IF EXISTS patients_ngrams;
DROP TABLE IF EXISTS flags;
DROP TABLE IF EXISTS allergy_intolerances;
DROP TABLE IF EXISTS medication_requests;
DROP TABLE IF EXISTS conditions;
DROP TABLE IF EXISTS observations;
DROP TABLE IF EXISTS encounters;
DROP TABLE IF EXISTS immunizations;
DROP TABLE IF EXISTS procedures;
DROP TRIGGER IF EXISTS patients_ai;
DROP TRIGGER IF EXISTS patients_ad;
DROP TRIGGER IF EXISTS patients_au;
DROP TABLE IF EXISTS patients_fts;
DROP TABLE IF EXISTS patients;
DROP TABLE IF EXISTS measure_reports;
DROP TABLE IF EXISTS detected_issues;
DROP TABLE IF EXISTS practitioners;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS locations;
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
    git_blob_hash TEXT NOT NULL
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
    git_blob_hash TEXT NOT NULL
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
    git_blob_hash TEXT NOT NULL
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
    git_blob_hash TEXT NOT NULL
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
    git_blob_hash TEXT NOT NULL
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
    git_blob_hash TEXT NOT NULL
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
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_flag_patient ON flags(patient_id);
CREATE INDEX IF NOT EXISTS idx_flag_status ON flags(status);
CREATE INDEX IF NOT EXISTS idx_flag_category ON flags(category);

CREATE TABLE IF NOT EXISTS immunizations (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    vaccine_code TEXT NOT NULL,
    vaccine_display TEXT,
    occurrence_datetime TEXT NOT NULL,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_imm_patient ON immunizations(patient_id);
CREATE INDEX IF NOT EXISTS idx_imm_status ON immunizations(status);
CREATE INDEX IF NOT EXISTS idx_imm_vaccine ON immunizations(vaccine_code);
CREATE INDEX IF NOT EXISTS idx_imm_date ON immunizations(occurrence_datetime);

CREATE TABLE IF NOT EXISTS procedures (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    code TEXT NOT NULL,
    code_display TEXT,
    performed_datetime TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_proc_patient ON procedures(patient_id);
CREATE INDEX IF NOT EXISTS idx_proc_status ON procedures(status);
CREATE INDEX IF NOT EXISTS idx_proc_code ON procedures(code);
CREATE INDEX IF NOT EXISTS idx_proc_date ON procedures(performed_datetime);

CREATE TABLE IF NOT EXISTS practitioners (
    id TEXT PRIMARY KEY,
    family_name TEXT NOT NULL,
    given_names TEXT NOT NULL,
    active INTEGER DEFAULT 1,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pract_name ON practitioners(family_name);
CREATE INDEX IF NOT EXISTS idx_pract_active ON practitioners(active);

CREATE TABLE IF NOT EXISTS organizations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT,
    active INTEGER DEFAULT 1,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_org_name ON organizations(name);
CREATE INDEX IF NOT EXISTS idx_org_active ON organizations(active);

CREATE TABLE IF NOT EXISTS locations (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT,
    status TEXT NOT NULL,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_loc_name ON locations(name);
CREATE INDEX IF NOT EXISTS idx_loc_status ON locations(status);

CREATE TABLE IF NOT EXISTS measure_reports (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    type TEXT NOT NULL,
    period_start TEXT NOT NULL,
    period_end TEXT,
    reporter TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_mr_status ON measure_reports(status);
CREATE INDEX IF NOT EXISTS idx_mr_type ON measure_reports(type);
CREATE INDEX IF NOT EXISTS idx_mr_period ON measure_reports(period_start);

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
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_di_severity ON detected_issues(severity);
CREATE INDEX IF NOT EXISTS idx_di_status ON detected_issues(status);
CREATE INDEX IF NOT EXISTS idx_di_date ON detected_issues(identified_datetime);

CREATE TABLE IF NOT EXISTS patients_ngrams (
    patient_id TEXT NOT NULL REFERENCES patients(id),
    ngram_hash TEXT NOT NULL,
    field TEXT NOT NULL,
    PRIMARY KEY (patient_id, ngram_hash, field)
);

CREATE INDEX IF NOT EXISTS idx_ngram_hash ON patients_ngrams(ngram_hash);

CREATE TABLE IF NOT EXISTS consents (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL,
    status TEXT NOT NULL,
    scope_code TEXT NOT NULL,
    performer_id TEXT NOT NULL,
    provision_type TEXT NOT NULL,
    period_start TEXT,
    period_end TEXT,
    category TEXT,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_consent_patient ON consents(patient_id);
CREATE INDEX IF NOT EXISTS idx_consent_status ON consents(status);
CREATE INDEX IF NOT EXISTS idx_consent_performer ON consents(performer_id);
CREATE INDEX IF NOT EXISTS idx_consent_scope ON consents(scope_code);

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
