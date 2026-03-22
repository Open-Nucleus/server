import { useEffect, useRef, useState } from "react";
import { Search, X } from "lucide-react";
import { cn } from "@/lib/utils";

interface SearchFieldProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function SearchField({
  value,
  onChange,
  placeholder = "Search...",
  className,
}: SearchFieldProps) {
  const [local, setLocal] = useState(value);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Sync external value changes
  useEffect(() => {
    setLocal(value);
  }, [value]);

  const handleChange = (next: string) => {
    setLocal(next);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      onChange(next);
    }, 300);
  };

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const handleClear = () => {
    setLocal("");
    if (timerRef.current) clearTimeout(timerRef.current);
    onChange("");
  };

  return (
    <div
      className={cn(
        "flex items-center gap-2 px-3 py-1.5 rounded-[var(--radius-md)]",
        "border border-[var(--color-border)] dark:border-[var(--color-border-dark)]",
        "bg-[var(--color-surface)] dark:bg-[var(--color-surface-dark)]",
        "focus-within:border-[var(--color-ink)] dark:focus-within:border-[var(--color-sidebar-text)]",
        "transition-colors duration-150",
        className,
      )}
    >
      <Search size={14} className="shrink-0 text-[var(--color-muted)]" />
      <input
        type="text"
        value={local}
        onChange={(e) => handleChange(e.target.value)}
        placeholder={placeholder}
        className={cn(
          "flex-1 bg-transparent text-sm font-mono outline-none",
          "text-[var(--color-ink)] dark:text-[var(--color-sidebar-text)]",
          "placeholder:text-[var(--color-muted)] placeholder:opacity-60",
        )}
      />
      {local.length > 0 && (
        <button
          onClick={handleClear}
          className="shrink-0 text-[var(--color-muted)] hover:text-[var(--color-ink)] dark:hover:text-[var(--color-sidebar-text)] transition-colors cursor-pointer"
          aria-label="Clear search"
        >
          <X size={14} />
        </button>
      )}
    </div>
  );
}
