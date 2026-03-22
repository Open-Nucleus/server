import { cn } from "@/lib/utils";

interface StatusIndicatorProps {
  status: "active" | "inactive" | "pending" | "error" | "connected" | "disconnected";
  label?: string;
  size?: "sm" | "md";
}

const DOT_COLORS: Record<StatusIndicatorProps["status"], string> = {
  active: "bg-[var(--color-success)]",
  connected: "bg-[var(--color-success)]",
  error: "bg-[var(--color-error)]",
  disconnected: "bg-[var(--color-error)]",
  pending: "bg-[var(--color-warning)]",
  inactive: "bg-[var(--color-muted)]",
};

export function StatusIndicator({ status, label, size = "md" }: StatusIndicatorProps) {
  const dotSize = size === "sm" ? "w-2 h-2" : "w-2.5 h-2.5";

  return (
    <span className="inline-flex items-center gap-1.5">
      <span
        className={cn(
          "rounded-full shrink-0",
          dotSize,
          DOT_COLORS[status],
        )}
      />
      {label && (
        <span
          className={cn(
            "font-mono uppercase tracking-wider",
            size === "sm" ? "text-[10px]" : "text-xs",
            "text-[var(--color-muted)]",
          )}
        >
          {label}
        </span>
      )}
    </span>
  );
}
