'use client';

import { useState, useRef, useCallback } from 'react';
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
  Upload,
} from 'lucide-react';
import { useResumes } from '@/hooks/use-resumes';
import { useCreditBalance } from '@/hooks/use-payments';
import { useATSChecks, useCheckCompany, useCheckJob, useCheckResume, useCheckResumeTempUpload, isATSComplete } from '@/hooks/use-ats';
import { CompanyCombobox } from '@/components/companies/company-combobox';
import type { ATSCheck } from '@/types/api';

type ResumeSource = 'slot' | 'upload';

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
  if (type === 'company') return <Building2 className="h-4 w-4 text-[var(--color-primary)]" />;
  if (type === 'job') return <FileText className="h-4 w-4 text-[#e040fb]" />;
  return <FileScan className="h-4 w-4 text-[#39ff14]" />;
}

function ATSCheckRow({ check }: { check: ATSCheck }) {
  const router = useRouter();
  const complete = isATSComplete(check.result);

  return (
    <button
      onClick={() => router.push(`/ats/${check.id}`)}
      className="flex w-full items-center gap-4 rounded-lg border border-edge bg-card p-4 text-left transition-all card-hover"
    >
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-edge">
        {checkIcon(check.check_type)}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-[var(--color-text)] truncate">
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
        <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">
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

      <ChevronRight className="h-4 w-4 shrink-0 text-[var(--color-text-muted)]" />
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
  const checkResumeTempUpload = useCheckResumeTempUpload();

  const [mode, setMode] = useState<CheckMode>('company');
  const [resumeSource, setResumeSource] = useState<ResumeSource>('slot');
  const [resumeId, setResumeId] = useState('');
  const [tempFile, setTempFile] = useState<File | null>(null);
  const [tempFileDragOver, setTempFileDragOver] = useState(false);
  const [tempFileError, setTempFileError] = useState<string | null>(null);
  const tempFileRef = useRef<HTMLInputElement>(null);
  const [company, setCompany] = useState<{ id: string; name: string } | null>(null);
  const [jobDescription, setJobDescription] = useState('');
  const [error, setError] = useState<string | null>(null);

  const readyResumes = resumes?.filter((r) => r.status === 'ready') ?? [];
  const isPending = checkCompany.isPending || checkJob.isPending || checkResume.isPending || checkResumeTempUpload.isPending;
  const jdLength = jobDescription.length;

  const handleTempFile = useCallback((file: File) => {
    setTempFileError(null);
    if (file.type !== 'application/pdf') {
      setTempFileError('Only PDF files are supported');
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setTempFileError('File must be under 5 MB');
      return;
    }
    setTempFile(file);
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // For company/job modes we always need a slot resume
    if (mode !== 'resume' && !resumeId) {
      setError('Please select a resume');
      return;
    }
    // For resume-only mode: validate the chosen source
    if (mode === 'resume') {
      if (resumeSource === 'slot' && !resumeId) {
        setError('Please select a resume');
        return;
      }
      if (resumeSource === 'upload' && !tempFile) {
        setError('Please select a PDF file to upload');
        return;
      }
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
        // resume-only mode
        if (resumeSource === 'upload' && tempFile) {
          result = await checkResumeTempUpload.mutateAsync(tempFile);
        } else {
          result = await checkResume.mutateAsync({ resumeId });
        }
      }
      router.push(`/ats/${result.id}`);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Request failed');
    }
  };

  const modeButtonClass = (m: CheckMode, active: boolean) => {
    const base = 'flex flex-1 items-center justify-center gap-2 py-2 text-sm font-medium transition-all';
    if (active) {
      if (m === 'company') return `${base} bg-[var(--color-primary)]/10 text-[var(--color-primary)]`;
      if (m === 'job') return `${base} bg-[#e040fb]/10 text-[#e040fb]`;
      return `${base} bg-[#39ff14]/10 text-[#39ff14]`;
    }
    return `${base} text-[var(--color-text-muted)] hover:text-[var(--color-text)]`;
  };

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text)]">ATS Check</h1>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">
            Score your resume against a company, job description, or on its own.
          </p>
        </div>
        {credits && (
          <div className="hidden sm:flex items-center gap-2 rounded-lg border border-edge bg-card px-3 py-2">
            <ScanSearch className="h-4 w-4 text-[var(--color-text-muted)]" />
            <span className="text-xs text-[var(--color-text-muted)]">ATS credits:</span>
            <span className="text-sm font-bold text-[var(--color-primary)]">{credits.ats_check}</span>
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

        {/* Resume selector — hidden in resume-only + upload source */}
        {(mode !== 'resume' || resumeSource === 'slot') && (
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
              Resume
            </label>
            {readyResumes.length === 0 ? (
              <p className="text-sm text-[var(--color-text-muted)]">
                No ready resumes.{' '}
                <a href="/resumes" className="text-[var(--color-primary)] hover:underline">
                  Upload one first
                </a>
                .
              </p>
            ) : (
              <select
                value={resumeId}
                onChange={(e) => setResumeId(e.target.value)}
                className="block w-full rounded-md border border-edge-input bg-input py-2 px-3 text-sm text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
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
        )}

        {/* Mode toggle — 3 modes */}
        <div>
          <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
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
            <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
              Company
            </label>
            <CompanyCombobox value={company} onChange={setCompany} />
          </div>
        )}

        {mode === 'job' && (
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)] mb-2">
              Job Description
            </label>
            <textarea
              value={jobDescription}
              onChange={(e) => setJobDescription(e.target.value)}
              placeholder="Paste the full job description here…"
              rows={8}
              className="block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30 resize-y"
            />
            <p
              className={`mt-1 text-right text-xs ${
                jdLength > 10000
                  ? 'text-red-400'
                  : jdLength < 100 && jdLength > 0
                    ? 'text-[#ffb800]'
                    : 'text-[var(--color-text-muted)]'
              }`}
            >
              {jdLength.toLocaleString()} / 10,000
            </p>
          </div>
        )}

        {mode === 'resume' && (
          <div className="space-y-3">
            <p className="text-xs text-[var(--color-text-muted)]">
              Evaluates general ATS compatibility — formatting, keyword density, structure — without targeting a specific company or role.
            </p>

            {/* Source toggle: slot vs fresh upload */}
            <div className="flex rounded-md border border-edge overflow-hidden text-xs font-medium">
              <button
                type="button"
                onClick={() => { setResumeSource('slot'); setTempFile(null); setTempFileError(null); }}
                className={`flex flex-1 items-center justify-center gap-1.5 py-2 transition-all ${
                  resumeSource === 'slot'
                    ? 'bg-[#39ff14]/10 text-[#39ff14]'
                    : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'
                }`}
              >
                <FileScan className="h-3.5 w-3.5" /> From my slots
              </button>
              <button
                type="button"
                onClick={() => { setResumeSource('upload'); setResumeId(''); }}
                className={`flex flex-1 items-center justify-center gap-1.5 py-2 border-l border-edge transition-all ${
                  resumeSource === 'upload'
                    ? 'bg-[#39ff14]/10 text-[#39ff14]'
                    : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'
                }`}
              >
                <Upload className="h-3.5 w-3.5" /> Upload a PDF
              </button>
            </div>

            {/* PDF drop zone — shown when upload source is selected */}
            {resumeSource === 'upload' && (
              <div
                className={`relative rounded-lg border-2 border-dashed p-5 text-center transition-all ${
                  tempFileDragOver
                    ? 'border-[#39ff14] bg-[#39ff14]/5'
                    : tempFile
                      ? 'border-[#39ff14]/40 bg-[#39ff14]/5'
                      : 'border-edge hover:border-[#39ff14]/30'
                }`}
                onDragOver={(e) => { e.preventDefault(); setTempFileDragOver(true); }}
                onDragLeave={() => setTempFileDragOver(false)}
                onDrop={(e) => {
                  e.preventDefault();
                  setTempFileDragOver(false);
                  const file = e.dataTransfer.files[0];
                  if (file) handleTempFile(file);
                }}
              >
                <input
                  ref={tempFileRef}
                  type="file"
                  accept=".pdf,application/pdf"
                  className="hidden"
                  onChange={(e) => {
                    const file = e.target.files?.[0];
                    if (file) handleTempFile(file);
                    e.target.value = '';
                  }}
                />
                {tempFile ? (
                  <div className="flex flex-col items-center gap-1">
                    <FileScan className="h-6 w-6 text-[#39ff14]" />
                    <p className="text-sm font-medium text-[var(--color-text)]">{tempFile.name}</p>
                    <p className="text-xs text-[var(--color-text-muted)]">
                      {(tempFile.size / 1024).toFixed(0)} KB —{' '}
                      <button
                        type="button"
                        onClick={() => { setTempFile(null); setTempFileError(null); }}
                        className="text-[var(--color-primary)] hover:underline"
                      >
                        change
                      </button>
                    </p>
                  </div>
                ) : (
                  <>
                    <Upload className="mx-auto h-7 w-7 text-[var(--color-text-muted)]" />
                    <p className="mt-2 text-sm text-[var(--color-text-muted)]">
                      Drop PDF here or{' '}
                      <button
                        type="button"
                        onClick={() => tempFileRef.current?.click()}
                        className="text-[#39ff14] hover:underline"
                      >
                        browse
                      </button>
                    </p>
                    <p className="mt-1 text-xs text-[var(--color-text-muted)]">PDF, max 5 MB — not saved to your slots</p>
                  </>
                )}
                {tempFileError && (
                  <p className="mt-2 flex items-center justify-center gap-1 text-xs text-red-400">
                    <AlertCircle className="h-3 w-3" />
                    {tempFileError}
                  </p>
                )}
              </div>
            )}
          </div>
        )}

        <button
          type="submit"
          disabled={
            isPending ||
            (credits?.ats_check ?? 0) === 0 ||
            // Need at least one ready resume unless in resume+upload mode
            (mode !== 'resume' || resumeSource !== 'upload') && readyResumes.length === 0
          }
          className="btn-primary w-full rounded-md py-2.5 text-sm font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
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
        <h2 className="text-sm font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
          Recent Checks
        </h2>
        <div className="mt-3 space-y-3">
          {checksLoading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-[var(--color-primary)]" />
            </div>
          ) : !checks || checks.length === 0 ? (
            <div className="rounded-lg border border-dashed border-edge p-8 text-center">
              <ScanSearch className="mx-auto h-10 w-10 text-[var(--color-text-muted)]" />
              <p className="mt-3 text-sm text-[var(--color-text-muted)]">No checks yet. Run your first ATS check above.</p>
            </div>
          ) : (
            checks.map((check) => <ATSCheckRow key={check.id} check={check} />)
          )}
        </div>
      </div>
    </div>
  );
}
