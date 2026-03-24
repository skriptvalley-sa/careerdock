'use client';

import Link from 'next/link';
import {
  ScanSearch,
  Sparkles,
  FileText,
  ChevronRight,
  Building2,
  Clock,
  CheckCircle2,
} from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { useDashboardCounts, useLists } from '@/hooks/use-lists';
import { useResumes } from '@/hooks/use-resumes';
import { useCreditBalance } from '@/hooks/use-payments';
import { useATSChecks, isATSComplete } from '@/hooks/use-ats';
import type { ATSCheck } from '@/types/api';

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

function scoreColor(score: number) {
  if (score >= 80) return 'text-[#39ff14]';
  if (score >= 60) return 'text-[#ffb800]';
  return 'text-red-400';
}

function ATSCheckMini({ check }: { check: ATSCheck }) {
  const complete = isATSComplete(check.result);
  return (
    <Link
      href={`/ats/${check.id}`}
      className="flex items-center gap-3 rounded-lg border border-edge bg-card p-3 transition-all card-neon-hover"
    >
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-edge">
        {check.check_type === 'company' ? (
          <Building2 className="h-3.5 w-3.5 text-[#00f0ff]" />
        ) : (
          <FileText className="h-3.5 w-3.5 text-[#e040fb]" />
        )}
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium text-slate-200">
          {check.check_type === 'company' ? 'Company Check' : 'Job Check'}
        </p>
        <p className="text-[10px] text-slate-500">
          {new Date(check.created_at).toLocaleDateString('en-IN', {
            day: 'numeric',
            month: 'short',
          })}
        </p>
      </div>
      {complete ? (
        <span className={`text-sm font-bold ${scoreColor(check.result.score)}`}>
          {check.result.score}%
        </span>
      ) : (
        <Clock className="h-3.5 w-3.5 animate-spin text-[#ffb800]" />
      )}
      <ChevronRight className="h-3.5 w-3.5 shrink-0 text-slate-600" />
    </Link>
  );
}

export default function DashboardPage() {
  const { user } = useAuthStore();
  const isPremium = !!user?.premium_since;

  const { data: counts, isLoading: countsLoading } = useDashboardCounts();
  const { data: lists, isLoading: listsLoading } = useLists();
  const { data: resumes } = useResumes();
  const { data: credits } = useCreditBalance();
  const { data: atsChecks } = useATSChecks();

  const defaultResume = resumes?.find((r) => r.is_default);
  const recentATSChecks = atsChecks?.slice(0, 3) ?? [];

  return (
    <div>
      <div>
        <h1 className="text-2xl font-bold text-slate-100">Dashboard</h1>
        <p className="mt-1 text-sm text-slate-500">
          Welcome back, {user?.name ?? 'User'}
        </p>
      </div>

      {/* Premium section */}
      {isPremium && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold text-slate-100">AI Tools</h2>

          <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {/* Resume health */}
            <div className="rounded-lg border border-edge bg-card p-4">
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-slate-500">
                <FileText className="h-3.5 w-3.5" /> Resume Health
              </div>
              {defaultResume ? (
                <div className="mt-3">
                  <p className="truncate text-sm text-slate-300">{defaultResume.file_name}</p>
                  {defaultResume.ats_general_score != null ? (
                    <div className="mt-2 flex items-center gap-2">
                      <span className="text-xs text-slate-500">General ATS:</span>
                      <span
                        className={`text-lg font-black ${scoreColor(defaultResume.ats_general_score)}`}
                      >
                        {defaultResume.ats_general_score}%
                      </span>
                    </div>
                  ) : (
                    <p className="mt-2 text-xs text-slate-500">
                      {defaultResume.status === 'parsing'
                        ? 'Parsing in progress…'
                        : 'Score not available'}
                    </p>
                  )}
                  <Link
                    href="/resumes"
                    className="mt-3 inline-flex items-center gap-1 text-xs text-[#00f0ff] hover:underline"
                  >
                    Manage resumes <ChevronRight className="h-3 w-3" />
                  </Link>
                </div>
              ) : (
                <div className="mt-3">
                  <p className="text-sm text-slate-500">No default resume set.</p>
                  <Link
                    href="/resumes"
                    className="mt-2 inline-flex items-center gap-1 text-xs text-[#00f0ff] hover:underline"
                  >
                    Upload a resume <ChevronRight className="h-3 w-3" />
                  </Link>
                </div>
              )}
            </div>

            {/* Credit tracker */}
            {credits && (
              <div className="rounded-lg border border-edge bg-card p-4">
                <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-slate-500">
                  <CheckCircle2 className="h-3.5 w-3.5" /> Credits
                </div>
                <div className="mt-3 space-y-2">
                  {[
                    { label: 'Resume Upload', value: credits.resume_upload },
                    { label: 'ATS Check', value: credits.ats_check },
                    { label: 'Curated List', value: credits.curated_list },
                  ].map(({ label, value }) => (
                    <div key={label} className="flex items-center justify-between">
                      <span className="text-xs text-slate-500">{label}</span>
                      <span
                        className={`text-sm font-bold ${value > 0 ? 'text-slate-100' : 'text-slate-600'}`}
                      >
                        {value}
                      </span>
                    </div>
                  ))}
                </div>
                <Link
                  href="/pricing"
                  className="mt-3 inline-flex items-center gap-1 text-xs text-[#00f0ff] hover:underline"
                >
                  Buy more <ChevronRight className="h-3 w-3" />
                </Link>
              </div>
            )}

            {/* Quick actions */}
            <div className="rounded-lg border border-edge bg-card p-4">
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-slate-500">
                Quick Actions
              </div>
              <div className="mt-3 space-y-2">
                <Link
                  href="/ats"
                  className="flex items-center gap-2 rounded-md border border-edge px-3 py-2 text-sm text-slate-300 hover:border-[#00f0ff]/30 hover:text-[#00f0ff] transition-all"
                >
                  <ScanSearch className="h-4 w-4 shrink-0" />
                  Run ATS Check
                </Link>
                <Link
                  href="/curated-lists"
                  className="flex items-center gap-2 rounded-md border border-edge px-3 py-2 text-sm text-slate-300 hover:border-[#00f0ff]/30 hover:text-[#00f0ff] transition-all"
                >
                  <Sparkles className="h-4 w-4 shrink-0" />
                  Generate Curated List
                </Link>
              </div>
            </div>
          </div>

          {/* Recent ATS checks */}
          {recentATSChecks.length > 0 && (
            <div className="mt-6">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-slate-400">Recent ATS Checks</h3>
                <Link href="/ats" className="text-xs text-[#00f0ff] hover:text-[#00f0ff]/80">
                  View all &rarr;
                </Link>
              </div>
              <div className="mt-2 space-y-2">
                {recentATSChecks.map((check) => (
                  <ATSCheckMini key={check.id} check={check} />
                ))}
              </div>
            </div>
          )}
        </div>
      )}

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
