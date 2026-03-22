import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { useQuery } from '@tanstack/react-query';
import { Users, Plus, ChevronDown, ChevronUp } from 'lucide-react';
import { apiGet } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { useUIStore } from '@/stores/ui-store';
import { timeAgo } from '@/lib/date-utils';
import { capitalize } from '@/lib/string-utils';
import { cn } from '@/lib/utils';
import { ADMINISTRATIVE_GENDERS } from '@/lib/fhir-codes';
import {
  DataTableCard,
  PaginationControls,
  StatusIndicator,
} from '@/components';
import type { ApiEnvelope, Pagination } from '@/types';

/* ---------- types ---------- */

/** Raw FHIR Patient resource as returned by the Go backend. */
type FhirPatient = Record<string, unknown>;

interface Filters {
  gender: string;
  active: string;
}

/* ---------- FHIR helpers ---------- */

function extractName(r: FhirPatient): string {
  const nameArr = r.name as
    | Array<{ use?: string; family?: string; given?: string[] }>
    | undefined;
  if (!nameArr || nameArr.length === 0) return 'Unknown';
  const entry = nameArr[0];
  const parts: string[] = [];
  if (entry.family) parts.push(entry.family);
  if (entry.given?.length) parts.push(entry.given.join(' '));
  return parts.join(', ') || 'Unknown';
}

/* ---------- columns ---------- */

const columns = [
  {
    key: 'name',
    header: 'Name',
    render: (r: FhirPatient) => (
      <span className="font-medium">{extractName(r)}</span>
    ),
  },
  {
    key: 'gender',
    header: 'Gender',
    render: (r: FhirPatient) => (
      <span className="font-mono text-xs">
        {capitalize(String(r.gender ?? '--'))}
      </span>
    ),
  },
  {
    key: 'birthDate',
    header: 'Birth Date',
    render: (r: FhirPatient) => (
      <span className="font-mono text-xs tabular-nums">
        {String(r.birthDate ?? '--')}
      </span>
    ),
  },
  {
    key: 'active',
    header: 'Status',
    render: (r: FhirPatient) => {
      const active = r.active as boolean | undefined;
      return (
        <StatusIndicator
          status={active ? 'active' : 'inactive'}
          label={active ? 'Active' : 'Inactive'}
          size="sm"
        />
      );
    },
  },
  {
    key: 'lastUpdated',
    header: 'Last Updated',
    render: (r: FhirPatient) => {
      const meta = r.meta as { lastUpdated?: string } | undefined;
      return (
        <span className="font-mono text-xs text-[var(--color-muted)]">
          {meta?.lastUpdated ? timeAgo(meta.lastUpdated) : '--'}
        </span>
      );
    },
  },
];

/* ---------- component ---------- */

export default function PatientsListPage() {
  const navigate = useNavigate();
  const setPageTitle = useUIStore((s) => s.setPageTitle);

  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const [filters, setFilters] = useState<Filters>({ gender: '', active: '' });
  const [filtersOpen, setFiltersOpen] = useState(false);

  useEffect(() => {
    setPageTitle('Patients');
  }, [setPageTitle]);

  /* Keyboard shortcut: Ctrl+N -> new patient */
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'n') {
        e.preventDefault();
        navigate({ to: '/patients/new' });
      }
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [navigate]);

  /* Build query params */
  const buildParams = useCallback(() => {
    const params: Record<string, string> = {
      page: String(page),
      per_page: '20',
    };
    if (search.trim()) params.q = search.trim();
    if (filters.gender) params.gender = filters.gender;
    if (filters.active) params.active = filters.active;
    return params;
  }, [page, search, filters]);

  /* Fetch patients */
  const {
    data: envelope,
    isLoading,
    error,
    refetch,
  } = useQuery<ApiEnvelope<FhirPatient[]>>({
    queryKey: ['patients', page, search, filters],
    queryFn: () => {
      const params = buildParams();
      const path = search.trim() ? API.patients.search : API.patients.list;
      return apiGet<FhirPatient[]>(path, params);
    },
  });

  const patients: FhirPatient[] = envelope?.data ?? [];
  const pagination: Pagination | undefined = envelope?.pagination;
  const totalPages = pagination?.total_pages ?? 1;

  /* Handlers */
  const handleSearchChange = (value: string) => {
    setSearch(value);
    setPage(1);
  };

  const handleRowClick = (r: FhirPatient) => {
    navigate({ to: '/patients/$id', params: { id: String(r.id) } });
  };

  const handleFilterChange = (key: keyof Filters, value: string) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1);
  };

  return (
    <div className="page-padding space-y-4">
      {/* Header row */}
      <div className="flex items-center justify-between">
        <h1 className="font-mono text-lg font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
          Patients
        </h1>
        <button
          type="button"
          onClick={() => navigate({ to: '/patients/new' })}
          className={cn(
            'inline-flex items-center gap-2 px-4 py-2 text-xs font-mono uppercase tracking-wider cursor-pointer',
            'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]',
            'hover:opacity-90 transition-opacity duration-150 rounded-[var(--radius-sm)]',
          )}
        >
          <Plus size={14} />
          New Patient
        </button>
      </div>

      {/* Filter panel (collapsible) */}
      <div
        className={cn(
          'rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
          'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
        )}
      >
        <button
          type="button"
          onClick={() => setFiltersOpen((v) => !v)}
          className={cn(
            'flex w-full items-center justify-between px-4 py-2.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
            'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
            'transition-colors duration-150',
          )}
        >
          Filters
          {filtersOpen ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
        </button>

        {filtersOpen && (
          <div className="flex flex-wrap gap-4 px-4 pb-4 border-t border-[var(--color-border)] dark:border-[var(--color-border-dark)] pt-3">
            {/* Gender */}
            <label className="flex flex-col gap-1">
              <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
                Gender
              </span>
              <select
                value={filters.gender}
                onChange={(e) => handleFilterChange('gender', e.target.value)}
                className={cn(
                  'px-3 py-1.5 text-sm font-mono rounded-[var(--radius-sm)]',
                  'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
                  'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
                  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                  'outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
                )}
              >
                <option value="">All</option>
                {ADMINISTRATIVE_GENDERS.map((g) => (
                  <option key={g} value={g}>
                    {capitalize(g)}
                  </option>
                ))}
              </select>
            </label>

            {/* Status */}
            <label className="flex flex-col gap-1">
              <span className="font-mono text-[10px] uppercase tracking-wider text-[var(--color-muted)]">
                Status
              </span>
              <select
                value={filters.active}
                onChange={(e) => handleFilterChange('active', e.target.value)}
                className={cn(
                  'px-3 py-1.5 text-sm font-mono rounded-[var(--radius-sm)]',
                  'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
                  'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
                  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
                  'outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
                )}
              >
                <option value="">All</option>
                <option value="true">Active</option>
                <option value="false">Inactive</option>
              </select>
            </label>

            {/* Clear */}
            {(filters.gender || filters.active) && (
              <button
                type="button"
                onClick={() => setFilters({ gender: '', active: '' })}
                className={cn(
                  'self-end px-3 py-1.5 text-xs font-mono uppercase tracking-wider cursor-pointer',
                  'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
                  'transition-colors duration-150',
                )}
              >
                Clear Filters
              </button>
            )}
          </div>
        )}
      </div>

      {/* Data table */}
      <DataTableCard<FhirPatient>
        title="Patient Registry"
        columns={columns}
        data={patients}
        keyExtractor={(r) => String(r.id)}
        onRowClick={handleRowClick}
        searchValue={search}
        onSearchChange={handleSearchChange}
        searchPlaceholder="Search patients..."
        loading={isLoading}
        error={error ? (error as Error).message : undefined}
        onRetry={() => refetch()}
        emptyIcon={<Users size={36} strokeWidth={1.5} />}
        emptyTitle="No patients found"
        emptySubtitle={
          search
            ? `No results for "${search}". Try a different search term.`
            : 'Register your first patient to get started.'
        }
      />

      {/* Pagination */}
      {totalPages > 1 && (
        <PaginationControls
          page={page}
          totalPages={totalPages}
          onPageChange={setPage}
        />
      )}
    </div>
  );
}
