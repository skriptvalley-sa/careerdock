'use client';

import { useState, useRef, useEffect } from 'react';
import {
  Sparkles,
  Loader2,
  AlertCircle,
  Clock,
  ChevronDown,
  ChevronUp,
  Building2,
  Pencil,
  Trash2,
  Check,
  X,
  Plus,
} from 'lucide-react';
import { useResumes } from '@/hooks/use-resumes';
import { useCreditBalance } from '@/hooks/use-payments';
import {
  useCuratedLists,
  useCuratedList,
  useGenerateCuratedList,
  useRenameCuratedList,
  useDeleteCuratedList,
  isCuratedListComplete,
} from '@/hooks/use-curated-lists';
import { useLists, useCreateEntry } from '@/hooks/use-lists';
import type { CuratedList, RankedCompany } from '@/types/api';

function scoreColor(score: number) {
  if (score >= 80) return 'text-[#39ff14]';
  if (score >= 60) return 'text-[#ffb800]';
  return 'text-red-400';
}

function AddToListButton({ companyId, companyName }: { companyId: string; companyName: string }) {
  const { data: lists } = useLists();
  const createEntry = useCreateEntry();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    if (open) document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [open]);

  const handleAdd = async (listId: string) => {
    try {
      await createEntry.mutateAsync({ listId, company_id: companyId });
      setOpen(false);
    } catch {
      // entry may already exist — ignore
      setOpen(false);
    }
  };

  if (!lists || lists.length === 0) return null;

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        title={`Add ${companyName} to a list`}
        className="inline-flex items-center gap-1 rounded-md border border-edge px-2 py-1 text-xs font-medium text-[var(--color-text-muted)] hover:border-[var(--color-primary)]/30 hover:bg-surface hover:text-[var(--color-primary)] transition-colors"
      >
        <Plus className="h-3 w-3" />
      </button>
      {open && (
        <div className="absolute right-0 top-full z-20 mt-1 w-48 rounded-lg border border-edge bg-overlay shadow-lg">
          <p className="px-3 py-2 text-[10px] font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
            Add to list
          </p>
          {lists.map((list) => (
            <button
              key={list.id}
              onClick={() => handleAdd(list.id)}
              disabled={createEntry.isPending}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-[var(--color-text)] hover:bg-surface hover:text-[var(--color-primary)] transition-colors"
            >
              {list.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function RankedCompanyCard({ company, rank }: { company: RankedCompany; rank: number }) {
  return (
    <div className="flex items-start gap-4 rounded-lg border border-edge bg-card p-4">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-edge text-xs font-bold text-[var(--color-text-muted)]">
        {rank}
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-3">
          <h4 className="font-semibold text-[var(--color-text)]">{company.name}</h4>
          <div className="flex items-center gap-2">
            <span className={`shrink-0 text-sm font-bold ${scoreColor(company.match_score)}`}>
              {company.match_score}%
            </span>
            <AddToListButton companyId={company.company_id} companyName={company.name} />
          </div>
        </div>
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">{company.recommendation}</p>
        {company.match_reasons.length > 0 && (
          <ul className="mt-2 space-y-0.5">
            {company.match_reasons.map((reason, i) => (
              <li key={i} className="flex items-start gap-1.5 text-xs text-[var(--color-text-muted)]">
                <span className="mt-1 h-1 w-1 shrink-0 rounded-full bg-[var(--color-primary)]/50" />
                {reason}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

function InlineNameEditor({ list }: { list: CuratedList }) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(list.name || 'Curated List');
  const rename = useRenameCuratedList();
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing) inputRef.current?.focus();
  }, [editing]);

  const handleSave = async () => {
    const trimmed = name.trim();
    if (trimmed && trimmed !== list.name) {
      await rename.mutateAsync({ id: list.id, name: trimmed });
    }
    setEditing(false);
  };

  if (editing) {
    return (
      <div className="flex items-center gap-1.5">
        <input
          ref={inputRef}
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSave();
            if (e.key === 'Escape') {
              setName(list.name || 'Curated List');
              setEditing(false);
            }
          }}
          className="rounded-md border border-edge-input bg-input px-2 py-0.5 text-sm font-semibold text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none"
        />
        <button onClick={handleSave} className="text-[#39ff14] hover:text-[#39ff14]/80">
          <Check className="h-3.5 w-3.5" />
        </button>
        <button
          onClick={() => {
            setName(list.name || 'Curated List');
            setEditing(false);
          }}
          className="text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-1.5">
      <p className="text-sm font-semibold text-[var(--color-text)]">{list.name || 'Curated List'}</p>
      <button
        onClick={() => setEditing(true)}
        className="text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
        title="Rename"
      >
        <Pencil className="h-3 w-3" />
      </button>
    </div>
  );
}

function CuratedListCard({ list }: { list: CuratedList }) {
  const [expanded, setExpanded] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const deleteMutation = useDeleteCuratedList();
  const complete = isCuratedListComplete(list.result);

  const companies = complete ? list.result.companies : [];
  const visible = expanded ? companies : companies.slice(0, 5);
  const hasMore = companies.length > 5;

  const handleDelete = async () => {
    await deleteMutation.mutateAsync(list.id);
    setConfirmDelete(false);
  };

  return (
    <div className="rounded-lg border border-edge bg-card">
      {/* Header */}
      <div className="flex items-center justify-between p-4">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-full border border-edge">
            <Sparkles className="h-4 w-4 text-[var(--color-primary)]" />
          </div>
          <div>
            <InlineNameEditor list={list} />
            <p className="text-xs text-[var(--color-text-muted)]">
              {new Date(list.created_at).toLocaleDateString('en-IN', {
                day: 'numeric',
                month: 'short',
                year: 'numeric',
              })}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          {complete ? (
            <span className="text-sm font-semibold text-[var(--color-primary)]">
              {companies.length} companies
            </span>
          ) : (
            <span className="inline-flex items-center gap-1.5 rounded-full bg-[#ffb800]/10 px-2.5 py-1 text-xs font-medium text-[#ffb800]">
              <Clock className="h-3 w-3 animate-spin" /> Generating…
            </span>
          )}
          {/* Delete button */}
          {confirmDelete ? (
            <div className="flex items-center gap-1.5">
              <button
                onClick={handleDelete}
                disabled={deleteMutation.isPending}
                className="rounded-md bg-red-500/20 px-2 py-1 text-xs font-medium text-red-400 hover:bg-red-500/30"
              >
                {deleteMutation.isPending ? 'Deleting…' : 'Confirm'}
              </button>
              <button
                onClick={() => setConfirmDelete(false)}
                className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
              >
                Cancel
              </button>
            </div>
          ) : (
            <button
              onClick={() => setConfirmDelete(true)}
              className="text-[var(--color-text-muted)] hover:text-red-400 transition-colors"
              title="Delete list"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* Pending state */}
      {!complete && (
        <div className="border-t border-edge px-4 py-6 text-center">
          <Loader2 className="mx-auto h-6 w-6 animate-spin text-[#ffb800]" />
          <p className="mt-2 text-xs text-[var(--color-text-muted)]">
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
              className="flex w-full items-center justify-center gap-2 rounded-md border border-edge py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-all"
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
          <h1 className="text-2xl font-bold text-[var(--color-text)]">Curated Lists</h1>
          <p className="mt-1 text-sm text-[var(--color-text-muted)]">
            AI-ranked company recommendations based on your resume.
          </p>
        </div>
        <div className="flex items-center gap-3">
          {credits && (
            <div className="hidden sm:flex items-center gap-2 rounded-lg border border-edge bg-card px-3 py-2">
              <Sparkles className="h-4 w-4 text-[var(--color-text-muted)]" />
              <span className="text-xs text-[var(--color-text-muted)]">Credits:</span>
              <span className="text-sm font-bold text-[var(--color-primary)]">{credits.curated_list}</span>
            </div>
          )}
          <button
            onClick={() => setShowForm((v) => !v)}
            disabled={(credits?.curated_list ?? 0) === 0}
            className="btn-primary rounded-md px-4 py-2 text-sm font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
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
          className="mt-4 rounded-lg border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/5 p-5 space-y-4"
        >
          <h2 className="text-sm font-semibold text-[var(--color-text)]">
            Select resume to curate against
          </h2>

          {error && (
            <div className="flex items-center gap-2 rounded-md border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
              <AlertCircle className="h-4 w-4 shrink-0" />
              {error}
            </div>
          )}

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

          <div className="flex gap-3">
            <button
              type="submit"
              disabled={generate.isPending || readyResumes.length === 0}
              className="btn-primary flex-1 rounded-md py-2 text-sm font-semibold disabled:opacity-50"
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
              className="rounded-md border border-edge px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-all"
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
            <Loader2 className="h-8 w-8 animate-spin text-[var(--color-primary)]" />
          </div>
        ) : !lists || lists.length === 0 ? (
          <div className="rounded-lg border border-dashed border-edge p-12 text-center">
            <Building2 className="mx-auto h-12 w-12 text-[var(--color-text-muted)]" />
            <p className="mt-3 text-sm text-[var(--color-text-muted)]">No curated lists yet.</p>
            <p className="mt-1 text-xs text-[var(--color-text-muted)]">
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
