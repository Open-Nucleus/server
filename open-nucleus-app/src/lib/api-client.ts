import type { ApiEnvelope } from '../types/api-envelope';
import { API } from './api-paths';

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

const BASE_URL_KEY = 'nucleus:base_url';
const TOKEN_KEY = 'nucleus:token';
const REFRESH_TOKEN_KEY = 'nucleus:refresh_token';

/** Returns the API base URL from localStorage or the default. */
export function getBaseUrl(): string {
  if (typeof window !== 'undefined') {
    return localStorage.getItem(BASE_URL_KEY) || 'http://localhost:8080';
  }
  return 'http://localhost:8080';
}

/** Persists the API base URL to localStorage. */
export function setBaseUrl(url: string): void {
  localStorage.setItem(BASE_URL_KEY, url);
}

// ---------------------------------------------------------------------------
// Token helpers (simple localStorage store — zustand store may wrap these)
// ---------------------------------------------------------------------------

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setRefreshToken(token: string): void {
  localStorage.setItem(REFRESH_TOKEN_KEY, token);
}

export function clearTokens(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
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
 * 1. Attaches the Bearer token
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
    const refreshed = await attemptRefresh();
    if (refreshed) {
      const retryHeaders = buildHeaders();
      const retryRes = await fetch(url, {
        method,
        headers: retryHeaders,
        body: body != null ? JSON.stringify(body) : undefined,
      });
      return handleResponse<T>(retryRes);
    }
    // Refresh failed — clear tokens (router auth guard will redirect to /login)
    clearTokens();
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

async function attemptRefresh(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;

  try {
    const url = buildUrl(API.auth.refresh);
    const res = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!res.ok) return false;

    const envelope = (await res.json()) as ApiEnvelope<{
      token: string;
      refresh_token: string;
    }>;

    if (envelope.status === 'success' && envelope.data) {
      setToken(envelope.data.token);
      setRefreshToken(envelope.data.refresh_token);
      return true;
    }
    return false;
  } catch {
    return false;
  }
}
