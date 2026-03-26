'use client';

import { X } from 'lucide-react';

const SIZE_OPTIONS = [
  { value: 'startup', label: 'Startup' },
  { value: 'small', label: 'Small' },
  { value: 'mid', label: 'Mid-size' },
  { value: 'large', label: 'Large' },
  { value: 'enterprise', label: 'Enterprise' },
];

const HIRING_OPTIONS = [
  { value: 'active', label: 'Hiring' },
  { value: 'paused', label: 'Paused' },
];

const TIER_OPTIONS = [
  { value: 'tier_1', label: 'T1 (>40L)', color: 'border-[var(--color-warning)] text-[var(--color-warning)] bg-[var(--color-warning)]/15' },
  { value: 'tier_2', label: 'T2 (20-40L)', color: 'border-green-500 text-green-400 bg-green-900/20' },
  { value: 'tier_3', label: 'T3 (10-20L)', color: 'border-[var(--color-primary)]/50 text-[var(--color-primary)] bg-[var(--color-primary)]/10' },
  { value: 'tier_4', label: 'T4 (5-10L)', color: 'border-[var(--color-edge-hover)] text-[var(--color-text-muted)] bg-[var(--color-primary)]/10' },
];

const SORT_OPTIONS = [
  { value: 'name', label: 'Name' },
  { value: 'updated_at', label: 'Recently Updated' },
  { value: 'compensation_tier', label: 'Compensation' },
  { value: 'size', label: 'Company Size' },
];

export interface FilterValues {
  sizes: string[];
  hiring_status: string;
  compensation_tiers: string[];
  has_rsu: string;
  sort: string;
  order: string;
}

interface CompanyFiltersProps {
  values: FilterValues;
  onChange: (values: FilterValues) => void;
}

function toggleArrayValue(arr: string[], value: string): string[] {
  return arr.includes(value) ? arr.filter((v) => v !== value) : [...arr, value];
}

export function CompanyFilters({ values, onChange }: CompanyFiltersProps) {
  const hasActiveFilters =
    values.sizes.length > 0 ||
    values.hiring_status ||
    values.compensation_tiers.length > 0 ||
    values.has_rsu;

  const clearFilters = () => {
    onChange({
      sizes: [],
      hiring_status: '',
      compensation_tiers: [],
      has_rsu: '',
      sort: values.sort,
      order: values.order,
    });
  };

  return (
    <div className="space-y-3">
      {/* Row 1: Tier chips + RSU toggles + Sort */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Tier chips */}
        {TIER_OPTIONS.map((tier) => {
          const active = values.compensation_tiers.includes(tier.value);
          return (
            <button
              key={tier.value}
              type="button"
              onClick={() =>
                onChange({
                  ...values,
                  compensation_tiers: toggleArrayValue(values.compensation_tiers, tier.value),
                })
              }
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
                active
                  ? tier.color
                  : 'border-edge text-[var(--color-text-muted)] hover:border-edge-hover hover:text-[var(--color-text-muted)]'
              }`}
            >
              {tier.label}
            </button>
          );
        })}

        {/* Divider */}
        <div className="mx-1 h-5 w-px bg-edge" />

        {/* RSU toggle */}
        <button
          type="button"
          onClick={() =>
            onChange({
              ...values,
              has_rsu: values.has_rsu === 'true' ? '' : 'true',
            })
          }
          className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
            values.has_rsu === 'true'
              ? 'border-green-700 bg-green-900/30 text-green-400'
              : 'border-edge text-[var(--color-text-muted)] hover:border-edge-hover hover:text-[var(--color-text-muted)]'
          }`}
        >
          RSU
        </button>

        {/* Divider */}
        <div className="mx-1 h-5 w-px bg-edge" />

        {/* Size chips */}
        {SIZE_OPTIONS.map((opt) => {
          const active = values.sizes.includes(opt.value);
          return (
            <button
              key={opt.value}
              type="button"
              onClick={() =>
                onChange({
                  ...values,
                  sizes: toggleArrayValue(values.sizes, opt.value),
                })
              }
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
                active
                  ? 'border-slate-500 bg-slate-800 text-[var(--color-text)]'
                  : 'border-edge text-[var(--color-text-muted)] hover:border-edge-hover hover:text-[var(--color-text-muted)]'
              }`}
            >
              {opt.label}
            </button>
          );
        })}

        {/* Divider */}
        <div className="mx-1 h-5 w-px bg-edge" />

        {/* Hiring status chips */}
        {HIRING_OPTIONS.map((opt) => {
          const active = values.hiring_status === opt.value;
          return (
            <button
              key={opt.value}
              type="button"
              onClick={() =>
                onChange({
                  ...values,
                  hiring_status: active ? '' : opt.value,
                })
              }
              className={`rounded-full border px-3 py-1 text-xs font-medium transition-all ${
                active
                  ? opt.value === 'active'
                    ? 'border-green-700 bg-green-900/30 text-green-400'
                    : 'border-[var(--color-warning)] bg-[var(--color-warning)]/15 text-[var(--color-warning)]'
                  : 'border-edge text-[var(--color-text-muted)] hover:border-edge-hover hover:text-[var(--color-text-muted)]'
              }`}
            >
              {opt.label}
            </button>
          );
        })}

        {/* Spacer */}
        <div className="flex-1" />

        {/* Sort controls */}
        <div className="flex items-center gap-2">
          <select
            value={values.sort}
            onChange={(e) => onChange({ ...values, sort: e.target.value })}
            className="rounded-md border border-edge bg-input px-2.5 py-1 text-xs text-[var(--color-text)] focus:border-[var(--color-primary)]/50 focus:outline-none"
          >
            {SORT_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={() => onChange({ ...values, order: values.order === 'asc' ? 'desc' : 'asc' })}
            className="rounded-md border border-edge px-2.5 py-1 text-xs text-[var(--color-text-muted)] hover:bg-card"
          >
            {values.order === 'desc' ? '↓' : '↑'}
          </button>
        </div>
      </div>

      {/* Active filter summary + clear */}
      {hasActiveFilters && (
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={clearFilters}
            className="flex items-center gap-1 text-xs text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
          >
            <X className="h-3 w-3" />
            Clear filters
          </button>
        </div>
      )}
    </div>
  );
}
