'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type {
  UserList,
  ListDetail,
  ListEntry,
  ListCompanyFlag,
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
    }) => apiClient.post<ListEntry>(`/api/lists/${listId}/entries`, data),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list() });
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
      position?: number;
    }) =>
      apiClient.put<ListEntry>(
        `/api/lists/${listId}/entries/${entryId}`,
        data,
      ),
    onSuccess: (_, vars) => {
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(vars.listId) });
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
