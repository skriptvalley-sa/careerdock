'use client';

import { Filter } from 'lucide-react';
import { useState } from 'react';

const SIZE_OPTIONS = [
  { value: 'startup', label: 'Startup' },
  { value: 'small', label: 'Small' },
  { value: 'mid', label: 'Mid-size' },
  { value: 'large', label: 'Large' },
  { value: 'enterprise', label: 'Enterprise' },
];

const HIRING_OPTIONS = [
  { value: 'active', label: 'Actively Hiring' },
  { value: 'paused', label: 'Paused' },
  { value: 'unknown', label: 'Unknown' },
];

const TIER_OPTIONS = [
  { value: 'tier_1', label: 'Tier 1 (>40L)' },
  { value: 'tier_2', label: 'Tier 2 (20-40L)' },
  { value: 'tier_3', label: 'Tier 3 (10-20L)' },
  { value: 'tier_4', label: 'Tier 4 (5-10L)' },
];

const SORT_OPTIONS = [
  { value: 'name', label: 'Name' },
  { value: 'updated_at', label: 'Recently Updated' },
  { value: 'compensation_tier', label: 'Compensation' },
  { value: 'size', label: 'Company Size' },
];

export interface FilterValues {
  size: string;
  hiring_status: string;
  compensation_tier: string;
  has_rsu: string;
  sort: string;
  order: string;
}

interface CompanyFiltersProps {
  values: FilterValues;
  onChange: (values: FilterValues) => void;
}

export function CompanyFilters({ values, onChange }: CompanyFiltersProps) {
  const [isOpen, setIsOpen] = useState(false);

  const update = (key: keyof FilterValues, value: string) => {
    onChange({ ...values, [key]: value });
  };

  const hasActiveFilters =
    values.size || values.hiring_status || values.compensation_tier || values.has_rsu;

  const clearFilters = () => {
    onChange({
      size: '',
      hiring_status: '',
      compensation_tier: '',
      has_rsu: '',
      sort: values.sort,
      order: values.order,
    });
  };

  return (
    <div>
      <div className="flex items-center justify-between gap-4">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center gap-2 rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
        >
          <Filter className="h-4 w-4" />
          Filters
          {hasActiveFilters && (
            <span className="rounded-full bg-blue-600 px-1.5 py-0.5 text-xs text-white">!</span>
          )}
        </button>

        <div className="flex items-center gap-2">
          <select
            value={values.sort}
            onChange={(e) => update('sort', e.target.value)}
            className="rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
          >
            {SORT_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <button
            onClick={() => update('order', values.order === 'asc' ? 'desc' : 'asc')}
            className="rounded-lg border border-gray-300 px-3 py-2 text-sm hover:bg-gray-50"
          >
            {values.order === 'desc' ? 'Z-A' : 'A-Z'}
          </button>
        </div>
      </div>

      {isOpen && (
        <div className="mt-4 grid grid-cols-2 gap-4 rounded-lg border border-gray-200 bg-gray-50 p-4 md:grid-cols-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-700">Company Size</label>
            <select
              value={values.size}
              onChange={(e) => update('size', e.target.value)}
              className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm"
            >
              <option value="">All sizes</option>
              {SIZE_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-gray-700">Hiring Status</label>
            <select
              value={values.hiring_status}
              onChange={(e) => update('hiring_status', e.target.value)}
              className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm"
            >
              <option value="">All statuses</option>
              {HIRING_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-gray-700">
              Compensation Tier
            </label>
            <select
              value={values.compensation_tier}
              onChange={(e) => update('compensation_tier', e.target.value)}
              className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm"
            >
              <option value="">All tiers</option>
              {TIER_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-gray-700">RSU</label>
            <select
              value={values.has_rsu}
              onChange={(e) => update('has_rsu', e.target.value)}
              className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm"
            >
              <option value="">Any</option>
              <option value="true">Has RSU</option>
              <option value="false">No RSU</option>
            </select>
          </div>

          {hasActiveFilters && (
            <div className="col-span-full">
              <button onClick={clearFilters} className="text-sm text-blue-600 hover:text-blue-700">
                Clear all filters
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
