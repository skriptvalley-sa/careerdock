'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type { ATSCheck, ATSResult } from '@/types/api';

export function isATSComplete(result: ATSCheck['result']): result is ATSResult {
  return 'score' in result;
}

// --- Queries ---

export function useATSChecks() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.ats.list(userId!),
    queryFn: () => apiClient.get<ATSCheck[]>('/api/ats/'),
    staleTime: staleTimes.atsResults,
    enabled: !!userId,
  });
}

export function useATSCheck(id: string) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.ats.detail(userId!, id),
    queryFn: () => apiClient.get<ATSCheck>(`/api/ats/${id}`),
    enabled: !!id && !!userId,
    staleTime: staleTimes.atsResults,
    // Poll every 5s while pending; SSE will also trigger invalidation
    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data || !isATSComplete(data.result)) return 5_000;
      return false;
    },
  });
}

// --- Mutations ---

export function useCheckCompany() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: ({ resumeId, companyId }: { resumeId: string; companyId: string }) =>
      apiClient.post<ATSCheck>('/api/ats/company', {
        resume_id: resumeId,
        company_id: companyId,
      }),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.ats.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance(userId) });
    },
  });
}

export function useCheckJob() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: ({
      resumeId,
      jobDescription,
    }: {
      resumeId: string;
      jobDescription: string;
    }) =>
      apiClient.post<ATSCheck>('/api/ats/job', {
        resume_id: resumeId,
        job_description: jobDescription,
      }),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.ats.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance(userId) });
    },
  });
}

export function useCheckResume() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: ({ resumeId }: { resumeId: string }) =>
      apiClient.post<ATSCheck>('/api/ats/resume', {
        resume_id: resumeId,
      }),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.ats.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance(userId) });
    },
  });
}

export function useCheckResumeTempUpload() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (file: File) => {
      const formData = new FormData();
      formData.append('file', file);
      return apiClient.upload<ATSCheck>('/api/ats/resume/upload', formData);
    },
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.ats.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance(userId) });
    },
  });
}
