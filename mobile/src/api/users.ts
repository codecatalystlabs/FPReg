import api from './client';
import type { ApiResponse, PaginationMeta, User } from '../types';

export interface CreateUserPayload {
  email: string;
  password: string;
  full_name: string;
  role: string;
  facility_id?: string | null;
  district?: string;
}

export type UserListParams = {
  page?: number;
  per_page?: number;
  facility_id?: string;
};

export const usersApi = {
  async list(params: UserListParams = {}): Promise<{ items: User[]; meta: PaginationMeta }> {
    const { data } = await api.get<ApiResponse<User[]>>('/users', { params });
    return {
      items: data.data || [],
      meta: data.meta || { page: 1, per_page: 25, total: 0, total_pages: 0 },
    };
  },

  async create(payload: CreateUserPayload): Promise<User> {
    const { data } = await api.post<ApiResponse<User>>('/users', payload);
    return data.data!;
  },
};
