import { Fragment, useEffect, useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Bell,
  AlertTriangle,
  Info,
  ShieldAlert,
  CheckCircle,
  XCircle,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { useUIStore } from '@/stores/ui-store';
import { apiGet, apiPost } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { timeAgo } from '@/lib/date-utils';
import { cn } from '@/lib/utils';
import { LoadingSkeleton } from '@/components/loading-skeleton';
import { ErrorState } from '@/components/error-state';
import { SeverityBadge } from '@/components/severity-badge';
import { PaginationControls } from '@/components/pagination-controls';
import { PageHeader } from '@/components/page-header';
import type { AlertDetail, AlertSummary } from '@/types/alert';

// ---------------------------------------------------------------------------
// Severity filter options
// ---------------------------------------------------------------------------

type SeverityFilter = 'all' | 'critical' | 'high' | 'warning' | 'info';

const SEVERITY_OPTIONS: { value: SeverityFilter; label: string }[] = [
  { value: 'all', label: 'All Severities' },
  { value: 'critical', label: 'Critical' },
  { value: 'high', label: 'High' },
  { value: 'warning', label: 'Warning' },
  { value: 'info', label: 'Info' },
];

// ---------------------------------------------------------------------------
// Summary card
// ---------------------------------------------------------------------------

function SummaryCard({
  label,
  count,
  icon,
  colorClass,
}: {
  label: string;
  count: number;
  icon: React.ReactNode;
  colorClass: string;
}) {
  return (
    <div className="rounded-[var(--radius-md)] typewriter-border bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)] card-padding flex items-center gap-3">
      <span className={colorClass}>{icon}</span>
      <div>
        <p className="font-mono text-2xl font-bold tabular-nums text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
          {count}
        </p>
        <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)]">
          {label}
        </p>
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Alerts Page
// ---------------------------------------------------------------------------

export default function AlertsPage() {
  const setAlertCount = useUIStore((s) => s.setAlertCount);
  const queryClient = useQueryClient();

  const [page, setPage] = useState(1);
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all');
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const perPage = 20;


  // ---- Alert summary ----
  const summaryQuery = useQuery({
    queryKey: ['alerts', 'summary'],
    queryFn: () => apiGet<AlertSummary>(API.alerts.summary),
  });

  const summary = summaryQuery.data?.data;

  // Update sidebar badge
  useEffect(() => {
    if (summary) {
      setAlertCount(summary.unacknowledged);
    }
  }, [summary, setAlertCount]);

  // ---- Alert list ----
  const listParams: Record<string, string> = {
    page: String(page),
    per_page: String(perPage),
  };
  if (severityFilter !== 'all') {
    listParams.severity = severityFilter;
  }

  const listQuery = useQuery({
    queryKey: ['alerts', 'list', page, severityFilter],
    queryFn: () => apiGet<AlertDetail[]>(API.alerts.list, listParams),
  });

  const alerts = listQuery.data?.data ?? [];
  const pagination = listQuery.data?.pagination;
  const totalPages = pagination?.total_pages ?? 1;

  // ---- Acknowledge mutation ----
  const acknowledgeMutation = useMutation({
    mutationFn: (alertId: string) =>
      apiPost(API.alerts.acknowledge(alertId), {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'alertSummary'] });
    },
  });

  // ---- Dismiss mutation ----
  const dismissMutation = useMutation({
    mutationFn: (alertId: string) =>
      apiPost(API.alerts.dismiss(alertId), {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'alertSummary'] });
    },
  });

  // ---- Toggle expand ----
  const toggleExpand = (id: string) => {
    setExpandedId((prev) => (prev === id ? null : id));
  };

  // ---- Map severity to badge type ----
  const toBadgeSeverity = (
    s: string,
  ): 'critical' | 'high' | 'warning' | 'moderate' | 'info' | 'low' => {
    const lower = s.toLowerCase();
    if (lower === 'critical') return 'critical';
    if (lower === 'high') return 'high';
    if (lower === 'warning') return 'warning';
    if (lower === 'info') return 'info';
    if (lower === 'low') return 'low';
    return 'info';
  };

  return (
    <div className="page-padding space-y-4">
      <PageHeader
        title="Alerts"
        breadcrumbs={[
          { label: 'Dashboard', path: '/dashboard' },
          { label: 'Alerts' },
        ]}
      />

      {/* Summary cards */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
        {summaryQuery.isLoading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <div
              key={i}
              className="rounded-[var(--radius-md)] typewriter-border bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)] card-padding"
            >
              <LoadingSkeleton count={2} />
            </div>
          ))
        ) : summaryQuery.isError ? (
          <div className="col-span-full">
            <ErrorState
              message="Failed to load alert summary"
              onRetry={() => summaryQuery.refetch()}
            />
          </div>
        ) : summary ? (
          <>
            <SummaryCard
              label="Total"
              count={summary.total}
              icon={<Bell size={20} />}
              colorClass="text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
            />
            <SummaryCard
              label="Critical"
              count={summary.critical}
              icon={<ShieldAlert size={20} />}
              colorClass="text-[var(--color-critical)]"
            />
            <SummaryCard
              label="Warning"
              count={summary.warning}
              icon={<AlertTriangle size={20} />}
              colorClass="text-[var(--color-moderate)]"
            />
            <SummaryCard
              label="Info"
              count={summary.info}
              icon={<Info size={20} />}
              colorClass="text-[var(--color-info)]"
            />
            <SummaryCard
              label="Unacknowledged"
              count={summary.unacknowledged}
              icon={<Bell size={20} />}
              colorClass="text-[var(--color-error)]"
            />
          </>
        ) : null}
      </div>

      {/* Alert table */}
      <div className="rounded-[var(--radius-md)] typewriter-border bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]">
        {/* Table header */}
        <div className="flex items-center justify-between gap-4 px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
          <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-[var(--color-muted)]">
            Alerts
          </h2>
          <select
            value={severityFilter}
            onChange={(e) => {
              setSeverityFilter(e.target.value as SeverityFilter);
              setPage(1);
            }}
            className={cn(
              'px-3 py-1.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
              'typewriter-border rounded-[var(--radius-sm)]',
              'bg-transparent text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
              'focus:outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
            )}
          >
            {SEVERITY_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* Table body */}
        <div className="overflow-x-auto">
          {listQuery.isLoading ? (
            <div className="p-6">
              <LoadingSkeleton count={8} />
            </div>
          ) : listQuery.isError ? (
            <ErrorState
              message="Failed to load alerts"
              onRetry={() => listQuery.refetch()}
            />
          ) : alerts.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
              <div className="mb-4 text-[var(--color-muted)] opacity-50">
                <Bell size={36} />
              </div>
              <h3 className="font-mono text-sm font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                No alerts
              </h3>
              <p className="mt-1.5 text-sm text-[var(--color-muted)] max-w-xs">
                {severityFilter !== 'all'
                  ? `No ${severityFilter} alerts found.`
                  : 'All clear. No alerts have been generated.'}
              </p>
            </div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)] w-8" />
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Severity
                  </th>
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Type
                  </th>
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Message
                  </th>
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Patient
                  </th>
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Time
                  </th>
                  <th className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Status
                  </th>
                  <th className="px-4 py-2.5 text-right font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {alerts.map((alert) => {
                  const isExpanded = expandedId === alert.id;
                  return (
                    <Fragment key={alert.id}>
                      <tr
                        onClick={() => toggleExpand(alert.id)}
                        className={cn(
                          'border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50',
                          'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                          'cursor-pointer hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)] transition-colors duration-100',
                          isExpanded &&
                            'bg-[var(--color-surface-hover)]/50 dark:bg-[var(--color-surface-dark-hover)]/50',
                        )}
                      >
                        <td className="px-4 py-2.5 text-[var(--color-muted)]">
                          {isExpanded ? (
                            <ChevronUp size={14} />
                          ) : (
                            <ChevronDown size={14} />
                          )}
                        </td>
                        <td className="px-4 py-2.5">
                          <SeverityBadge
                            severity={toBadgeSeverity(alert.severity)}
                          />
                        </td>
                        <td className="px-4 py-2.5 font-mono text-xs">
                          {alert.type}
                        </td>
                        <td className="px-4 py-2.5 max-w-[240px] truncate">
                          {alert.message}
                        </td>
                        <td className="px-4 py-2.5 font-mono text-xs text-[var(--color-muted)]">
                          {alert.patient_id
                            ? `${alert.patient_id.slice(0, 8)}...`
                            : '--'}
                        </td>
                        <td className="px-4 py-2.5 font-mono text-xs text-[var(--color-muted)] whitespace-nowrap">
                          {timeAgo(alert.created_at)}
                        </td>
                        <td className="px-4 py-2.5">
                          {alert.acknowledged ? (
                            <span className="inline-flex items-center gap-1 text-[10px] font-mono uppercase tracking-wider text-[var(--color-success)]">
                              <CheckCircle size={12} />
                              Acknowledged
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1 text-[10px] font-mono uppercase tracking-wider text-[var(--color-error)]">
                              <XCircle size={12} />
                              Pending
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-2.5 text-right">
                          <div
                            className="flex items-center justify-end gap-2"
                            onClick={(e) => e.stopPropagation()}
                          >
                            {!alert.acknowledged && (
                              <button
                                type="button"
                                onClick={() =>
                                  acknowledgeMutation.mutate(alert.id)
                                }
                                disabled={acknowledgeMutation.isPending}
                                className={cn(
                                  'px-2.5 py-1 text-[10px] font-mono uppercase tracking-wider cursor-pointer',
                                  'typewriter-border rounded-[var(--radius-sm)]',
                                  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                                  'hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]',
                                  'dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]',
                                  'transition-colors duration-150',
                                  'disabled:opacity-40 disabled:cursor-not-allowed',
                                )}
                              >
                                Ack
                              </button>
                            )}
                            <button
                              type="button"
                              onClick={() =>
                                dismissMutation.mutate(alert.id)
                              }
                              disabled={dismissMutation.isPending}
                              className={cn(
                                'px-2.5 py-1 text-[10px] font-mono uppercase tracking-wider cursor-pointer',
                                'border border-[var(--color-error)] text-[var(--color-error)] rounded-[var(--radius-sm)]',
                                'hover:bg-[var(--color-error)] hover:text-white',
                                'transition-colors duration-150',
                                'disabled:opacity-40 disabled:cursor-not-allowed',
                              )}
                            >
                              Dismiss
                            </button>
                          </div>
                        </td>
                      </tr>

                      {/* Expanded detail row */}
                      {isExpanded && (
                        <tr className="border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50">
                          <td colSpan={8} className="px-4 py-4">
                            <div className="rounded-[var(--radius-sm)] typewriter-border p-4 space-y-2 bg-[var(--color-paper)] dark:bg-[var(--color-paper-dark)]">
                              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                <div>
                                  <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                                    Alert ID
                                  </p>
                                  <p className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] break-all">
                                    {alert.id}
                                  </p>
                                </div>
                                <div>
                                  <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                                    Source
                                  </p>
                                  <p className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                                    {alert.source || '--'}
                                  </p>
                                </div>
                                <div>
                                  <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                                    Patient ID
                                  </p>
                                  <p className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] break-all">
                                    {alert.patient_id || 'N/A'}
                                  </p>
                                </div>
                                <div>
                                  <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                                    Created At
                                  </p>
                                  <p className="font-mono text-xs text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                                    {new Date(alert.created_at).toLocaleString()}
                                  </p>
                                </div>
                              </div>
                              <div>
                                <p className="text-[10px] font-mono uppercase tracking-wider text-[var(--color-muted)] mb-0.5">
                                  Full Message
                                </p>
                                <p className="text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] leading-relaxed">
                                  {alert.message}
                                </p>
                              </div>
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="border-t border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
            <PaginationControls
              page={page}
              totalPages={totalPages}
              onPageChange={setPage}
            />
          </div>
        )}
      </div>
    </div>
  );
}
