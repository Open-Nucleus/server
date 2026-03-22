import { useEffect } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import {
  Monitor,
  Users,
  Bell,
  RefreshCw,
  Shield,
  Zap,
  Activity,
} from 'lucide-react';
import { useUIStore } from '@/stores/ui-store';
import { useAuthStore } from '@/stores/auth-store';
import { apiGet, apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { timeAgo } from '@/lib/date-utils';
import { cn } from '@/lib/utils';
import { LoadingSkeleton } from '@/components/loading-skeleton';
import { ErrorState } from '@/components/error-state';
import { StatusIndicator } from '@/components/status-indicator';
import { useConnection } from '@/hooks/use-connection';
import type { PatientSummary } from '@/types/patient';
import type { AlertSummary } from '@/types/alert';
import type { SyncStatusResponse } from '@/types/sync';
import type { AnchorStatus } from '@/types/anchor';

// ---------------------------------------------------------------------------
// Card wrapper
// ---------------------------------------------------------------------------

function Card({
  title,
  icon,
  children,
  className,
}: {
  title: string;
  icon: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        'rounded-[var(--radius-md)] typewriter-border bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
        className,
      )}
    >
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
        <span className="text-[var(--color-muted)]">{icon}</span>
        <h2 className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
          {title}
        </h2>
      </div>
      <div className="card-padding">{children}</div>
    </div>
  );
}

function HealthRow({ label, status, statusLabel }: { label: string; status: 'active' | 'inactive' | 'pending' | 'error'; statusLabel: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">{label}</span>
      <StatusIndicator status={status} label={statusLabel} size="sm" />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Dashboard Page
// ---------------------------------------------------------------------------

export default function DashboardPage() {
  const setPageTitle = useUIStore((s) => s.setPageTitle);
  const setAlertCount = useUIStore((s) => s.setAlertCount);
  const navigate = useNavigate();

  const { nodeId, siteId, deviceId } = useAuthStore();

  useEffect(() => {
    setPageTitle('Dashboard');
  }, [setPageTitle]);

  // ---- Patient count ----
  const patients = useQuery({
    queryKey: ['dashboard', 'patientCount'],
    queryFn: () =>
      apiGet<PatientSummary[]>(API.patients.list, { per_page: '1' }),
  });

  // ---- Alert summary ----
  const alertSummary = useQuery({
    queryKey: ['dashboard', 'alertSummary'],
    queryFn: () => apiGet<AlertSummary>(API.alerts.summary),
  });

  // Update sidebar badge whenever alert summary loads
  useEffect(() => {
    if (alertSummary.data?.data) {
      setAlertCount(alertSummary.data.data.unacknowledged);
    }
  }, [alertSummary.data, setAlertCount]);

  // ---- Sync status ----
  const syncStatus = useQuery({
    queryKey: ['dashboard', 'syncStatus'],
    queryFn: () => apiGet<SyncStatusResponse>(API.sync.status),
  });

  // ---- Anchor status ----
  const anchorStatus = useQuery({
    queryKey: ['dashboard', 'anchorStatus'],
    queryFn: () => apiGet<AnchorStatus>(API.anchor.status),
  });

  // ---- Handlers ----
  const handleSyncNow = async () => {
    try {
      await apiPost(API.sync.trigger, {});
      syncStatus.refetch();
    } catch {
      // handled by query refetch
    }
  };

  const handleAnchorNow = async () => {
    try {
      await apiPost(API.anchor.trigger, {});
      anchorStatus.refetch();
    } catch {
      // handled by query refetch
    }
  };

  // ---- Derived values ----
  const patientTotal = patients.data?.pagination?.total ?? 0;
  const summary = alertSummary.data?.data;
  const sync = syncStatus.data?.data;
  const anchor = anchorStatus.data?.data;

  const { connected: apiConnected } = useConnection();
  const syncConnected = sync?.state?.toLowerCase() === 'connected' || sync?.state?.toLowerCase() === 'idle';
  const anchorActive = anchor?.has_been_anchored ?? false;

  return (
    <div className="page-padding space-y-4">
      {/* Row 1: 3 columns */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {/* Node Identity */}
        <Card title="Node Identity" icon={<Monitor size={14} />}>
          <div className="space-y-3">
            <div>
              <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                Node ID
              </p>
              <p className="font-mono text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] break-all">
                {nodeId || '--'}
              </p>
            </div>
            <div>
              <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                Site ID
              </p>
              <p className="font-mono text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] break-all">
                {siteId || '--'}
              </p>
            </div>
            <div>
              <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                Device ID
              </p>
              <p className="font-mono text-xs text-[var(--color-muted)] break-all">
                {deviceId || '--'}
              </p>
            </div>
          </div>
        </Card>

        {/* Patient Statistics */}
        <Card title="Patient Statistics" icon={<Users size={14} />}>
          {patients.isLoading ? (
            <LoadingSkeleton count={2} />
          ) : patients.isError ? (
            <ErrorState
              message="Failed to load patients"
              onRetry={() => patients.refetch()}
            />
          ) : (
            <div className="flex flex-col items-center justify-center py-2">
              <p className="text-5xl font-bold font-mono tabular-nums text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                {patientTotal}
              </p>
              <p className="text-xs text-[var(--color-muted)] mt-1 uppercase tracking-wider font-mono">
                Total Patients
              </p>
              <button
                type="button"
                onClick={() => navigate({ to: '/patients' })}
                className={cn(
                  'mt-4 px-4 py-1.5 text-[10px] font-mono uppercase tracking-wider cursor-pointer',
                  'typewriter-border rounded-[var(--radius-sm)]',
                  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                  'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                  'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                  'transition-colors duration-150',
                )}
              >
                View All
              </button>
            </div>
          )}
        </Card>

        {/* Alert Summary */}
        <Card title="Alert Summary" icon={<Bell size={14} />}>
          {alertSummary.isLoading ? (
            <LoadingSkeleton count={4} />
          ) : alertSummary.isError ? (
            <ErrorState
              message="Failed to load alerts"
              onRetry={() => alertSummary.refetch()}
            />
          ) : summary ? (
            <div className="space-y-2.5">
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Critical
                </span>
                <span className="font-mono text-sm font-bold tabular-nums text-[var(--color-critical)]">
                  {summary.critical}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Warning
                </span>
                <span className="font-mono text-sm font-bold tabular-nums text-[var(--color-moderate)]">
                  {summary.warning}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Info
                </span>
                <span className="font-mono text-sm font-bold tabular-nums text-[var(--color-info)]">
                  {summary.info}
                </span>
              </div>
              <hr className="border-[var(--color-border)] dark:border-[var(--color-border-dark)]" />
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Unacknowledged
                </span>
                <span className="font-mono text-sm font-bold tabular-nums text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  {summary.unacknowledged}
                </span>
              </div>
            </div>
          ) : null}
        </Card>
      </div>

      {/* Row 2: 2 columns */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Sync Status */}
        <Card title="Sync Status" icon={<RefreshCw size={14} />}>
          {syncStatus.isLoading ? (
            <LoadingSkeleton count={3} />
          ) : syncStatus.isError ? (
            <ErrorState
              message="Failed to load sync status"
              onRetry={() => syncStatus.refetch()}
            />
          ) : sync ? (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  State
                </span>
                <StatusIndicator
                  status={syncConnected ? 'connected' : 'disconnected'}
                  label={syncConnected ? 'Connected' : 'Disconnected'}
                  size="sm"
                />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Last Sync
                </span>
                <span className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  {sync.last_sync ? timeAgo(sync.last_sync) : 'Never'}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Pending Changes
                </span>
                <span className="font-mono text-sm font-bold tabular-nums text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  {sync.pending_changes}
                </span>
              </div>
              <button
                type="button"
                onClick={handleSyncNow}
                className={cn(
                  'w-full mt-1 px-4 py-2 text-[10px] font-mono uppercase tracking-wider cursor-pointer',
                  'typewriter-border rounded-[var(--radius-sm)]',
                  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                  'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                  'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                  'transition-colors duration-150',
                )}
              >
                Sync Now
              </button>
            </div>
          ) : null}
        </Card>

        {/* Anchor Integrity */}
        <Card title="Anchor Integrity" icon={<Shield size={14} />}>
          {anchorStatus.isLoading ? (
            <LoadingSkeleton count={4} />
          ) : anchorStatus.isError ? (
            <ErrorState
              message="Failed to load anchor status"
              onRetry={() => anchorStatus.refetch()}
            />
          ) : anchor ? (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Status
                </span>
                <StatusIndicator
                  status={anchorActive ? 'active' : 'inactive'}
                  label={anchorActive ? 'Anchored' : 'Not Anchored'}
                  size="sm"
                />
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Last Anchor
                </span>
                <span className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  {anchor.last_anchor_time
                    ? timeAgo(anchor.last_anchor_time)
                    : 'Never'}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Merkle Root
                </span>
                <span
                  className="font-mono text-[10px] text-[var(--color-muted)] truncate max-w-[160px]"
                  title={anchor.merkle_root}
                >
                  {anchor.merkle_root
                    ? `${anchor.merkle_root.slice(0, 8)}...${anchor.merkle_root.slice(-8)}`
                    : '--'}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase tracking-wider text-[var(--color-muted)]">
                  Queue Depth
                </span>
                <span className="font-mono text-sm font-bold tabular-nums text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  {anchor.queue_depth ?? 0}
                </span>
              </div>
              <button
                type="button"
                onClick={handleAnchorNow}
                className={cn(
                  'w-full mt-1 px-4 py-2 text-[10px] font-mono uppercase tracking-wider cursor-pointer',
                  'typewriter-border rounded-[var(--radius-sm)]',
                  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                  'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                  'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                  'transition-colors duration-150',
                )}
              >
                Anchor Now
              </button>
            </div>
          ) : null}
        </Card>
      </div>

      {/* Row 3: 2 columns */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Quick Actions */}
        <Card title="Quick Actions" icon={<Zap size={14} />}>
          <div className="grid grid-cols-1 gap-2">
            <button
              type="button"
              onClick={() => navigate({ to: '/patients/new' })}
              className={cn(
                'w-full px-4 py-2.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
                'typewriter-border rounded-[var(--radius-sm)]',
                'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                'transition-colors duration-150',
              )}
            >
              New Patient
            </button>
            <button
              type="button"
              onClick={handleSyncNow}
              className={cn(
                'w-full px-4 py-2.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
                'typewriter-border rounded-[var(--radius-sm)]',
                'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                'transition-colors duration-150',
              )}
            >
              Trigger Sync
            </button>
            <button
              type="button"
              onClick={() => navigate({ to: '/alerts' })}
              className={cn(
                'w-full px-4 py-2.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
                'typewriter-border rounded-[var(--radius-sm)]',
                'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                'transition-colors duration-150',
              )}
            >
              View Alerts
            </button>
          </div>
        </Card>

        {/* System Health */}
        <Card title="System Health" icon={<Activity size={14} />}>
          <div className="space-y-3">
            <HealthRow
              label="API Connection"
              status={apiConnected ? 'active' : 'error'}
              statusLabel={apiConnected ? 'Online' : 'Offline'}
            />
            <HealthRow
              label="Sync Engine"
              status={syncStatus.isLoading ? 'pending' : syncConnected ? 'active' : 'inactive'}
              statusLabel={syncStatus.isLoading ? 'Checking...' : syncConnected ? 'Ready' : 'Idle'}
            />
            <HealthRow
              label="Anchor Service"
              status={anchorStatus.isLoading ? 'pending' : anchorActive ? 'active' : 'inactive'}
              statusLabel={anchorStatus.isLoading ? 'Checking...' : anchorActive ? 'Active' : 'Standby'}
            />
            <HealthRow
              label="Sentinel Agent"
              status="inactive"
              statusLabel="Standby"
            />
          </div>
        </Card>
      </div>
    </div>
  );
}
