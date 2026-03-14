'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useCompanyList } from '@/hooks/use-companies';
import { useCompanyListCounts } from '@/hooks/use-lists';
import { useAuthStore } from '@/store/auth-store';
import { CompanyCard } from '@/components/companies/company-card';
import { CompanySearchBar } from '@/components/companies/company-search-bar';
import { CompanyFilters, type FilterValues } from '@/components/companies/company-filters';

export default function CompaniesPage() {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [filters, setFilters] = useState<FilterValues>({
    sizes: [],
    hiring_status: '',
    compensation_tiers: [],
    has_rsu: '',
    sort: 'name',
    order: 'asc',
  });

  // Debounce search
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  const params = {
    q: debouncedSearch || undefined,
    size: filters.sizes.length > 0 ? filters.sizes.join(',') : undefined,
    hiring_status: filters.hiring_status || undefined,
    compensation_tier: filters.compensation_tiers.length > 0 ? filters.compensation_tiers.join(',') : undefined,
    has_rsu: filters.has_rsu || undefined,
    sort: filters.sort || undefined,
    order: filters.order || undefined,
    limit: '20',
  };

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading, isError, error } =
    useCompanyList(params);

  // Load list membership counts for authenticated users (for card list indicators)
  const { isAuthenticated } = useAuthStore();
  const { data: listCounts } = useCompanyListCounts();
  const countsMap = isAuthenticated ? listCounts ?? {} : {};

  // Infinite scroll observer
  const loadMoreRef = useRef<HTMLDivElement>(null);
  const handleObserver = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (entry.isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage();
      }
    },
    [fetchNextPage, hasNextPage, isFetchingNextPage],
  );

  useEffect(() => {
    const el = loadMoreRef.current;
    if (!el) return;
    const observer = new IntersectionObserver(handleObserver, { threshold: 0.1 });
    observer.observe(el);
    return () => observer.disconnect();
  }, [handleObserver]);

  // Deduplicate across pages to prevent duplicate React key errors
  const companies = (() => {
    const all = data?.pages.flatMap((page) => page.data) ?? [];
    const seen = new Set<string>();
    return all.filter((c) => {
      if (seen.has(c.id)) return false;
      seen.add(c.id);
      return true;
    });
  })();

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-slate-100">Company Directory</h1>
        <p className="mt-2 text-slate-400">
          Browse {companies.length > 0 ? `${companies.length}+` : ''} Indian tech companies.
          Filter by size, tech stack, compensation, and more.
        </p>
      </div>

      <div className="mb-6 space-y-4">
        <CompanySearchBar value={search} onChange={setSearch} />
        <CompanyFilters values={filters} onChange={setFilters} />
      </div>

      {isLoading && (
        <div className="py-12 text-center text-slate-500">Loading companies...</div>
      )}

      {isError && (
        <div className="rounded-lg bg-red-900/30 p-4 text-sm text-red-400">
          Failed to load companies: {(error as Error).message}
        </div>
      )}

      {!isLoading && companies.length === 0 && (
        <div className="py-12 text-center text-slate-500">
          No companies found. Try adjusting your search or filters.
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {companies.map((company) => (
          <CompanyCard key={company.id} company={company} listCount={countsMap[company.id] ?? 0} />
        ))}
      </div>

      {/* Infinite scroll trigger */}
      <div ref={loadMoreRef} className="py-8 text-center">
        {isFetchingNextPage && (
          <span className="text-sm text-slate-500">Loading more...</span>
        )}
      </div>
    </div>
  );
}
