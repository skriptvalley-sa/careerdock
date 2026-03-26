import type { ApplicationStatus } from '@/types/api';

const statusConfig: Record<
  ApplicationStatus,
  { label: string; className: string }
> = {
  not_applied: { label: 'Not Applied', className: 'bg-slate-800/60 text-[var(--color-text-muted)] border border-slate-700' },
  applied: { label: 'Applied', className: 'bg-[var(--color-primary)]/10 text-[var(--color-primary)] border border-[var(--color-primary)]/20' },
  phone_screen: { label: 'Phone Screen', className: 'bg-[#ff00e5]/10 text-[#e040fb] border border-[#ff00e5]/20' },
  interview: { label: 'Interview', className: 'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20' },
  offer: { label: 'Offer', className: 'bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20' },
  rejected: { label: 'Rejected', className: 'bg-red-900/30 text-red-400 border border-red-800/30' },
  accepted: { label: 'Accepted', className: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border border-[var(--color-success)]/30' },
  withdrawn: { label: 'Withdrawn', className: 'bg-orange-900/30 text-orange-400 border border-orange-800/30' },
};

export const ALL_STATUSES: ApplicationStatus[] = [
  'not_applied',
  'applied',
  'phone_screen',
  'interview',
  'offer',
  'rejected',
  'accepted',
  'withdrawn',
];

export function StatusBadge({ status }: { status: ApplicationStatus }) {
  const config = statusConfig[status] || statusConfig.not_applied;
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${config.className}`}
    >
      {config.label}
    </span>
  );
}

export function getStatusLabel(status: ApplicationStatus): string {
  return statusConfig[status]?.label || status;
}
