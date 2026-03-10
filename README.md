# Open Nucleus

Open-source, offline-first electronic health record (EHR) system for military forward operating bases, disaster relief zones, and small clinics in sub-Saharan Africa. Assumes zero connectivity as the default and treats network access as a bonus.

**Architecture:** [`ARCHITECTURE.md`](./ARCHITECTURE.md) | **Security:** [`SECURITY.md`](./SECURITY.md) | **Privacy:** [`PRIVACY.md`](./PRIVACY.md) | **Internals:** [`agents.md`](./agents.md)

## Quick Start

```bash
# Build the monolith (single binary — all Go services in-process)
make build-nucleus

# Run (starts on :8080, HTTPS if TLS configured)
./bin/nucleus --config config.yaml

# Run Sentinel Agent separately (Python, :50056 gRPC + :8090 HTTP)
make run-sentinel

# Test everything
make test-all

# Test Sentinel Agent (Python)
make test-sentinel
```

### Minimal config.yaml

```yaml
data:
  repo_path: data/repo     # Git repository (source of truth)
  db_path: data/nucleus.db  # SQLite search index (rebuildable)

encryption:
  enabled: true
  # master_key_file: /path/to/master.key  # or set NUCLEUS_MASTER_KEY env var

tls:
  mode: auto  # auto-generates self-signed Ed25519 cert; use "provided" for your own
```

## Architecture

```
Flutter App (HTTPS REST/JSON)
        |
        v
  +------------------+
  |    nucleus        |  Single Go binary — all services in-process
  |                   |  HTTP :8080 (TLS optional)
  |  +-----------+    |
  |  | Patient   |    |  FHIR R4 write pipeline, 18 resource types
  |  | Auth      |    |  Ed25519 challenge-response, SMART on FHIR
  |  | Sync      |    |  Git-based sync, FHIR-aware merge driver
  |  | Formulary |    |  WHO essential medicines, drug interactions
  |  | Anchor    |    |  Merkle tree, DID, Verifiable Credentials
  |  +-----------+    |
  +--------+----------+
           |
  +--------+----------+
  |  Git repo         |  Source of truth — encrypted FHIR JSON files
  |  SQLite index     |  Rebuildable search index (no PII in full records)
  +-------------------+

  Sentinel Agent (separate Python process, optional)
  gRPC :50056 / HTTP :8090
```

**Dual-layer data model:** FHIR R4 resources are stored as encrypted JSON files in a Git repository (source of truth) with SQLite as a rebuildable search index. Every clinical write validates, extracts search fields, encrypts, commits to Git, then upserts SQLite with extracted fields only. If SQLite is lost, it rebuilds from Git.

**Per-patient encryption:** Each patient's data is encrypted with a unique AES-256-GCM data encryption key (DEK), wrapped by a master key. Per-provider ECDH key grants allow individual devices to receive their own wrapped DEK copy — revoking a provider's grant immediately cuts off their access without re-encrypting data. Non-patient resources (Practitioner, Organization, etc.) use a system-level key. Destroying a patient's key renders their Git data permanently unreadable — crypto-erasure for privacy law compliance.

**Consent-based access control:** FHIR Consent resources gate patient data access. Each provider device must have an active consent grant before accessing a patient's records. The ConsentCheck middleware enforces this after JWT authentication. Break-glass emergency access creates a time-limited (4h) consent with mandatory audit logging. Consent grants can be exported as W3C Verifiable Credentials for offline verification during sync.

**Blind indexes:** SQLite stores HMAC-SHA256 blind indexes instead of plaintext PII. Patient names are indexed as blind n-gram hashes, enabling substring search without exposing names in the database. Dates are blinded by year-month prefix. An attacker with SQLite access sees only opaque hashes.

**Git-based sync:** Nodes discover each other via Wi-Fi Direct, Bluetooth, or local network and sync using Git fetch/merge/push over ECDH-encrypted channels. A FHIR-aware merge driver classifies conflicts into auto-merge (safe), review (flag for clinician), or block (clinical safety risk).

**Merkle anchoring:** Git Merkle roots are periodically anchored for cryptographic proof of data integrity. V1 uses a stub backend (queued, not submitted); real blockchain integration planned.

## Supported FHIR Resources (18 types)

| Resource | Scope | Indexed |
|----------|-------|---------|
| Patient | — | Yes |
| Encounter, Observation, Condition | Patient-scoped | Yes |
| MedicationRequest, AllergyIntolerance | Patient-scoped | Yes |
| Flag, Immunization, Procedure | Patient-scoped | Yes |
| Consent | Patient-scoped | Yes |
| Practitioner, Organization, Location | Top-level | Yes |
| MeasureReport | Top-level | Yes |
| Provenance | Auto-generated | No |
| DetectedIssue, SupplyDelivery | System-scoped | No |
| StructureDefinition | Read-only | No |

5 custom Open Nucleus FHIR profiles for African healthcare: national IDs, WHO vaccines, growth monitoring, AI provenance, DHIS2 reporting.

## Key Design Decisions

- **Single binary** — All Go services run in one process. No service mesh, no gRPC between components (except optional Sentinel). Runs on a Raspberry Pi 4.
- **Pure Go** — No CGO. SQLite via `modernc.org/sqlite`, Git via `go-git/v5`.
- **Encryption at rest** — Per-patient AES-256-GCM envelope encryption. Master key wraps per-patient DEKs. Per-provider ECDH key grants for access delegation.
- **Consent-gated access** — FHIR Consent resources enforce per-provider, per-patient access control. Break-glass emergency override with audit trail. Offline-verifiable consent via W3C Verifiable Credentials.
- **TLS** — Auto-generated self-signed Ed25519 certs, or bring your own. Never plain HTTP by default.
- **Git as source of truth** — SQLite is a rebuildable index with search fields only (no full FHIR JSON). Full records live in Git, encrypted.
- **Offline-first** — Every feature works without network. Sync is opportunistic.
- **FHIR R4** — Interoperable with global health systems. CapabilityStatement at `/fhir/metadata`.
- **SMART on FHIR** — OAuth2 authorization code + PKCE with SMART v2 scopes. All flows execute locally.
- **Crypto-erasure** — `DELETE /api/v1/patients/{id}/erase` destroys encryption key + purges index. Compliant with GDPR Art 17, SA POPIA, Kenya DPA, Nigeria NDPA.

## Honest Limitations

- **Sentinel is rule-based V1** — Uses WHO IDSR thresholds for outbreak detection. Not AI/LLM-powered. Ollama sidecar is future infrastructure, disabled by default.
- **Anchor backend is stubbed** — Merkle trees are computed and queued but not submitted to any blockchain. Real IOTA Tangle integration is planned.
- **Blind indexes reduce but don't eliminate PII exposure** — Patient names and dates are stored as HMAC blind hashes in SQLite. Gender, clinical codes, and resource IDs remain in plaintext for query performance. Deployed environments should still use disk-level encryption (LUKS, FileVault) for defense in depth.
- **Git commit metadata is unencrypted** — Commit messages and timestamps are visible. File contents are encrypted but paths contain resource type and patient ID. Disk encryption recommended.
- **Dosing engine is stubbed** — Returns `configured=false`. Real WHO dosing guidelines integration is planned.
- **No WebSocket support yet** — Real-time push notifications are planned for Phase 6.

## Testing

```bash
make test-all        # All Go tests with race detection
make test-patient    # Patient service + FHIR + gitstore + sqliteindex
make test-auth       # Auth service + pkg/auth
make test-sync       # Sync service + pkg/merge
make test-formulary  # Formulary service
make test-anchor     # Anchor service + pkg/merge/openanchor
make test-sentinel   # Sentinel agent (Python)
make test-fhir       # FHIR REST API handler tests
make smoke           # Interactive smoke test (27 steps)
```

## Development

```bash
make build-nucleus   # Build monolith binary
make build-all       # Build monolith + legacy microservice binaries
make build-sentinel  # Install Sentinel Agent (Python)
make run-sentinel    # Run Sentinel Agent
make proto-gen       # Regenerate Go protobuf code (requires buf)
make lint            # Run golangci-lint
make clean           # Remove build artifacts
```

## Middleware Pipeline

Every protected request passes through:

| # | Stage | What it does |
|---|-------|--------------|
| 1 | Rate Limiter | Per-device token bucket (200/min reads, 60/min writes, 10/min auth) |
| 2 | Request ID | UUID v4 in `X-Request-ID` header + context |
| 3 | JWT Validator | Ed25519 signature check, expiry, deny list. Skipped for `/auth/*` |
| 4 | Consent Check | Verifies active FHIR Consent grant for patient-scoped routes. Supports `X-Break-Glass: true` for emergencies |
| 5 | RBAC Enforcer | Role permissions against endpoint requirements |
| 6 | Schema Validator | JSON schema validation for POST/PUT FHIR bodies |
| 7 | CORS | Configurable allowed origins |
| 8 | SMART Scope | SMART v2 scope enforcement on FHIR endpoints |
| 9 | Audit Logger | JSON structured log of every request |

## RBAC Roles

| Role | Read | Write | Admin |
|------|------|-------|-------|
| Community Health Worker | patient, observation | observation | — |
| Nurse | patient, encounter, medication | encounter, observation | — |
| Physician | All clinical, consent | All clinical, consent | conflict:resolve |
| Site Administrator | All, consent | All, consent | sync, anchor, supply |
| Regional Administrator | All (cross-site), consent | All (cross-site), consent | All admin |

## License

AGPLv3 — FibrinLab
