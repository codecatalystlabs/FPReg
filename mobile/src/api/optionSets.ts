import api from './client';
import type { ApiResponse, OptionSetsGrouped } from '../types';

export const optionSetsApi = {
  async getAllGrouped(): Promise<OptionSetsGrouped> {
    const { data } = await api.get<ApiResponse<OptionSetsGrouped>>('/option-sets');
    return data.data || {};
  },

  async getCategories(): Promise<string[]> {
    const { data } = await api.get<ApiResponse<string[]>>('/option-sets/categories');
    return data.data || [];
  },
};
