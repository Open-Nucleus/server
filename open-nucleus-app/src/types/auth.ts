/** Challenge-response payload for Ed25519 login. */
export interface ChallengeResponse {
  nonce: string;
  signature: string;
  timestamp: string;
}

/** Login request body sent to POST /api/v1/auth/login. */
export interface LoginRequest {
  device_id: string;
  public_key: string;
  challenge_response: ChallengeResponse;
  practitioner_id: string;
}

/** Role descriptor returned by the backend. */
export interface RoleDTO {
  code: string;
  name: string;
  permissions: string[];
}

/** Successful login response. */
export interface LoginResponse {
  token: string;
  expires_at: string;
  refresh_token: string;
  role: RoleDTO;
  site_id: string;
  node_id: string;
}

/** Response from GET /api/v1/auth/whoami. */
export interface WhoamiResponse {
  subject: string;
  node_id: string;
  site_id: string;
  role: RoleDTO;
}
