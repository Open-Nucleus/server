import { useUIStore } from "@/stores/ui-store";
import { useAuthStore } from "@/stores/auth-store";
import { useConnection } from "@/hooks/use-connection";
import { StatusIndicator } from "@/components/status-indicator";
import { cn } from "@/lib/utils";

export function TopBar() {
  const pageTitle = useUIStore((s) => s.pageTitle);
  const nodeId = useAuthStore((s) => s.nodeId);
  const siteId = useAuthStore((s) => s.siteId);
  const practitionerId = useAuthStore((s) => s.practitionerId);
  const { connected } = useConnection();

  return (
    <header
      className={cn(
        "flex items-center justify-between h-[var(--topbar-height)] px-4",
        "border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
        "bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]",
      )}
    >
      {/* Left: page title */}
      <h1 className="font-mono text-sm font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
        {pageTitle}
      </h1>

      {/* Right: connection + identifiers */}
      <div className="flex items-center gap-4">
        {/* Connection status */}
        <StatusIndicator
          status={connected ? "connected" : "disconnected"}
          label={connected ? "Connected" : "Disconnected"}
          size="sm"
        />

        {/* Practitioner */}
        {practitionerId && (
          <Chip label="User" value={practitionerId} />
        )}

        {/* Node ID */}
        {nodeId && (
          <Chip label="Node" value={nodeId} />
        )}

        {/* Site ID */}
        {siteId && (
          <Chip label="Site" value={siteId} />
        )}
      </div>
    </header>
  );
}

/* ---------- helper chip ---------- */

function Chip({ label, value }: { label: string; value: string }) {
  // Show truncated ID for long values
  const display = value.length > 12 ? value.slice(0, 8) + "..." : value;

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1 rounded font-mono text-[11px]",
        "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
        "text-[var(--color-muted)] dark:text-[var(--color-sidebar-text)]",
        "bg-[var(--color-paper)] dark:bg-[var(--color-surface-dark)]",
      )}
      title={`${label}: ${value}`}
    >
      <span className="font-semibold uppercase text-[10px] tracking-wider opacity-60">
        {label}
      </span>
      {display}
    </span>
  );
}
