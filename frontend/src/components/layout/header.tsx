'use client';

import { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { useAuth } from '@/hooks/use-auth';
import { useSidebar } from '@/hooks/use-sidebar';
import {
  ChevronDown,
  LogOut,
  Menu,
  Settings,
  Shield,
  ShieldCheck,
} from 'lucide-react';
import { NotificationBell } from '@/components/notifications/notification-bell';

/** Routes where the mobile hamburger should be hidden (auth pages). */
const NO_HAMBURGER_PREFIXES = ['/login', '/register', '/forgot-password', '/reset-password', '/verify-email'];

export function Header() {
  const { isAuthenticated, user } = useAuthStore();
  const { logout } = useAuth();
  const { toggleMobile } = useSidebar();
  const pathname = usePathname();
  const [accountMenuOpen, setAccountMenuOpen] = useState(false);
  const accountMenuRef = useRef<HTMLDivElement>(null);

  const showHamburger = !NO_HAMBURGER_PREFIXES.some((p) => pathname.startsWith(p));
  const isAdmin = user?.role === 'admin';
  const isModerator = user?.role === 'moderator' || user?.role === 'admin';
  const accountNavItems = isAuthenticated
    ? [
        { href: '/settings', label: 'Settings', icon: Settings },
        ...(isModerator ? [{ href: '/moderator', label: 'Moderator', icon: Shield }] : []),
        ...(isAdmin ? [{ href: '/admin', label: 'Admin', icon: ShieldCheck }] : []),
      ]
    : [];

  useEffect(() => {
    if (!accountMenuOpen) return;

    function handleClickOutside(event: MouseEvent) {
      if (accountMenuRef.current && !accountMenuRef.current.contains(event.target as Node)) {
        setAccountMenuOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [accountMenuOpen]);

  useEffect(() => {
    setAccountMenuOpen(false);
  }, [pathname]);

  return (
    <header className="fixed left-0 right-0 top-0 z-40 border-b border-edge bg-overlay">
      <div className="flex h-14 items-center justify-between px-4 sm:px-6">
        <div className="flex items-center gap-3">
          {/* Mobile hamburger */}
          {showHamburger && (
            <button
              onClick={toggleMobile}
              className="rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-card hover:text-[var(--color-text)] lg:hidden"
              aria-label="Toggle navigation"
            >
              <Menu className="h-5 w-5" />
            </button>
          )}

          <Link
            href={isAuthenticated ? '/dashboard' : '/'}
            className="text-lg font-bold text-[var(--color-primary)]"
          >
            CareerDock
          </Link>
        </div>

        <div className="flex items-center gap-3">
          {isAuthenticated ? (
            <>
              <NotificationBell />
              <div className="relative" ref={accountMenuRef}>
                <button
                  onClick={() => setAccountMenuOpen((open) => !open)}
                  className={`flex items-center gap-2 rounded-md border px-2.5 py-1.5 text-sm font-medium transition-colors ${
                    accountMenuOpen
                      ? 'border-[var(--color-primary)]/40 bg-[var(--color-primary)]/10 text-[var(--color-text)]'
                      : 'border-edge text-[var(--color-text)] hover:bg-card hover:text-[var(--color-text)]'
                  }`}
                  aria-label="Account menu"
                  type="button"
                >
                  <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary)]/15 text-xs font-semibold text-[var(--color-primary)]">
                    {user?.name?.charAt(0)?.toUpperCase() || 'U'}
                  </div>
                  <span className="hidden max-w-32 truncate sm:inline">{user?.name ?? 'Account'}</span>
                  <ChevronDown
                    className={`h-4 w-4 shrink-0 text-[var(--color-text-muted)] transition-transform ${
                      accountMenuOpen ? 'rotate-180' : ''
                    }`}
                  />
                </button>

                {accountMenuOpen && (
                  <div className="absolute right-0 top-full z-40 mt-2 w-80 max-w-[calc(100vw-2rem)] rounded-xl border border-edge bg-card shadow-xl">
                    <div className="border-b border-edge px-4 py-4">
                      <div className="flex items-start gap-3">
                        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-[var(--color-primary)]/15 text-base font-medium text-[var(--color-primary)]">
                          {user?.name?.charAt(0)?.toUpperCase() || 'U'}
                        </div>
                        <div className="min-w-0">
                          <div className="flex flex-wrap items-center gap-2">
                            <p className="text-sm font-semibold text-[var(--color-text)]">
                              {user?.name ?? 'Account'}
                            </p>
                            {(user?.role === 'moderator' || user?.role === 'admin') && (
                              <span className={`rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase ${
                                user?.role === 'admin'
                                  ? 'bg-[var(--color-warning)]/15 text-[var(--color-warning)]'
                                  : 'bg-[var(--color-primary)]/10 text-[var(--color-primary)]'
                              }`}>
                                {user?.role === 'admin' ? 'Admin' : 'Mod'}
                              </span>
                            )}
                          </div>
                          <p className="mt-1 break-all text-xs text-[var(--color-text-muted)]">
                            {user?.email}
                          </p>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-1 p-2">
                      {accountNavItems.map((item) => {
                        const Icon = item.icon;
                        const isActive = pathname === item.href || pathname.startsWith(item.href + '/');

                        return (
                          <Link
                            key={item.href}
                            href={item.href}
                            onClick={() => setAccountMenuOpen(false)}
                            className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all ${
                              isActive
                                ? 'bg-[var(--color-primary)]/10 nav-link-active glow-primary'
                                : 'text-[var(--color-text-muted)] hover:bg-overlay hover:text-[var(--color-text)]'
                            }`}
                          >
                            <Icon className="h-4 w-4 shrink-0" />
                            <span>{item.label}</span>
                          </Link>
                        );
                      })}
                    </div>

                    <div className="border-t border-edge p-2">
                      <button
                        onClick={() => {
                          setAccountMenuOpen(false);
                          logout();
                        }}
                        className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-[var(--color-text-muted)] transition-colors hover:bg-overlay hover:text-[var(--color-text)]"
                        type="button"
                      >
                        <LogOut className="h-4 w-4 shrink-0" />
                        <span>Sign out</span>
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </>
          ) : (
            <>
              <Link
                href="/login"
                className="text-sm font-medium text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
              >
                Log in
              </Link>
              <Link
                href="/register"
                className="btn-primary rounded-md px-4 py-1.5 text-sm font-medium"
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
