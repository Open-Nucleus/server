import { cn } from "@/lib/utils";

interface SeverityBadgeProps {
  severity: "critical" | "high" | "warning" | "moderate" | "info" | "low";
}

const SEVERITY_COLORS: Record<SeverityBadgeProps["severity"], string> = {
  critical: "bg-[var(--color-critical)]",
  high: "bg-[var(--color-high)]",
  warning: "bg-[var(--color-moderate)]",
  moderate: "bg-[var(--color-moderate)]",
  info: "bg-[var(--color-info)]",
  low: "bg-[var(--color-low)]",
};

export function SeverityBadge({ severity }: SeverityBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center px-2 py-0.5 rounded-full",
        "text-[10px] font-mono font-semibold uppercase tracking-wider text-white",
        SEVERITY_COLORS[severity],
      )}
    >
      {severity}
    </span>
  );
}
