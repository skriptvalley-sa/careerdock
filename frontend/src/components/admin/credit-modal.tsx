'use client';

import { useState, useEffect } from 'react';
import { X } from 'lucide-react';
import { useAdminAllocateCredits } from '@/hooks/use-admin';

const CREDIT_TYPES = [
  { value: 'resume_upload', label: 'Resume Upload' },
  { value: 'ats_check', label: 'ATS Check' },
  { value: 'curated_list', label: 'Curated List' },
  { value: 'cv_generation', label: 'Cover Letter' },
];

interface CreditModalProps {
  userId: string;
  userName: string;
  onClose: () => void;
}

export function CreditModal({ userId, userName, onClose }: CreditModalProps) {
  const [creditType, setCreditType] = useState('resume_upload');
  const [amount, setAmount] = useState(1);
  const [reason, setReason] = useState('');
  const [error, setError] = useState('');
  const allocate = useAdminAllocateCredits();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [onClose]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      await allocate.mutateAsync({
        userId,
        credit_type: creditType,
        amount,
        reason,
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    }
  };

  const inputClass =
    'mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        className="w-full max-w-sm rounded-lg border border-edge bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-edge px-5 py-4">
          <div>
            <h3 className="text-sm font-semibold text-[var(--color-text)]">Allocate Credits</h3>
            <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">{userName}</p>
          </div>
          <button type="button" onClick={onClose} className="text-[var(--color-text-muted)] hover:text-[var(--color-text)]">
            <X className="h-5 w-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-5 py-4">
          {error && (
            <div className="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-400">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="creditType" className="block text-sm font-medium text-[var(--color-text)]">
              Credit Type
            </label>
            <select
              id="creditType"
              value={creditType}
              onChange={(e) => setCreditType(e.target.value)}
              className={inputClass}
            >
              {CREDIT_TYPES.map((ct) => (
                <option key={ct.value} value={ct.value}>
                  {ct.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="amount" className="block text-sm font-medium text-[var(--color-text)]">
              Amount
            </label>
            <input
              id="amount"
              type="number"
              min={1}
              value={amount}
              onChange={(e) => setAmount(parseInt(e.target.value) || 1)}
              className={inputClass}
            />
          </div>

          <div>
            <label htmlFor="reason" className="block text-sm font-medium text-[var(--color-text)]">
              Reason *
            </label>
            <input
              id="reason"
              required
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="e.g. Customer support request"
              className={inputClass}
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-card hover:text-[var(--color-text)]"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={allocate.isPending}
              className="btn-primary rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
            >
              {allocate.isPending ? 'Allocating...' : 'Allocate'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
