'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys } from '@/lib/query-keys';

// --- Types ---

export interface CompanyDraft {
  name: string;
  slug?: string;
  description?: string;
  size?: string;
  headquarters?: string;
  founded_year?: number;
  careers_page_url?: string;
  linkedin_url?: string;
  tech_stack?: string[];
  domains?: string[];
  hiring_status?: string;
  office_modes?: string[];
  compensation_tier?: string;
  has_rsu?: boolean;
  has_rsu_refresher?: boolean;
  compensation_bands?: unknown;
}

export interface EditLock {
  company_id: string;
  locked_by: string;
  locked_at: string;
  expires_at: string;
}

// --- Mutations ---

export function useGenerateCompanyDraft() {
  return useMutation({
    mutationFn: (params: { name: string; careers_url?: string; linkedin_url?: string }) =>
      apiClient.post<CompanyDraft>('/api/moderator/companies/generate', params),
  });
}

export function useSubmitCompanyDraft() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (draft: CompanyDraft) =>
      apiClient.post('/api/moderator/companies', draft),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.companies.all });
    },
  });
}

export function useEditLockStatus(companyId: string) {
  return useQuery({
    queryKey: queryKeys.moderator.editLock(companyId),
    queryFn: () => apiClient.get<EditLock | null>(`/api/moderator/companies/${companyId}/lock`),
    enabled: !!companyId,
    staleTime: 10_000,
  });
}

export function useAcquireEditLock() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (companyId: string) =>
      apiClient.post<EditLock>(`/api/moderator/companies/${companyId}/lock`, {}),
    onSuccess: (_, companyId) => {
      qc.invalidateQueries({ queryKey: queryKeys.moderator.editLock(companyId) });
    },
  });
}

export function useReleaseEditLock() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (companyId: string) =>
      apiClient.delete(`/api/moderator/companies/${companyId}/lock`),
    onSuccess: (_, companyId) => {
      qc.invalidateQueries({ queryKey: queryKeys.moderator.editLock(companyId) });
    },
  });
}

export function useSubmitCompanyEdit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ companyId, changes }: { companyId: string; changes: Record<string, unknown> }) =>
      apiClient.post(`/api/moderator/companies/${companyId}/edit`, changes),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.companies.all });
      qc.invalidateQueries({ queryKey: queryKeys.moderator.all });
    },
  });
}
