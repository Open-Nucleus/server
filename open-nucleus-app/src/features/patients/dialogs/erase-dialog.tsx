import { useEffect, useRef } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiDelete } from '@/lib/api-client';
import { API } from '@/lib/api-paths';
import { cn } from '@/lib/utils';

interface EraseDialogProps {
  open: boolean;
  onClose: () => void;
  patientId: string;
  patientName?: string;
}

/**
 * Crypto-erasure confirmation dialog.
 * Destroys the patient's encryption key, making all clinical data unrecoverable.
 */
export function EraseDialog({ open, onClose, patientId, patientName }: EraseDialogProps) {
  const overlayRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => apiDelete<void>(API.patients.erase(patientId)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['patients'] });
      onClose();
    },
  });

  /* ---- close on Escape ---- */
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !mutation.isPending) onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, onClose, mutation.isPending]);

  /* ---- close on overlay click ---- */
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === overlayRef.current && !mutation.isPending) {
      onClose();
    }
  };

  if (!open) return null;

  return (
    <div
      ref={overlayRef}
      onClick={handleOverlayClick}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
    >
      <div
        className={cn(
          'w-full max-w-md mx-4 p-6 rounded-[var(--radius-lg)]',
          'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
          'border border-[var(--color-error)]/40',
          'shadow-lg',
        )}
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="erase-dialog-title"
        aria-describedby="erase-dialog-description"
      >
        {/* Warning icon */}
        <div className="flex justify-center mb-4">
          <div className="w-12 h-12 rounded-full bg-[var(--color-error)]/10 flex items-center justify-center">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-[var(--color-error)]">
              <path d="M3 6h18" /><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" /><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" /><line x1="10" x2="10" y1="11" y2="17" /><line x1="14" x2="14" y1="11" y2="17" />
            </svg>
          </div>
        </div>

        <h2
          id="erase-dialog-title"
          className="font-mono text-sm font-bold uppercase tracking-wider text-center text-[var(--color-error)]"
        >
          Crypto-Erasure
        </h2>

        <p
          id="erase-dialog-description"
          className="mt-3 text-sm text-center text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] leading-relaxed"
        >
          This will permanently destroy the encryption key for{' '}
          <strong className="font-semibold">{patientName || patientId}</strong>.
          All clinical data will become unrecoverable. This action cannot be undone.
        </p>

        <div className="mt-2 p-3 rounded-[var(--radius-sm)] bg-[var(--color-error)]/5 border border-[var(--color-error)]/20">
          <p className="text-xs text-[var(--color-error)] text-center font-mono leading-relaxed">
            The patient's encryption key will be destroyed. All FHIR resources,
            encounters, observations, conditions, and other clinical data tied
            to this patient will become permanently inaccessible.
          </p>
        </div>

        {mutation.isError && (
          <div className="mt-3 p-2 text-sm text-[var(--color-error)] text-center">
            {mutation.error instanceof Error ? mutation.error.message : 'Erasure failed'}
          </div>
        )}

        <div className="flex justify-end gap-3 mt-6">
          <button
            type="button"
            onClick={onClose}
            disabled={mutation.isPending}
            className={cn(
              'px-4 py-2 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
              'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
              'text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]',
              'hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]',
              'disabled:opacity-40 disabled:cursor-not-allowed transition-colors duration-150',
            )}
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => mutation.mutate()}
            disabled={mutation.isPending}
            className={cn(
              'px-4 py-2 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
              'bg-[var(--color-error)] text-white',
              'hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed transition-opacity',
            )}
          >
            {mutation.isPending ? 'Erasing...' : 'Permanently Erase'}
          </button>
        </div>
      </div>
    </div>
  );
}
