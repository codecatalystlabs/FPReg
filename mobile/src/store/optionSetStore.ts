import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { optionSetsApi } from '../api/optionSets';
import { STORAGE_KEYS } from '../constants';
import { logger } from '../utils/logger';
import type { OptionSetsGrouped, OptionSetItem } from '../types';

interface OptionSetState {
  sets: OptionSetsGrouped;
  isLoaded: boolean;
  isLoading: boolean;

  fetchAll: () => Promise<void>;
  getCategory: (category: string) => OptionSetItem[];
  getLabelByCode: (category: string, code: string) => string;
}

export const useOptionSetStore = create<OptionSetState>((set, get) => ({
  sets: {},
  isLoaded: false,
  isLoading: false,

  fetchAll: async () => {
    if (get().isLoading) return;
    set({ isLoading: true });

    try {
      const data = await optionSetsApi.getAllGrouped();
      set({ sets: data, isLoaded: true, isLoading: false });
      await AsyncStorage.setItem(STORAGE_KEYS.OPTION_SETS, JSON.stringify(data));
      logger.info('OptionSets', 'Fetched and cached');
    } catch {
      // Fall back to cached version
      try {
        const cached = await AsyncStorage.getItem(STORAGE_KEYS.OPTION_SETS);
        if (cached) {
          set({ sets: JSON.parse(cached), isLoaded: true, isLoading: false });
          logger.info('OptionSets', 'Loaded from cache');
          return;
        }
      } catch {}
      set({ isLoading: false });
      logger.error('OptionSets', 'Failed to load option sets');
    }
  },

  getCategory: (category: string) => get().sets[category] || [],

  getLabelByCode: (category: string, code: string) => {
    const items = get().sets[category] || [];
    return items.find((i) => i.code === code)?.label || code;
  },
}));
