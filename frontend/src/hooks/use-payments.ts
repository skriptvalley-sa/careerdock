'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api';
import { queryKeys, staleTimes } from '@/lib/query-keys';
import { useAuthStore } from '@/store/auth-store';
import type { ProductType } from '@/lib/payment-products';
import type {
  CreditBalances,
  CreditTransaction,
  PaymentOrder,
  PaymentRecord,
} from '@/types/api';

// --- Queries ---

export function useCreditBalance() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.credits.balance(userId!),
    queryFn: () => apiClient.get<CreditBalances>('/api/credits'),
    staleTime: staleTimes.credits,
    refetchInterval: 60_000, // poll every minute as a fallback to SSE
    enabled: !!userId,
  });
}

export function useCreditTransactions() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.credits.transactions(userId!),
    queryFn: () => apiClient.get<CreditTransaction[]>('/api/credits/transactions'),
    staleTime: staleTimes.credits,
    enabled: !!userId,
  });
}

export function usePaymentHistory() {
  const userId = useAuthStore((s) => s.user?.id);
  return useQuery({
    queryKey: queryKeys.payments.list(userId!),
    queryFn: () => apiClient.get<PaymentRecord[]>('/api/payments'),
    staleTime: 60_000,
    enabled: !!userId,
  });
}

// --- Mutations ---

export interface CreateOrderPayload {
  productType?: ProductType;
  items?: Array<{
    productType: ProductType;
    quantity: number;
  }>;
}

export function useCreateOrder() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (payload: CreateOrderPayload) => {
      if (!userId) throw new Error('Authentication required');

      const body =
        payload.items && payload.items.length > 0
          ? {
              items: payload.items.map((item) => ({
                product_type: item.productType,
                quantity: item.quantity,
              })),
            }
          : {
              product_type: payload.productType,
            };

      return apiClient.post<PaymentOrder>('/api/payments/orders', {
        ...body,
      });
    },
    onSuccess: () => {
      if (!userId) return;
      qc.invalidateQueries({ queryKey: queryKeys.payments.all(userId) });
    },
  });
}

export function useConfirmPayment() {
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.user?.id);
  return useMutation({
    mutationFn: (data: {
      razorpay_payment_id: string;
      razorpay_order_id: string;
      razorpay_signature: string;
    }) => apiClient.post<void>('/api/payments/confirm', data),
    onSuccess: () => {
      if (!userId) return;
      // Refresh credits & payments after successful payment
      qc.invalidateQueries({ queryKey: queryKeys.credits.all(userId) });
      qc.invalidateQueries({ queryKey: queryKeys.payments.all(userId) });
      // Refresh auth to pick up premium_since
      qc.invalidateQueries({ queryKey: queryKeys.auth.me });
    },
  });
}
