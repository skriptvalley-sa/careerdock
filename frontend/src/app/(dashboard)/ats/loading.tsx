import { Skeleton } from '@/components/ui/skeleton';

export default function ATSLoading() {
  return (
    <div>
      <Skeleton className="h-8 w-32" />
      <Skeleton className="mt-2 h-4 w-72" />

      {/* New check form skeleton */}
      <div className="mt-6 rounded-lg border border-edge bg-card p-5">
        <Skeleton className="h-5 w-28" />
        <div className="mt-4 flex gap-3">
          <Skeleton className="h-10 flex-1 rounded-md" />
          <Skeleton className="h-10 w-32 rounded-md" />
        </div>
      </div>

      {/* ATS check history */}
      <div className="mt-8">
        <Skeleton className="h-5 w-24" />
        <div className="mt-4 space-y-3">
          {Array.from({ length: 5 }).map((_, i) => (
            <div
              key={i}
              className="flex items-center gap-4 rounded-lg border border-edge bg-card p-4"
            >
              <Skeleton className="h-8 w-8 rounded-full" />
              <div className="flex-1">
                <Skeleton className="h-4 w-40" />
                <Skeleton className="mt-1.5 h-3 w-24" />
              </div>
              <Skeleton className="h-6 w-12" />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
