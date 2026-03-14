'use client';

import Link from 'next/link';
import { useAuthStore } from '@/store/auth-store';
import { useDashboardCounts, useLists } from '@/hooks/use-lists';

const funnelStages = [
  { key: 'not_applied' as const, label: 'Not Applied', color: 'bg-slate-600' },
  { key: 'applied' as const, label: 'Applied', color: 'bg-[#00f0ff]' },
  { key: 'phone_screen' as const, label: 'Phone Screen', color: 'bg-[#e040fb]' },
  { key: 'interview' as const, label: 'Interview', color: 'bg-[#ffb800]' },
  { key: 'offer' as const, label: 'Offer', color: 'bg-[#39ff14]' },
  { key: 'accepted' as const, label: 'Accepted', color: 'bg-[#39ff14]' },
  { key: 'rejected' as const, label: 'Rejected', color: 'bg-red-500' },
  { key: 'withdrawn' as const, label: 'Withdrawn', color: 'bg-orange-500' },
];

export default function DashboardPage() {
  const { user } = useAuthStore();
  const { data: counts, isLoading: countsLoading } = useDashboardCounts();
  const { data: lists, isLoading: listsLoading } = useLists();

  return (
    <div>
      <div>
        <h1 className="text-2xl font-bold text-slate-100">Dashboard</h1>
        <p className="mt-1 text-sm text-slate-500">
          Welcome back, {user?.name ?? 'User'}
        </p>
      </div>

      {/* Funnel stats */}
      <div className="mt-8">
        <h2 className="text-lg font-semibold text-slate-100">Application Funnel</h2>
        {countsLoading ? (
          <div className="mt-4 flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-4 border-[#00f0ff] border-t-transparent" />
          </div>
        ) : counts ? (
          <>
            <p className="mt-1 text-sm text-slate-500">
              {counts.total} total application{counts.total !== 1 ? 's' : ''} tracked
            </p>
            <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
              {funnelStages.map((stage) => (
                <Link
                  key={stage.key}
                  href={`/applications?status=${stage.key}`}
                  className="card-neon-hover rounded-lg border border-edge bg-card p-4 text-center transition-all"
                >
                  <div className={`mx-auto mb-2 h-2 w-12 rounded-full ${stage.color}`} />
                  <p className="text-2xl font-bold text-slate-100">
                    {counts[stage.key]}
                  </p>
                  <p className="mt-1 text-xs text-slate-500">{stage.label}</p>
                </Link>
              ))}
            </div>
          </>
        ) : (
          <p className="mt-4 text-sm text-slate-500">No data yet.</p>
        )}
      </div>

      {/* Quick access to lists */}
      <div className="mt-10">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-100">Your Lists</h2>
          <Link href="/lists" className="text-sm text-[#00f0ff] hover:text-[#00f0ff]/80">
            View all &rarr;
          </Link>
        </div>
        {listsLoading ? (
          <div className="mt-4 flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-4 border-[#00f0ff] border-t-transparent" />
          </div>
        ) : !lists || lists.length === 0 ? (
          <div className="mt-4 rounded-lg border border-dashed border-edge p-8 text-center">
            <p className="text-sm text-slate-500">
              No lists yet.{' '}
              <Link href="/lists" className="text-[#00f0ff] hover:text-[#00f0ff]/80">
                Create your first list
              </Link>{' '}
              to start tracking applications.
            </p>
          </div>
        ) : (
          <div className="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {lists.map((list) => (
              <Link
                key={list.id}
                href={`/lists/${list.id}`}
                className="card-neon-hover rounded-lg border border-edge bg-card p-4 transition-all"
              >
                <h3 className="font-medium text-slate-100">{list.name}</h3>
                <p className="mt-1 text-sm text-slate-600">
                  {list.entry_count} {list.entry_count === 1 ? 'entry' : 'entries'}
                </p>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
