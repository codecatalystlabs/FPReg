import api from './client';
import type { ApiResponse, Facility } from '../types';

export const facilitiesApi = {
  async list(): Promise<Facility[]> {
    const { data } = await api.get<ApiResponse<Facility[]>>('/facilities', {
      params: { per_page: 100 },
    });
    return data.data || [];
  },

  async getById(id: string): Promise<Facility> {
    const { data } = await api.get<ApiResponse<Facility>>(`/facilities/${id}`);
    return data.data!;
  },
};
