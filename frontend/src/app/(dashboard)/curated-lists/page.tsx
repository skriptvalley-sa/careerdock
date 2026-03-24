'use client';

import { useState } from 'react';
import {
  Sparkles,
  Loader2,
  AlertCircle,
  Clock,
  ChevronDown,
  ChevronUp,
  Building2,
} from 'lucide-react';
import { useResumes } from '@/hooks/use-resumes';
import { useCreditBalance } from '@/hooks/use-payments';
import {
  useCuratedLists,
  useCuratedList,
  useGenerateCuratedList,
  isCuratedListComplete,
} from '@/hooks/use-curated-lists';
import type { CuratedList, RankedCompany } from '@/types/api';

function scoreColor(score: number) {
  if (score >= 80) return 'text-[#39ff14]';
  if (score >= 60) return 'text-[#ffb800]';
  return 'text-red-400';
}

function RankedCompanyCard({ company, rank }: { company: RankedCompany; rank: number }) {
  return (
    <div className="flex items-start gap-4 rounded-lg border border-edge bg-card p-4">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-edge text-xs font-bold text-slate-500">
        {rank}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-3">
          <h4 className="font-semibold text-slate-100">{company.name}</h4>
          <span className={`shrink-0 text-sm font-bold ${scoreColor(company.match_score)}`}>
            {company.match_score}%
          </span>
        </div>
        <p className="mt-1 text-sm text-slate-400">{company.recommendation}</p>
        {company.match_reasons.length > 0 && (
          <ul className="mt-2 space-y-0.5">
            {company.match_reasons.map((reason, i) => (
              <li key={i} className="flex items-start gap-1.5 text-xs text-slate-500">
                <span className="mt-1 h-1 w-1 shrink-0 rounded-full bg-[#00f0ff]/50" />
                {reason}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

function CuratedListCard({ list }: { list: CuratedList }) {
  const [expanded, setExpanded] = useState(false);
  const complete = isCuratedListComplete(list.result);

  // When expanded + complete, use full result; otherwise limit to 5
  const companies = complete ? list.result.companies : [];
  const visible = expanded ? companies : companies.slice(0, 5);
  const hasMore = companies.length > 5;

  return (
    <div className="rounded-lg border border-edge bg-card">
      {/* Header */}
      <div className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-full border border-edge">
            <Sparkles className="h-4 w-4 text-[#00f0ff]" />
          </div>
          <div>
            <p className="text-sm font-semibold text-slate-100">
              Curated List
            </p>
            <p className="text-xs text-slate-500">
              {new Date(list.created_at).toLocaleDateString('en-IN', {
                day: 'numeric',
                month: 'short',
                year: 'numeric',
              })}
            </p>
          </div>
        </div>
        {complete ? (
          <span className="text-sm font-semibold text-[#00f0ff]">
            {companies.length} companies
          </span>
        ) : (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-[#ffb800]/10 px-2.5 py-1 text-xs font-medium text-[#ffb800]">
            <Clock className="h-3 w-3 animate-spin" /> Generating…
          </span>
        )}
      </div>

      {/* Pending state */}
      {!complete && (
        <div className="border-t border-edge px-4 py-6 text-center">
          <Loader2 className="mx-auto h-6 w-6 animate-spin text-[#ffb800]" />
          <p className="mt-2 text-xs text-slate-500">
            AI is ranking companies for your profile. This page will update automatically.
          </p>
        </div>
      )}

      {/* Complete — company list */}
      {complete && companies.length > 0 && (
        <div className="border-t border-edge p-4 space-y-3">
          {visible.map((company, i) => (
            <RankedCompanyCard key={company.company_id} company={company} rank={i + 1} />
          ))}

          {hasMore && (
            <button
              onClick={() => setExpanded((v) => !v)}
              className="flex w-full items-center justify-center gap-2 rounded-md border border-edge py-2 text-sm text-slate-400 hover:text-slate-200 transition-all"
            >
              {expanded ? (
                <>
                  <ChevronUp className="h-4 w-4" /> Show less
                </>
              ) : (
                <>
                  <ChevronDown className="h-4 w-4" /> Show all {companies.length} companies
                </>
              )}
            </button>
          )}
        </div>
      )}
    </div>
  );
}

/** Wrapper that uses the polling hook for a single pending list. */
function PendingListPoller({ list }: { list: CuratedList }) {
  // Subscribe to the detail endpoint for polling/SSE-driven updates
  const { data: fresh } = useCuratedList(list.id);
  return <CuratedListCard list={fresh ?? list} />;
}

export default function CuratedListsPage() {
  const { data: resumes } = useResumes();
  const { data: credits } = useCreditBalance();
  const { data: lists, isLoading } = useCuratedLists();
  const generate = useGenerateCuratedList();

  const [resumeId, setResumeId] = useState('');
  const [showForm, setShowForm] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const readyResumes = resumes?.filter((r) => r.status === 'ready') ?? [];

  const handleGenerate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!resumeId) {
      setError('Please select a resume');
      return;
    }
    try {
      await generate.mutateAsync(resumeId);
      setShowForm(false);
      setResumeId('');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to generate list');
    }
  };

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">Curated Lists</h1>
          <p className="mt-1 text-sm text-slate-500">
            AI-ranked company recommendations based on your resume.
          </p>
        </div>
        <div className="flex items-center gap-3">
          {credits && (
            <div className="hidden sm:flex items-center gap-2 rounded-lg border border-edge bg-card px-3 py-2">
              <Sparkles className="h-4 w-4 text-slate-500" />
              <span className="text-xs text-slate-500">Credits:</span>
              <span className="text-sm font-bold text-[#00f0ff]">{credits.curated_list}</span>
            </div>
          )}
          <button
            onClick={() => setShowForm((v) => !v)}
            disabled={(credits?.curated_list ?? 0) === 0}
            className="btn-neon rounded-md px-4 py-2 text-sm font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span className="flex items-center gap-2">
              <Sparkles className="h-4 w-4" />
              Generate New
            </span>
          </button>
        </div>
      </div>

      {/* Generate form (inline) */}
      {showForm && (
        <form
          onSubmit={handleGenerate}
          className="mt-4 rounded-lg border border-[#00f0ff]/30 bg-[#00f0ff]/5 p-5 space-y-4"
        >
          <h2 className="text-sm font-semibold text-slate-200">
            Select resume to curate against
          </h2>

          {error && (
            <div className="flex items-center gap-2 rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
              <AlertCircle className="h-4 w-4 shrink-0" />
              {error}
            </div>
          )}

          {readyResumes.length === 0 ? (
            <p className="text-sm text-slate-500">
              No ready resumes.{' '}
              <a href="/resumes" className="text-[#00f0ff] hover:underline">
                Upload one first
              </a>
              .
            </p>
          ) : (
            <select
              value={resumeId}
              onChange={(e) => setResumeId(e.target.value)}
              className="block w-full rounded-md border border-edge-input bg-input py-2 px-3 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            >
              <option value="">Select a resume…</option>
              {readyResumes.map((r) => (
                <option key={r.id} value={r.id}>
                  {r.file_name}
                  {r.is_default ? ' (default)' : ''}
                </option>
              ))}
            </select>
          )}

          <div className="flex gap-3">
            <button
              type="submit"
              disabled={generate.isPending || readyResumes.length === 0}
              className="btn-neon flex-1 rounded-md py-2 text-sm font-semibold disabled:opacity-50"
            >
              {generate.isPending ? (
                <span className="flex items-center justify-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" /> Generating…
                </span>
              ) : (
                'Generate'
              )}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowForm(false);
                setError(null);
              }}
              className="rounded-md border border-edge px-4 py-2 text-sm text-slate-400 hover:text-slate-200 transition-all"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {(credits?.curated_list ?? 0) === 0 && !showForm && (
        <div className="mt-4 rounded-lg border border-[#ffb800]/30 bg-[#ffb800]/10 px-4 py-3 text-sm text-[#ffb800]">
          No curated list credits.{' '}
          <a href="/pricing" className="underline hover:text-[#ffb800]/80">
            Buy more credits
          </a>
          .
        </div>
      )}

      {/* List of curated results */}
      <div className="mt-8 space-y-4">
        {isLoading ? (
          <div className="flex justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-[#00f0ff]" />
          </div>
        ) : !lists || lists.length === 0 ? (
          <div className="rounded-lg border border-dashed border-edge p-12 text-center">
            <Building2 className="mx-auto h-12 w-12 text-slate-600" />
            <p className="mt-3 text-sm text-slate-400">No curated lists yet.</p>
            <p className="mt-1 text-xs text-slate-600">
              Generate your first list to see AI-ranked company recommendations.
            </p>
          </div>
        ) : (
          lists.map((list) =>
            isCuratedListComplete(list.result) ? (
              <CuratedListCard key={list.id} list={list} />
            ) : (
              <PendingListPoller key={list.id} list={list} />
            ),
          )
        )}
      </div>
    </div>
  );
}
