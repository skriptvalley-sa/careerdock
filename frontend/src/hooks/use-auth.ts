'use client';

import { useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient, ApiError } from '@/lib/api';
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
 */
export function useAuth() {
  const { setUser, logout: storeLogout } = useAuthStore();
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
    storeLogout();
    router.push('/login');
  }, [storeLogout, router]);

  const refreshSession = useCallback(async () => {
    try {
      await apiClient.post('/api/auth/refresh');
      const user = await apiClient.get<User>('/api/auth/me');
      setUser(user);
      return true;
    } catch {
      storeLogout();
      return false;
    }
  }, [setUser, storeLogout]);

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
   * Called once on app mount via AuthProvider.
   */
  const checkSession = useCallback(async () => {
    try {
      const user = await apiClient.get<User>('/api/auth/me');
      setUser(user);
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        // Try refresh once
        try {
          await apiClient.post('/api/auth/refresh');
          const user = await apiClient.get<User>('/api/auth/me');
          setUser(user);
          return;
        } catch {
          // Refresh also failed — not authenticated
        }
      }
      setUser(null);
    }
  }, [setUser]);

  return {
    register,
    login,
    logout,
    refreshSession,
    forgotPassword,
    resetPassword,
    verifyEmail,
    checkSession,
  };
}
