'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type { CuratedList, CuratedListResult } from '@/types/api';

export function isCuratedListComplete(
  result: CuratedList['result'],
): result is CuratedListResult {
  return 'companies' in result;
}

// --- Queries ---

export function useCuratedLists() {
  return useQuery({
    queryKey: queryKeys.curatedLists.list(),
    queryFn: () => apiClient.get<CuratedList[]>('/api/curated-lists/'),
    staleTime: staleTimes.curatedLists,
  });
}

export function useCuratedList(id: string) {
  return useQuery({
    queryKey: queryKeys.curatedLists.detail(id),
    queryFn: () => apiClient.get<CuratedList>(`/api/curated-lists/${id}`),
    enabled: !!id,
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
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.post<CuratedList>('/api/curated-lists/', { resume_id: resumeId }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.curatedLists.all });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
    },
  });
}

export function useRenameCuratedList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) =>
      apiClient.put(`/api/curated-lists/${id}`, { name }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.curatedLists.all });
    },
  });
}

export function useDeleteCuratedList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/api/curated-lists/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.curatedLists.all });
    },
  });
}
