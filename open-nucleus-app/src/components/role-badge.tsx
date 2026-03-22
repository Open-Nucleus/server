import { cn } from "@/lib/utils";

interface RoleBadgeProps {
  role: string;
}

export function RoleBadge({ role }: RoleBadgeProps) {
  // Capitalize first letter of each word
  const display = role
    .split(/[_\s-]+/)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1).toLowerCase())
    .join(" ");

  return (
    <span
      className={cn(
        "inline-flex items-center px-2.5 py-0.5 rounded-full",
        "text-[10px] font-mono font-semibold uppercase tracking-wider",
        "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
        "text-[var(--color-muted)]",
        "bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]",
      )}
    >
      {display}
    </span>
  );
}
