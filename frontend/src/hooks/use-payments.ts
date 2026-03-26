'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import type {
  CreditBalances,
  CreditTransaction,
  PaymentOrder,
  PaymentRecord,
} from '@/types/api';

// --- Queries ---

export function useCreditBalance() {
  return useQuery({
    queryKey: queryKeys.credits.balance,
    queryFn: () => apiClient.get<CreditBalances>('/api/credits'),
    staleTime: staleTimes.credits,
    refetchInterval: 60_000, // poll every minute as a fallback to SSE
  });
}

export function useCreditTransactions() {
  return useQuery({
    queryKey: ['credits', 'transactions'] as const,
    queryFn: () =>
      apiClient.get<CreditTransaction[]>('/api/credits/transactions'),
    staleTime: staleTimes.credits,
  });
}

export function usePaymentHistory() {
  return useQuery({
    queryKey: ['payments', 'list'] as const,
    queryFn: () => apiClient.get<PaymentRecord[]>('/api/payments'),
    staleTime: 60_000,
  });
}

// --- Mutations ---

export function useCreateOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (productType: string) =>
      apiClient.post<PaymentOrder>('/api/payments/orders', {
        product_type: productType,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['payments'] });
    },
  });
}

export function useConfirmPayment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      razorpay_payment_id: string;
      razorpay_order_id: string;
      razorpay_signature: string;
    }) => apiClient.post<void>('/api/payments/confirm', data),
    onSuccess: () => {
      // Refresh credits & payments after successful payment
      qc.invalidateQueries({ queryKey: queryKeys.credits.balance });
      qc.invalidateQueries({ queryKey: ['credits'] });
      qc.invalidateQueries({ queryKey: ['payments'] });
      // Refresh auth to pick up premium_since
      qc.invalidateQueries({ queryKey: queryKeys.auth.me });
    },
  });
}
