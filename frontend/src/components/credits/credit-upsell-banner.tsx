'use client';

import Link from 'next/link';
import { AlertCircle } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import {
  creditUpsellProducts,
  getCreditPurchaseHref,
  type CreditTypeKey,
} from '@/lib/payment-products';

export function CreditUpsellBanner({ creditType }: { creditType: CreditTypeKey }) {
  const isPremium = useAuthStore((s) => s.isPremium);
  const product = creditUpsellProducts[creditType];
  const href = getCreditPurchaseHref(isPremium);
  const creditLabels: Record<CreditTypeKey, string> = {
    resume_upload: 'resume upload',
    ats_check: 'ATS check',
    curated_list: 'curated list',
    cv_generation: 'cover letter',
  };

  return (
    <div className="mt-4 flex items-start gap-2 rounded-lg border border-[var(--color-warning)]/30 bg-[var(--color-warning)]/10 px-4 py-3 text-sm text-[var(--color-warning)]">
      <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
      {isPremium ? (
        <p>
          No {creditLabels[creditType]} credits remaining.{' '}
          {product.available === false ? (
            <span>{product.name} is coming soon.</span>
          ) : (
            <>
              <Link href={href} className="underline hover:text-[var(--color-warning)]/80">
                {product.name} (₹{product.price} for {product.quantity} {product.unitLabel})
              </Link>
              .
            </>
          )}
        </p>
      ) : (
        <p>
          No {creditLabels[creditType]} credits remaining.{' '}
          <Link href={href} className="underline hover:text-[var(--color-warning)]/80">
            Buy the Starter Pack
          </Link>{' '}
          to unlock premium credits.
        </p>
      )}
    </div>
  );
}
