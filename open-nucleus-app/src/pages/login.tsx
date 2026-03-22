import { useState, useCallback, useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useAuthStore, getServerUrl, setServerUrl } from '../stores/auth-store';
import {
  loadStoredKeypair,
  generateKeypair,
  saveKeypair,
  fingerprint,
  type StoredKeypair,
} from '../lib/ed25519';
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

  useEffect(() => {
    if (status === 'authenticated') {
      navigate({ to: '/dashboard' });
    }
  }, [status, navigate]);

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

  const handleGenerateKeypair = useCallback(() => {
    const raw = generateKeypair();
    saveKeypair(raw);
    const stored: StoredKeypair = {
      publicKey: encodeBase64(raw.publicKey),
      secretKey: encodeBase64(raw.secretKey),
    };
    setKeypair(stored);
  }, []);

  const canSignIn =
    connStatus === 'connected' &&
    keypair !== null &&
    practitionerId.trim().length > 0 &&
    status !== 'loading';

  const handleSignIn = useCallback(async () => {
    if (!canSignIn) return;
    await login(serverUrl, practitionerId.trim());
  }, [canSignIn, login, serverUrl, practitionerId]);

  const dotColor =
    connStatus === 'connected' ? '#2E7D32'
    : connStatus === 'failed' ? '#D32F2F'
    : connStatus === 'testing' ? '#F57F17'
    : '#999999';

  const statusLabel =
    connStatus === 'connected' ? 'Connected'
    : connStatus === 'failed' ? 'Failed'
    : connStatus === 'testing' ? 'Testing...'
    : 'Not tested';

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      backgroundColor: 'var(--color-paper)',
      color: 'var(--color-ink)',
      padding: '24px',
    }}>
      <div style={{
        width: '100%',
        maxWidth: '420px',
        border: '1px solid var(--color-border)',
        borderRadius: '8px',
        padding: '40px 32px',
        backgroundColor: 'var(--color-surface)',
      }}>
        {/* Header */}
        <h1 style={{
          fontSize: '28px',
          fontWeight: 700,
          textAlign: 'center',
          fontFamily: 'var(--font-mono)',
          margin: '0 0 4px 0',
          letterSpacing: '-0.5px',
        }}>
          Open Nucleus
        </h1>
        <p style={{
          textAlign: 'center',
          color: 'var(--color-muted)',
          fontSize: '14px',
          margin: '0 0 32px 0',
        }}>
          Electronic Health Record
        </p>

        {/* Server URL */}
        <label style={{
          display: 'block',
          fontSize: '11px',
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: '1px',
          color: 'var(--color-muted)',
          marginBottom: '6px',
        }}>
          Server URL
        </label>
        <div style={{ display: 'flex', gap: '8px', marginBottom: '8px' }}>
          <input
            type="text"
            value={serverUrl}
            onChange={(e) => { setUrl(e.target.value); setConnStatus('idle'); }}
            placeholder="http://localhost:8080"
            style={{
              flex: 1,
              padding: '8px 12px',
              fontSize: '14px',
              border: '1px solid var(--color-border)',
              borderRadius: '4px',
              backgroundColor: 'transparent',
              color: 'var(--color-ink)',
              outline: 'none',
              fontFamily: 'var(--font-mono)',
            }}
          />
          <button
            type="button"
            onClick={testConnection}
            disabled={connStatus === 'testing' || !serverUrl.trim()}
            style={{
              padding: '8px 16px',
              fontSize: '12px',
              fontWeight: 600,
              textTransform: 'uppercase',
              letterSpacing: '0.5px',
              border: '1px solid var(--color-border)',
              borderRadius: '4px',
              backgroundColor: 'transparent',
              color: 'var(--color-ink)',
              cursor: connStatus === 'testing' ? 'wait' : 'pointer',
              opacity: !serverUrl.trim() ? 0.4 : 1,
            }}
          >
            Test
          </button>
        </div>

        {/* Connection status */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '12px', color: 'var(--color-muted)', marginBottom: '24px' }}>
          <span style={{
            width: '8px',
            height: '8px',
            borderRadius: '50%',
            backgroundColor: dotColor,
            display: 'inline-block',
          }} />
          <span>{statusLabel}</span>
          {connMessage && <span style={{ marginLeft: '4px' }}>— {connMessage}</span>}
        </div>

        {/* Divider */}
        <hr style={{ border: 'none', borderTop: '1px solid var(--color-border)', margin: '0 0 24px 0' }} />

        {/* Device Keypair */}
        <label style={{
          display: 'block',
          fontSize: '11px',
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: '1px',
          color: 'var(--color-muted)',
          marginBottom: '6px',
        }}>
          Device Keypair
        </label>
        {keypair ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '24px' }}>
            <span style={{ fontFamily: 'var(--font-mono)', fontSize: '14px', letterSpacing: '2px' }}>
              {fingerprint(keypair.publicKey)}
            </span>
            <span style={{ fontSize: '12px', color: 'var(--color-muted)' }}>
              Ed25519 loaded
            </span>
          </div>
        ) : (
          <button
            type="button"
            onClick={handleGenerateKeypair}
            style={{
              display: 'block',
              width: '100%',
              padding: '10px',
              fontSize: '14px',
              fontWeight: 600,
              border: '1px solid var(--color-border)',
              borderRadius: '4px',
              backgroundColor: 'transparent',
              color: 'var(--color-ink)',
              cursor: 'pointer',
              marginBottom: '24px',
            }}
          >
            Generate Keypair
          </button>
        )}

        {/* Practitioner ID */}
        <label style={{
          display: 'block',
          fontSize: '11px',
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: '1px',
          color: 'var(--color-muted)',
          marginBottom: '6px',
        }}>
          Practitioner ID
        </label>
        <input
          type="text"
          value={practitionerId}
          onChange={(e) => setPractitionerId(e.target.value)}
          placeholder="e.g. demo-clinician"
          style={{
            display: 'block',
            width: '100%',
            padding: '8px 12px',
            fontSize: '14px',
            border: '1px solid var(--color-border)',
            borderRadius: '4px',
            backgroundColor: 'transparent',
            color: 'var(--color-ink)',
            outline: 'none',
            marginBottom: '24px',
          }}
        />

        {/* Sign In */}
        <button
          type="button"
          onClick={handleSignIn}
          disabled={!canSignIn}
          style={{
            display: 'block',
            width: '100%',
            padding: '12px',
            fontSize: '14px',
            fontWeight: 700,
            textTransform: 'uppercase',
            letterSpacing: '1px',
            backgroundColor: '#111111',
            color: '#FAFAF8',
            border: 'none',
            borderRadius: '4px',
            cursor: canSignIn ? 'pointer' : 'not-allowed',
            opacity: canSignIn ? 1 : 0.3,
          }}
        >
          {status === 'loading' ? 'Signing in...' : 'Sign In'}
        </button>

        {/* Error */}
        {errorMessage && (
          <div style={{
            marginTop: '16px',
            padding: '12px',
            fontSize: '13px',
            color: '#D32F2F',
            border: '1px solid #D32F2F',
            borderRadius: '4px',
          }}>
            {errorMessage}
          </div>
        )}
      </div>
    </div>
  );
}
