'use client';

import { X } from 'lucide-react';
import type { Application } from '@/types/api';

interface ApplicationNotesModalProps {
  application: Application;
  companyName?: string;
  onClose: () => void;
}

export function ApplicationNotesModal({
  application,
  companyName,
  onClose,
}: ApplicationNotesModalProps) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 px-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-2xl rounded-lg border border-edge bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-edge px-5 py-4">
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Application Notes</h3>
            <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">
              {companyName ?? application.company_name ?? 'Application'}
              {application.role_title ? ` · ${application.role_title}` : ''}
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

        <div className="px-5 py-4">
          <div className="rounded-lg border border-edge bg-overlay px-4 py-4">
            <p className="whitespace-pre-wrap text-sm leading-6 text-[var(--color-text)]">
              {application.notes || 'No notes added for this application yet.'}
            </p>
          </div>
        </div>

        <div className="flex justify-end border-t border-edge px-5 py-3">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-overlay"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
