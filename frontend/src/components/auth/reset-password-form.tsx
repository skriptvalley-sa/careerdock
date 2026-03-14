'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { useAuth } from '@/hooks/use-auth';
import { ApiError } from '@/lib/api';
import { resetPasswordSchema, type ResetPasswordInput } from '@/lib/validations';

interface ResetPasswordFormProps {
  token: string;
}

export function ResetPasswordForm({ token }: ResetPasswordFormProps) {
  const { resetPassword } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordInput>({
    resolver: zodResolver(resetPasswordSchema),
  });

  const onSubmit = async (data: ResetPasswordInput) => {
    setError(null);
    try {
      await resetPassword(token, data.password);
      setSuccess(true);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
    }
  };

  if (success) {
    return (
      <div className="rounded-lg border border-edge bg-card p-8 shadow-sm text-center">
        <h1 className="text-2xl font-bold text-slate-100">Password reset</h1>
        <p className="mt-2 text-sm text-slate-500">
          Your password has been reset successfully.
        </p>
        <Link
          href="/login"
          className="btn-neon mt-4 inline-block rounded-md px-4 py-2 text-sm font-medium"
        >
          Sign in
        </Link>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-edge bg-card p-8 shadow-sm">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-bold text-slate-100">Set a new password</h1>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-900/30 p-3 text-sm text-red-400">{error}</div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label htmlFor="password" className="block text-sm font-medium text-slate-300">
            New password
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            {...register('password')}
            className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 shadow-sm focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
          />
          {errors.password && (
            <p className="mt-1 text-xs text-red-600">{errors.password.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="confirmPassword" className="block text-sm font-medium text-slate-300">
            Confirm password
          </label>
          <input
            id="confirmPassword"
            type="password"
            autoComplete="new-password"
            {...register('confirmPassword')}
            className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 shadow-sm focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
          />
          {errors.confirmPassword && (
            <p className="mt-1 text-xs text-red-600">{errors.confirmPassword.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="btn-neon w-full rounded-md px-4 py-2 text-sm font-medium shadow-sm disabled:opacity-50"
        >
          {isSubmitting ? 'Resetting...' : 'Reset password'}
        </button>
      </form>
    </div>
  );
}
