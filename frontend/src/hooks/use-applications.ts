'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type {
  Application,
  ApplicationStatus,
  StatusHistoryItem,
  InterviewRound,
  DashboardCounts,
} from '@/types/api';

// --- Queries ---

export function useApplications(status?: string) {
  const params: Record<string, string> = {};
  if (status) params.status = status;
  return useQuery({
    queryKey: queryKeys.applications.list(status),
    queryFn: () => apiClient.get<Application[]>('/api/applications', params),
    staleTime: staleTimes.userLists,
  });
}

export function useApplicationsByCompany(companyId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.applications.byCompany(companyId!),
    queryFn: () =>
      apiClient.get<Application[]>(`/api/applications/by-company/${companyId!}`),
    enabled: !!companyId,
    staleTime: staleTimes.userLists,
  });
}

export function useApplicationHistory(applicationId: string) {
  return useQuery({
    queryKey: queryKeys.applications.detail(applicationId),
    queryFn: () =>
      apiClient.get<StatusHistoryItem[]>(`/api/applications/${applicationId}`),
    enabled: !!applicationId,
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

export function useCreateApplication() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      company_id: string;
      role_title?: string;
      status?: ApplicationStatus;
      date_applied?: string;
      notes?: string;
    }) => apiClient.post<Application>('/api/applications', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.applications.all });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useUpdateApplication() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      ...data
    }: {
      id: string;
      role_title?: string;
      status?: ApplicationStatus;
      date_applied?: string;
      notes?: string;
    }) => apiClient.put<Application>(`/api/applications/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.applications.all });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

export function useDeleteApplication() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/api/applications/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.applications.all });
      qc.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
}

// --- Interview Rounds ---

export function useCreateRound() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      applicationId,
      ...data
    }: {
      applicationId: string;
      round_number: number;
      round_type: string;
      scheduled_date?: string;
      outcome?: string;
      notes?: string;
    }) =>
      apiClient.post<InterviewRound>(
        `/api/applications/${applicationId}/rounds`,
        data,
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.applications.all });
    },
  });
}
