import type { ApiEnvelope } from '../types/api-envelope';
import { useAuthStore, getServerUrl } from '@/stores/auth-store';

// ---------------------------------------------------------------------------
// Token & URL helpers — read directly from Zustand (same store login writes to)
// ---------------------------------------------------------------------------

function getToken(): string | null {
  // Try Zustand state first, then localStorage fallback (handles rehydration race)
  return useAuthStore.getState().token || localStorage.getItem('nucleus:token');
}

function getBaseUrl(): string {
  return getServerUrl() || 'https://server-6xbu.onrender.com';
}

// ---------------------------------------------------------------------------
// AppError
// ---------------------------------------------------------------------------

export class AppError extends Error {
  code: string;
  statusCode: number;
  details: unknown;

  constructor(message: string, code: string, statusCode: number, details?: unknown) {
    super(message);
    this.name = 'AppError';
    this.code = code;
    this.statusCode = statusCode;
    this.details = details;
  }
}

// ---------------------------------------------------------------------------
// Core request function
// ---------------------------------------------------------------------------

/**
 * Low-level fetch wrapper that:
 * 1. Attaches the Bearer token from Zustand auth store
 * 2. Handles 401 with a single refresh-and-retry
 * 3. Unwraps ApiEnvelope<T>
 * 4. Throws AppError on failure
 */
export async function apiRequest<T>(
  method: string,
  path: string,
  body?: unknown,
  params?: Record<string, string>,
): Promise<ApiEnvelope<T>> {
  const url = buildUrl(path, params);
  const headers = buildHeaders();

  const res = await fetch(url, {
    method,
    headers,
    body: body != null ? JSON.stringify(body) : undefined,
  });

  // Attempt a single token refresh on 401
  if (res.status === 401) {
    const refreshed = await useAuthStore.getState().refresh();
    if (refreshed) {
      const retryHeaders = buildHeaders();
      const retryRes = await fetch(url, {
        method,
        headers: retryHeaders,
        body: body != null ? JSON.stringify(body) : undefined,
      });
      return handleResponse<T>(retryRes);
    }
    // Refresh failed — force logout (router auth guard will redirect to /login)
    useAuthStore.getState().logout();
    throw new AppError('Session expired', 'AUTH_EXPIRED', 401);
  }

  return handleResponse<T>(res);
}

// ---------------------------------------------------------------------------
// Convenience methods
// ---------------------------------------------------------------------------

export function apiGet<T>(
  path: string,
  params?: Record<string, string>,
): Promise<ApiEnvelope<T>> {
  return apiRequest<T>('GET', path, undefined, params);
}

export function apiPost<T>(
  path: string,
  body: unknown,
): Promise<ApiEnvelope<T>> {
  return apiRequest<T>('POST', path, body);
}

export function apiPut<T>(
  path: string,
  body: unknown,
): Promise<ApiEnvelope<T>> {
  return apiRequest<T>('PUT', path, body);
}

export function apiDelete<T>(
  path: string,
): Promise<ApiEnvelope<T>> {
  return apiRequest<T>('DELETE', path);
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

function buildUrl(path: string, params?: Record<string, string>): string {
  const base = getBaseUrl();
  const url = new URL(path, base);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      url.searchParams.set(key, value);
    }
  }
  return url.toString();
}

function buildHeaders(): HeadersInit {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Accept: 'application/json',
    'X-Break-Glass': 'true', // Emergency consent bypass for demo
  };
  const token = getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

async function handleResponse<T>(res: Response): Promise<ApiEnvelope<T>> {
  let envelope: ApiEnvelope<T>;

  try {
    envelope = (await res.json()) as ApiEnvelope<T>;
  } catch {
    throw new AppError(
      `Unexpected response (HTTP ${res.status})`,
      'PARSE_ERROR',
      res.status,
    );
  }

  if (!res.ok || envelope.status === 'error') {
    throw new AppError(
      envelope.error?.message ?? `Request failed (HTTP ${res.status})`,
      envelope.error?.code ?? 'UNKNOWN',
      res.status,
      envelope.error?.details,
    );
  }

  return envelope;
}
