import nacl from 'tweetnacl';
import { encodeBase64, decodeBase64 } from 'tweetnacl-util';

// ---------------------------------------------------------------------------
// Key types
// ---------------------------------------------------------------------------

export interface Ed25519Keypair {
  publicKey: Uint8Array;
  secretKey: Uint8Array;
}

/**
 * Serialized form stored in localStorage (base64 strings).
 * Kept for backwards-compatibility with prior callers.
 */
export interface StoredKeypair {
  publicKey: string;
  secretKey: string;
}

// ---------------------------------------------------------------------------
// localStorage keys
// ---------------------------------------------------------------------------

const PUB_KEY = 'nucleus:ed25519_pub';
const SEC_KEY = 'nucleus:ed25519_sec';

// Legacy key used by the earlier implementation — checked during load.
const LEGACY_KEY = 'nucleus-ed25519-keypair';

// ---------------------------------------------------------------------------
// Core operations
// ---------------------------------------------------------------------------

/** Generate a fresh Ed25519 keypair. */
export function generateKeypair(): Ed25519Keypair {
  const kp = nacl.sign.keyPair();
  return { publicKey: kp.publicKey, secretKey: kp.secretKey };
}

/**
 * Sign a UTF-8 message with the secret key.
 * Returns a base64url-encoded detached signature.
 */
export function sign(secretKey: Uint8Array, message: string): string {
  const msgBytes = new TextEncoder().encode(message);
  const sig = nacl.sign.detached(msgBytes, secretKey);
  return toBase64Url(encodeBase64(sig));
}

/**
 * Sign a UTF-8 message using a base64 secret key string.
 * Returns a base64-encoded detached signature.
 * Convenience overload for callers that store keys as strings.
 */
export function signWithB64(message: string, secretKeyB64: string): string {
  const skBytes = decodeBase64(secretKeyB64);
  const msgBytes = new TextEncoder().encode(message);
  const sig = nacl.sign.detached(msgBytes, skBytes);
  return encodeBase64(sig);
}

/** Encode a public key as a base64url string. */
export function encodePublicKey(publicKey: Uint8Array): string {
  return toBase64Url(encodeBase64(publicKey));
}

/** First 8 hex characters of the public key (device fingerprint). */
export function getFingerprint(publicKey: Uint8Array): string {
  return Array.from(publicKey.slice(0, 4))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * Compute a fingerprint from a base64-encoded public key string.
 * Returns uppercase hex. Kept for backwards-compatibility.
 */
export function fingerprint(publicKeyB64: string): string {
  const bytes = decodeBase64(publicKeyB64);
  return Array.from(bytes.slice(0, 4))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
    .toUpperCase();
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------

/** Save a keypair to localStorage (base64-encoded). */
export function saveKeypair(keypair: Ed25519Keypair): void {
  localStorage.setItem(PUB_KEY, encodeBase64(keypair.publicKey));
  localStorage.setItem(SEC_KEY, encodeBase64(keypair.secretKey));
}

/** Load the keypair from localStorage, or return null if absent. */
export function loadKeypair(): Ed25519Keypair | null {
  let pub = localStorage.getItem(PUB_KEY);
  let sec = localStorage.getItem(SEC_KEY);

  // Migrate from legacy single-key storage if present
  if (!pub || !sec) {
    const legacy = localStorage.getItem(LEGACY_KEY);
    if (legacy) {
      try {
        const parsed = JSON.parse(legacy) as StoredKeypair;
        if (parsed.publicKey && parsed.secretKey) {
          pub = parsed.publicKey;
          sec = parsed.secretKey;
          // Migrate to new keys
          localStorage.setItem(PUB_KEY, pub);
          localStorage.setItem(SEC_KEY, sec);
          localStorage.removeItem(LEGACY_KEY);
        }
      } catch {
        // ignore corrupt legacy data
      }
    }
  }

  if (!pub || !sec) return null;
  return {
    publicKey: decodeBase64(pub),
    secretKey: decodeBase64(sec),
  };
}

/** Load a keypair as base64 strings (for callers that prefer strings). */
export function loadStoredKeypair(): StoredKeypair | null {
  const kp = loadKeypair();
  if (!kp) return null;
  return {
    publicKey: encodeBase64(kp.publicKey),
    secretKey: encodeBase64(kp.secretKey),
  };
}

/** Check whether a keypair exists in localStorage. */
export function keypairExists(): boolean {
  if (localStorage.getItem(PUB_KEY) && localStorage.getItem(SEC_KEY)) return true;
  // Also check legacy key
  const legacy = localStorage.getItem(LEGACY_KEY);
  if (legacy) {
    try {
      const parsed = JSON.parse(legacy) as StoredKeypair;
      return !!(parsed.publicKey && parsed.secretKey);
    } catch {
      return false;
    }
  }
  return false;
}

/** Generate a keypair and immediately persist it. Returns the stored form. */
export function loadOrGenerateKeypair(): StoredKeypair {
  const existing = loadStoredKeypair();
  if (existing) return existing;
  const kp = generateKeypair();
  saveKeypair(kp);
  return {
    publicKey: encodeBase64(kp.publicKey),
    secretKey: encodeBase64(kp.secretKey),
  };
}

/** Delete the stored keypair from localStorage. */
export function clearKeypair(): void {
  localStorage.removeItem(PUB_KEY);
  localStorage.removeItem(SEC_KEY);
  localStorage.removeItem(LEGACY_KEY);
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function toBase64Url(b64: string): string {
  return b64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}
