'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type { ATSCheck, ATSResult } from '@/types/api';

export function isATSComplete(result: ATSCheck['result']): result is ATSResult {
  return 'score' in result;
}

// --- Queries ---

export function useATSChecks() {
  return useQuery({
    queryKey: queryKeys.ats.list(),
    queryFn: () => apiClient.get<ATSCheck[]>('/api/ats/'),
    staleTime: staleTimes.atsResults,
  });
}

export function useATSCheck(id: string) {
  return useQuery({
    queryKey: queryKeys.ats.detail(id),
    queryFn: () => apiClient.get<ATSCheck>(`/api/ats/${id}`),
    enabled: !!id,
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
  return useMutation({
    mutationFn: ({ resumeId, companyId }: { resumeId: string; companyId: string }) =>
      apiClient.post<ATSCheck>('/api/ats/company', {
        resume_id: resumeId,
        company_id: companyId,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.ats.all });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
    },
  });
}

export function useCheckJob() {
  const qc = useQueryClient();
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
      qc.invalidateQueries({ queryKey: queryKeys.ats.all });
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
    },
  });
}
