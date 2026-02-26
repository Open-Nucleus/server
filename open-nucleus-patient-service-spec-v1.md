# Open Nucleus — Patient Service Specification V1

**Version:** 1.0  
**Date:** February 2026  
**Author:** Dr Akanimoh Osutuk — FibrinLab  
**Repo:** github.com/FibrinLab/open-nucleus  
**Service:** `services/patient/`  
**Status:** Draft — V1 Specification

---

## 1. Service Overview

### 1.1 Role

The Patient Service is the primary write path for all clinical data in Open Nucleus. It owns CRUD operations on FHIR R4 resources, manages the Git commit lifecycle, maintains the SQLite query index, and enforces FHIR validation. No other service writes clinical data to Git — the Sync Service writes merge commits, and the Sentinel Agent writes alerts *through* the Patient Service API.

### 1.2 Service Identity

| Property | Value |
|----------|-------|
| Language | Go |
| gRPC Port | 50051 |
| Dependencies | `pkg/fhir`, `pkg/gitstore`, `pkg/sqliteindex`, `pkg/merge` |
| Writes to | Git repository (commits), SQLite (index upserts) |
| Reads from | SQLite (queries), Git (version history, diffs) |
| Consumed by | API Gateway (via gRPC), Sentinel Agent (alert write-back) |

### 1.3 Design Principles

- **Single writer for clinical data:** All FHIR resource mutations flow through this service. This eliminates Git write conflicts from concurrent local services.
- **Git-first writes:** The Git commit is the authoritative write. SQLite is updated post-commit. If SQLite fails, the data is safe in Git and the index is rebuilt.
- **Validate before write:** No invalid FHIR resource ever reaches Git. Validation is strict and happens before any I/O.
- **Stateless service logic:** All state lives in Git and SQLite. The service can be restarted at any time without data loss.

---

## 2. gRPC Service Definition

### 2.1 Proto File

```protobuf
syntax = "proto3";
package opennucleus.patient.v1;

import "google/protobuf/timestamp.proto";

service PatientService {
  // Patient CRUD
  rpc ListPatients(ListPatientsRequest) returns (ListPatientsResponse);
  rpc GetPatient(GetPatientRequest) returns (GetPatientResponse);
  rpc GetPatientBundle(GetPatientBundleRequest) returns (GetPatientBundleResponse);
  rpc CreatePatient(CreatePatientRequest) returns (MutationResponse);
  rpc UpdatePatient(UpdatePatientRequest) returns (MutationResponse);
  rpc DeletePatient(DeletePatientRequest) returns (MutationResponse);
  
  // Patient search and matching
  rpc SearchPatients(SearchPatientsRequest) returns (ListPatientsResponse);
  rpc MatchPatient(MatchPatientRequest) returns (MatchPatientResponse);
  
  // Patient history
  rpc GetPatientHistory(GetPatientHistoryRequest) returns (PatientHistoryResponse);
  rpc GetPatientTimeline(GetPatientTimelineRequest) returns (PatientTimelineResponse);
  
  // Encounter CRUD
  rpc ListEncounters(ListEncountersRequest) returns (ListEncountersResponse);
  rpc GetEncounter(GetEncounterRequest) returns (FhirResourceResponse);
  rpc CreateEncounter(CreateResourceRequest) returns (MutationResponse);
  rpc UpdateEncounter(UpdateResourceRequest) returns (MutationResponse);
  
  // Observation CRUD
  rpc ListObservations(ListObservationsRequest) returns (ListObservationsResponse);
  rpc GetObservation(GetObservationRequest) returns (FhirResourceResponse);
  rpc CreateObservation(CreateResourceRequest) returns (MutationResponse);
  
  // Condition CRUD
  rpc ListConditions(ListConditionsRequest) returns (ListConditionsResponse);
  rpc GetCondition(GetConditionRequest) returns (FhirResourceResponse);
  rpc CreateCondition(CreateResourceRequest) returns (MutationResponse);
  rpc UpdateCondition(UpdateResourceRequest) returns (MutationResponse);
  
  // MedicationRequest CRUD
  rpc ListMedicationRequests(ListMedicationRequestsRequest) returns (ListMedicationRequestsResponse);
  rpc GetMedicationRequest(GetMedicationRequestRequest) returns (FhirResourceResponse);
  rpc CreateMedicationRequest(CreateResourceRequest) returns (MutationResponse);
  rpc UpdateMedicationRequest(UpdateResourceRequest) returns (MutationResponse);
  
  // AllergyIntolerance CRUD
  rpc ListAllergyIntolerances(ListAllergyIntolerancesRequest) returns (ListAllergyIntolerancesResponse);
  rpc GetAllergyIntolerance(GetAllergyIntoleranceRequest) returns (FhirResourceResponse);
  rpc CreateAllergyIntolerance(CreateResourceRequest) returns (MutationResponse);
  rpc UpdateAllergyIntolerance(UpdateResourceRequest) returns (MutationResponse);
  
  // Flag CRUD (used by Sentinel Agent write-back)
  rpc CreateFlag(CreateResourceRequest) returns (MutationResponse);
  rpc UpdateFlag(UpdateResourceRequest) returns (MutationResponse);
  
  // Batch operations
  rpc CreateBatch(CreateBatchRequest) returns (BatchResponse);
  
  // Index management
  rpc RebuildIndex(RebuildIndexRequest) returns (RebuildIndexResponse);
  rpc CheckIndexHealth(CheckIndexHealthRequest) returns (IndexHealthResponse);
  
  // Health check
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

### 2.2 Core Message Types

```protobuf
// Context passed on every mutating request
message MutationContext {
  string practitioner_id = 1;    // Who is making this change
  string node_id = 2;            // Which device
  string site_id = 3;            // Which facility
  google.protobuf.Timestamp timestamp = 4;
}

// Generic FHIR resource wrapper (JSON bytes)
message FhirResource {
  string resource_type = 1;      // "Patient", "Encounter", etc.
  string id = 2;                 // Resource UUID
  bytes fhir_json = 3;           // Complete FHIR R4 JSON
}

// Standard mutation response
message MutationResponse {
  FhirResource resource = 1;
  GitCommitInfo git = 2;
}

message GitCommitInfo {
  string commit_hash = 1;
  string message = 2;
  google.protobuf.Timestamp timestamp = 3;
}

// Generic create/update for child resources
message CreateResourceRequest {
  string patient_id = 1;
  bytes fhir_json = 2;           // FHIR resource as JSON bytes
  MutationContext context = 3;
}

message UpdateResourceRequest {
  string patient_id = 1;
  string resource_id = 2;
  bytes fhir_json = 3;
  MutationContext context = 4;
}

message FhirResourceResponse {
  bytes fhir_json = 1;
}
```

---

## 3. Write Pipeline

### 3.1 Write Flow (Every Mutation)

```
gRPC Request
    │
    ▼
┌──────────────────────────────────┐
│  1. Deserialise FHIR JSON        │
│  2. Validate against R4 schema   │
│  3. Assign UUID if CREATE        │
│  4. Set meta.lastUpdated         │
│  5. Set meta.source (site_id)    │
│  6. Set meta.versionId (short hash) │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│  7. Acquire write lock           │
│  8. Write JSON to Git worktree   │
│  9. git add <path>               │
│ 10. git commit (structured msg)  │
│ 11. Upsert SQLite index          │
│ 12. Release write lock           │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│ 13. Return resource + commit     │
└──────────────────────────────────┘
```

### 3.2 Write Lock

The Patient Service uses a single-writer mutex to serialise all Git writes. This ensures:

- No concurrent commits from different goroutines
- Git working tree is never in a dirty state between operations
- SQLite upsert always follows the corresponding Git commit

The lock is held for the duration of steps 8–12. Read operations (queries, history) are not blocked by the write lock — SQLite supports concurrent reads, and Git reads go against committed objects.

**Lock scope:** Process-level `sync.Mutex`. In V1, only one Patient Service instance runs per node, so process-level locking is sufficient. Distributed locking is not needed.

### 3.3 Git Commit Details

**File path derivation:**

| Resource Type | Git Path |
|---------------|----------|
| Patient | `patients/{patient-id}/Patient.json` |
| Encounter | `patients/{patient-id}/encounters/{id}.json` |
| Observation | `patients/{patient-id}/observations/{id}.json` |
| Condition | `patients/{patient-id}/conditions/{id}.json` |
| MedicationRequest | `patients/{patient-id}/medication-requests/{id}.json` |
| AllergyIntolerance | `patients/{patient-id}/allergy-intolerances/{id}.json` |
| Flag | `patients/{patient-id}/flags/{id}.json` |
| DetectedIssue | `alerts/{id}.json` |
| SupplyDelivery | `supply/deliveries/{id}.json` |

**Commit message format:**

```
[{ResourceType}] {OPERATION} {resource-id}

node: {node-id}
author: {practitioner-id}
site: {site-id}
timestamp: {ISO-8601}
fhir_version: R4
```

**Operations:** `CREATE`, `UPDATE`, `DELETE` (soft-delete marks resource inactive, does not remove file)

**Example:**

```
[Encounter] CREATE enc-a1b2c3d4

node: node-sheffield-01
author: dr-osutuk
site: clinic-maiduguri-03
timestamp: 2026-03-15T09:42:00Z
fhir_version: R4
```

### 3.4 Soft Delete

DELETE operations do not remove files from Git. Instead, the resource is updated with:

- Patient: `active` set to `false`
- Encounter: `status` set to `entered-in-error`
- Condition: `clinicalStatus` set to `inactive`, `verificationStatus` set to `entered-in-error`
- MedicationRequest: `status` set to `entered-in-error`
- AllergyIntolerance: `verificationStatus` set to `entered-in-error`

This preserves the full Git history and ensures deleted records are visible in audit trails.

---

## 4. FHIR Validation

### 4.1 Validation Layers

Every incoming FHIR resource passes through three validation layers before any write:

**Layer 1 — Structural Validation:**
- Valid JSON
- `resourceType` field present and matches expected type
- All required FHIR R4 fields present
- Data types correct (strings, dates, codings, quantities)
- No unknown fields (strict mode)

**Layer 2 — Referential Validation:**
- `subject` reference resolves to an existing Patient (for child resources)
- `encounter` reference resolves to an existing Encounter (for observations, etc.)
- `participant.individual` references valid Practitioner IDs
- Cross-references within the same patient scope only

**Layer 3 — Clinical Validation:**
- ICD-10 codes are valid (checked against embedded code list)
- LOINC codes are valid for observations
- SNOMED CT codes are valid where used
- Quantity units match expected units for the observation code (e.g. temperature must be in Cel or [degF])
- Date ranges are logical (encounter end not before start, birth date not in the future)

### 4.2 Validation Response

On failure, the service returns a structured error with field-level detail:

```json
{
  "code": "VALIDATION_ERROR",
  "message": "FHIR resource validation failed",
  "field_errors": [
    {
      "path": "birthDate",
      "rule": "required",
      "message": "Patient.birthDate is required"
    },
    {
      "path": "name[0].family",
      "rule": "min_length",
      "message": "Family name must be a non-empty string"
    },
    {
      "path": "gender",
      "rule": "value_set",
      "message": "Must be one of: male, female, other, unknown"
    }
  ]
}
```

### 4.3 Required Fields by Resource Type

**Patient:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `name[].family` | string | Yes | At least one name with family |
| `name[].given[]` | string[] | Yes | At least one given name |
| `gender` | code | Yes | male, female, other, unknown |
| `birthDate` | date | Yes | Full or partial (year only accepted) |
| `identifier[]` | Identifier[] | Auto | `open-nucleus:local` identifier auto-assigned |

**Encounter:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `status` | code | Yes | planned, arrived, triaged, in-progress, onleave, finished, cancelled, entered-in-error |
| `class` | Coding | Yes | AMB, EMER, IMP, etc. |
| `subject` | Reference | Yes | Must reference existing Patient |
| `period.start` | dateTime | Yes | |
| `participant[].individual` | Reference | Yes | At least one participant |

**Observation:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `status` | code | Yes | registered, preliminary, final, amended, corrected, cancelled, entered-in-error |
| `code` | CodeableConcept | Yes | LOINC code required |
| `subject` | Reference | Yes | Must reference existing Patient |
| `effectiveDateTime` | dateTime | Yes | When the observation was made |
| `value[x]` | varies | Yes* | valueQuantity, valueString, valueCodeableConcept, etc. (*except for panels) |

**Condition:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `clinicalStatus` | CodeableConcept | Yes | active, recurrence, relapse, inactive, remission, resolved |
| `verificationStatus` | CodeableConcept | Yes | unconfirmed, provisional, differential, confirmed, refuted, entered-in-error |
| `code` | CodeableConcept | Yes | ICD-10 code required |
| `subject` | Reference | Yes | Must reference existing Patient |

**MedicationRequest:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `status` | code | Yes | active, on-hold, cancelled, completed, entered-in-error, stopped, draft |
| `intent` | code | Yes | proposal, plan, order, original-order, reflex-order, filler-order, instance-order, option |
| `medicationCodeableConcept` | CodeableConcept | Yes | ATC code preferred |
| `subject` | Reference | Yes | Must reference existing Patient |
| `dosageInstruction[]` | Dosage[] | Yes | At least one dosage instruction |

**AllergyIntolerance:**

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `clinicalStatus` | CodeableConcept | Yes | active, inactive, resolved |
| `verificationStatus` | CodeableConcept | Yes | unconfirmed, confirmed, refuted, entered-in-error |
| `type` | code | No | allergy, intolerance |
| `code` | CodeableConcept | Yes | Substance/product code |
| `patient` | Reference | Yes | Must reference existing Patient |
| `criticality` | code | No | low, high, unable-to-assess |

---

## 5. SQLite Index Schema

### 5.1 Tables

```sql
-- Core patient index
CREATE TABLE patients (
    id TEXT PRIMARY KEY,
    family_name TEXT NOT NULL,
    given_names TEXT NOT NULL,      -- JSON array as text
    gender TEXT NOT NULL,
    birth_date TEXT NOT NULL,       -- ISO date or partial (year)
    site_id TEXT NOT NULL,
    active INTEGER DEFAULT 1,       -- 0 = soft-deleted
    last_updated TEXT NOT NULL,      -- ISO datetime
    git_blob_hash TEXT NOT NULL,     -- SHA for traceability
    fhir_json TEXT NOT NULL          -- Complete FHIR resource
);

CREATE INDEX idx_patients_name ON patients(family_name, given_names);
CREATE INDEX idx_patients_gender ON patients(gender);
CREATE INDEX idx_patients_birth ON patients(birth_date);
CREATE INDEX idx_patients_site ON patients(site_id);
CREATE INDEX idx_patients_updated ON patients(last_updated);

-- Full-text search
CREATE VIRTUAL TABLE patients_fts USING fts5(
    id, family_name, given_names, 
    content='patients', content_rowid='rowid'
);

-- Encounters
CREATE TABLE encounters (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    class_code TEXT NOT NULL,
    type_code TEXT,                  -- SNOMED code
    period_start TEXT NOT NULL,
    period_end TEXT,
    site_id TEXT NOT NULL,
    reason_code TEXT,                -- ICD-10
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX idx_enc_patient ON encounters(patient_id);
CREATE INDEX idx_enc_status ON encounters(status);
CREATE INDEX idx_enc_date ON encounters(period_start);
CREATE INDEX idx_enc_site ON encounters(site_id);

-- Observations
CREATE TABLE observations (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    encounter_id TEXT REFERENCES encounters(id),
    status TEXT NOT NULL,
    category TEXT,                   -- vital-signs, laboratory, etc.
    code TEXT NOT NULL,              -- LOINC code
    code_display TEXT,
    effective_datetime TEXT NOT NULL,
    value_quantity_value REAL,
    value_quantity_unit TEXT,
    value_string TEXT,
    value_codeable_concept TEXT,     -- JSON
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX idx_obs_patient ON observations(patient_id);
CREATE INDEX idx_obs_encounter ON observations(encounter_id);
CREATE INDEX idx_obs_code ON observations(code);
CREATE INDEX idx_obs_category ON observations(category);
CREATE INDEX idx_obs_date ON observations(effective_datetime);

-- Conditions
CREATE TABLE conditions (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    clinical_status TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    code TEXT NOT NULL,              -- ICD-10
    code_display TEXT,
    onset_datetime TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX idx_cond_patient ON conditions(patient_id);
CREATE INDEX idx_cond_status ON conditions(clinical_status);
CREATE INDEX idx_cond_code ON conditions(code);

-- Medication Requests
CREATE TABLE medication_requests (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,
    intent TEXT NOT NULL,
    medication_code TEXT NOT NULL,    -- ATC code
    medication_display TEXT,
    authored_on TEXT,
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX idx_medrq_patient ON medication_requests(patient_id);
CREATE INDEX idx_medrq_status ON medication_requests(status);
CREATE INDEX idx_medrq_medication ON medication_requests(medication_code);

-- Allergy Intolerances
CREATE TABLE allergy_intolerances (
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

CREATE INDEX idx_allergy_patient ON allergy_intolerances(patient_id);
CREATE INDEX idx_allergy_substance ON allergy_intolerances(substance_code);
CREATE INDEX idx_allergy_criticality ON allergy_intolerances(criticality);

-- Flags (Sentinel alerts on patients)
CREATE TABLE flags (
    id TEXT PRIMARY KEY,
    patient_id TEXT NOT NULL REFERENCES patients(id),
    status TEXT NOT NULL,            -- active, inactive, entered-in-error
    category TEXT,
    code TEXT,
    period_start TEXT,
    period_end TEXT,
    generated_by TEXT,               -- sentinel rule ID
    site_id TEXT NOT NULL,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX idx_flag_patient ON flags(patient_id);
CREATE INDEX idx_flag_status ON flags(status);
CREATE INDEX idx_flag_category ON flags(category);

-- System-wide alerts (DetectedIssue)
CREATE TABLE detected_issues (
    id TEXT PRIMARY KEY,
    severity TEXT NOT NULL,
    code TEXT,
    detail TEXT,
    identified_datetime TEXT NOT NULL,
    status TEXT NOT NULL,
    implicated_sites TEXT,           -- JSON array
    implicated_patients TEXT,        -- JSON array
    generated_by TEXT,
    last_updated TEXT NOT NULL,
    git_blob_hash TEXT NOT NULL,
    fhir_json TEXT NOT NULL
);

CREATE INDEX idx_di_severity ON detected_issues(severity);
CREATE INDEX idx_di_status ON detected_issues(status);
CREATE INDEX idx_di_date ON detected_issues(identified_datetime);

-- Patient summary cache (denormalised for list views)
CREATE TABLE patient_summaries (
    patient_id TEXT PRIMARY KEY REFERENCES patients(id),
    encounter_count INTEGER DEFAULT 0,
    active_conditions INTEGER DEFAULT 0,
    active_medications INTEGER DEFAULT 0,
    active_allergies INTEGER DEFAULT 0,
    unresolved_alerts INTEGER DEFAULT 0,
    last_encounter_date TEXT,
    last_updated TEXT NOT NULL
);

-- Index rebuild metadata
CREATE TABLE index_meta (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- Stores: git_head (last indexed commit), rebuilt_at, resource_count
```

### 5.2 Index Upsert Strategy

When a FHIR resource is committed to Git, the service extracts indexed fields and performs an `INSERT OR REPLACE` into the corresponding table. The `git_blob_hash` column links the SQLite row back to the Git object for traceability.

The `patient_summaries` table is updated on every child resource mutation — incrementing/decrementing counts as encounters, conditions, medications, allergies, and flags are created or updated. This avoids expensive COUNT queries on the list endpoint.

### 5.3 FTS5 Sync

The `patients_fts` full-text search table is kept in sync with the `patients` table via triggers:

```sql
CREATE TRIGGER patients_ai AFTER INSERT ON patients BEGIN
    INSERT INTO patients_fts(rowid, id, family_name, given_names) 
    VALUES (new.rowid, new.id, new.family_name, new.given_names);
END;

CREATE TRIGGER patients_ad AFTER DELETE ON patients BEGIN
    INSERT INTO patients_fts(patients_fts, rowid, id, family_name, given_names) 
    VALUES ('delete', old.rowid, old.id, old.family_name, old.given_names);
END;

CREATE TRIGGER patients_au AFTER UPDATE ON patients BEGIN
    INSERT INTO patients_fts(patients_fts, rowid, id, family_name, given_names) 
    VALUES ('delete', old.rowid, old.id, old.family_name, old.given_names);
    INSERT INTO patients_fts(rowid, id, family_name, given_names) 
    VALUES (new.rowid, new.id, new.family_name, new.given_names);
END;
```

---

## 6. Read Operations

### 6.1 List Queries

All list endpoints query SQLite with pagination, filtering, and sorting. The response includes the `fhir_json` column directly — no Git read required for standard list/detail views.

**Standard pagination parameters:**

| Parameter | Default | Max |
|-----------|---------|-----|
| `page` | 1 | — |
| `per_page` | 25 | 100 |
| `sort` | `last_updated:desc` | Any indexed column |

### 6.2 Patient Bundle (GetPatientBundle)

Returns the Patient resource plus all child resources. Implementation:

```sql
-- Single patient
SELECT fhir_json FROM patients WHERE id = ?;

-- All child resources (parallel queries)
SELECT fhir_json FROM encounters WHERE patient_id = ? ORDER BY period_start DESC;
SELECT fhir_json FROM observations WHERE patient_id = ? ORDER BY effective_datetime DESC;
SELECT fhir_json FROM conditions WHERE patient_id = ? AND clinical_status = 'active';
SELECT fhir_json FROM medication_requests WHERE patient_id = ? AND status = 'active';
SELECT fhir_json FROM allergy_intolerances WHERE patient_id = ? AND clinical_status = 'active';
SELECT fhir_json FROM flags WHERE patient_id = ? AND status = 'active';
```

These six queries run concurrently using goroutines and are assembled into the bundle response.

### 6.3 Patient Search (FTS5)

```sql
SELECT p.fhir_json, ps.* 
FROM patients_fts fts
JOIN patients p ON p.id = fts.id
JOIN patient_summaries ps ON ps.patient_id = p.id
WHERE patients_fts MATCH ?
AND p.active = 1
ORDER BY rank
LIMIT ? OFFSET ?;
```

The FTS5 query supports prefix matching (`Ibra*`), phrase matching (`"Fatima Ibrahim"`), and boolean operators (`Ibrahim OR Okafor`).

### 6.4 Patient History (Git)

Version history reads directly from Git, not SQLite:

```go
// Pseudocode
func GetPatientHistory(patientID string) []CommitInfo {
    path := fmt.Sprintf("patients/%s/", patientID)
    return gitstore.LogPath(path, LogOptions{
        OrderBy: TimeDescending,
        Limit:   100,
    })
}
```

Each commit entry includes the hash, timestamp, author, site, operation, resource type, resource ID, and a diff summary.

### 6.5 Patient Timeline

The timeline endpoint returns a chronological view of all clinical events for a patient, assembled from encounters, observations, conditions, and flags, sorted by date:

```sql
SELECT 'encounter' as type, id, period_start as date, fhir_json FROM encounters WHERE patient_id = ?
UNION ALL
SELECT 'observation' as type, id, effective_datetime as date, fhir_json FROM observations WHERE patient_id = ?
UNION ALL
SELECT 'condition' as type, id, onset_datetime as date, fhir_json FROM conditions WHERE patient_id = ?
UNION ALL
SELECT 'flag' as type, id, period_start as date, fhir_json FROM flags WHERE patient_id = ?
ORDER BY date DESC
LIMIT ? OFFSET ?;
```

---

## 7. Patient Identity Matching

### 7.1 Algorithm

The `MatchPatient` RPC implements probabilistic identity matching for environments where patients lack formal identification. It uses a weighted scoring model:

| Factor | Weight | Match Logic |
|--------|--------|-------------|
| Family name exact | 0.30 | Case-insensitive, diacritics-normalised |
| Family name fuzzy | 0.20 | Levenshtein distance ≤ 2, or Soundex match |
| Given name exact | 0.15 | Any given name matches any given name |
| Given name fuzzy | 0.10 | Levenshtein distance ≤ 2 |
| Gender | 0.10 | Exact match |
| Birth year | 0.10 | Exact year match (from full or partial birthDate) |
| District/location | 0.05 | Exact match on address.district |

**Scoring:**
- Total score = sum of matched factor weights
- Default threshold: 0.7 (configurable per request)
- Returns all matches above threshold, sorted by confidence descending
- Maximum 10 matches returned

### 7.2 Implementation Strategy

```sql
-- Step 1: Broad candidate selection (fast, over-inclusive)
SELECT id, family_name, given_names, gender, birth_date, fhir_json
FROM patients
WHERE active = 1
AND (
    family_name LIKE ? || '%'              -- prefix match
    OR family_name IN (SELECT term FROM fuzzy_names(?))  -- pre-computed Soundex
    OR birth_date LIKE ? || '%'            -- year match
);

-- Step 2: Score each candidate in Go code using the weighted model
-- Step 3: Filter by threshold, sort by score, return top 10
```

The broad SQL query is intentionally over-inclusive to avoid false negatives. The expensive fuzzy matching and scoring happens in Go on the reduced candidate set (typically < 50 rows).

---

## 8. Batch Operations

### 8.1 CreateBatch

The batch endpoint allows creating multiple resources in a single atomic Git commit. This is used for:

- Recording a complete encounter (Encounter + Observations + Conditions + MedicationRequests in one commit)
- Bulk import during initial data migration
- Sentinel Agent writing multiple alerts from a single analysis cycle

```protobuf
message CreateBatchRequest {
  string patient_id = 1;
  repeated FhirResource resources = 2;  // Multiple resources
  MutationContext context = 3;
  bool atomic = 4;                       // If true, all-or-nothing
}

message BatchResponse {
  repeated BatchItemResult results = 1;
  GitCommitInfo git = 2;                 // Single commit for all
}

message BatchItemResult {
  string resource_type = 1;
  string resource_id = 2;
  bool success = 3;
  string error = 4;                      // Empty if success
}
```

**Behaviour:**
- If `atomic = true`: All resources are validated first. If any fail validation, none are written. All successful resources are committed in a single Git commit.
- If `atomic = false`: Each resource is validated and written independently. Failures don't block other resources. Still a single Git commit for all successful writes.

---

## 9. Index Rebuild

### 9.1 RebuildIndex RPC

Triggered manually, on startup health check failure, or by the Sync Service after a merge.

```
1. Drop all SQLite tables (except index_meta)
2. Recreate schema
3. Walk Git tree at HEAD
4. For each FHIR JSON file:
   a. Parse resource
   b. Extract indexed fields
   c. INSERT into appropriate table
5. Rebuild patient_summaries from counts
6. Rebuild FTS5 index
7. Write index_meta: git_head = current HEAD, resource_count = N
8. Return: duration, resources_indexed, git_head
```

**Performance target:** < 5 seconds per 10,000 resources on Raspberry Pi 4.

### 9.2 CheckIndexHealth RPC

Runs on service startup and can be called on demand:

```
1. Read index_meta.git_head
2. Compare against current Git HEAD
3. If different:
   a. Count resources in Git tree
   b. Count resources in SQLite
   c. If mismatch > 0: index is stale, return unhealthy
4. If same: return healthy with resource count
```

If unhealthy, the service auto-triggers a rebuild before accepting write requests.

---

## 10. Interaction with Other Services

### 10.1 Sync Service → Patient Service

After a successful merge, the Sync Service calls `RebuildIndex` or performs incremental re-indexing by calling the Patient Service with each new/modified resource path from the Git diff. The Patient Service reads the file from Git and upserts SQLite.

The Sync Service does NOT call `CreatePatient` / `CreateEncounter` etc. for merged records — those have already been committed to Git by the merge. The Patient Service just needs to re-index them.

```protobuf
// Called by Sync Service after merge
rpc ReindexResources(ReindexRequest) returns (ReindexResponse);

message ReindexRequest {
  repeated string resource_paths = 1;  // Git paths of new/modified files
}

message ReindexResponse {
  int32 indexed = 1;
  int32 failed = 2;
  repeated string errors = 3;
}
```

### 10.2 Sentinel Agent → Patient Service

The Sentinel Agent writes alerts by calling `CreateFlag` (patient-level) and creating DetectedIssue resources (system-level) through the Patient Service. This maintains the single-writer principle.

```
Sentinel Agent
    │
    │  gRPC: CreateFlag / CreateBatch
    ▼
Patient Service
    │
    │  Git commit + SQLite index
    ▼
Data Layer
```

### 10.3 Formulary Service ← Patient Service

The Patient Service does NOT call the Formulary Service directly. Drug interaction checking is the API Gateway's responsibility — it calls the Formulary Service before forwarding the `CreateMedicationRequest` to the Patient Service. The Patient Service receives pre-validated requests only.

---

## 11. Error Handling

### 11.1 Error Types

| gRPC Code | Condition | Recovery |
|-----------|-----------|----------|
| `INVALID_ARGUMENT` | FHIR validation failure | Return field-level errors |
| `NOT_FOUND` | Patient or resource doesn't exist | — |
| `ALREADY_EXISTS` | Duplicate resource ID | Return existing resource |
| `FAILED_PRECONDITION` | Index unhealthy, rebuild in progress | Retry after rebuild |
| `ABORTED` | Write lock timeout (5s) | Retry |
| `INTERNAL` | Git write failure | Log, alert, do not retry automatically |
| `INTERNAL` | SQLite write failure after Git success | Log, schedule re-index, return success with warning |

### 11.2 Git Write Failure Recovery

If a Git commit fails (disk full, corrupted repo):

1. Rollback the working tree to the last good commit (`git checkout HEAD`)
2. Do NOT upsert SQLite
3. Return `INTERNAL` error with `GIT_WRITE_FAILED`
4. Log with high severity for ops alerting

### 11.3 SQLite Failure After Git Success

If Git commit succeeds but SQLite upsert fails:

1. Return success to the caller (data is safe in Git)
2. Include a warning: `SQLITE_INDEX_STALE`
3. Schedule an async re-index for the affected resource
4. If re-index also fails, schedule a full rebuild

The data is never lost — Git is the source of truth.

---

## 12. Performance Targets

All targets measured on Raspberry Pi 4 (4GB RAM).

| Operation | Target | Notes |
|-----------|--------|-------|
| Create Patient (full pipeline) | < 200ms | Validate + Git commit + SQLite upsert |
| Create Encounter | < 200ms | Same pipeline |
| Create Observation | < 150ms | Simpler validation |
| Get Patient Bundle | < 150ms | 6 parallel SQLite queries |
| List Patients (25 per page) | < 100ms | SQLite index scan |
| Search Patients (FTS5) | < 100ms | Full-text search |
| Match Patient | < 300ms | Broad query + in-memory scoring |
| Get Patient History (Git log) | < 500ms | Git log on patient directory |
| Create Batch (10 resources) | < 500ms | Single commit, 10 SQLite upserts |
| Rebuild Index (10,000 resources) | < 5s | Full drop and rebuild |
| Write lock acquisition | < 5s timeout | Fails with ABORTED if exceeded |

### 12.1 Memory Budget

| Component | Target |
|-----------|--------|
| Service process RSS | < 80MB |
| SQLite database file | ~1KB per resource (10,000 resources ≈ 10MB) |
| Git repository | ~2KB per resource with delta compression |
| Write lock contention | < 1% of requests wait at expected load |

---

## 13. Configuration

```yaml
patient_service:
  grpc_port: 50051
  
  git:
    repo_path: /var/lib/open-nucleus/data
    author_name: "open-nucleus"
    author_email: "system@open-nucleus.local"
    
  sqlite:
    db_path: /var/lib/open-nucleus/index.db
    journal_mode: WAL               # Write-Ahead Logging for concurrent reads
    busy_timeout: 5000              # ms
    cache_size: -20000              # 20MB page cache
    
  validation:
    strict_mode: true               # Reject unknown fields
    require_icd10: true             # Require ICD-10 codes on conditions
    require_loinc: true             # Require LOINC codes on observations
    
  write_lock:
    timeout: 5s
    
  index:
    auto_rebuild_on_drift: true
    health_check_on_startup: true
    
  matching:
    default_threshold: 0.7
    max_results: 10
    fuzzy_max_distance: 2           # Levenshtein distance for fuzzy name matching
    
  logging:
    level: info
    format: json
```

---

## 14. Testing Strategy

### 14.1 Unit Tests

| Area | Coverage Target | Focus |
|------|----------------|-------|
| FHIR validation | 95% | Every required field, every resource type, edge cases (partial dates, missing codes) |
| Git path derivation | 100% | Correct path for every resource type |
| SQLite indexing | 90% | Extraction of all indexed fields from all resource types |
| Patient matching | 90% | Scoring accuracy across name variations, missing fields |
| Commit message formatting | 100% | Structured format parsing roundtrip |

### 14.2 Integration Tests

| Test | Description |
|------|-------------|
| Write roundtrip | Create resource → verify Git commit → verify SQLite row → read back via query |
| Batch atomicity | Batch with one invalid resource → verify nothing written (atomic=true) |
| Soft delete | Delete patient → verify still in Git → verify excluded from active queries |
| Index rebuild | Write 1000 resources → drop SQLite → rebuild → verify all present |
| Concurrent reads | Write under load + concurrent read queries → verify no lock contention on reads |
| Health check | Manually desync SQLite → verify health check detects → verify auto-rebuild |

### 14.3 Benchmark Tests

Run on Raspberry Pi 4 as part of CI:

- `BenchmarkCreatePatient` — target < 200ms p99
- `BenchmarkListPatients100` — target < 100ms p99
- `BenchmarkSearchFTS5` — target < 100ms p99
- `BenchmarkGetPatientBundle` — target < 150ms p99
- `BenchmarkRebuildIndex10k` — target < 5s

---

*Open Nucleus • Patient Service Specification V1 • FibrinLab*  
*github.com/FibrinLab/open-nucleus*
