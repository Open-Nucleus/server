import { cn } from "@/lib/utils";

interface PaginationControlsProps {
  page: number;
  totalPages: number;
  onPageChange: (page: number) => void;
}

export function PaginationControls({
  page,
  totalPages,
  onPageChange,
}: PaginationControlsProps) {
  const isFirst = page <= 1;
  const isLast = page >= totalPages;

  return (
    <div className="flex items-center justify-center gap-4 py-3">
      <button
        onClick={() => onPageChange(page - 1)}
        disabled={isFirst}
        className={cn(
          "px-3 py-1.5 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer",
          "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
          "transition-colors duration-150",
          isFirst
            ? "opacity-40 cursor-not-allowed text-[var(--color-muted)]"
            : "text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]",
        )}
      >
        Previous
      </button>

      <span className="font-mono text-xs text-[var(--color-muted)] tabular-nums">
        Page {page} of {totalPages}
      </span>

      <button
        onClick={() => onPageChange(page + 1)}
        disabled={isLast}
        className={cn(
          "px-3 py-1.5 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer",
          "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
          "transition-colors duration-150",
          isLast
            ? "opacity-40 cursor-not-allowed text-[var(--color-muted)]"
            : "text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)] hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]",
        )}
      >
        Next
      </button>
    </div>
  );
}
