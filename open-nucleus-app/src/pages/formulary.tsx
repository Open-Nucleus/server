import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { apiGet } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { useUIStore } from '@/stores/ui-store';
import { LoadingSkeleton, EmptyState, ErrorState } from '@/components';
import { cn } from '@/lib/utils';
import { Pill, Star, Package } from 'lucide-react';
import type { MedicationDetail } from '@/types';

/* ---------- hardcoded category list (mirrors common WHO EML groups) ---------- */
const CATEGORIES = [
  'Analgesics',
  'Antibiotics',
  'Antimalarials',
  'Antiretrovirals',
  'Cardiovascular',
  'Dermatological',
  'Diabetic',
  'Gastrointestinal',
  'Ophthalmic',
  'Psychotropic',
  'Respiratory',
  'Vaccines',
  'Vitamins',
] as const;

export default function FormularyPage() {
  const setPageTitle = useUIStore((s) => s.setPageTitle);
  useEffect(() => setPageTitle('Formulary'), [setPageTitle]);

  const [selectedCategory, setSelectedCategory] = useState<string>(CATEGORIES[0]);
  const [selectedMedication, setSelectedMedication] = useState<MedicationDetail | null>(null);

  /* ---- medications for selected category ---- */
  const medsQuery = useQuery({
    queryKey: ['formulary', 'medications', selectedCategory],
    queryFn: () =>
      apiGet<MedicationDetail[]>(API.formulary.medicationsByCategory(selectedCategory)),
    enabled: !!selectedCategory,
  });

  const medications = medsQuery.data?.data ?? [];

  /* ---- reset selection when category changes ---- */
  useEffect(() => {
    setSelectedMedication(null);
  }, [selectedCategory]);

  return (
    <div className="page-padding h-full flex flex-col">
      {/* header */}
      <div className="mb-4">
        <h1 className="font-mono text-lg font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
          Formulary
        </h1>
        <p className="text-xs text-[var(--color-muted)] mt-1">
          Browse medications by therapeutic category
        </p>
      </div>

      {/* 3-pane layout */}
      <div className="flex-1 flex gap-0 rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)] overflow-hidden min-h-0">
        {/* ---- Left pane: categories ---- */}
        <div className="w-64 shrink-0 border-r border-[var(--color-border)] dark:border-[var(--color-border-dark)] bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)] overflow-y-auto">
          <div className="px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
            <h2 className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
              Categories
            </h2>
          </div>
          <ul>
            {CATEGORIES.map((cat) => (
              <li key={cat}>
                <button
                  onClick={() => setSelectedCategory(cat)}
                  className={cn(
                    'w-full text-left px-4 py-2.5 text-sm font-mono transition-colors duration-100 cursor-pointer',
                    selectedCategory === cat
                      ? 'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]'
                      : 'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
                  )}
                >
                  {cat}
                </button>
              </li>
            ))}
          </ul>
        </div>

        {/* ---- Middle pane: medication list ---- */}
        <div className="flex-1 min-w-0 border-r border-[var(--color-border)] dark:border-[var(--color-border-dark)] bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)] overflow-y-auto">
          <div className="px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
            <h2 className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
              {selectedCategory} ({medications.length})
            </h2>
          </div>

          {medsQuery.isLoading ? (
            <div className="p-4">
              <LoadingSkeleton count={6} />
            </div>
          ) : medsQuery.isError ? (
            <ErrorState
              message="Failed to load medications"
              details={medsQuery.error instanceof Error ? medsQuery.error.message : undefined}
              onRetry={() => medsQuery.refetch()}
            />
          ) : medications.length === 0 ? (
            <EmptyState
              icon={<Pill size={32} strokeWidth={1.5} />}
              title="No medications"
              subtitle={`No medications found in ${selectedCategory}`}
            />
          ) : (
            <ul>
              {medications.map((med) => (
                <li key={med.code}>
                  <button
                    onClick={() => setSelectedMedication(med)}
                    className={cn(
                      'w-full text-left px-4 py-3 border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50 transition-colors duration-100 cursor-pointer',
                      selectedMedication?.code === med.code
                        ? 'bg-[var(--color-surface-hover)] dark:bg-[var(--color-surface-dark-hover)]'
                        : 'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
                    )}
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                        {med.display}
                      </span>
                      <div className="flex items-center gap-2">
                        {med.who_essential && (
                          <Star size={12} className="text-[var(--color-warning)] fill-[var(--color-warning)]" />
                        )}
                        <span
                          className={cn(
                            'inline-block w-2 h-2 rounded-full',
                            med.available ? 'bg-[var(--color-success)]' : 'bg-[var(--color-error)]',
                          )}
                        />
                      </div>
                    </div>
                    <span className="text-[10px] font-mono text-[var(--color-muted)] mt-0.5 block">
                      {med.code}
                      {med.strength ? ` / ${med.strength}` : ''}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* ---- Right pane: medication detail ---- */}
        <div className="w-80 shrink-0 bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)] overflow-y-auto">
          <div className="px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
            <h2 className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)]">
              Detail
            </h2>
          </div>

          {selectedMedication ? (
            <div className="p-4 space-y-4">
              {/* Name + code */}
              <div>
                <h3 className="text-base font-bold text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
                  {selectedMedication.display}
                </h3>
                <p className="font-mono text-xs text-[var(--color-muted)] mt-0.5">
                  {selectedMedication.code}
                </p>
              </div>

              {/* WHO Essential badge */}
              {selectedMedication.who_essential && (
                <div className="flex items-center gap-2 px-3 py-2 rounded-[var(--radius-sm)] border border-[var(--color-warning)]/40 bg-[var(--color-warning)]/5">
                  <Star size={14} className="text-[var(--color-warning)] fill-[var(--color-warning)]" />
                  <span className="text-xs font-mono font-semibold uppercase tracking-wider text-[var(--color-warning)]">
                    WHO Essential Medicine
                  </span>
                </div>
              )}

              {/* Properties */}
              <div className="space-y-3">
                <DetailRow label="Form" value={selectedMedication.form} />
                <DetailRow label="Route" value={selectedMedication.route} />
                <DetailRow label="Strength" value={selectedMedication.strength} />
                <DetailRow label="Unit" value={selectedMedication.unit} />
                <DetailRow label="Category" value={selectedMedication.category} />
                <DetailRow label="Therapeutic Class" value={selectedMedication.therapeutic_class} />
                <DetailRow
                  label="Availability"
                  value={
                    <span className="flex items-center gap-1.5">
                      <span
                        className={cn(
                          'w-2 h-2 rounded-full',
                          selectedMedication.available ? 'bg-[var(--color-success)]' : 'bg-[var(--color-error)]',
                        )}
                      />
                      {selectedMedication.available ? 'In Stock' : 'Out of Stock'}
                    </span>
                  }
                />
              </div>

              {/* Common frequencies */}
              {selectedMedication.common_frequencies && selectedMedication.common_frequencies.length > 0 && (
                <div>
                  <p className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)] mb-2">
                    Common Frequencies
                  </p>
                  <div className="flex flex-wrap gap-1.5">
                    {selectedMedication.common_frequencies.map((freq) => (
                      <span
                        key={freq}
                        className="px-2 py-1 text-[10px] font-mono rounded-[var(--radius-sm)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)] text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
                      >
                        {freq}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center h-full py-16 text-center px-4">
              <Package size={32} strokeWidth={1.5} className="text-[var(--color-muted)] opacity-40 mb-3" />
              <p className="font-mono text-xs uppercase tracking-wider text-[var(--color-muted)]">
                Select a medication
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

/* ---------- helper ---------- */

function DetailRow({ label, value }: { label: string; value?: React.ReactNode }) {
  if (value === undefined || value === null || value === '') return null;
  return (
    <div className="flex justify-between items-start gap-2">
      <span className="font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)] shrink-0">
        {label}
      </span>
      <span className="text-sm text-right text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
        {value}
      </span>
    </div>
  );
}
