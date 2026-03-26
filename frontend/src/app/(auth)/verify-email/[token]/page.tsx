'use client';

import { use, useEffect, useState } from 'react';
import Link from 'next/link';
import { useAuth } from '@/hooks/use-auth';
import { ApiError } from '@/lib/api';

export default function VerifyEmailPage({
  params,
}: {
  params: Promise<{ token: string }>;
}) {
  const { token } = use(params);
  const { verifyEmail } = useAuth();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    verifyEmail(token)
      .then(() => setStatus('success'))
      .catch((err) => {
        setStatus('error');
        if (err instanceof ApiError) {
          setError(err.message);
        } else {
          setError('An unexpected error occurred');
        }
      });
  }, [token, verifyEmail]);

  return (
    <div className="rounded-lg border border-edge bg-card p-8 shadow-sm text-center">
      {status === 'loading' && (
        <>
          <div className="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-primary)] border-t-transparent" />
          <p className="mt-4 text-sm text-[var(--color-text-muted)]">Verifying your email...</p>
        </>
      )}

      {status === 'success' && (
        <>
          <h1 className="text-2xl font-bold text-[var(--color-text)]">Email verified!</h1>
          <p className="mt-2 text-sm text-[var(--color-text-muted)]">
            Your email has been verified. You can now use all features.
          </p>
          <Link
            href="/dashboard"
            className="btn-primary mt-4 inline-block rounded-md px-4 py-2 text-sm font-medium"
          >
            Go to dashboard
          </Link>
        </>
      )}

      {status === 'error' && (
        <>
          <h1 className="text-2xl font-bold text-[var(--color-text)]">Verification failed</h1>
          <p className="mt-2 text-sm text-red-400">{error}</p>
          <Link
            href="/login"
            className="mt-4 inline-block text-sm font-medium text-[var(--color-primary)] hover:text-[var(--color-primary)]/80"
          >
            Back to sign in
          </Link>
        </>
      )}
    </div>
  );
}
