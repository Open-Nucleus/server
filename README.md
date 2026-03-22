# Open Nucleus

Open-source, offline-first electronic health record (EHR) system for military forward operating bases, disaster relief zones, and small clinics in sub-Saharan Africa. Assumes zero connectivity as the default and treats network access as a bonus.

**Architecture:** [`ARCHITECTURE.md`](./ARCHITECTURE.md) | **Security:** [`SECURITY.md`](./SECURITY.md) | **Privacy:** [`PRIVACY.md`](./PRIVACY.md) | **Internals:** [`agents.md`](./agents.md)

## Quick Start

```bash
# 1. Seed demo data (6 patients, cholera outbreak scenario)
go run ./cmd/seed

# 2. Start the server (port 8080)
NUCLEUS_BOOTSTRAP_SECRET=demo go run ./cmd/nucleus

# 3. Start the desktop app (in a separate terminal)
cd open-nucleus-app && pnpm install && pnpm tauri dev
```

Login with practitioner ID `demo-clinician` (or any ID — the app auto-registers new devices via bootstrap secret).

### Build from source

```bash
# Build server binary
go build -o nucleus ./cmd/nucleus

# Run with config
NUCLEUS_BOOTSTRAP_SECRET=demo ./nucleus --config config.yaml

# Build desktop app (macOS .app / Windows .msi)
cd open-nucleus-app && pnpm tauri build
```

### Minimal config.yaml

```yaml
data:
  repo_path: data/repo     # Git repository (source of truth)
  db_path: data/nucleus.db  # SQLite search index (rebuildable)

encryption:
  enabled: false  # set true + provide master key for production

tls:
  mode: "off"  # "auto" for self-signed, "provided" for your own

anchor:
  backend: hedera  # "hedera" or "stub"
  network: testnet
  operator_id: "0.0.XXXXX"
  operator_key: ""  # or set NUCLEUS_HEDERA_KEY env var
  topic_id: "0.0.XXXXX"
```

## Ecosystem

Open Nucleus is a 5-repo ecosystem:

| Repo | Language | Purpose |
|------|----------|---------|
| **[server](https://github.com/Open-Nucleus/server)** (this repo) | Go | Core EHR monolith — all services in-process |
| **[open-anchor](https://github.com/Open-Nucleus/open-anchor)** | Go | Blockchain-agnostic data integrity anchoring, DIDs, VCs. Hedera HCS + IOTA backends |
| **[open-sentinel](https://github.com/Open-Nucleus/open-sentinel)** | Python | LLM-powered sleeper agent for clinical surveillance (13 skills, IDSR outbreak detection) |
| **[open-engram](https://github.com/Open-Nucleus/open-engram)** | TypeScript | Brain-inspired memory architecture for AI agents |
| **[open-pharm-dosing](https://github.com/Open-Nucleus/open-pharm-dosing)** | Go | Medication dosing frequency encoding + FHIR R4 Timing conversion |

## Architecture

```
Desktop App (Tauri + React)
        |  REST/JSON on :8080
        v
  +------------------+
  |    nucleus        |  Single Go binary — all services in-process
  |                   |
  |  +-----------+    |
  |  | Patient   |    |  FHIR R4 write pipeline, 18 resource types
  |  | Auth      |    |  Ed25519 challenge-response, SMART on FHIR
  |  | Sync      |    |  Git-based sync, FHIR-aware merge driver
  |  | Formulary |    |  WHO essential medicines, drug interactions
  |  | Anchor    |    |  Hedera HCS anchoring, DIDs, Verifiable Credentials
  |  +-----------+    |
  +--------+----------+
           |
  +--------+----------+
  |  Git repo         |  Source of truth — encrypted FHIR JSON files
  |  SQLite index     |  Rebuildable search index (no PII stored)
  +-------------------+

  Sentinel Agent (separate Python process, optional)
  Rule-based + LLM-powered clinical surveillance
```

**Dual-layer data model:** FHIR R4 resources are stored as encrypted JSON files in a Git repository (source of truth) with SQLite as a rebuildable search index. Every clinical write validates, extracts search fields, encrypts, commits to Git, then upserts SQLite with extracted fields only. If SQLite is lost, it rebuilds from Git.

**Per-patient encryption:** Each patient's data is encrypted with a unique AES-256-GCM data encryption key (DEK), wrapped by a master key. Per-provider ECDH key grants allow individual devices to receive their own wrapped DEK copy — revoking a provider's grant immediately cuts off their access without re-encrypting data. Destroying a patient's key renders their Git data permanently unreadable — crypto-erasure for privacy law compliance.

**Consent-based access control:** FHIR Consent resources gate patient data access. ConsentCheck middleware enforces consent after JWT auth. Break-glass emergency access creates time-limited (4h) consents with audit. Consent grants exportable as W3C Verifiable Credentials for offline verification.

**Blind indexes:** SQLite stores HMAC-SHA256 blind indexes of PII. Patient names indexed as n-gram hashes for substring search without exposing plaintext.

**Hedera HCS anchoring:** Git Merkle roots submitted as messages to Hedera Consensus Service topics. Verification via Mirror Node REST API. `did:hedera` DIDs use HCS topics as append-only document logs.

**Git-based sync:** Nodes sync via Git fetch/merge/push over ECDH-encrypted channels. FHIR-aware merge driver classifies conflicts into auto-merge, review, or block.

## Desktop App (Tauri)

The desktop app lives in `open-nucleus-app/` — built with Tauri v2 + React + TypeScript.

```bash
cd open-nucleus-app
pnpm install          # Install dependencies
pnpm tauri dev        # Development mode (hot reload)
pnpm tauri build      # Production build (macOS .app / Windows .msi)
```

**Tech stack:** Tauri v2, React 19, TypeScript, Vite, Tailwind CSS v4, TanStack Router/Query, Zustand, Recharts, TweetNaCl.js

**Screens:** Login, Dashboard, Patient List/Detail/Form, Formulary, Sync & Conflicts, Alerts, Anchor/Integrity, Settings

## Sentinel Agent

```bash
# Run outbreak detection demo (requires seeded data)
cd ../open-sentinel
python scripts/demo.py --repo /path/to/open-nucleus/data/repo
```

Detects cholera outbreaks using WHO IDSR thresholds. Runs in rule-based mode (no LLM required) or with Ollama for AI-powered analysis.

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

## Testing

```bash
make test-all        # All Go tests with race detection
make test-patient    # Patient service + FHIR + gitstore + sqliteindex
make test-auth       # Auth service + pkg/auth
make test-sync       # Sync service + pkg/merge
make test-formulary  # Formulary service
make test-anchor     # Anchor service + pkg/merge/openanchor
make test-fhir       # FHIR REST API handler tests
make smoke           # Interactive smoke test (27 steps)
```

## Development

```bash
make build-nucleus   # Build monolith binary
go run ./cmd/seed    # Seed demo data (6 patients + cholera outbreak)
go run ./cmd/nucleus # Run server directly
make proto-gen       # Regenerate protobuf code (requires buf)
make lint            # Run golangci-lint
make clean           # Remove build artifacts
```

## Middleware Pipeline

| # | Stage | What it does |
|---|-------|--------------|
| 1 | Rate Limiter | Per-device token bucket (200/min reads, 60/min writes, 10/min auth) |
| 2 | Request ID | UUID v4 in `X-Request-ID` header + context |
| 3 | JWT Validator | Ed25519 signature check, expiry, deny list. Skipped for `/auth/*` |
| 4 | Consent Check | Verifies active FHIR Consent grant for patient-scoped routes. `X-Break-Glass: true` for emergencies |
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

## Honest Limitations

- **Sentinel has dual modes** — Rule-based (WHO IDSR thresholds) works offline. LLM mode (Ollama) adds AI reasoning but requires 8GB+ RAM.
- **Hedera anchoring requires testnet account** — Free to create at [portal.hedera.com](https://portal.hedera.com). Mainnet costs real HBAR.
- **Blind indexes reduce but don't eliminate PII exposure** — Gender, clinical codes, and resource IDs remain in plaintext for query performance. Use disk encryption for defense in depth.
- **Git commit metadata is unencrypted** — Commit messages and file paths visible. Disk encryption recommended.
- **Dosing engine is stubbed** — Returns `configured=false`. Real WHO dosing guidelines integration planned.

## License

AGPLv3 — FibrinLab
