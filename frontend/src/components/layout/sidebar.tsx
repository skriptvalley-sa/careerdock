'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import {
  LayoutDashboard,
  List,
  Briefcase,
  Building2,
  Settings,
  ChevronLeft,
  ChevronRight,
  LogIn,
  DollarSign,
  FileText,
  X,
} from 'lucide-react';
import { CreditBalance } from '@/components/credit-balance';

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  mobileOpen: boolean;
  onCloseMobile: () => void;
}

const authedNavItems = [
  { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { href: '/lists', label: 'My Lists', icon: List },
  { href: '/applications', label: 'Applications', icon: Briefcase },
  { href: '/resumes', label: 'Resumes', icon: FileText },
  { href: '/companies', label: 'Companies', icon: Building2 },
  { href: '/pricing', label: 'Pricing', icon: DollarSign },
  { href: '/settings', label: 'Settings', icon: Settings },
];

const publicNavItems = [
  { href: '/companies', label: 'Companies', icon: Building2 },
  { href: '/pricing', label: 'Pricing', icon: DollarSign },
];

export function Sidebar({ collapsed, onToggle, mobileOpen, onCloseMobile }: SidebarProps) {
  const pathname = usePathname();
  const { isAuthenticated, user } = useAuthStore();

  const navItems = isAuthenticated ? authedNavItems : publicNavItems;

  const sidebarContent = (
    <div className="flex h-full flex-col">
      {/* Toggle button */}
      <div className={`flex h-14 items-center border-b border-edge ${collapsed ? 'justify-center px-2' : 'justify-between px-4'}`}>
        {/* Mobile close button (shown only in mobile overlay) */}
        <button
          onClick={onCloseMobile}
          className="flex items-center justify-center rounded-md p-1.5 text-slate-400 hover:bg-card hover:text-slate-200 lg:hidden"
          aria-label="Close sidebar"
        >
          <X className="h-5 w-5" />
        </button>

        {/* Desktop collapse toggle */}
        <button
          onClick={onToggle}
          className="hidden items-center justify-center rounded-md p-1.5 text-slate-400 hover:bg-card hover:text-slate-200 lg:flex"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </button>
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
                  ? 'bg-[#00f0ff]/10 text-[#00f0ff] border-l-2 border-[#00f0ff] glow-cyan'
                  : 'text-slate-400 hover:bg-card hover:text-slate-200'
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

      {/* User info at bottom (authenticated only) */}
      {isAuthenticated && user && (
        <div className="border-t border-edge p-3">
          <div className={`flex items-center ${collapsed ? 'justify-center' : 'gap-3'}`}>
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#00f0ff]/15 text-sm font-medium text-[#00f0ff]">
              {user.name?.charAt(0)?.toUpperCase() || 'U'}
            </div>
            {!collapsed && (
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-slate-100">
                  {user.name}
                </p>
                <p className="truncate text-xs text-slate-500">{user.email}</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Sign in CTA (unauthenticated only) */}
      {!isAuthenticated && (
        <div className="border-t border-edge p-3">
          <Link
            href="/login"
            onClick={onCloseMobile}
            className={`btn-neon flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium ${collapsed ? 'justify-center px-2' : ''}`}
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
