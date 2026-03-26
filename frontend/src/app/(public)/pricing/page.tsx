'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Check, Loader2, ShoppingCart } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { useCreateOrder, useConfirmPayment } from '@/hooks/use-payments';
import { openRazorpayCheckout } from '@/lib/razorpay';
import type { RazorpayPaymentResponse } from '@/lib/razorpay';

const freeTier = {
  name: 'Free',
  price: '0',
  description: 'Get started with the essentials',
  features: [
    'Full company directory access',
    'Up to 3 custom company lists',
    'Application tracking (status, dates, notes)',
    'Basic search & filters',
  ],
  highlight: false,
};

const starterPack = {
  name: 'Starter Pack',
  price: '299',
  productType: 'starter_pack',
  description: 'One-time purchase — no subscription',
  features: [
    'Everything in Free',
    'Upload up to 3 resumes (PDF)',
    'AI resume analysis & skill extraction',
    'General ATS score',
    'Company-specific ATS score',
    'Job-specific ATS score',
    'AI-curated company matching',
  ],
  highlight: true,
};

interface CreditPack {
  name: string;
  credits: number;
  price: string;
  productType: string;
  description: string;
}

const creditPacks: CreditPack[] = [
  {
    name: 'Resume Upload',
    credits: 1,
    price: '49',
    productType: 'resume_upload',
    description: '1 resume upload credit',
  },
  {
    name: 'ATS Bundle',
    credits: 5,
    price: '99',
    productType: 'ats_bundle',
    description: '5 ATS check credits',
  },
  {
    name: 'Rebuy Pack',
    credits: 15,
    price: '399',
    productType: 'rebuy_pack',
    description: 'Full credit refill (premium only)',
  },
];

function PricingCard({
  plan,
  onBuy,
  buying,
  showBuyButton,
}: {
  plan: typeof freeTier | typeof starterPack;
  onBuy?: () => void;
  buying?: boolean;
  showBuyButton?: boolean;
}) {
  const { isAuthenticated, isPremium } = useAuthStore();

  const getCTA = () => {
    if ('productType' in plan) {
      // Starter pack
      if (!isAuthenticated) return { label: 'Sign up to buy', href: '/register' };
      if (isPremium) return { label: 'Already purchased', href: undefined, disabled: true };
      return { label: 'Buy Starter Pack', action: onBuy };
    }
    // Free tier
    if (isAuthenticated) return { label: 'Current plan', href: undefined, disabled: true };
    return { label: 'Create Free Account', href: '/register' };
  };

  const cta = getCTA();

  return (
    <div
      className={`flex flex-col rounded-xl border p-8 ${
        plan.highlight
          ? 'border-[var(--color-primary)]/40 shadow-lg shadow-[var(--color-primary)]/5 ring-1 ring-[var(--color-primary)]/30 glow-primary'
          : 'border-edge card-hover'
      }`}
    >
      {plan.highlight && (
        <span className="-mt-12 mb-4 self-start rounded-full bg-[var(--color-primary)]/15 border border-[var(--color-primary)]/30 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]">
          Most Popular
        </span>
      )}
      <h3 className="text-lg font-semibold text-[var(--color-text)]">{plan.name}</h3>
      <p className="mt-1 text-sm text-[var(--color-text-muted)]">{plan.description}</p>
      <div className="mt-6">
        <span className="text-4xl font-bold text-[var(--color-text)]">
          {plan.price === '0' ? 'Free' : `₹${plan.price}`}
        </span>
        {plan.price !== '0' && (
          <span className="text-sm text-[var(--color-text-muted)]"> one-time</span>
        )}
      </div>
      <ul className="mt-8 flex-1 space-y-3">
        {plan.features.map((f) => (
          <li key={f} className="flex items-start gap-2 text-sm text-[var(--color-text)]">
            <Check className="mt-0.5 h-4 w-4 shrink-0 text-[var(--color-primary)]" />
            {f}
          </li>
        ))}
      </ul>

      {'action' in (cta as Record<string, unknown>) && showBuyButton ? (
        <button
          onClick={(cta as { action?: () => void }).action}
          disabled={buying}
          className="mt-8 flex items-center justify-center gap-2 rounded-lg px-4 py-2.5 text-sm font-semibold btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {buying ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Processing...
            </>
          ) : (
            <>
              <ShoppingCart className="h-4 w-4" />
              {cta.label}
            </>
          )}
        </button>
      ) : cta.href ? (
        <Link
          href={cta.href}
          className={`mt-8 block rounded-lg px-4 py-2.5 text-center text-sm font-semibold ${
            plan.highlight
              ? 'btn-primary'
              : 'border border-[var(--color-primary)]/20 text-[var(--color-primary)] hover:bg-[var(--color-primary)]/5 hover:border-[var(--color-primary)]/40 transition-all'
          }`}
        >
          {cta.label}
        </Link>
      ) : (
        <span
          className={`mt-8 block rounded-lg px-4 py-2.5 text-center text-sm font-semibold opacity-50 ${
            plan.highlight
              ? 'border border-[var(--color-primary)]/30 text-[var(--color-primary)]'
              : 'border border-edge text-[var(--color-text-muted)]'
          }`}
        >
          {cta.label}
        </span>
      )}
    </div>
  );
}

export default function PricingPage() {
  const { isAuthenticated, isPremium, user } = useAuthStore();
  const createOrder = useCreateOrder();
  const confirmPayment = useConfirmPayment();
  const [buyingProduct, setBuyingProduct] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleBuy = async (productType: string) => {
    setError(null);
    setSuccess(null);
    setBuyingProduct(productType);

    try {
      const order = await createOrder.mutateAsync(productType);

      await openRazorpayCheckout({
        key: order.razorpay_key_id,
        amount: order.amount_paise,
        currency: order.currency || 'INR',
        name: 'CareerDock',
        description: productType.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase()),
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
            setSuccess('Payment successful! Your credits have been added.');
          } catch {
            setError('Payment was received but confirmation failed. Please contact support.');
          } finally {
            setBuyingProduct(null);
          }
        },
        modal: {
          ondismiss: () => {
            setBuyingProduct(null);
          },
        },
      });
    } catch {
      setError('Failed to create order. Please try again.');
      setBuyingProduct(null);
    }
  };

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="text-center">
        <h1 className="text-3xl font-bold text-[var(--color-text)]">
          Simple, transparent pricing
        </h1>
        <p className="mt-4 text-lg text-[var(--color-text-muted)]">
          Start free. Pay once for AI features. No subscriptions.
        </p>
      </div>

      {/* Status messages */}
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

      {/* Plans */}
      <div className="mt-12 grid gap-8 sm:grid-cols-2">
        <PricingCard plan={freeTier} />
        <PricingCard
          plan={starterPack}
          onBuy={() => handleBuy('starter_pack')}
          buying={buyingProduct === 'starter_pack'}
          showBuyButton={isAuthenticated && !isPremium}
        />
      </div>

      {/* Credit Packs — only shown for premium users */}
      <section className="mt-16">
        <h2 className="text-center text-xl font-bold text-[var(--color-text)]">
          Need more credits?
        </h2>
        <p className="mt-2 text-center text-sm text-[var(--color-text-muted)]">
          Buy additional AI credits a la carte after your Starter Pack.
        </p>
        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          {creditPacks.map((pack) => (
            <div
              key={pack.productType}
              className="card-hover rounded-lg border border-edge p-6 text-center"
            >
              <div className="text-lg font-semibold text-[var(--color-text)]">
                {pack.name}
              </div>
              <div className="mt-1 text-2xl font-bold text-[var(--color-primary)]">
                ₹{pack.price}
              </div>
              <p className="mt-1 text-xs text-[var(--color-text-muted)]">{pack.description}</p>

              {isAuthenticated && isPremium ? (
                <button
                  onClick={() => handleBuy(pack.productType)}
                  disabled={buyingProduct === pack.productType}
                  className="mt-4 inline-flex items-center gap-1.5 rounded-md border border-[var(--color-primary)]/20 px-3 py-1.5 text-xs font-medium text-[var(--color-primary)] hover:bg-[var(--color-primary)]/5 hover:border-[var(--color-primary)]/40 transition-all disabled:opacity-50"
                >
                  {buyingProduct === pack.productType ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <ShoppingCart className="h-3 w-3" />
                  )}
                  Buy
                </button>
              ) : isAuthenticated && !isPremium ? (
                <p className="mt-4 text-xs text-[var(--color-text-muted)]">
                  Get the Starter Pack first
                </p>
              ) : (
                <Link
                  href="/register"
                  className="mt-4 inline-block text-xs text-[var(--color-primary)] hover:underline"
                >
                  Sign up
                </Link>
              )}
            </div>
          ))}
        </div>
      </section>

      {/* FAQ hint */}
      <section className="mt-16 text-center">
        <h2 className="text-xl font-bold text-[var(--color-text)]">Questions?</h2>
        <p className="mt-2 text-sm text-[var(--color-text-muted)]">
          Reach out at{' '}
          <span className="font-medium text-[var(--color-text)]">
            support@careerdock.in
          </span>{' '}
          and we&apos;ll get back to you within 24 hours.
        </p>
      </section>
    </div>
  );
}
