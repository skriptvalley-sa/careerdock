'use client';

import { use } from 'react';
import Link from 'next/link';
import {
  ArrowLeft,
  Building2,
  FileText,
  FileCheck2,
  Loader2,
  CheckCircle2,
  AlertCircle,
} from 'lucide-react';
import { useATSCheck, isATSComplete } from '@/hooks/use-ats';

function scoreColor(score: number) {
  if (score >= 80) return 'text-[#39ff14]';
  if (score >= 60) return 'text-[#ffb800]';
  return 'text-red-400';
}

function scoreBorderColor(score: number) {
  if (score >= 80) return 'border-[#39ff14]/40';
  if (score >= 60) return 'border-[#ffb800]/40';
  return 'border-red-400/40';
}

function formatCategoryName(key: string) {
  return key
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

function CategoryBar({ score }: { score: number }) {
  return (
    <div className="mt-1.5 h-1.5 w-full rounded-full bg-surface">
      <div
        className={`h-1.5 rounded-full transition-all ${
          score >= 80 ? 'bg-[#39ff14]' : score >= 60 ? 'bg-[#ffb800]' : 'bg-red-400'
        }`}
        style={{ width: `${score}%` }}
      />
    </div>
  );
}

export default function ATSResultPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);
  const { data: check, isLoading } = useATSCheck(id);

  if (isLoading || !check) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-[#00f0ff]" />
      </div>
    );
  }

  const complete = isATSComplete(check.result);

  return (
    <div>
      {/* Back */}
      <Link
        href="/ats"
        className="inline-flex items-center gap-1.5 text-sm text-slate-500 hover:text-slate-300"
      >
        <ArrowLeft className="h-4 w-4" /> Back to ATS Checks
      </Link>

      {/* Header */}
      <div className="mt-4 flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full border border-edge">
          {check.check_type === 'company' ? (
            <Building2 className="h-5 w-5 text-[#00f0ff]" />
          ) : check.check_type === 'job' ? (
            <FileText className="h-5 w-5 text-[#e040fb]" />
          ) : (
            <FileCheck2 className="h-5 w-5 text-[#39ff14]" />
          )}
        </div>
        <div>
          <h1 className="text-xl font-bold text-slate-100">
            {check.check_type === 'company'
              ? 'Company ATS Check'
              : check.check_type === 'job'
                ? 'Job ATS Check'
                : 'General ATS Check'}
          </h1>
          <p className="text-xs text-slate-500">
            {new Date(check.created_at).toLocaleDateString('en-IN', {
              day: 'numeric',
              month: 'long',
              year: 'numeric',
            })}
          </p>
        </div>
      </div>

      {/* Pending state */}
      {!complete && (
        <div className="mt-8 flex flex-col items-center justify-center rounded-lg border border-dashed border-[#ffb800]/40 bg-[#ffb800]/5 py-16">
          <Loader2 className="h-10 w-10 animate-spin text-[#ffb800]" />
          <p className="mt-4 text-sm font-medium text-[#ffb800]">Analysis in progress…</p>
          <p className="mt-1 text-xs text-slate-500">
            This usually takes 15–30 seconds. This page will update automatically.
          </p>
        </div>
      )}

      {/* Complete state */}
      {complete && (
        <div className="mt-6 space-y-6">
          {/* Score hero */}
          <div className="flex items-center gap-6 rounded-lg border border-edge bg-card p-6">
            <div
              className={`flex h-24 w-24 shrink-0 items-center justify-center rounded-full border-4 ${scoreBorderColor(check.result.score)}`}
            >
              <span className={`text-3xl font-black ${scoreColor(check.result.score)}`}>
                {check.result.score}
              </span>
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wider text-slate-500">
                Overall Score
              </p>
              <p className={`mt-1 text-2xl font-bold ${scoreColor(check.result.score)}`}>
                {check.result.score >= 80
                  ? 'Strong Match'
                  : check.result.score >= 60
                    ? 'Moderate Match'
                    : 'Weak Match'}
              </p>
              <p className="mt-1 text-sm text-slate-500">
                {check.result.score >= 80
                  ? check.check_type === 'resume'
                    ? 'Your resume is well-optimised for ATS systems.'
                    : 'Your resume is well-optimised for this target.'
                  : check.result.score >= 60
                    ? 'Some gaps exist — review the suggestions below.'
                    : 'Significant gaps found — work through the suggestions.'}
              </p>
            </div>
          </div>

          {/* Breakdown */}
          {check.result.breakdown && Object.keys(check.result.breakdown).length > 0 && (
            <div>
              <h2 className="text-sm font-semibold uppercase tracking-wider text-slate-400">
                Score Breakdown
              </h2>
              <div className="mt-3 grid gap-3 sm:grid-cols-2">
                {Object.entries(check.result.breakdown).map(([key, detail]) => (
                  <div
                    key={key}
                    className="rounded-lg border border-edge bg-card p-4"
                  >
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-slate-200">
                        {formatCategoryName(key)}
                      </span>
                      <span className={`text-sm font-bold ${scoreColor(detail.score)}`}>
                        {detail.score}%
                      </span>
                    </div>
                    <CategoryBar score={detail.score} />
                    {detail.feedback && (
                      <p className="mt-2 text-xs text-slate-500">{detail.feedback}</p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Suggestions */}
          {check.result.suggestions && check.result.suggestions.length > 0 && (
            <div>
              <h2 className="text-sm font-semibold uppercase tracking-wider text-slate-400">
                Recommendations
              </h2>
              <ul className="mt-3 space-y-2">
                {check.result.suggestions.map((suggestion, i) => (
                  <li
                    key={i}
                    className="flex items-start gap-3 rounded-lg border border-edge bg-card p-4"
                  >
                    <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-[#ffb800]" />
                    <span className="text-sm text-slate-300">{suggestion}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-3 pt-2">
            <Link
              href="/ats"
              className="inline-flex items-center gap-2 rounded-md border border-edge px-4 py-2 text-sm text-slate-400 hover:border-[#00f0ff]/30 hover:text-[#00f0ff] transition-all"
            >
              <CheckCircle2 className="h-4 w-4" /> Run Another Check
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
