# Open Nucleus — API Gateway Specification V1

**Version:** 1.0  
**Date:** February 2026  
**Author:** Akanimoh Osutuk — FibrinLab  
**Repo:** github.com/FibrinLab/open-nucleus  
**Status:** Draft — V1 Specification

---

## 1. Gateway Overview

### 1.1 Role

The API Gateway is the sole entry point for the Flutter frontend and any future external consumers. It is a stateless Go HTTP server that translates REST/JSON requests into gRPC calls on the backend microservices. The gateway owns no business logic beyond authentication, authorisation, request validation, rate limiting, and response formatting.

### 1.2 Base Configuration

| Property | Value |
|----------|-------|
| Base URL | `http://localhost:8080/api/v1` |
| Protocol | HTTP/1.1 + HTTP/2 (REST/JSON) |
| Authentication | Bearer JWT (Ed25519-signed) |
| Content-Type | `application/json` (all requests and responses) |
| FHIR Version | R4 (4.0.1) |

### 1.3 Design Constraints

- **Stateless:** No session storage. All state lives in the JWT token and backend services.
- **Single port:** All traffic enters through port 8080. No service is directly exposed to the frontend.
- **FHIR-native responses:** Clinical data in response payloads is valid FHIR R4 JSON. The gateway does not transform FHIR resources.
- **Offline-compatible:** The gateway runs on the same device as the frontend. Latency is sub-millisecond. No internet dependency.

### 1.4 Request Lifecycle (Middleware Pipeline)

```
Flutter App
    │
    │  HTTP REST/JSON
    ▼
┌────────────────────────────────────────────────┐
│  1. Rate Limiter                               │
│  2. Request ID Generator (X-Request-ID)        │
│  3. JWT Validator (except /auth/*)              │
│  4. RBAC Enforcer (check role vs endpoint)      │
│  5. Request Validator (JSON schema)             │
│  6. gRPC Router (→ backend service)             │
│  7. Response Formatter (envelope + status)      │
│  8. Audit Logger                                │
└────────────────────────────────────────────────┘
    │
    │  gRPC
    ▼
Backend Microservice
```

### 1.5 Backend Service Map

| Service | gRPC Port | Language |
|---------|-----------|----------|
| Patient Service | 50051 | Go |
| Sync Service | 50052 | Go |
| Auth Service | 50053 | Go |
| Formulary Service | 50054 | Go |
| Anchor Service | 50055 | Go |
| Sentinel Agent | 50056 | Python |

---

## 2. Authentication

### 2.1 Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/login` | Authenticate device and receive JWT |
| `POST` | `/api/v1/auth/refresh` | Refresh an expiring JWT token |
| `POST` | `/api/v1/auth/logout` | Invalidate current token (local only) |
| `GET` | `/api/v1/auth/whoami` | Return current user identity and role |

### 2.2 POST /api/v1/auth/login

Authenticates the device using its Ed25519 keypair. The device signs a challenge nonce with its private key. The Auth Service verifies the signature against the registered public key.

**Request:**

```json
{
  "device_id": "node-sheffield-01",
  "public_key": "MCowBQYDK2VwAyEA...",
  "challenge_response": {
    "nonce": "a1b2c3d4e5f6...",
    "signature": "SGVsbG8gV29ybGQ...",
    "timestamp": "2026-02-25T09:00:00Z"
  },
  "practitioner_id": "dr-adeleye"
}
```

**Response (200):**

```json
{
  "status": "success",
  "data": {
    "token": "eyJhbGciOiJFZERTQSIs...",
    "expires_at": "2026-02-26T09:00:00Z",
    "refresh_token": "dGhpcyBpcyBhIHJl...",
    "role": {
      "code": "physician",
      "display": "Physician",
      "permissions": [
        "patient:read", "patient:write", "encounter:write",
        "medication:write", "conflict:resolve", "alert:read"
      ]
    },
    "site_id": "clinic-maiduguri-03",
    "node_id": "node-sheffield-01"
  }
}
```

### 2.3 JWT Structure

Tokens are Ed25519-signed JWTs verified locally without network access.

```json
{
  "sub": "dr-adeleye",
  "node": "node-sheffield-01",
  "site": "clinic-maiduguri-03",
  "role": "physician",
  "permissions": ["patient:read", "patient:write", "..."],
  "iat": 1740470400,
  "exp": 1740556800,
  "iss": "open-nucleus-auth"
}
```

- **Algorithm:** EdDSA (Ed25519)
- **Lifetime:** 24 hours (configurable per deployment)
- **Refresh window:** Final 2 hours before expiry
- **Revocation:** Local deny list, propagated via Git sync

---

## 3. Patient Endpoints

### 3.1 Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/patients` | List patients (paginated, filterable) |
| `POST` | `/api/v1/patients` | Create a new patient |
| `GET` | `/api/v1/patients/:id` | Get full patient bundle |
| `PUT` | `/api/v1/patients/:id` | Update patient demographics |
| `DELETE` | `/api/v1/patients/:id` | Soft-delete (mark inactive) |
| `GET` | `/api/v1/patients/:id/history` | Git version history |
| `GET` | `/api/v1/patients/:id/timeline` | Chronological encounter timeline |
| `GET` | `/api/v1/patients/search` | Full-text search (FTS5) |
| `POST` | `/api/v1/patients/match` | Probabilistic identity matching |

### 3.2 GET /api/v1/patients

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-indexed) |
| `per_page` | integer | 25 (max 100) | Results per page |
| `sort` | string | `updated_at:desc` | Sort field and direction |
| `gender` | string | — | Filter: male, female, other, unknown |
| `birth_date_from` | date | — | Filter: born on or after this date |
| `birth_date_to` | date | — | Filter: born on or before this date |
| `site_id` | string | Current site | Filter by originating site |
| `status` | string | `active` | Filter: active, inactive, all |
| `has_alerts` | boolean | — | Filter patients with active Sentinel flags |

**Response (200):**

```json
{
  "status": "success",
  "data": [
    {
      "id": "patient-uuid-001",
      "resourceType": "Patient",
      "name": [{ "family": "Okafor", "given": ["Chidi"] }],
      "gender": "male",
      "birthDate": "2022-06-15",
      "meta": {
        "lastUpdated": "2026-02-24T14:30:00Z",
        "source": "clinic-maiduguri-03",
        "versionId": "a1b2c3d4"
      },
      "_summary": {
        "encounter_count": 7,
        "active_conditions": 2,
        "active_medications": 1,
        "unresolved_alerts": 1,
        "last_encounter": "2026-02-20T09:15:00Z"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 25,
    "total": 342,
    "total_pages": 14
  }
}
```

### 3.3 POST /api/v1/patients

**Request:**

```json
{
  "resourceType": "Patient",
  "name": [{ "family": "Ibrahim", "given": ["Fatima"] }],
  "gender": "female",
  "birthDate": "2024-03-10",
  "address": [{ "district": "Maiduguri", "state": "Borno" }],
  "identifier": [
    { "system": "open-nucleus:local", "value": "auto-generated" },
    { "system": "national-id", "value": "NG-BRN-2024-00451" }
  ]
}
```

**Response (201):**

```json
{
  "status": "success",
  "data": {
    "id": "patient-uuid-002",
    "resourceType": "Patient",
    "...": "full FHIR Patient resource",
    "meta": {
      "versionId": "initial",
      "lastUpdated": "2026-02-25T09:42:00Z",
      "source": "clinic-maiduguri-03"
    }
  },
  "git": {
    "commit": "e4f5a6b7c8d9...",
    "message": "[Patient] CREATE patient-uuid-002"
  }
}
```

### 3.4 GET /api/v1/patients/:id

Returns the complete patient bundle: the Patient resource plus all child resources (encounters, observations, conditions, medication requests, allergies, and flags).

**Response (200):**

```json
{
  "status": "success",
  "data": {
    "patient": { "resourceType": "Patient", "..." },
    "encounters": [{ "resourceType": "Encounter", "..." }],
    "observations": [{ "resourceType": "Observation", "..." }],
    "conditions": [{ "resourceType": "Condition", "..." }],
    "medication_requests": [{ "resourceType": "MedicationRequest", "..." }],
    "allergy_intolerances": [{ "resourceType": "AllergyIntolerance", "..." }],
    "flags": [{ "resourceType": "Flag", "..." }]
  }
}
```

### 3.5 GET /api/v1/patients/:id/history

Returns the Git version history for the patient directory, providing a complete audit trail.

**Response (200):**

```json
{
  "status": "success",
  "data": [
    {
      "commit_hash": "e4f5a6b7c8d9...",
      "timestamp": "2026-02-24T14:30:00Z",
      "author": "dr-adeleye",
      "node": "node-sheffield-01",
      "site": "clinic-maiduguri-03",
      "operation": "UPDATE",
      "resource_type": "Observation",
      "resource_id": "obs-uuid-015",
      "message": "[Observation] CREATE obs-uuid-015",
      "diff_summary": { "files_changed": 1, "insertions": 24, "deletions": 0 }
    }
  ]
}
```

### 3.6 POST /api/v1/patients/match

Probabilistic identity matching for patients without formal identification.

**Request:**

```json
{
  "name": { "family": "Ibrahim", "given": ["Fatima"] },
  "gender": "female",
  "birth_date_approx": "2024",
  "district": "Maiduguri",
  "threshold": 0.7
}
```

**Response (200):**

```json
{
  "status": "success",
  "data": {
    "matches": [
      {
        "patient_id": "patient-uuid-002",
        "confidence": 0.92,
        "match_factors": ["name_exact", "gender", "district", "birth_year"]
      },
      {
        "patient_id": "patient-uuid-119",
        "confidence": 0.74,
        "match_factors": ["name_fuzzy", "gender", "birth_year"]
      }
    ]
  }
}
```

---

## 4. Clinical Resource Endpoints

All clinical resource endpoints follow the same pattern. They are nested under the patient path, accept FHIR R4 JSON, and return the committed resource with Git metadata.

### 4.1 Encounters

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/patients/:id/encounters` | List encounters for patient |
| `POST` | `/api/v1/patients/:id/encounters` | Record new encounter |
| `GET` | `/api/v1/patients/:id/encounters/:eid` | Get specific encounter |
| `PUT` | `/api/v1/patients/:id/encounters/:eid` | Update encounter |

**POST Request Body:**

```json
{
  "resourceType": "Encounter",
  "status": "in-progress",
  "class": { "code": "AMB", "display": "Ambulatory" },
  "type": [{
    "coding": [{
      "system": "http://snomed.info/sct",
      "code": "185345009",
      "display": "Encounter for symptom"
    }]
  }],
  "subject": { "reference": "Patient/patient-uuid-002" },
  "period": { "start": "2026-02-25T10:00:00Z" },
  "reasonCode": [{
    "coding": [{
      "system": "http://hl7.org/fhir/sid/icd-10",
      "code": "J06.9",
      "display": "Acute upper respiratory infection"
    }]
  }],
  "participant": [{
    "individual": { "reference": "Practitioner/dr-adeleye" }
  }]
}
```

### 4.2 Observations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/patients/:id/observations` | List observations (filterable by code, date) |
| `POST` | `/api/v1/patients/:id/observations` | Record new observation |
| `GET` | `/api/v1/patients/:id/observations/:oid` | Get specific observation |

**Query Parameters for GET:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `code` | string | LOINC code filter |
| `category` | string | vital-signs, laboratory, social-history, etc. |
| `date_from` | datetime | Observations after this date |
| `date_to` | datetime | Observations before this date |
| `encounter_id` | string | Filter by encounter |

**POST Request Body (Vital Signs Example):**

```json
{
  "resourceType": "Observation",
  "status": "final",
  "category": [{
    "coding": [{
      "system": "http://terminology.hl7.org/CodeSystem/observation-category",
      "code": "vital-signs"
    }]
  }],
  "code": {
    "coding": [{
      "system": "http://loinc.org",
      "code": "8310-5",
      "display": "Body temperature"
    }]
  },
  "subject": { "reference": "Patient/patient-uuid-002" },
  "encounter": { "reference": "Encounter/enc-uuid-010" },
  "effectiveDateTime": "2026-02-25T10:15:00Z",
  "valueQuantity": {
    "value": 38.4,
    "unit": "Cel",
    "system": "http://unitsofmeasure.org",
    "code": "Cel"
  }
}
```

### 4.3 Conditions

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/patients/:id/conditions` | List conditions (active, resolved, all) |
| `POST` | `/api/v1/patients/:id/conditions` | Record diagnosis |
| `PUT` | `/api/v1/patients/:id/conditions/:cid` | Update condition status |

**Query Parameters for GET:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `clinical_status` | string | active, recurrence, relapse, inactive, remission, resolved |
| `category` | string | encounter-diagnosis, problem-list-item |
| `code` | string | ICD-10 code filter |

### 4.4 Medication Requests

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/patients/:id/medication-requests` | List prescriptions |
| `POST` | `/api/v1/patients/:id/medication-requests` | Create prescription |
| `PUT` | `/api/v1/patients/:id/medication-requests/:mid` | Update prescription status |

When a MedicationRequest is created, the gateway automatically calls the Formulary Service to check for drug interactions against the patient's active medications and allergies. If a major interaction is detected, the response includes a warnings array:

**Response (201) with Drug Interaction Warning:**

```json
{
  "status": "success",
  "data": { "resourceType": "MedicationRequest", "..." },
  "warnings": [
    {
      "severity": "high",
      "type": "drug-interaction",
      "description": "Major interaction: Methotrexate + Trimethoprim. Risk of bone marrow suppression.",
      "interacting_medication": "medication-request-uuid-005",
      "source": "formulary-service"
    }
  ],
  "git": { "commit": "..." }
}
```

### 4.5 Allergy Intolerances

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/patients/:id/allergy-intolerances` | List allergies |
| `POST` | `/api/v1/patients/:id/allergy-intolerances` | Record allergy |
| `PUT` | `/api/v1/patients/:id/allergy-intolerances/:aid` | Update allergy |

---

## 5. Synchronisation Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/sync/status` | Current sync state |
| `GET` | `/api/v1/sync/peers` | List discovered peer nodes |
| `POST` | `/api/v1/sync/trigger` | Manually trigger sync |
| `GET` | `/api/v1/sync/history` | Sync event log |
| `POST` | `/api/v1/sync/bundle/export` | Export Git bundle to file |
| `POST` | `/api/v1/sync/bundle/import` | Import Git bundle from file |

### 5.1 GET /api/v1/sync/status

```json
{
  "status": "success",
  "data": {
    "node_id": "node-sheffield-01",
    "git_head": "e4f5a6b7c8d9...",
    "total_commits": 1847,
    "total_fhir_resources": 12340,
    "last_sync": {
      "timestamp": "2026-02-25T08:30:00Z",
      "peer": "node-maiduguri-hub",
      "transport": "wifi-direct",
      "records_received": 45,
      "records_sent": 23,
      "conflicts": 2,
      "duration_ms": 4200
    },
    "available_transports": [
      { "type": "wifi-direct", "status": "active", "peers_found": 2 },
      { "type": "bluetooth", "status": "scanning", "peers_found": 0 },
      { "type": "local-network", "status": "active", "peers_found": 1 }
    ],
    "pending_conflicts": 2,
    "queued_outbound_records": 0
  }
}
```

### 5.2 GET /api/v1/sync/peers

```json
{
  "status": "success",
  "data": [
    {
      "peer_id": "node-maiduguri-hub",
      "peer_name": "Maiduguri Regional Hub",
      "transport": "wifi-direct",
      "signal_strength": "strong",
      "last_synced": "2026-02-25T08:30:00Z",
      "their_head": "f6g7h8i9j0...",
      "commits_behind": 23,
      "commits_ahead": 12
    }
  ]
}
```

### 5.3 POST /api/v1/sync/trigger

**Request:**

```json
{
  "peer_id": "node-maiduguri-hub",
  "transport": "wifi-direct",
  "priority": "normal"
}
```

Both `peer_id` and `transport` are optional. If omitted, the Sync Service selects the best available peer and transport automatically. Setting `priority` to `"critical"` causes only high-priority records (active encounters, alerts) to sync — useful for bandwidth-constrained transports.

**Response (202):**

```json
{
  "status": "success",
  "data": {
    "sync_id": "sync-uuid-001",
    "state": "initiated",
    "peer": "node-maiduguri-hub",
    "transport": "wifi-direct"
  }
}
```

### 5.4 POST /api/v1/sync/bundle/export

**Request:**

```json
{
  "since_commit": "a1b2c3d4...",
  "output_path": "/media/usb/nucleus-bundle-20260225.bundle"
}
```

**Response (200):**

```json
{
  "status": "success",
  "data": {
    "bundle_path": "/media/usb/nucleus-bundle-20260225.bundle",
    "commits_included": 145,
    "size_bytes": 2340000,
    "from_commit": "a1b2c3d4...",
    "to_commit": "e4f5a6b7..."
  }
}
```

### 5.5 POST /api/v1/sync/bundle/import

**Request:**

```json
{
  "bundle_path": "/media/usb/nucleus-bundle-20260225.bundle"
}
```

---

## 6. Conflict Resolution Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/conflicts` | List all unresolved conflicts |
| `GET` | `/api/v1/conflicts/:id` | Get conflict detail with diff |
| `POST` | `/api/v1/conflicts/:id/resolve` | Resolve a conflict |
| `POST` | `/api/v1/conflicts/:id/defer` | Defer resolution (propagates as unresolved) |

### 6.1 Conflict Levels

| Level | Condition | Action |
|-------|-----------|--------|
| `auto-merge` | Additive, non-overlapping changes | Merged automatically, logged |
| `review` | Overlapping changes, no clinical risk | Merged, flagged for clinician review |
| `block` | Contradictory clinical data with safety implications | Rejected, requires explicit resolution |

### 6.2 GET /api/v1/conflicts/:id

```json
{
  "status": "success",
  "data": {
    "id": "conflict-uuid-001",
    "level": "review",
    "resource_type": "MedicationRequest",
    "patient_id": "patient-uuid-002",
    "created_at": "2026-02-25T08:30:00Z",
    "local_version": {
      "resource": { "resourceType": "MedicationRequest", "..." },
      "commit": "a1b2c3...",
      "author": "dr-adeleye",
      "site": "clinic-maiduguri-03",
      "timestamp": "2026-02-24T16:00:00Z"
    },
    "incoming_version": {
      "resource": { "resourceType": "MedicationRequest", "..." },
      "commit": "d4e5f6...",
      "author": "nurse-aisha",
      "site": "clinic-bama-01",
      "timestamp": "2026-02-24T17:30:00Z"
    },
    "diff": {
      "changed_fields": ["dosageInstruction[0].doseAndRate[0].doseQuantity.value"],
      "local_value": "250",
      "incoming_value": "500"
    }
  }
}
```

### 6.3 POST /api/v1/conflicts/:id/resolve

**Request:**

```json
{
  "resolution": "accept_incoming",
  "reason": "Confirmed with Dr. Aisha - dose increase clinically appropriate",
  "reviewed_by": "dr-adeleye"
}
```

The `resolution` field accepts:
- `accept_local` — keep local version
- `accept_incoming` — accept remote version
- `merge_custom` — provide a manually merged resource in a `custom_resource` field

All resolutions are committed to Git with full audit metadata.

### 6.4 Clinical Safety Rules (Always Block)

These resource types and field combinations always escalate to `block` level:

- **AllergyIntolerance:** Any conflicting modification to substance or criticality fields
- **MedicationRequest:** Concurrent active prescriptions where the drug interaction database flags a major interaction
- **Condition:** Conflicting modifications to the same active diagnosis (different clinicalStatus or verificationStatus)
- **Patient demographics:** Conflicting changes to identifiers, birth date, or gender

---

## 7. Sentinel Alert Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/alerts` | List all alerts (paginated, filterable) |
| `GET` | `/api/v1/alerts/:id` | Get alert detail |
| `POST` | `/api/v1/alerts/:id/acknowledge` | Acknowledge an alert |
| `POST` | `/api/v1/alerts/:id/dismiss` | Dismiss (false positive) |
| `GET` | `/api/v1/alerts/summary` | Dashboard summary counts |

### 7.1 GET /api/v1/alerts

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `severity` | string | — | critical, high, moderate, low |
| `category` | string | — | outbreak, medication-safety, missed-referral, stockout, vital-trend |
| `status` | string | `active` | active, acknowledged, dismissed |
| `patient_id` | string | — | Filter alerts for specific patient |
| `site_id` | string | — | Filter alerts from specific site |
| `since` | datetime | — | Alerts generated after this timestamp |

**Response (200):**

```json
{
  "status": "success",
  "data": [
    {
      "id": "alert-uuid-001",
      "resourceType": "DetectedIssue",
      "severity": "high",
      "category": "outbreak",
      "code": {
        "coding": [{
          "system": "http://hl7.org/fhir/sid/icd-10",
          "code": "A00.9",
          "display": "Cholera, unspecified"
        }]
      },
      "detail": "8 cases of acute watery diarrhoea across 3 sites in 10 days. Matches WHO IDSR cholera threshold.",
      "identified": "2026-02-25T08:35:00Z",
      "implicated_sites": ["clinic-maiduguri-03", "clinic-bama-01", "clinic-dikwa-02"],
      "implicated_patients": 8,
      "status": "active",
      "generated_by": "sentinel-rule:outbreak-clustering-v1"
    }
  ]
}
```

### 7.2 GET /api/v1/alerts/summary

```json
{
  "status": "success",
  "data": {
    "total_active": 5,
    "by_severity": { "critical": 1, "high": 2, "moderate": 1, "low": 1 },
    "by_category": {
      "outbreak": 1,
      "medication-safety": 2,
      "missed-referral": 1,
      "stockout": 1,
      "vital-trend": 0
    },
    "last_generated": "2026-02-25T08:35:00Z"
  }
}
```

### 7.3 POST /api/v1/alerts/:id/acknowledge

**Request:**

```json
{
  "acknowledged_by": "dr-adeleye",
  "action_taken": "Notified district health office. Initiated cholera response protocol.",
  "follow_up_required": true
}
```

---

## 8. Formulary Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/formulary/medications` | Search medication database |
| `GET` | `/api/v1/formulary/medications/:code` | Get medication detail |
| `POST` | `/api/v1/formulary/check-interactions` | Check drug interactions |
| `GET` | `/api/v1/formulary/availability/:site_id` | Site stock levels |
| `PUT` | `/api/v1/formulary/availability/:site_id` | Update site stock levels |

### 8.1 GET /api/v1/formulary/medications

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Free-text search (name, generic name, ATC code) |
| `atc_code` | string | ATC classification filter |
| `in_stock` | boolean | Filter by current site availability |
| `who_essential` | boolean | Filter WHO Essential Medicines List only |

### 8.2 POST /api/v1/formulary/check-interactions

**Request:**

```json
{
  "patient_id": "patient-uuid-002",
  "proposed_medication": {
    "code": "J01EA01",
    "display": "Trimethoprim",
    "dose": "200mg",
    "frequency": "BD"
  }
}
```

**Response (200):**

```json
{
  "status": "success",
  "data": {
    "safe": false,
    "interactions": [
      {
        "severity": "major",
        "type": "drug-drug",
        "existing_medication": "Methotrexate (L04AX03)",
        "mechanism": "Trimethoprim inhibits renal excretion of methotrexate, increasing toxicity risk",
        "recommendation": "Avoid combination. Consider alternative antibiotic."
      }
    ],
    "allergy_alerts": [],
    "stock_available": true,
    "stock_quantity": 240
  }
}
```

---

## 9. IOTA Integrity Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/anchor/status` | Current anchoring status |
| `POST` | `/api/v1/anchor/verify` | Verify data integrity |
| `GET` | `/api/v1/anchor/history` | List all anchoring events |
| `POST` | `/api/v1/anchor/trigger` | Manually trigger anchoring (hub only) |

### 9.1 GET /api/v1/anchor/status

```json
{
  "status": "success",
  "data": {
    "last_anchor": {
      "timestamp": "2026-02-24T00:00:00Z",
      "git_commit": "e4f5a6b7c8d9...",
      "merkle_root": "3f4a5b6c7d8e9f...",
      "tangle_message_id": "iota1qz7...",
      "network": "iota-mainnet",
      "verified": true
    },
    "local_integrity": {
      "computed_merkle_root": "3f4a5b6c7d8e9f...",
      "matches_anchor": true,
      "commits_since_anchor": 23,
      "resources_since_anchor": 45
    },
    "next_anchor_scheduled": "2026-02-25T00:00:00Z"
  }
}
```

### 9.2 POST /api/v1/anchor/verify

Recomputes the Merkle root from the local Git object store and compares it against the most recent IOTA-anchored proof.

**Response (200):**

```json
{
  "status": "success",
  "data": {
    "verified": true,
    "commit_verified": "e4f5a6b7c8d9...",
    "computed_root": "3f4a5b6c7d8e9f...",
    "anchored_root": "3f4a5b6c7d8e9f...",
    "tangle_message_id": "iota1qz7...",
    "verification_method": "cached_proof",
    "timestamp": "2026-02-25T09:50:00Z"
  }
}
```

### 9.3 POST /api/v1/anchor/trigger

**Request:**

```json
{
  "include_metadata": {
    "facility_name": "Maiduguri Regional Hub",
    "reporting_period": "2026-02",
    "deployment_id": "deploy-borno-001"
  }
}
```

**Response (202):**

```json
{
  "status": "success",
  "data": {
    "anchor_id": "anchor-uuid-001",
    "state": "pending",
    "git_commit": "e4f5a6b7c8d9...",
    "merkle_root": "3f4a5b6c7d8e9f..."
  }
}
```

---

## 10. Supply Chain Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/supply/inventory` | Current inventory by site |
| `GET` | `/api/v1/supply/inventory/:item_code` | Stock detail for specific item |
| `POST` | `/api/v1/supply/deliveries` | Record a supply delivery |
| `GET` | `/api/v1/supply/predictions` | Stockout predictions (Sentinel) |
| `GET` | `/api/v1/supply/redistribution` | Redistribution suggestions (Sentinel) |

### 10.1 GET /api/v1/supply/inventory

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `site_id` | string | Filter by site (default: current site) |
| `category` | string | antibiotics, antimalarials, vaccines, consumables, etc. |
| `low_stock` | boolean | Filter items below reorder threshold |

### 10.2 GET /api/v1/supply/predictions

```json
{
  "status": "success",
  "data": [
    {
      "item_code": "J01CA04",
      "item_display": "Amoxicillin 250mg capsules",
      "site_id": "clinic-dikwa-02",
      "current_stock": 45,
      "daily_consumption_rate": 4.2,
      "projected_stockout_date": "2026-03-07",
      "confidence": 0.85,
      "recommendation": "Reorder 200 units or request redistribution from clinic-maiduguri-03 (surplus: 380 units)"
    }
  ]
}
```

### 10.3 GET /api/v1/supply/redistribution

```json
{
  "status": "success",
  "data": [
    {
      "item_code": "J01CA04",
      "item_display": "Amoxicillin 250mg capsules",
      "from_site": "clinic-maiduguri-03",
      "from_stock": 380,
      "to_site": "clinic-dikwa-02",
      "to_stock": 45,
      "suggested_transfer": 150,
      "urgency": "high",
      "rationale": "clinic-dikwa-02 projected stockout in 10 days; clinic-maiduguri-03 has 90-day surplus"
    }
  ]
}
```

### 10.4 POST /api/v1/supply/deliveries

**Request:**

```json
{
  "resourceType": "SupplyDelivery",
  "status": "completed",
  "type": { "coding": [{ "code": "medication", "display": "Medication" }] },
  "suppliedItem": {
    "quantity": { "value": 500, "unit": "capsules" },
    "itemCodeableConcept": {
      "coding": [{
        "system": "http://www.whocc.no/atc",
        "code": "J01CA04",
        "display": "Amoxicillin"
      }]
    }
  },
  "occurrenceDateTime": "2026-02-25T11:00:00Z",
  "destination": { "reference": "Location/clinic-maiduguri-03" },
  "supplier": { "display": "UNICEF Supply Division" }
}
```

---

## 11. Response Envelope

### 11.1 Standard Envelope

Every API response is wrapped in a consistent envelope:

```json
{
  "status": "success | error",
  "data": {},
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description",
    "details": {}
  },
  "pagination": {
    "page": 1,
    "per_page": 25,
    "total": 342,
    "total_pages": 14
  },
  "warnings": [],
  "git": {
    "commit": "e4f5a6b7...",
    "message": "[Resource] OPERATION id"
  },
  "meta": {
    "request_id": "req-uuid-001",
    "duration_ms": 42,
    "node_id": "node-sheffield-01"
  }
}
```

Fields `data`, `error`, `pagination`, `warnings`, `git`, and `meta` are all optional depending on the endpoint and outcome. The `status` field is always present.

### 11.2 Error Codes

| HTTP | Error Code | Description |
|------|------------|-------------|
| 400 | `VALIDATION_ERROR` | Request body fails JSON schema or FHIR R4 validation |
| 400 | `INVALID_FHIR_RESOURCE` | FHIR resource is syntactically valid but clinically incomplete |
| 401 | `AUTH_REQUIRED` | No JWT token provided |
| 401 | `TOKEN_EXPIRED` | JWT has expired, use /auth/refresh |
| 401 | `TOKEN_REVOKED` | Device has been decommissioned |
| 403 | `INSUFFICIENT_PERMISSIONS` | Role does not permit this operation |
| 403 | `SITE_SCOPE_VIOLATION` | Attempting to access data outside assigned site scope |
| 404 | `RESOURCE_NOT_FOUND` | Requested resource does not exist |
| 409 | `MERGE_CONFLICT` | Write conflicts with a pending unresolved merge conflict |
| 409 | `DUPLICATE_RESOURCE` | Resource with this identifier already exists |
| 422 | `CLINICAL_SAFETY_BLOCK` | Operation blocked by clinical safety rules |
| 429 | `RATE_LIMITED` | Too many requests, retry after indicated duration |
| 500 | `INTERNAL_ERROR` | Unexpected server error |
| 500 | `GIT_WRITE_FAILED` | Failed to commit to Git repository |
| 500 | `SQLITE_INDEX_FAILED` | Git commit succeeded but SQLite indexing failed (data safe, index stale) |
| 503 | `SERVICE_UNAVAILABLE` | Backend microservice is down or restarting |

### 11.3 Error Response Format

```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "FHIR resource validation failed",
    "details": {
      "fields": [
        { "path": "birthDate", "error": "Required field missing" },
        { "path": "name[0].family", "error": "Must be a non-empty string" }
      ]
    }
  },
  "meta": {
    "request_id": "req-uuid-042",
    "duration_ms": 3,
    "node_id": "node-sheffield-01"
  }
}
```

---

## 12. Middleware Configuration

### 12.1 Rate Limiting

Rate limiting is per-device (identified by JWT `sub` claim). Limits are generous because the gateway is local but exist as a safety net.

| Endpoint Category | Rate Limit | Burst |
|-------------------|-----------|-------|
| Read endpoints (GET) | 200 req/min | 50 |
| Write endpoints (POST/PUT) | 60 req/min | 20 |
| Auth endpoints | 10 req/min | 5 |
| Sync trigger | 5 req/min | 2 |
| Anchor trigger | 1 req/min | 1 |

Rate limit headers on every response:

```
X-RateLimit-Limit: 200
X-RateLimit-Remaining: 187
X-RateLimit-Reset: 1740470460
```

When rate limited (429), response includes `Retry-After` header.

### 12.2 RBAC Permission Matrix

| Role | Read | Write | Admin |
|------|------|-------|-------|
| Community Health Worker | patient:read, observation:read | observation:write | — |
| Nurse | patient:read, encounter:read, medication:read | encounter:write, observation:write | — |
| Physician | All clinical reads | All clinical writes | conflict:resolve |
| Site Administrator | All reads | All writes | sync:trigger, anchor:trigger, supply:write |
| Regional Administrator | All reads (cross-site) | All writes (cross-site) | All admin operations |

### 12.3 Request Validation

All POST and PUT request bodies are validated against JSON schemas derived from FHIR R4 StructureDefinitions. Validation occurs in the gateway before any gRPC call to backend services.

### 12.4 Audit Logging

Every authenticated request is logged to `/var/log/open-nucleus/audit.log`:

```json
{
  "request_id": "req-uuid-001",
  "timestamp": "2026-02-25T09:42:00Z",
  "user": "dr-adeleye",
  "method": "POST",
  "endpoint": "/api/v1/patients/patient-uuid-002/encounters",
  "status_code": 201,
  "duration_ms": 42,
  "git_commit": "e4f5a6b7...",
  "ip": "127.0.0.1"
}
```

### 12.5 CORS

Allowed origins: `http://localhost:*` and `http://127.0.0.1:*` only. Extended to local network IP range for distributed deployments.

---

## 13. Real-Time Events (WebSocket)

### 13.1 Connection

```
ws://localhost:8080/api/v1/ws
Authorization: Bearer <jwt_token>
```

### 13.2 Event Types

| Event | Payload |
|-------|---------|
| `sync.started` | `{ peer_id, transport }` |
| `sync.completed` | `{ peer_id, records_in, records_out, conflicts, duration_ms }` |
| `sync.failed` | `{ peer_id, error }` |
| `sync.peer_discovered` | `{ peer_id, peer_name, transport }` |
| `sync.peer_lost` | `{ peer_id }` |
| `conflict.new` | `{ conflict_id, level, resource_type, patient_id }` |
| `conflict.resolved` | `{ conflict_id, resolution, resolved_by }` |
| `alert.new` | `{ alert_id, severity, category, detail }` |
| `anchor.completed` | `{ commit, merkle_root, tangle_message_id }` |
| `anchor.failed` | `{ error }` |
| `supply.stockout_warning` | `{ item_code, site_id, projected_date }` |

### 13.3 Event Message Format

```json
{
  "event": "alert.new",
  "timestamp": "2026-02-25T08:35:00Z",
  "payload": {
    "alert_id": "alert-uuid-001",
    "severity": "high",
    "category": "outbreak",
    "detail": "Potential cholera cluster detected across 3 sites"
  }
}
```

### 13.4 Client Subscription

Clients can filter events by subscribing to specific channels:

```json
{
  "action": "subscribe",
  "channels": ["sync.*", "alert.new", "conflict.new"]
}
```

Wildcard `*` is supported at the category level.

---

## 14. Gateway Configuration

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  max_request_body: 5MB

auth:
  jwt_issuer: open-nucleus-auth
  token_lifetime: 24h
  refresh_window: 2h

grpc:
  patient_service: localhost:50051
  sync_service: localhost:50052
  auth_service: localhost:50053
  formulary_service: localhost:50054
  anchor_service: localhost:50055
  sentinel_agent: localhost:50056
  dial_timeout: 5s
  request_timeout: 30s

rate_limit:
  read_rpm: 200
  write_rpm: 60
  auth_rpm: 10

cors:
  allowed_origins:
    - "http://localhost:*"
    - "http://127.0.0.1:*"

websocket:
  ping_interval: 30s
  max_connections: 10

logging:
  level: info
  audit_file: /var/log/open-nucleus/audit.log
  format: json
```

---

## 15. Git Data Model Reference

For context on how the API maps to storage, every write endpoint commits FHIR resources to a Git repository with this structure:

```
open-nucleus-data/
├── .nucleus/
│   ├── node.json
│   ├── roles/
│   └── formulary/
├── patients/
│   └── {patient-uuid}/
│       ├── Patient.json
│       ├── encounters/{uuid}.json
│       ├── observations/{uuid}.json
│       ├── conditions/{uuid}.json
│       ├── medication-requests/{uuid}.json
│       ├── allergy-intolerances/{uuid}.json
│       └── flags/{uuid}.json
├── supply/
│   ├── inventory/
│   └── deliveries/
└── alerts/
    └── {uuid}.json
```

### Commit Message Convention

```
[RESOURCE_TYPE] [OPERATION] [RESOURCE_ID]

node: {node-uuid}
author: {practitioner-uuid}
site: {site-identifier}
timestamp: {ISO-8601}
fhir_version: R4
```

### Dual-Layer Principle

- **Git** = source of truth (versioned, synced, Merkle-hashed, IOTA-anchored)
- **SQLite** = query index (rebuildable from Git, disposable)
- Every API write commits to Git first, then upserts SQLite
- If SQLite is lost, it is fully reconstructed by walking the Git tree

---

*Open Nucleus • API Gateway Specification V1 • FibrinLab*
*github.com/FibrinLab/open-nucleus*