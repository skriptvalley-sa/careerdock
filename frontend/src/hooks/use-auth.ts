'use client';

import { useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api';
import { useAuthStore } from '@/store/auth-store';
import type { User } from '@/types/api';

interface RegisterInput {
  email: string;
  password: string;
  name: string;
}

interface LoginInput {
  email: string;
  password: string;
}

/**
 * useAuth provides authentication actions that call the backend API
 * and update the Zustand auth store.
 *
 * Note: 401 → refresh → retry is handled transparently by the API client
 * (see lib/api.ts). Hooks here only need to call the API; if the access
 * token is expired the client will silently refresh and retry.
 */
export function useAuth() {
  const { setUser, logout: storeLogout } = useAuthStore();
  const qc = useQueryClient();
  const router = useRouter();

  const register = useCallback(
    async (input: RegisterInput) => {
      const user = await apiClient.post<User>('/api/auth/register', input);
      setUser(user);
      return user;
    },
    [setUser],
  );

  const login = useCallback(
    async (input: LoginInput) => {
      const user = await apiClient.post<User>('/api/auth/login', input);
      setUser(user);
      return user;
    },
    [setUser],
  );

  const logout = useCallback(async () => {
    try {
      await apiClient.post('/api/auth/logout');
    } catch {
      // Ignore errors — clear local state regardless
    }
    await qc.cancelQueries();
    qc.clear();
    storeLogout();
    router.push('/login');
  }, [qc, storeLogout, router]);

  const forgotPassword = useCallback(async (email: string) => {
    await apiClient.post('/api/auth/forgot-password', { email });
  }, []);

  const resetPassword = useCallback(
    async (token: string, newPassword: string) => {
      await apiClient.post('/api/auth/reset-password', {
        token,
        new_password: newPassword,
      });
    },
    [],
  );

  const verifyEmail = useCallback(async (token: string) => {
    await apiClient.post('/api/auth/verify-email', { token });
  }, []);

  /**
   * Check current session by calling GET /api/auth/me.
   * The API client handles 401 → refresh → retry automatically,
   * so this just needs to fetch and update the store.
   * Called on app mount and on window focus via AuthProvider.
   */
  const checkSession = useCallback(async () => {
    try {
      const user = await apiClient.get<User>('/api/auth/me');
      setUser(user);
    } catch {
      // Any failure (including post-refresh 401) means not authenticated.
      // The API client's onAuthFailure handler also clears state,
      // but we set null here too for the initial mount case.
      setUser(null);
    }
  }, [setUser]);

  return {
    register,
    login,
    logout,
    forgotPassword,
    resetPassword,
    verifyEmail,
    checkSession,
  };
}
