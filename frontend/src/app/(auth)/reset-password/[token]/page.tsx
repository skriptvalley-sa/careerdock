'use client';

import { use } from 'react';
import { ResetPasswordForm } from '@/components/auth/reset-password-form';

export default function ResetPasswordPage({
  params,
}: {
  params: Promise<{ token: string }>;
}) {
  const { token } = use(params);
  return <ResetPasswordForm token={token} />;
}
