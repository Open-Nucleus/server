import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { useUIStore } from '@/stores/ui-store';
import { LoadingSkeleton, EmptyState, ErrorState } from '@/components';
import { cn } from '@/lib/utils';
import { Sun, Moon, Globe, Info } from 'lucide-react';
import type { SmartClient } from '@/types';

export default function SettingsPage() {
  const setPageTitle = useUIStore((s) => s.setPageTitle);
  useEffect(() => setPageTitle('Settings'), [setPageTitle]);

  const { theme, toggleTheme } = useUIStore();

  /* ---- SMART clients ---- */
  const smartQuery = useQuery({
    queryKey: ['smart', 'clients'],
    queryFn: () => apiGet<SmartClient[]>(API.smartClients.list),
  });

  const smartClients = smartQuery.data?.data ?? [];

  return (
    <div className="page-padding space-y-6 max-w-3xl">
      {/* header */}
      <div>
        <h1 className="font-mono text-lg font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
          Settings
        </h1>
        <p className="text-xs text-[var(--color-muted)] mt-1">
          Application preferences and configuration
        </p>
      </div>

      {/* ---- Appearance ---- */}
      <section className="rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)] bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]">
        <div className="px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
          <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            Appearance
          </h2>
        </div>
        <div className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {theme === 'light' ? (
                <Sun size={18} className="text-[var(--color-muted)]" />
              ) : (
                <Moon size={18} className="text-[var(--color-muted)]" />
              )}
              <div>
                <p className="text-sm font-medium text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  Theme
                </p>
                <p className="text-xs text-[var(--color-muted)]">
                  Currently using {theme} mode
                </p>
              </div>
            </div>
            <button
              onClick={toggleTheme}
              className={cn(
                'relative w-12 h-6 rounded-full transition-colors duration-200 cursor-pointer',
                theme === 'dark'
                  ? 'bg-[var(--color-ink)]'
                  : 'bg-[var(--color-border)]',
              )}
              role="switch"
              aria-checked={theme === 'dark'}
              aria-label="Toggle dark mode"
            >
              <span
                className={cn(
                  'absolute top-0.5 w-5 h-5 rounded-full transition-transform duration-200',
                  'bg-[var(--color-surface)] dark:bg-[var(--color-sidebar-text)]',
                  theme === 'dark' ? 'translate-x-6' : 'translate-x-0.5',
                )}
              />
            </button>
          </div>
        </div>
      </section>

      {/* ---- SMART on FHIR ---- */}
      <section className="rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)] bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]">
        <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
          <div className="flex items-center gap-2">
            <Globe size={14} className="text-[var(--color-muted)]" />
            <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
              SMART on FHIR Clients
            </h2>
          </div>
          <button
            className={cn(
              'px-3 py-1.5 text-[10px] font-mono font-semibold uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
              'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
              'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
              'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
              'transition-colors duration-150',
            )}
          >
            Register Client
          </button>
        </div>

        <div className="overflow-x-auto">
          {smartQuery.isLoading ? (
            <div className="p-4">
              <LoadingSkeleton count={3} />
            </div>
          ) : smartQuery.isError ? (
            <ErrorState
              message="Failed to load SMART clients"
              onRetry={() => smartQuery.refetch()}
            />
          ) : smartClients.length === 0 ? (
            <EmptyState
              icon={<Globe size={32} strokeWidth={1.5} />}
              title="No SMART clients"
              subtitle="Register a SMART on FHIR client to enable third-party app access"
            />
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
                  {['Client ID', 'Name', 'Redirect URI', 'Scope'].map((h) => (
                    <th
                      key={h}
                      className="px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]"
                    >
                      {h}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {smartClients.map((client) => (
                  <tr
                    key={client.id}
                    className="border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50 last:border-0 text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
                  >
                    <td className="px-4 py-2.5 font-mono text-xs">{client.id}</td>
                    <td className="px-4 py-2.5 text-sm">{client.name}</td>
                    <td className="px-4 py-2.5 font-mono text-xs text-[var(--color-muted)] max-w-[200px] truncate">
                      {client.redirect_uri ?? '--'}
                    </td>
                    <td className="px-4 py-2.5 font-mono text-[10px] text-[var(--color-muted)] max-w-[200px] truncate">
                      {client.scope ?? '--'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </section>

      {/* ---- About ---- */}
      <section className="rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)] bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
          <Info size={14} className="text-[var(--color-muted)]" />
          <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
            About
          </h2>
        </div>
        <div className="p-4 space-y-3">
          <AboutRow label="Application" value="Open Nucleus" />
          <AboutRow label="Version" value="0.1.0" />
          <AboutRow label="Platform" value="Tauri + React" />
          <AboutRow label="FHIR Version" value="R4 (4.0.1)" />
          <AboutRow label="License" value="AGPLv3" />
          <AboutRow
            label="Repository"
            value={
              <span className="font-mono text-xs text-[var(--color-muted)]">
                github.com/open-nucleus
              </span>
            }
          />
        </div>
      </section>
    </div>
  );
}

/* ---------- helper ---------- */
function AboutRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between py-1.5 border-b border-[var(--color-border)]/30 dark:border-[var(--color-border-dark)]/30 last:border-0">
      <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
        {label}
      </span>
      <span className="text-sm text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
        {value}
      </span>
    </div>
  );
}
