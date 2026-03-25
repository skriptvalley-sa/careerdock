/** Reusable skeleton primitives for loading states. */

interface SkeletonProps {
  className?: string;
}

/** A single animated placeholder block. */
export function Skeleton({ className = '' }: SkeletonProps) {
  return <div className={`animate-pulse rounded bg-slate-800 ${className}`} />;
}

/** Skeleton row for table-like layouts. */
export function SkeletonRow({ cols = 4 }: { cols?: number }) {
  return (
    <div className="flex items-center gap-4 px-4 py-3">
      {Array.from({ length: cols }).map((_, i) => (
        <Skeleton
          key={i}
          className={`h-4 ${i === 0 ? 'w-32' : i === cols - 1 ? 'w-16' : 'w-24'}`}
        />
      ))}
    </div>
  );
}

/** Card skeleton placeholder. */
export function SkeletonCard() {
  return (
    <div className="rounded-lg border border-edge bg-card p-4">
      <Skeleton className="mb-3 h-5 w-3/4" />
      <Skeleton className="mb-2 h-3 w-1/2" />
      <Skeleton className="h-3 w-2/3" />
    </div>
  );
}

/** Table skeleton with header and rows. */
export function SkeletonTable({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div className="overflow-hidden rounded-lg border border-edge">
      {/* Header */}
      <div className="border-b border-edge bg-overlay px-4 py-3">
        <div className="flex items-center gap-4">
          {Array.from({ length: cols }).map((_, i) => (
            <Skeleton key={i} className="h-3 w-20" />
          ))}
        </div>
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="border-b border-edge last:border-b-0">
          <SkeletonRow cols={cols} />
        </div>
      ))}
    </div>
  );
}
