'use client';

import { useState } from 'react';
import { X } from 'lucide-react';
import { useCreateApplication } from '@/hooks/use-applications';
import { ALL_STATUSES, getStatusLabel } from '@/components/lists/status-badge';
import type { ApplicationStatus, ListEntry } from '@/types/api';

interface AddApplicationModalProps {
  entry: ListEntry;
  onClose: () => void;
}

export function AddApplicationModal({ entry, onClose }: AddApplicationModalProps) {
  const [roleTitle, setRoleTitle] = useState('');
  const [status, setStatus] = useState<ApplicationStatus>('applied');
  const [dateApplied, setDateApplied] = useState('');
  const createApp = useCreateApplication();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createApp.mutateAsync({
      company_id: entry.company_id,
      role_title: roleTitle || undefined,
      status,
      date_applied: dateApplied || undefined,
    });
    onClose();
  };

  // Filter out not_applied from status options — this is an application form
  const statusOptions = ALL_STATUSES.filter((s) => s !== 'not_applied');

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-full max-w-md rounded-lg border border-edge bg-card shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-edge px-5 py-4">
          <div>
            <h3 className="text-sm font-semibold text-slate-100">Add Application</h3>
            <p className="mt-0.5 text-xs text-slate-500">{entry.company_name}</p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-slate-400 hover:text-slate-200"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4 px-5 py-4">
          <div>
            <label htmlFor="roleTitle" className="block text-sm font-medium text-slate-300">
              Role Title
            </label>
            <input
              id="roleTitle"
              type="text"
              value={roleTitle}
              onChange={(e) => setRoleTitle(e.target.value)}
              placeholder="e.g. Senior Software Engineer"
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 placeholder:text-slate-600 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            />
          </div>

          <div>
            <label htmlFor="appStatus" className="block text-sm font-medium text-slate-300">
              Application Status
            </label>
            <select
              id="appStatus"
              value={status}
              onChange={(e) => setStatus(e.target.value as ApplicationStatus)}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            >
              {statusOptions.map((s) => (
                <option key={s} value={s}>
                  {getStatusLabel(s)}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="dateApplied" className="block text-sm font-medium text-slate-300">
              Date Applied
            </label>
            <input
              id="dateApplied"
              type="date"
              value={dateApplied}
              onChange={(e) => setDateApplied(e.target.value)}
              className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-slate-300 hover:bg-overlay"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createApp.isPending}
              className="btn-neon rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {createApp.isPending ? 'Saving...' : 'Save Application'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
