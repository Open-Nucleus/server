import { useEffect, useRef } from "react";
import { cn } from "@/lib/utils";

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  variant?: "default" | "destructive";
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  onConfirm,
  variant = "default",
}: ConfirmDialogProps) {
  const overlayRef = useRef<HTMLDivElement>(null);

  // Close on Escape key
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onOpenChange(false);
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, onOpenChange]);

  // Close on overlay click
  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === overlayRef.current) {
      onOpenChange(false);
    }
  };

  if (!open) return null;

  const isDestructive = variant === "destructive";

  return (
    <div
      ref={overlayRef}
      onClick={handleOverlayClick}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
    >
      <div
        className={cn(
          "w-full max-w-md mx-4 p-6 rounded-[var(--radius-lg)]",
          "bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]",
          "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
          "shadow-lg",
        )}
        role="dialog"
        aria-modal="true"
        aria-labelledby="confirm-dialog-title"
        aria-describedby="confirm-dialog-description"
      >
        <h2
          id="confirm-dialog-title"
          className="font-mono text-sm font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]"
        >
          {title}
        </h2>
        <p
          id="confirm-dialog-description"
          className="mt-2 text-sm text-[var(--color-muted)] leading-relaxed"
        >
          {description}
        </p>

        <div className="flex justify-end gap-3 mt-6">
          <button
            type="button"
            onClick={() => onOpenChange(false)}
            className={cn(
              "px-4 py-2 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer",
              "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
              "text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)]",
              "hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)]",
              "transition-colors duration-150",
            )}
          >
            {cancelLabel}
          </button>
          <button
            type="button"
            onClick={() => {
              onConfirm();
              onOpenChange(false);
            }}
            className={cn(
              "px-4 py-2 text-xs font-mono uppercase tracking-wider rounded-[var(--radius-sm)] cursor-pointer",
              "transition-colors duration-150",
              isDestructive
                ? "bg-[var(--color-error)] text-white hover:opacity-90"
                : "bg-[var(--color-ink)] text-[var(--color-paper)] dark:bg-[var(--color-sidebar-text)] dark:text-[var(--color-paper-dark)] hover:opacity-90",
            )}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
