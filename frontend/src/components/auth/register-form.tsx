'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/use-auth';
import { ApiError } from '@/lib/api';
import { registerSchema, type RegisterInput } from '@/lib/validations';

export function RegisterForm() {
  const { register: registerUser } = useAuth();
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterInput>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = async (data: RegisterInput) => {
    setError(null);
    try {
      await registerUser(data);
      router.push('/dashboard');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
    }
  };

  return (
    <div className="rounded-lg border border-edge bg-card p-8 shadow-sm">
      <div className="mb-6 text-center">
        <h1 className="text-2xl font-bold text-slate-100">Create your account</h1>
        <p className="mt-1 text-sm text-slate-500">Start tracking your job applications</p>
      </div>

      {error && (
        <div className="mb-4 rounded-md bg-red-900/30 p-3 text-sm text-red-400">{error}</div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-slate-300">
            Full name
          </label>
          <input
            id="name"
            type="text"
            autoComplete="name"
            {...register('name')}
            className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 shadow-sm focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
          />
          {errors.name && (
            <p className="mt-1 text-xs text-red-600">{errors.name.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="email" className="block text-sm font-medium text-slate-300">
            Email
          </label>
          <input
            id="email"
            type="email"
            autoComplete="email"
            {...register('email')}
            className="mt-1 block w-full rounded-md border border-edge-input bg-input px-3 py-2 text-sm text-slate-200 shadow-sm focus:border-[#00f0ff]/50 focus:outline-none focus:ring-1 focus:ring-[#00f0ff]/30"
          />
          {errors.email && (
            <p className="mt-1 text-xs text-red-600">{errors.email.message}</p>
          )}
        </div>

        <div>
          <label htmlFor="password" className="block text-sm font-medium text-slate-300">
            Password
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
          <p className="mt-1 text-xs text-slate-400">
            Min 8 chars, 1 uppercase, 1 lowercase, 1 digit
          </p>
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="btn-neon w-full rounded-md px-4 py-2 text-sm font-medium shadow-sm disabled:opacity-50"
        >
          {isSubmitting ? 'Creating account...' : 'Create account'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-slate-500">
        Already have an account?{' '}
        <Link href="/login" className="font-medium text-[#00f0ff] hover:text-[#00f0ff]/80">
          Sign in
        </Link>
      </p>
    </div>
  );
}
