'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { useAuth } from '@/hooks/use-auth';
import { ApiError } from '@/lib/api';
import { forgotPasswordSchema, type ForgotPasswordInput } from '@/lib/validations';

export function ForgotPasswordForm() {
  const { forgotPassword } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [sent, setSent] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordInput>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  const onSubmit = async (data: ForgotPasswordInput) => {
    setError(null);
    try {
      await forgotPassword(data.email);
      setSent(true);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
    }
  };

  if (sent) {
    return (
      <div className="rounded-lg border border-edge bg-card p-8 shadow-sm text-center">
        <h1 className="text-2xl font-bold text-[var(--color-text)]">Check your email</h1>
        <p className="mt-2 text-sm text-[var(--color-text-muted)]">
          If an account with that email exists, we&apos;ve sent a password reset link.
        </p>
        <Link
          href="/login"
          className="mt-4 inline-block text-sm font-medium text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
        >
          Back to sign in
        </Link>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-edge bg-card p-8 shadow-sm">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-bold text-[var(--color-text)]">Reset your password</h1>
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">
          Enter your email and we&apos;ll send you a reset link
        </p>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-900/30 p-3 text-sm text-red-400">{error}</div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label htmlFor="email" className="block text-sm font-medium text-[var(--color-text)]">
            Email
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            {...register('email')}
            className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-[var(--color-text)] shadow-sm focus:border-[var(--color-primary)]/50 focus:outline-none focus:ring-1 focus:ring-[var(--color-primary)]/30"
          />
          {errors.email && (
            <p className="mt-1 text-xs text-red-600">{errors.email.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="btn-primary w-full rounded-md px-4 py-2 text-sm font-medium shadow-sm disabled:opacity-50"
        >
          {isSubmitting ? 'Sending...' : 'Send reset link'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-[var(--color-text-muted)]">
        <Link href="/login" className="font-medium text-[var(--color-primary)] hover:text-[var(--color-primary)]/80">
          Back to sign in
        </Link>
      </p>
    </div>
  );
}
