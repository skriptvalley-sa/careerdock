import Link from 'next/link';
import type { CompanyListItem } from '@/types/api';
import { TechStackTags } from './tech-stack-tags';

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
  active: 'bg-green-50 text-green-700',
  paused: 'bg-yellow-50 text-yellow-700',
  unknown: 'bg-gray-100 text-gray-600',
};

interface CompanyCardProps {
  company: CompanyListItem;
}

export function CompanyCard({ company }: CompanyCardProps) {
  return (
    <Link
      href={`/companies/${company.slug}`}
      className="block rounded-lg border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md"
    >
      <div className="flex items-start justify-between">
        <div className="min-w-0 flex-1">
          <h3 className="truncate text-lg font-semibold text-gray-900">{company.name}</h3>
          <div className="mt-1 flex flex-wrap items-center gap-2 text-sm text-gray-500">
            {company.headquarters && <span>{company.headquarters}</span>}
            {company.size && (
              <>
                <span className="text-gray-300">|</span>
                <span>{sizeLabels[company.size] || company.size}</span>
              </>
            )}
            {company.compensation_tier && (
              <>
                <span className="text-gray-300">|</span>
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

      {company.description && (
        <p className="mt-2 line-clamp-2 text-sm text-gray-600">{company.description}</p>
      )}

      <div className="mt-3">
        <TechStackTags tags={company.tech_stack} limit={5} />
      </div>

      {company.domains.length > 0 && (
        <div className="mt-2 flex flex-wrap gap-1.5">
          {company.domains.slice(0, 3).map((d) => (
            <span
              key={d}
              className="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs text-gray-600"
            >
              {d}
            </span>
          ))}
        </div>
      )}

      <div className="mt-3 flex items-center gap-3 text-xs text-gray-400">
        {company.has_rsu && <span>RSU</span>}
        {company.has_rsu_refresher && <span>Refresher</span>}
      </div>
    </Link>
  );
}
