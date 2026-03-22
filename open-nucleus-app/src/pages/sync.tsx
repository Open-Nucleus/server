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
  SeverityBadge,
  PaginationControls,
} from '@/components';
import { cn } from '@/lib/utils';
import { timeAgo, toDisplayDateTime } from '@/lib/date-utils';
import { RefreshCw, Users, AlertTriangle, Clock } from 'lucide-react';
import type { SyncStatusResponse, PeerInfo, ConflictDetail } from '@/types';

type SyncTab = 'status' | 'peers' | 'conflicts' | 'history';

interface SyncEvent {
  id: string;
  type: string;
  peer_node_id?: string;
  status: string;
  started_at: string;
  completed_at?: string;
  changes_pushed?: number;
  changes_pulled?: number;
  error_message?: string;
}

export default function SyncPage() {
  const setPageTitle = useUIStore((s) => s.setPageTitle);
  useEffect(() => setPageTitle('Sync & Conflicts'), [setPageTitle]);

  const [activeTab, setActiveTab] = useState<SyncTab>('status');
  const [historyPage, setHistoryPage] = useState(1);
  const queryClient = useQueryClient();

  /* ---- sync status ---- */
  const statusQuery = useQuery({
    queryKey: ['sync', 'status'],
    queryFn: () => apiGet<SyncStatusResponse>(API.sync.status),
    refetchInterval: 5_000,
  });

  /* ---- peers ---- */
  const peersQuery = useQuery({
    queryKey: ['sync', 'peers'],
    queryFn: () => apiGet<PeerInfo[]>(API.sync.peers),
    refetchInterval: 10_000,
    enabled: activeTab === 'peers' || activeTab === 'status',
  });

  /* ---- conflicts ---- */
  const conflictsQuery = useQuery({
    queryKey: ['conflicts'],
    queryFn: () => apiGet<ConflictDetail[]>(API.conflicts.list),
    enabled: activeTab === 'conflicts',
  });

  /* ---- history ---- */
  const historyQuery = useQuery({
    queryKey: ['sync', 'history', historyPage],
    queryFn: () =>
      apiGet<SyncEvent[]>(API.sync.history, {
        page: String(historyPage),
        per_page: '20',
      }),
    enabled: activeTab === 'history',
  });

  /* ---- trigger sync ---- */
  const triggerMutation = useMutation({
    mutationFn: () => apiPost<void>(API.sync.trigger, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sync'] });
    },
  });

  /* ---- resolve / defer conflict ---- */
  const resolveMutation = useMutation({
    mutationFn: (id: string) => apiPost<void>(API.conflicts.resolve(id), {}),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['conflicts'] }),
  });

  const deferMutation = useMutation({
    mutationFn: (id: string) => apiPost<void>(API.conflicts.defer(id), {}),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['conflicts'] }),
  });

  const status = statusQuery.data?.data;
  const peers = peersQuery.data?.data ?? [];
  const conflicts = conflictsQuery.data?.data ?? [];
  const historyEvents = historyQuery.data?.data ?? [];
  const historyPagination = historyQuery.data?.pagination;

  const tabs: { key: SyncTab; label: string; icon: React.ReactNode }[] = [
    { key: 'status', label: 'Status', icon: <RefreshCw size={14} /> },
    { key: 'peers', label: 'Peers', icon: <Users size={14} /> },
    { key: 'conflicts', label: 'Conflicts', icon: <AlertTriangle size={14} /> },
    { key: 'history', label: 'History', icon: <Clock size={14} /> },
  ];

  return (
    <div className="page-padding space-y-4">
      {/* header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-mono text-lg font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            Sync & Conflicts
          </h1>
          <p className="text-xs text-[var(--color-muted)] mt-1">
            Manage peer synchronization and resolve merge conflicts
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
          <RefreshCw size={12} className={triggerMutation.isPending ? 'animate-spin' : ''} />
          {triggerMutation.isPending ? 'Syncing...' : 'Trigger Sync'}
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
          <StatusPanel status={status} loading={statusQuery.isLoading} error={statusQuery.isError} />
        )}

        {activeTab === 'peers' && (
          <PeersPanel
            peers={peers}
            loading={peersQuery.isLoading}
            error={peersQuery.isError}
            onRetry={() => peersQuery.refetch()}
          />
        )}

        {activeTab === 'conflicts' && (
          <ConflictsPanel
            conflicts={conflicts}
            loading={conflictsQuery.isLoading}
            error={conflictsQuery.isError}
            onRetry={() => conflictsQuery.refetch()}
            onResolve={(id) => resolveMutation.mutate(id)}
            onDefer={(id) => deferMutation.mutate(id)}
            resolvingId={resolveMutation.isPending ? (resolveMutation.variables as string) : null}
            deferringId={deferMutation.isPending ? (deferMutation.variables as string) : null}
          />
        )}

        {activeTab === 'history' && (
          <HistoryPanel
            events={historyEvents}
            loading={historyQuery.isLoading}
            error={historyQuery.isError}
            onRetry={() => historyQuery.refetch()}
            page={historyPage}
            totalPages={historyPagination?.total_pages ?? 1}
            onPageChange={setHistoryPage}
          />
        )}
      </div>
    </div>
  );
}

/* ================================================================
   Sub-panels
   ================================================================ */

function StatusPanel({
  status,
  loading,
  error,
}: {
  status?: SyncStatusResponse;
  loading: boolean;
  error: boolean;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={4} /></div>;
  if (error || !status) return <div className="p-6"><ErrorState message="Failed to load sync status" /></div>;

  const rows: { label: string; value: React.ReactNode }[] = [
    {
      label: 'State',
      value: (
        <StatusIndicator
          status={status.state === 'idle' ? 'active' : status.state === 'syncing' ? 'pending' : 'inactive'}
          label={status.state}
        />
      ),
    },
    { label: 'Last Sync', value: status.last_sync ? toDisplayDateTime(status.last_sync) : 'Never' },
    { label: 'Pending Changes', value: String(status.pending_changes) },
    { label: 'Node ID', value: <span className="font-mono text-xs">{status.node_id}</span> },
    { label: 'Site ID', value: <span className="font-mono text-xs">{status.site_id}</span> },
  ];

  return (
    <div className="p-6 space-y-4">
      {rows.map((row) => (
        <div key={row.label} className="flex items-center justify-between py-2 border-b border-[var(--color-border)]/30 dark:border-[var(--color-border-dark)]/30 last:border-0">
          <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
            {row.label}
          </span>
          <span className="text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            {row.value}
          </span>
        </div>
      ))}
    </div>
  );
}

function PeersPanel({
  peers,
  loading,
  error,
  onRetry,
}: {
  peers: PeerInfo[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={4} /></div>;
  if (error) return <ErrorState message="Failed to load peers" onRetry={onRetry} />;
  if (peers.length === 0)
    return (
      <EmptyState
        icon={<Users size={32} strokeWidth={1.5} />}
        title="No peers discovered"
        subtitle="Peers will appear when other nodes are reachable via Wi-Fi Direct, Bluetooth, or local network"
      />
    );

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
            {['Node ID', 'Site ID', 'State', 'Last Seen', 'Latency'].map((h) => (
              <th key={h} className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {peers.map((peer) => (
            <tr
              key={peer.node_id}
              className="border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50 last:border-0 text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
            >
              <td className="px-4 py-2.5 font-mono text-xs">{peer.node_id}</td>
              <td className="px-4 py-2.5 font-mono text-xs">{peer.site_id}</td>
              <td className="px-4 py-2.5">
                <StatusIndicator
                  status={peer.state === 'connected' ? 'connected' : peer.state === 'syncing' ? 'pending' : 'disconnected'}
                  label={peer.state}
                  size="sm"
                />
              </td>
              <td className="px-4 py-2.5 text-xs text-[var(--color-muted)]">
                {peer.last_seen ? timeAgo(peer.last_seen) : '--'}
              </td>
              <td className="px-4 py-2.5 font-mono text-xs">
                {peer.latency_ms != null ? `${peer.latency_ms}ms` : '--'}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ConflictsPanel({
  conflicts,
  loading,
  error,
  onRetry,
  onResolve,
  onDefer,
  resolvingId,
  deferringId,
}: {
  conflicts: ConflictDetail[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  onResolve: (id: string) => void;
  onDefer: (id: string) => void;
  resolvingId: string | null;
  deferringId: string | null;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={4} /></div>;
  if (error) return <ErrorState message="Failed to load conflicts" onRetry={onRetry} />;
  if (conflicts.length === 0)
    return (
      <EmptyState
        icon={<AlertTriangle size={32} strokeWidth={1.5} />}
        title="No conflicts"
        subtitle="All merge conflicts have been resolved"
      />
    );

  return (
    <div className="divide-y divide-[var(--color-border)]/50 dark:divide-[var(--color-border-dark)]/50">
      {conflicts.map((c) => (
        <div key={c.id} className="p-4 space-y-2">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <SeverityBadge severity={mapSeverity(c.severity)} />
              <span className="font-mono text-xs uppercase tracking-wider text-[var(--color-muted)]">
                {c.type}
              </span>
            </div>
            <span className="font-mono text-[10px] text-[var(--color-muted)]">{c.id}</span>
          </div>

          <p className="text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            {c.resources.length} affected resource{c.resources.length !== 1 ? 's' : ''}
          </p>

          {c.suggestions.length > 0 && (
            <ul className="text-xs text-[var(--color-muted)] list-disc list-inside">
              {c.suggestions.map((s, i) => (
                <li key={i}>{s}</li>
              ))}
            </ul>
          )}

          <div className="flex gap-2 pt-1">
            <button
              type="button"
              onClick={() => onResolve(c.id)}
              disabled={resolvingId === c.id}
              className={cn(
                'px-3 py-1.5 text-[10px] font-mono font-semibold uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
                'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]',
                'hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed transition-opacity',
              )}
            >
              {resolvingId === c.id ? 'Resolving...' : 'Resolve'}
            </button>
            <button
              type="button"
              onClick={() => onDefer(c.id)}
              disabled={deferringId === c.id}
              className={cn(
                'px-3 py-1.5 text-[10px] font-mono font-semibold uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
                'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
                'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
                'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
                'disabled:opacity-40 disabled:cursor-not-allowed transition-colors',
              )}
            >
              {deferringId === c.id ? 'Deferring...' : 'Defer'}
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}

function HistoryPanel({
  events,
  loading,
  error,
  onRetry,
  page,
  totalPages,
  onPageChange,
}: {
  events: SyncEvent[];
  loading: boolean;
  error: boolean;
  onRetry: () => void;
  page: number;
  totalPages: number;
  onPageChange: (p: number) => void;
}) {
  if (loading) return <div className="p-6"><LoadingSkeleton count={6} /></div>;
  if (error) return <ErrorState message="Failed to load sync history" onRetry={onRetry} />;
  if (events.length === 0)
    return (
      <EmptyState
        icon={<Clock size={32} strokeWidth={1.5} />}
        title="No sync history"
        subtitle="Sync events will appear here after the first synchronization"
      />
    );

  return (
    <div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
              {['Type', 'Peer', 'Status', 'Pushed', 'Pulled', 'Started', 'Duration'].map((h) => (
                <th key={h} className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {events.map((evt) => (
              <tr
                key={evt.id}
                className="border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50 last:border-0 text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
              >
                <td className="px-4 py-2.5 font-mono text-xs uppercase">{evt.type}</td>
                <td className="px-4 py-2.5 font-mono text-xs">{evt.peer_node_id ?? '--'}</td>
                <td className="px-4 py-2.5">
                  <StatusIndicator
                    status={evt.status === 'completed' ? 'active' : evt.status === 'failed' ? 'error' : 'pending'}
                    label={evt.status}
                    size="sm"
                  />
                </td>
                <td className="px-4 py-2.5 font-mono text-xs">{evt.changes_pushed ?? '--'}</td>
                <td className="px-4 py-2.5 font-mono text-xs">{evt.changes_pulled ?? '--'}</td>
                <td className="px-4 py-2.5 text-xs text-[var(--color-muted)]">
                  {toDisplayDateTime(evt.started_at)}
                </td>
                <td className="px-4 py-2.5 font-mono text-xs text-[var(--color-muted)]">
                  {evt.completed_at
                    ? `${Math.round((new Date(evt.completed_at).getTime() - new Date(evt.started_at).getTime()) / 1000)}s`
                    : '--'}
                </td>
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

/* ---------- helper ---------- */
function mapSeverity(s: string): 'critical' | 'high' | 'warning' | 'moderate' | 'info' | 'low' {
  const map: Record<string, 'critical' | 'high' | 'warning' | 'moderate' | 'info' | 'low'> = {
    critical: 'critical',
    high: 'high',
    warning: 'warning',
    moderate: 'moderate',
    medium: 'moderate',
    info: 'info',
    low: 'low',
  };
  return map[s.toLowerCase()] ?? 'info';
}
