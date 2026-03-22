/** Backend (IOTA / Hedera / etc.) connection info. */
export interface BackendInfo {
  name: string;
  connected: boolean;
  error_message?: string;
}

/** Current anchor status from GET /api/v1/anchor/status. */
export interface AnchorStatus {
  has_been_anchored: boolean;
  last_anchor_time?: string;
  status: string;
  merkle_root?: string;
  pending_commits?: number;
  queue_depth?: number;
  backends: BackendInfo[];
}

/** Verification method inside a DID Document. */
export interface VerificationMethodDTO {
  id: string;
  type: string;
  controller: string;
  public_key_multibase: string;
}

/** W3C DID Document for node or device identity. */
export interface DIDDocument {
  id: string;
  context: string[];
  verification_method: VerificationMethodDTO[];
  authentication: string[];
}
