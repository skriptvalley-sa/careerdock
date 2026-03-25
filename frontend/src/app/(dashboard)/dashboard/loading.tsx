import { Skeleton, SkeletonCard } from '@/components/ui/skeleton';

export default function DashboardLoading() {
  return (
    <div>
      {/* Header */}
      <Skeleton className="h-8 w-40" />
      <Skeleton className="mt-2 h-4 w-56" />

      {/* AI Tools cards */}
      <div className="mt-8">
        <Skeleton className="h-6 w-24" />
        <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      </div>

      {/* Funnel */}
      <div className="mt-8">
        <Skeleton className="h-6 w-40" />
        <Skeleton className="mt-2 h-4 w-48" />
        <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="rounded-lg border border-edge bg-card p-4 text-center">
              <Skeleton className="mx-auto mb-2 h-2 w-12" />
              <Skeleton className="mx-auto h-8 w-12" />
              <Skeleton className="mx-auto mt-2 h-3 w-16" />
            </div>
          ))}
        </div>
      </div>

      {/* Your Lists */}
      <div className="mt-10">
        <Skeleton className="h-6 w-28" />
        <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      </div>
    </div>
  );
}
