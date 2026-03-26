'use client';

import { useState, useRef, useCallback } from 'react';
import {
  Upload,
  FileText,
  Star,
  Trash2,
  Download,
  Loader2,
  AlertCircle,
  CheckCircle2,
  Clock,
  XCircle,
  RefreshCw,
} from 'lucide-react';
import { useResumes, useUploadResume, useSetDefaultResume, useArchiveResume, useResumeDownloadUrl, useRetryResume } from '@/hooks/use-resumes';
import { useCreditBalance } from '@/hooks/use-payments';
import type { ResumeListItem } from '@/types/api';

const MAX_SLOTS = 3;
const MAX_FILE_SIZE_MB = 5;

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function StatusBadge({ status }: { status: string }) {
  switch (status) {
    case 'ready':
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-[var(--color-success)]/10 px-2 py-0.5 text-xs font-medium text-[var(--color-success)]">
          <CheckCircle2 className="h-3 w-3" /> Ready
        </span>
      );
    case 'parsing':
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-[var(--color-warning)]/10 px-2 py-0.5 text-xs font-medium text-[var(--color-warning)]">
          <Clock className="h-3 w-3 animate-spin" /> Processing
        </span>
      );
    case 'failed':
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-red-500/10 px-2 py-0.5 text-xs font-medium text-red-400">
          <XCircle className="h-3 w-3" /> Failed
        </span>
      );
    default:
      return null;
  }
}

function ResumeCard({
  resume,
  onSetDefault,
  onArchive,
  onDownload,
  onRetry,
  settingDefault,
  archiving,
  downloading,
  retrying,
}: {
  resume: ResumeListItem;
  onSetDefault: (id: string) => void;
  onArchive: (id: string) => void;
  onDownload: (id: string) => void;
  onRetry: (id: string) => void;
  settingDefault: boolean;
  archiving: boolean;
  downloading: boolean;
  retrying: boolean;
}) {
  return (
    <div
      className={`rounded-lg border p-5 transition-all ${
        resume.is_default
          ? 'border-[var(--color-primary)]/40 bg-[var(--color-primary)]/5 ring-1 ring-[var(--color-primary)]/20'
          : 'border-edge card-hover'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <FileText className="h-5 w-5 shrink-0 text-[var(--color-primary)]" />
            <h3 className="truncate text-sm font-semibold text-[var(--color-text)]">
              {resume.file_name}
            </h3>
            {resume.is_default && (
              <span className="shrink-0 rounded-full bg-[var(--color-primary)]/15 px-2 py-0.5 text-[10px] font-semibold text-[var(--color-primary)]">
                DEFAULT
              </span>
            )}
          </div>
          <div className="mt-1 flex items-center gap-3 text-xs text-[var(--color-text-muted)]">
            <span>Slot {resume.slot_number}</span>
            <span>{formatFileSize(resume.file_size_bytes)}</span>
            <StatusBadge status={resume.status} />
          </div>
        </div>
      </div>

      {/* Failure reason */}
      {resume.status === 'failed' && resume.failure_reason && (
        <div className="mt-3 flex items-start gap-2 rounded-md border border-red-500/20 bg-red-500/5 px-3 py-2">
          <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-red-400" />
          <p className="text-xs text-red-400">{resume.failure_reason}</p>
        </div>
      )}

      {/* ATS Score + Parsed Summary */}
      {resume.status === 'ready' && (
        <div className="mt-3 flex flex-wrap items-center gap-3">
          {resume.ats_general_score != null && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-[var(--color-text-muted)]">ATS Score:</span>
              <span
                className={`text-sm font-bold ${
                  resume.ats_general_score >= 80
                    ? 'text-[var(--color-success)]'
                    : resume.ats_general_score >= 60
                      ? 'text-[var(--color-warning)]'
                      : 'text-red-400'
                }`}
              >
                {resume.ats_general_score}%
              </span>
            </div>
          )}
          {resume.parsed_data_summary?.top_skills &&
            resume.parsed_data_summary.top_skills.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {resume.parsed_data_summary.top_skills.slice(0, 4).map((skill) => (
                  <span
                    key={skill}
                    className="rounded-full border border-edge bg-card px-2 py-0.5 text-[10px] text-[var(--color-text-muted)]"
                  >
                    {skill}
                  </span>
                ))}
              </div>
            )}
        </div>
      )}

      {/* Actions */}
      <div className="mt-4 flex items-center gap-2">
        {!resume.is_default && resume.status === 'ready' && (
          <button
            onClick={() => onSetDefault(resume.id)}
            disabled={settingDefault}
            className="inline-flex items-center gap-1 rounded-md border border-edge px-2.5 py-1.5 text-xs text-[var(--color-text-muted)] hover:border-[var(--color-primary)]/30 hover:text-[var(--color-primary)] transition-all disabled:opacity-50"
            title="Set as default"
          >
            {settingDefault ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <Star className="h-3 w-3" />
            )}
            Set Default
          </button>
        )}
        {resume.status === 'failed' && (
          <button
            onClick={() => onRetry(resume.id)}
            disabled={retrying}
            className="inline-flex items-center gap-1 rounded-md border border-[var(--color-warning)]/30 px-2.5 py-1.5 text-xs text-[var(--color-warning)] hover:border-[var(--color-warning)]/60 hover:bg-[var(--color-warning)]/5 transition-all disabled:opacity-50"
            title="Retry processing"
          >
            {retrying ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <RefreshCw className="h-3 w-3" />
            )}
            Retry
          </button>
        )}
        <button
          onClick={() => onDownload(resume.id)}
          disabled={downloading}
          className="inline-flex items-center gap-1 rounded-md border border-edge px-2.5 py-1.5 text-xs text-[var(--color-text-muted)] hover:border-[var(--color-primary)]/30 hover:text-[var(--color-primary)] transition-all disabled:opacity-50"
          title="Download"
        >
          {downloading ? (
            <Loader2 className="h-3 w-3 animate-spin" />
          ) : (
            <Download className="h-3 w-3" />
          )}
          Download
        </button>
        <button
          onClick={() => onArchive(resume.id)}
          disabled={archiving}
          className="inline-flex items-center gap-1 rounded-md border border-edge px-2.5 py-1.5 text-xs text-red-400 hover:border-red-500/30 hover:bg-red-500/5 transition-all disabled:opacity-50"
          title="Archive"
        >
          {archiving ? (
            <Loader2 className="h-3 w-3 animate-spin" />
          ) : (
            <Trash2 className="h-3 w-3" />
          )}
          Archive
        </button>
      </div>
    </div>
  );
}

function UploadSlot({
  slotNumber,
  onUpload,
  uploading,
}: {
  slotNumber: number;
  onUpload: (file: File, slot: number) => void;
  uploading: boolean;
}) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);
  const [fileError, setFileError] = useState<string | null>(null);

  const handleFile = useCallback(
    (file: File) => {
      setFileError(null);
      if (file.type !== 'application/pdf') {
        setFileError('Only PDF files are supported');
        return;
      }
      if (file.size > MAX_FILE_SIZE_MB * 1024 * 1024) {
        setFileError(`File must be under ${MAX_FILE_SIZE_MB} MB`);
        return;
      }
      onUpload(file, slotNumber);
    },
    [onUpload, slotNumber],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setDragOver(false);
      const file = e.dataTransfer.files[0];
      if (file) handleFile(file);
    },
    [handleFile],
  );

  return (
    <div
      className={`relative rounded-lg border-2 border-dashed p-6 text-center transition-all ${
        dragOver
          ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/5'
          : 'border-edge hover:border-[var(--color-primary)]/30'
      }`}
      onDragOver={(e) => {
        e.preventDefault();
        setDragOver(true);
      }}
      onDragLeave={() => setDragOver(false)}
      onDrop={handleDrop}
    >
      <input
        ref={fileInputRef}
        type="file"
        accept=".pdf,application/pdf"
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (file) handleFile(file);
          e.target.value = '';
        }}
      />

      {uploading ? (
        <div className="flex flex-col items-center gap-2">
          <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary)]" />
          <p className="text-sm text-[var(--color-text-muted)]">Uploading...</p>
        </div>
      ) : (
        <>
          <Upload className="mx-auto h-8 w-8 text-[var(--color-text-muted)]" />
          <p className="mt-2 text-sm text-[var(--color-text-muted)]">
            Slot {slotNumber} — Drop PDF here or{' '}
            <button
              onClick={() => fileInputRef.current?.click()}
              className="text-[var(--color-primary)] hover:underline"
            >
              browse
            </button>
          </p>
          <p className="mt-1 text-xs text-[var(--color-text-muted)]">PDF, max {MAX_FILE_SIZE_MB} MB</p>
          {fileError && (
            <p className="mt-2 flex items-center justify-center gap-1 text-xs text-red-400">
              <AlertCircle className="h-3 w-3" />
              {fileError}
            </p>
          )}
        </>
      )}
    </div>
  );
}

export default function ResumesPage() {
  const { data: resumes, isLoading } = useResumes();
  const { data: credits } = useCreditBalance();
  const uploadResume = useUploadResume();
  const setDefault = useSetDefaultResume();
  const archiveResume = useArchiveResume();
  const downloadUrl = useResumeDownloadUrl();
  const retryResume = useRetryResume();
  const [uploadingSlot, setUploadingSlot] = useState<number | null>(null);
  const [actionId, setActionId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const occupiedSlots = new Set(resumes?.map((r) => r.slot_number) ?? []);
  const emptySlots = Array.from({ length: MAX_SLOTS }, (_, i) => i + 1).filter(
    (s) => !occupiedSlots.has(s),
  );

  const handleUpload = async (file: File, slotNumber: number) => {
    setError(null);
    setUploadingSlot(slotNumber);
    try {
      await uploadResume.mutateAsync({ file, slotNumber });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Upload failed';
      setError(message);
    } finally {
      setUploadingSlot(null);
    }
  };

  const handleSetDefault = async (id: string) => {
    setActionId(id);
    try {
      await setDefault.mutateAsync(id);
    } catch {
      setError('Failed to set default resume');
    } finally {
      setActionId(null);
    }
  };

  const handleArchive = async (id: string) => {
    if (!confirm('Archive this resume? It will be removed from your active slots.')) return;
    setActionId(id);
    try {
      await archiveResume.mutateAsync(id);
    } catch {
      setError('Failed to archive resume');
    } finally {
      setActionId(null);
    }
  };

  const handleDownload = async (id: string) => {
    setActionId(id);
    try {
      const result = await downloadUrl.mutateAsync(id);
      window.open(result.url, '_blank');
    } catch {
      setError('Failed to get download URL');
    } finally {
      setActionId(null);
    }
  };

  const handleRetry = async (id: string) => {
    setActionId(id);
    try {
      await retryResume.mutateAsync(id);
    } catch {
      setError('Failed to retry resume processing');
    } finally {
      setActionId(null);
    }
  };

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text)]">Resumes</h1>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">
            Upload up to {MAX_SLOTS} resumes. AI will parse and score them
            automatically.
          </p>
        </div>
        {credits && (
          <div className="hidden sm:flex items-center gap-2 rounded-lg border border-edge bg-card px-3 py-2">
            <span className="text-xs text-[var(--color-text-muted)]">Upload credits:</span>
            <span className="text-sm font-bold text-[var(--color-primary)]">
              {credits.resume_upload}
            </span>
          </div>
        )}
      </div>

      {error && (
        <div className="mt-4 flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
          <AlertCircle className="h-4 w-4 shrink-0" />
          {error}
          <button
            onClick={() => setError(null)}
            className="ml-auto text-xs text-red-400 hover:text-red-300"
          >
            Dismiss
          </button>
        </div>
      )}

      {credits && credits.resume_upload === 0 && emptySlots.length > 0 && (
        <div className="mt-4 rounded-lg border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-4 py-3 text-sm text-[var(--color-warning)]">
          No upload credits remaining.{' '}
          <a href="/pricing" className="underline hover:text-[var(--color-warning)]/80">
            Buy more credits
          </a>{' '}
          to upload resumes.
        </div>
      )}

      {/* Active Resumes */}
      {isLoading ? (
        <div className="mt-8 flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary)]" />
        </div>
      ) : (
        <div className="mt-8 space-y-4">
          {resumes && resumes.length > 0 && (
            <>
              <h2 className="text-sm font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
                Active Resumes
              </h2>
              {resumes.map((resume) => (
                <ResumeCard
                  key={resume.id}
                  resume={resume}
                  onSetDefault={handleSetDefault}
                  onArchive={handleArchive}
                  onDownload={handleDownload}
                  onRetry={handleRetry}
                  settingDefault={actionId === resume.id && setDefault.isPending}
                  archiving={actionId === resume.id && archiveResume.isPending}
                  downloading={actionId === resume.id && downloadUrl.isPending}
                  retrying={actionId === resume.id && retryResume.isPending}
                />
              ))}
            </>
          )}

          {/* Empty upload slots */}
          {emptySlots.length > 0 && (
            <>
              <h2 className="mt-6 text-sm font-semibold text-[var(--color-text-muted)] uppercase tracking-wider">
                Available Slots
              </h2>
              <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {emptySlots.map((slot) => (
                  <UploadSlot
                    key={slot}
                    slotNumber={slot}
                    onUpload={handleUpload}
                    uploading={uploadingSlot === slot}
                  />
                ))}
              </div>
            </>
          )}

          {resumes?.length === 0 && emptySlots.length === 0 && (
            <div className="rounded-lg border border-edge p-8 text-center">
              <FileText className="mx-auto h-12 w-12 text-[var(--color-text-muted)]" />
              <p className="mt-3 text-sm text-[var(--color-text-muted)]">
                No resumes yet. Upload your first resume to get started.
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
