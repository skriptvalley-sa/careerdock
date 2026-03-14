'use client';

import { createContext, useContext, useState, useEffect, useCallback } from 'react';

const SIDEBAR_KEY = 'careerdock:sidebar-collapsed';

export interface SidebarState {
  collapsed: boolean;
  toggle: () => void;
  mobileOpen: boolean;
  toggleMobile: () => void;
  closeMobile: () => void;
}

export const SidebarContext = createContext<SidebarState>({
  collapsed: false,
  toggle: () => {},
  mobileOpen: false,
  toggleMobile: () => {},
  closeMobile: () => {},
});

export function useSidebar() {
  return useContext(SidebarContext);
}

/** Hook that creates the sidebar state — used once in the provider. */
export function useSidebarState(): SidebarState {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  // Hydrate from localStorage after mount
  useEffect(() => {
    const stored = localStorage.getItem(SIDEBAR_KEY);
    if (stored === 'true') setCollapsed(true);
  }, []);

  const toggle = useCallback(() => {
    setCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem(SIDEBAR_KEY, String(next));
      return next;
    });
  }, []);

  const toggleMobile = useCallback(() => {
    setMobileOpen((prev) => !prev);
  }, []);

  const closeMobile = useCallback(() => {
    setMobileOpen(false);
  }, []);

  return { collapsed, toggle, mobileOpen, toggleMobile, closeMobile };
}
