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
    <div className="rounded-lg border border-gray-200 bg-white p-8 shadow-sm text-center">
      {status === 'loading' && (
        <>
          <div className="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
          <p className="mt-4 text-sm text-gray-500">Verifying your email...</p>
        </>
      )}

      {status === 'success' && (
        <>
          <h1 className="text-2xl font-bold text-gray-900">Email verified!</h1>
          <p className="mt-2 text-sm text-gray-500">
            Your email has been verified. You can now use all features.
          </p>
          <Link
            href="/dashboard"
            className="mt-4 inline-block rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            Go to dashboard
          </Link>
        </>
      )}

      {status === 'error' && (
        <>
          <h1 className="text-2xl font-bold text-gray-900">Verification failed</h1>
          <p className="mt-2 text-sm text-red-600">{error}</p>
          <Link
            href="/login"
            className="mt-4 inline-block text-sm font-medium text-blue-600 hover:text-blue-500"
          >
            Back to sign in
          </Link>
        </>
      )}
    </div>
  );
}
