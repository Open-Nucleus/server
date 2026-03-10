# Architecture

Technical architecture of Open Nucleus, an offline-first EHR system.

## System Overview

```
+-------------------------------------------------------------------+
|  nucleus (single Go binary)                                        |
|                                                                    |
|  HTTP :8080 (TLS)                                                 |
|  ┌─────────────────────────────────────┐                          |
|  │ Middleware Pipeline                  │                          |
|  │ Rate Limit → RequestID → JWT →      │                          |
|  │ RBAC → Schema → CORS → SMART →     │                          |
|  │ Audit                               │                          |
|  └──────────┬──────────────────────────┘                          |
|             │                                                      |
|  ┌──────────▼──────────────────────────┐                          |
|  │ Handlers (HTTP → Service Interface) │                          |
|  │ auth, patient, clinical, resource,  │                          |
|  │ sync, conflict, formulary, anchor,  │                          |
|  │ supply, sentinel, FHIR, SMART       │                          |
|  └──────────┬──────────────────────────┘                          |
|             │                                                      |
|  ┌──────────▼──────────────────────────┐                          |
|  │ Local Adapters (in-process calls)   │                          |
|  │ internal/service/local/*            │                          |
|  └──────────┬──────────────────────────┘                          |
|             │                                                      |
|  ┌──────────▼──────────────────────────────────────────┐          |
|  │ Business Logic                                       │          |
|  │  Patient:   pipeline.Writer (validate→encrypt→git)  │          |
|  │  Auth:      authservice (Ed25519, JWT, devices)     │          |
|  │  Sync:      syncservice (merge, conflicts, peers)   │          |
|  │  Formulary: formularyservice (drugs, interactions)  │          |
|  │  Anchor:    anchorservice (Merkle, DID, VCs)        │          |
|  └──────────┬──────────────────────────┬───────────────┘          |
|             │                          │                           |
|  ┌──────────▼────────┐  ┌─────────────▼──────────┐               |
|  │ Git Repository     │  │ SQLite (unified DB)     │               |
|  │ Encrypted FHIR JSON│  │ Search index + auth +   │               |
|  │ Wrapped DEKs       │  │ sync + formulary +      │               |
|  │ Device registry    │  │ anchor tables           │               |
|  └────────────────────┘  └────────────────────────┘               |
+-------------------------------------------------------------------+

  Sentinel Agent (optional, separate Python process)
  gRPC :50056 / HTTP :8090
```

## Write Pipeline

The core data path for all clinical writes:

```
HTTP Request (FHIR JSON)
    │
    ▼
1. Validate (pkg/fhir/validate.go)
    │  Structural validation per resource type
    │  Required fields, enums, references
    │
    ▼
2. Meta (pkg/fhir/meta.go)
    │  Assign UUID if absent
    │  Set meta.lastUpdated, versionId, source
    │
    ▼
3. Extract Search Fields (pkg/fhir/extract.go)
    │  Pull indexed fields from cleartext JSON
    │  (name, DOB, codes, dates, statuses)
    │
    ▼
4. Encrypt (pkg/envelope/)
    │  Get-or-create patient DEK
    │  AES-256-GCM encrypt full FHIR JSON
    │  Non-patient resources use system DEK
    │
    ▼
5. Git Commit (pkg/gitstore/)
    │  Write ciphertext to patients/{pid}/{type}/{id}.json
    │  Atomic commit with structured message
    │
    ▼
6. SQLite Upsert (pkg/sqliteindex/)
    │  Upsert extracted search fields (no full JSON)
    │
    ▼
7. Provenance (pkg/fhir/provenance.go)
    │  Auto-generate FHIR Provenance resource
    │  HL7 v3-DataOperation coding
    │
    ▼
Return: resource + git commit metadata
```

Steps 5-6 are serialized by a `sync.Mutex` to prevent concurrent Git writes. Steps 1-4 happen outside the lock.

## Read Path

```
HTTP Request (GET /api/v1/patients/{pid}/encounters/{eid})
    │
    ▼
1. SQLite Lookup
    │  Find resource by ID, get patient_id
    │  (Returns 404 if not indexed)
    │
    ▼
2. Git Read (pkg/gitstore/)
    │  Read ciphertext from Git path
    │
    ▼
3. Decrypt (pkg/envelope/)
    │  Unwrap patient DEK with master key
    │  AES-256-GCM decrypt
    │  (Falls back to cleartext for unencrypted repos)
    │
    ▼
Return: FHIR JSON
```

List endpoints return search fields from SQLite. Detail endpoints read full resources from Git and decrypt.

## Encryption Key Hierarchy

```
NUCLEUS_MASTER_KEY (AES-256, 32 bytes)
│
│  Loaded from env var or config file
│  Never written to disk by the application
│
├── AES-KW wrap ──► .nucleus/keys/{patient_id}.key
│                    │
│                    └── AES-256-GCM ──► patients/{pid}/**/*.json
│
├── AES-KW wrap ──► .nucleus/keys/_system.key
│                    │
│                    └── AES-256-GCM ──► practitioners/*.json
│                                        organizations/*.json
│                                        locations/*.json
│                                        measure_reports/*.json
│
└── Crypto-Erasure:
     DELETE .nucleus/keys/{patient_id}.key
     → All patient files become permanently unreadable
     → DELETE FROM sqlite WHERE patient_id = ?
```

## Monolith Design

### Why a single binary?

The original microservice architecture (5 Go services + gateway communicating via gRPC) was over-engineered for the target hardware (Raspberry Pi 4). The monolith:

- Eliminates 5 gRPC round-trips per request
- Uses one SQLite database instead of 5
- Simplifies deployment to a single binary
- Reduces memory footprint by ~60%
- Shares the Git repository lock directly

### Interface Preservation

The handler layer is identical between the old gateway and the new monolith. The key abstraction:

```
OLD:  Handler → service.Interface → gRPC Adapter → wire → Microservice
NEW:  Handler → service.Interface → Local Adapter → in-process call
```

`internal/service/interfaces.go` defines 9 interfaces. `internal/service/local/` implements them by constructing business logic directly instead of making gRPC calls.

### Sentinel Exception

The Sentinel Agent remains a separate Python process because it requires Python ML libraries. The monolith uses stub services by default and connects via gRPC when Sentinel is running.

## Sync Architecture

### Node-to-Node

```
Node A                          Node B
  │                                │
  │  1. Discover (Wi-Fi/BT/LAN)   │
  │◄──────────────────────────────►│
  │                                │
  │  2. ECDH Key Exchange          │
  │  (X25519 from Ed25519 keys)   │
  │◄──────────────────────────────►│
  │                                │
  │  3. Git fetch/merge/push       │
  │  (encrypted channel)           │
  │──────────────────────────────►│
  │                                │
  │  4. Merge Driver               │
  │  classify conflicts:           │
  │  - auto-merge (timestamps)     │
  │  - review (flag for clinician) │
  │  - block (safety risk)         │
  │                                │
  │  5. Post-sync: re-encrypt      │
  │  locally, update SQLite index  │
  │                                │
```

Sync transports exchange cleartext FHIR over the encrypted channel. The receiving node encrypts with its own keys before local Git commit.

### Conflict Classification

The FHIR-aware merge driver (`pkg/merge/`) classifies conflicts:

| Tier | Action | Example |
|------|--------|---------|
| Auto-merge | Apply latest timestamp | Demographics update, status change |
| Review | Flag for clinician | Conflicting diagnoses, medication changes |
| Block | Reject until resolved | Conflicting allergy records, safety-critical data |

## Data Storage Layout

### Git Repository

```
data/repo/
├── .nucleus/
│   └── keys/
│       ├── {patient_id}.key     # Wrapped per-patient DEK
│       └── _system.key          # Wrapped system DEK
├── patients/
│   └── {patient_id}/
│       ├── patient.json         # Encrypted Patient resource
│       ├── encounters/
│       │   └── {id}.json        # Encrypted Encounter
│       ├── observations/
│       │   └── {id}.json
│       ├── conditions/
│       │   └── {id}.json
│       ├── medication_requests/
│       │   └── {id}.json
│       ├── allergy_intolerances/
│       │   └── {id}.json
│       ├── immunizations/
│       │   └── {id}.json
│       ├── procedures/
│       │   └── {id}.json
│       ├── flags/
│       │   └── {id}.json
│       └── provenance/
│           └── {id}.json
├── practitioners/
│   └── {id}.json                # Encrypted with system key
├── organizations/
│   └── {id}.json
├── locations/
│   └── {id}.json
└── measure_reports/
    └── {id}.json
```

### SQLite Database

Single unified database at `data/nucleus.db` with tables for:

- **Patient index:** patients, encounters, observations, conditions, medication_requests, allergy_intolerances, immunizations, procedures, flags, patient_summaries, detected_issues, practitioners, organizations, locations, measure_reports
- **Auth:** deny_list, revocations
- **Sync:** conflicts, sync_history, peers
- **Formulary:** stock_levels
- **Anchor:** anchor_queue

FTS5 virtual table for patient full-text search. WAL journal mode, single-writer.

## Project Structure

```
cmd/
├── nucleus/main.go              Monolith entry point (recommended)
├── gateway/main.go              Legacy gateway (gRPC to microservices)
└── smoke/main.go                Interactive smoke test CLI
internal/
├── config/                      Koanf YAML config loader
├── server/                      HTTP server with graceful shutdown + TLS
├── router/                      chi route tree — ~95 REST + ~85 FHIR + 11 SMART endpoints
├── middleware/                   8-stage pipeline
├── handler/                     HTTP handlers
│   └── fhir/                    FHIR R4 REST API handlers
├── service/
│   ├── interfaces.go            9 service interfaces + DTOs
│   ├── local/                   In-process adapters (monolith)
│   ├── patient.go, auth.go ...  gRPC adapters (legacy gateway)
│   └── ...
├── grpcclient/                  gRPC pool (legacy gateway only)
└── model/                       Response envelope, errors, pagination, JWT claims, RBAC
pkg/
├── envelope/                    Per-patient AES-256-GCM encryption
├── tls/                         TLS certificate management
├── fhir/                        FHIR R4 utilities (17 resource types, 5 profiles)
├── gitstore/                    Git operations (go-git/v5)
├── sqliteindex/                 SQLite search index (modernc.org/sqlite)
├── auth/                        Ed25519 crypto, EdDSA JWT, RBAC
├── smart/                       SMART on FHIR v2
├── merge/                       FHIR-aware merge driver
└── merge/openanchor/            Merkle tree, DID, Verifiable Credentials
services/
├── patient/                     Patient Service (write pipeline)
├── auth/                        Auth Service (Ed25519, JWT)
├── sync/                        Sync Service (merge, conflicts)
├── formulary/                   Formulary Service (drugs, interactions)
├── anchor/                      Anchor Service (Merkle, DID)
└── sentinel/                    Sentinel Agent (Python)
```
