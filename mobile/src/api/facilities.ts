import api from './client';
import type { ApiResponse, Facility } from '../types';

export const facilitiesApi = {
  /** Fetches all pages (API max per_page is capped server-side; we page until done). */
  async listAll(perPage = 200): Promise<Facility[]> {
    const out: Facility[] = [];
    let page = 1;
    let totalPages = 1;
    do {
      const { data } = await api.get<ApiResponse<Facility[]>>('/facilities', {
        params: { page, per_page: perPage },
      });
      const batch = data.data || [];
      out.push(...batch);
      const meta = data.meta;
      totalPages = meta?.total_pages ?? 1;
      page += 1;
    } while (page <= totalPages);
    return out;
  },

  async getById(id: string): Promise<Facility> {
    const { data } = await api.get<ApiResponse<Facility>>(`/facilities/${id}`);
    return data.data!;
  },

  /** Superadmin only — used when creating a district biostatistician account. */
  async listDistricts(): Promise<string[]> {
    const { data } = await api.get<ApiResponse<string[]>>('/facilities/districts');
    return data.data ?? [];
  },
};
