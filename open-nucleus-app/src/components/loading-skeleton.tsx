import { cn } from "@/lib/utils";

interface LoadingSkeletonProps {
  className?: string;
  count?: number;
}

const WIDTHS = ["w-full", "w-3/4", "w-5/6", "w-2/3", "w-4/5", "w-1/2"];

export function LoadingSkeleton({ className, count = 1 }: LoadingSkeletonProps) {
  return (
    <div className={cn("space-y-3", className)}>
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className={cn(
            "h-4 rounded-[var(--radius-sm)] animate-pulse",
            "bg-[var(--color-border)] dark:bg-[var(--color-border-dark)]",
            count > 1 ? WIDTHS[i % WIDTHS.length] : "w-full",
          )}
        />
      ))}
    </div>
  );
}
