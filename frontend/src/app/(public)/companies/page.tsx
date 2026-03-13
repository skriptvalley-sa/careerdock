'use client';

import { useState, useCallback, useEffect, useRef } from 'react';
import { useCompanyList } from '@/hooks/use-companies';
import { CompanyCard } from '@/components/companies/company-card';
import { CompanySearchBar } from '@/components/companies/company-search-bar';
import { CompanyFilters, type FilterValues } from '@/components/companies/company-filters';

export default function CompaniesPage() {
  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [filters, setFilters] = useState<FilterValues>({
    size: '',
    hiring_status: '',
    compensation_tier: '',
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
    size: filters.size || undefined,
    hiring_status: filters.hiring_status || undefined,
    compensation_tier: filters.compensation_tier || undefined,
    has_rsu: filters.has_rsu || undefined,
    sort: filters.sort || undefined,
    order: filters.order || undefined,
    limit: '20',
  };

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading, isError, error } =
    useCompanyList(params);

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

  const companies = data?.pages.flatMap((page) => page.data) ?? [];

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Company Directory</h1>
        <p className="mt-2 text-gray-600">
          Browse {companies.length > 0 ? `${companies.length}+` : ''} Indian tech companies.
          Filter by size, tech stack, compensation, and more.
        </p>
      </div>

      <div className="mb-6 space-y-4">
        <CompanySearchBar value={search} onChange={setSearch} />
        <CompanyFilters values={filters} onChange={setFilters} />
      </div>

      {isLoading && (
        <div className="py-12 text-center text-gray-500">Loading companies...</div>
      )}

      {isError && (
        <div className="rounded-lg bg-red-50 p-4 text-sm text-red-700">
          Failed to load companies: {(error as Error).message}
        </div>
      )}

      {!isLoading && companies.length === 0 && (
        <div className="py-12 text-center text-gray-500">
          No companies found. Try adjusting your search or filters.
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {companies.map((company) => (
          <CompanyCard key={company.id} company={company} />
        ))}
      </div>

      {/* Infinite scroll trigger */}
      <div ref={loadMoreRef} className="py-8 text-center">
        {isFetchingNextPage && (
          <span className="text-sm text-gray-500">Loading more...</span>
        )}
      </div>
    </div>
  );
}
