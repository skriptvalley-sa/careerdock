'use client';

import { useQuery, useInfiniteQuery } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type {
  CompanyListItem,
  CompanyDetail,
  CompanyFilterParams,
  PaginatedResponse,
} from '@/types/api';

/** Fetch a paginated, filtered list of companies with infinite scroll support. */
export function useCompanyList(params: CompanyFilterParams = {}) {
  // Build clean params (omit empty strings)
  const cleanParams: Record<string, string> = {};
  for (const [key, value] of Object.entries(params)) {
    if (value) cleanParams[key] = value;
  }

  return useInfiniteQuery({
    queryKey: queryKeys.companies.list(cleanParams),
    queryFn: async ({ pageParam }) => {
      const fetchParams = { ...cleanParams };
      if (pageParam) fetchParams.cursor = pageParam as string;
      return apiClient.getPaginated<CompanyListItem>(
        '/api/companies',
        fetchParams,
      );
    },
    initialPageParam: '' as string,
    getNextPageParam: (lastPage: PaginatedResponse<CompanyListItem>) =>
      lastPage.pagination.has_more ? lastPage.pagination.next_cursor : undefined,
    staleTime: staleTimes.companyList,
  });
}

/** Fetch a single company by slug. */
export function useCompanyDetail(slug: string) {
  return useQuery({
    queryKey: queryKeys.companies.detail(slug),
    queryFn: () => apiClient.get<CompanyDetail>(`/api/companies/${slug}`),
    staleTime: staleTimes.companyDetail,
    enabled: !!slug,
  });
}
