# Open Nucleus — API Gateway

Stateless Go HTTP server that translates REST/JSON into gRPC calls to 6 backend microservices. Sole entry point for the Flutter frontend. Runs locally on-device with no internet dependency.

**Spec:** [`api_gateway_spec.md`](./api_gateway_spec.md) | **Architecture:** [`agents.md`](./agents.md)

## Quick Start

```bash
# Build
make build

# Run (starts on :8080)
make run

# Test (27 tests, race detection)
make test
```

## Configuration

All settings in [`config.yaml`](./config.yaml) — server port, gRPC service addresses, rate limits, CORS origins, JWT issuer, timeouts.

## Project Structure

```
cmd/gateway/main.go              Entry point — wires config, gRPC pool, services, middleware, router
internal/
├── config/                      Koanf YAML config loader
├── server/                      HTTP server with graceful shutdown (SIGINT/SIGTERM)
├── router/                      chi route tree — all 58 REST endpoints + middleware scoping
├── middleware/                   8-stage pipeline (see below)
├── handler/                     HTTP handlers — auth, patient, stubs (501 for unimplemented)
├── service/                     Interfaces + gRPC adapters (decouples handlers from transport)
├── grpcclient/                  Connection pool for 6 backend services
├── model/                       Response envelope, error codes, pagination, JWT claims, RBAC
└── websocket/                   (placeholder — Phase 5)
proto/                           Protobuf definitions (auth, patient, common)
schemas/                         JSON schemas for request validation (not yet populated)
```

## Middleware Pipeline

Every request passes through these stages in order:

| # | Stage | File | What it does |
|---|-------|------|--------------|
| 1 | Rate Limiter | `ratelimit.go` | Per-device token bucket. 200/min reads, 60/min writes, 10/min auth |
| 2 | Request ID | `requestid.go` | UUID v4 in `X-Request-ID` header + context |
| 3 | JWT Validator | `jwtauth.go` | Ed25519 signature check, expiry, deny list. Skipped for `/auth/*` |
| 4 | RBAC Enforcer | `rbac.go` | Checks role permissions against endpoint requirements |
| 5 | Request Validator | `validator.go` | JSON schema validation for POST/PUT bodies |
| 6 | gRPC Router | (router.go) | Dispatches to handler → service → gRPC backend |
| 7 | Response Formatter | (model/envelope.go) | Wraps all responses in standard envelope |
| 8 | Audit Logger | `audit.go` | JSON structured log of every request |

Auth routes (`/api/v1/auth/*`) only get stages 1, 2, and 8. All other routes get the full pipeline.

## Backend Services

| Service | gRPC Port | Proto Defined | Gateway Adapter |
|---------|-----------|---------------|-----------------|
| Auth | :50053 | Yes | Yes |
| Patient | :50051 | Yes | Yes |
| Sync | :50052 | No | No |
| Formulary | :50054 | No | No |
| Anchor | :50055 | No | No |
| Sentinel | :50056 | No | No |

The gateway starts even if backends are down — unavailable services return 503 `SERVICE_UNAVAILABLE`.

## API Endpoints

58 REST endpoints + 1 WebSocket. Currently 7 have real handlers, 52 return 501 `NOT_IMPLEMENTED`.

**Live endpoints:**
- `POST /api/v1/auth/login` — Device authentication
- `POST /api/v1/auth/refresh` — Token refresh
- `POST /api/v1/auth/logout` — Token invalidation
- `GET  /api/v1/auth/whoami` — Current identity from JWT claims
- `GET  /api/v1/patients` — List patients (paginated, filterable)
- `GET  /api/v1/patients/:id` — Full patient bundle
- `GET  /api/v1/patients/search` — Full-text search
- `GET  /health` — Health check

All other endpoints are registered with correct middleware (JWT, RBAC, rate limiting) but return 501 until their service adapters are built.

## RBAC Roles

| Role | Read | Write | Admin |
|------|------|-------|-------|
| Community Health Worker | patient, observation | observation | — |
| Nurse | patient, encounter, medication | encounter, observation | — |
| Physician | All clinical | All clinical | conflict:resolve |
| Site Administrator | All | All | sync, anchor, supply |
| Regional Administrator | All (cross-site) | All (cross-site) | All admin |

## Error Codes

All responses use a standard envelope. Errors include a typed code:

| HTTP | Code | When |
|------|------|------|
| 400 | `VALIDATION_ERROR` | Request body fails schema validation |
| 400 | `INVALID_FHIR_RESOURCE` | FHIR resource clinically incomplete |
| 401 | `AUTH_REQUIRED` | No JWT provided |
| 401 | `TOKEN_EXPIRED` | JWT expired — use /auth/refresh |
| 401 | `TOKEN_REVOKED` | Device decommissioned |
| 403 | `INSUFFICIENT_PERMISSIONS` | Role lacks required permission |
| 403 | `SITE_SCOPE_VIOLATION` | Cross-site access denied |
| 404 | `RESOURCE_NOT_FOUND` | Resource doesn't exist |
| 409 | `MERGE_CONFLICT` | Unresolved sync conflict |
| 409 | `DUPLICATE_RESOURCE` | Identifier already exists |
| 422 | `CLINICAL_SAFETY_BLOCK` | Blocked by safety rules |
| 429 | `RATE_LIMITED` | Too many requests |
| 501 | `NOT_IMPLEMENTED` | Endpoint not yet built |
| 503 | `SERVICE_UNAVAILABLE` | Backend gRPC service down |

## Development

```bash
make build       # Build binary to bin/gateway
make run         # Build + run
make test        # Run all tests with -race
make proto-gen   # Regenerate protobuf code (requires buf)
make lint        # Run golangci-lint
```

## License

FibrinLab — Open Nucleus
