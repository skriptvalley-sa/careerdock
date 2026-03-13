'use client';

import Link from 'next/link';
import { useAuthStore } from '@/store/auth-store';

export function Header() {
  const { isAuthenticated, user } = useAuthStore();

  return (
    <header className="border-b border-gray-200 bg-white">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-8">
          <Link href="/" className="text-xl font-bold text-gray-900">
            CareerDock
          </Link>
          <nav className="hidden items-center gap-6 md:flex">
            <Link
              href="/companies"
              className="text-sm font-medium text-gray-600 hover:text-gray-900"
            >
              Companies
            </Link>
            <Link
              href="/pricing"
              className="text-sm font-medium text-gray-600 hover:text-gray-900"
            >
              Pricing
            </Link>
          </nav>
        </div>

        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <Link
              href="/dashboard"
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              {user?.name?.split(' ')[0] || 'Dashboard'}
            </Link>
          ) : (
            <>
              <Link
                href="/login"
                className="text-sm font-medium text-gray-600 hover:text-gray-900"
              >
                Log in
              </Link>
              <Link
                href="/register"
                className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
              >
                Sign up
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
