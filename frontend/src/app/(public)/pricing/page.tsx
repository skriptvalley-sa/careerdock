'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Check, Loader2, ShoppingBag } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { useProductCheckout } from '@/hooks/use-product-checkout';
import {
  SUPPORT_EMAIL,
  starterPackProduct,
  premiumShopProducts,
  formatCreditLineItems,
} from '@/lib/payment-products';

const freeTier = {
  name: 'Free',
  priceLabel: 'Free',
  description: 'Get started with the essentials',
  highlight: false,
  features: [
    'Full company directory access',
    'Up to 3 custom company lists',
    'Application tracking (status, dates, notes)',
    'Basic search and filters',
  ],
};

const starterPack = {
  name: starterPackProduct.name,
  priceLabel: `₹${starterPackProduct.price}`,
  description: starterPackProduct.description,
  highlight: true,
  features: [
    ...formatCreditLineItems(starterPackProduct.credits).map((item) => `${item} included`),
    'Premium dashboard and AI tools',
    'Manage up to 3 active resumes at a time',
    'One-time purchase, no subscription',
  ],
};

function PricingCard({
  plan,
  ctaLabel,
  ctaHref,
  onClick,
  disabled,
  loading,
}: {
  plan: typeof freeTier | typeof starterPack;
  ctaLabel: string;
  ctaHref?: string;
  onClick?: () => void;
  disabled?: boolean;
  loading?: boolean;
}) {
  return (
    <div
      className={`flex flex-col rounded-xl border p-8 ${
        plan.highlight
          ? 'border-[var(--color-primary)]/40 shadow-lg shadow-[var(--color-primary)]/5 ring-1 ring-[var(--color-primary)]/30 glow-primary'
          : 'border-edge card-hover'
      }`}
    >
      {plan.highlight && (
        <span className="-mt-12 mb-4 self-start rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/15 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]">
          Most Popular
        </span>
      )}

      <h3 className="text-lg font-semibold text-[var(--color-text)]">{plan.name}</h3>
      <p className="mt-1 text-sm text-[var(--color-text-muted)]">{plan.description}</p>

      <div className="mt-6">
        <span className="text-4xl font-bold text-[var(--color-text)]">{plan.priceLabel}</span>
        {plan.priceLabel !== 'Free' && (
          <span className="text-sm text-[var(--color-text-muted)]"> one-time</span>
        )}
      </div>

      <ul className="mt-8 flex-1 space-y-3">
        {plan.features.map((feature) => (
          <li key={feature} className="flex items-start gap-2 text-sm text-[var(--color-text)]">
            <Check className="mt-0.5 h-4 w-4 shrink-0 text-[var(--color-primary)]" />
            {feature}
          </li>
        ))}
      </ul>

      {ctaHref ? (
        <Link
          href={ctaHref}
          className={`mt-8 block rounded-lg px-4 py-2.5 text-center text-sm font-semibold ${
            plan.highlight
              ? 'btn-primary'
              : 'border border-[var(--color-primary)]/20 text-[var(--color-primary)] transition-all hover:border-[var(--color-primary)]/40 hover:bg-[var(--color-primary)]/5'
          }`}
        >
          {ctaLabel}
        </Link>
      ) : (
        <button
          onClick={onClick}
          disabled={disabled || loading}
          className={`mt-8 flex items-center justify-center gap-2 rounded-lg px-4 py-2.5 text-sm font-semibold ${
            plan.highlight
              ? 'btn-primary disabled:cursor-not-allowed disabled:opacity-50'
              : 'border border-edge text-[var(--color-text)] disabled:opacity-50'
          }`}
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Processing...
            </>
          ) : (
            <>
              <ShoppingBag className="h-4 w-4" />
              {ctaLabel}
            </>
          )}
        </button>
      )}
    </div>
  );
}

export default function PricingPage() {
  const { isAuthenticated, isPremium } = useAuthStore();
  const { checkoutProduct, isPending } = useProductCheckout();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleStarterCheckout = async () => {
    setError(null);
    setSuccess(null);

    try {
      await checkoutProduct('starter_pack');
      setSuccess('Payment successful. Your premium credits have been added.');
    } catch (checkoutError) {
      const message =
        checkoutError instanceof Error
          ? checkoutError.message
          : 'Failed to create order. Please try again.';
      setError(message);
    }
  };

  const starterCTA = (() => {
    if (!isAuthenticated) {
      return { label: 'Sign up to buy', href: '/register' };
    }
    if (isPremium) {
      return { label: 'Open Credit Shop', href: '/shop' };
    }
    return { label: 'Buy Starter Pack', onClick: handleStarterCheckout };
  })() as { label: string; href?: string; onClick?: () => void };

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold text-[var(--color-text)]">
          Simple, transparent pricing
        </h1>
        <p className="mt-4 text-lg text-[var(--color-text-muted)]">
          Start free. Upgrade once. Refill credits only when you need them.
        </p>
      </div>

      {error && (
        <div className="mt-6 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
          {error}
        </div>
      )}
      {success && (
        <div className="mt-6 rounded-lg border border-[#39ff14]/30 bg-[#39ff14]/10 px-4 py-3 text-sm text-[#39ff14]">
          {success}
        </div>
      )}

      <div className="mt-12 grid gap-8 sm:grid-cols-2">
        <PricingCard
          plan={freeTier}
          ctaLabel={isAuthenticated ? 'Current plan' : 'Create Free Account'}
          ctaHref={isAuthenticated ? undefined : '/register'}
          disabled={isAuthenticated}
        />
        <PricingCard
          plan={starterPack}
          ctaLabel={starterCTA.label}
          ctaHref={starterCTA.href}
          onClick={starterCTA.onClick}
          disabled={!starterCTA.onClick}
          loading={isPending}
        />
      </div>

      <section className="mt-16 rounded-2xl border border-edge bg-card/60 p-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 className="text-xl font-bold text-[var(--color-text)]">Credit Shop</h2>
            <p className="mt-2 text-sm text-[var(--color-text-muted)]">
              Premium users can buy refills and bundles without revisiting the Starter Pack.
            </p>
          </div>
          {isPremium && (
            <Link href="/shop" className="btn-primary inline-flex items-center justify-center rounded-lg px-4 py-2.5 text-sm font-semibold">
              Open Credit Shop
            </Link>
          )}
        </div>

        <div className="mt-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {premiumShopProducts.map((product) => (
            <div key={product.productType} className="rounded-lg border border-edge bg-overlay px-4 py-4">
              <p className="text-sm font-semibold text-[var(--color-text)]">{product.name}</p>
              <p className="mt-1 text-sm text-[var(--color-primary)]">
                {product.available ? `₹${product.price}` : product.statusLabel ?? 'Coming soon'}
              </p>
              <p className="mt-1 text-xs text-[var(--color-text-muted)]">{product.description}</p>
            </div>
          ))}
        </div>
      </section>

      <section className="mt-16 text-center">
        <h2 className="text-xl font-bold text-[var(--color-text)]">Questions?</h2>
        <p className="mt-2 text-sm text-[var(--color-text-muted)]">
          Reach out at{' '}
          <a
            href={`mailto:${SUPPORT_EMAIL}`}
            className="font-medium text-[var(--color-text)] underline decoration-[var(--color-primary)]/40 underline-offset-4 transition-colors hover:text-[var(--color-primary)]"
          >
            {SUPPORT_EMAIL}
          </a>{' '}
          and we&apos;ll get back to you within 24 hours.
        </p>
      </section>
    </div>
  );
}
