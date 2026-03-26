'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type { ResumeListItem, ResumeDetail } from '@/types/api';

// --- Queries ---

export function useResumes() {
  return useQuery({
    queryKey: queryKeys.resumes.list(),
    queryFn: () => apiClient.get<ResumeListItem[]>('/api/resumes'),
    staleTime: staleTimes.resumes,
  });
}

export function useResumeDetail(id: string) {
  return useQuery({
    queryKey: queryKeys.resumes.detail(id),
    queryFn: () => apiClient.get<ResumeDetail>(`/api/resumes/${id}`),
    staleTime: staleTimes.resumes,
    enabled: !!id,
  });
}

// --- Mutations ---

export function useUploadResume() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ file, slotNumber }: { file: File; slotNumber: number }) => {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('slot_number', String(slotNumber));
      return apiClient.upload<ResumeDetail>('/api/resumes', formData);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
    },
  });
}

export function useSetDefaultResume() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.put<void>(`/api/resumes/${resumeId}/default`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all });
      qc.invalidateQueries({ queryKey: queryKeys.auth.me });
    },
  });
}

export function useArchiveResume() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.delete<void>(`/api/resumes/${resumeId}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all });
    },
  });
}

export function useResumeDownloadUrl() {
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.get<{ url: string }>(`/api/resumes/${resumeId}/download`),
  });
}

export function useRetryResume() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.post<void>(`/api/resumes/${resumeId}/retry`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all });
    },
  });
}
