import type { CompanyTrackingStatus } from '@/types/api';

const statusConfig: Record<
  CompanyTrackingStatus,
  { label: string; className: string }
> = {
  marked: { label: 'Marked', className: 'bg-slate-800/60 text-slate-300 border border-slate-700' },
  researching: { label: 'Researching', className: 'bg-[#ff00e5]/10 text-[#e040fb] border border-[#ff00e5]/20' },
  applied: { label: 'Applied', className: 'bg-[#00f0ff]/10 text-[#00f0ff] border border-[#00f0ff]/20' },
  interviewing: { label: 'Interviewing', className: 'bg-[#ffb800]/10 text-[#ffb800] border border-[#ffb800]/20' },
  offered: { label: 'Offered', className: 'bg-[#39ff14]/10 text-[#39ff14] border border-[#39ff14]/20' },
  accepted: { label: 'Accepted', className: 'bg-[#39ff14]/15 text-[#39ff14] border border-[#39ff14]/30 glow-green' },
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
