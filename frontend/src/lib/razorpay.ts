// Razorpay Checkout.js dynamic loader + TypeScript declarations.
// Razorpay doesn't provide an npm package — it's loaded via a <script> tag.

declare global {
  interface Window {
    Razorpay: new (options: RazorpayOptions) => RazorpayInstance;
  }
}

export interface RazorpayOptions {
  key: string;
  amount: number; // in paise
  currency: string;
  name: string;
  description: string;
  order_id: string;
  handler: (response: RazorpayPaymentResponse) => void;
  prefill?: {
    name?: string;
    email?: string;
  };
  theme?: {
    color?: string;
  };
  modal?: {
    ondismiss?: () => void;
    escape?: boolean;
  };
}

export interface RazorpayPaymentResponse {
  razorpay_payment_id: string;
  razorpay_order_id: string;
  razorpay_signature: string;
}

export interface RazorpayInstance {
  open: () => void;
  close: () => void;
  on: (event: string, handler: (...args: unknown[]) => void) => void;
}

let loadPromise: Promise<void> | null = null;

/**
 * Dynamically loads the Razorpay Checkout.js script. Safe to call multiple
 * times — the script is loaded only once.
 */
export function loadRazorpayScript(): Promise<void> {
  if (loadPromise) return loadPromise;

  loadPromise = new Promise<void>((resolve, reject) => {
    if (typeof window !== 'undefined' && window.Razorpay) {
      resolve();
      return;
    }

    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.async = true;
    script.onload = () => resolve();
    script.onerror = () => {
      loadPromise = null;
      reject(new Error('Failed to load Razorpay Checkout.js'));
    };
    document.body.appendChild(script);
  });

  return loadPromise;
}

/**
 * Opens the Razorpay checkout modal with the given options.
 * Ensures the script is loaded first.
 */
export async function openRazorpayCheckout(options: RazorpayOptions): Promise<void> {
  await loadRazorpayScript();
  const rzp = new window.Razorpay(options);
  rzp.open();
}
