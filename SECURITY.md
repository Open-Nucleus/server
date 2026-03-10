# Security

Open Nucleus security model for offline-first deployment in austere environments.

## Encryption at Rest

### Per-Patient Envelope Encryption

Every patient's clinical data is encrypted with a unique data encryption key (DEK) before being written to Git.

```
Master Key (AES-256, env var or file)
    |
    +--- wraps ---> Patient DEK (AES-256-GCM, random 32 bytes)
    |                   |
    |                   +--- encrypts ---> Patient's FHIR JSON files in Git
    |
    +--- wraps ---> System DEK
                        |
                        +--- encrypts ---> Non-patient resources (Practitioner, Organization, Location)
```

**Key storage:** Wrapped DEKs are stored in Git at `.nucleus/keys/{patient_id}.key`, so they sync with the data they protect. Unwrapped keys exist only in memory during the process lifetime.

**Key format:** `[1-byte version][wrapped-key-bytes]` — the version byte enables future key rotation without breaking existing data.

### What is encrypted

| Data | Encrypted | Notes |
|------|-----------|-------|
| FHIR JSON files in Git | Yes | AES-256-GCM per patient |
| SQLite search index | No | Contains extracted search fields only (name, DOB, gender, clinical codes) |
| Git commit metadata | No | Timestamps and commit messages visible |
| Git file paths | No | Contain resource type and patient UUID |
| Wrapped encryption keys | Yes | AES-KW wrapped with master key |

**Recommendation:** Deploy with disk-level encryption (LUKS on Linux, FileVault on macOS) for defense in depth. This protects SQLite index fields, Git metadata, and file paths.

### Master Key Management

The master key is loaded from:
1. `NUCLEUS_MASTER_KEY` environment variable (hex-encoded 32 bytes), or
2. File path specified in `config.yaml` at `encryption.master_key_file`

**If the master key is lost, all patient data is permanently unreadable.** Back up the master key separately from the data.

## Transport Security

### TLS for HTTP

The monolith supports three TLS modes via `config.yaml`:

| Mode | Behavior |
|------|----------|
| `auto` | Auto-generates a self-signed Ed25519 certificate on first start. Stored in `data/certs/`. |
| `provided` | Uses user-supplied PEM certificate and key files. |
| `off` | Plain HTTP. Only for development or when TLS is terminated upstream. |

### Node-to-Node Sync

Inter-node sync uses ECDH key exchange (X25519 derived from Ed25519 node keys) with AES-256-GCM for payload encryption. Each node's identity key serves dual purpose: JWT signing and sync transport encryption.

## Authentication

### Ed25519 Challenge-Response

Devices register with an Ed25519 public key. Authentication is a two-step challenge-response:

1. **Challenge:** Server generates a random nonce (60s TTL, single-use).
2. **Authenticate:** Device signs the nonce with its private key. Server verifies signature, issues JWT.

No passwords are stored or transmitted. The Ed25519 keypair is the device's identity.

### JWT Tokens

- **Algorithm:** EdDSA (Ed25519)
- **Access token lifetime:** 24 hours (configurable)
- **Refresh window:** 2 hours (configurable)
- **Offline-verifiable:** Any node with the issuer's public key can verify tokens without network access
- **Deny list:** SQLite-backed revocation list for logout and device decommission

### SMART on FHIR

OAuth2 authorization code flow with PKCE for third-party app integration:

- SMART v2 scopes (patient/\*.read, user/\*.write, etc.)
- EHR launch and standalone launch
- Client registration and management
- All flows execute locally — no cloud identity provider required

### Brute-Force Protection

- Per-device failure counter with configurable window (default: 10 failures / 60s)
- Nonce single-use prevents replay attacks
- Rate limiting at the HTTP layer (10 auth attempts/min)

## Authorization

5-role RBAC model enforced at middleware layer:

| Role | Scope |
|------|-------|
| Community Health Worker | Read/write own patient observations |
| Nurse | Read/write encounters, observations |
| Physician | All clinical data, conflict resolution |
| Site Administrator | All data + sync, anchor, supply admin |
| Regional Administrator | Cross-site access, all admin functions |

SMART v2 scopes provide additional fine-grained access control on FHIR endpoints, restricting third-party apps to specific resource types and patient contexts.

## Threat Model

### In Scope

| Threat | Mitigation |
|--------|-----------|
| Device theft/loss | Encryption at rest, short JWT lifetime |
| Unauthorized access | Ed25519 auth, RBAC, rate limiting |
| Data tampering | Git integrity (SHA-1 hash chain), Merkle anchoring |
| Eavesdropping on sync | ECDH + AES-256-GCM transport encryption |
| Patient data breach | Per-patient encryption, crypto-erasure |
| Replay attacks | Single-use nonces, JWT expiry |

### Out of Scope (V1)

| Threat | Status |
|--------|--------|
| Side-channel attacks | Not addressed |
| Physical key extraction (cold boot) | Use hardware security module in production |
| Compromised master key | Document backup procedures; HSM support planned |
| Traffic analysis | File paths and Git metadata reveal resource types |
| Malicious node in sync mesh | Planned: mutual TLS with certificate pinning |

## Vulnerability Reporting

Report security vulnerabilities to the maintainers via GitHub Security Advisories on the repository. Do not open public issues for security bugs.
