export type Role = 'superadmin' | 'facility_admin' | 'facility_user' | 'reviewer';

export interface Facility {
  id: string;
  uid?: string;
  name: string;
  code: string;
  level: string;
  subcounty: string;
  hsd: string;
  district: string;
  client_code_prefix: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  email: string;
  full_name: string;
  role: Role;
  facility_id?: string;
  facility?: Facility;
  is_active: boolean;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface LoginResponse {
  tokens: TokenPair;
  user: User;
}

export interface OptionSetItem {
  id: string;
  category: string;
  code: string;
  label: string;
  description?: string;
  sort_order: number;
  is_active: boolean;
}

export type OptionSetsGrouped = Record<string, OptionSetItem[]>;

export interface FPRegistration {
  id: string;
  facility_id: string;
  facility?: Facility;
  created_by: string;
  creator?: User;
  visit_date: string;
  serial_number: number;
  client_number?: string;
  is_visitor: boolean;
  nin?: string;
  surname: string;
  given_name: string;
  phone_number?: string;
  village?: string;
  parish?: string;
  subcounty?: string;
  district?: string;
  sex: 'M' | 'F';
  age: number;
  is_new_user: boolean;
  is_revisit: boolean;
  previous_method?: string;
  hts_code?: string;
  counseling_individual: boolean;
  counseling_as_couple: boolean;
  counseling_om: boolean;
  counseling_se: boolean;
  counseling_wd: boolean;
  counseling_ms: boolean;
  is_switching: boolean;
  switching_reason?: string;
  pills_coc_cycles: number;
  pills_pop_cycles: number;
  pills_ecp_pieces: number;
  condoms_male_units: number;
  condoms_female_units: number;
  injectable_dmpa_im_doses: number;
  injectable_dmpa_sc_pa_doses: number;
  injectable_dmpa_sc_si_doses: number;
  implant_3_years: boolean;
  implant_5_years: boolean;
  iud_copper_t: boolean;
  iud_hormonal_3_years: boolean;
  iud_hormonal_5_years: boolean;
  tubal_ligation: boolean;
  vasectomy: boolean;
  fam_standard_days: boolean;
  fam_lam: boolean;
  fam_two_day: boolean;
  postpartum_fp_timing?: string;
  post_abortion_fp_timing?: string;
  implant_removal_reason?: string;
  implant_removal_timing?: string;
  iud_removal_reason?: string;
  iud_removal_timing?: string;
  side_effects?: string;
  cervical_screening_method?: string;
  cervical_cancer_status?: string;
  cervical_cancer_treatment?: string;
  breast_cancer_screening?: string;
  screened_for_sti?: boolean;
  referral_number?: string;
  referral_reason?: string;
  remarks?: string;
  created_at: string;
  updated_at: string;
}

export interface RegistrationInput {
  visit_date: string;
  is_visitor: boolean;
  nin: string;
  surname: string;
  given_name: string;
  phone_number: string;
  village: string;
  parish: string;
  subcounty: string;
  district: string;
  sex: string;
  age: number;
  is_new_user: boolean;
  is_revisit: boolean;
  previous_method: string;
  hts_code: string;
  counseling_individual: boolean;
  counseling_as_couple: boolean;
  counseling_om: boolean;
  counseling_se: boolean;
  counseling_wd: boolean;
  counseling_ms: boolean;
  is_switching: boolean;
  switching_reason: string;
  pills_coc_cycles: number;
  pills_pop_cycles: number;
  pills_ecp_pieces: number;
  condoms_male_units: number;
  condoms_female_units: number;
  injectable_dmpa_im_doses: number;
  injectable_dmpa_sc_pa_doses: number;
  injectable_dmpa_sc_si_doses: number;
  implant_3_years: boolean;
  implant_5_years: boolean;
  iud_copper_t: boolean;
  iud_hormonal_3_years: boolean;
  iud_hormonal_5_years: boolean;
  tubal_ligation: boolean;
  vasectomy: boolean;
  fam_standard_days: boolean;
  fam_lam: boolean;
  fam_two_day: boolean;
  postpartum_fp_timing: string;
  post_abortion_fp_timing: string;
  implant_removal_reason: string;
  implant_removal_timing: string;
  iud_removal_reason: string;
  iud_removal_timing: string;
  side_effects: string;
  cervical_screening_method: string;
  cervical_cancer_status: string;
  cervical_cancer_treatment: string;
  breast_cancer_screening: string;
  screened_for_sti: boolean | null;
  referral_number: string;
  referral_reason: string;
  remarks: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  meta?: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface ApiError {
  success: false;
  message: string;
  errors?: Array<{ field?: string; message: string }>;
}
