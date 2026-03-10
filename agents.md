# Open Nucleus — Architectural Memory

> Living document. Updated after every major feature or structural change.
> Last updated: Overhaul Phases 0–2 — Monolith, Encryption, Crypto-Erasure (2026-03-10)

---

## System Overview

### Open Nucleus
Open Nucleus is an open-source, offline-first electronic health record (EHR) system designed for military forward operating bases, disaster relief zones, and small clinics in sub-Saharan Africa. It assumes zero connectivity as the default and treats network access as a bonus.

### Core Architecture
**Single Go binary** (`cmd/nucleus/main.go`) with all services running in-process. The Python Sentinel Agent runs as a separate process on :50056 (gRPC) / :8090 (HTTP). The Flutter frontend lives in a separate repo (open-nucleus-app) and consumes the HTTP API as a pure REST client.

**Dual-layer data model:** FHIR R4 resources are stored as **encrypted** JSON files in a Git repository (source of truth) with SQLite as a rebuildable search index containing extracted fields only (no full FHIR JSON). Every clinical write validates, extracts search fields, encrypts, commits to Git, then upserts SQLite.

**Per-patient encryption:** AES-256-GCM envelope encryption with master key wrapping per-patient DEKs. Destroying a patient's key renders their Git data permanently unreadable (crypto-erasure).

**Git-based sync:** Nodes sync using Git fetch/merge/push over ECDH-encrypted channels. A FHIR-aware merge driver classifies conflicts into auto-merge (safe), review (flag for clinician), or block (clinical safety risk).

**Sentinel Agent:** Rule-based V1 using WHO IDSR thresholds for outbreak detection. Not AI/LLM-powered. Ollama sidecar is future infrastructure.

**Merkle anchoring:** Git Merkle roots queued for anchoring. V1 uses a stub backend; real blockchain integration planned.

```
Flutter App (HTTPS REST/JSON)
        │
        ▼
  ┌─────────────────────────────┐
  │  nucleus (single binary)     │  HTTP :8080 (TLS)
  │  Patient, Auth, Sync,       │
  │  Formulary, Anchor          │
  │  ┌─────────┐  ┌──────────┐ │
  │  │ Git repo │  │ SQLite   │ │
  │  │(encrypted│  │(index    │ │
  │  │ FHIR)   │  │ only)    │ │
  │  └─────────┘  └──────────┘ │
  └─────────────────────────────┘
        │ gRPC (optional)
        ▼
  ┌─────────────────────┐
  │ Sentinel :50056     │  Python (separate process)
  └─────────────────────┘
```

---

## Dependency Wiring (main.go)

`cmd/nucleus/main.go` is the composition root. It constructs all services in-process:

```
config.Load(path)
    │
    ▼
gitstore.NewStore(cfg.Data.RepoPath)   ← shared Git repository
sql.Open("sqlite", cfg.Data.DBPath)    ← shared unified SQLite DB
sqliteindex.InitUnifiedSchema(db)      ← all tables in one schema
    │
    ├─► pipeline.NewWriter(git, idx)   ← FHIR write pipeline (validate→encrypt→git→sqlite)
    │       ▼
    │   local.NewPatientService(pw, idx, git)
    │       ▼
    │   handler.NewPatientHandler(patientSvc)
    │
    ├─► authservice.NewAuthService(cfg, git, keystore, denyList)
    │       ▼
    │   local.NewAuthService(authImpl)
    │       ▼
    │   handler.NewAuthHandler(authSvc)
    │
    ├─► authservice.NewSmartService(authImpl, clientStore)
    │       ▼
    │   local.NewSmartService(smartImpl)
    │       ▼
    │   handler.NewSmartHandler(smartSvc)
    │
    ├─► syncservice.NewSyncEngine(cfg, git, conflicts, history, peers, mergeDriver, eventBus)
    │       ▼
    │   local.NewSyncService(syncEngine, historyStore, peerStore)
    │   local.NewConflictService(conflictStore, eventBus)
    │       ▼
    │   handler.NewSyncHandler(syncSvc) + handler.NewConflictHandler(conflictSvc)
    │
    ├─► formularyservice.New(drugDB, interactions, stockStore, dosingEngine)
    │       ▼
    │   local.NewFormularyService(formularyImpl)
    │       ▼
    │   handler.NewFormularyHandler(formularySvc)
    │
    ├─► anchorservice.New(git, backend, identity, queue, store, creds, dids, nodeKey)
    │       ▼
    │   local.NewAnchorService(anchorImpl)
    │       ▼
    │   handler.NewAnchorHandler(anchorSvc)
    │
    ├─► local.NewStubSentinelService()   ← stubs when Sentinel not running
    │   local.NewStubSupplyService()
    │
    ├─► middleware.NewSchemaValidator() + load 8 JSON schemas
    ├─► middleware.NewJWTAuth(pubKey, issuer)
    ├─► middleware.NewRateLimiter(cfg.RateLimit)
    │
    ▼
router.New(Config{all handlers, middleware, schemaValidator, auditLogger, corsOrigins})
    │
    ▼
server.New(cfg, mux, logger).WithTLS(tlsCfg).Run()
```

---

## Package Dependency Graph

Arrows mean "imports / depends on". No circular dependencies exist.

```
cmd/nucleus/main (monolith)
    ├── internal/config
    ├── internal/service/local   ── services/*/   (direct business logic)
    │                            ── pkg/envelope   (encryption)
    ├── internal/handler         ── internal/service (interfaces only)
    │                            ── internal/model
    ├── internal/middleware       ── internal/config  (ratelimit only)
    │                            ── internal/model    (all middleware)
    ├── internal/router          ── internal/handler
    │                            ── internal/middleware
    │                            ── internal/model
    ├── internal/server          ── internal/config
    │                            ── pkg/tls
    ├── pkg/gitstore             ── (go-git/v5)
    ├── pkg/sqliteindex          ── pkg/fhir
    └── pkg/envelope             ── (crypto/aes, crypto/cipher)

cmd/gateway/main (legacy, still builds)
    ├── internal/grpcclient      ── internal/config
    └── internal/service         ── internal/grpcclient (gRPC adapters)
```

**internal/model** is the leaf package — imported by nearly everything, imports nothing internal.

---

## Module Details

### internal/config
- **config.go** — `Config` struct matching `config.yaml` / spec section 14. Loaded via koanf.
- Consumed by: main (passed to pool, server, rate limiter), grpcclient (dial addresses/timeouts), server (port, timeouts).

### internal/model (leaf — no internal imports)
- **envelope.go** — `Envelope` struct + `JSON()`, `Success()`, `ErrorResponse()` response writers. Every HTTP response flows through here.
- **errors.go** — 16 error code constants (`ErrAuthRequired`, `ErrRateLimited`, etc.) + `ErrorHTTPStatus` map + `WriteError()` + `NotImplementedError()`.
- **pagination.go** — `Pagination` struct, `PaginationFromRequest(r)` query parser, `NewPagination()` constructor.
- **auth.go** — `NucleusClaims` (JWT claims struct), `LoginRequest`, `RefreshRequest`, `LogoutRequest`.
- **rbac.go** — 5 role constants, 24 permission constants, `RolePermissions` matrix map, `HasPermission(role, perm)`.
- **context.go** — Context keys (`CtxRequestID`, `CtxClaims`) + extraction helpers `RequestIDFromContext()`, `ClaimsFromContext()`. This is the glue that lets middleware pass data to handlers without coupling.

### internal/middleware

Each middleware is a `func(http.Handler) http.Handler` or a method that returns one. They compose via chi's `r.Use()` and `r.With()`.

| File | What it writes to context | What it reads from context | External deps |
|------|---------------------------|----------------------------|---------------|
| **requestid.go** | `CtxRequestID` (UUID v4) | — | `github.com/google/uuid` |
| **jwtauth.go** | `CtxClaims` (*NucleusClaims) | — | `github.com/golang-jwt/jwt/v5` |
| **rbac.go** | — | `CtxClaims` (reads role + permissions) | — |
| **ratelimit.go** | — | `CtxClaims` (reads Subject for device ID) | `golang.org/x/time/rate` |
| **validator.go** | — | — (reads r.Body) | `github.com/santhosh-tekuri/jsonschema/v5` |
| **cors.go** | — | — (reads Origin header) | — |
| **audit.go** | — | `CtxRequestID`, `CtxClaims` | `log/slog` |
| **smartscope.go** | — | `CtxClaims` (reads Scope, LaunchPatient) | `pkg/smart` |

**Context data flow:**
```
requestid.go  ──writes──►  CtxRequestID  ──read by──►  audit.go, handlers (via Meta)
jwtauth.go    ──writes──►  CtxClaims     ──read by──►  rbac.go, ratelimit.go, audit.go, handlers
```

**Middleware pipeline order on protected routes:**
```
CORS → RequestID → AuditLog → JWTAuth → [per-route: RateLimiter → RequirePermission → SchemaValidator] → Handler
```

**Auth routes skip** JWTAuth and RBAC — they only get CORS + RequestID + AuditLog + RateLimiter(CategoryAuth).

### internal/grpcclient
- **pool.go** — `Pool` holds a `map[string]*grpc.ClientConn` for 6 named services. `NewPool()` dials all with timeout (non-blocking on failure — stores nil, returns SERVICE_UNAVAILABLE at call time). `Conn(name)` returns connection or error.
- Consumed by: service adapters call `pool.Conn("auth")`, `pool.Conn("patient")`, etc.

### internal/service
- **interfaces.go** — 9 service interfaces (`AuthService`, `PatientService`, `SyncService`, `ConflictService`, `SentinelService`, `FormularyService`, `AnchorService`, `SupplyService`, `SmartService`) + all DTOs including `EraseResponse`. Handlers depend only on these interfaces.

#### internal/service/local/ (monolith — recommended)
In-process adapters that call business logic directly without gRPC:
- **patient.go** — `patientService` wraps `pipeline.Writer` + `sqliteindex.Index` + `gitstore.Store`. Reads FHIR JSON from Git and decrypts via `pw.DecryptFromGit()`. Implements `ErasePatient()` for crypto-erasure.
- **auth.go** — `authService` wraps `authservice.AuthService` directly.
- **smart.go** — `smartService` wraps `authservice.SmartService` directly.
- **sync.go** — `syncService` wraps `syncservice.SyncEngine` + history/peer stores. Also `conflictService` wraps conflict store + event bus.
- **formulary.go** — `formularyService` wraps `formularyservice.FormularyService` directly.
- **anchor.go** — `anchorService` wraps `anchorservice.AnchorService` directly.
- **stubs.go** — `stubSentinelService` + `stubSupplyService` return 503 when Sentinel is not running.

#### internal/service/*.go (legacy gRPC adapters — still builds)
- **auth.go**, **patient.go**, **sync.go**, etc. — gRPC adapters via `pool.Conn(name)`. Used by `cmd/gateway/main.go` when running in distributed microservice mode.

**Key pattern:** Handlers never touch business logic or gRPC directly. The service interface layer enables both monolith (local adapters) and distributed (gRPC adapters) deployment from the same handler code.

### internal/handler
- **auth.go** — `AuthHandler` holds `service.AuthService`. Methods: `Login`, `Refresh`, `Logout`, `Whoami`. Whoami short-circuits from JWT claims in context if available.
- **patient.go** — `PatientHandler` holds `service.PatientService`. Methods: `List`, `GetByID`, `Search`, `Create`, `Update`, `Delete`, `History`, `Timeline`, `Match`. Write methods use `writeResponseWithGit()` to include git metadata in the response envelope.
- **clinical.go** — Additional methods on `PatientHandler` for all 22 clinical sub-resource endpoints: `ListEncounters`, `GetEncounter`, `CreateEncounter`, `UpdateEncounter`, `ListObservations`, `GetObservation`, `CreateObservation`, `ListConditions`, `CreateCondition`, `UpdateCondition`, `ListMedicationRequests`, `CreateMedicationRequest`, `UpdateMedicationRequest`, `ListAllergyIntolerances`, `CreateAllergyIntolerance`, `UpdateAllergyIntolerance`, `ListImmunizations`, `GetImmunization`, `CreateImmunization`, `ListProcedures`, `GetProcedure`, `CreateProcedure`.
- **resource.go** — `ResourceHandler` with factory methods (`ListFactory`, `GetFactory`, `CreateFactory`, `UpdateFactory`) for top-level CRUD (Practitioner, Organization, Location). `CapabilityStatementHandler()` serves FHIR R4 CapabilityStatement at `/fhir/metadata`.
- **sync.go** — `SyncHandler` holds `service.SyncService`. Methods: `Status`, `Peers`, `Trigger`, `History`, `ExportBundle`, `ImportBundle`.
- **conflict.go** — `ConflictHandler` holds `service.ConflictService`. Methods: `List`, `GetByID`, `Resolve`, `Defer`.
- **sentinel.go** — `SentinelHandler` holds `service.SentinelService`. Methods: `ListAlerts`, `Summary`, `GetAlert`, `Acknowledge`, `Dismiss`.
- **formulary.go** — `FormularyHandler` holds `service.FormularyService`. 16 methods: `SearchMedications`, `GetMedication`, `ListMedicationsByCategory`, `CheckInteractions`, `CheckAllergyConflicts`, `ValidateDosing`, `GetDosingOptions`, `GenerateSchedule`, `GetStockLevel`, `UpdateStockLevel`, `RecordDelivery`, `GetStockPrediction`, `GetRedistributionSuggestions`, `GetFormularyInfo`.
- **anchor.go** — `AnchorHandler` holds `service.AnchorService`. 13 methods: `Status`, `Verify`, `History`, `Trigger`, `NodeDID`, `DeviceDID`, `ResolveDID`, `IssueCredential`, `VerifyCredentialHandler`, `ListCredentials`, `ListBackends`, `BackendStatus`, `QueueStatus`.
- **supply.go** — `SupplyHandler` holds `service.SupplyService`. Methods: `Inventory`, `InventoryItem`, `RecordDelivery`, `Predictions`, `Redistribution`.
- **stubs.go** — `StubHandler()` returns 501 via `model.NotImplementedError()`. Only used for WebSocket endpoint (Phase 5).

### internal/router
- **router.go** — `New(Config)` builds the chi route tree. Config now includes all 8 handler types + `SchemaValidator`. `validatorMiddleware()` helper returns a no-op if SchemaValidator is nil (for tests without schemas). Owns middleware scoping:
  - `/health` — no middleware beyond global
  - `/api/v1/auth/*` — global + RateLimiter(CategoryAuth), NO JWT/RBAC
  - `/api/v1/*` (everything else) — global + JWTAuth, then per-route RateLimiter + RequirePermission + optional SchemaValidator
  - `/fhir/metadata` — no auth, serves FHIR CapabilityStatement
  - `/api/v1/patients/{id}/immunizations`, `/api/v1/patients/{id}/procedures` — patient-scoped clinical
  - `/api/v1/practitioners`, `/api/v1/organizations`, `/api/v1/locations` — top-level FHIR resources
- ~70 REST endpoints wired to real handlers. Only `/ws` remains stubbed.

### internal/server
- **server.go** — `Server` wraps `http.Server` with config-driven timeouts. `Run()` starts listener and blocks until SIGINT/SIGTERM, then calls `Shutdown()` with 10s grace period.

### schemas/
All 8 schemas use inline `$defs` for reusable `Reference` (`{ reference: string minLength:1 }`) and `CodeableConcept` (`anyOf: [ has coding[], has text ]`) patterns. They mirror the validation rules in `pkg/fhir/validate.go` so malformed payloads are rejected at the gateway before the gRPC round-trip.

- **patient.json** — Requires `resourceType: "Patient"`, `name` array (items: `{ family: string, given: string[] }`), `gender` enum, `birthDate` string.
- **encounter.json** — Requires `resourceType: "Encounter"`, `status` enum (8 FHIR values), `class` object with `code`, `subject` Reference, `period` with `start`.
- **observation.json** — Requires `resourceType: "Observation"`, `status` enum (7 values), `code` CodeableConcept, `subject` Reference, `effectiveDateTime`.
- **condition.json** — Requires `resourceType: "Condition"`, `clinicalStatus` CodeableConcept, `verificationStatus` CodeableConcept, `code` CodeableConcept, `subject` Reference.
- **medication_request.json** — Requires `resourceType: "MedicationRequest"`, `status`, `intent`, `medicationCodeableConcept` CodeableConcept, `subject` Reference, `dosageInstruction` array (minItems:1).
- **allergy_intolerance.json** — Requires `resourceType: "AllergyIntolerance"`, `clinicalStatus` CodeableConcept, `verificationStatus` CodeableConcept, `code` CodeableConcept, `patient` Reference.
- **immunization.json** — Requires `resourceType: "Immunization"`, `status` enum (3 values), `vaccineCode` CodeableConcept, `patient` Reference, `occurrenceDateTime`.
- **procedure.json** — Requires `resourceType: "Procedure"`, `status` enum (8 values), `code` CodeableConcept, `subject` Reference.

---

## Proto Structure

```
proto/
├── common/v1/
│   ├── metadata.proto   ← GitMetadata (+ Timestamp), PaginationRequest/Response, NodeInfo
│   └── fhir.proto       ← FHIRResource{resource_type, id, json_payload bytes}
├── auth/v1/
│   └── auth.proto       ← AuthService: 15 RPCs (register, challenge, authenticate, refresh, logout, identity, devices, roles, validate, health)
├── patient/v1/
│   └── patient.proto    ← PatientService: 49 RPCs (CRUD + clinical + immunization + procedure + generic CRUD + batch + index + health)
├── sync/v1/
│   └── sync.proto       ← SyncService (14 RPCs) + ConflictService (4 RPCs) + NodeSyncService (3 RPCs)
├── formulary/v1/
│   └── formulary.proto  ← FormularyService: 16 RPCs (drug lookup, interactions, allergy, dosing stub, stock, redistribution, info, health)
├── anchor/v1/
│   └── anchor.proto     ← AnchorService: 14 RPCs (anchoring, DID, credentials, backend, health)
└── sentinel/v1/
    └── sentinel.proto   ← SentinelService: 5 alert RPCs + 5 supply chain RPCs
```

FHIR resources are opaque `bytes json_payload` — the gateway never parses or transforms them.

Generated Go code lives in `gen/proto/` (protoc with go + go-grpc plugins).

---

## Shared Libraries (pkg/)

### pkg/envelope — Per-Patient Encryption
AES-256-GCM envelope encryption with master key wrapping.
- **envelope.go** — `KeyManager` interface: `GetOrCreateKey()`, `DestroyKey()`, `Encrypt()`, `Decrypt()`, `IsKeyDestroyed()`. `FileKeyManager` impl stores wrapped DEKs in Git at `.nucleus/keys/`. In-memory cache with `sync.RWMutex`.

### pkg/tls — TLS Certificate Management
Auto-generate or load TLS certificates.
- **certs.go** — `Config{Mode, CertFile, KeyFile, CertDir}`. `LoadOrGenerate()` returns `*tls.Config`. Modes: "auto" (self-signed Ed25519), "provided" (user PEM), "off" (nil).

### pkg/fhir — FHIR R4 Utilities
Pure functions for working with FHIR resources. No I/O.
- **types.go** — Resource type constants for 17 types, operation constants (`OpCreate`, etc.), row structs for 14 indexed types. **No `FHIRJson` field** — row structs contain only extracted search fields.
- **path.go** — `GitPath(resourceType, patientID, resourceID)` returns Git file path. Patient-scoped: `patients/{pid}/immunizations/{id}.json`, etc. Top-level: `practitioners/{id}.json`.
- **meta.go** — `SetMeta()` writes `meta.lastUpdated/versionId/source`. `AssignID()` assigns UUID if absent.
- **validate.go** — `Validate(resourceType, json)` structural validation for 12 resource types.
- **extract.go** — Extract functions for all 14 indexed types. Returns row structs with search fields only (no full FHIR JSON).
- **softdelete.go** — `ApplySoftDelete()` for all types.
- **registry.go** — Central resource registry: 17 resource types with scope, interactions, search params.
- **outcome.go** — FHIR R4 OperationOutcome builder.
- **bundle.go** — FHIR R4 Bundle builder.
- **capability.go** — Auto-generates CapabilityStatement from registry.
- **provenance.go** — Auto-generates Provenance with HL7 v3-DataOperation coding.

### pkg/gitstore — Git Operations
Wraps `go-git/v5` for clinical data Git repository management.
- **store.go** — `Store` interface: `WriteAndCommit()`, `Read()`, `LogPath()`, `Head()`, `TreeWalk()`, `Rollback()`. `NewStore(repoPath)` opens or inits repo.
- **commit.go** — `CommitMessage` struct with `Format()` and `ParseCommitMessage()` for structured commit messages per spec §3.3.

### pkg/sqliteindex — SQLite Search Index
Uses `modernc.org/sqlite` (pure Go, no CGO) for Raspberry Pi 4 deployment. **Pure search index** — no full FHIR JSON stored. All `fhir_json` columns have been removed.
- **schema.go** — `InitSchema()` creates 14 resource tables + index_meta + FTS5 + triggers. `InitUnifiedSchema()` additionally creates auth (deny_list, revocations), sync (conflicts, sync_history, peers), formulary (stock_levels), and anchor (anchor_queue) tables. `DropAll()` for rebuild.
- **index.go** — `Index` interface: Upsert/Get/List methods for all 14 resource types + `DeletePatientData()` for crypto-erasure + bundle + search + timeline + match + meta + summary. `NewIndex(dbPath)` opens DB with WAL mode. `NewIndexFromDB(*sql.DB)` for shared DB in monolith.
- **erase.go** — `DeletePatientData(patientID)` deletes from 10 tables in a transaction for crypto-erasure.
- **search.go** — FTS5 patient search via `patients_fts` virtual table.
- **timeline.go** — `GetTimeline()` UNION ALL query across encounters, observations, conditions, flags.
- **match.go** — `GetMatchCandidates()` broad SQL query for patient identity matching.
- **summary.go** — `UpdateSummary()` recomputes `patient_summaries` counts. `GetPatientBundle()` returns patient + all active child resources.

## Patient Service (services/patient/)

The clinical data write pipeline. Single writer for all FHIR data: validate → extract search fields → encrypt → Git commit → SQLite upsert (fields only) → return resource + commit metadata. Supports optional per-patient envelope encryption via `WithEncryption(keys)` and crypto-erasure via `DestroyPatientKey(patientID)`.

```
services/patient/
├── cmd/main.go                          ← gRPC server entrypoint, port :50051
├── config.yaml                          ← default config
├── internal/
│   ├── config/config.go                 ← koanf config loader
│   ├── pipeline/writer.go               ← Write pipeline (sync.Mutex serialized)
│   └── server/
│       ├── server.go                    ← gRPC server struct + helpers (levenshtein, soundex)
│       ├── patient_rpcs.go              ← List/Get/Bundle/Create/Update/Delete/Search/Match/History/Timeline
│       ├── encounter_rpcs.go            ← List/Get/Create/Update
│       ├── observation_rpcs.go          ← List/Get/Create
│       ├── condition_rpcs.go            ← List/Get/Create/Update
│       ├── medrq_rpcs.go               ← List/Get/Create/Update (MedicationRequest)
│       ├── allergy_rpcs.go              ← List/Get/Create/Update (AllergyIntolerance)
│       ├── immunization_rpcs.go         ← List/Get/Create (Immunization — patient-scoped)
│       ├── procedure_rpcs.go           ← List/Get/Create (Procedure — patient-scoped)
│       ├── generic_rpcs.go             ← Create/Get/List/Update/Delete (Practitioner/Organization/Location — top-level)
│       ├── flag_rpcs.go                 ← Create/Update (Sentinel write-back)
│       ├── batch_rpcs.go               ← CreateBatch (atomic multi-resource commit)
│       ├── index_rpcs.go               ← RebuildIndex, CheckIndexHealth, ReindexResources
│       └── health_rpcs.go              ← Health check
└── patient_test.go                      ← Integration tests (full gRPC roundtrip)
```

**Write pipeline (pipeline/writer.go):**
1. Validate FHIR JSON (pkg/fhir)
2. Assign UUID if CREATE
3. Set meta.lastUpdated/versionId/source
4. Acquire sync.Mutex (5s timeout)
5. Write JSON to Git + commit (pkg/gitstore)
6. Extract fields + upsert SQLite (pkg/fhir + pkg/sqliteindex)
7. Update patient_summaries
8. **Auto-generate FHIR Provenance** (target ref, activity coding, agents) → write to Git (skip if resourceType == "Provenance")
9. Release mutex, return resource + git metadata

**Error handling (spec §11):** Validation→INVALID_ARGUMENT, NotFound→NOT_FOUND, LockTimeout→ABORTED, GitFail→INTERNAL+rollback, SQLiteFail→log warning (data safe in Git).

**Patient matching (spec §7):** Weighted scoring (family 0.30, fuzzy 0.20, given 0.15, gender 0.10, birth year 0.10, district 0.05) with Levenshtein distance and Soundex phonetic matching.

---

## Cross-Cutting Patterns

### Response Envelope
Every response (success or error) goes through `model.JSON()` → `model.Envelope{}`. Handlers call `model.Success()`, `model.SuccessWithPagination()`, or `model.WriteError()`. Write operations use `writeResponseWithGit()` to include git metadata in the envelope. Never write raw JSON.

### Error Propagation
```
Service returns error  →  Handler calls model.WriteError(code, msg)  →  Envelope with status:"error"
```
gRPC unavailable errors map to `ErrServiceUnavailable` (503). Validation errors map to `ErrValidation` (400). The `ErrorHTTPStatus` map in `model/errors.go` is the single source of truth for code→status mapping.

### JSON Schema Validation
POST/PUT requests for FHIR resources are validated against JSON schemas loaded at startup. The `SchemaValidator` middleware reads the request body, validates against the registered schema, resets the body for downstream handlers, and returns 400 with VALIDATION_ERROR on failure.

### Testing Strategy
- Middleware tests: pass `httptest.Request` through middleware, assert on `httptest.Recorder` status + body + context values.
- Handler tests: inject mock service implementations (function fields), assert on response envelope. Mock types use embedded interface for convenience.
- Integration tests (router_test.go): wire real middleware + mock services, test full request flow (login → list patients, 401 without JWT, 503 for service unavailable, no more 501s on stubbed routes).
- **E2E smoke tests** (`test/e2e/smoke_test.go`): Boot all 3 microservices (Auth, Patient, Sync) in-process on dynamic ports, wire the full gateway HTTP handler with real JWT validation, test the complete REST flow (auth → CRUD → sync). 11 tests covering health, auth enforcement, CRUD, sync status, token refresh, and logout. Run via `make test-e2e`.

### Test Helper Packages
Exported test helpers that wrap internal service setup for E2E tests (Go's `internal` package restriction prevents direct imports from `test/e2e/`):
- `services/auth/authtest/` — Starts in-process Auth Service, exposes `Addr`, `PublicKey`, `GetChallenge()`, `AuthenticateWithNonce()`
- `services/patient/patienttest/` — Starts in-process Patient Service, exposes `Addr`
- `services/sync/synctest/` — Starts in-process Sync Service, exposes `Addr`

Each package also exports a `StartStandalone()` function that returns `(env, cleanup, error)` instead of requiring `*testing.T`. Used by the smoke test CLI.

### Interactive Smoke Test CLI (`cmd/smoke/`)
Standalone Go program that boots all 5 services (Auth, Patient, Sync, Formulary, Anchor) + gateway in-process, runs 27 REST steps with colored PASS/FAIL output. No external deps, no `*testing.T` — just `go run ./cmd/smoke` or `make smoke`. Exercises: health, auth enforcement, full CRUD (patient + 5 clinical resources), timeline, history, sync, conflicts, formulary (search, interactions, allergy), anchor (status, trigger, DID, backends, queue), schema rejection, and delete. Exit code 0/1 for CI.

---

## What's Implemented vs Stubbed

| Area | Status | Handler | Service Adapter |
|------|--------|---------|-----------------|
| Auth (register/challenge/authenticate/refresh/logout/validate/roles/devices) | Handler complete, gRPC adapter wired to auth service :50053 | auth.go | auth.go |
| Patient reads (list/get/search) | Handler complete, gRPC adapter wired to patient service :50051 | patient.go | patient.go |
| Patient writes (create/update/delete) | Handler complete, gRPC adapter wired to patient service :50051 | patient.go | patient.go |
| Patient match/history/timeline | Handler complete, gRPC adapter wired to patient service :50051 | patient.go | patient.go |
| Encounters (list/get/create/update) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Observations (list/get/create) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Conditions (list/create/update) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Medication Requests (list/create/update) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Allergy Intolerances (list/create/update) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Immunizations (list/get/create) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Procedures (list/get/create) | Handler complete, gRPC adapter wired to patient service :50051 | clinical.go | patient.go |
| Practitioners (list/get/create/update) | Handler complete (ResourceHandler factory), gRPC adapter wired to patient service :50051 | resource.go | patient.go |
| Organizations (list/get/create/update) | Handler complete (ResourceHandler factory), gRPC adapter wired to patient service :50051 | resource.go | patient.go |
| Locations (list/get/create/update) | Handler complete (ResourceHandler factory), gRPC adapter wired to patient service :50051 | resource.go | patient.go |
| FHIR CapabilityStatement (/fhir/metadata) | Auto-generated from resource registry, no auth | resource.go | — |
| FHIR Bundle/OperationOutcome builders | Library-only (pkg/fhir), ready for Phase 2 /fhir/ routes | — | — |
| Provenance auto-generation | Auto-generated after every write in pipeline, committed to Git | — | writer.go |
| Resource Registry | Central registry of 15 FHIR types with scope, interactions, search params | — | registry.go |
| Sync (status/peers/trigger/cancel/history/bundle/transports/events) | Handler complete, gRPC adapter wired to sync service :50052 | sync.go | sync.go |
| Conflicts (list/get/resolve/defer) | Handler complete, gRPC adapter wired to sync service :50052 | conflict.go | conflict.go |
| Alerts (list/get/acknowledge/dismiss/summary) | Handler complete, gRPC adapter wired to sentinel service :50056 | sentinel.go | sentinel.go |
| Formulary (16 RPCs: drug lookup, interactions, allergy, dosing, stock, redistribution, info) | Handler complete, gRPC adapter wired to formulary service :50054 | formulary.go | formulary.go |
| Anchor (14 RPCs: anchoring, DID, credentials, backend, queue, health) | Handler complete, gRPC adapter wired to anchor service :50055 | anchor.go | anchor.go |
| Supply chain (inventory/deliveries/predictions/redistribution) | Handler complete, gRPC adapter wired to sentinel service :50056 | supply.go | supply.go |
| JSON Schema Validation | 8 hardened schemas (Reference, CodeableConcept, status enums, required fields mirror validate.go) | — | validator.go |
| WebSocket (/ws) | 501 stub | stubs.go | — |

---

## Adding a New Endpoint (Checklist)

1. **Proto:** Define RPC + request/response messages in the appropriate `proto/*/v1/*.proto`
2. **Service interface:** Add method to interface in `service/interfaces.go`, add DTOs
3. **Service adapter:** Implement in `service/<domain>.go` using `pool.Conn("<service>")`
4. **Handler:** Add method to handler struct in `handler/<domain>.go`
5. **Router:** Wire the handler method in `router/router.go`
6. **Schema:** If POST/PUT with FHIR body, add JSON schema in `schemas/` and register in `main.go`
7. **Tests:** Unit test handler with mock service, add integration case in `router_test.go`
8. **Update this file**

---

## Auth Service (services/auth/)

Ed25519 challenge-response authentication, EdDSA JWT issuance, device registry in Git, RBAC with 5 roles.

```
services/auth/
├── cmd/main.go                          ← gRPC server entrypoint, port :50053
├── config.yaml                          ← default config
├── internal/
│   ├── config/config.go                 ← koanf config loader
│   ├── store/
│   │   ├── schema.go                    ← SQLite tables: deny_list, revocations, node_info
│   │   └── denylist.go                  ← In-memory + SQLite deny list for JTI revocation
│   ├── service/
│   │   ├── auth.go                      ← AuthService: register, challenge, authenticate, refresh, logout, validate, revoke
│   │   └── device.go                    ← Git-backed device registry (CRUD .nucleus/devices/*.json)
│   └── server/
│       ├── server.go                    ← gRPC server struct + error mapping
│       ├── auth_rpcs.go                 ← RegisterDevice, GetChallenge, Authenticate, RefreshToken, Logout, GetCurrentIdentity
│       ├── device_rpcs.go               ← ListDevices, RevokeDevice, CheckRevocation
│       ├── role_rpcs.go                 ← ListRoles, GetRole, AssignRole
│       ├── validation_rpcs.go           ← ValidateToken, CheckPermission
│       └── health_rpcs.go              ← Health
└── auth_test.go                         ← 12 integration tests (bootstrap, full auth cycle, brute force, revocation, etc.)
```

**Auth flow:** RegisterDevice → GetChallenge (32-byte nonce) → Authenticate (Ed25519 sig of nonce) → JWT issued → ValidateToken (<1ms, all in-memory)

**Token validation:** VerifyToken parses JWT → check deny list (in-memory map) → check device revocation list. All O(1), no I/O.

**RBAC:** 5 roles (CHW, Nurse, Physician, SiteAdmin, RegionalAdmin) × 37 permissions. Site scope: "local" (single site) or "regional" (cross-site).

---

## Sync Service (services/sync/)

Transport-agnostic Git sync, FHIR-aware merge driver, conflict resolution, event bus.

```
services/sync/
├── cmd/main.go                          ← gRPC server entrypoint, port :50052
├── config.yaml                          ← default config
├── internal/
│   ├── config/config.go                 ← koanf config loader
│   ├── store/
│   │   ├── schema.go                    ← SQLite tables: conflicts, sync_history, peer_state
│   │   ├── conflicts.go                 ← ConflictStore: Create, Get, List (with filters), Resolve, Defer
│   │   ├── history.go                   ← HistoryStore: Record, List, Get, RecordCompleted, RecordFailed
│   │   └── peers.go                     ← PeerStore: Upsert, Get, List, Trust, Untrust, MarkRevoked
│   ├── transport/
│   │   ├── adapter.go                   ← Adapter interface (Name, Capabilities, Start, Stop, Discover, Connect)
│   │   ├── stubs.go                     ← StubAdapter for unimplemented transports
│   │   └── localnet/localnet.go         ← Local network adapter (mDNS + gRPC over TCP)
│   ├── service/
│   │   ├── eventbus.go                  ← EventBus: pub/sub with type filtering, 7 event types
│   │   ├── syncengine.go               ← SyncEngine: orchestrator, TriggerSync, CancelSync, ExportBundle, ImportBundle
│   │   ├── syncqueue.go                ← SyncQueue: priority queue for sync jobs
│   │   └── bundle.go                   ← Bundle format placeholder
│   └── server/
│       ├── server.go                    ← gRPC server struct + error mapping
│       ├── sync_rpcs.go                 ← GetStatus, TriggerSync, CancelSync, ListPeers, TrustPeer, UntrustPeer, GetHistory
│       ├── conflict_rpcs.go             ← ListConflicts, GetConflict, ResolveConflict, DeferConflict
│       ├── transport_rpcs.go            ← ListTransports, EnableTransport, DisableTransport
│       ├── event_rpcs.go               ← SubscribeEvents (server-streaming)
│       ├── bundle_rpcs.go              ← ExportBundle, ImportBundle
│       ├── nodesync_rpcs.go            ← Handshake, RequestPack, SendPack (stubs for node-to-node)
│       └── health_rpcs.go              ← Health
└── sync_test.go                         ← 12 integration tests
```

**Merge Driver:** Three-tier classification: AutoMerge (non-overlapping) → Review (overlapping non-clinical) → Block (clinical safety risk). Block rules: allergy criticality, drug interaction, diagnosis conflict, patient identity, contradictory vitals.

**Transport:** Pluggable via Adapter interface. Local network (mDNS discovery), Wi-Fi Direct, Bluetooth, USB (stubs). Transport selection is automatic.

**Event Bus:** 7 event types (sync.started/completed/failed, peer.discovered/lost, conflict.new/resolved). Server-streaming gRPC for real-time updates.

---

## Shared Libraries — Auth + Merge

### pkg/auth — Shared Auth Utilities
- **crypto.go** — Ed25519 `GenerateKeypair()`, `Sign()`, `Verify()`, `EncodePublicKey()`, `DecodePublicKey()`
- **jwt.go** — `NucleusClaims`, `SignToken()`, `VerifyToken()` — EdDSA JWT via golang-jwt/v5
- **nonce.go** — `NonceStore` with TTL, `Generate()`, `Consume()`, `Cleanup()`
- **keystore.go** — `KeyStore` interface, `MemoryKeyStore`, `FileKeyStore` (0600 perms)
- **roles.go** — 37 permission constants, 5 role definitions, `HasPermission()`, `AllRoles()`
- **bruteforce.go** — `BruteForceGuard` with sliding window (N fails / M seconds)
- **auth_test.go** — 19 tests

### pkg/merge — FHIR-Aware Merge Driver
- **types.go** — `ConflictLevel` (AutoMerge/Review/Block), `FieldMergeStrategy`, `SyncPriority` (5 tiers)
- **diff.go** — `DiffResources()`, `DiffResourcesWithBase()`, `OverlappingFields()`, `NonOverlappingFields()`
- **classify.go** — `Classifier` with block rules per resource type, optional `FormularyChecker`
- **strategy.go** — Field merge strategies (LatestTimestamp, KeepBoth, PreferLocal) per resource type
- **driver.go** — `Driver` with `MergeFile()` and `MergeFields()` for three-way merge
- **priority.go** — `ClassifyResource()` → 5-tier sync priority based on resource type and status
- **merge_test.go** — 19 tests

### pkg/sync — Transport-Layer Cryptography

ECDH-based key exchange and AES-256-GCM authenticated encryption for node-to-node sync bundles. Replaces the previous broken scheme that prepended the AES key to ciphertext.

- **transport_crypto.go** — `DeriveSharedKey()` (Ed25519 → X25519 → ECDH → HKDF-SHA256), `EncryptPayload()`, `DecryptPayload()` (AES-256-GCM)
- **transport_crypto_test.go** — 11 tests (shared key derivation, round-trip, wrong-key rejection, determinism, nonce uniqueness, edge cases)

**Key design decisions:**
- Ed25519 → X25519 conversion: private key via SHA-512 + clamping (RFC 8032), public key via Edwards → Montgomery (`u = (1+y)/(1-y) mod p`)
- HKDF salt: `open-nucleus-sync-v1`, info: `transport-encryption`
- Bundle export uses ECIES pattern: ephemeral keypair per bundle, ephemeral public key prepended to ciphertext
- No external deps beyond `golang.org/x/crypto` (curve25519, hkdf)

## Formulary Service (services/formulary/)

Port :50054, 16 RPCs. Drug database, interaction checking, allergy cross-reactivity, stock management. Dosing RPCs return "not configured" cleanly (awaiting open-pharm-dosing integration).

```
services/formulary/
├── cmd/main.go                  ← gRPC entrypoint
├── config.yaml                  ← default config (root: formulary_service)
├── internal/
│   ├── config/config.go         ← koanf loader
│   ├── store/
│   │   ├── schema.go            ← SQLite: stock_levels + deliveries tables
│   │   ├── stock.go             ← StockStore CRUD
│   │   ├── drugdb.go            ← In-memory DrugDB from JSON seed data
│   │   └── interaction.go       ← InteractionIndex: O(1) pair lookup + class + allergy
│   ├── dosing/engine.go         ← Engine interface + StubEngine
│   ├── service/formulary.go     ← Core business logic (search, interactions, stock, predictions)
│   └── server/
│       ├── server.go            ← gRPC server + mapError
│       ├── medication_rpcs.go   ← Search, Get, ListByCategory
│       ├── interaction_rpcs.go  ← CheckInteractions, CheckAllergyConflicts
│       ├── dosing_rpcs.go       ← Validate, Options, Schedule (stub)
│       ├── stock_rpcs.go        ← StockLevel, Update, Delivery, Prediction, Redistribution
│       ├── formulary_rpcs.go    ← GetFormularyInfo
│       └── health_rpcs.go       ← Health
├── formulary_test.go            ← 26 integration tests
├── formularytest/
│   ├── setup.go                 ← Start(*testing.T, tmpDir)
│   └── standalone.go            ← StartStandalone(tmpDir)
└── testdata/
    ├── medications/             ← 20 WHO essential medicine JSONs
    └── interactions/            ← 17 interaction rules + 4 allergy cross-reactivity rules
```

**Key design decisions:**
- **DrugDB**: In-memory map loaded from embedded JSON. Case-insensitive substring search.
- **InteractionIndex**: Canonical key `min(a,b):max(a,b)` for O(1) pair lookup. Separate class-level and allergy indexes.
- **CheckInteractions**: pair lookup → class lookup → allergy check → stock check → classify overall risk.
- **Stock prediction**: `daysRemaining = quantity / dailyRate`, risk classification (critical/high/moderate/low).
- **Redistribution**: surplus (>90 days supply) vs shortage (<14 days), suggests transfers.
- **Dosing**: `Engine` interface with `StubEngine` that returns `configured=false`. 3 dosing RPCs cleanly signal "not configured" without gRPC errors.

## pkg/merge/openanchor — Anchor Cryptography Library

Interfaces + local implementations for Merkle trees, DID:key, and Verifiable Credentials. No external dependencies beyond Go stdlib. Designed to be replaced by the real `open-anchor` library later.

- **interfaces.go** — `AnchorEngine`, `IdentityEngine`, `MerkleTree` interfaces + all types (`DIDDocument`, `VerifiableCredential`, `CredentialProof`, `AnchorReceipt`, `CredentialClaims`, `VerificationResult`, `AnchorResult`, `FileEntry`) + sentinel errors
- **merkle.go** — SHA-256 Merkle tree: sort by path, `H(path||fileHash)` per leaf, binary tree bottom-up, duplicate odd leaf
- **base58.go** — Base58btc encoder/decoder (Bitcoin alphabet, ~60 lines)
- **didkey.go** — `did:key` from Ed25519: multicodec prefix `0xed01` + pubkey → base58btc → `did:key:z...`. `ResolveDIDKey()` parses back to `DIDDocument`
- **credential.go** — `IssueCredentialLocal()` — build VC, sign canonicalized payload with Ed25519. `VerifyCredentialLocal()` — resolve issuer DID, verify signature
- **stub_backend.go** — `StubBackend`: `Anchor()` returns `ErrBackendNotConfigured`, `Available()` returns false, `Name()` returns "none"
- **local_identity.go** — `LocalIdentityEngine`: delegates to DIDKeyFromEd25519, ResolveDIDKey, IssueCredentialLocal, VerifyCredentialLocal
- **openanchor_test.go** — 13 unit tests (Merkle, base58, DID:key, VC, stub backend)

## Anchor Service (services/anchor/)

Port :50055, 14 RPCs. Merkle anchoring, DID management, Verifiable Credentials, queue management. Blockchain backend uses StubBackend (anchors queued in SQLite but never submitted).

```
services/anchor/
├── cmd/main.go                          ← gRPC entrypoint
├── config.yaml                          ← default config (root: anchor_service)
├── internal/
│   ├── config/config.go                 ← koanf loader
│   ├── store/
│   │   ├── schema.go                    ← SQLite: anchor_queue table + indexes
│   │   ├── queue.go                     ← AnchorQueue: Enqueue, ListPending, CountPending, CountTotal
│   │   ├── anchors.go                   ← Git-backed anchor record CRUD (.nucleus/anchors/)
│   │   ├── credentials.go              ← Git-backed credential CRUD (.nucleus/credentials/)
│   │   └── dids.go                      ← Git-backed DID document CRUD (.nucleus/dids/)
│   ├── service/anchor.go               ← Core business logic (14 methods)
│   └── server/
│       ├── server.go                    ← gRPC server struct + mapError
│       ├── anchor_rpcs.go              ← GetStatus, TriggerAnchor, Verify, GetHistory
│       ├── did_rpcs.go                 ← GetNodeDID, GetDeviceDID, ResolveDID
│       ├── credential_rpcs.go          ← IssueDataIntegrityCredential, VerifyCredential, ListCredentials
│       ├── backend_rpcs.go             ← ListBackends, GetBackendStatus, GetQueueStatus
│       └── health_rpcs.go             ← Health
├── anchor_test.go                       ← 19 integration tests
├── anchortest/
│   ├── setup.go                         ← Start(*testing.T, tmpDir)
│   └── standalone.go                    ← StartStandalone(tmpDir)
```

**Key design decisions:**
- **Crypto in `pkg/merge/openanchor/`**: Clean swap to real open-anchor later; service codes to interfaces.
- **did:key only** (no ledger DIDs in V1): Fully offline, deterministic from Ed25519.
- **SQLite for queue, Git for records/credentials/DIDs**: Queue is transient; records are source of truth (syncs via Git).
- **StubBackend**: Returns `ErrBackendNotConfigured`. Queue fills, never drains. Same pattern as formulary dosing stub.
- **Merkle tree excludes `.nucleus/`**: Only clinical data files are included in the tree; internal metadata is excluded.
- **TriggerAnchor workflow**: TreeWalk → SHA-256 each file → Merkle root → skip if unchanged (unless manual) → attempt engine.Anchor() → enqueue on failure → save record in Git.

## Sentinel Agent Service (services/sentinel/) — Python

Port :50056 (gRPC), :8090 (HTTP management). The first Python microservice. Implements all 10 sentinel proto RPCs (5 alert + 5 supply) with in-memory stores and seed data. Stubs `open-sentinel` interfaces for future swap.

```
services/sentinel/
├── pyproject.toml                       ← Python project config
├── requirements.txt                     ← Pinned deps
├── config.yaml                          ← Default config
├── proto_gen.sh                         ← Generate Python proto stubs
├── src/sentinel/
│   ├── main.py                          ← Async entrypoint (gRPC + HTTP + background tasks)
│   ├── config.py                        ← SentinelConfig + OllamaConfig dataclasses, YAML loader
│   ├── sync_subscriber.py               ← Sync Service event stream skeleton (stub)
│   ├── fhir_output.py                   ← Alert → FHIR DetectedIssue conversion, EmissionQueue
│   ├── gen/                             ← Generated proto Python code (committed)
│   │   ├── common/v1/                   ← PaginationRequest/Response
│   │   └── sentinel/v1/                 ← SentinelService stub/servicer, all message types
│   ├── server/
│   │   ├── servicer.py                  ← SentinelServiceServicer (10 RPCs)
│   │   └── converters.py                ← Proto ↔ domain model converters
│   ├── http/
│   │   └── health_server.py             ← aiohttp server (13 HTTP endpoints)
│   ├── store/
│   │   ├── models.py                    ← Alert, InventoryItem, DeliveryRecord, SupplyPrediction, etc.
│   │   ├── alert_store.py               ← Thread-safe in-memory alert store
│   │   ├── inventory_store.py           ← Thread-safe in-memory inventory store
│   │   └── seed.py                      ← 5 alerts + 10 inventory items + predictions + redistributions
│   ├── ollama/
│   │   └── sidecar.py                   ← OllamaSidecar: start/stop/watchdog/health
│   └── agent/
│       ├── interfaces.py                ← ABCs: SentinelSkill, DataAdapter, AlertOutput, MemoryStore, LLMEngine
│       └── stub.py                      ← StubAgent (logs "open-sentinel not configured")
└── tests/                               ← 68 pytest tests
    ├── conftest.py                      ← Fixtures: seeded stores, in-process gRPC server
    ├── test_config.py                   ← 4 tests
    ├── test_alert_store.py              ← 11 tests
    ├── test_inventory_store.py          ← 11 tests
    ├── test_grpc_servicer.py            ← 17 tests (all 10 RPCs)
    ├── test_health_server.py            ← 13 tests (all HTTP endpoints)
    └── test_fhir_output.py              ← 12 tests (FHIR conversion, provenance, queue)
```

**Key design decisions:**
- **In-memory stores**: Thread-safe dicts with seed data. No SQLite/Git yet — stores are populated at startup and persist for session lifetime.
- **Seed data**: 5 realistic alerts (cholera cluster, measles, stockout, drug interaction, BP trend) + 10 WHO essential medicines across 2 sites + supply predictions + redistribution suggestions.
- **StubAgent pattern**: Same as formulary dosing stub — clean interfaces with stub implementations that log "not configured". When `open-sentinel` exists, swap StubAgent for real SentinelAgent.
- **FHIR output**: Full DetectedIssue conversion with AI provenance tags (rule-only vs ai-generated), severity mapping, reasoning extensions. EmissionQueue stubs the Patient Service write-back.
- **Ollama sidecar**: Process manager with crash recovery (max 5 restarts), health monitoring, watchdog loop. Disabled by default.

---

## FHIR Phase 2 — REST API Layer

**Goal:** Standards-compliant FHIR R4 REST API at `/fhir/{Type}` running parallel to the existing `/api/v1/` endpoints.

**Key differences from `/api/v1/`:**
- Raw FHIR JSON responses (no envelope wrapper)
- FHIR Bundle for search results (not arrays)
- OperationOutcome for errors (not custom error codes)
- `Content-Type: application/fhir+json` on all responses
- ETag / If-None-Match for conditional reads (304 Not Modified)
- XML requests rejected with 406

**Architecture:**

```
internal/handler/fhir/
├── fhir.go          ← FHIRHandler struct + dynamic route registration
├── response.go      ← FHIR response writers (resource, bundle, error, 304)
├── middleware.go     ← Content negotiation middleware (JSON only)
├── params.go        ← FHIR search parameter parser (_count, _offset, patient)
├── dispatch.go      ← Resource type → service call dispatch table
├── read.go          ← GET /fhir/{Type}/{id}
├── search.go        ← GET /fhir/{Type} → Bundle
├── write.go         ← POST/PUT/DELETE handlers
├── everything.go    ← GET /fhir/Patient/{id}/$everything
└── fhir_test.go     ← 22 tests
```

**Dispatch pattern:** `map[string]*ResourceDispatch` built at init, each entry closes over `PatientService` methods. Reads go through expanded `GetResource` RPC (all 15 types). Searches call type-specific list methods. Writes extract patient reference from body for patient-scoped types.

**ID-only lookups:** 8 new `GetXByID(id)` methods on SQLite Index (drop `AND patient_id = ?`) enabling FHIR-standard `GET /fhir/Encounter/{id}` without patient ID in URL.

**Route count:** ~50 new FHIR endpoints auto-generated from 15 resource type definitions.

---

## FHIR Phase 3 — Open Nucleus FHIR Profiles

**Goal:** FHIR profiles specific to African healthcare deployment — custom extensions for national IDs, WHO vaccine codes, AI provenance, growth monitoring, and DHIS2 reporting. Adds MeasureReport as a new resource type and StructureDefinition as a read-only endpoint for profile discovery.

**Five profiles:**

| Profile | Base | Extensions |
|---------|------|------------|
| OpenNucleus-Patient | Patient | national-health-id (valueIdentifier), ethnic-group (valueCoding) |
| OpenNucleus-Immunization | Immunization | dose-schedule-name (valueString), dose-expected-age (valueString) + CVX/ATC warning |
| OpenNucleus-GrowthObservation | Observation | who-zscore (valueDecimal), nutritional-classification (valueCoding) + growth code + vital-signs constraints |
| OpenNucleus-DetectedIssue | DetectedIssue | ai-model-name, ai-confidence-score, ai-reflection-count, ai-reasoning-chain |
| OpenNucleus-MeasureReport | MeasureReport | dhis2-data-element, dhis2-org-unit, dhis2-period |

**New resource types:** MeasureReport (full stack: type → registry → validation → extraction → Git path → soft delete → SQLite schema/index → pipeline → RPCs → dispatch), StructureDefinition (read-only, served from profile registry).

**Architecture:**

```
pkg/fhir/
├── extension.go              ← ExtensionDef, ExtractExtension, HasExtension, ValidateExtensions
├── profile.go                ← Profile registry (GetProfileDef, AllProfileDefs, ProfilesForResource, GetMetaProfiles)
├── profile_defs.go           ← 5 profile builders with validation functions
├── structuredefinition.go    ← GenerateStructureDefinition, GenerateAllStructureDefinitions
├── validate.go               ← +ValidateWithProfile, +validateMeasureReport (profile-aware validation)
├── types.go                  ← +ResourceMeasureReport, +ResourceStructureDefinition, +MeasureReportRow
├── registry.go               ← +MeasureReport (SystemScoped), +StructureDefinition (SystemScoped, read-only)
├── extract.go                ← +ExtractMeasureReportFields
├── path.go                   ← +measure-reports/, +.nucleus/profiles/
├── softdelete.go             ← +MeasureReport → status="error"
└── capability.go             ← +supportedProfile per resource type
```

**Profile validation:** `ValidateWithProfile` runs base `Validate` then checks `meta.profile` URLs against the profile registry. Each profile can have required extensions, value type checks, and custom constraint functions (e.g. growth code whitelist, CVX/ATC warning). Unknown extensions pass through (FHIR open model).

**StructureDefinition endpoint:** `GET /fhir/StructureDefinition` returns all 5 profiles as FHIR R4 StructureDefinition resources generated from ProfileDef metadata.

**Resource count:** 15 → 17 (MeasureReport + StructureDefinition). 58 pkg/fhir tests (26 new).

---

## FHIR Phase 4 — SMART on FHIR

**Goal:** OAuth2 authorization code flow with SMART on FHIR v2 scopes, enabling third-party clinical apps (growth chart widgets, immunization trackers, DHIS2 connectors) to connect securely via standardized launch protocols. All OAuth2 flows execute on the local node — no cloud IdP required.

**Coexistence model:** Internal devices use Ed25519 challenge-response. SMART apps use OAuth2 auth code + PKCE. Both produce EdDSA JWTs — SMART tokens carry additional `scope`, `client_id`, and launch context claims. FHIR endpoints enforce SMART scopes when present, otherwise fall back to existing RBAC.

**Architecture:**

```
pkg/smart/
├── scope.go          ← SMART v2 scope parser (patient/Resource.cruds)
├── client.go         ← Client model + validation (pending/approved/revoked)
├── authcode.go       ← Auth code + PKCE (S256, one-shot exchange)
├── launch.go         ← EHR launch token store (one-shot consume)
└── config.go         ← SMART configuration builder (/.well-known/smart-configuration)

proto/smart/v1/
└── smart.proto       ← SmartService (11 RPCs: OAuth2, client mgmt, launch, health)

services/auth/
├── internal/store/clients.go   ← Client storage (Git + SQLite dual store)
├── internal/service/smart.go   ← SmartService implementation
└── internal/server/smart_rpcs.go ← gRPC server adapter

internal/
├── service/smart.go           ← SmartService interface + gRPC adapter
├── handler/smart.go           ← 11 HTTP endpoints (OAuth2 + admin)
└── middleware/smartscope.go   ← SMART scope enforcement on FHIR routes
```

**OAuth2 endpoints:**

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| GET | `/.well-known/smart-configuration` | Public | SMART discovery |
| GET | `/auth/smart/authorize` | JWT | Authorization (returns redirect with code) |
| POST | `/auth/smart/token` | Public | Token exchange (client auth via body/Basic) |
| POST | `/auth/smart/revoke` | JWT | Token revocation |
| POST | `/auth/smart/introspect` | JWT | Token introspection |
| POST | `/auth/smart/register` | JWT (admin) | Dynamic client registration |
| POST | `/auth/smart/launch` | JWT | Create EHR launch token |
| GET/PUT/DELETE | `/api/v1/smart/clients/{id}` | JWT (admin) | Client management |

**SMART scope middleware:** `SmartScope(resourceType, interaction)` enforces v2 scopes on all FHIR endpoints. Patient-context scopes restrict access to launch patient only. Wildcard resource (`*`) supported. No-scope tokens pass through to existing RBAC.

**CapabilityStatement:** Security section includes oauth-uris extension (authorize, token, revoke, register endpoints) and SMART-on-FHIR service coding when `SmartEnabled=true`.

**New permissions:** `smart:launch` (physician, site-admin, regional-admin), `smart:register` (site-admin, regional-admin).

**Test count:** 408 total (37 new SMART tests: 27 pkg/smart, 3 pkg/auth SMART claims, 6 smartscope middleware, 8 handler, 2 capability).

---

## Phase Roadmap

| Phase | Scope | Status |
|-------|-------|--------|
| 1 — Walking Skeleton | Middleware pipeline, auth + patient read handlers, all stubs | COMPLETE |
| 2 — Gateway Gaps | All handler/service/proto definitions, clinical sub-resources, JSON schema validation, zero stubs (except /ws) | COMPLETE |
| 3 — Patient Service | First real backend: `services/patient/` + `pkg/fhir` + `pkg/gitstore` + `pkg/sqliteindex`. 38 gRPC RPCs, full write pipeline, 40 tests passing | COMPLETE |
| 4 — Auth + Sync Services | Auth Service (15 RPCs, Ed25519 + JWT + RBAC) + Sync Service (~25 RPCs + NodeSyncService, FHIR merge driver, event bus) + `pkg/auth` + `pkg/merge`. 62 tests passing | COMPLETE |
| 4.5 — E2E Smoke Tests | Full-stack E2E tests (11 cases), JWT claims fix, patient gRPC adapter wiring, test helper packages | COMPLETE |
| 5 — Formulary + Anchor + Sentinel | Formulary COMPLETE (16 RPCs, 26 tests). Anchor COMPLETE (14 RPCs, 19 tests). Sentinel Agent COMPLETE (10 RPCs, 13 HTTP endpoints, 68 tests). Go gateway adapters wired for all 3. | COMPLETE |
| FHIR Phase 1 — Core Foundation | 5 new resource types (Immunization, Procedure, Practitioner, Organization, Location) + Provenance auto-generation. Resource registry (15 types), CapabilityStatement, Bundle/OperationOutcome builders. 49 Patient Service RPCs, ~70 gateway endpoints. 36 pkg/fhir tests. | COMPLETE |
| FHIR Phase 2 — REST API Layer | Standards-compliant `/fhir/{Type}` REST API. Raw FHIR JSON (no envelope), Bundle for search, OperationOutcome for errors, ETag/conditional reads. ~50 new endpoints auto-generated from resource registry. Dispatch table, content negotiation, $everything. 22 handler tests. | COMPLETE |
| FHIR Phase 3 — FHIR Profiles | 5 Open Nucleus profiles (Patient, Immunization, GrowthObservation, DetectedIssue, MeasureReport). Extension utilities, profile registry, profile-aware validation. MeasureReport full stack (17 resource types). StructureDefinition read-only endpoint. CapabilityStatement supportedProfile. 58 pkg/fhir tests. | COMPLETE |
| FHIR Phase 4 — SMART on FHIR | OAuth2 auth code + PKCE, SMART v2 scopes, EHR launch, client registration, scope middleware on FHIR endpoints. 11 gRPC RPCs, 11 HTTP endpoints, CapabilityStatement SMART security, 37 new tests (408 total). | COMPLETE |
| Overhaul Phase 3 — Sync Crypto Fix | Replaced broken AES-GCM (key-in-ciphertext) with ECDH X25519 + HKDF-SHA256 + AES-256-GCM. New `pkg/sync` (transport_crypto.go), ECIES-pattern bundle encryption in SyncEngine. 11 new crypto tests, 23 total sync tests. | COMPLETE |
| 6 — WebSocket + Hardening | Real-time events, production config, TLS, metrics | Not started |
