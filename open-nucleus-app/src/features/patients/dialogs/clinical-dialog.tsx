import { useEffect, useRef } from 'react';
import { cn } from '@/lib/utils';

export interface ClinicalDialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
  onSubmit: () => void;
  submitting?: boolean;
  error?: string | null;
}

/**
 * Base modal wrapper for all clinical form dialogs.
 * Renders a centered overlay with title, form content, Cancel/Save buttons.
 */
export function ClinicalDialog({
  open,
  onClose,
  title,
  children,
  onSubmit,
  submitting = false,
  error = null,
}: ClinicalDialogProps) {
  const overlayRef = useRef<HTMLDivElement>(null);

  /* ---- close on Escape ---- */
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !submitting) onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, onClose, submitting]);

  /* ---- close on overlay click ---- */
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === overlayRef.current && !submitting) {
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
          'w-full max-w-lg mx-4 rounded-[var(--radius-lg)]',
          'bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]',
          'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
          'shadow-lg',
        )}
        role="dialog"
        aria-modal="true"
        aria-labelledby="clinical-dialog-title"
      >
        {/* Header */}
        <div className="px-6 py-4 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
          <h2
            id="clinical-dialog-title"
            className="font-mono text-sm font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
          >
            {title}
          </h2>
        </div>

        {/* Body */}
        <div className="px-6 py-4 space-y-4 max-h-[60vh] overflow-y-auto">
          {children}
        </div>

        {/* Error */}
        {error && (
          <div className="px-6 pb-2">
            <p className="text-sm text-[var(--color-error)]">{error}</p>
          </div>
        )}

        {/* Footer */}
        <div className="flex justify-end gap-3 px-6 py-4 border-t border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
          <button
            onClick={onClose}
            disabled={submitting}
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
            onClick={onSubmit}
            disabled={submitting}
            className={cn(
              'px-4 py-2 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer',
              'bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)]',
              'hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed transition-opacity',
            )}
          >
            {submitting ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ---------- shared form field helpers ---------- */

export function FieldLabel({ children }: { children: React.ReactNode }) {
  return (
    <label className="block font-mono text-[10px] font-semibold uppercase tracking-wider text-[var(--color-muted)] mb-1.5">
      {children}
    </label>
  );
}

export const inputClass = cn(
  'w-full px-3 py-2 text-sm bg-transparent rounded-[var(--radius-sm)]',
  'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
  'placeholder:text-[var(--color-muted)] placeholder:opacity-60',
  'focus:outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
);

export const selectClass = cn(
  'w-full px-3 py-2 text-sm bg-transparent rounded-[var(--radius-sm)] appearance-none',
  'border border-[var(--color-border)] dark:border-[var(--color-border-dark)]',
  'text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]',
  'focus:outline-none focus:border-[var(--color-ink)] dark:focus:border-[var(--color-sidebar-text)]',
);
