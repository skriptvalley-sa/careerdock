import { create } from 'zustand';
import type { User } from '@/types/api';

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isPremium: boolean;
  isAdmin: boolean;
  isModerator: boolean;
  setUser: (user: User | null) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoading: true,
  isAuthenticated: false,
  isPremium: false,
  isAdmin: false,
  isModerator: false,

  setUser: (user) =>
    set({
      user,
      isLoading: false,
      isAuthenticated: !!user,
      isPremium: !!user?.premium_since,
      isAdmin: user?.role === 'admin',
      isModerator: user?.role === 'moderator' || user?.role === 'admin',
    }),

  logout: () =>
    set({
      user: null,
      isLoading: false,
      isAuthenticated: false,
      isPremium: false,
      isAdmin: false,
      isModerator: false,
    }),
}));
