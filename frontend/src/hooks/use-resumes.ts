'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type { ResumeListItem, ResumeDetail } from '@/types/api';

// --- Queries ---

export function useResumes() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.resumes.list(userId!),
    queryFn: () => apiClient.get<ResumeListItem[]>('/api/resumes'),
    staleTime: staleTimes.resumes,
    enabled: !!userId,
  });
}

export function useResumeDetail(id: string) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.resumes.detail(userId!, id),
    queryFn: () => apiClient.get<ResumeDetail>(`/api/resumes/${id}`),
    staleTime: staleTimes.resumes,
    enabled: !!id && !!userId,
  });
}

// --- Mutations ---

export function useUploadResume() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: ({ file, slotNumber }: { file: File; slotNumber: number }) => {
      if (!userId) throw new Error('Authentication required');
      const formData = new FormData();
      formData.append('file', file);
      formData.append('slot_number', String(slotNumber));
      return apiClient.upload<ResumeDetail>('/api/resumes', formData);
    },
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance(userId) });
    },
  });
}

export function useSetDefaultResume() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.put<void>(`/api/resumes/${resumeId}/default`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.auth.me });
    },
  });
}

export function useArchiveResume() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.delete<void>(`/api/resumes/${resumeId}`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all(userId) });
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
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (resumeId: string) =>
      apiClient.post<void>(`/api/resumes/${resumeId}/retry`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.resumes.all(userId) });
    },
  });
}
