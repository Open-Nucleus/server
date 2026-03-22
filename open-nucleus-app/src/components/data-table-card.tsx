import { cn } from "@/lib/utils";
import { SearchField } from "./search-field";
import { LoadingSkeleton } from "./loading-skeleton";
import { ErrorState } from "./error-state";
import { EmptyState } from "./empty-state";

interface Column<T> {
  key: string;
  header: string;
  render?: (item: T) => React.ReactNode;
  sortable?: boolean;
  className?: string;
}

interface DataTableCardProps<T> {
  title: string;
  columns: Column<T>[];
  data: T[];
  keyExtractor: (item: T) => string;
  onRowClick?: (item: T) => void;
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  actions?: React.ReactNode;
  emptyIcon?: React.ReactNode;
  emptyTitle?: string;
  emptySubtitle?: string;
  loading?: boolean;
  error?: string;
  onRetry?: () => void;
}

export function DataTableCard<T>({
  title,
  columns,
  data,
  keyExtractor,
  onRowClick,
  searchValue,
  onSearchChange,
  searchPlaceholder,
  actions,
  emptyIcon,
  emptyTitle = "No data",
  emptySubtitle,
  loading = false,
  error,
  onRetry,
}: DataTableCardProps<T>) {
  /* ---------- helper to get cell value ---------- */
  const getCellValue = (item: T, col: Column<T>): React.ReactNode => {
    if (col.render) return col.render(item);
    // Fallback: access item[key] for plain objects
    const record = item as Record<string, unknown>;
    const val = record[col.key];
    if (val === null || val === undefined) return "-";
    return String(val);
  };

  return (
    <div
      className={cn(
        "rounded-[var(--radius-lg)] border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
        "bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]",
      )}
    >
      {/* Header */}
      <div className="flex items-center justify-between gap-4 px-4 py-3 border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
        <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]">
          {title}
        </h2>

        <div className="flex items-center gap-3">
          {onSearchChange && (
            <SearchField
              value={searchValue ?? ""}
              onChange={onSearchChange}
              placeholder={searchPlaceholder}
              className="w-56"
            />
          )}
          {actions}
        </div>
      </div>

      {/* Body */}
      <div className="overflow-x-auto">
        {loading ? (
          <div className="p-6">
            <LoadingSkeleton count={5} />
          </div>
        ) : error ? (
          <ErrorState message={error} onRetry={onRetry} />
        ) : data.length === 0 ? (
          <EmptyState
            icon={emptyIcon}
            title={emptyTitle}
            subtitle={emptySubtitle}
          />
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-[var(--color-border)] dark:border-[var(--color-border-dark)]">
                {columns.map((col) => (
                  <th
                    key={col.key}
                    className={cn(
                      "px-4 py-2.5 text-left font-mono text-[10px] font-semibold uppercase tracking-wider",
                      "text-[var(--color-muted)]",
                      col.className,
                    )}
                  >
                    {col.header}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.map((item) => (
                <tr
                  key={keyExtractor(item)}
                  onClick={onRowClick ? () => onRowClick(item) : undefined}
                  className={cn(
                    "border-b border-[var(--color-border)]/50 dark:border-[var(--color-border-dark)]/50 last:border-b-0",
                    "text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]",
                    onRowClick &&
                      "cursor-pointer hover:bg-[var(--color-surface-hover)] dark:hover:bg-[var(--color-surface-dark-hover)] transition-colors duration-100",
                  )}
                >
                  {columns.map((col) => (
                    <td
                      key={col.key}
                      className={cn("px-4 py-2.5", col.className)}
                    >
                      {getCellValue(item, col)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
