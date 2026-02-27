# Open Nucleus — Sync Service Specification V1

**Version:** 1.0  
**Date:** February 2026  
**Author:** Dr Akanimoh Osutuk — FibrinLab  
**Repo:** github.com/FibrinLab/open-nucleus  
**Service:** `services/sync/`  
**Status:** Draft — V1 Specification

---

## 1. Service Overview

### 1.1 Role

The Sync Service is the core differentiator of Open Nucleus. It manages node discovery, transport negotiation, Git-based data synchronisation between nodes, and FHIR-aware conflict resolution. It is the only service (besides Patient Service) that writes to the Git repository — specifically merge commits from incoming sync data. After every successful merge, it emits a `SyncCompleted` event consumed by the Sentinel Agent and triggers re-indexing via the Patient Service.

### 1.2 Service Identity

| Property | Value |
|----------|-------|
| Language | Go |
| gRPC Port | 50052 |
| Additional Ports | Dynamic per transport adapter |
| Dependencies | `pkg/gitstore`, `pkg/merge`, `pkg/fhir`, `pkg/auth` |
| Writes to | Git repository (merge commits only) |
| Reads from | Git (HEAD, refs, pack data) |
| Emits | `SyncCompleted`, `SyncFailed`, `PeerDiscovered`, `PeerLost`, `ConflictDetected` events |
| Consumed by | API Gateway (status, trigger), Patient Service (re-index), Sentinel Agent (analysis trigger) |

### 1.3 Design Principles

- **Transport-agnostic sync:** Git operations are completely decoupled from how bytes move between nodes. Adding a new transport requires implementing one interface.
- **Automatic and invisible:** Sync happens in the background. Clinicians never initiate or manage sync — they just see updated data.
- **Safety over speed:** The FHIR-aware merge driver never silently resolves a clinically dangerous conflict. When in doubt, it blocks and escalates.
- **Bandwidth-aware:** On constrained transports, critical data syncs first.

---

## 2. gRPC Service Definition

```protobuf
syntax = "proto3";
package opennucleus.sync.v1;

import "google/protobuf/timestamp.proto";

service SyncService {
  // Sync status and control
  rpc GetStatus(GetStatusRequest) returns (SyncStatusResponse);
  rpc TriggerSync(TriggerSyncRequest) returns (TriggerSyncResponse);
  rpc CancelSync(CancelSyncRequest) returns (CancelSyncResponse);
  
  // Peer management
  rpc ListPeers(ListPeersRequest) returns (ListPeersResponse);
  rpc TrustPeer(TrustPeerRequest) returns (TrustPeerResponse);
  rpc UntrustPeer(UntrustPeerRequest) returns (UntrustPeerResponse);
  
  // Conflict management
  rpc ListConflicts(ListConflictsRequest) returns (ListConflictsResponse);
  rpc GetConflict(GetConflictRequest) returns (ConflictDetailResponse);
  rpc ResolveConflict(ResolveConflictRequest) returns (ResolveConflictResponse);
  rpc DeferConflict(DeferConflictRequest) returns (DeferConflictResponse);
  
  // Bundle import/export (USB fallback)
  rpc ExportBundle(ExportBundleRequest) returns (ExportBundleResponse);
  rpc ImportBundle(ImportBundleRequest) returns (ImportBundleResponse);
  
  // Sync history
  rpc GetSyncHistory(GetSyncHistoryRequest) returns (SyncHistoryResponse);
  
  // Transport management
  rpc ListTransports(ListTransportsRequest) returns (ListTransportsResponse);
  rpc EnableTransport(EnableTransportRequest) returns (TransportResponse);
  rpc DisableTransport(DisableTransportRequest) returns (TransportResponse);
  
  // Event subscription (for Sentinel and Gateway WebSocket)
  rpc SubscribeEvents(SubscribeEventsRequest) returns (stream SyncEvent);
  
  // Node-to-node sync protocol (called by remote nodes)
  rpc Handshake(HandshakeRequest) returns (HandshakeResponse);
  rpc RequestPack(RequestPackRequest) returns (stream PackChunk);
  rpc SendPack(stream PackChunk) returns (SendPackResponse);
  
  // Health
  rpc Health(HealthRequest) returns (HealthResponse);
}
```

---

## 3. Transport Adapter Interface

### 3.1 Interface Definition

Every transport adapter must implement this Go interface:

```go
package transport

import (
    "context"
    "io"
)

type Adapter interface {
    // Metadata
    Name() string                              // "wifi-direct", "bluetooth", "local-network", "usb"
    Capabilities() Capabilities

    // Lifecycle
    Start(ctx context.Context) error           // Begin scanning/listening
    Stop() error                               // Shutdown gracefully
    
    // Discovery
    Discover(ctx context.Context) (<-chan PeerNode, error)  // Stream of discovered peers
    
    // Connection
    Connect(ctx context.Context, peer PeerNode) (SyncConn, error)
    
    // Accept incoming connections (for peers that discover us)
    Accept(ctx context.Context) (<-chan SyncConn, error)
}

type SyncConn interface {
    io.ReadWriteCloser
    PeerInfo() PeerNode
    Transport() string
    EstimatedBandwidth() int64    // bytes/sec, 0 if unknown
}

type Capabilities struct {
    MaxBandwidth    int64         // bytes/sec theoretical max
    Latency         time.Duration // typical RTT
    Reliability     float64       // 0.0-1.0
    AutoDiscovery   bool          // Can find peers automatically
    Bidirectional   bool          // Both sides can initiate
    MaxPeers        int           // Concurrent connections
}

type PeerNode struct {
    NodeID     string
    NodeName   string
    PublicKey  ed25519.PublicKey
    Transport  string
    Address    string             // Transport-specific address
    LastSeen   time.Time
    Metadata   map[string]string  // Transport-specific metadata (RSSI, etc.)
}
```

### 3.2 V1 Transport Implementations

**Wi-Fi Direct Adapter** (`transports/wifidirect/`)

| Property | Value |
|----------|-------|
| Discovery | mDNS service advertisement (`_opennucleus._tcp`) |
| Connection | TCP over Wi-Fi Direct group |
| Encryption | TLS 1.3 with mutual Ed25519 authentication |
| Auto-discovery | Yes |
| Bandwidth | ~50 MB/s |
| Platform | Android (Wi-Fi P2P API), Linux (wpa_supplicant) |

**Local Network Adapter** (`transports/localnet/`)

| Property | Value |
|----------|-------|
| Discovery | mDNS service advertisement (`_opennucleus._tcp`) |
| Connection | TCP over existing LAN |
| Encryption | TLS 1.3 |
| Auto-discovery | Yes |
| Bandwidth | ~100 MB/s (LAN dependent) |
| Platform | All |

**Bluetooth Adapter** (`transports/bluetooth/`)

| Property | Value |
|----------|-------|
| Discovery | BLE advertisement with service UUID |
| Connection | Bluetooth RFCOMM / BLE GATT |
| Encryption | Noise Protocol Framework (NNpsk0) |
| Auto-discovery | Yes |
| Bandwidth | ~2 MB/s (BLE 5.0) |
| Platform | Android (Bluetooth API), Linux (BlueZ) |

**USB Bundle Adapter** (`transports/usb/`)

| Property | Value |
|----------|-------|
| Discovery | File system watch on `/media/` mount points |
| Connection | N/A (file-based) |
| Encryption | AES-256-GCM with deployment key |
| Auto-discovery | No (manual export/import) |
| Bandwidth | Batch (limited by storage I/O) |
| Platform | All |

### 3.3 Transport Priority and Selection

When multiple transports are available for the same peer, the Sync Service selects based on a priority score:

```go
func (s *SyncService) selectTransport(peer PeerNode, available []TransportOption) TransportOption {
    // Score = (bandwidth * 0.4) + (reliability * 0.3) + (1/latency * 0.2) + (autoDiscovery * 0.1)
    // Higher score wins
}
```

**Default priority order:**

| Priority | Transport | Score Rationale |
|----------|-----------|-----------------|
| 1 | Local Network | Highest bandwidth + reliability |
| 2 | Wi-Fi Direct | High bandwidth, no infrastructure needed |
| 3 | Bluetooth | Lower bandwidth but works everywhere |
| 4 | USB Bundle | Manual fallback only |

If a sync fails on the selected transport, the service automatically falls back to the next available transport.

---

## 4. Sync Protocol

### 4.1 Five-Step Handshake

```
Node A (initiator)                    Node B (responder)
    │                                      │
    │  ── 1. HANDSHAKE ──────────────────▶ │
    │     { node_id, public_key,           │
    │       git_head, commit_count }       │
    │                                      │
    │  ◀── 2. HANDSHAKE_ACK ───────────── │
    │     { node_id, public_key,           │
    │       git_head, commit_count,        │
    │       trusted: bool }                │
    │                                      │
    │  Compare HEADs                       │
    │  If identical → DONE (no sync)       │
    │  If different → continue             │
    │                                      │
    │  ── 3. REQUEST_PACK ───────────────▶ │
    │     { want: [commits B has],         │
    │       have: [commits A has] }        │
    │                                      │
    │  ◀── 4. PACK_DATA (stream) ──────── │
    │     { git pack file chunks }         │
    │                                      │
    │  Apply pack, run merge               │
    │                                      │
    │  ── 5. SYNC_ACK ──────────────────▶  │
    │     { new_head, conflicts: [],       │
    │       resources_merged: N }          │
    │                                      │
    │  (Node B then requests A's           │
    │   new commits — bidirectional)       │
```

### 4.2 Handshake Authentication

The handshake includes mutual authentication:

1. Node A signs the handshake payload with its node private key
2. Node B verifies against A's public key
3. Node B checks if A's public key is in its trusted nodes list
4. If NOT trusted: sync is rejected unless manually approved
5. If the node's device has been revoked: sync is rejected

### 4.3 Pack Generation

The Sync Service uses libgit2 to generate minimal pack files:

```go
func generatePack(repo *git.Repository, wants []git.Oid, haves []git.Oid) (io.Reader, error) {
    // 1. Compute missing commits: wants minus haves
    // 2. Walk commit graph to find minimal set of objects
    // 3. Generate Git pack file with delta compression
    // 4. Return as streaming reader
}
```

Pack files are streamed in chunks (default 64KB) to avoid buffering the entire pack in memory.

### 4.4 Bidirectional Sync

After Node A receives and merges Node B's data, the roles reverse:

- Node B requests Node A's new commits (including any that were created locally since the handshake, plus the merge commit itself)
- This ensures both nodes end up with the same data after a single sync session

### 4.5 Bandwidth-Aware Priority Sync

On constrained transports (Bluetooth, satellite), the Sync Service prioritises which records to sync first:

**Priority tiers:**

| Tier | Content | Rationale |
|------|---------|-----------|
| 1 (Critical) | Active alerts (DetectedIssue, Flag), device revocations | Safety-critical, small payload |
| 2 (High) | Active encounters, new patients, medication requests | Current clinical workflow |
| 3 (Normal) | Observations, conditions, allergy updates | Important but not urgent |
| 4 (Low) | Closed encounters, resolved conditions, supply data | Historical/operational |
| 5 (Bulk) | Version history, old encounters | Only on high-bandwidth transport |

Implementation: the pack generator walks the commit graph and sorts objects by resource type priority. On a constrained transport, it stops after Tier 2 if bandwidth is exhausted. Remaining tiers sync on the next opportunity.

```go
type SyncPriority int

const (
    PriorityCritical SyncPriority = iota
    PriorityHigh
    PriorityNormal
    PriorityLow
    PriorityBulk
)

func classifyResource(path string, fhirJSON []byte) SyncPriority {
    // Parse resource type from path
    // Check status fields for active vs closed
    // Return priority tier
}
```

---

## 5. FHIR-Aware Merge Driver

### 5.1 Overview

The merge driver is the most critical component of the Sync Service. It replaces Git's default text-based merge with clinical-context-aware conflict resolution. It lives in `pkg/merge/` and is registered as a custom Git merge driver via libgit2.

### 5.2 Merge Classification

For every file that has concurrent modifications on both sides of a merge, the driver classifies the conflict:

```go
type ConflictLevel int

const (
    AutoMerge ConflictLevel = iota  // Safe to merge automatically
    Review                           // Merge but flag for clinician review
    Block                            // Do not merge, require explicit resolution
)

type ConflictResult struct {
    Level         ConflictLevel
    ResourceType  string
    ResourceID    string
    PatientID     string
    LocalVersion  []byte           // FHIR JSON
    RemoteVersion []byte           // FHIR JSON
    MergedVersion []byte           // nil if Block
    ChangedFields []string         // JSON paths that differ
    Reason        string           // Human-readable explanation
}
```

### 5.3 Classification Rules

**Auto-Merge (safe — merge automatically, log only):**

| Scenario | Example |
|----------|---------|
| Different resource types for same patient | Site A adds Observation, Site B adds MedicationRequest |
| Same resource type, different resources | Site A adds Observation-001, Site B adds Observation-002 |
| Non-overlapping field changes in same resource | Site A updates `address`, Site B updates `telecom` |
| Additive array changes | Site A adds an encounter, Site B adds a different encounter |
| Status progression (forward only) | `planned` → `in-progress` (both sides agree on direction) |

**Review (merge both, flag for clinician review):**

| Scenario | Example |
|----------|---------|
| Same resource, overlapping non-clinical fields | Both sites update patient address to different values |
| Same resource, different non-critical values | Both sites record different contact numbers |
| Encounter timing overlap | Site A and Site B both record encounters with overlapping periods |
| Status disagreement (non-dangerous) | One site marks condition `active`, other marks `remission` |

**Block (reject merge for this resource, queue for resolution):**

| Scenario | Example |
|----------|---------|
| Conflicting allergy: substance or criticality | Site A: penicillin/high, Site B: penicillin/low |
| Conflicting active medications + interaction | Site A prescribes Drug X, Site B prescribes Drug Y, X+Y = major interaction |
| Conflicting diagnosis on same condition | Site A: confirmed malaria, Site B: refuted malaria |
| Patient identity conflict | Different birth dates, different gender values |
| Contradictory vital signs at same timestamp | Same observation code, same effectiveDateTime, different values |

### 5.4 Merge Algorithm

```
For each conflicting file in the Git merge:
    │
    ├─ 1. Parse both versions as FHIR JSON
    │
    ├─ 2. Identify resource type
    │
    ├─ 3. Compute field-level diff (JSON path comparison)
    │
    ├─ 4. Check BLOCK rules first (highest priority):
    │     If any block rule matches → ConflictLevel = Block
    │     Store both versions, do NOT merge this file
    │
    ├─ 5. Check REVIEW rules:
    │     If any review rule matches → ConflictLevel = Review
    │     Merge using field-level strategy (latest timestamp wins for non-clinical fields)
    │     Flag for clinician review
    │
    ├─ 6. Default → AutoMerge:
    │     Merge using standard JSON merge (non-overlapping changes combine)
    │     Log the auto-resolution
    │
    └─ 7. Return ConflictResult
```

### 5.5 Field-Level Merge Strategy (Review Level)

When a resource has overlapping changes classified as Review, the merge uses this strategy:

```go
type FieldMergeStrategy int

const (
    LatestTimestamp FieldMergeStrategy = iota  // Last write wins based on meta.lastUpdated
    KeepBoth                                    // For array fields, keep both entries
    PreferLocal                                 // Default for unresolved fields
)

var fieldStrategies = map[string]FieldMergeStrategy{
    "Patient.name":              LatestTimestamp,
    "Patient.address":           LatestTimestamp,
    "Patient.telecom":           KeepBoth,
    "Encounter.participant":     KeepBoth,
    "Observation.note":          KeepBoth,
    "Condition.note":            KeepBoth,
    "MedicationRequest.note":    KeepBoth,
    // Everything else: PreferLocal
}
```

### 5.6 Drug Interaction Check During Merge

When the merge introduces new MedicationRequests from a remote site, the merge driver calls the Formulary Service to check for interactions against the local patient's active medications:

```
Remote site added MedicationRequest for Patient X
    │
    ├─ Fetch all active MedicationRequests for Patient X (local)
    ├─ Check new medication against existing via Formulary Service
    │
    ├─ No interaction → AutoMerge
    ├─ Minor interaction → Review (merge + flag)
    └─ Major interaction → Block (do not merge, queue for resolution)
```

---

## 6. Conflict Storage and Resolution

### 6.1 Conflict Storage

Unresolved conflicts are stored in a local SQLite table (separate from the main index):

```sql
CREATE TABLE conflicts (
    id TEXT PRIMARY KEY,
    level TEXT NOT NULL,              -- "review" or "block"
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    patient_id TEXT,
    local_commit TEXT NOT NULL,       -- Git commit hash
    remote_commit TEXT NOT NULL,
    local_version TEXT NOT NULL,      -- FHIR JSON
    remote_version TEXT NOT NULL,     -- FHIR JSON
    merged_version TEXT,              -- FHIR JSON (null for block)
    changed_fields TEXT NOT NULL,     -- JSON array of paths
    reason TEXT NOT NULL,
    peer_node_id TEXT NOT NULL,
    peer_site_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    resolved_at TEXT,
    resolved_by TEXT,
    resolution TEXT,                  -- "accept_local", "accept_incoming", "merge_custom"
    resolution_reason TEXT,
    status TEXT NOT NULL DEFAULT 'open'  -- "open", "resolved", "deferred"
);

CREATE INDEX idx_conflict_status ON conflicts(status);
CREATE INDEX idx_conflict_patient ON conflicts(patient_id);
CREATE INDEX idx_conflict_level ON conflicts(level);
```

### 6.2 Conflict Resolution Flow

```
Clinician sees conflict in Flutter UI
    │
    ├─ GET /api/v1/conflicts/:id (via Gateway → Sync Service)
    │   Returns: local version, remote version, diff, context
    │
    ├─ Clinician chooses resolution:
    │   ├─ "accept_local"     → Keep local version
    │   ├─ "accept_incoming"  → Accept remote version
    │   └─ "merge_custom"     → Provide hand-merged resource
    │
    ├─ POST /api/v1/conflicts/:id/resolve
    │   │
    │   ├─ Sync Service validates the chosen/custom resource
    │   ├─ Writes resolved resource via Patient Service (Git commit + SQLite index)
    │   ├─ Marks conflict as resolved in conflicts table
    │   ├─ Emits ConflictResolved event
    │   └─ Resolution propagates to other nodes on next sync
    │
    └─ Or: POST /api/v1/conflicts/:id/defer
        ├─ Conflict remains open
        └─ Propagates as unresolved to other nodes
```

### 6.3 Conflict Propagation

Unresolved conflicts are committed to Git as metadata files at `.nucleus/conflicts/{conflict-id}.json`. When nodes sync:

- If Node A has an unresolved conflict for Resource X, Node B receives the conflict metadata
- Node B knows not to auto-merge Resource X until the conflict is resolved
- Any node with the appropriate permissions can resolve the conflict
- The resolution (not just the result, but the full audit trail including who resolved it and why) syncs to all nodes

---

## 7. Event System

### 7.1 Event Bus

The Sync Service runs a local event bus for internal communication. Events are published to in-memory channels consumed by:

- **API Gateway** (via `SubscribeEvents` gRPC stream → WebSocket to Flutter)
- **Sentinel Agent** (via `SubscribeEvents` gRPC stream → trigger analysis)
- **Anchor Service** (via `SubscribeEvents` → trigger anchoring)

```go
type EventBus interface {
    Publish(event SyncEvent)
    Subscribe(filter EventFilter) <-chan SyncEvent
    Unsubscribe(ch <-chan SyncEvent)
}

type SyncEvent struct {
    Type      EventType
    Timestamp time.Time
    Payload   interface{}       // Type-specific payload
}

type EventType string

const (
    EventSyncStarted      EventType = "sync.started"
    EventSyncCompleted    EventType = "sync.completed"
    EventSyncFailed       EventType = "sync.failed"
    EventPeerDiscovered   EventType = "sync.peer_discovered"
    EventPeerLost         EventType = "sync.peer_lost"
    EventConflictDetected EventType = "conflict.new"
    EventConflictResolved EventType = "conflict.resolved"
)
```

### 7.2 SyncCompleted Event Payload

This is the critical event that triggers Sentinel Agent analysis and SQLite re-indexing:

```go
type SyncCompletedPayload struct {
    SyncID          string
    PeerNodeID      string
    PeerSiteID      string
    Transport       string
    Duration        time.Duration
    RecordsReceived int
    RecordsSent     int
    ConflictsFound  int
    
    // Delta information for Sentinel and re-indexing
    NewResources    []ResourceDelta    // Resources added by this sync
    ModifiedResources []ResourceDelta  // Resources modified by merge
    
    NewGitHead      string             // Current HEAD after merge
    PreviousGitHead string             // HEAD before sync started
}

type ResourceDelta struct {
    Path         string    // Git path: patients/{id}/observations/{id}.json
    ResourceType string    // "Observation", "Encounter", etc.
    ResourceID   string
    PatientID    string    // Extracted from path
    Operation    string    // "added", "modified"
    SiteOrigin   string    // Which site created this resource
}
```

---

## 8. Discovery and Connection Lifecycle

### 8.1 Discovery Loop

```go
func (s *SyncService) discoveryLoop(ctx context.Context) {
    for _, adapter := range s.transports {
        peers, _ := adapter.Discover(ctx)
        go func(peers <-chan PeerNode) {
            for peer := range peers {
                s.handleDiscoveredPeer(peer)
            }
        }(peers)
    }
}

func (s *SyncService) handleDiscoveredPeer(peer PeerNode) {
    // 1. Check if peer is trusted
    // 2. Check if peer is revoked
    // 3. Check if we've synced recently (cooldown: 5 min default)
    // 4. If sync needed → add to sync queue
    // 5. Emit PeerDiscovered event
}
```

### 8.2 Sync Queue

Discovered peers that need syncing are added to a priority queue:

```go
type SyncQueue struct {
    mu    sync.Mutex
    items []SyncJob
}

type SyncJob struct {
    Peer      PeerNode
    Priority  int              // Based on transport quality + time since last sync
    Transport transport.Adapter
    Attempts  int              // Retry count
    NextRetry time.Time
}
```

A worker goroutine processes the queue, running one sync at a time (V1 — no concurrent syncs to avoid Git conflicts):

```go
func (s *SyncService) syncWorker(ctx context.Context) {
    for {
        job := s.queue.Pop()     // Blocks until job available
        err := s.executeSync(ctx, job)
        if err != nil {
            job.Attempts++
            if job.Attempts < 3 {
                job.NextRetry = time.Now().Add(backoff(job.Attempts))
                s.queue.Push(job)  // Re-queue with backoff
            } else {
                s.emitEvent(EventSyncFailed, ...)
            }
        }
    }
}
```

### 8.3 Sync Cooldown

After a successful sync with a peer, a cooldown period prevents redundant re-syncing:

| Transport | Default Cooldown |
|-----------|-----------------|
| Local Network | 2 minutes |
| Wi-Fi Direct | 5 minutes |
| Bluetooth | 10 minutes |
| USB | No cooldown (manual trigger) |

Cooldown is bypassed if:
- A manual sync trigger is issued via the API
- A critical alert (device revocation) needs propagation
- The peer has a significantly different HEAD (detected during discovery metadata exchange)

---

## 9. Bundle Import/Export (USB Fallback)

### 9.1 Export

```go
func (s *SyncService) ExportBundle(req *ExportBundleRequest) (*ExportBundleResponse, error) {
    // 1. Determine commit range (since_commit → HEAD, or full if no since_commit)
    // 2. Generate Git bundle file using libgit2
    // 3. Encrypt bundle with deployment AES-256-GCM key
    // 4. Write to specified output path
    // 5. Return: path, commit range, size, checksum
}
```

**Bundle file format:**

```
[8 bytes]   Magic: "ONBUNDLE"
[4 bytes]   Version: 1
[32 bytes]  SHA-256 of encrypted payload
[4 bytes]   Metadata length
[N bytes]   Metadata JSON (node_id, from_commit, to_commit, created_at)
[...]       AES-256-GCM encrypted Git bundle
```

### 9.2 Import

```go
func (s *SyncService) ImportBundle(req *ImportBundleRequest) (*ImportBundleResponse, error) {
    // 1. Read bundle file, verify magic and checksum
    // 2. Decrypt with deployment key
    // 3. Parse metadata, verify source node is trusted
    // 4. Apply Git bundle (equivalent to git fetch)
    // 5. Run FHIR-aware merge
    // 6. Re-index new resources via Patient Service
    // 7. Emit SyncCompleted event
    // 8. Return: resources merged, conflicts found
}
```

---

## 10. Sync Logging and History

### 10.1 Sync History Table

```sql
CREATE TABLE sync_history (
    id TEXT PRIMARY KEY,
    peer_node_id TEXT NOT NULL,
    peer_site_id TEXT,
    transport TEXT NOT NULL,
    direction TEXT NOT NULL,          -- "bidirectional", "inbound", "outbound"
    started_at TEXT NOT NULL,
    completed_at TEXT,
    duration_ms INTEGER,
    status TEXT NOT NULL,             -- "completed", "failed", "partial"
    records_received INTEGER DEFAULT 0,
    records_sent INTEGER DEFAULT 0,
    conflicts_found INTEGER DEFAULT 0,
    conflicts_auto_resolved INTEGER DEFAULT 0,
    bytes_transferred INTEGER DEFAULT 0,
    local_head_before TEXT,
    local_head_after TEXT,
    remote_head TEXT,
    error_message TEXT,
    metadata TEXT                     -- JSON: transport-specific details
);

CREATE INDEX idx_sync_peer ON sync_history(peer_node_id);
CREATE INDEX idx_sync_date ON sync_history(started_at);
CREATE INDEX idx_sync_status ON sync_history(status);
```

### 10.2 Peer State Table

```sql
CREATE TABLE peer_state (
    node_id TEXT PRIMARY KEY,
    node_name TEXT,
    public_key TEXT NOT NULL,
    last_seen TEXT,
    last_synced TEXT,
    last_sync_transport TEXT,
    their_head TEXT,                  -- Last known HEAD of this peer
    trusted INTEGER DEFAULT 0,        -- 0 = untrusted, 1 = trusted
    revoked INTEGER DEFAULT 0,
    discovered_via TEXT,              -- Transport that last discovered this peer
    metadata TEXT                     -- JSON
);
```

---

## 11. Interaction with Other Services

### 11.1 → Patient Service (Re-indexing)

After a successful merge, the Sync Service calls the Patient Service to re-index new and modified resources:

```protobuf
// Called on Patient Service
rpc ReindexResources(ReindexRequest) returns (ReindexResponse);

message ReindexRequest {
  repeated string resource_paths = 1;  // Git paths from the merge delta
}
```

The Sync Service sends only the paths that changed in the merge — NOT a full re-index. Full re-index is only triggered if incremental re-indexing fails.

### 11.2 → Auth Service (Peer Verification)

During handshake, the Sync Service calls the Auth Service to:

1. Verify the remote node's public key is trusted
2. Check if the remote node has been revoked
3. Authenticate the handshake signature

```protobuf
// Called on Auth Service
rpc CheckRevocation(CheckRevocationRequest) returns (CheckRevocationResponse);
```

### 11.3 → Formulary Service (Merge-Time Interaction Check)

When merging MedicationRequest resources, the Sync Service calls the Formulary Service:

```protobuf
// Called on Formulary Service
rpc CheckInteractions(CheckInteractionsRequest) returns (CheckInteractionsResponse);

message CheckInteractionsRequest {
  string patient_id = 1;
  bytes new_medication_json = 2;         // Incoming MedicationRequest
  repeated bytes existing_medications = 3; // Current active MedicationRequests
}
```

### 11.4 ← Sentinel Agent (Event Consumer)

The Sentinel Agent subscribes to `SyncCompleted` events via the `SubscribeEvents` RPC. It receives the full delta (list of new/modified resource paths and types) and uses this to scope its analysis.

### 11.5 ← Anchor Service (Event Consumer)

The Anchor Service subscribes to `SyncCompleted` events to trigger Merkle root recomputation and IOTA anchoring.

---

## 12. Error Handling

### 12.1 Sync Failure Modes

| Failure | Impact | Recovery |
|---------|--------|----------|
| Transport disconnect mid-sync | Partial data transfer | Retry from last known good state (Git is safe) |
| Merge conflict (Block level) | Resource not merged | Queue conflict for resolution, continue with other resources |
| Remote node revoked | Sync rejected | Log, alert admin, do not retry |
| Remote node untrusted | Sync rejected | Prompt admin for trust decision |
| Pack file corrupted | Cannot apply | Retry transfer, fall back to different transport |
| Git repository locked | Cannot merge | Wait + retry (Patient Service write lock) |
| Disk full | Cannot write pack/merge | Alert, prioritise critical-only sync |
| Re-indexing failure | SQLite stale | Schedule full rebuild, data safe in Git |

### 12.2 Retry Strategy

```go
func backoff(attempt int) time.Duration {
    base := 5 * time.Second
    max := 5 * time.Minute
    delay := base * time.Duration(1<<uint(attempt))  // Exponential
    if delay > max {
        delay = max
    }
    // Add jitter: ±25%
    jitter := time.Duration(rand.Int63n(int64(delay) / 2)) - delay/4
    return delay + jitter
}
```

- Max 3 retries per sync attempt
- After 3 failures: emit `SyncFailed` event, remove from queue, try again on next discovery

---

## 13. Configuration

```yaml
sync_service:
  grpc_port: 50052
  
  git:
    repo_path: /var/lib/open-nucleus/data
    
  transports:
    wifi_direct:
      enabled: true
      service_name: "_opennucleus._tcp"
      
    local_network:
      enabled: true
      service_name: "_opennucleus._tcp"
      
    bluetooth:
      enabled: true
      service_uuid: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
      
    usb:
      enabled: true
      watch_paths:
        - /media/
        - /mnt/usb/
      deployment_key_path: /var/lib/open-nucleus/keys/deployment.key
      
  sync:
    max_concurrent_syncs: 1          # V1: sequential only
    cooldown:
      local_network: 2m
      wifi_direct: 5m
      bluetooth: 10m
    max_retries: 3
    handshake_timeout: 10s
    transfer_timeout: 300s           # 5 min max per sync session
    
  priority:
    enabled: true                    # Enable bandwidth-aware prioritisation
    constrained_transports:          # Transports that trigger priority sync
      - bluetooth
      
  merge:
    drug_interaction_check: true     # Call Formulary during merge
    
  discovery:
    scan_interval: 30s               # How often to scan for peers
    peer_ttl: 300s                   # Mark peer as lost after 5 min unseen
    
  history:
    db_path: /var/lib/open-nucleus/sync.db
    max_history_entries: 10000
    
  events:
    buffer_size: 100                 # Event channel buffer
    
  logging:
    level: info
    format: json
```

---

## 14. Testing Strategy

### 14.1 Unit Tests

| Area | Coverage | Focus |
|------|----------|-------|
| Merge driver | 95% | Every conflict classification scenario |
| Transport adapter interface | 100% | Mock adapter satisfies interface |
| Priority classification | 100% | Every resource type mapped to correct tier |
| Pack generation | 90% | Correct Git objects included, delta compression works |
| Conflict storage | 90% | CRUD on conflicts table, state transitions |
| Event bus | 90% | Publish/subscribe, filtering, cleanup |

### 14.2 Integration Tests

| Test | Description |
|------|-------------|
| Two-node sync | Start two in-process nodes, write data on each, sync, verify both have all data |
| Conflict detection | Write conflicting data on two nodes, sync, verify correct classification |
| Conflict resolution | Create block conflict, resolve via API, verify resolution propagates |
| Transport fallback | Start sync on mock transport that fails, verify fallback to second transport |
| Bundle roundtrip | Export bundle on Node A, import on Node B, verify data matches |
| Revocation enforcement | Revoke Node B on Node A, attempt sync, verify rejection |
| Priority sync | Sync over mock constrained transport, verify critical data arrives first |
| Re-index trigger | Sync successfully, verify Patient Service re-indexes new resources |
| Event propagation | Sync successfully, verify SyncCompleted event reaches subscriber |

### 14.3 Stress Tests

| Test | Description |
|------|-------------|
| Large sync | 10,000 resources on each side, merge, measure time and memory |
| Rapid reconnect | Peer appears/disappears every 10s, verify no goroutine leaks |
| Concurrent discovery | 10 peers discovered simultaneously, verify orderly sync queue |

---

## 15. Performance Targets

All targets measured on Raspberry Pi 4 (4GB RAM).

| Operation | Target | Notes |
|-----------|--------|-------|
| Peer discovery (mDNS) | < 5s | Time from peer appearing to discovery event |
| Handshake + auth | < 500ms | Including Ed25519 verification |
| Sync 100 resources (Wi-Fi) | < 10s | Pack generation + transfer + merge + re-index |
| Sync 1,000 resources (Wi-Fi) | < 30s | |
| Sync 100 resources (Bluetooth) | < 60s | BLE bandwidth limited |
| Merge driver per file | < 10ms | Classification + merge of single FHIR resource |
| Conflict resolution | < 200ms | Write resolved resource via Patient Service |
| Bundle export (1,000 resources) | < 5s | Including encryption |
| Bundle import (1,000 resources) | < 15s | Including decryption + merge + re-index |
| Event emission | < 1ms | Publish to in-memory channel |
| Memory footprint | < 100MB RSS | During active sync of 1,000 resources |

---

*Open Nucleus • Sync Service Specification V1 • FibrinLab*  
*github.com/FibrinLab/open-nucleus*
