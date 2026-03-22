import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { RoleDTO, LoginRequest, LoginResponse } from '../types/auth.ts';
import type { ApiEnvelope } from '../types/api-envelope.ts';
import { loadOrGenerateKeypair, signWithB64 } from '../lib/ed25519.ts';

/* ---------- helpers ---------- */

const DEVICE_ID_KEY = 'nucleus-device-id';
const SERVER_URL_KEY = 'nucleus-server-url';

function getOrCreateDeviceId(): string {
  let id = localStorage.getItem(DEVICE_ID_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(DEVICE_ID_KEY, id);
  }
  return id;
}

export function getServerUrl(): string {
  return localStorage.getItem(SERVER_URL_KEY) ?? '';
}

export function setServerUrl(url: string): void {
  localStorage.setItem(SERVER_URL_KEY, url.replace(/\/+$/, ''));
}

/* ---------- state shape ---------- */

export interface AuthState {
  status: 'initial' | 'loading' | 'authenticated' | 'error';
  token: string | null;
  refreshToken: string | null;
  role: RoleDTO | null;
  siteId: string | null;
  nodeId: string | null;
  deviceId: string | null;
  practitionerId: string | null;
  errorMessage: string | null;

  // Actions
  login: (serverUrl: string, practitionerId: string) => Promise<void>;
  logout: () => void;
  refresh: () => Promise<boolean>;
  setError: (message: string) => void;
  clearError: () => void;
  loadSavedAuth: () => void;
}

/* ---------- store ---------- */

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      status: 'initial',
      token: null,
      refreshToken: null,
      role: null,
      siteId: null,
      nodeId: null,
      deviceId: null,
      practitionerId: null,
      errorMessage: null,

      login: async (serverUrl: string, practitionerId: string) => {
        set({ status: 'loading', errorMessage: null });

        try {
          // 1. Persist server URL
          setServerUrl(serverUrl);

          // 2. Load or generate Ed25519 keypair
          const kp = loadOrGenerateKeypair();

          // 3. Get or create device ID
          const deviceId = getOrCreateDeviceId();

          // 4. Build challenge-response
          const nonce = new Date().toISOString();
          const signature = signWithB64(nonce, kp.secretKey);

          // Convert public key to base64url (no padding) — Go backend uses base64.RawURLEncoding
          const pubKeyB64Url = kp.publicKey
            .replace(/\+/g, '-')
            .replace(/\//g, '_')
            .replace(/=+$/, '');

          const body: LoginRequest = {
            device_id: deviceId,
            public_key: pubKeyB64Url,
            challenge_response: {
              nonce,
              signature,
              timestamp: nonce,
            },
            practitioner_id: practitionerId,
            bootstrap_secret: 'demo',
          };

          // 5. POST login (raw fetch — api-client depends on this store)
          const baseUrl = serverUrl.replace(/\/+$/, '');
          const res = await fetch(`${baseUrl}/api/v1/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
          });

          const envelope: ApiEnvelope<LoginResponse> = await res.json();

          if (!res.ok || envelope.status === 'error') {
            const msg =
              envelope.error?.message ?? `Login failed (${res.status})`;
            set({ status: 'error', errorMessage: msg });
            return;
          }

          const data = envelope.data!;

          // Write token to simple localStorage for synchronous access by api-client
          localStorage.setItem('nucleus:token', data.token);
          localStorage.setItem('nucleus:refresh_token', data.refresh_token);

          set({
            status: 'authenticated',
            token: data.token,
            refreshToken: data.refresh_token,
            role: data.role,
            siteId: data.site_id,
            nodeId: data.node_id,
            deviceId,
            practitionerId,
            errorMessage: null,
          });
        } catch (err) {
          const msg =
            err instanceof Error ? err.message : 'Unknown login error';
          set({ status: 'error', errorMessage: msg });
        }
      },

      logout: () => {
        localStorage.removeItem('nucleus:token');
        localStorage.removeItem('nucleus:refresh_token');
        set({
          status: 'initial',
          token: null,
          refreshToken: null,
          role: null,
          siteId: null,
          nodeId: null,
          deviceId: null,
          practitionerId: null,
          errorMessage: null,
        });
      },

      refresh: async (): Promise<boolean> => {
        const { refreshToken } = get();
        const baseUrl = getServerUrl();
        if (!refreshToken || !baseUrl) return false;

        try {
          const res = await fetch(`${baseUrl}/api/v1/auth/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
          });

          const envelope: ApiEnvelope<LoginResponse> = await res.json();

          if (!res.ok || envelope.status === 'error') {
            // Refresh failed — force logout
            get().logout();
            return false;
          }

          const data = envelope.data!;

          set({
            token: data.token,
            refreshToken: data.refresh_token,
            role: data.role,
            siteId: data.site_id,
            nodeId: data.node_id,
          });
          return true;
        } catch {
          get().logout();
          return false;
        }
      },

      setError: (message: string) => {
        set({ status: 'error', errorMessage: message });
      },

      clearError: () => {
        set({ errorMessage: null, status: get().token ? 'authenticated' : 'initial' });
      },

      loadSavedAuth: () => {
        // persist middleware rehydrates automatically,
        // but callers can use this to re-derive status after hydration.
        const { token } = get();
        if (token) {
          set({ status: 'authenticated' });
        }
      },
    }),
    {
      name: 'nucleus-auth',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        role: state.role,
        siteId: state.siteId,
        nodeId: state.nodeId,
        deviceId: state.deviceId,
        practitionerId: state.practitionerId,
      }),
      onRehydrateStorage: () => {
        return (state) => {
          // Sync token to simple localStorage for api-client synchronous access
          if (state?.token) {
            localStorage.setItem('nucleus:token', state.token);
            if (state.refreshToken) {
              localStorage.setItem('nucleus:refresh_token', state.refreshToken);
            }
          }
        };
      },
    },
  ),
);
