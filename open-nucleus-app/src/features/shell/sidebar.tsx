import { useNavigate, useLocation } from "@tanstack/react-router";
import {
  LayoutDashboard,
  Users,
  Pill,
  RefreshCw,
  Bell,
  Shield,
  Settings,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { useUIStore } from "@/stores/ui-store";
import { cn } from "@/lib/utils";

/* ---------- nav config ---------- */

interface NavItem {
  label: string;
  icon: React.ReactNode;
  path: string;
}

interface NavSection {
  title: string;
  items: NavItem[];
}

const ICON_SIZE = 20;

const sections: NavSection[] = [
  {
    title: "Operations",
    items: [
      { label: "Dashboard", icon: <LayoutDashboard size={ICON_SIZE} />, path: "/dashboard" },
      { label: "Patients", icon: <Users size={ICON_SIZE} />, path: "/patients" },
      { label: "Formulary", icon: <Pill size={ICON_SIZE} />, path: "/formulary" },
    ],
  },
  {
    title: "System",
    items: [
      { label: "Sync", icon: <RefreshCw size={ICON_SIZE} />, path: "/sync" },
      { label: "Alerts", icon: <Bell size={ICON_SIZE} />, path: "/alerts" },
      { label: "Integrity", icon: <Shield size={ICON_SIZE} />, path: "/integrity" },
    ],
  },
  {
    title: "Config",
    items: [
      { label: "Settings", icon: <Settings size={ICON_SIZE} />, path: "/settings" },
    ],
  },
];

/* ---------- component ---------- */

export function Sidebar() {
  const expanded = useUIStore((s) => s.sidebarExpanded);
  const toggle = useUIStore((s) => s.toggleSidebar);
  const alertCount = useUIStore((s) => s.unacknowledgedAlerts);
  const navigate = useNavigate();
  const location = useLocation();

  const isActive = (path: string) => location.pathname.startsWith(path);

  return (
    <aside
      className={cn(
        "flex flex-col h-full bg-[var(--color-sidebar)] text-[var(--color-sidebar-text)] transition-all duration-200 ease-in-out select-none",
        expanded ? "w-[var(--sidebar-width)]" : "w-[var(--sidebar-collapsed)]",
      )}
    >
      {/* Brand */}
      <div className="flex items-center h-[var(--topbar-height)] px-4 border-b border-white/10">
        {expanded ? (
          <span className="font-mono text-lg font-bold tracking-[0.2em] text-[var(--color-sidebar-active)] uppercase">
            Nucleus
          </span>
        ) : (
          <span className="font-mono text-lg font-bold text-[var(--color-sidebar-active)] mx-auto">
            N
          </span>
        )}
      </div>

      {/* Nav sections */}
      <nav className="flex-1 overflow-y-auto py-2">
        {sections.map((section) => (
          <div key={section.title} className="mb-2">
            {expanded && (
              <p className="px-4 py-2 text-[10px] font-mono font-semibold uppercase tracking-[0.15em] text-white/40">
                {section.title}
              </p>
            )}
            {section.items.map((item) => {
              const active = isActive(item.path);
              const isAlerts = item.path === "/alerts";

              return (
                <button
                  key={item.path}
                  onClick={() => navigate({ to: item.path })}
                  className={cn(
                    "flex items-center w-full gap-3 px-4 py-2.5 text-sm transition-colors duration-150 cursor-pointer relative",
                    expanded ? "justify-start" : "justify-center",
                    active
                      ? "bg-white/10 text-[var(--color-sidebar-active)] font-semibold"
                      : "text-[var(--color-sidebar-text)] hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-sidebar-active)]",
                  )}
                  title={expanded ? undefined : item.label}
                >
                  <span className="shrink-0">{item.icon}</span>
                  {expanded && (
                    <span className="font-mono text-xs uppercase tracking-wider truncate">
                      {item.label}
                    </span>
                  )}
                  {/* Alert badge */}
                  {isAlerts && alertCount > 0 && (
                    <span
                      className={cn(
                        "flex items-center justify-center bg-[var(--color-error)] text-white text-[10px] font-bold rounded-full min-w-[18px] h-[18px] px-1",
                        expanded ? "ml-auto" : "absolute top-1 right-1",
                      )}
                    >
                      {alertCount > 99 ? "99+" : alertCount}
                    </span>
                  )}
                  {/* Active indicator bar */}
                  {active && (
                    <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 bg-[var(--color-sidebar-active)] rounded-r" />
                  )}
                </button>
              );
            })}
          </div>
        ))}
      </nav>

      {/* Collapse toggle */}
      <div className="border-t border-white/10">
        <button
          onClick={toggle}
          className="flex items-center justify-center w-full py-3 text-white/50 hover:text-white/80 hover:bg-[var(--color-sidebar-hover)] transition-colors duration-150 cursor-pointer"
          title={expanded ? "Collapse sidebar" : "Expand sidebar"}
        >
          {expanded ? <ChevronLeft size={18} /> : <ChevronRight size={18} />}
        </button>
      </div>
    </aside>
  );
}
