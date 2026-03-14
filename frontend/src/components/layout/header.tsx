'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { useAuth } from '@/hooks/use-auth';
import { useSidebar } from '@/hooks/use-sidebar';
import { LogOut, Menu } from 'lucide-react';

/** Routes where the mobile hamburger should be hidden (auth pages). */
const NO_HAMBURGER_PREFIXES = ['/login', '/register', '/forgot-password', '/reset-password', '/verify-email'];

export function Header() {
  const { isAuthenticated } = useAuthStore();
  const { logout } = useAuth();
  const { toggleMobile } = useSidebar();
  const pathname = usePathname();

  const showHamburger = !NO_HAMBURGER_PREFIXES.some((p) => pathname.startsWith(p));

  return (
    <header className="fixed left-0 right-0 top-0 z-40 border-b border-edge bg-overlay">
      <div className="flex h-14 items-center justify-between px-4 sm:px-6">
        <div className="flex items-center gap-3">
          {/* Mobile hamburger */}
          {showHamburger && (
            <button
              onClick={toggleMobile}
              className="rounded-md p-1.5 text-slate-400 hover:bg-card hover:text-slate-200 lg:hidden"
              aria-label="Toggle navigation"
            >
              <Menu className="h-5 w-5" />
            </button>
          )}

          <Link
            href={isAuthenticated ? '/dashboard' : '/'}
            className="text-lg font-bold text-[#00f0ff] text-glow-cyan"
          >
            CareerDock
          </Link>
        </div>

        <div className="flex items-center gap-4">
          {isAuthenticated ? (
            <button
              onClick={logout}
              className="flex items-center gap-2 rounded-md border border-edge px-3 py-1.5 text-sm font-medium text-slate-300 hover:bg-card hover:text-slate-100"
            >
              <LogOut className="h-4 w-4" />
              <span className="hidden sm:inline">Sign out</span>
            </button>
          ) : (
            <>
              <Link
                href="/login"
                className="text-sm font-medium text-slate-400 hover:text-slate-200"
              >
                Log in
              </Link>
              <Link
                href="/register"
                className="btn-neon rounded-md px-4 py-1.5 text-sm font-medium"
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
