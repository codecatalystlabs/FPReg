import { create } from 'zustand';
import * as SecureStore from 'expo-secure-store';
import { authApi } from '../api/auth';
import { STORAGE_KEYS } from '../constants';
import { logger } from '../utils/logger';
import type { User, TokenPair } from '../types';

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  isRestoringSession: boolean;

  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  restoreSession: () => Promise<void>;
  setUser: (user: User) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  isLoading: false,
  isRestoringSession: true,

  login: async (email: string, password: string) => {
    set({ isLoading: true });
    try {
      const response = await authApi.login(email, password);
      const { tokens, user } = response;

      await persistTokens(tokens);
      await SecureStore.setItemAsync(STORAGE_KEYS.USER, JSON.stringify(user));

      set({ user, isAuthenticated: true, isLoading: false });
      logger.auth('Login', user.email);
    } catch (error) {
      set({ isLoading: false });
      logger.error('Auth', 'Login failed');
      throw error;
    }
  },

  logout: async () => {
    const refreshToken = await SecureStore.getItemAsync(STORAGE_KEYS.REFRESH_TOKEN);
    try {
      if (refreshToken) {
        await authApi.logout(refreshToken);
      }
    } catch {
      // Server logout is best-effort
    }

    await clearTokens();
    set({ user: null, isAuthenticated: false });
    logger.auth('Logout');
  },

  restoreSession: async () => {
    try {
      const accessToken = await SecureStore.getItemAsync(STORAGE_KEYS.ACCESS_TOKEN);
      const userJson = await SecureStore.getItemAsync(STORAGE_KEYS.USER);

      if (accessToken && userJson) {
        const user = JSON.parse(userJson) as User;
        set({ user, isAuthenticated: true, isRestoringSession: false });
        logger.auth('Session restored', user.email);
      } else {
        set({ isRestoringSession: false });
      }
    } catch {
      await clearTokens();
      set({ isRestoringSession: false });
      logger.error('Auth', 'Session restore failed');
    }
  },

  setUser: (user: User) => set({ user }),
}));

async function persistTokens(tokens: TokenPair) {
  await SecureStore.setItemAsync(STORAGE_KEYS.ACCESS_TOKEN, tokens.access_token);
  await SecureStore.setItemAsync(STORAGE_KEYS.REFRESH_TOKEN, tokens.refresh_token);
}

async function clearTokens() {
  await SecureStore.deleteItemAsync(STORAGE_KEYS.ACCESS_TOKEN);
  await SecureStore.deleteItemAsync(STORAGE_KEYS.REFRESH_TOKEN);
  await SecureStore.deleteItemAsync(STORAGE_KEYS.USER);
}
