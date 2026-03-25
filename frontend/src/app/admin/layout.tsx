'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/store/auth-store';
import { Building2, Users, CreditCard, ChevronLeft } from 'lucide-react';

const adminNav = [
  { href: '/admin/companies', label: 'Companies', icon: Building2 },
  { href: '/admin/users', label: 'Users', icon: Users },
  { href: '/admin/payments', label: 'Payments', icon: CreditCard },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isAdmin, isLoading } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && (!isAuthenticated || !isAdmin)) {
      router.push('/dashboard');
    }
  }, [isLoading, isAuthenticated, isAdmin, router]);

  if (isLoading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-[#00f0ff] border-t-transparent" />
      </div>
    );
  }

  if (!isAuthenticated || !isAdmin) {
    return null;
  }

  return (
    <div className="flex min-h-[calc(100vh-3.5rem)]">
      {/* Admin sidebar */}
      <aside className="hidden w-56 shrink-0 border-r border-edge bg-overlay lg:block">
        <div className="flex h-full flex-col">
          <div className="border-b border-edge px-4 py-3">
            <Link
              href="/dashboard"
              className="flex items-center gap-2 text-xs text-slate-500 hover:text-slate-300"
            >
              <ChevronLeft className="h-3 w-3" />
              Back to app
            </Link>
            <h2 className="mt-1 text-sm font-semibold text-[#00f0ff] text-glow-cyan">
              Admin Panel
            </h2>
          </div>
          <nav className="mt-2 flex-1 space-y-1 px-2">
            {adminNav.map((item) => {
              const Icon = item.icon;
              const isActive =
                pathname === item.href || pathname.startsWith(item.href + '/');
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all ${
                    isActive
                      ? 'bg-[#00f0ff]/10 text-[#00f0ff] border-l-2 border-[#00f0ff] glow-cyan'
                      : 'text-slate-400 hover:bg-card hover:text-slate-200'
                  }`}
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>
        </div>
      </aside>

      {/* Mobile admin nav */}
      <div className="border-b border-edge px-4 py-2 lg:hidden">
        <div className="flex items-center gap-4 overflow-x-auto">
          {adminNav.map((item) => {
            const isActive =
              pathname === item.href || pathname.startsWith(item.href + '/');
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`shrink-0 rounded-md px-3 py-1.5 text-sm font-medium ${
                  isActive
                    ? 'bg-[#00f0ff]/10 text-[#00f0ff]'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-6xl">{children}</div>
      </div>
    </div>
  );
}
