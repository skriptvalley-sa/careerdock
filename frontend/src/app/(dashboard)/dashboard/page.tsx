'use client';

import { useAuthStore } from '@/store/auth-store';
import { useAuth } from '@/hooks/use-auth';

export default function DashboardPage() {
  const { user } = useAuthStore();
  const { logout } = useAuth();

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Welcome back, {user?.name ?? 'User'}
          </p>
        </div>
        <button
          onClick={logout}
          className="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50"
        >
          Sign out
        </button>
      </div>

      {/* TODO (Sprint 2): Funnel view, recent activity, quick add */}
      <div className="mt-8 rounded-lg border border-dashed border-gray-300 p-12 text-center">
        <p className="text-sm text-gray-500">
          Dashboard features coming in Sprint 2
        </p>
      </div>
    </div>
  );
}
