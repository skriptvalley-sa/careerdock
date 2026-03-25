'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  ScanSearch,
  Building2,
  FileText,
  FileScan,
  Loader2,
  AlertCircle,
  CheckCircle2,
  Clock,
  ChevronRight,
} from 'lucide-react';
import { useResumes } from '@/hooks/use-resumes';
import { useCreditBalance } from '@/hooks/use-payments';
import { useATSChecks, useCheckCompany, useCheckJob, useCheckResume, isATSComplete } from '@/hooks/use-ats';
import { CompanyCombobox } from '@/components/companies/company-combobox';
import type { ATSCheck } from '@/types/api';

type CheckMode = 'company' | 'job' | 'resume';

function scoreColor(score: number) {
  if (score >= 80) return 'text-[#39ff14]';
  if (score >= 60) return 'text-[#ffb800]';
  return 'text-red-400';
}

function scoreBg(score: number) {
  if (score >= 80) return 'bg-[#39ff14]/10 border-[#39ff14]/30';
  if (score >= 60) return 'bg-[#ffb800]/10 border-[#ffb800]/30';
  return 'bg-red-500/10 border-red-500/30';
}

function checkLabel(check: ATSCheck) {
  if (check.check_type === 'company') {
    return check.company_name ? `Company Check — ${check.company_name}` : 'Company Check';
  }
  if (check.check_type === 'job') {
    return 'Job Description Check';
  }
  return 'Resume Check';
}

function checkIcon(type: ATSCheck['check_type']) {
  if (type === 'company') return <Building2 className="h-4 w-4 text-[#00f0ff]" />;
  if (type === 'job') return <FileText className="h-4 w-4 text-[#e040fb]" />;
  return <FileScan className="h-4 w-4 text-[#39ff14]" />;
}

function ATSCheckRow({ check }: { check: ATSCheck }) {
  const router = useRouter();
  const complete = isATSComplete(check.result);

  return (
    <button
      onClick={() => router.push(`/ats/${check.id}`)}
      className="flex w-full items-center gap-4 rounded-lg border border-edge bg-card p-4 text-left transition-all card-neon-hover"
    >
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-edge">
        {checkIcon(check.check_type)}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-slate-200 truncate">
            {checkLabel(check)}
          </span>
          {complete ? (
            <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-[#39ff14]/10 px-2 py-0.5 text-[10px] font-medium text-[#39ff14]">
              <CheckCircle2 className="h-3 w-3" /> Done
            </span>
          ) : (
            <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-[#ffb800]/10 px-2 py-0.5 text-[10px] font-medium text-[#ffb800]">
              <Clock className="h-3 w-3 animate-spin" /> Processing
            </span>
          )}
        </div>
        <p className="mt-0.5 text-xs text-slate-500">
          {new Date(check.created_at).toLocaleDateString('en-IN', {
            day: 'numeric',
            month: 'short',
            year: 'numeric',
          })}
        </p>
      </div>

      {complete && (
        <div className={`rounded-md border px-2.5 py-1 text-sm font-bold ${scoreBg(check.result.score)} ${scoreColor(check.result.score)}`}>
          {check.result.score}%
        </div>
      )}

      <ChevronRight className="h-4 w-4 shrink-0 text-slate-600" />
    </button>
  );
}

export default function ATSPage() {
  const router = useRouter();
  const { data: resumes } = useResumes();
  const { data: credits } = useCreditBalance();
  const { data: checks, isLoading: checksLoading } = useATSChecks();
  const checkCompany = useCheckCompany();
  const checkJob = useCheckJob();
  const checkResume = useCheckResume();

  const [mode, setMode] = useState<CheckMode>('company');
  const [resumeId, setResumeId] = useState('');
  const [company, setCompany] = useState<{ id: string; name: string } | null>(null);
  const [jobDescription, setJobDescription] = useState('');
  const [error, setError] = useState<string | null>(null);

  const readyResumes = resumes?.filter((r) => r.status === 'ready') ?? [];
  const isPending = checkCompany.isPending || checkJob.isPending || checkResume.isPending;
  const jdLength = jobDescription.length;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!resumeId) {
      setError('Please select a resume');
      return;
    }

    try {
      let result;
      if (mode === 'company') {
        if (!company) {
          setError('Please select a company');
          return;
        }
        result = await checkCompany.mutateAsync({ resumeId, companyId: company.id });
      } else if (mode === 'job') {
        if (jdLength < 100) {
          setError('Job description must be at least 100 characters');
          return;
        }
        if (jdLength > 10000) {
          setError('Job description must be under 10,000 characters');
          return;
        }
        result = await checkJob.mutateAsync({ resumeId, jobDescription });
      } else {
        result = await checkResume.mutateAsync({ resumeId });
      }
      router.push(`/ats/${result.id}`);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Request failed');
    }
  };

  const modeButtonClass = (m: CheckMode, active: boolean) => {
    const base = 'flex flex-1 items-center justify-center gap-2 py-2 text-sm font-medium transition-all';
    if (active) {
      if (m === 'company') return `${base} bg-[#00f0ff]/10 text-[#00f0ff]`;
      if (m === 'job') return `${base} bg-[#e040fb]/10 text-[#e040fb]`;
      return `${base} bg-[#39ff14]/10 text-[#39ff14]`;
    }
    return `${base} text-slate-400 hover:text-slate-200`;
  };

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-100">ATS Check</h1>
          <p className="mt-1 text-sm text-slate-500">
            Score your resume against a company, job description, or on its own.
          </p>
        </div>
        {credits && (
          <div className="hidden sm:flex items-center gap-2 rounded-lg border border-edge bg-card px-3 py-2">
            <ScanSearch className="h-4 w-4 text-slate-500" />
            <span className="text-xs text-slate-500">ATS credits:</span>
            <span className="text-sm font-bold text-[#00f0ff]">{credits.ats_check}</span>
          </div>
        )}
      </div>

      {/* Form */}
      <form
        onSubmit={handleSubmit}
        className="mt-8 rounded-lg border border-edge bg-card p-6 space-y-5"
      >
        {error && (
          <div className="flex items-center gap-2 rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
            <AlertCircle className="h-4 w-4 shrink-0" />
            {error}
          </div>
        )}

        {/* Resume selector */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
            Resume
          </label>
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
        </div>

        {/* Mode toggle — 3 modes */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
            Check type
          </label>
          <div className="flex rounded-md border border-edge overflow-hidden">
            <button
              type="button"
              onClick={() => setMode('company')}
              className={modeButtonClass('company', mode === 'company')}
            >
              <Building2 className="h-4 w-4" /> vs Company
            </button>
            <button
              type="button"
              onClick={() => setMode('job')}
              className={`${modeButtonClass('job', mode === 'job')} border-l border-edge`}
            >
              <FileText className="h-4 w-4" /> vs Job Description
            </button>
            <button
              type="button"
              onClick={() => setMode('resume')}
              className={`${modeButtonClass('resume', mode === 'resume')} border-l border-edge`}
            >
              <FileScan className="h-4 w-4" /> Resume Only
            </button>
          </div>
        </div>

        {/* Company or JD input — hidden for resume-only mode */}
        {mode === 'company' && (
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
              Company
            </label>
            <CompanyCombobox value={company} onChange={setCompany} />
          </div>
        )}

        {mode === 'job' && (
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
              Job Description
            </label>
            <textarea
              value={jobDescription}
              onChange={(e) => setJobDescription(e.target.value)}
              placeholder="Paste the full job description here…"
              rows={8}
              className="block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 placeholder:text-slate-600 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30 resize-y"
            />
            <p
              className={`mt-1 text-right text-xs ${
                jdLength > 10000
                  ? 'text-red-400'
                  : jdLength < 100 && jdLength > 0
                    ? 'text-[#ffb800]'
                    : 'text-slate-600'
              }`}
            >
              {jdLength.toLocaleString()} / 10,000
            </p>
          </div>
        )}

        {mode === 'resume' && (
          <div className="rounded-md border border-edge bg-surface/50 px-4 py-3">
            <p className="text-sm text-slate-400">
              Resume-only mode evaluates your resume&apos;s general ATS compatibility — formatting,
              keyword density, structure, and readability — without targeting a specific company or role.
            </p>
          </div>
        )}

        <button
          type="submit"
          disabled={isPending || readyResumes.length === 0 || (credits?.ats_check ?? 0) === 0}
          className="btn-neon w-full rounded-md py-2.5 text-sm font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isPending ? (
            <span className="flex items-center justify-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" /> Submitting…
            </span>
          ) : (
            'Run ATS Check'
          )}
        </button>

        {(credits?.ats_check ?? 0) === 0 && (
          <p className="text-center text-xs text-[#ffb800]">
            No ATS credits.{' '}
            <a href="/pricing" className="underline hover:text-[#ffb800]/80">
              Buy more
            </a>
            .
          </p>
        )}
      </form>

      {/* History */}
      <div className="mt-10">
        <h2 className="text-sm font-semibold uppercase tracking-wider text-slate-400">
          Recent Checks
        </h2>
        <div className="mt-3 space-y-3">
          {checksLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-[#00f0ff]" />
            </div>
          ) : !checks || checks.length === 0 ? (
            <div className="rounded-lg border border-dashed border-edge p-8 text-center">
              <ScanSearch className="mx-auto h-10 w-10 text-slate-600" />
              <p className="mt-3 text-sm text-slate-500">No checks yet. Run your first ATS check above.</p>
            </div>
          ) : (
            checks.map((check) => <ATSCheckRow key={check.id} check={check} />)
          )}
        </div>
      </div>
    </div>
  );
}
