import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { useUIStore } from '@/stores/ui-store';
import {
  LoadingSkeleton,
  EmptyState,
  ErrorState,
  StatusIndicator,
  PaginationControls,
} from '@/components';
import { cn } from '@/lib/utils';
import { toDisplayDateTime } from '@/lib/date-utils';
import { Shield, CheckCircle, XCircle, History, Server, Anchor } from 'lucide-react';
import type { AnchorStatus, BackendInfo, DIDDocument } from '@/types';

type IntegrityTab = 'status' | 'verification' | 'history' | 'backends';

interface AnchorRecord {
  id: string;
  merkle_root: string;
  commit_hash: string;
  anchored_at: string;
  backend: string;
  transaction_id?: string;
}

interface VerifyResult {
  valid: boolean;
  commit_hash: string;
  merkle_root?: string;
  anchored_at?: string;
  message?: string;
}

export default function IntegrityPage() {
  const setPageTitle = useUIStore((s) => s.setPageTitle);
  useEffect(() => setPageTitle('Integrity & Anchor'), [setPageTitle]);

  const [activeTab, setActiveTab] = useState<IntegrityTab>('status');
  const [historyPage, setHistoryPage] = useState(1);
  const queryClient = useQueryClient();

  /* ---- anchor status ---- */
  const statusQuery = useQuery({
    queryKey: ['anchor', 'status'],
    queryFn: () => apiGet<AnchorStatus>(API.anchor.status),
    refetchInterval: 15_000,
  });

  /* ---- node DID ---- */
  const didQuery = useQuery({
    queryKey: ['anchor', 'did', 'node'],
    queryFn: () => apiGet<DIDDocument>(API.anchor.didNode),
    enabled: activeTab === 'status',
  });

  /* ---- anchor history ---- */
  const historyQuery = useQuery({
    queryKey: ['anchor', 'history', historyPage],
    queryFn: () =>
      apiGet<AnchorRecord[]>(API.anchor.history, {
        page: String(historyPage),
        per_page: '20',
      }),
    enabled: activeTab === 'history',
  });

  /* ---- backends ---- */
  const backendsQuery = useQuery({
    queryKey: ['anchor', 'backends'],
    queryFn: () => apiGet<BackendInfo[]>(API.anchor.backends),
    enabled: activeTab === 'backends',
  });

  /* ---- trigger anchor ---- */
  const triggerMutation = useMutation({
    mutationFn: () => apiPost<void>(API.anchor.trigger, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['anchor'] });
    },
  });

  const tabs: { key: IntegrityTab; label: string; icon: React.ReactNode }[] = [
    { key: 'status', label: 'Status', icon: <Shield size={14} /> },
    { key: 'verification', label: 'Verification', icon: <CheckCircle size={14} /> },
    { key: 'history', label: 'History', icon: <History size={14} /> },
    { key: 'backends', label: 'Backends', icon: <Server size={14} /> },
  ];

  return (
    <div className="page-padding space-y-4">
      {/* header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-mono text-lg font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            Integrity & Anchor
          </h1>
          <p className="text-xs text-[var(--color-muted)] mt-1">
            Cryptographic anchoring and data integrity verification
          </p>
        </div>
        <button
          type="button"
          onClick={() => triggerMutation.mutate()}
          disabled={triggerMutation.isPending}
          className={cn(
            'flex items-center gap-2 px-4 py-2 text-xs font-mono font-semibold uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
            'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]',
            'hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed transition-opacity',
          )}
        >
          <Anchor size={12} />
          {triggerMutation.isPending ? 'Anchoring...' : 'Trigger Anchor'}
        </button>
      </div>

      {/* tabs */}
      <div className="flex gap-0 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
        {tabs.map((tab) => (
          <button
            type="button"
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={cn(
              'flex items-center gap-1.5 px-4 py-2.5 text-xs font-mono uppercase tracking-wider transition-colors duration-100 cursor-pointer',
              activeTab === tab.key
                ? 'border-b-2 border-[var(--color-ink)] dark:border-[var(--color-sidebar-text)] text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] font-semibold'
                : 'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
            )}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* tab content */}
      <div className="rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)] bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]">
        {activeTab === 'status' && (
          <AnchorStatusPanel
            status={statusQuery.data?.data}
            did={didQuery.data?.data}
            loading={statusQuery.isLoading}
            error={statusQuery.isError}
          />
        )}

        {activeTab === 'verification' && <VerificationPanel />}

        {activeTab === 'history' && (
          <AnchorHistoryPanel
            records={historyQuery.data?.data ?? []}
            loading={historyQuery.isLoading}
            error={historyQuery.isError}
            onRetry={() => historyQuery.refetch()}
            page={historyPage}
            totalPages={historyQuery.data?.pagination?.total_pages ?? 1}
            onPageChange={setHistoryPage}
          />
        )}

        {activeTab === 'backends' && (
          <BackendsPanel
            backends={backendsQuery.data?.data ?? []}
            loading={backendsQuery.isLoading}
            error={backendsQuery.isError}
            onRetry={() => backendsQuery.refetch()}
          />
        )}
      </div>
    </div>
  );
}

/* ================================================================
   Sub-panels
   ================================================================ */

function AnchorStatusPanel({
  status,
  did,
  loading,
  error,
}: {
  status?: AnchorStatus;
  did?: DIDDocument;
  loading: boolean;
  error: boolean;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={5} /></div>;
  if (error || !status) return <div className="p-6"><ErrorState message="Failed to load anchor status" /></div>;

  const rows: { label: string; value: React.ReactNode }[] = [
    {
      label: 'Anchor State',
      value: (
        <StatusIndicator
          status={status.has_been_anchored ? 'active' : 'inactive'}
          label={status.has_been_anchored ? 'Anchored' : 'Not Anchored'}
        />
      ),
    },
    { label: 'Status', value: status.status },
    {
      label: 'Last Anchor',
      value: status.last_anchor_time ? toDisplayDateTime(status.last_anchor_time) : 'Never',
    },
    {
      label: 'Merkle Root',
      value: status.merkle_root ? (
        <span className="font-mono text-[10px] break-all">{status.merkle_root}</span>
      ) : (
        '--'
      ),
    },
    { label: 'Pending Commits', value: String(status.pending_commits ?? 0) },
    { label: 'Queue Depth', value: String(status.queue_depth ?? 0) },
  ];

  if (did) {
    rows.push({
      label: 'Node DID',
      value: <span className="font-mono text-[10px] break-all">{did.id}</span>,
    });
  }

  return (
    <div className="p-6 space-y-4">
      {rows.map((row) => (
        <div
          key={row.label}
          className="flex items-start justify-between py-2 border-b border-[var(--color-border)]/30 dark:border-[var(--color-border-dark)]/30 last:border-0 gap-4"
        >
          <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)] shrink-0">
            {row.label}
          </span>
          <span className="text-sm text-right text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            {row.value}
          </span>
        </div>
      ))}
    </div>
  );
}

function VerificationPanel() {
  const [commitHash, setCommitHash] = useState('');
  const [result, setResult] = useState<VerifyResult | null>(null);

  const verifyMutation = useMutation({
    mutationFn: (hash: string) =>
      apiPost<VerifyResult>(API.anchor.verify, { commit_hash: hash }),
    onSuccess: (envelope) => {
      setResult(envelope.data ?? null);
    },
    onError: () => {
      setResult(null);
    },
  });

  const handleVerify = () => {
    if (!commitHash.trim()) return;
    verifyMutation.mutate(commitHash.trim());
  };

  return (
    <div className="p-6 space-y-6">
      <div>
        <label className="block font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)] mb-2">
          Commit Hash
        </label>
        <div className="flex gap-2">
          <input
            type="text"
            value={commitHash}
            onChange={(e) => setCommitHash(e.target.value)}
            placeholder="Enter git commit hash to verify..."
            className={cn(
              'flex-1 px-3 py-2 text-sm font-mono bg-transparent rounded-[var(--radius-sm)]',
              'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
              'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
              'placeholder:text-[var(--color-muted)] placeholder:opacity-60',
              'focus:outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
            )}
            onKeyDown={(e) => e.key === 'Enter' && handleVerify()}
          />
          <button
            type="button"
            onClick={handleVerify}
            disabled={!commitHash.trim() || verifyMutation.isPending}
            className={cn(
              'px-4 py-2 text-xs font-mono font-semibold uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
              'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]',
              'hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed transition-opacity',
            )}
          >
            {verifyMutation.isPending ? 'Verifying...' : 'Verify'}
          </button>
        </div>
      </div>

      {verifyMutation.isError && (
        <div className="p-4 rounded-[var(--radius-sm)] border border-[var(--color-error)] bg-[var(--color-error)]/5">
          <p className="text-sm text-[var(--color-error)]">
            Verification failed: {verifyMutation.error instanceof Error ? verifyMutation.error.message : 'Unknown error'}
          </p>
        </div>
      )}

      {result && (
        <div
          className={cn(
            'p-4 rounded-[var(--radius-sm)] border space-y-3',
            result.valid
              ? 'border-[var(--color-success)]/40 bg-[var(--color-success)]/5'
              : 'border-[var(--color-error)]/40 bg-[var(--color-error)]/5',
          )}
        >
          <div className="flex items-center gap-2">
            {result.valid ? (
              <CheckCircle size={18} className="text-[var(--color-success)]" />
            ) : (
              <XCircle size={18} className="text-[var(--color-error)]" />
            )}
            <span
              className={cn(
                'font-mono text-sm font-semibold uppercase tracking-wider',
                result.valid ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]',
              )}
            >
              {result.valid ? 'Verified' : 'Not Verified'}
            </span>
          </div>

          {result.message && (
            <p className="text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
              {result.message}
            </p>
          )}

          <div className="space-y-2 text-xs">
            <div className="flex justify-between">
              <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">Commit</span>
              <span className="font-mono text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">{result.commit_hash}</span>
            </div>
            {result.merkle_root && (
              <div className="flex justify-between">
                <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">Merkle Root</span>
                <span className="font-mono text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] break-all text-right max-w-xs">{result.merkle_root}</span>
              </div>
            )}
            {result.anchored_at && (
              <div className="flex justify-between">
                <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">Anchored At</span>
                <span className="text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">{toDisplayDateTime(result.anchored_at)}</span>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function AnchorHistoryPanel({
  records,
  loading,
  error,
  onRetry,
  page,
  totalPages,
  onPageChange,
}: {
  records: AnchorRecord[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  page: number;
  totalPages: number;
  onPageChange: (p: number) => void;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={6} /></div>;
  if (error) return <ErrorState message="Failed to load anchor history" onRetry={onRetry} />;
  if (records.length === 0)
    return (
      <EmptyState
        icon={<History size={32} strokeWidth={1.5} />}
        title="No anchor records"
        subtitle="Anchor records will appear after the first anchoring operation"
      />
    );

  return (
    <div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
              {['Commit', 'Merkle Root', 'Backend', 'Transaction', 'Anchored At'].map((h) => (
                <th key={h} className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {records.map((rec) => (
              <tr
                key={rec.id}
                className="border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50 last:border-0 text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
              >
                <td className="px-4 py-2.5 font-mono text-xs">{rec.commit_hash.substring(0, 12)}</td>
                <td className="px-4 py-2.5 font-mono text-[10px] max-w-[200px] truncate">{rec.merkle_root}</td>
                <td className="px-4 py-2.5 font-mono text-xs uppercase">{rec.backend}</td>
                <td className="px-4 py-2.5 font-mono text-[10px] max-w-[150px] truncate">{rec.transaction_id ?? '--'}</td>
                <td className="px-4 py-2.5 text-xs text-[var(--color-muted)]">{toDisplayDateTime(rec.anchored_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {totalPages > 1 && (
        <PaginationControls page={page} totalPages={totalPages} onPageChange={onPageChange} />
      )}
    </div>
  );
}

function BackendsPanel({
  backends,
  loading,
  error,
  onRetry,
}: {
  backends: BackendInfo[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={3} /></div>;
  if (error) return <ErrorState message="Failed to load backends" onRetry={onRetry} />;
  if (backends.length === 0)
    return (
      <EmptyState
        icon={<Server size={32} strokeWidth={1.5} />}
        title="No backends configured"
        subtitle="Configure anchor backends in the server configuration"
      />
    );

  return (
    <div className="divide-y divide-[var(--color-border)]/50 dark:divide-[var(--color-border-dark)]/50">
      {backends.map((be) => (
        <div key={be.name} className="flex items-center justify-between p-4">
          <div className="flex items-center gap-3">
            <Server size={16} className="text-[var(--color-muted)]" />
            <div>
              <p className="text-sm font-semibold text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] font-mono uppercase">
                {be.name}
              </p>
              {be.error_message && (
                <p className="text-xs text-[var(--color-error)] mt-0.5">{be.error_message}</p>
              )}
            </div>
          </div>
          <StatusIndicator
            status={be.connected ? 'connected' : 'disconnected'}
            label={be.connected ? 'Connected' : 'Disconnected'}
          />
        </div>
      ))}
    </div>
  );
}
