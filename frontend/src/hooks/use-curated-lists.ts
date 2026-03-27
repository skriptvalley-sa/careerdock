'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type { CuratedList, CuratedListResult } from '@/types/api';

export function isCuratedListComplete(
  result: CuratedList['result'],
): result is CuratedListResult {
  return 'companies' in result;
}

// --- Queries ---

export function useCuratedLists() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.curatedLists.list(userId!),
    queryFn: () => apiClient.get<CuratedList[]>('/api/curated-lists/'),
    staleTime: staleTimes.curatedLists,
    enabled: !!userId,
  });
}

export function useCuratedList(id: string) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.curatedLists.detail(userId!, id),
    queryFn: () => apiClient.get<CuratedList>(`/api/curated-lists/${id}`),
    enabled: !!id && !!userId,
    staleTime: staleTimes.curatedLists,
    // Poll every 8s while pending; SSE will also trigger invalidation
    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data || !isCuratedListComplete(data.result)) return 8_000;
      return false;
    },
  });
}

// --- Mutations ---

export function useGenerateCuratedList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.post<CuratedList>('/api/curated-lists/', { resume_id: resumeId }),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.curatedLists.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance(userId) });
    },
  });
}

export function useRenameCuratedList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      apiClient.put(`/api/curated-lists/${id}`, { name }),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.curatedLists.all(userId) });
    },
  });
}

export function useDeleteCuratedList() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/api/curated-lists/${id}`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.curatedLists.all(userId) });
    },
  });
}
