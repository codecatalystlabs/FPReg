import api from './client';
import type {
  ApiResponse,
  FPRegistration,
  PaginationMeta,
  RegistrationInput,
} from '../types';

export interface RegistrationListParams {
  page?: number;
  per_page?: number;
  search?: string;
  visit_date?: string;
  date_from?: string;
  date_to?: string;
  sex?: string;
}

export const registrationsApi = {
  async list(params: RegistrationListParams = {}): Promise<{
    items: FPRegistration[];
    meta: PaginationMeta;
  }> {
    const { data } = await api.get<ApiResponse<FPRegistration[]>>('/registrations', { params });
    return {
      items: data.data || [],
      meta: data.meta || { page: 1, per_page: 25, total: 0, total_pages: 0 },
    };
  },

  async getById(id: string): Promise<FPRegistration> {
    const { data } = await api.get<ApiResponse<FPRegistration>>(`/registrations/${id}`);
    return data.data!;
  },

  async create(input: RegistrationInput): Promise<FPRegistration> {
    const { data } = await api.post<ApiResponse<FPRegistration>>('/registrations', input);
    return data.data!;
  },

  async update(id: string, input: RegistrationInput): Promise<FPRegistration> {
    const { data } = await api.put<ApiResponse<FPRegistration>>(`/registrations/${id}`, input);
    return data.data!;
  },

  async remove(id: string): Promise<void> {
    await api.delete(`/registrations/${id}`);
  },
};
