# Open Nucleus

Open-source, offline-first electronic health record (EHR) system for military forward operating bases, disaster relief zones, and small clinics in sub-Saharan Africa. Assumes zero connectivity as the default and treats network access as a bonus.

**Architecture:** [`agents.md`](./agents.md)

## Quick Start

```bash
# Build all (gateway + 5 Go services)
make build-all

# Run gateway (starts on :8080)
make run

# Run Sentinel Agent (Python, starts on :50056 gRPC + :8090 HTTP)
make run-sentinel

# Test everything (race detection)
make test-all

# Test Sentinel Agent (Python, 68 tests)
make test-sentinel

# Interactive smoke test (boots all services in-process, 27 REST steps)
make smoke
```

## Architecture

```
Flutter App (HTTP REST/JSON)
        |
        v
   +---------+
   | Gateway  |  Stateless Go HTTP server — REST/JSON to gRPC translation
   +----+-----+
        | gRPC
        v
  +------------------------------------------------------+
  | Auth :50053  | Patient :50051  | Sync :50052          |
  | Formulary :50054 | Anchor :50055 | Sentinel :50056    |
  +------------------------------------------------------+
```

**Dual-layer data model:** FHIR R4 resources stored as JSON files in a Git repository (source of truth) with SQLite as a rebuildable query index. Every clinical write commits to Git first, then upserts SQLite. If SQLite is lost, it rebuilds from Git. Supports 17 resource types (Patient, Encounter, Observation, Condition, MedicationRequest, AllergyIntolerance, Flag, DetectedIssue, Immunization, Procedure, Practitioner, Organization, Location, Provenance, SupplyDelivery, MeasureReport, StructureDefinition) with automatic FHIR Provenance generation on every write. Includes 5 Open Nucleus FHIR profiles for African healthcare deployment (national IDs, WHO vaccines, growth monitoring, AI provenance, DHIS2 reporting).

**Git-based sync:** Nodes discover each other via Wi-Fi Direct, Bluetooth, or local network and sync using Git fetch/merge/push. A FHIR-aware merge driver classifies conflicts into auto-merge (safe), review (flag for clinician), or block (clinical safety risk).

**Merkle anchoring:** Git Merkle roots are periodically anchored for cryptographic proof of data integrity. V1 uses a stub backend (queued, not submitted); real blockchain integration planned.

## Backend Services

| Service | Port | RPCs | Status |
|---------|------|------|--------|
| **Auth** | :50053 | 15 | Ed25519 challenge-response, EdDSA JWT, RBAC (5 roles), device registry |
| **Patient** | :50051 | 49 | FHIR R4 CRUD, 12 resource types, clinical sub-resources, generic top-level CRUD, FTS5 search, patient matching |
| **Sync** | :50052 | ~25 | Transport-agnostic sync, FHIR-aware merge driver, conflict resolution, event bus |
| **Formulary** | :50054 | 16 | WHO essential medicines, drug interactions, allergy cross-reactivity, stock management |
| **Anchor** | :50055 | 14 | Merkle tree, did:key, Verifiable Credentials, queue management |
| **Sentinel** | :50056 | 10 | Python agent — alerts, supply intelligence, Ollama sidecar, HTTP mgmt API (:8090) |

The gateway starts even if backends are down — unavailable services return 503.

## Project Structure

```
cmd/
├── gateway/main.go              Gateway entry point
└── smoke/main.go                Interactive smoke test CLI
internal/
├── config/                      Koanf YAML config loader
├── server/                      HTTP server with graceful shutdown
├── router/                      chi route tree — ~70 REST endpoints + middleware scoping
├── middleware/                   8-stage pipeline (ratelimit, requestid, jwt, rbac, validator, cors, audit)
├── handler/                     HTTP handlers (auth, patient, clinical, resource, sync, conflict, sentinel, formulary, anchor, supply)
│   └── fhir/                   FHIR R4 REST API handlers (read, search, write, $everything, content negotiation)
├── service/                     8 interfaces + gRPC adapters (decouples handlers from transport)
├── grpcclient/                  Connection pool for 6 backend services
└── model/                       Response envelope, error codes, pagination, JWT claims, RBAC
pkg/
├── fhir/                        FHIR R4 utilities (validation, extraction, meta, paths, soft delete, registry, bundle, outcome, capability, provenance)
├── gitstore/                    Git operations via go-git/v5 (pure Go)
├── sqliteindex/                 SQLite query index via modernc.org/sqlite (pure Go, no CGO)
├── auth/                        Ed25519 crypto, EdDSA JWT, nonce store, RBAC, brute-force guard
├── merge/                       FHIR-aware merge driver (3-tier conflict classification)
└── merge/openanchor/            Merkle tree, did:key, Verifiable Credentials, base58btc
services/
├── patient/                     Patient Service (FHIR R4 write pipeline)
├── auth/                        Auth Service (Ed25519 challenge-response)
├── sync/                        Sync Service (transport-agnostic, conflict resolution)
├── formulary/                   Formulary Service (drug DB, interactions, stock)
├── anchor/                      Anchor Service (Merkle anchoring, DID, VCs)
└── sentinel/                    Sentinel Agent (Python — alerts, supply, Ollama sidecar)
proto/                           Protobuf definitions (common, auth, patient, sync, formulary, anchor, sentinel)
schemas/                         8 JSON schemas for FHIR resource validation
```

## Middleware Pipeline

Every protected request passes through these stages in order:

| # | Stage | What it does |
|---|-------|--------------|
| 1 | Rate Limiter | Per-device token bucket (200/min reads, 60/min writes, 10/min auth) |
| 2 | Request ID | UUID v4 in `X-Request-ID` header + context |
| 3 | JWT Validator | Ed25519 signature check, expiry, deny list. Skipped for `/auth/*` |
| 4 | RBAC Enforcer | Role permissions against endpoint requirements |
| 5 | Schema Validator | JSON schema validation for POST/PUT FHIR bodies |
| 6 | CORS | Configurable allowed origins |
| 7 | Audit Logger | JSON structured log of every request |

## RBAC Roles

| Role | Read | Write | Admin |
|------|------|-------|-------|
| Community Health Worker | patient, observation | observation | — |
| Nurse | patient, encounter, medication | encounter, observation | — |
| Physician | All clinical | All clinical | conflict:resolve |
| Site Administrator | All | All | sync, anchor, supply |
| Regional Administrator | All (cross-site) | All (cross-site) | All admin |

## Testing

```bash
make test-all        # All tests with race detection
make test-patient    # Patient service + pkg/fhir + pkg/gitstore + pkg/sqliteindex
make test-auth       # Auth service + pkg/auth
make test-sync       # Sync service + pkg/merge
make test-formulary  # Formulary service
make test-anchor     # Anchor service + pkg/merge/openanchor
make test-sentinel   # Sentinel agent (Python, 68 tests)
make test-fhir       # FHIR REST API handler tests
make test-e2e        # End-to-end tests
make smoke           # Interactive smoke test (27 steps, colored output)
```

## Configuration

All settings in [`config.yaml`](./config.yaml) — server port, gRPC service addresses, rate limits, CORS origins, JWT issuer, timeouts.

## Key Design Decisions

- **Pure Go** — No CGO. Runs on Raspberry Pi 4 and Android tablets.
- **Git as source of truth** — All clinical data in a Git repository. SQLite is a rebuildable index.
- **Offline-first** — Every feature works without network. Sync is opportunistic.
- **FHIR R4** — Interoperable with global health systems. CapabilityStatement at `/fhir/metadata`.
- **Provenance by default** — Every clinical write auto-generates a FHIR Provenance resource (HL7 v3-DataOperation coding, author/custodian agents).
- **No new module deps for anchor crypto** — Merkle trees, did:key, and VCs use only Go stdlib (crypto/ed25519, crypto/sha256).

## Development

```bash
make build           # Build gateway binary
make build-all       # Build all 6 Go binaries
make build-sentinel  # Install Sentinel Agent (Python)
make run             # Build + run gateway
make run-sentinel    # Run Sentinel Agent (Python)
make proto-gen       # Regenerate Go protobuf code (requires buf)
make proto-gen-python # Regenerate Python protobuf code
make lint            # Run golangci-lint
make clean           # Remove build artifacts
```

## License

AGPLv3 — FibrinLab
