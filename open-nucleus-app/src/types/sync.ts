/** Response from GET /api/v1/sync/status. */
export interface SyncStatusResponse {
  state: string;
  last_sync?: string;
  pending_changes: number;
  node_id: string;
  site_id: string;
}

/** Information about a discovered sync peer. */
export interface PeerInfo {
  node_id: string;
  site_id: string;
  last_seen?: string;
  state: string;
  latency_ms?: number;
}
