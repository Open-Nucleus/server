import { useState, useCallback, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useAuthStore, getServerUrl, setServerUrl } from '../stores/auth-store.ts';
import {
  loadStoredKeypair,
  generateKeypair,
  saveKeypair,
  fingerprint,
  type StoredKeypair,
} from '../lib/ed25519.ts';
import { encodeBase64 } from 'tweetnacl-util';

type ConnectionStatus = 'idle' | 'testing' | 'connected' | 'failed';

export default function LoginPage() {
  const { status, errorMessage, login, clearError } = useAuthStore();
  const navigate = useNavigate();

  const [serverUrl, setUrl] = useState(getServerUrl() || 'http://localhost:8080');
  const [connStatus, setConnStatus] = useState<ConnectionStatus>('idle');
  const [connMessage, setConnMessage] = useState('');
  const [keypair, setKeypair] = useState<StoredKeypair | null>(loadStoredKeypair());
  const [practitionerId, setPractitionerId] = useState('');

  // Redirect if already authenticated
  useEffect(() => {
    if (status === 'authenticated') {
      navigate({ to: '/dashboard' });
    }
  }, [status, navigate]);

  /* ---- connection test ---- */
  const testConnection = useCallback(async () => {
    setConnStatus('testing');
    setConnMessage('');
    clearError();

    const base = serverUrl.replace(/\/+$/, '');
    try {
      const res = await fetch(`${base}/health`, { signal: AbortSignal.timeout(5000) });
      if (res.ok) {
        setConnStatus('connected');
        setConnMessage('Server reachable');
        setServerUrl(base);
      } else {
        setConnStatus('failed');
        setConnMessage(`Server responded with ${res.status}`);
      }
    } catch (err) {
      setConnStatus('failed');
      setConnMessage(err instanceof Error ? err.message : 'Connection failed');
    }
  }, [serverUrl, clearError]);

  /* ---- keypair generation ---- */
  const handleGenerateKeypair = useCallback(() => {
    const raw = generateKeypair();
    saveKeypair(raw);
    const stored: StoredKeypair = {
      publicKey: encodeBase64(raw.publicKey),
      secretKey: encodeBase64(raw.secretKey),
    };
    setKeypair(stored);
  }, []);

  /* ---- sign in ---- */
  const canSignIn =
    connStatus === 'connected' &&
    keypair !== null &&
    practitionerId.trim().length > 0 &&
    status !== 'loading';

  const handleSignIn = useCallback(async () => {
    if (!canSignIn) return;
    await login(serverUrl, practitionerId.trim());
  }, [canSignIn, login, serverUrl, practitionerId]);

  /* ---- connection indicator ---- */
  const dotColor =
    connStatus === 'connected'
      ? 'bg-[var(--color-success)]'
      : connStatus === 'failed'
        ? 'bg-[var(--color-error)]'
        : connStatus === 'testing'
          ? 'bg-[var(--color-warning)]'
          : 'bg-[var(--color-muted)]';

  const statusLabel =
    connStatus === 'connected'
      ? 'Connected'
      : connStatus === 'failed'
        ? 'Failed'
        : connStatus === 'testing'
          ? 'Testing...'
          : 'Not tested';

  return (
    <div className="flex items-center justify-center min-h-screen bg-[var(--color-paper)] dark:bg-[var(--color-paper-dark)]">
      <div className="w-full max-w-md typewriter-border rounded-[var(--radius-md)] p-8 bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]">
        {/* Header */}
        <h1 className="text-3xl font-bold tracking-tight text-center font-mono">
          Open Nucleus
        </h1>
        <p className="text-center text-[var(--color-muted)] text-sm mt-1">
          Electronic Health Record
        </p>

        {/* Server URL */}
        <div className="mt-8 space-y-2">
          <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-muted)]">
            Server URL
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              value={serverUrl}
              onChange={(e) => {
                setUrl(e.target.value);
                setConnStatus('idle');
              }}
              placeholder="http://localhost:8080"
              className="flex-1 px-3 py-2 text-sm typewriter-border rounded-[var(--radius-sm)] bg-transparent focus:outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]"
            />
            <button
              onClick={testConnection}
              disabled={connStatus === 'testing' || !serverUrl.trim()}
              className="px-3 py-2 text-xs font-semibold uppercase typewriter-border rounded-[var(--radius-sm)] hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)] disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            >
              Test
            </button>
          </div>

          {/* Connection indicator */}
          <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
            <span className={`inline-block w-2 h-2 rounded-full ${dotColor}`} />
            <span>{statusLabel}</span>
            {connMessage && (
              <span className="ml-1 text-[var(--color-muted)]">
                — {connMessage}
              </span>
            )}
          </div>
        </div>

        {/* Separator */}
        <hr className="my-6 border-[var(--color-border)] dark:border-[var(--color-border-dark)]" />

        {/* Device / Keypair */}
        <div className="space-y-2">
          <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-muted)]">
            Device Keypair
          </label>
          {keypair ? (
            <div className="flex items-center gap-3">
              <span className="font-mono text-sm tracking-widest">
                {fingerprint(keypair.publicKey)}
              </span>
              <span className="text-xs text-[var(--color-muted)]">
                Ed25519 keypair loaded
              </span>
            </div>
          ) : (
            <button
              onClick={handleGenerateKeypair}
              className="w-full px-3 py-2 text-sm font-semibold typewriter-border rounded-[var(--radius-sm)] hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)] transition-colors"
            >
              Generate Keypair
            </button>
          )}
        </div>

        {/* Practitioner ID */}
        <div className="mt-6 space-y-2">
          <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-muted)]">
            Practitioner ID
          </label>
          <input
            type="text"
            value={practitionerId}
            onChange={(e) => setPractitionerId(e.target.value)}
            placeholder="e.g. practitioner-001"
            className="w-full px-3 py-2 text-sm typewriter-border rounded-[var(--radius-sm)] bg-transparent focus:outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]"
          />
        </div>

        {/* Sign In */}
        <button
          onClick={handleSignIn}
          disabled={!canSignIn}
          className="mt-6 w-full py-2.5 text-sm font-bold uppercase tracking-wider bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)] rounded-[var(--radius-sm)] hover:opacity-90 disabled:opacity-30 disabled:cursor-not-allowed transition-opacity"
        >
          {status === 'loading' ? 'Signing in...' : 'Sign In'}
        </button>

        {/* Error display */}
        {errorMessage && (
          <div className="mt-4 p-3 text-sm text-[var(--color-error)] typewriter-border border-[var(--color-error)] rounded-[var(--radius-sm)]">
            {errorMessage}
          </div>
        )}
      </div>
    </div>
  );
}
