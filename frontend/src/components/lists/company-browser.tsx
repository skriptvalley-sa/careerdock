'use client';

import { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { Search, X, Check, Minus } from 'lucide-react';
import { useCompanyList } from '@/hooks/use-companies';
import { CompanyFilters, type FilterValues } from '@/components/companies/company-filters';
import type { CompanyFilterParams } from '@/types/api';

interface CompanyBrowserProps {
  /** Company IDs already in the list (pre-selected for edit mode) */
  existingCompanyIds: Set<string>;
  /** Called with the full desired set of company IDs (additions + retained) */
  onSave: (companyIds: string[]) => void;
  onCancel: () => void;
  isSaving?: boolean;
  /** Fires when the diff status (has unsaved changes) changes */
  onDiffChange?: (hasChanges: boolean) => void;
  /** Fires when the selected set changes (parent can use for header Save) */
  onSelectedChange?: (selectedIds: string[]) => void;
}

export function CompanyBrowserPanel({
  existingCompanyIds,
  onSave,
  onCancel,
  isSaving,
  onDiffChange,
  onSelectedChange,
}: CompanyBrowserProps) {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  // selected = the full desired set; initialized from existingCompanyIds
  const [selected, setSelected] = useState<Set<string>>(() => new Set(existingCompanyIds));
  const [filters, setFilters] = useState<FilterValues>({
    sizes: [],
    hiring_status: '',
    compensation_tiers: [],
    has_rsu: '',
    sort: 'name',
    order: 'asc',
  });
  const bottomRef = useRef<HTMLDivElement>(null);

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(query), 300);
    return () => clearTimeout(timer);
  }, [query]);

  // Build params for the company query
  const params: CompanyFilterParams = {
    q: debouncedQuery || undefined,
    size: filters.sizes.join(',') || undefined,
    hiring_status: filters.hiring_status || undefined,
    compensation_tier: filters.compensation_tiers.join(',') || undefined,
    has_rsu: filters.has_rsu || undefined,
    sort: filters.sort,
    order: filters.order,
    limit: '20',
  };

  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useCompanyList(params);

  // Infinite scroll: load more when user scrolls near bottom
  useEffect(() => {
    if (!bottomRef.current || !hasNextPage) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
          fetchNextPage();
        }
      },
      { threshold: 0.5 },
    );
    observer.observe(bottomRef.current);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  const companies = data?.pages.flatMap((p) => p.data) ?? [];

  const toggleCompany = useCallback(
    (companyId: string) => {
      setSelected((prev) => {
        const next = new Set(prev);
        if (next.has(companyId)) {
          next.delete(companyId);
        } else {
          next.add(companyId);
        }
        return next;
      });
    },
    [],
  );

  const handleSave = () => {
    onSave(Array.from(selected));
  };

  // Compute diff from original
  const diff = useMemo(() => {
    let added = 0;
    let removed = 0;
    // Count new additions (in selected but not in existing)
    for (const id of selected) {
      if (!existingCompanyIds.has(id)) added++;
    }
    // Count removals (in existing but not in selected)
    for (const id of existingCompanyIds) {
      if (!selected.has(id)) removed++;
    }
    return { added, removed, hasChanges: added > 0 || removed > 0 };
  }, [selected, existingCompanyIds]);

  // Notify parent when diff status changes
  useEffect(() => {
    onDiffChange?.(diff.hasChanges);
  }, [diff.hasChanges, onDiffChange]);

  // Notify parent when selected set changes
  useEffect(() => {
    onSelectedChange?.(Array.from(selected));
  }, [selected, onSelectedChange]);

  return (
    <div className="mt-6 rounded-lg border border-edge bg-card">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-edge px-4 py-3">
        <h3 className="text-sm font-semibold text-[var(--color-text)]">Edit list companies</h3>
        <button
          type="button"
          onClick={onCancel}
          className="text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        >
          <X className="h-5 w-5" />
        </button>
      </div>

      {/* Search */}
      <div className="border-b border-edge px-4 py-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--color-text-muted)]" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search companies..."
            className="block w-full rounded-md border border-edge-input bg-input py-2 pl-9 pr-3 text-sm text-[var(--color-text)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
          />
        </div>
      </div>

      {/* Filters */}
      <div className="border-b border-edge px-4 py-3">
        <CompanyFilters values={filters} onChange={setFilters} />
      </div>

      {/* Company grid */}
      <div className="max-h-[420px] overflow-y-auto p-4">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
          </div>
        ) : companies.length === 0 ? (
          <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">
            No companies found. Try adjusting your search or filters.
          </div>
        ) : (
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {companies.map((company) => {
              const isSelected = selected.has(company.id);
              const wasExisting = existingCompanyIds.has(company.id);
              const isRemoved = wasExisting && !isSelected;

              return (
                <button
                  key={company.id}
                  type="button"
                  onClick={() => toggleCompany(company.id)}
                  className={`group relative flex items-start gap-3 rounded-lg border p-3 text-left transition-all ${
                    isRemoved
                      ? 'border-red-500/50 bg-red-500/10 opacity-70'
                      : isSelected
                        ? 'border-[var(--color-primary)]/50 bg-[var(--color-primary)]/10 ring-1 ring-[var(--color-primary)]/30'
                        : 'border-edge hover:border-[var(--color-primary)]/30 hover:bg-surface'
                  }`}
                >
                  {/* Selection indicator */}
                  <div
                    className={`mt-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded ${
                      isRemoved
                        ? 'bg-red-500/20 text-red-400'
                        : isSelected
                          ? 'bg-[var(--color-primary)] text-[#06080f]'
                          : 'border border-edge group-hover:border-slate-500'
                    }`}
                  >
                    {isRemoved && <Minus className="h-3 w-3" />}
                    {isSelected && !isRemoved && <Check className="h-3 w-3" />}
                  </div>

                  {/* Company info */}
                  <div className="min-w-0 flex-1">
                    <div className={`truncate text-sm font-medium ${isRemoved ? 'text-red-400 line-through' : 'text-[var(--color-text)]'}`}>
                      {company.name}
                    </div>
                    <div className="mt-0.5 flex items-center gap-1.5 text-xs text-[var(--color-text-muted)]">
                      {company.headquarters && <span>{company.headquarters}</span>}
                      {company.headquarters && company.size && <span>·</span>}
                      {company.size && <span>{company.size}</span>}
                    </div>
                    {company.hiring_status === 'active' && (
                      <span className="mt-1 inline-block rounded-full bg-green-900/30 px-1.5 py-0.5 text-[10px] font-medium text-green-400">
                        Hiring
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        )}

        {/* Infinite scroll sentinel */}
        <div ref={bottomRef} className="h-1" />
        {isFetchingNextPage && (
          <div className="flex justify-center py-3">
            <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--color-primary)] border-t-transparent" />
          </div>
        )}
      </div>

      {/* Footer: diff summary + save */}
      <div className="flex items-center justify-between border-t border-edge px-4 py-3">
        <div className="flex items-center gap-3 text-sm text-[var(--color-text-muted)]">
          {diff.hasChanges ? (
            <>
              {diff.added > 0 && (
                <span className="font-medium text-[#39ff14]">+{diff.added} added</span>
              )}
              {diff.removed > 0 && (
                <span className="font-medium text-red-400">-{diff.removed} removed</span>
              )}
            </>
          ) : (
            'Select or deselect companies to edit'
          )}
        </div>
        <div className="flex gap-3">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md border border-edge px-4 py-2 text-sm font-medium text-[var(--color-text)] hover:bg-overlay"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleSave}
            disabled={!diff.hasChanges || isSaving}
            className="btn-primary rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50"
          >
            {isSaving ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </div>
    </div>
  );
}
