'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys } from '@/lib/query-keys';
import type {
  AdminUser,
  AdminListResponse,
  AdminPayment,
  AdminCreditTransaction,
  CompanyDetail,
} from '@/types/api';

// Admin list endpoints return {data: T[], total: N} — not the standard {data: T} envelope.
// Use apiClient.getRaw which reuses the same 401 auto-refresh logic as all other calls.

// --- Admin Users ---

export interface AdminUserFilter {
  q?: string;
  role?: string;
  limit?: string;
  offset?: string;
}

export function useAdminUsers(filter: AdminUserFilter = {}) {
  const params: Record<string, string> = {};
  if (filter.q) params.q = filter.q;
  if (filter.role) params.role = filter.role;
  if (filter.limit) params.limit = filter.limit;
  if (filter.offset) params.offset = filter.offset;

  return useQuery({
    queryKey: queryKeys.admin.users(params),
    queryFn: () =>
      apiClient.getRaw<AdminListResponse<AdminUser>>('/api/admin/users', params),
    staleTime: 15_000,
  });
}

export function useAdminUpdateUser() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      ...data
    }: {
      userId: string;
      role?: string;
      set_premium?: boolean;
      banned?: boolean;
    }) => apiClient.put<AdminUser>(`/api/admin/users/${userId}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.admin.all });
    },
  });
}

export function useAdminAllocateCredits() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      userId,
      ...data
    }: {
      userId: string;
      credit_type: string;
      amount: number;
      reason: string;
    }) => apiClient.post(`/api/admin/users/${userId}/credits`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.admin.all });
    },
  });
}

// --- Admin Companies ---

export function useAdminCreateCompany() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Record<string, unknown>) =>
      apiClient.post<CompanyDetail>('/api/admin/companies', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.companies.all });
    },
  });
}

export function useAdminUpdateCompany() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      companyId,
      ...data
    }: Record<string, unknown> & { companyId: string }) =>
      apiClient.put<CompanyDetail>(
        `/api/admin/companies/${companyId}`,
        data,
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.companies.all });
    },
  });
}

export function useAdminUploadLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      companyId,
      formData,
    }: {
      companyId: string;
      formData: FormData;
    }) => apiClient.upload<{ key: string }>(`/api/admin/companies/${companyId}/logo`, formData),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.companies.all });
    },
  });
}

// --- Admin Payments ---

export interface AdminPaymentFilter {
  user_id?: string;
  status?: string;
  limit?: string;
  offset?: string;
}

export function useAdminPayments(filter: AdminPaymentFilter = {}) {
  const params: Record<string, string> = {};
  if (filter.user_id) params.user_id = filter.user_id;
  if (filter.status) params.status = filter.status;
  if (filter.limit) params.limit = filter.limit;
  if (filter.offset) params.offset = filter.offset;

  return useQuery({
    queryKey: queryKeys.admin.payments(params),
    queryFn: () =>
      apiClient.getRaw<AdminListResponse<AdminPayment>>('/api/admin/payments', params),
    staleTime: 15_000,
  });
}

// --- Admin Credit Transactions ---

export interface AdminCreditTxnFilter {
  user_id?: string;
  credit_type?: string;
  limit?: string;
  offset?: string;
}

export function useAdminCreditTransactions(
  filter: AdminCreditTxnFilter = {},
) {
  const params: Record<string, string> = {};
  if (filter.user_id) params.user_id = filter.user_id;
  if (filter.credit_type) params.credit_type = filter.credit_type;
  if (filter.limit) params.limit = filter.limit;
  if (filter.offset) params.offset = filter.offset;

  return useQuery({
    queryKey: queryKeys.admin.creditTransactions(params),
    queryFn: () =>
      apiClient.getRaw<AdminListResponse<AdminCreditTransaction>>(
        '/api/admin/credits/transactions',
        params,
      ),
    staleTime: 15_000,
  });
}
