'use client';

import { useState } from 'react';
import {
  Shield,
  Loader2,
  CheckCircle2,
  AlertCircle,
  Pencil,
  Lock,
  Unlock,
  Plus,
  Sparkles,
} from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { CompanyCombobox } from '@/components/companies/company-combobox';
import {
  useGenerateCompanyDraft,
  useSubmitCompanyDraft,
  useAcquireEditLock,
  useReleaseEditLock,
  useSubmitCompanyEdit,
  useEditLockStatus,
  type CompanyDraft,
} from '@/hooks/use-moderator';

// --- Company Draft Generator ---

function GenerateCompanySection() {
  const [name, setName] = useState('');
  const [careersUrl, setCareersUrl] = useState('');
  const [linkedinUrl, setLinkedinUrl] = useState('');
  const [draft, setDraft] = useState<CompanyDraft | null>(null);
  const [submitted, setSubmitted] = useState(false);

  const generate = useGenerateCompanyDraft();
  const submit = useSubmitCompanyDraft();

  const handleGenerate = () => {
    if (!name.trim()) return;
    generate.mutate(
      { name: name.trim(), careers_url: careersUrl.trim() || undefined, linkedin_url: linkedinUrl.trim() || undefined },
      {
        onSuccess: (data) => {
          setDraft(data);
          setSubmitted(false);
        },
      },
    );
  };

  const handleSubmit = () => {
    if (!draft) return;
    submit.mutate(draft, {
      onSuccess: () => {
        setSubmitted(true);
        setDraft(null);
        setName('');
        setCareersUrl('');
        setLinkedinUrl('');
      },
    });
  };

  const updateDraft = (field: string, value: unknown) => {
    if (!draft) return;
    setDraft({ ...draft, [field]: value });
  };

  return (
    <div className="rounded-lg border border-edge bg-card p-6">
      <div className="mb-4 flex items-center gap-2">
        <Plus className="h-5 w-5 text-[var(--color-primary)]" />
        <h2 className="text-lg font-semibold text-[var(--color-text)]">Add New Company</h2>
      </div>
      <p className="mb-4 text-sm text-[var(--color-text-muted)]">
        Enter a company name and optionally provide URLs. AI will generate a draft profile for review.
      </p>

      {/* Input form */}
      <div className="space-y-3">
        <input
          type="text"
          placeholder="Company name *"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-full rounded-md border border-edge bg-overlay px-3 py-2 text-sm text-[var(--color-text)] placeholder-slate-500 focus:border-[var(--color-primary)] focus:outline-none"
        />
        <div className="grid gap-3 sm:grid-cols-2">
          <input
            type="url"
            placeholder="Careers page URL (optional)"
            value={careersUrl}
            onChange={(e) => setCareersUrl(e.target.value)}
            className="w-full rounded-md border border-edge bg-overlay px-3 py-2 text-sm text-[var(--color-text)] placeholder-slate-500 focus:border-[var(--color-primary)] focus:outline-none"
          />
          <input
            type="url"
            placeholder="LinkedIn URL (optional)"
            value={linkedinUrl}
            onChange={(e) => setLinkedinUrl(e.target.value)}
            className="w-full rounded-md border border-edge bg-overlay px-3 py-2 text-sm text-[var(--color-text)] placeholder-slate-500 focus:border-[var(--color-primary)] focus:outline-none"
          />
        </div>
        <button
          onClick={handleGenerate}
          disabled={!name.trim() || generate.isPending}
          className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
        >
          {generate.isPending ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Sparkles className="h-4 w-4" />
          )}
          Generate Draft
        </button>
        {generate.isError && (
          <p className="text-sm text-red-400">
            <AlertCircle className="mr-1 inline h-4 w-4" />
            {generate.error?.message || 'Failed to generate draft'}
          </p>
        )}
      </div>

      {/* Draft review */}
      {draft && (
        <div className="mt-6 space-y-4 border-t border-edge pt-4">
          <h3 className="text-sm font-medium text-[var(--color-text)]">Review Generated Draft</h3>
          <div className="grid gap-3 sm:grid-cols-2">
            <DraftField label="Name" value={draft.name} onChange={(v) => updateDraft('name', v)} />
            <DraftField label="Slug" value={draft.slug || ''} onChange={(v) => updateDraft('slug', v)} />
            <DraftField label="Headquarters" value={draft.headquarters || ''} onChange={(v) => updateDraft('headquarters', v)} />
            <DraftField label="Size" value={draft.size || ''} onChange={(v) => updateDraft('size', v)} />
            <DraftField label="Hiring Status" value={draft.hiring_status || ''} onChange={(v) => updateDraft('hiring_status', v)} />
            <DraftField label="Compensation Tier" value={draft.compensation_tier || ''} onChange={(v) => updateDraft('compensation_tier', v)} />
            <DraftField label="Founded Year" value={String(draft.founded_year || '')} onChange={(v) => updateDraft('founded_year', v ? parseInt(v) : undefined)} />
            <DraftField label="Careers URL" value={draft.careers_page_url || ''} onChange={(v) => updateDraft('careers_page_url', v)} />
          </div>
          <DraftField
            label="Description"
            value={draft.description || ''}
            onChange={(v) => updateDraft('description', v)}
            multiline
          />
          <DraftTagField label="Tech Stack" values={draft.tech_stack || []} onChange={(v) => updateDraft('tech_stack', v)} />
          <DraftTagField label="Domains" values={draft.domains || []} onChange={(v) => updateDraft('domains', v)} />
          <DraftTagField label="Office Modes" values={draft.office_modes || []} onChange={(v) => updateDraft('office_modes', v)} />

          <div className="flex items-center gap-3 border-t border-edge pt-4">
            <button
              onClick={handleSubmit}
              disabled={submit.isPending}
              className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {submit.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <CheckCircle2 className="h-4 w-4" />}
              Submit to Directory
            </button>
            <button
              onClick={() => setDraft(null)}
              className="rounded-md border border-edge px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
            >
              Discard
            </button>
          </div>
          {submit.isError && (
            <p className="text-sm text-red-400">
              <AlertCircle className="mr-1 inline h-4 w-4" />
              {submit.error?.message || 'Failed to submit'}
            </p>
          )}
        </div>
      )}

      {/* Success */}
      {submitted && (
        <div className="mt-4 flex items-center gap-2 rounded-md bg-[#39ff14]/10 border border-[#39ff14]/30 p-3 text-sm text-[#39ff14]">
          <CheckCircle2 className="h-4 w-4" />
          Company added to directory successfully.
        </div>
      )}
    </div>
  );
}

// --- Company Editor with Lock ---

function EditCompanySection() {
  const [selectedCompany, setSelectedCompany] = useState<{ id: string; name: string } | null>(null);
  const [editFields, setEditFields] = useState<Record<string, string>>({});
  const [editTechStack, setEditTechStack] = useState<string[]>([]);
  const [editDomains, setEditDomains] = useState<string[]>([]);

  const acquireLock = useAcquireEditLock();
  const releaseLock = useReleaseEditLock();
  const submitEdit = useSubmitCompanyEdit();
  const { data: lockStatus } = useEditLockStatus(selectedCompany?.id || '');

  const hasLock = !!lockStatus;

  const handleAcquireLock = () => {
    if (!selectedCompany) return;
    acquireLock.mutate(selectedCompany.id, {
      onSuccess: () => {
        setEditFields({});
        setEditTechStack([]);
        setEditDomains([]);
      },
    });
  };

  const handleReleaseLock = () => {
    if (!selectedCompany) return;
    releaseLock.mutate(selectedCompany.id);
  };

  const handleSubmitEdit = () => {
    if (!selectedCompany) return;
    const changes: Record<string, unknown> = {};
    if (editFields.description) changes.description = editFields.description;
    if (editFields.size) changes.size = editFields.size;
    if (editFields.headquarters) changes.headquarters = editFields.headquarters;
    if (editFields.careers_page_url) changes.careers_page_url = editFields.careers_page_url;
    if (editFields.linkedin_url) changes.linkedin_url = editFields.linkedin_url;
    if (editFields.hiring_status) changes.hiring_status = editFields.hiring_status;
    if (editFields.compensation_tier) changes.compensation_tier = editFields.compensation_tier;
    if (editTechStack.length > 0) changes.tech_stack = editTechStack;
    if (editDomains.length > 0) changes.domains = editDomains;

    if (Object.keys(changes).length === 0) return;

    submitEdit.mutate(
      { companyId: selectedCompany.id, changes },
      {
        onSuccess: () => {
          setSelectedCompany(null);
          setEditFields({});
        },
      },
    );
  };

  return (
    <div className="rounded-lg border border-edge bg-card p-6">
      <div className="mb-4 flex items-center gap-2">
        <Pencil className="h-5 w-5 text-[#e040fb]" />
        <h2 className="text-lg font-semibold text-[var(--color-text)]">Edit Company</h2>
      </div>
      <p className="mb-4 text-sm text-[var(--color-text-muted)]">
        Search for a company, acquire an edit lock, then submit changes. A 10-minute cooldown applies after each edit.
      </p>

      <div className="mb-4">
        <CompanyCombobox
          value={selectedCompany}
          onChange={(c) => {
            setSelectedCompany(c);
            setEditFields({});
            setEditTechStack([]);
            setEditDomains([]);
          }}
          placeholder="Search company to edit..."
        />
      </div>

      {selectedCompany && !hasLock && (
        <div>
          <button
            onClick={handleAcquireLock}
            disabled={acquireLock.isPending}
            className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
          >
            {acquireLock.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Lock className="h-4 w-4" />}
            Acquire Edit Lock
          </button>
          {acquireLock.isError && (
            <p className="mt-2 text-sm text-red-400">
              <AlertCircle className="mr-1 inline h-4 w-4" />
              {acquireLock.error?.message || 'Failed to acquire lock'}
            </p>
          )}
        </div>
      )}

      {selectedCompany && hasLock && (
        <div className="space-y-4 border-t border-edge pt-4">
          <div className="flex items-center gap-2 rounded-md bg-[var(--color-primary)]/10 border border-[var(--color-primary)]/30 p-3 text-sm text-[var(--color-primary)]">
            <Lock className="h-4 w-4" />
            Editing {selectedCompany.name} — lock active
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <DraftField label="Description" value={editFields.description || ''} onChange={(v) => setEditFields((f) => ({ ...f, description: v }))} multiline />
            <DraftField label="Size" value={editFields.size || ''} onChange={(v) => setEditFields((f) => ({ ...f, size: v }))} />
            <DraftField label="Headquarters" value={editFields.headquarters || ''} onChange={(v) => setEditFields((f) => ({ ...f, headquarters: v }))} />
            <DraftField label="Hiring Status" value={editFields.hiring_status || ''} onChange={(v) => setEditFields((f) => ({ ...f, hiring_status: v }))} />
            <DraftField label="Careers URL" value={editFields.careers_page_url || ''} onChange={(v) => setEditFields((f) => ({ ...f, careers_page_url: v }))} />
            <DraftField label="LinkedIn URL" value={editFields.linkedin_url || ''} onChange={(v) => setEditFields((f) => ({ ...f, linkedin_url: v }))} />
            <DraftField label="Compensation Tier" value={editFields.compensation_tier || ''} onChange={(v) => setEditFields((f) => ({ ...f, compensation_tier: v }))} />
          </div>
          <DraftTagField label="Tech Stack" values={editTechStack} onChange={setEditTechStack} />
          <DraftTagField label="Domains" values={editDomains} onChange={setEditDomains} />

          <div className="flex items-center gap-3 border-t border-edge pt-4">
            <button
              onClick={handleSubmitEdit}
              disabled={submitEdit.isPending}
              className="btn-primary flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {submitEdit.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <CheckCircle2 className="h-4 w-4" />}
              Submit Changes
            </button>
            <button
              onClick={handleReleaseLock}
              disabled={releaseLock.isPending}
              className="rounded-md border border-edge px-4 py-2 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
            >
              <Unlock className="mr-1 inline h-4 w-4" />
              Release Lock
            </button>
          </div>
          {submitEdit.isError && (
            <p className="text-sm text-red-400">
              <AlertCircle className="mr-1 inline h-4 w-4" />
              {submitEdit.error?.message || 'Failed to submit edit'}
            </p>
          )}
          {submitEdit.isSuccess && (
            <div className="flex items-center gap-2 rounded-md bg-[#39ff14]/10 border border-[#39ff14]/30 p-3 text-sm text-[#39ff14]">
              <CheckCircle2 className="h-4 w-4" />
              Changes submitted successfully.
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// --- Shared field components ---

function DraftField({
  label,
  value,
  onChange,
  multiline,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  multiline?: boolean;
}) {
  const cls = 'w-full rounded-md border border-edge bg-overlay px-3 py-2 text-sm text-[var(--color-text)] placeholder-slate-500 focus:border-[var(--color-primary)] focus:outline-none';
  return (
    <div>
      <label className="mb-1 block text-xs font-medium text-[var(--color-text-muted)]">{label}</label>
      {multiline ? (
        <textarea value={value} onChange={(e) => onChange(e.target.value)} rows={3} className={cls} />
      ) : (
        <input type="text" value={value} onChange={(e) => onChange(e.target.value)} className={cls} />
      )}
    </div>
  );
}

function DraftTagField({
  label,
  values,
  onChange,
}: {
  label: string;
  values: string[];
  onChange: (v: string[]) => void;
}) {
  const [input, setInput] = useState('');

  const addTag = () => {
    const tag = input.trim();
    if (tag && !values.includes(tag)) {
      onChange([...values, tag]);
    }
    setInput('');
  };

  return (
    <div>
      <label className="mb-1 block text-xs font-medium text-[var(--color-text-muted)]">{label}</label>
      <div className="flex flex-wrap gap-1.5 mb-2">
        {values.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 rounded-full bg-[var(--color-primary)]/10 px-2 py-0.5 text-xs text-[var(--color-primary)]"
          >
            {tag}
            <button
              onClick={() => onChange(values.filter((t) => t !== tag))}
              className="text-[var(--color-primary)]/50 hover:text-[var(--color-primary)]"
            >
              ×
            </button>
          </span>
        ))}
      </div>
      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addTag(); } }}
          placeholder={`Add ${label.toLowerCase()}...`}
          className="flex-1 rounded-md border border-edge bg-overlay px-3 py-1.5 text-sm text-[var(--color-text)] placeholder-slate-500 focus:border-[var(--color-primary)] focus:outline-none"
        />
        <button onClick={addTag} className="rounded-md border border-edge px-3 py-1.5 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)]">
          Add
        </button>
      </div>
    </div>
  );
}

// --- Page ---

export default function ModeratorPage() {
  const { isModerator, user } = useAuthStore();

  if (!isModerator) {
    return (
      <div className="flex items-center justify-center p-12">
        <div className="text-center">
          <Shield className="mx-auto h-12 w-12 text-[var(--color-text-muted)]" />
          <p className="mt-4 text-[var(--color-text-muted)]">Moderator access required.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6 p-6">
      <div className="flex items-center gap-3">
        <Shield className="h-6 w-6 text-[var(--color-primary)]" />
        <div>
          <h1 className="text-xl font-bold text-[var(--color-text)]">Moderator Tools</h1>
          <p className="text-sm text-[var(--color-text-muted)]">
            Signed in as{' '}
            <span className="rounded bg-[var(--color-primary)]/10 px-1.5 py-0.5 text-xs font-medium text-[var(--color-primary)]">
              {user?.role}
            </span>{' '}
            — {user?.name}
          </p>
        </div>
      </div>

      <GenerateCompanySection />
      <EditCompanySection />
    </div>
  );
}
