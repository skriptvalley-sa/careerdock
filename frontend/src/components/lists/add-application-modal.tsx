'use client';

import { useState } from 'react';
import { X } from 'lucide-react';
import { useCreateApplication } from '@/hooks/use-applications';
import { ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import { CompanyCombobox } from '@/components/companies/company-combobox';
import type { ApplicationStatus } from '@/types/api';

interface PreselectedCompany {
  id: string;
  name: string;
}

interface AddApplicationModalProps {
  /** Pre-selected company (from a list entry). When absent, a company combobox is shown. */
  company?: PreselectedCompany;
  onClose: () => void;
}

export function AddApplicationModal({ company, onClose }: AddApplicationModalProps) {
  const [selectedCompany, setSelectedCompany] = useState<PreselectedCompany | null>(
    company ?? null,
  );
  const [roleTitle, setRoleTitle] = useState('');
  const [status, setStatus] = useState<ApplicationStatus>('applied');
  const [dateApplied, setDateApplied] = useState('');
  const [notes, setNotes] = useState('');
  const createApp = useCreateApplication();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedCompany) return;
    await createApp.mutateAsync({
      company_id: selectedCompany.id,
      role_title: roleTitle || undefined,
      status,
      date_applied: dateApplied || undefined,
      notes: notes || undefined,
    });
    onClose();
  };

  // Filter out not_applied — this is an application creation form
  const statusOptions = ALL_STATUSES.filter((s) => s !== 'not_applied');

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-full max-w-md rounded-lg border border-edge bg-card shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-edge px-5 py-4">
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Add Application</h3>
            {company && (
              <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">{company.name}</p>
            )}
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4 px-5 py-4">
          {/* Company selector — only shown when no company is pre-selected */}
          {!company && (
            <div>
              <label className="block text-sm font-medium text-[var(--color-text)]">
                Company <span className="text-red-400">*</span>
              </label>
              <div className="mt-1">
                <CompanyCombobox
                  value={selectedCompany}
                  onChange={setSelectedCompany}
                  placeholder="Search companies..."
                />
              </div>
            </div>
          )}

          <div>
            <label htmlFor="roleTitle" className="block text-sm font-medium text-[var(--color-text)]">
              Role Title
            </label>
            <input
              id="roleTitle"
              type="text"
              value={roleTitle}
              onChange={(e) => setRoleTitle(e.target.value)}
              placeholder="e.g. Senior Software Engineer"
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            />
          </div>

          <div>
            <label htmlFor="appStatus" className="block text-sm font-medium text-[var(--color-text)]">
              Application Status
            </label>
            <select
              id="appStatus"
              value={status}
              onChange={(e) => setStatus(e.target.value as ApplicationStatus)}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            >
              {statusOptions.map((s) => (
                <option key={s} value={s}>
                  {getStatusLabel(s)}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="dateApplied" className="block text-sm font-medium text-[var(--color-text)]">
              Date Applied
            </label>
            <input
              id="dateApplied"
              type="date"
              value={dateApplied}
              onChange={(e) => setDateApplied(e.target.value)}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            />
          </div>

          <div>
            <label htmlFor="notes" className="block text-sm font-medium text-[var(--color-text)]">
              Notes
            </label>
            <textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Any notes about this application..."
              rows={3}
              className="mt-1 block w-full resize-none rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-overlay"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createApp.isPending || !selectedCompany}
              className="btn-primary rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {createApp.isPending ? 'Saving...' : 'Save Application'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
