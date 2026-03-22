'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState, useEffect, useCallback } from 'react';
import { useAuth } from '@/hooks/use-auth';
import { setAuthFailureHandler } from '@/lib/api';
import { useAuthStore } from '@/store/auth-store';
import { SidebarContext, useSidebarState } from '@/hooks/use-sidebar';

function AuthProvider({ children }: { children: React.ReactNode }) {
  const { checkSession } = useAuth();
  const logout = useAuthStore((s) => s.logout);

  // Wire up the API client's auth failure callback to clear Zustand state.
  // This fires when a 401 occurs and refresh also fails.
  useEffect(() => {
    setAuthFailureHandler(() => {
      logout();
    });
  }, [logout]);

  // Check session on initial mount
  useEffect(() => {
    checkSession();
  }, [checkSession]);

  // Re-check session when the window regains focus (catches stale sessions
  // after the user leaves the tab idle beyond the access token TTL).
  const handleFocus = useCallback(() => {
    checkSession();
  }, [checkSession]);

  useEffect(() => {
    window.addEventListener('focus', handleFocus);
    return () => window.removeEventListener('focus', handleFocus);
  }, [handleFocus]);

  return <>{children}</>;
}

/** Register the service worker for offline company directory support.
 *  Disabled in development to prevent stale asset caching on public dev deployments. */
function useServiceWorker() {
  useEffect(() => {
    if (typeof window === 'undefined' || !('serviceWorker' in navigator)) return;

    if (process.env.NODE_ENV === 'development') {
      // Unregister any existing SW in dev mode to avoid stale caching
      navigator.serviceWorker.getRegistrations().then((registrations) => {
        for (const reg of registrations) {
          reg.unregister();
        }
      });
      return;
    }

    navigator.serviceWorker.register('/sw.js').catch(() => {
      // SW registration failed — offline caching unavailable, non-critical
    });
  }, []);
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            refetchOnWindowFocus: false,
            retry: 1,
          },
        },
      }),
  );

  const sidebar = useSidebarState();

  useServiceWorker();

  return (
    <QueryClientProvider client={queryClient}>
      <SidebarContext.Provider value={sidebar}>
        <AuthProvider>{children}</AuthProvider>
      </SidebarContext.Provider>
    </QueryClientProvider>
  );
}
