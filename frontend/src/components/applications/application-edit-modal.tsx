'use client';

import { useEffect, useState } from 'react';
import { X } from 'lucide-react';
import { useUpdateApplication } from '@/hooks/use-applications';
import { ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import type { Application, ApplicationStatus } from '@/types/api';

interface ApplicationEditModalProps {
  application: Application;
  companyName?: string;
  onClose: () => void;
}

export function ApplicationEditModal({
  application,
  companyName,
  onClose,
}: ApplicationEditModalProps) {
  const updateApp = useUpdateApplication();
  const [roleTitle, setRoleTitle] = useState(application.role_title ?? '');
  const [status, setStatus] = useState<ApplicationStatus>(application.status);
  const [dateApplied, setDateApplied] = useState(application.date_applied ?? '');
  const [notes, setNotes] = useState(application.notes ?? '');

  useEffect(() => {
    setRoleTitle(application.role_title ?? '');
    setStatus(application.status);
    setDateApplied(application.date_applied ?? '');
    setNotes(application.notes ?? '');
  }, [application]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await updateApp.mutateAsync({
      id: application.id,
      role_title: roleTitle,
      status,
      date_applied: dateApplied,
      notes,
    });
    onClose();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-lg rounded-lg border border-edge bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-edge px-5 py-4">
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Edit Application</h3>
            <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">
              {companyName ?? application.company_name ?? 'Application details'}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-5 py-4">
          <div>
            <label htmlFor="editRoleTitle" className="block text-sm font-medium text-[var(--color-text)]">
              Role Title
            </label>
            <input
              id="editRoleTitle"
              type="text"
              value={roleTitle}
              onChange={(e) => setRoleTitle(e.target.value)}
              placeholder="e.g. Senior Software Engineer"
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            />
          </div>

          <div>
            <label htmlFor="editStatus" className="block text-sm font-medium text-[var(--color-text)]">
              Status
            </label>
            <select
              id="editStatus"
              value={status}
              onChange={(e) => setStatus(e.target.value as ApplicationStatus)}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            >
              {ALL_STATUSES.map((item) => (
                <option key={item} value={item}>
                  {getStatusLabel(item)}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="editDateApplied" className="block text-sm font-medium text-[var(--color-text)]">
              Date Applied
            </label>
            <input
              id="editDateApplied"
              type="date"
              value={dateApplied}
              onChange={(e) => setDateApplied(e.target.value)}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
            />
          </div>

          <div>
            <label htmlFor="editNotes" className="block text-sm font-medium text-[var(--color-text)]">
              Notes
            </label>
            <textarea
              id="editNotes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Add notes about this application..."
              rows={6}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
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
              disabled={updateApp.isPending}
              className="btn-primary rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {updateApp.isPending ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
