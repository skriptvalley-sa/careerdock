'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import {
  LayoutDashboard,
  List,
  Briefcase,
  Building2,
  ChevronLeft,
  ChevronRight,
  LogIn,
  FileText,
  ScanSearch,
  ShoppingBag,
  Sparkles,
  Tag,
  X,
} from 'lucide-react';
import { CreditBalance } from '@/components/credit-balance';
import { ThemeToggle } from '@/components/ui/theme-toggle';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  mobileOpen: boolean;
  onCloseMobile: () => void;
}

function getAuthedNavItems(isPremium: boolean) {
  return [
    { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { href: '/lists', label: 'My Lists', icon: List },
    { href: '/applications', label: 'Applications', icon: Briefcase },
    { href: '/resumes', label: 'Resumes', icon: FileText },
    { href: '/ats', label: 'ATS Check', icon: ScanSearch },
    { href: '/curated-lists', label: 'Curated Lists', icon: Sparkles },
    { href: '/companies', label: 'Companies', icon: Building2 },
    ...(isPremium
      ? [{ href: '/shop', label: 'Credit Shop', icon: ShoppingBag }]
      : [{ href: '/pricing', label: 'Pricing', icon: Tag }]),
  ];
}

const publicNavItems = [
  { href: '/companies', label: 'Companies', icon: Building2 },
  { href: '/pricing', label: 'Pricing', icon: Tag },
];

export function Sidebar({ collapsed, onToggle, mobileOpen, onCloseMobile }: SidebarProps) {
  const pathname = usePathname();
  const { isAuthenticated, user } = useAuthStore();
  const navItems = isAuthenticated ? getAuthedNavItems(!!user?.premium_since) : publicNavItems;

  const sidebarContent = (
    <div className="flex h-full flex-col">
      {/* Toggle button */}
      <div className={`flex h-14 items-center border-b border-edge ${collapsed ? 'justify-center px-2' : 'justify-between px-4'}`}>
        {/* Mobile close button (shown only in mobile overlay) */}
        <button
          onClick={onCloseMobile}
          className="flex items-center justify-center rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-card hover:text-[var(--color-text)] lg:hidden"
          aria-label="Close sidebar"
        >
          <X className="h-5 w-5" />
        </button>

        {/* Desktop collapse toggle */}
        <button
          onClick={onToggle}
          className="hidden items-center justify-center rounded-md p-1.5 text-[var(--color-text-muted)] hover:bg-card hover:text-[var(--color-text)] lg:flex"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </button>

        {!collapsed && <ThemeToggle />}
      </div>

      {/* Navigation */}
      <nav className="mt-3 flex-1 space-y-1 overflow-y-auto px-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href || pathname.startsWith(item.href + '/');
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={onCloseMobile}
              title={collapsed ? item.label : undefined}
              className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-all ${
                isActive
                  ? 'bg-[var(--color-primary)]/10 nav-link-active glow-primary'
                  : 'text-[var(--color-text-muted)] hover:bg-card hover:text-[var(--color-text)]'
              } ${collapsed ? 'justify-center px-2' : ''}`}
            >
              <Icon className="h-4 w-4 shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      {/* Credit balance (premium users) */}
      {isAuthenticated && (
        <div className="px-3 pb-2">
          <CreditBalance collapsed={collapsed} />
        </div>
      )}

      {/* Sign in CTA (unauthenticated only) */}
      {!isAuthenticated && (
        <div className="border-t border-edge p-3">
          <Link
            href="/login"
            onClick={onCloseMobile}
            className={`btn-primary flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium ${collapsed ? 'justify-center px-2' : ''}`}
          >
            <LogIn className="h-4 w-4 shrink-0" />
            {!collapsed && <span>Sign in</span>}
          </Link>
        </div>
      )}
    </div>
  );

  return (
    <>
      {/* Desktop sidebar — fixed */}
      <aside
        className={`fixed left-0 top-14 z-30 hidden h-[calc(100vh-3.5rem)] border-r border-edge bg-overlay transition-all duration-200 lg:flex lg:flex-col ${
          collapsed ? 'w-16' : 'w-56'
        }`}
      >
        {sidebarContent}
      </aside>

      {/* Mobile overlay */}
      {mobileOpen && (
        <>
          <div
            className="fixed inset-0 z-40 bg-black/60 lg:hidden"
            onClick={onCloseMobile}
          />
          <aside className="fixed left-0 top-0 z-50 flex h-screen w-56 flex-col border-r border-edge bg-overlay lg:hidden">
            {sidebarContent}
          </aside>
        </>
      )}
    </>
  );
}
