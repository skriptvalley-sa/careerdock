'use client';

import { ThemeProvider } from 'next-themes';
import { QueryClient, QueryClientProvider, useQueryClient } from '@tanstack/react-query';
import { useState, useEffect, useCallback, useRef } from 'react';
import { useAuth } from '@/hooks/use-auth';
import { useSSE } from '@/hooks/use-sse';
import { setAuthFailureHandler } from '@/lib/api';
import { useAuthStore } from '@/store/auth-store';
import { SidebarContext, useSidebarState } from '@/hooks/use-sidebar';

function AuthProvider({ children }: { children: React.ReactNode }) {
  const { checkSession } = useAuth();
  const logout = useAuthStore((s) => s.logout);
  const userId = useAuthStore((s) => s.user?.id ?? null);
  const queryClient = useQueryClient();
  const previousUserIDRef = useRef<string | null>(null);

  // Wire up the API client's auth failure callback to clear all user-scoped
  // client state when refresh fails.
  // This fires when a 401 occurs and refresh also fails.
  useEffect(() => {
    setAuthFailureHandler(() => {
      void queryClient.cancelQueries().finally(() => {
        queryClient.clear();
        logout();
      });
    });
  }, [logout, queryClient]);

  // Defensive session isolation: if the authenticated identity changes,
  // clear any user-scoped queries before the next user starts rendering.
  useEffect(() => {
    const previousUserID = previousUserIDRef.current;
    if (previousUserID && previousUserID !== userId) {
      void queryClient.cancelQueries().finally(() => {
        queryClient.clear();
      });
    }
    previousUserIDRef.current = userId;
  }, [queryClient, userId]);

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

  // Connect to SSE for real-time updates (resume processing, etc.)
  useSSE();

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
    <ThemeProvider attribute="class" defaultTheme="dark" disableTransitionOnChange={false}>
      <QueryClientProvider client={queryClient}>
        <SidebarContext.Provider value={sidebar}>
          <AuthProvider>{children}</AuthProvider>
        </SidebarContext.Provider>
      </QueryClientProvider>
    </ThemeProvider>
  );
}
