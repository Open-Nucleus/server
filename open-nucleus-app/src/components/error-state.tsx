import { AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";

interface ErrorStateProps {
  message: string;
  details?: string;
  onRetry?: () => void;
}

export function ErrorState({ message, details, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
      <div className="mb-4 text-[var(--color-error)]">
        <AlertCircle size={36} strokeWidth={1.5} />
      </div>
      <h3 className="font-mono text-sm font-bold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
        {message}
      </h3>
      {details && (
        <p className="mt-1.5 text-sm text-[var(--color-muted)] max-w-md">
          {details}
        </p>
      )}
      {onRetry && (
        <button
          onClick={onRetry}
          className={cn(
            "mt-4 px-4 py-2 text-xs font-mono uppercase tracking-wider cursor-pointer",
            "border border-[var(--color-error)] text-[var(--color-error)]",
            "hover:bg-[var(--color-error)] hover:text-white",
            "transition-colors duration-150 rounded-[var(--radius-sm)]",
          )}
        >
          Retry
        </button>
      )}
    </div>
  );
}
