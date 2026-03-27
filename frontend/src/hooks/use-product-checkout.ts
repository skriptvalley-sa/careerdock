'use client';

import { useCallback } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { useCreateOrder, useConfirmPayment } from '@/hooks/use-payments';
import { openRazorpayCheckout, type RazorpayPaymentResponse } from '@/lib/razorpay';
import { productCatalog, type ProductType } from '@/lib/payment-products';

export function useProductCheckout() {
  const user = useAuthStore((s) => s.user);
  const createOrder = useCreateOrder();
  const confirmPayment = useConfirmPayment();

  const openCheckout = useCallback(
    async (order: {
      razorpay_key_id: string;
      amount_paise: number;
      currency: string;
      razorpay_order_id: string;
    }, description: string) => {
      await new Promise<void>((resolve, reject) => {
        void openRazorpayCheckout({
          key: order.razorpay_key_id,
          amount: order.amount_paise,
          currency: order.currency || 'INR',
          name: 'CareerDock',
          description,
          order_id: order.razorpay_order_id,
          prefill: {
            name: user?.name,
            email: user?.email,
          },
          theme: {
            color: '#FD802E',
          },
          handler: async (response: RazorpayPaymentResponse) => {
            try {
              await confirmPayment.mutateAsync({
                razorpay_payment_id: response.razorpay_payment_id,
                razorpay_order_id: response.razorpay_order_id,
                razorpay_signature: response.razorpay_signature,
              });
              resolve();
            } catch {
              reject(new Error('Payment was received but confirmation failed.'));
            }
          },
          modal: {
            ondismiss: () => {
              reject(new Error('Checkout cancelled.'));
            },
          },
        }).catch(() => {
          reject(new Error('Failed to open checkout.'));
        });
      });
    },
    [confirmPayment, user?.email, user?.name],
  );

  const checkoutProduct = useCallback(
    async (productType: ProductType) => {
      const product = productCatalog[productType];
      const order = await createOrder.mutateAsync({ productType });
      await openCheckout(order, product.name);
    },
    [createOrder, openCheckout],
  );

  const checkoutCart = useCallback(
    async (items: Array<{ productType: ProductType; quantity: number }>) => {
      const order = await createOrder.mutateAsync({ items });
      await openCheckout(order, 'Credit Shop Cart');
    },
    [createOrder, openCheckout],
  );

  return {
    checkoutProduct,
    checkoutCart,
    isPending: createOrder.isPending || confirmPayment.isPending,
  };
}
