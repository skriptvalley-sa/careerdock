import type { CompanyTrackingStatus } from '@/types/api';

const statusConfig: Record<
  CompanyTrackingStatus,
  { label: string; className: string }
> = {
  marked: { label: 'Marked', className: 'bg-slate-800/60 text-[var(--color-text)] border border-slate-700' },
  researching: { label: 'Researching', className: 'bg-[#ff00e5]/10 text-[#e040fb] border border-[#ff00e5]/20' },
  applied: { label: 'Applied', className: 'bg-[var(--color-primary)]/10 text-[var(--color-primary)] border border-[var(--color-primary)]/20' },
  interviewing: { label: 'Interviewing', className: 'bg-[var(--color-warning)]/10 text-[var(--color-warning)] border border-[var(--color-warning)]/20' },
  offered: { label: 'Offered', className: 'bg-[var(--color-success)]/10 text-[var(--color-success)] border border-[var(--color-success)]/20' },
  accepted: { label: 'Accepted', className: 'bg-[var(--color-success)]/15 text-[var(--color-success)] border border-[var(--color-success)]/30' },
  rejected: { label: 'Rejected', className: 'bg-red-900/30 text-red-400 border border-red-800/30' },
};

export const ALL_COMPANY_STATUSES: CompanyTrackingStatus[] = [
  'marked',
  'researching',
  'applied',
  'interviewing',
  'offered',
  'accepted',
  'rejected',
];

export function CompanyStatusBadge({ status }: { status: CompanyTrackingStatus }) {
  const config = statusConfig[status] || statusConfig.marked;
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${config.className}`}
    >
      {config.label}
    </span>
  );
}

export function getCompanyStatusLabel(status: CompanyTrackingStatus): string {
  return statusConfig[status]?.label || status;
}
