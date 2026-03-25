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

// We use apiRaw-style fetch for admin list endpoints that return {data: [], total: N}
// instead of the standard {data: T} envelope. We use apiClient.get which unwraps .data,
// but admin list endpoints return {data: T[], total: N} at top level without an envelope.
// So we need to fetch raw. We'll use a small helper.

async function adminGet<T>(
  path: string,
  params?: Record<string, string>,
): Promise<T> {
  const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  let url = `${API_BASE}${path}`;
  if (params) {
    const cleaned: Record<string, string> = {};
    for (const [k, v] of Object.entries(params)) {
      if (v) cleaned[k] = v;
    }
    if (Object.keys(cleaned).length > 0) {
      url += `?${new URLSearchParams(cleaned).toString()}`;
    }
  }
  const resp = await fetch(url, { credentials: 'include' });
  if (!resp.ok) {
    const json = await resp.json();
    throw new Error(json.error?.message || 'Request failed');
  }
  return resp.json() as Promise<T>;
}

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
      adminGet<AdminListResponse<AdminUser>>('/api/admin/users', params),
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
      adminGet<AdminListResponse<AdminPayment>>('/api/admin/payments', params),
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
      adminGet<AdminListResponse<AdminCreditTransaction>>(
        '/api/admin/credits/transactions',
        params,
      ),
    staleTime: 15_000,
  });
}
