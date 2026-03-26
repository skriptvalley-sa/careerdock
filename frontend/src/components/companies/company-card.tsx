'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Plus } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import type { CompanyListItem } from '@/types/api';
import { TechStackTags } from './tech-stack-tags';
import { QuickAddToListModal } from './quick-add-to-list-modal';

const sizeLabels: Record<string, string> = {
  startup: 'Startup',
  small: 'Small',
  mid: 'Mid-size',
  large: 'Large',
  enterprise: 'Enterprise',
};

const tierLabels: Record<string, string> = {
  tier_1: 'Tier 1',
  tier_2: 'Tier 2',
  tier_3: 'Tier 3',
  tier_4: 'Tier 4',
};

const hiringColors: Record<string, string> = {
  active: 'bg-green-900/30 text-green-400',
  paused: 'bg-yellow-900/30 text-yellow-400',
  unknown: 'bg-slate-800 text-[var(--color-text-muted)]',
};

const officeModeLabels: Record<string, string> = {
  remote: 'Remote',
  hybrid: 'Hybrid',
  onsite: 'On-site',
};

interface CompanyCardProps {
  company: CompanyListItem;
  /** Number of lists this company belongs to (0 = show +, >0 = show count) */
  listCount?: number;
}

export function CompanyCard({ company, listCount }: CompanyCardProps) {
  const { isAuthenticated } = useAuthStore();
  const [showAddModal, setShowAddModal] = useState(false);

  const handleListClick = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setShowAddModal(true);
  };

  return (
    <>
      <Link
        href={`/companies/${company.slug}`}
        className="card-hover group relative flex flex-col rounded-lg border border-edge bg-card p-5 transition-all"
      >
        {/* Header row: name + hiring badge */}
        <div className="flex items-start justify-between">
          <div className="min-w-0 flex-1">
            <h3 className="truncate text-lg font-semibold text-[var(--color-text)]">{company.name}</h3>
            <div className="mt-1 flex flex-wrap items-center gap-2 text-sm text-[var(--color-text-muted)]">
              {company.headquarters && <span>{company.headquarters}</span>}
              {company.size && (
                <>
                  <span className="text-[var(--color-text-muted)]">|</span>
                  <span>{sizeLabels[company.size] || company.size}</span>
                </>
              )}
              {company.compensation_tier && (
                <>
                  <span className="text-[var(--color-text-muted)]">|</span>
                  <span>{tierLabels[company.compensation_tier] || company.compensation_tier}</span>
                </>
              )}
            </div>
          </div>
          <span
            className={`ml-3 shrink-0 rounded-full px-2.5 py-0.5 text-xs font-medium ${hiringColors[company.hiring_status] || hiringColors.unknown}`}
          >
            {company.hiring_status === 'active'
              ? 'Hiring'
              : company.hiring_status === 'paused'
                ? 'Paused'
                : 'Unknown'}
          </span>
        </div>

        {/* Domains */}
        {company.domains.length > 0 && (
          <div className="mt-2.5 flex flex-wrap gap-1.5">
            {company.domains.slice(0, 4).map((d) => (
              <span
                key={d}
                className="inline-flex items-center rounded-full bg-[#ff00e5]/10 px-2.5 py-0.5 text-xs font-medium text-[#ff00e5]/80"
              >
                {d}
              </span>
            ))}
          </div>
        )}

        {/* Tech stack */}
        <div className="mt-2.5">
          <TechStackTags tags={company.tech_stack} limit={5} />
        </div>

        {/* Spacer to push bottom section down */}
        <div className="flex-1" />

        {/* Bottom row: RSU chips + office mode + list indicator */}
        <div className="mt-3 flex items-center justify-between">
          {/* Left: RSU + Refresher chips */}
          <div className="flex items-center gap-2">
            {company.has_rsu && (
              <span className="inline-flex items-center rounded-full bg-[#39ff14]/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-[#39ff14]">
                RSU
              </span>
            )}
            {company.has_rsu_refresher && (
              <span className="inline-flex items-center rounded-full bg-[#ffb800]/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-[#ffb800]">
                Refresher
              </span>
            )}
          </div>

          {/* Right: office mode + list indicator */}
          <div className="flex items-center gap-2">
            {/* Office mode chip */}
            {company.office_modes.length > 0 && (
              <span className="inline-flex items-center rounded-full bg-slate-800 px-2 py-0.5 text-[10px] font-medium text-[var(--color-text-muted)]">
                {company.office_modes.map((m) => officeModeLabels[m] || m).join(' / ')}
              </span>
            )}

            {/* List indicator (auth-only) */}
            {isAuthenticated && (
              <button
                type="button"
                onClick={handleListClick}
                className={`flex h-7 min-w-[28px] items-center justify-center rounded-md border text-xs font-medium transition-all ${
                  listCount && listCount > 0
                    ? 'border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 text-[var(--color-primary)] hover:bg-[var(--color-primary)]/20'
                    : 'border-edge bg-overlay text-[var(--color-text-muted)] hover:border-[var(--color-primary)]/50 hover:text-[var(--color-primary)]'
                }`}
                title={listCount && listCount > 0 ? `In ${listCount} list(s)` : 'Add to list'}
              >
                {listCount && listCount > 0 ? (
                  <span className="px-1">{listCount}</span>
                ) : (
                  <Plus className="h-4 w-4" />
                )}
              </button>
            )}
          </div>
        </div>
      </Link>

      {/* Quick-add modal */}
      {showAddModal && (
        <QuickAddToListModal
          companyId={company.id}
          companyName={company.name}
          onClose={() => setShowAddModal(false)}
        />
      )}
    </>
  );
}
