'use client';

import { usePathname } from 'next/navigation';
import { Sidebar } from './sidebar';
import { Footer } from './footer';
import { useSidebar } from '@/hooks/use-sidebar';

/** Routes where the sidebar is hidden (auth pages + landing). */
const NO_SIDEBAR_PREFIXES = ['/login', '/register', '/forgot-password', '/reset-password', '/verify-email'];
const NO_SIDEBAR_EXACT = ['/'];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { collapsed, toggle, mobileOpen, closeMobile } = useSidebar();

  const hideSidebar =
    NO_SIDEBAR_EXACT.includes(pathname) ||
    NO_SIDEBAR_PREFIXES.some((p) => pathname.startsWith(p));

  if (hideSidebar) {
    return (
      <div className="flex min-h-[calc(100vh-3.5rem)] flex-col">
        <main className="flex-1">{children}</main>
        <Footer />
      </div>
    );
  }

  return (
    <>
      <Sidebar
        collapsed={collapsed}
        onToggle={toggle}
        mobileOpen={mobileOpen}
        onCloseMobile={closeMobile}
      />
      <div
        className={`flex min-h-[calc(100vh-3.5rem)] flex-col transition-all duration-200 ${
          collapsed ? 'lg:ml-16' : 'lg:ml-56'
        }`}
      >
        <main className="flex-1">{children}</main>
        <Footer />
      </div>
    </>
  );
}
