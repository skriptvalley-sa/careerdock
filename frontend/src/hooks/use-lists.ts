'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type {
  UserList,
  ListDetail,
  ListEntry,
  ListCompanyFlag,
  StatusHistoryItem,
  InterviewRound,
  DashboardCounts,
  ApplicationStatus,
  CompanyTrackingStatus,
} from '@/types/api';

// --- Queries ---

export function useLists() {
  return useQuery({
    queryKey: queryKeys.lists.list(),
    queryFn: () => apiClient.get<UserList[]>('/api/lists'),
    staleTime: staleTimes.userLists,
  });
}

export function useListDetail(id: string) {
  return useQuery({
    queryKey: queryKeys.lists.detail(id),
    queryFn: () => apiClient.get<ListDetail>(`/api/lists/${id}`),
    staleTime: staleTimes.userLists,
    enabled: !!id,
  });
}

export function useEntryHistory(listId: string, entryId: string) {
  return useQuery({
    queryKey: ['lists', 'history', entryId] as const,
    queryFn: () =>
      apiClient.get<StatusHistoryItem[]>(
        `/api/lists/${listId}/entries/${entryId}/history`,
      ),
    enabled: !!entryId,
  });
}

export function useDashboardCounts() {
  return useQuery({
    queryKey: ['dashboard', 'counts'] as const,
    queryFn: () => apiClient.get<DashboardCounts>('/api/dashboard'),
    staleTime: staleTimes.userLists,
  });
}

// --- Mutations ---

export function useCreateList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; description?: string }) =>
      apiClient.post<UserList>('/api/lists', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.all });
    },
  });
}

export function useUpdateList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      ...data
    }: {
      id: string;
      name?: string;
      description?: string;
      position?: number;
    }) => apiClient.put<UserList>(`/api/lists/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.all });
    },
  });
}

export function useDeleteList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/api/lists/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.all });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useCreateEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      ...data
    }: {
      listId: string;
      company_id: string;
      company_status?: CompanyTrackingStatus;
      role_title?: string;
      status?: ApplicationStatus;
      date_applied?: string;
      notes?: string;
    }) => apiClient.post<ListEntry>(`/api/lists/${listId}/entries`, data),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list() });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useBatchCreateEntries() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      company_ids,
    }: {
      listId: string;
      company_ids: string[];
    }) =>
      apiClient.post<ListEntry[]>(
        `/api/lists/${listId}/entries/batch`,
        { company_ids },
      ),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list() });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useUpdateEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      entryId,
      ...data
    }: {
      listId: string;
      entryId: string;
      company_status?: CompanyTrackingStatus;
      status?: ApplicationStatus;
      role_title?: string;
      notes?: string;
      date_applied?: string;
    }) =>
      apiClient.put<ListEntry>(
        `/api/lists/${listId}/entries/${entryId}`,
        data,
      ),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useDeleteEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ listId, entryId }: { listId: string; entryId: string }) =>
      apiClient.delete(`/api/lists/${listId}/entries/${entryId}`),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list() });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useSyncListEntries() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      company_ids,
    }: {
      listId: string;
      company_ids: string[];
    }) =>
      apiClient.put<ListEntry[]>(
        `/api/lists/${listId}/entries/sync`,
        { company_ids },
      ),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list() });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
      qc.invalidateQueries({ queryKey: ['entries'] });
    },
  });
}

export function useListsForCompany(companyId: string | undefined) {
  return useQuery({
    queryKey: ['lists', 'by-company', companyId] as const,
    queryFn: () =>
      apiClient.get<ListCompanyFlag[]>(`/api/lists/by-company/${companyId!}`),
    enabled: !!companyId,
    staleTime: staleTimes.userLists,
  });
}

export function useCompanyListCounts() {
  return useQuery({
    queryKey: ['lists', 'company-counts'] as const,
    queryFn: () =>
      apiClient.get<Record<string, number>>('/api/lists/company-counts'),
    staleTime: staleTimes.userLists,
  });
}

export function useAddCompanyToList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      companyId,
    }: {
      listId: string;
      companyId: string;
    }) =>
      apiClient.post<ListEntry>(`/api/lists/${listId}/entries`, {
        company_id: companyId,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.all });
      qc.invalidateQueries({ queryKey: ['lists', 'by-company'] });
      qc.invalidateQueries({ queryKey: ['lists', 'company-counts'] });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useRemoveCompanyFromList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      companyId,
    }: {
      listId: string;
      companyId: string;
    }) =>
      apiClient.delete(`/api/lists/${listId}/entries/by-company/${companyId}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.all });
      qc.invalidateQueries({ queryKey: ['lists', 'by-company'] });
      qc.invalidateQueries({ queryKey: ['lists', 'company-counts'] });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useCreateRound() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      listId,
      entryId,
      ...data
    }: {
      listId: string;
      entryId: string;
      round_number: number;
      round_type: string;
      scheduled_date?: string;
      outcome?: string;
      notes?: string;
    }) =>
      apiClient.post<InterviewRound>(
        `/api/lists/${listId}/entries/${entryId}/rounds`,
        data,
      ),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
    },
  });
}

// --- Cross-list entry lookup ---

/** Fetch entries for a specific company across all the user's lists. */
export function useEntriesByCompany(companyId: string | undefined) {
  return useQuery({
    queryKey: ['entries', 'by-company', companyId] as const,
    queryFn: () =>
      apiClient.get<
        (ListEntry & { list_name: string })[]
      >('/api/entries', { company_id: companyId! }),
    enabled: !!companyId,
    staleTime: staleTimes.userLists,
  });
}

/** Fetch all entries across all lists, optionally filtered by status. */
export function useAllEntries(status?: string) {
  const params: Record<string, string> = {};
  if (status) params.status = status;
  return useQuery({
    queryKey: ['entries', 'all', status ?? ''] as const,
    queryFn: () =>
      apiClient.get<
        (ListEntry & { list_name: string; company_name: string })[]
      >('/api/entries', params),
    staleTime: staleTimes.userLists,
  });
}

// --- User settings ---

export function useUpdateProfile() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      name?: string;
      current_title?: string;
      experience_level?: string;
      preferred_tech_stacks?: string[];
      target_domains?: string[];
      target_locations?: string[];
    }) => apiClient.put('/api/users/me', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.auth.me });
    },
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (data: { current_password: string; new_password: string }) =>
      apiClient.put('/api/users/me/password', data),
  });
}

export function useDeleteAccount() {
  return useMutation({
    mutationFn: (data: { password: string }) =>
      apiClient.delete('/api/users/me', data),
  });
}
