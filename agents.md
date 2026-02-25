# Open Nucleus — Architectural Memory

> Living document. Updated after every major feature or structural change.
> Last updated: Phase 1 — Walking Skeleton (2026-02-25)

---

## System Overview

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
    │   handler.NewPatientHandler(patientSvc)
    │
    ├─► middleware.NewJWTAuth(pubKey, issuer)
    │
    ├─► middleware.NewRateLimiter(cfg.RateLimit)
    │
    ▼
router.New(Config{handlers, middleware, auditLogger, corsOrigins})
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
CORS → RequestID → AuditLog → JWTAuth → [per-route: RateLimiter → RequirePermission] → Handler
```

**Auth routes skip** JWTAuth and RBAC — they only get CORS + RequestID + AuditLog + RateLimiter(CategoryAuth).

### internal/grpcclient
- **pool.go** — `Pool` holds a `map[string]*grpc.ClientConn` for 6 named services. `NewPool()` dials all with timeout (non-blocking on failure — stores nil, returns SERVICE_UNAVAILABLE at call time). `Conn(name)` returns connection or error.
- Consumed by: service adapters call `pool.Conn("auth")`, `pool.Conn("patient")`, etc.

### internal/service
- **interfaces.go** — `AuthService` and `PatientService` interfaces + all DTOs. Handlers depend only on these interfaces, enabling mock-based testing.
- **auth.go** — `authAdapter` struct implements `AuthService` by calling `pool.Conn("auth")` and making gRPC calls. Currently returns SERVICE_UNAVAILABLE since backends don't exist yet.
- **patient.go** — `patientAdapter` struct implements `PatientService` via `pool.Conn("patient")`. Same stub behavior.

**Key pattern:** Handlers never touch gRPC directly. The service layer translates between HTTP DTOs and gRPC request/response types. This is where multi-service orchestration will live (e.g., MedRequest → Formulary check).

### internal/handler
- **auth.go** — `AuthHandler` holds `service.AuthService`. Methods: `Login`, `Refresh`, `Logout`, `Whoami`. Whoami short-circuits from JWT claims in context if available.
- **patient.go** — `PatientHandler` holds `service.PatientService`. Methods: `List`, `GetByID`, `Search`. Uses `model.PaginationFromRequest()` + chi's `URLParam()`.
- **stubs.go** — `StubHandler()` returns 501 via `model.NotImplementedError()`. Used for all unimplemented endpoints.

### internal/router
- **router.go** — `New(Config)` builds the chi route tree. Owns middleware scoping:
  - `/health` — no middleware beyond global
  - `/api/v1/auth/*` — global + RateLimiter(CategoryAuth), NO JWT/RBAC
  - `/api/v1/*` (everything else) — global + JWTAuth, then per-route RateLimiter + RequirePermission
- All 58 REST endpoints + 1 WebSocket endpoint registered. Unimplemented ones point to `StubHandler`.
- Imports handler, middleware, and model — the only package that knows the full route topology.

### internal/server
- **server.go** — `Server` wraps `http.Server` with config-driven timeouts. `Run()` starts listener and blocks until SIGINT/SIGTERM, then calls `Shutdown()` with 10s grace period.

---

## Cross-Cutting Patterns

### Response Envelope
Every response (success or error) goes through `model.JSON()` → `model.Envelope{}`. Handlers call `model.Success()`, `model.SuccessWithPagination()`, or `model.WriteError()`. Never write raw JSON.

### Error Propagation
```
Service returns error  →  Handler calls model.WriteError(code, msg)  →  Envelope with status:"error"
```
gRPC unavailable errors map to `ErrServiceUnavailable` (503). Validation errors map to `ErrValidation` (400). The `ErrorHTTPStatus` map in `model/errors.go` is the single source of truth for code→status mapping.

### Testing Strategy
- Middleware tests: pass `httptest.Request` through middleware, assert on `httptest.Recorder` status + body + context values.
- Handler tests: inject mock service implementations (function fields), assert on response envelope.
- Integration tests (router_test.go): wire real middleware + mock services, test full request flow (login → list patients, 401 without JWT, 501 for stubs).

---

## Proto Structure

```
proto/
├── common/v1/
│   ├── metadata.proto   ← GitMetadata, PaginationRequest/Response, NodeInfo
│   └── fhir.proto       ← FHIRResource{resource_type, id, json_payload bytes}
├── auth/v1/
│   └── auth.proto       ← AuthService: Login, Refresh, Logout, Whoami RPCs
└── patient/v1/
    └── patient.proto    ← PatientService: CRUD + clinical sub-resources (27 RPCs)
```

FHIR resources are opaque `bytes json_payload` — the gateway never parses or transforms them.

---

## What's Implemented vs Stubbed

| Area | Status | Handler | Service Adapter |
|------|--------|---------|-----------------|
| Auth (login/refresh/logout/whoami) | Handler complete, gRPC adapter stubbed | auth.go | auth.go |
| Patient (list/get/search) | Handler complete, gRPC adapter stubbed | patient.go | patient.go |
| Patient writes (create/update/delete) | 501 stub | stubs.go | — |
| Clinical resources (encounters, observations, conditions, meds, allergies) | 501 stub | stubs.go | — |
| Patient history/timeline | 501 stub | stubs.go | — |
| Patient match | 501 stub | stubs.go | — |
| Sync (status/peers/trigger/history/bundle) | 501 stub | stubs.go | — |
| Conflicts (list/get/resolve/defer) | 501 stub | stubs.go | — |
| Alerts (list/get/acknowledge/dismiss/summary) | 501 stub | stubs.go | — |
| Formulary (medications/interactions/availability) | 501 stub | stubs.go | — |
| Anchor/IOTA (status/verify/history/trigger) | 501 stub | stubs.go | — |
| Supply chain (inventory/deliveries/predictions/redistribution) | 501 stub | stubs.go | — |
| WebSocket (/ws) | 501 stub | stubs.go | — |

---

## Adding a New Endpoint (Checklist)

1. **Proto:** Define RPC + request/response messages in the appropriate `proto/*/v1/*.proto`
2. **Service interface:** Add method to interface in `service/interfaces.go`, add DTOs
3. **Service adapter:** Implement in `service/<domain>.go` using `pool.Conn("<service>")`
4. **Handler:** Add method to handler struct in `handler/<domain>.go`
5. **Router:** Replace `StubHandler()` with the real handler method in `router/router.go`
6. **Tests:** Unit test handler with mock service, add integration case in `router_test.go`
7. **Update this file**

---

## Phase Roadmap

| Phase | Scope | Status |
|-------|-------|--------|
| 1 — Walking Skeleton | Middleware pipeline, auth + patient read handlers, all stubs | COMPLETE |
| 2 — Patient Writes + Clinical | Patient CRUD, encounters, observations, conditions, meds, allergies, formulary interaction checks | Not started |
| 3 — Sync + Conflicts + Sentinel | Sync endpoints, conflict resolution, alert endpoints | Not started |
| 4 — Formulary + Anchor + Supply | Formulary search, IOTA anchoring, supply chain | Not started |
| 5 — WebSocket + Hardening | Real-time events, production config, TLS, metrics | Not started |
