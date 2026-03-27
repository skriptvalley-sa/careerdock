'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type {
  Application,
  ApplicationStatus,
  StatusHistoryItem,
  InterviewRound,
  DashboardCounts,
} from '@/types/api';

// --- Queries ---

export function useApplications(status?: string) {
  const userId = useAuthStore((s) => s.user?.id);
  const params: Record<string, string> = {};
  if (status) params.status = status;
  return useQuery({
    queryKey: queryKeys.applications.list(userId!, status),
    queryFn: () => apiClient.get<Application[]>('/api/applications', params),
    staleTime: staleTimes.userLists,
    enabled: !!userId,
  });
}

export function useApplicationsByCompany(companyId: string | undefined) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.applications.byCompany(userId!, companyId!),
    queryFn: () =>
      apiClient.get<Application[]>(`/api/applications/by-company/${companyId!}`),
    enabled: !!companyId && !!userId,
    staleTime: staleTimes.userLists,
  });
}

export function useApplicationHistory(applicationId: string) {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.applications.detail(userId!, applicationId),
    queryFn: () =>
      apiClient.get<StatusHistoryItem[]>(`/api/applications/${applicationId}`),
    enabled: !!applicationId && !!userId,
  });
}

export function useDashboardCounts() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.dashboard.counts(userId!),
    queryFn: () => apiClient.get<DashboardCounts>('/api/dashboard'),
    staleTime: staleTimes.userLists,
    enabled: !!userId,
  });
}

// --- Mutations ---

export function useCreateApplication() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (data: {
      company_id: string;
      role_title?: string;
      status?: ApplicationStatus;
      date_applied?: string;
      notes?: string;
    }) => apiClient.post<Application>('/api/applications', data),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.applications.all(userId) });
      // Invalidate list queries so application_count in list detail updates immediately
      qc.invalidateQueries({ queryKey: queryKeys.lists.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

export function useUpdateApplication() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
    onSuccess: (updatedApp) => {
      if (!userId) return;
      // Immediately patch every cached applications list so the status badge
      // updates without waiting for a background refetch.
      qc.setQueriesData<Application[]>(
        { queryKey: queryKeys.applications.all(userId) },
        (old) => old?.map((a) => (a.id === updatedApp.id ? updatedApp : a)),
      );
      qc.invalidateQueries({ queryKey: queryKeys.applications.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

export function useDeleteApplication() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (id: string) => apiClient.delete(`/api/applications/${id}`),
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.applications.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.dashboard.all(userId) });
    },
  });
}

// --- Interview Rounds ---

export function useCreateRound() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
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
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.applications.all(userId) });
    },
  });
}
