# Open Nucleus — Architectural Memory

> Living document. Updated after every major feature or structural change.
> Last updated: Phase 3 — Patient Service (2026-02-26)

---

## System Overview

### Open Nucleus
Open Nucleus is an open-source, offline-first electronic health record (EHR) system designed for military forward operating bases, disaster relief zones, and small clinics in sub-Saharan Africa. It assumes zero connectivity as the default and treats network access as a bonus.

### Core Architecture
Microservices in Go (Patient, Sync, Auth, Formulary, Anchor services) plus a Python Sentinel Agent, fronted by a Go API Gateway on port 8080 (REST/JSON). The Flutter frontend lives in a separate repo (open-nucleus-app) and consumes the gateway as a pure REST client.
Dual-layer data model: FHIR R4 resources are stored as JSON files in a Git repository (source of truth) with a SQLite database as a rebuildable query index. Every clinical write commits to Git first, then upserts SQLite. If SQLite is lost, it rebuilds from Git.
Git-based sync: Nodes discover each other via Wi-Fi Direct, Bluetooth, or local network and sync using Git fetch/merge/push. A FHIR-aware merge driver classifies conflicts into auto-merge (safe), review (flag for clinician), or block (clinical safety risk). Transport is pluggable and automatic.
Sentinel Agent: A "sleeper" AI agent that wakes on sync events, crawls the merged dataset for epidemiological outbreak signals, cross-site medication conflicts, missed referral follow-ups, and supply stockout predictions. V1 is rule-based using WHO IDSR thresholds.
IOTA Tangle anchoring: Git Merkle roots are periodically anchored to the IOTA Tangle (feeless), providing cryptographic proof of data integrity for regulatory compliance, humanitarian accountability, and supply chain provenance.

The API Gateway is a stateless Go HTTP server that sits between the Flutter frontend and 6 backend gRPC microservices. It owns no business logic beyond auth, authorization, validation, rate limiting, and response formatting. All clinical data passes through as opaque FHIR R4 JSON.

```
Flutter App (HTTP REST/JSON)
        │
        ▼
   ┌─────────┐
   │ Gateway  │  ← this repo
   └────┬─────┘
        │ gRPC
        ▼
  ┌──────────────────────────────────────────────┐
  │ Auth :50053  │ Patient :50051  │ Sync :50052  │
  │ Formulary :50054 │ Anchor :50055 │ Sentinel :50056 │
  └──────────────────────────────────────────────┘
```

---

## Dependency Wiring (main.go)

`cmd/gateway/main.go` is the composition root. It wires everything together in this order:

```
config.Load(path)
    │
    ▼
grpcclient.NewPool(cfg.GRPC)          ← dials 6 backend services (non-blocking)
    │
    ├─► service.NewAuthService(pool)   ← implements service.AuthService interface
    │       │
    │       ▼
    │   handler.NewAuthHandler(authSvc)
    │
    ├─► service.NewPatientService(pool) ← implements service.PatientService interface
    │       │
    │       ▼
    │   handler.NewPatientHandler(patientSvc)   ← also handles clinical sub-resources
    │
    ├─► service.NewSyncService(pool)
    │       ▼
    │   handler.NewSyncHandler(syncSvc)
    │
    ├─► service.NewConflictService(pool)
    │       ▼
    │   handler.NewConflictHandler(conflictSvc)
    │
    ├─► service.NewSentinelService(pool)
    │       ▼
    │   handler.NewSentinelHandler(sentinelSvc)
    │
    ├─► service.NewFormularyService(pool)
    │       ▼
    │   handler.NewFormularyHandler(formularySvc)
    │
    ├─► service.NewAnchorService(pool)
    │       ▼
    │   handler.NewAnchorHandler(anchorSvc)
    │
    ├─► service.NewSupplyService(pool)
    │       ▼
    │   handler.NewSupplyHandler(supplySvc)
    │
    ├─► middleware.NewSchemaValidator() + load 6 JSON schemas from schemas/
    │
    ├─► middleware.NewJWTAuth(pubKey, issuer)
    │
    ├─► middleware.NewRateLimiter(cfg.RateLimit)
    │
    ▼
router.New(Config{all handlers, middleware, schemaValidator, auditLogger, corsOrigins})
    │
    ▼
server.New(cfg, mux, logger).Run()    ← graceful shutdown on SIGINT/SIGTERM
```

---

## Package Dependency Graph

Arrows mean "imports / depends on". No circular dependencies exist.

```
cmd/gateway/main
    ├── internal/config
    ├── internal/grpcclient  ── internal/config
    ├── internal/service     ── internal/grpcclient
    ├── internal/handler     ── internal/service
    │                        ── internal/model
    ├── internal/middleware   ── internal/config  (ratelimit only)
    │                        ── internal/model    (all middleware)
    ├── internal/router      ── internal/handler
    │                        ── internal/middleware
    │                        ── internal/model
    └── internal/server      ── internal/config
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
- **interfaces.go** — 8 service interfaces (`AuthService`, `PatientService`, `SyncService`, `ConflictService`, `SentinelService`, `FormularyService`, `AnchorService`, `SupplyService`) + all DTOs. Handlers depend only on these interfaces, enabling mock-based testing.
- **auth.go** — `authAdapter` implements `AuthService` via `pool.Conn("auth")`.
- **patient.go** — `patientAdapter` implements `PatientService` (24 methods: list/get/search/create/update/delete + match/history/timeline + 15 clinical sub-resource methods) via `pool.Conn("patient")`.
- **sync.go** — `syncAdapter` implements `SyncService` (6 methods) via `pool.Conn("sync")`.
- **conflict.go** — `conflictAdapter` implements `ConflictService` (4 methods) via `pool.Conn("sync")` (conflicts are a sync sub-domain).
- **sentinel.go** — `sentinelAdapter` implements `SentinelService` (5 methods) via `pool.Conn("sentinel")`.
- **formulary.go** — `formularyAdapter` implements `FormularyService` (5 methods) via `pool.Conn("formulary")`.
- **anchor.go** — `anchorAdapter` implements `AnchorService` (4 methods) via `pool.Conn("anchor")`.
- **supply.go** — `supplyAdapter` implements `SupplyService` (5 methods) via `pool.Conn("sentinel")` (supply intelligence from Sentinel).

**Key pattern:** Handlers never touch gRPC directly. The service layer translates between HTTP DTOs and gRPC request/response types. This is where multi-service orchestration will live (e.g., MedRequest → Formulary check).

### internal/handler
- **auth.go** — `AuthHandler` holds `service.AuthService`. Methods: `Login`, `Refresh`, `Logout`, `Whoami`. Whoami short-circuits from JWT claims in context if available.
- **patient.go** — `PatientHandler` holds `service.PatientService`. Methods: `List`, `GetByID`, `Search`, `Create`, `Update`, `Delete`, `History`, `Timeline`, `Match`. Write methods use `writeResponseWithGit()` to include git metadata in the response envelope.
- **clinical.go** — Additional methods on `PatientHandler` for all 16 clinical sub-resource endpoints: `ListEncounters`, `GetEncounter`, `CreateEncounter`, `UpdateEncounter`, `ListObservations`, `GetObservation`, `CreateObservation`, `ListConditions`, `CreateCondition`, `UpdateCondition`, `ListMedicationRequests`, `CreateMedicationRequest`, `UpdateMedicationRequest`, `ListAllergyIntolerances`, `CreateAllergyIntolerance`, `UpdateAllergyIntolerance`.
- **sync.go** — `SyncHandler` holds `service.SyncService`. Methods: `Status`, `Peers`, `Trigger`, `History`, `ExportBundle`, `ImportBundle`.
- **conflict.go** — `ConflictHandler` holds `service.ConflictService`. Methods: `List`, `GetByID`, `Resolve`, `Defer`.
- **sentinel.go** — `SentinelHandler` holds `service.SentinelService`. Methods: `ListAlerts`, `Summary`, `GetAlert`, `Acknowledge`, `Dismiss`.
- **formulary.go** — `FormularyHandler` holds `service.FormularyService`. Methods: `SearchMedications`, `GetMedication`, `CheckInteractions`, `GetAvailability`, `UpdateAvailability`.
- **anchor.go** — `AnchorHandler` holds `service.AnchorService`. Methods: `Status`, `Verify`, `History`, `Trigger`.
- **supply.go** — `SupplyHandler` holds `service.SupplyService`. Methods: `Inventory`, `InventoryItem`, `RecordDelivery`, `Predictions`, `Redistribution`.
- **stubs.go** — `StubHandler()` returns 501 via `model.NotImplementedError()`. Only used for WebSocket endpoint (Phase 5).

### internal/router
- **router.go** — `New(Config)` builds the chi route tree. Config now includes all 8 handler types + `SchemaValidator`. `validatorMiddleware()` helper returns a no-op if SchemaValidator is nil (for tests without schemas). Owns middleware scoping:
  - `/health` — no middleware beyond global
  - `/api/v1/auth/*` — global + RateLimiter(CategoryAuth), NO JWT/RBAC
  - `/api/v1/*` (everything else) — global + JWTAuth, then per-route RateLimiter + RequirePermission + optional SchemaValidator
- All 58 REST endpoints wired to real handlers. Only `/ws` remains stubbed (Phase 5).

### internal/server
- **server.go** — `Server` wraps `http.Server` with config-driven timeouts. `Run()` starts listener and blocks until SIGINT/SIGTERM, then calls `Shutdown()` with 10s grace period.

### schemas/
- **patient.json** — Requires `resourceType: "Patient"`, `name` array (minItems 1), `gender` enum.
- **encounter.json** — Requires `resourceType: "Encounter"`, `status`, `class` object.
- **observation.json** — Requires `resourceType: "Observation"`, `status`, `code` object.
- **condition.json** — Requires `resourceType: "Condition"`, `code` object.
- **medication_request.json** — Requires `resourceType: "MedicationRequest"`, `status`, `medicationCodeableConcept` object.
- **allergy_intolerance.json** — Requires `resourceType: "AllergyIntolerance"`.

---

## Proto Structure

```
proto/
├── common/v1/
│   ├── metadata.proto   ← GitMetadata (+ Timestamp), PaginationRequest/Response, NodeInfo
│   └── fhir.proto       ← FHIRResource{resource_type, id, json_payload bytes}
├── auth/v1/
│   └── auth.proto       ← AuthService: Login, Refresh, Logout, Whoami RPCs
├── patient/v1/
│   └── patient.proto    ← PatientService: 38 RPCs (CRUD + clinical + batch + index + health)
├── sync/v1/
│   └── sync.proto       ← SyncService (6 RPCs) + ConflictService (4 RPCs)
├── formulary/v1/
│   └── formulary.proto  ← FormularyService: 5 RPCs (search, get, interactions, availability)
├── anchor/v1/
│   └── anchor.proto     ← AnchorService: 4 RPCs (status, verify, history, trigger)
└── sentinel/v1/
    └── sentinel.proto   ← SentinelService: 5 alert RPCs + 5 supply chain RPCs
```

FHIR resources are opaque `bytes json_payload` — the gateway never parses or transforms them.

Generated Go code lives in `gen/proto/` (protoc with go + go-grpc plugins).

---

## Shared Libraries (pkg/)

### pkg/fhir — FHIR R4 Utilities
Pure functions for working with FHIR resources. No I/O.
- **types.go** — Resource type constants (`ResourcePatient`, etc.), operation constants (`OpCreate`, etc.), row structs for all 7 resource types (`PatientRow`, `EncounterRow`, etc.), `FieldError`, `Pagination`, `PaginationOpts`, `TimelineEvent`.
- **path.go** — `GitPath(resourceType, patientID, resourceID)` returns Git file path per spec §3.3. `PatientDirPath(patientID)` for history queries.
- **meta.go** — `SetMeta()` writes `meta.lastUpdated/versionId/source`. `AssignID()` assigns UUID if absent. `GetResourceType()`, `GetID()`.
- **validate.go** — `Validate(resourceType, json)` performs Layer 1 structural validation. Per-type validators enforce required fields from spec §4.3.
- **extract.go** — `ExtractPatientFields()`, `ExtractEncounterFields()`, etc. Extract SQLite indexed columns from FHIR JSON.
- **softdelete.go** — `ApplySoftDelete()` mutates resource fields per spec §3.4 (Patient→active:false, Encounter→status:entered-in-error, etc.).

### pkg/gitstore — Git Operations
Wraps `go-git/v5` for clinical data Git repository management.
- **store.go** — `Store` interface: `WriteAndCommit()`, `Read()`, `LogPath()`, `Head()`, `TreeWalk()`, `Rollback()`. `NewStore(repoPath)` opens or inits repo.
- **commit.go** — `CommitMessage` struct with `Format()` and `ParseCommitMessage()` for structured commit messages per spec §3.3.

### pkg/sqliteindex — SQLite Query Index
Uses `modernc.org/sqlite` (pure Go, no CGO) for Raspberry Pi 4 deployment.
- **schema.go** — `InitSchema()` creates 9 tables (patients, encounters, observations, conditions, medication_requests, allergy_intolerances, flags, detected_issues, patient_summaries) + index_meta + FTS5 + triggers. `DropAll()` for rebuild.
- **index.go** — `Index` interface: Upsert/Get/List methods for all 7 resource types + bundle + search + timeline + match + meta + summary. `NewIndex(dbPath)` opens DB with WAL mode.
- **search.go** — FTS5 patient search via `patients_fts` virtual table.
- **timeline.go** — `GetTimeline()` UNION ALL query across encounters, observations, conditions, flags.
- **match.go** — `GetMatchCandidates()` broad SQL query for patient identity matching.
- **summary.go** — `UpdateSummary()` recomputes `patient_summaries` counts. `GetPatientBundle()` returns patient + all active child resources.

## Patient Service (services/patient/)

The first real backend microservice. Single writer for all clinical FHIR data: validate → Git commit → SQLite upsert → return resource + commit metadata.

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
8. Release mutex, return resource + git metadata

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

---

## What's Implemented vs Stubbed

| Area | Status | Handler | Service Adapter |
|------|--------|---------|-----------------|
| Auth (login/refresh/logout/whoami) | Handler complete, gRPC adapter stubbed | auth.go | auth.go |
| Patient reads (list/get/search) | Handler complete, gRPC adapter stubbed | patient.go | patient.go |
| Patient writes (create/update/delete) | Handler complete, gRPC adapter stubbed | patient.go | patient.go |
| Patient match/history/timeline | Handler complete, gRPC adapter stubbed | patient.go | patient.go |
| Encounters (list/get/create/update) | Handler complete, gRPC adapter stubbed | clinical.go | patient.go |
| Observations (list/get/create) | Handler complete, gRPC adapter stubbed | clinical.go | patient.go |
| Conditions (list/create/update) | Handler complete, gRPC adapter stubbed | clinical.go | patient.go |
| Medication Requests (list/create/update) | Handler complete, gRPC adapter stubbed | clinical.go | patient.go |
| Allergy Intolerances (list/create/update) | Handler complete, gRPC adapter stubbed | clinical.go | patient.go |
| Sync (status/peers/trigger/history/bundle) | Handler complete, gRPC adapter stubbed | sync.go | sync.go |
| Conflicts (list/get/resolve/defer) | Handler complete, gRPC adapter stubbed | conflict.go | conflict.go |
| Alerts (list/get/acknowledge/dismiss/summary) | Handler complete, gRPC adapter stubbed | sentinel.go | sentinel.go |
| Formulary (medications/interactions/availability) | Handler complete, gRPC adapter stubbed | formulary.go | formulary.go |
| Anchor/IOTA (status/verify/history/trigger) | Handler complete, gRPC adapter stubbed | anchor.go | anchor.go |
| Supply chain (inventory/deliveries/predictions/redistribution) | Handler complete, gRPC adapter stubbed | supply.go | supply.go |
| JSON Schema Validation | 6 schemas loaded, wired on POST/PUT clinical routes | — | validator.go |
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

## Phase Roadmap

| Phase | Scope | Status |
|-------|-------|--------|
| 1 — Walking Skeleton | Middleware pipeline, auth + patient read handlers, all stubs | COMPLETE |
| 2 — Gateway Gaps | All handler/service/proto definitions, clinical sub-resources, JSON schema validation, zero stubs (except /ws) | COMPLETE |
| 3 — Patient Service | First real backend: `services/patient/` + `pkg/fhir` + `pkg/gitstore` + `pkg/sqliteindex`. 38 gRPC RPCs, full write pipeline, 40 tests passing | COMPLETE |
| 4 — Sync + Conflicts + Sentinel | Real gRPC backend integration for sync, conflict resolution, alerts | Not started |
| 5 — Formulary + Anchor + Supply | Real gRPC backend integration for formulary, IOTA anchoring, supply chain | Not started |
| 6 — WebSocket + Hardening | Real-time events, production config, TLS, metrics | Not started |
