import { cn } from "@/lib/utils";

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  subtitle?: string;
  action?: { label: string; onClick: () => void };
}

export function EmptyState({ icon, title, subtitle, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
      {icon && (
        <div className="mb-4 text-[var(--color-muted)] opacity-50">
          {icon}
        </div>
      )}
      <h3 className="font-mono text-sm font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
        {title}
      </h3>
      {subtitle && (
        <p className="mt-1.5 text-sm text-[var(--color-muted)] max-w-xs">
          {subtitle}
        </p>
      )}
      {action && (
        <button
          onClick={action.onClick}
          className={cn(
            "mt-4 px-4 py-2 text-xs font-mono uppercase tracking-wider",
            "border border-[var(--color-ink)] dark:border-[var(--color-sidebar-text)]",
            "text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]",
            "hover:bg-[var(--color-ink)] hover:text-[var(--color-paper)]",
            "dark:hover:bg-[var(--color-sidebar-text)] dark:hover:text-[var(--color-paper-dark)]",
            "transition-colors duration-150 rounded-[var(--radius-sm)] cursor-pointer",
          )}
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
