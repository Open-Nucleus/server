# Open Nucleus — Auth Service Specification V1

**Version:** 1.0  
**Date:** February 2026  
**Author:** Dr Akanimoh Osutuk — FibrinLab  
**Repo:** github.com/FibrinLab/open-nucleus  
**Service:** `services/auth/`  
**Status:** Draft — V1 Specification

---

## 1. Service Overview

### 1.1 Role

The Auth Service manages device identity, practitioner authentication, JWT token issuance, and role-based access control definitions. It is the trust anchor for the entire Open Nucleus node. All authentication is offline-capable — no external identity provider, no network dependency. Trust is established through Ed25519 cryptographic keypairs.

### 1.2 Service Identity

| Property | Value |
|----------|-------|
| Language | Go |
| gRPC Port | 50053 |
| Dependencies | `pkg/auth`, `pkg/gitstore` |
| Writes to | Git repository (`.nucleus/` config directory) |
| Reads from | Git (role definitions, device registry, revocation list) |
| Consumed by | API Gateway (token validation), all services (permission checks) |

### 1.3 Design Principles

- **Zero network dependency:** Authentication and authorisation work entirely offline using local cryptographic verification.
- **Device-based identity:** The device is the authentication unit, not a username/password. Each device holds an Ed25519 private key in secure storage.
- **Roles from Git:** Role definitions are FHIR PractitionerRole resources stored in Git, synced across nodes like any other data.
- **Revocation propagates via sync:** When a device is decommissioned, the revocation record spreads through the Git sync network.

---

## 2. gRPC Service Definition

```protobuf
syntax = "proto3";
package opennucleus.auth.v1;

import "google/protobuf/timestamp.proto";

service AuthService {
  // Device registration and authentication
  rpc RegisterDevice(RegisterDeviceRequest) returns (RegisterDeviceResponse);
  rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (AuthenticateResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc GetCurrentIdentity(GetIdentityRequest) returns (IdentityResponse);
  
  // Challenge-response flow
  rpc GetChallenge(GetChallengeRequest) returns (ChallengeResponse);
  
  // Device management
  rpc ListDevices(ListDevicesRequest) returns (ListDevicesResponse);
  rpc RevokeDevice(RevokeDeviceRequest) returns (RevokeDeviceResponse);
  rpc CheckRevocation(CheckRevocationRequest) returns (CheckRevocationResponse);
  
  // Role management
  rpc ListRoles(ListRolesRequest) returns (ListRolesResponse);
  rpc GetRole(GetRoleRequest) returns (RoleResponse);
  rpc AssignRole(AssignRoleRequest) returns (AssignRoleResponse);
  
  // Token validation (called by Gateway on every request)
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  
  // Health
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## 3. Authentication Flow

### 3.1 Device Registration (First-Time Setup)

A new device must be registered before it can authenticate. Registration is performed by a site administrator on an already-authenticated device.

```
New Device                          Auth Service
    │                                    │
    │  1. Generate Ed25519 keypair       │
    │     (stored in secure enclave)     │
    │                                    │
    │  2. RegisterDevice(public_key,     │
    │     practitioner_id, site_id)      │
    │  ──────────────────────────────▶   │
    │                                    │  3. Validate admin JWT
    │                                    │  4. Create DeviceRegistration
    │                                    │     resource in Git
    │                                    │  5. Map device to PractitionerRole
    │   ◀──────────────────────────────  │
    │  6. Registration confirmed         │
    │     device_id assigned             │
```

**Registration request:**

```json
{
  "public_key": "MCowBQYDK2VwAyEA...",
  "practitioner_id": "dr-osutuk",
  "site_id": "clinic-maiduguri-03",
  "device_name": "Tablet-Field-07",
  "requested_role": "physician"
}
```

**Registration is an admin-only operation.** The request must include a valid JWT from a device with `device:register` permission. On the very first node setup (bootstrap), a one-time setup key is used to register the first administrator device.

### 3.2 Challenge-Response Authentication

```
Device                              Auth Service
    │                                    │
    │  1. GetChallenge(device_id)        │
    │  ──────────────────────────────▶   │
    │                                    │  2. Generate random nonce
    │                                    │     (32 bytes, stored in memory
    │                                    │      with 60s TTL)
    │   ◀──────────────────────────────  │
    │  3. Receive nonce                  │
    │                                    │
    │  4. Sign nonce with Ed25519        │
    │     private key                    │
    │                                    │
    │  5. Authenticate(device_id,        │
    │     signature, practitioner_id)    │
    │  ──────────────────────────────▶   │
    │                                    │  6. Look up device public key
    │                                    │  7. Check revocation list
    │                                    │  8. Verify signature
    │                                    │  9. Look up PractitionerRole
    │                                    │ 10. Issue JWT + refresh token
    │   ◀──────────────────────────────  │
    │ 11. Receive tokens                 │
```

### 3.3 Challenge Nonce Management

- Nonces are stored in an in-memory map with 60-second TTL
- Each device can have at most one active nonce
- Nonces are single-use — consumed on successful authentication or expiry
- No persistence needed — if the service restarts, devices simply request a new challenge

---

## 4. JWT Token Specification

### 4.1 Token Structure

```json
{
  "header": {
    "alg": "EdDSA",
    "typ": "JWT",
    "kid": "node-sheffield-01"
  },
  "payload": {
    "sub": "dr-osutuk",
    "device_id": "device-tablet-field-07",
    "node_id": "node-sheffield-01",
    "site_id": "clinic-maiduguri-03",
    "role": "physician",
    "permissions": [
      "patient:read",
      "patient:write",
      "encounter:read",
      "encounter:write",
      "observation:read",
      "observation:write",
      "condition:read",
      "condition:write",
      "medication:read",
      "medication:write",
      "allergy:read",
      "allergy:write",
      "alert:read",
      "conflict:resolve",
      "formulary:read",
      "supply:read",
      "sync:read",
      "anchor:read"
    ],
    "site_scope": ["clinic-maiduguri-03"],
    "iat": 1740470400,
    "exp": 1740556800,
    "jti": "jwt-uuid-001",
    "iss": "open-nucleus-auth"
  }
}
```

### 4.2 Token Signing

- **Algorithm:** EdDSA (Ed25519)
- **Signing key:** The node's Ed25519 private key (NOT the device key — the node signs tokens)
- **Verification:** Any service can verify by reading the node's public key from `.nucleus/node.json`
- **No network call required:** The Gateway validates tokens locally using the cached node public key

### 4.3 Token Lifecycle

| Parameter | Value | Configurable |
|-----------|-------|-------------|
| Access token lifetime | 24 hours | Yes |
| Refresh token lifetime | 7 days | Yes |
| Refresh window | Final 2 hours of access token | Yes |
| Max concurrent sessions per device | 1 | No |
| Token format | JWT (RFC 7519) | No |

### 4.4 Refresh Flow

```
Device                              Auth Service
    │                                    │
    │  RefreshToken(refresh_token)       │
    │  ──────────────────────────────▶   │
    │                                    │  1. Validate refresh token signature
    │                                    │  2. Check refresh token not revoked
    │                                    │  3. Check device not revoked
    │                                    │  4. Check within refresh window
    │                                    │  5. Issue new access + refresh tokens
    │                                    │  6. Invalidate old refresh token
    │   ◀──────────────────────────────  │
    │  New tokens                        │
```

### 4.5 Token Revocation

Revoked tokens are tracked in a local deny list (in-memory set backed by a small SQLite table). The deny list contains:

- Explicitly logged-out JTIs
- All JTIs issued to revoked devices

On service restart, the deny list is rebuilt from the SQLite backing table.

---

## 5. Role-Based Access Control

### 5.1 Permission Model

Permissions follow a `resource:action` format. The full permission set:

```
# Patient data
patient:read          patient:write         patient:delete
encounter:read        encounter:write
observation:read      observation:write
condition:read        condition:write
medication:read       medication:write
allergy:read          allergy:write

# Alerts and flags
alert:read            alert:acknowledge     alert:dismiss

# Conflicts
conflict:read         conflict:resolve

# Formulary
formulary:read        formulary:write

# Supply chain
supply:read           supply:write

# Sync
sync:read             sync:trigger

# IOTA anchoring
anchor:read           anchor:trigger

# Administration
device:register       device:revoke
role:read             role:assign
```

### 5.2 Role Definitions

Roles are stored as FHIR PractitionerRole resources in `.nucleus/roles/`:

```
.nucleus/roles/
├── community-health-worker.json
├── nurse.json
├── physician.json
├── site-administrator.json
└── regional-administrator.json
```

**Community Health Worker:**

```json
{
  "resourceType": "PractitionerRole",
  "id": "role-chw",
  "code": [{ "coding": [{ "code": "chw", "display": "Community Health Worker" }] }],
  "extension": [{
    "url": "https://open-nucleus.dev/fhir/StructureDefinition/permissions",
    "valueString": "patient:read,observation:read,observation:write,alert:read,sync:read"
  }],
  "location": [{ "reference": "Location/clinic-maiduguri-03" }]
}
```

**Full role-permission matrix:**

| Permission | CHW | Nurse | Physician | Site Admin | Regional Admin |
|------------|-----|-------|-----------|------------|----------------|
| patient:read | ✓ | ✓ | ✓ | ✓ | ✓ (cross-site) |
| patient:write | — | — | ✓ | ✓ | ✓ |
| patient:delete | — | — | — | ✓ | ✓ |
| encounter:read | — | ✓ | ✓ | ✓ | ✓ |
| encounter:write | — | ✓ | ✓ | ✓ | ✓ |
| observation:read | ✓ | ✓ | ✓ | ✓ | ✓ |
| observation:write | ✓ | ✓ | ✓ | ✓ | ✓ |
| condition:read | — | ✓ | ✓ | ✓ | ✓ |
| condition:write | — | — | ✓ | ✓ | ✓ |
| medication:read | — | ✓ | ✓ | ✓ | ✓ |
| medication:write | — | — | ✓ | ✓ | ✓ |
| allergy:read | — | ✓ | ✓ | ✓ | ✓ |
| allergy:write | — | — | ✓ | ✓ | ✓ |
| alert:read | ✓ | ✓ | ✓ | ✓ | ✓ |
| alert:acknowledge | — | — | ✓ | ✓ | ✓ |
| alert:dismiss | — | — | ✓ | ✓ | ✓ |
| conflict:read | — | — | ✓ | ✓ | ✓ |
| conflict:resolve | — | — | ✓ | ✓ | ✓ |
| formulary:read | ✓ | ✓ | ✓ | ✓ | ✓ |
| formulary:write | — | — | — | ✓ | ✓ |
| supply:read | — | ✓ | ✓ | ✓ | ✓ |
| supply:write | — | — | — | ✓ | ✓ |
| sync:read | ✓ | ✓ | ✓ | ✓ | ✓ |
| sync:trigger | — | — | — | ✓ | ✓ |
| anchor:read | — | — | ✓ | ✓ | ✓ |
| anchor:trigger | — | — | — | ✓ | ✓ |
| device:register | — | — | — | ✓ | ✓ |
| device:revoke | — | — | — | ✓ | ✓ |
| role:read | — | — | — | ✓ | ✓ |
| role:assign | — | — | — | — | ✓ |

### 5.3 Site Scoping

Each role assignment includes a `site_scope` — the set of sites whose data the practitioner can access. This is encoded in the JWT:

- **CHW, Nurse, Physician:** Scoped to their assigned site only
- **Site Administrator:** Scoped to their site only
- **Regional Administrator:** Scoped to all sites in their region (multiple site IDs)

The API Gateway checks `site_scope` on every request that filters by site. A request for patient data from a site not in the JWT's `site_scope` returns `403 SITE_SCOPE_VIOLATION`.

---

## 6. Device Registry

### 6.1 Storage

Device registrations are stored in Git at `.nucleus/devices/`:

```
.nucleus/devices/
├── device-tablet-field-07.json
├── device-pi-hub-01.json
└── device-phone-chw-12.json
```

**Device registration resource:**

```json
{
  "device_id": "device-tablet-field-07",
  "device_name": "Tablet-Field-07",
  "public_key": "MCowBQYDK2VwAyEA...",
  "practitioner_id": "dr-osutuk",
  "site_id": "clinic-maiduguri-03",
  "role": "physician",
  "registered_at": "2026-02-25T09:00:00Z",
  "registered_by": "admin-device-01",
  "status": "active",
  "last_authenticated": "2026-02-25T09:42:00Z"
}
```

### 6.2 Device Revocation

When a device is lost, stolen, or decommissioned:

```json
{
  "device_id": "device-tablet-field-07",
  "status": "revoked",
  "revoked_at": "2026-02-26T14:00:00Z",
  "revoked_by": "admin-device-01",
  "reason": "Device reported lost in field"
}
```

The revocation record is committed to Git and propagates through sync. On receipt, every node:

1. Adds the device's public key to the local deny list
2. Invalidates all JTIs issued to that device
3. Rejects future authentication attempts from that device
4. Rejects future sync attempts from that device

**Revocation is permanent in V1.** A revoked device cannot be re-registered — a new keypair and device ID must be created.

---

## 7. Bootstrap (First Node Setup)

The very first device on a new Open Nucleus deployment has no existing admin to register it. The bootstrap process handles this:

```
1. Generate node Ed25519 keypair → stored in .nucleus/node.json
2. Generate bootstrap secret (32-byte random, displayed once on screen)
3. First device calls RegisterDevice with bootstrap secret instead of admin JWT
4. First device is auto-assigned "regional-administrator" role
5. Bootstrap secret is consumed and cannot be reused
6. .nucleus/bootstrap_consumed marker file committed to Git
```

After bootstrap, all subsequent device registrations require a valid admin JWT.

**Bootstrap request:**

```json
{
  "public_key": "MCowBQYDK2VwAyEA...",
  "practitioner_id": "dr-osutuk",
  "site_id": "clinic-maiduguri-03",
  "device_name": "Admin-Tablet-01",
  "bootstrap_secret": "a1b2c3d4e5f6..."
}
```

---

## 8. Node Identity

### 8.1 Node Key

Each Open Nucleus node has its own Ed25519 keypair, separate from device keys. The node key is used for:

- Signing JWTs (all tokens issued by this node are signed with the node key)
- Authenticating the node during Git sync handshakes with other nodes
- Signing IOTA anchoring payloads

Stored at `.nucleus/node.json`:

```json
{
  "node_id": "node-sheffield-01",
  "node_name": "Sheffield Development Node",
  "site_id": "clinic-maiduguri-03",
  "public_key": "MCowBQYDK2VwAyEA...",
  "created_at": "2026-02-25T08:00:00Z",
  "fhir_version": "R4",
  "nucleus_version": "1.0.0"
}
```

The private key is stored outside Git in the OS secure enclave or an encrypted keyfile at `/var/lib/open-nucleus/keys/node.key`.

### 8.2 Node Trust

When two nodes sync, they exchange node public keys. A node trusts another node if:

1. The remote node's public key is in the local `.nucleus/trusted-nodes/` directory, OR
2. The remote node's public key is signed by a trusted regional administrator device

First sync between two nodes requires manual trust establishment (admin approves the remote node's public key). After that, the trust record syncs to all other nodes in the network.

---

## 9. Encryption

### 9.1 Key Storage

| Key | Storage Location | Protection |
|-----|-----------------|------------|
| Device Ed25519 private key | Android Keystore / Linux keyring | Hardware-backed where available |
| Node Ed25519 private key | `/var/lib/open-nucleus/keys/node.key` | File-system encryption + 0600 permissions |
| Refresh token signing key | Derived from node key | — |
| Nonce secret (HMAC) | In-memory only | Lost on restart (by design) |

### 9.2 Secure Storage Interface

```go
type KeyStore interface {
    // Store a private key
    StoreKey(keyID string, privateKey ed25519.PrivateKey) error
    
    // Retrieve a private key
    LoadKey(keyID string) (ed25519.PrivateKey, error)
    
    // Delete a private key
    DeleteKey(keyID string) error
    
    // Check if key exists
    HasKey(keyID string) bool
}
```

Implementations:
- `AndroidKeyStore` — uses Android Keystore API via platform channel
- `LinuxKeyStore` — encrypted file in `/var/lib/open-nucleus/keys/`
- `MemoryKeyStore` — for testing only

---

## 10. Gateway Integration

### 10.1 ValidateToken RPC

Called by the API Gateway on every authenticated request. Must be extremely fast.

```protobuf
message ValidateTokenRequest {
  string token = 1;  // Raw JWT string
}

message ValidateTokenResponse {
  bool valid = 1;
  string sub = 2;                    // Practitioner ID
  string device_id = 3;
  string node_id = 4;
  string site_id = 5;
  string role = 6;
  repeated string permissions = 7;
  repeated string site_scope = 8;
  string error_code = 9;            // Empty if valid
}
```

**Validation steps:**
1. Parse JWT header and payload
2. Verify EdDSA signature against node public key
3. Check `exp` not passed
4. Check `jti` not in deny list
5. Check `device_id` not in revocation list
6. Return parsed claims

**Performance target:** < 1ms. This is called on every single request. The node public key and deny list are cached in memory.

### 10.2 CheckPermission RPC

Optional fine-grained check, used when the Gateway needs to verify a specific permission beyond what's in the JWT:

```protobuf
message CheckPermissionRequest {
  string token = 1;
  string required_permission = 2;   // e.g. "patient:write"
  string target_site_id = 3;        // Site being accessed
}

message CheckPermissionResponse {
  bool allowed = 1;
  string reason = 2;                // Why denied, if denied
}
```

---

## 11. Error Handling

| gRPC Code | Condition | Notes |
|-----------|-----------|-------|
| `UNAUTHENTICATED` | Invalid signature, expired token, revoked device | Maps to HTTP 401 |
| `PERMISSION_DENIED` | Valid token but insufficient permissions | Maps to HTTP 403 |
| `NOT_FOUND` | Device ID not registered | — |
| `ALREADY_EXISTS` | Device ID already registered | — |
| `INVALID_ARGUMENT` | Malformed public key, invalid role | — |
| `FAILED_PRECONDITION` | Bootstrap already consumed, nonce expired | — |
| `RESOURCE_EXHAUSTED` | Too many failed auth attempts (10/min per device) | Brute-force protection |

### 11.1 Brute-Force Protection

Failed authentication attempts are rate-limited per device_id:

- **10 failed attempts per minute** → subsequent attempts return `RESOURCE_EXHAUSTED`
- Counter resets after 1 minute of no attempts
- Counter is in-memory only (resets on service restart — acceptable for V1)
- Successful authentication resets the counter

---

## 12. Configuration

```yaml
auth_service:
  grpc_port: 50053
  
  jwt:
    issuer: "open-nucleus-auth"
    access_token_lifetime: 24h
    refresh_token_lifetime: 7d
    refresh_window: 2h
    
  node:
    config_path: /var/lib/open-nucleus/data/.nucleus/node.json
    private_key_path: /var/lib/open-nucleus/keys/node.key
    
  devices:
    registry_path: /var/lib/open-nucleus/data/.nucleus/devices/
    revocation_check_interval: 30s
    
  roles:
    definitions_path: /var/lib/open-nucleus/data/.nucleus/roles/
    
  security:
    nonce_ttl: 60s
    max_failed_attempts: 10
    failed_attempt_window: 60s
    deny_list_db: /var/lib/open-nucleus/auth-deny.db
    
  keystore:
    type: linux                     # linux, android, memory
    path: /var/lib/open-nucleus/keys/
    
  logging:
    level: info
    format: json
```

---

## 13. Testing Strategy

### 13.1 Unit Tests

| Area | Coverage | Focus |
|------|----------|-------|
| JWT signing/verification | 100% | Ed25519 roundtrip, expiry, claims extraction |
| Challenge-response | 95% | Nonce lifecycle, signature verification, replay prevention |
| Permission checking | 100% | Every role × every permission combination |
| Site scoping | 100% | Cross-site access denial, multi-site regional admin access |
| Token revocation | 95% | JTI deny list, device revocation cascade |

### 13.2 Integration Tests

| Test | Description |
|------|-------------|
| Bootstrap flow | Fresh node → bootstrap → first admin device registered |
| Full auth cycle | Register device → challenge → authenticate → validate → refresh → logout |
| Revocation propagation | Revoke device → verify all tokens invalidated → verify auth rejected |
| Role assignment | Admin assigns role → device re-authenticates → verify new permissions |
| Concurrent validation | 1000 concurrent ValidateToken calls → verify < 1ms p99 |

### 13.3 Security Tests

| Test | Description |
|------|-------------|
| Replay attack | Reuse a consumed nonce → verify rejection |
| Forged signature | Sign with wrong key → verify rejection |
| Expired token | Use token past expiry → verify rejection |
| Tampered claims | Modify JWT payload → verify signature fails |
| Revoked device auth | Authenticate as revoked device → verify rejection |
| Brute force | 11 rapid failed attempts → verify rate limiting |

---

## 14. Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| GetChallenge | < 5ms | Nonce generation + memory store |
| Authenticate | < 50ms | Signature verification + JWT signing |
| ValidateToken | < 1ms | In-memory key + deny list check |
| CheckPermission | < 1ms | Array lookup on JWT claims |
| RefreshToken | < 50ms | Verify + re-sign |
| Device registration | < 200ms | Git commit for device record |
| Device revocation | < 200ms | Git commit + deny list update |
| Memory footprint | < 30MB RSS | Primarily deny list + nonce map |

---

*Open Nucleus • Auth Service Specification V1 • FibrinLab*  
*github.com/FibrinLab/open-nucleus*
