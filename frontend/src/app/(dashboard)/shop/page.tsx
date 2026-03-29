'use client';

import { type ComponentType, useEffect, useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import {
  FileText,
  FilePenLine,
  Loader2,
  Package,
  ScanSearch,
  ShoppingBag,
} from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { useCreditBalance } from '@/hooks/use-payments';
import { useProductCheckout } from '@/hooks/use-product-checkout';
import {
  formatCreditLineItems,
  premiumShopProducts,
  SUPPORT_EMAIL,
  type PremiumProductType,
} from '@/lib/payment-products';

type CartItem = {
  productId: PremiumProductType;
  label: string;
  price: number;
  qty: number;
};

const MAX_QTY_PER_PRODUCT = 5;

const productIcons: Record<PremiumProductType, ComponentType<{ className?: string }>> = {
  starter_refill: Package,
  resume_bundle: FileText,
  ats_bundle: ScanSearch,
  curated_list_bundle: ShoppingBag,
  cv_bundle: FilePenLine,
};

function updateCartItem(
  cart: CartItem[],
  item: Omit<CartItem, 'qty'>,
  qty: number,
): CartItem[] {
  if (qty <= 0) {
    return cart.filter((cartItem) => cartItem.productId !== item.productId);
  }

  const nextQty = Math.min(qty, MAX_QTY_PER_PRODUCT);
  const existing = cart.find((cartItem) => cartItem.productId === item.productId);

  if (!existing) {
    return [...cart, { ...item, qty: nextQty }];
  }

  return cart.map((cartItem) =>
    cartItem.productId === item.productId ? { ...cartItem, qty: nextQty } : cartItem,
  );
}

function PackCard({
  product,
  quantity,
  disabled,
  onQuantityChange,
}: {
  product: (typeof premiumShopProducts)[number];
  quantity: number;
  disabled: boolean;
  onQuantityChange: (qty: number) => void;
}) {
  const Icon = productIcons[product.productType];

  return (
    <div className="flex h-full flex-col rounded-2xl border border-edge bg-card p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-3">
          <div className="rounded-xl border border-edge bg-overlay p-3">
            <Icon className="h-5 w-5 text-[var(--color-primary)]" />
          </div>
          <div className="min-w-0">
            <h2 className="text-base font-semibold text-[var(--color-text)]">{product.name}</h2>
            <p className="mt-1 text-sm text-[var(--color-text-muted)]">{product.description}</p>
          </div>
        </div>
        {product.available && (
          <div className="text-right">
            <p className="text-lg font-bold text-[var(--color-primary)]">₹{product.price}</p>
          </div>
        )}
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {formatCreditLineItems(product.credits).map((line) => (
          <span
            key={line}
            className="rounded-full border border-edge bg-overlay px-3 py-1 text-xs text-[var(--color-text-muted)]"
          >
            {line}
          </span>
        ))}
      </div>

      <div className="mt-auto flex items-center justify-between gap-3 pt-5">
        <p className="text-xs text-[var(--color-text-muted)]">Max {MAX_QTY_PER_PRODUCT} per checkout</p>
        {!product.available ? (
          <button
            disabled
            className="rounded-lg border border-edge bg-overlay px-4 py-2 text-sm font-semibold text-[var(--color-text-muted)] opacity-70"
          >
            {product.statusLabel ?? 'Coming soon'}
          </button>
        ) : quantity === 0 ? (
          <button
            onClick={() => onQuantityChange(1)}
            disabled={disabled}
            className="btn-primary rounded-lg px-4 py-2 text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-50"
          >
            Add to Cart
          </button>
        ) : (
          <div className="flex items-center rounded-lg border border-edge bg-overlay">
            <button
              onClick={() => onQuantityChange(quantity - 1)}
              disabled={disabled}
              className="px-3 py-2 text-sm text-[var(--color-text)] disabled:cursor-not-allowed disabled:opacity-50"
            >
              -
            </button>
            <span className="min-w-10 text-center text-sm font-semibold text-[var(--color-text)]">
              {quantity}
            </span>
            <button
              onClick={() => onQuantityChange(quantity + 1)}
              disabled={disabled || quantity >= MAX_QTY_PER_PRODUCT}
              className="px-3 py-2 text-sm text-[var(--color-text)] disabled:cursor-not-allowed disabled:opacity-50"
            >
              +
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function ShopCart({
  items,
  total,
  checkoutLabel,
  disabled,
  onCheckout,
}: {
  items: CartItem[];
  total: number;
  checkoutLabel: string;
  disabled: boolean;
  onCheckout: () => void;
}) {
  return (
    <aside className="rounded-2xl border border-edge bg-card p-5 lg:sticky lg:top-24">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-[var(--color-text)]">Cart</h2>
        <span className="text-sm text-[var(--color-text-muted)]">
          {items.reduce((sum, item) => sum + item.qty, 0)} items
        </span>
      </div>

      {items.length === 0 ? (
        <p className="mt-4 text-sm text-[var(--color-text-muted)]">
          Add a pack to start your refill checkout.
        </p>
      ) : (
        <div className="mt-4 space-y-3">
          {items.map((item) => (
            <div key={item.productId} className="rounded-xl border border-edge bg-overlay px-4 py-3">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="text-sm font-medium text-[var(--color-text)]">{item.label}</p>
                  <p className="text-xs text-[var(--color-text-muted)]">
                    ₹{item.price} x {item.qty}
                  </p>
                </div>
                <p className="text-sm font-semibold text-[var(--color-text)]">
                  ₹{item.price * item.qty}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}

      <div className="mt-5 border-t border-edge pt-4">
        <div className="flex items-center justify-between">
          <span className="text-sm text-[var(--color-text-muted)]">Total</span>
          <span className="text-lg font-bold text-[var(--color-text)]">₹{total}</span>
        </div>

        <button
          onClick={onCheckout}
          disabled={disabled || items.length === 0}
          className="btn-primary mt-4 flex w-full items-center justify-center rounded-lg px-4 py-2.5 text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-50"
        >
          {checkoutLabel}
        </button>

        <p className="mt-3 text-xs text-[var(--color-text-muted)]">
          Need billing help?{' '}
          <a
            href={`mailto:${SUPPORT_EMAIL}`}
            className="text-[var(--color-text)] underline decoration-[var(--color-primary)]/40 underline-offset-4 transition-colors hover:text-[var(--color-primary)]"
          >
            {SUPPORT_EMAIL}
          </a>
        </p>
      </div>
    </aside>
  );
}

export default function ShopPage() {
  const router = useRouter();
  const { isPremium } = useAuthStore();
  const { data: credits, isLoading: creditsLoading } = useCreditBalance();
  const { checkoutCart, isPending } = useProductCheckout();
  const [cart, setCart] = useState<CartItem[]>([]);
  const [checkoutLabel, setCheckoutLabel] = useState('Checkout');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    if (!isPremium) {
      router.replace('/pricing');
    }
  }, [isPremium, router]);

  const total = useMemo(
    () => cart.reduce((sum, item) => sum + item.price * item.qty, 0),
    [cart],
  );
  const balanceItems = credits
    ? [
        { label: 'Resume uploads', value: credits.resume_upload },
        { label: 'ATS checks', value: credits.ats_check },
        { label: 'Curated lists', value: credits.curated_list },
        { label: 'Cover letters', value: credits.cv_generation },
      ]
    : [];

  if (!isPremium) {
    return null;
  }

  const handleQuantityChange = (
    product: (typeof premiumShopProducts)[number],
    qty: number,
  ) => {
    if (!product.available) return;

    setCart((currentCart) =>
      updateCartItem(
        currentCart,
        {
          productId: product.productType,
          label: product.name,
          price: product.price,
        },
        qty,
      ),
    );
  };

  const handleCheckout = async () => {
    if (cart.length === 0) return;

    setError(null);
    setSuccess(null);
    setCheckoutLabel('Opening checkout...');

    try {
      await checkoutCart(
        cart.map((item) => ({
          productType: item.productId,
          quantity: item.qty,
        })),
      );
      setCart([]);
      setSuccess('Purchase successful. Your credits have been added.');
    } catch (checkoutError) {
      const message =
        checkoutError instanceof Error ? checkoutError.message : 'Checkout failed.';
      setError(message);
    } finally {
      setCheckoutLabel('Checkout');
    }
  };

  return (
    <div>
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text)]">Credit Shop</h1>
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">
          Refill premium credits whenever you need more AI usage.
        </p>
      </div>

      <div className="mt-4 rounded-2xl border border-edge bg-card px-5 py-4">
        <div className="w-full">
          <p className="text-sm font-medium text-[var(--color-text)]">Your current balance</p>
          {creditsLoading ? (
            <div className="mt-2 flex items-center gap-2 text-sm text-[var(--color-text-muted)]">
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading credits...
            </div>
          ) : credits ? (
            <div className="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
              {balanceItems.map((item) => (
                <div
                  key={item.label}
                  className="rounded-lg bg-overlay px-3 py-2"
                >
                  <div className="flex items-center justify-between gap-3">
                    <p className="text-xs text-[var(--color-text-muted)]">{item.label}</p>
                    <p className="text-sm font-semibold text-[var(--color-text)]">
                      {item.value}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="mt-2 text-sm text-[var(--color-text-muted)]">
              Credit balance unavailable right now.
            </p>
          )}
        </div>
      </div>

      {error && (
        <div className="mt-4 rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400">
          {error}
        </div>
      )}
      {success && (
        <div className="mt-4 rounded-lg border border-[#39ff14]/30 bg-[#39ff14]/10 px-4 py-3 text-sm text-[#39ff14]">
          {success}
        </div>
      )}

      <div className="mt-8 grid gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <div className="grid gap-4 md:grid-cols-2">
          {premiumShopProducts.map((product) => {
            const quantity = cart.find((item) => item.productId === product.productType)?.qty ?? 0;
            return (
              <PackCard
                key={product.productType}
                product={product}
                quantity={quantity}
                disabled={isPending}
                onQuantityChange={(qty) => handleQuantityChange(product, qty)}
              />
            );
          })}
        </div>

        <ShopCart
          items={cart}
          total={total}
          checkoutLabel={isPending ? 'Waiting for payment...' : checkoutLabel}
          disabled={isPending}
          onCheckout={handleCheckout}
        />
      </div>
    </div>
  );
}
