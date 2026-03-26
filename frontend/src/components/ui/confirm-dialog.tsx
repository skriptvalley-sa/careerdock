'use client';

import { useEffect } from 'react';

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  /** Optional secondary action label (e.g. "Discard") */
  secondaryLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
  /** Optional secondary action handler */
  onSecondary?: () => void;
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  secondaryLabel,
  onConfirm,
  onCancel,
  onSecondary,
}: ConfirmDialogProps) {
  // Close on Escape
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open, onCancel]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      onClick={onCancel}
    >
      <div
        className="w-full max-w-sm rounded-lg border border-edge bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="border-b border-edge px-5 py-4">
          <h3 className="text-sm font-semibold text-[var(--color-text)]">{title}</h3>
        </div>

        {/* Body */}
        <div className="px-5 py-4">
          <p className="text-sm text-[var(--color-text)]">{message}</p>
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-3 border-t border-edge px-5 py-3">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-overlay"
          >
            {cancelLabel}
          </button>
          {secondaryLabel && onSecondary && (
            <button
              type="button"
              onClick={onSecondary}
              className="rounded-md border border-red-500/30 px-4 py-2 text-sm font-medium text-red-400 hover:bg-red-500/10"
            >
              {secondaryLabel}
            </button>
          )}
          <button
            type="button"
            onClick={onConfirm}
            className="btn-primary rounded-md px-4 py-2 text-sm font-medium"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
