'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type {
  UserList,
  ListDetail,
  ListEntry,
  ListCompanyFlag,
  CompanyTrackingStatus,
} from '@/types/api';

// Helper: patch a single entry in a cached ListDetail without a full refetch.
function patchEntry(old: ListDetail | undefined, updated: ListEntry): ListDetail | undefined {
  if (!old) return old;
  return { ...old, entries: old.entries.map((e) => (e.id === updated.id ? updated : e)) };
}

// --- Queries ---

export function useLists() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.lists.list(userId!),
    queryFn: () => apiClient.get<UserList[]>('/api/lists'),
    staleTime: staleTimes.userLists,
    enabled: !!userId,
  });
}

export function useListDetail(id: string) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.lists.detail(userId!, id),
    queryFn: () => apiClient.get<ListDetail>(`/api/lists/${id}`),
    staleTime: staleTimes.userLists,
    enabled: !!id && !!userId,
  });
}

// --- Mutations ---

export function useCreateList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (data: { name: string; description?: string }) =>
      apiClient.post<UserList>('/api/lists', data),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
    },
  });
}

export function useUpdateList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
    },
  });
}

export function useDeleteList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/api/lists/${id}`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

export function useCreateEntry() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(userId, vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list(userId) });
    },
  });
}

export function useBatchCreateEntries() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(userId, vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

export function useUpdateEntry() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
    onSuccess: (updatedEntry, vars) => {
      if (!userId) return;
      // Immediately patch the current list's cache so the UI updates without
      // waiting for a round-trip refetch.
      qc.setQueryData<ListDetail>(
        queryKeys.lists.detail(userId, vars.listId),
        (old) => patchEntry(old, updatedEntry),
      );
      // Company status is synced across ALL lists for this user+company, so
      // invalidate every list detail to reflect the change everywhere.
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
    },
  });
}

export function useDeleteEntry() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: ({ listId, entryId }: { listId: string; entryId: string }) =>
      apiClient.delete(`/api/lists/${listId}/entries/${entryId}`),
    onSuccess: (_, vars) => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(userId, vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

export function useSyncListEntries() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.detail(userId, vars.listId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.list(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.companyEntries.all(userId) });
    },
  });
}

export function useListsForCompany(companyId: string | undefined) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.lists.byCompany(userId!, companyId!),
    queryFn: () =>
      apiClient.get<ListCompanyFlag[]>(`/api/lists/by-company/${companyId!}`),
    enabled: !!companyId && !!userId,
    staleTime: staleTimes.userLists,
  });
}

export function useCompanyListCounts() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.lists.companyCounts(userId!),
    queryFn: () =>
      apiClient.get<Record<string, number>>('/api/lists/company-counts'),
    staleTime: staleTimes.userLists,
    enabled: !!userId,
  });
}

export function useAddCompanyToList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.byCompanyAll(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.companyCounts(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

export function useRemoveCompanyFromList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.byCompanyAll(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.lists.companyCounts(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
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
