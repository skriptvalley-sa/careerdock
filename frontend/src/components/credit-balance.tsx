'use client';

import { Coins } from 'lucide-react';
import { useCreditBalance } from '@/hooks/use-payments';
import { useAuthStore } from '@/store/auth-store';

/**
 * Compact credit balance display for sidebar. Shows total credits
 * across all types. Only renders for premium users.
 */
export function CreditBalance({ collapsed }: { collapsed?: boolean }) {
  const { isPremium } = useAuthStore();
  const { data: credits } = useCreditBalance();

  if (!isPremium || !credits) return null;

  const total =
    credits.resume_upload +
    credits.ats_check +
    credits.curated_list +
    credits.cv_generation;

  if (collapsed) {
    return (
      <div
        className="flex items-center justify-center rounded-md px-2 py-1.5"
        title={`${total} credits`}
      >
        <div className="relative">
          <Coins className="h-4 w-4 text-[var(--color-warning)]" />
          <span className="absolute -right-1.5 -top-1.5 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-[var(--color-warning)] text-[8px] font-bold text-black">
            {total > 99 ? '99' : total}
          </span>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-md border border-edge bg-card/50 px-3 py-2">
      <div className="flex items-center gap-2">
        <Coins className="h-4 w-4 text-[var(--color-warning)]" />
        <span className="text-xs font-medium text-[var(--color-text-muted)]">Credits</span>
        <span className="ml-auto text-sm font-bold text-[var(--color-warning)]">{total}</span>
      </div>
      <div className="mt-2 space-y-1 text-[10px] text-[var(--color-text-muted)]">
        <div className="flex items-center justify-between gap-3">
          <span>Resume</span>
          <span className="font-medium text-[var(--color-text)]">{credits.resume_upload}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <span>ATS</span>
          <span className="font-medium text-[var(--color-text)]">{credits.ats_check}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <span>Lists</span>
          <span className="font-medium text-[var(--color-text)]">{credits.curated_list}</span>
        </div>
        <div className="flex items-center justify-between gap-3">
          <span>Cover letters</span>
          <span className="font-medium text-[var(--color-text)]">{credits.cv_generation}</span>
        </div>
      </div>
    </div>
  );
}
