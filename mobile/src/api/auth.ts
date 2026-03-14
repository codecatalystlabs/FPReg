import api from './client';
import type { ApiResponse, LoginResponse, TokenPair, User } from '../types';

export const authApi = {
  async login(email: string, password: string): Promise<LoginResponse> {
    const { data } = await api.post<ApiResponse<LoginResponse>>('/auth/login', {
      email,
      password,
    });
    return data.data!;
  },

  async refresh(refreshToken: string): Promise<{ tokens: TokenPair }> {
    const { data } = await api.post<ApiResponse<{ tokens: TokenPair }>>('/auth/refresh', {
      refresh_token: refreshToken,
    });
    return data.data!;
  },

  async logout(refreshToken: string): Promise<void> {
    await api.post('/auth/logout', { refresh_token: refreshToken });
  },

  async me(): Promise<User> {
    const { data } = await api.get<ApiResponse<User>>('/auth/me');
    return data.data!;
  },
};
