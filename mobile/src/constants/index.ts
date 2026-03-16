 export const API_BASE_URL = 'https://fpscore.health.go.ug/fpreg/api/v1';
//export const API_BASE_URL = 'http://10.66.166.32:8080/api/v1';

export const STORAGE_KEYS = {
  ACCESS_TOKEN: 'fpreg_access_token',
  REFRESH_TOKEN: 'fpreg_refresh_token',
  USER: 'fpreg_user',
  OPTION_SETS: 'fpreg_option_sets',
} as const;

export const ROLES = {
  SUPERADMIN: 'superadmin' as const,
  FACILITY_ADMIN: 'facility_admin' as const,
  FACILITY_USER: 'facility_user' as const,
  REVIEWER: 'reviewer' as const,
};

export const DEFAULT_PAGE_SIZE = 25;
