import React, { useState, useEffect, useMemo } from 'react';
import {
  ScrollView,
  View,
  Text,
  Alert,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { AppInput } from '../components/AppInput';
import { AppSelect } from '../components/AppSelect';
import { AppCheckbox } from '../components/AppCheckbox';
import { AppButton } from '../components/AppButton';
import { SectionHeader } from '../components/SectionHeader';
import { useOptionSetStore } from '../store/optionSetStore';
import { registrationsApi } from '../api/registrations';
import { todayISO } from '../utils/format';
import { logger } from '../utils/logger';
import { colors, spacing, radii, shadows } from '../theme';
import type { RegistrationInput, OptionSetItem } from '../types';
import { AxiosError } from 'axios';

const schema = z.object({
  visit_date: z.string().min(1, 'Visit date is required'),
  surname: z.string().min(1, 'Surname is required'),
  given_name: z.string().min(1, 'Given name is required'),
  sex: z.string().min(1, 'Sex is required'),
  age: z.number().min(0).max(120, 'Age must be 0–120'),
  is_new_user: z.boolean(),
  is_revisit: z.boolean(),
}).refine((d) => d.is_new_user !== d.is_revisit, {
  message: 'Must be either new user or revisit',
  path: ['is_new_user'],
});

type FormData = z.infer<typeof schema>;

const defaultValues: RegistrationInput = {
  visit_date: todayISO(), is_visitor: false, nin: '', surname: '', given_name: '',
  phone_number: '', village: '', parish: '', subcounty: '', district: '',
  sex: '', age: 0, is_new_user: false, is_revisit: false, previous_method: '',
  hts_code: '', counseling_individual: false, counseling_as_couple: false,
  counseling_om: false, counseling_se: false, counseling_wd: false, counseling_ms: false,
  is_switching: false, switching_reason: '',
  pills_coc_cycles: 0, pills_pop_cycles: 0, pills_ecp_pieces: 0,
  condoms_male_units: 0, condoms_female_units: 0,
  injectable_dmpa_im_doses: 0, injectable_dmpa_sc_pa_doses: 0, injectable_dmpa_sc_si_doses: 0,
  implant_3_years: false, implant_5_years: false,
  iud_copper_t: false, iud_hormonal_3_years: false, iud_hormonal_5_years: false,
  tubal_ligation: false, vasectomy: false,
  fam_standard_days: false, fam_lam: false, fam_two_day: false,
  postpartum_fp_timing: '', post_abortion_fp_timing: '',
  implant_removal_reason: '', implant_removal_timing: '',
  iud_removal_reason: '', iud_removal_timing: '',
  side_effects: '', cervical_screening_method: '', cervical_cancer_status: '',
  cervical_cancer_treatment: '', breast_cancer_screening: '',
  screened_for_sti: null, referral_number: '', referral_reason: '', remarks: '',
};

function toOptions(items: OptionSetItem[]) {
  return items.map((i) => ({ value: i.code, label: `${i.code} – ${i.label}` }));
}

export function NewRegistrationScreen() {
  const nav = useNavigation();
  const sets = useOptionSetStore((s) => s.sets);
  const fetchSets = useOptionSetStore((s) => s.fetchAll);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState<RegistrationInput>({ ...defaultValues });

  useEffect(() => { fetchSets(); }, []);

  const { control, handleSubmit, watch, setValue, formState: { errors } } = useForm<RegistrationInput>({
    defaultValues,
    resolver: zodResolver(schema) as any,
  });

  const watchSex = watch('sex');
  const watchRevisit = watch('is_revisit');
  const watchNewUser = watch('is_new_user');
  const watchSwitching = watch('is_switching');
  const watchCervStatus = watch('cervical_cancer_status');
  const watchImplRemoval = watch('implant_removal_reason');
  const watchIudRemoval = watch('iud_removal_reason');

  const opts = useMemo(() => ({
    hts: toOptions(sets.hts_code || []),
    fpMethod: toOptions(sets.fp_method || []),
    switchReason: toOptions(sets.switching_reason || []),
    ppTiming: toOptions(sets.postpartum_fp_timing || []),
    paTiming: toOptions(sets.post_abortion_fp_timing || []),
    larcReason: toOptions(sets.larc_removal_reason || []),
    larcTiming: toOptions(sets.larc_removal_timing || []),
    cervMethod: toOptions(sets.cervical_screening_method || []),
    cervStatus: toOptions(sets.cervical_cancer_status || []),
    cervTreat: toOptions(sets.cervical_cancer_treatment || []),
    breastScreen: toOptions(sets.breast_cancer_screening || []),
    sideEffects: sets.side_effect || [],
  }), [sets]);

  const onSubmit = async (data: RegistrationInput) => {
    setSubmitting(true);
    try {
      const result = await registrationsApi.create(data);
      Alert.alert('Saved', `Client #: ${result.client_number || 'Visitor'}`, [
        { text: 'OK', onPress: () => nav.goBack() },
      ]);
    } catch (e) {
      const msg = e instanceof AxiosError ? e.response?.data?.message || 'Submission failed' : 'Submission failed';
      Alert.alert('Error', msg);
      logger.error('Registration', 'Create failed', e);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView style={styles.container} contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">

        {/* Visit Info */}
        <SectionHeader title="Visit Information" icon="calendar" />
        <View style={styles.section}>
          <Controller control={control} name="visit_date" render={({ field }) => (
            <AppInput label="Visit Date" value={field.value} onChangeText={field.onChange} placeholder="YYYY-MM-DD" error={errors.visit_date?.message} required />
          )} />
          <Controller control={control} name="is_visitor" render={({ field }) => (
            <AppCheckbox label="Visitor (no client number)" value={field.value} onChange={field.onChange} helpText="Tick if client will not return for refills" />
          )} />
        </View>

        {/* Client Info */}
        <SectionHeader title="Client Information" icon="person" subtitle="Columns 2–7" />
        <View style={styles.section}>
          <Controller control={control} name="nin" render={({ field }) => (
            <AppInput label="NIN" value={field.value} onChangeText={field.onChange} placeholder="National ID" />
          )} />
          <Controller control={control} name="surname" render={({ field }) => (
            <AppInput label="Surname" value={field.value} onChangeText={field.onChange} error={errors.surname?.message} required />
          )} />
          <Controller control={control} name="given_name" render={({ field }) => (
            <AppInput label="Given Name" value={field.value} onChangeText={field.onChange} error={errors.given_name?.message} required />
          )} />
          <Controller control={control} name="phone_number" render={({ field }) => (
            <AppInput label="Phone" value={field.value} onChangeText={field.onChange} keyboardType="phone-pad" />
          )} />
          <Controller control={control} name="sex" render={({ field }) => (
            <AppSelect label="Sex" options={[{ value: 'F', label: 'Female' }, { value: 'M', label: 'Male' }]} value={field.value} onChange={field.onChange} error={errors.sex?.message} required />
          )} />
          <Controller control={control} name="age" render={({ field }) => (
            <AppInput label="Age (years)" value={field.value ? String(field.value) : ''} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" error={errors.age?.message} required />
          )} />
          <Controller control={control} name="village" render={({ field }) => (
            <AppInput label="Village/Cell/Zone" value={field.value} onChangeText={field.onChange} />
          )} />
          <Controller control={control} name="parish" render={({ field }) => (
            <AppInput label="Parish/Ward" value={field.value} onChangeText={field.onChange} />
          )} />
          <Controller control={control} name="subcounty" render={({ field }) => (
            <AppInput label="Subcounty/Division" value={field.value} onChangeText={field.onChange} />
          )} />
          <Controller control={control} name="district" render={({ field }) => (
            <AppInput label="District/City" value={field.value} onChangeText={field.onChange} />
          )} />
        </View>

        {/* Visit Type */}
        <SectionHeader title="Visit Type" icon="clipboard" subtitle="Columns 8–11" />
        <View style={styles.section}>
          <Controller control={control} name="is_new_user" render={({ field }) => (
            <AppCheckbox label="New User (1st time FP)" value={field.value} onChange={(v) => { field.onChange(v); if (v) setValue('is_revisit', false); }} />
          )} />
          <Controller control={control} name="is_revisit" render={({ field }) => (
            <AppCheckbox label="Revisit" value={field.value} onChange={(v) => { field.onChange(v); if (v) setValue('is_new_user', false); }} />
          )} />
          {errors.is_new_user && <Text style={styles.formError}>{errors.is_new_user.message}</Text>}
          {watchRevisit && (
            <Controller control={control} name="previous_method" render={({ field }) => (
              <AppSelect label="Previous Method" options={opts.fpMethod} value={field.value} onChange={field.onChange} required />
            )} />
          )}
          <Controller control={control} name="hts_code" render={({ field }) => (
            <AppSelect label="HTS Code" options={opts.hts} value={field.value} onChange={field.onChange} />
          )} />
        </View>

        {/* Counseling */}
        <SectionHeader title="FP Counseling" icon="chatbubbles" subtitle="Columns 12–13" />
        <View style={styles.section}>
          <Controller control={control} name="counseling_individual" render={({ field }) => (
            <AppCheckbox label="Individually" value={field.value} onChange={field.onChange} />
          )} />
          <Controller control={control} name="counseling_as_couple" render={({ field }) => (
            <AppCheckbox label="As Couple" value={field.value} onChange={field.onChange} />
          )} />
          <Text style={styles.subLabel}>Topics Counseled</Text>
          <View style={styles.checkRow}>
            <Controller control={control} name="counseling_om" render={({ field }) => (<AppCheckbox label="OM" value={field.value} onChange={field.onChange} />)} />
            <Controller control={control} name="counseling_se" render={({ field }) => (<AppCheckbox label="SE" value={field.value} onChange={field.onChange} />)} />
            <Controller control={control} name="counseling_wd" render={({ field }) => (<AppCheckbox label="WD" value={field.value} onChange={field.onChange} />)} />
            <Controller control={control} name="counseling_ms" render={({ field }) => (<AppCheckbox label="MS" value={field.value} onChange={field.onChange} />)} />
          </View>
          <Controller control={control} name="is_switching" render={({ field }) => (
            <AppCheckbox label="Switching Method" value={field.value} onChange={field.onChange} />
          )} />
          {watchSwitching && (
            <Controller control={control} name="switching_reason" render={({ field }) => (
              <AppSelect label="Reason for Switching" options={opts.switchReason} value={field.value} onChange={field.onChange} required />
            )} />
          )}
        </View>

        {/* Contraceptives */}
        <SectionHeader title="Contraceptives" icon="medkit" subtitle="Columns 14–20" />
        <View style={styles.section}>
          <Text style={styles.subLabel}>Oral Pills (Cycles)</Text>
          <Controller control={control} name="pills_coc_cycles" render={({ field }) => (
            <AppInput label="CoCs" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" />
          )} />
          <Controller control={control} name="pills_pop_cycles" render={({ field }) => (
            <AppInput label="POP" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" />
          )} />
          <Controller control={control} name="pills_ecp_pieces" render={({ field }) => (
            <AppInput label="ECP (pieces)" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" />
          )} />
          <Text style={styles.subLabel}>Condoms (Units)</Text>
          <Controller control={control} name="condoms_male_units" render={({ field }) => (
            <AppInput label="Male" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" />
          )} />
          <Controller control={control} name="condoms_female_units" render={({ field }) => (
            <AppInput label="Female" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" />
          )} />
          <Text style={styles.subLabel}>Injectables (Doses)</Text>
          <Controller control={control} name="injectable_dmpa_im_doses" render={({ field }) => (
            <AppInput label="DMPA-IM" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" />
          )} />
          <Controller control={control} name="injectable_dmpa_sc_pa_doses" render={({ field }) => (
            <AppInput label="DMPA-SC PA" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" helpText="Provider Administered" />
          )} />
          <Controller control={control} name="injectable_dmpa_sc_si_doses" render={({ field }) => (
            <AppInput label="DMPA-SC SI" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" helpText="Self-Injected" />
          )} />
          <Text style={styles.subLabel}>Implants & IUDs</Text>
          <Controller control={control} name="implant_3_years" render={({ field }) => (<AppCheckbox label="Implant 3 Years" value={field.value} onChange={field.onChange} />)} />
          <Controller control={control} name="implant_5_years" render={({ field }) => (<AppCheckbox label="Implant 5 Years" value={field.value} onChange={field.onChange} />)} />
          <Controller control={control} name="iud_copper_t" render={({ field }) => (<AppCheckbox label="IUD Copper-T" value={field.value} onChange={field.onChange} />)} />
          <Controller control={control} name="iud_hormonal_3_years" render={({ field }) => (<AppCheckbox label="IUD Hormonal 3 Yr" value={field.value} onChange={field.onChange} />)} />
          <Controller control={control} name="iud_hormonal_5_years" render={({ field }) => (<AppCheckbox label="IUD Hormonal 5 Yr" value={field.value} onChange={field.onChange} />)} />
          <Text style={styles.subLabel}>Permanent & FAM</Text>
          {watchSex !== 'M' && <Controller control={control} name="tubal_ligation" render={({ field }) => (<AppCheckbox label="Tubal Ligation" value={field.value} onChange={field.onChange} />)} />}
          {watchSex !== 'F' && <Controller control={control} name="vasectomy" render={({ field }) => (<AppCheckbox label="Vasectomy" value={field.value} onChange={field.onChange} />)} />}
          <Controller control={control} name="fam_standard_days" render={({ field }) => (<AppCheckbox label="Standard Days" value={field.value} onChange={field.onChange} />)} />
          <Controller control={control} name="fam_lam" render={({ field }) => (<AppCheckbox label="LAM" value={field.value} onChange={field.onChange} />)} />
          <Controller control={control} name="fam_two_day" render={({ field }) => (<AppCheckbox label="Two Day Method" value={field.value} onChange={field.onChange} />)} />
        </View>

        {/* Post-pregnancy & LARC */}
        <SectionHeader title="Post-Pregnancy & LARC" icon="heart" subtitle="Columns 21–22" />
        <View style={styles.section}>
          <Controller control={control} name="postpartum_fp_timing" render={({ field }) => (
            <AppSelect label="Postpartum FP Timing" options={opts.ppTiming} value={field.value} onChange={field.onChange} />
          )} />
          <Controller control={control} name="post_abortion_fp_timing" render={({ field }) => (
            <AppSelect label="Post-Abortion FP Timing" options={opts.paTiming} value={field.value} onChange={field.onChange} />
          )} />
          <Controller control={control} name="implant_removal_reason" render={({ field }) => (
            <AppSelect label="Implant Removal Reason" options={opts.larcReason} value={field.value} onChange={field.onChange} />
          )} />
          {!!watchImplRemoval && (
            <Controller control={control} name="implant_removal_timing" render={({ field }) => (
              <AppSelect label="Implant Removal Timing" options={opts.larcTiming} value={field.value} onChange={field.onChange} />
            )} />
          )}
          <Controller control={control} name="iud_removal_reason" render={({ field }) => (
            <AppSelect label="IUD Removal Reason" options={opts.larcReason} value={field.value} onChange={field.onChange} />
          )} />
          {!!watchIudRemoval && (
            <Controller control={control} name="iud_removal_timing" render={({ field }) => (
              <AppSelect label="IUD Removal Timing" options={opts.larcTiming} value={field.value} onChange={field.onChange} />
            )} />
          )}
        </View>

        {/* Side Effects & Other Services */}
        <SectionHeader title="Side Effects & Services" icon="alert-circle" subtitle="Columns 23–25" />
        <View style={styles.section}>
          <Text style={styles.subLabel}>Side Effects</Text>
          <Controller control={control} name="side_effects" render={({ field }) => {
            const selected = field.value ? field.value.split(',').filter(Boolean) : [];
            return (
              <View style={styles.checkGrid}>
                {opts.sideEffects.map((se) => (
                  <AppCheckbox
                    key={se.code}
                    label={se.code}
                    value={selected.includes(se.code)}
                    onChange={(v) => {
                      const next = v ? [...selected, se.code] : selected.filter((c) => c !== se.code);
                      field.onChange(next.join(','));
                    }}
                  />
                ))}
              </View>
            );
          }} />
          {watchSex === 'F' && (
            <>
              <Controller control={control} name="cervical_screening_method" render={({ field }) => (
                <AppSelect label="Cervical Screening Method" options={opts.cervMethod} value={field.value} onChange={field.onChange} />
              )} />
              <Controller control={control} name="cervical_cancer_status" render={({ field }) => (
                <AppSelect label="Cervical Cancer Status" options={opts.cervStatus} value={field.value} onChange={field.onChange} />
              )} />
              {watchCervStatus === '2' && (
                <Controller control={control} name="cervical_cancer_treatment" render={({ field }) => (
                  <AppSelect label="Cervical Treatment" options={opts.cervTreat} value={field.value} onChange={field.onChange} />
                )} />
              )}
              <Controller control={control} name="breast_cancer_screening" render={({ field }) => (
                <AppSelect label="Breast Cancer Screening" options={opts.breastScreen} value={field.value} onChange={field.onChange} />
              )} />
            </>
          )}
          <Controller control={control} name="screened_for_sti" render={({ field }) => (
            <AppSelect
              label="Screened for STIs"
              options={[{ value: 'true', label: 'Yes' }, { value: 'false', label: 'No' }]}
              value={field.value === true ? 'true' : field.value === false ? 'false' : ''}
              onChange={(v) => field.onChange(v === 'true' ? true : v === 'false' ? false : null)}
            />
          )} />
        </View>

        {/* Referral */}
        <SectionHeader title="Referral & Remarks" icon="share" subtitle="Columns 26–27" />
        <View style={styles.section}>
          <Controller control={control} name="referral_number" render={({ field }) => (
            <AppInput label="Referral Number" value={field.value} onChangeText={field.onChange} />
          )} />
          <Controller control={control} name="referral_reason" render={({ field }) => (
            <AppInput label="Reason for Referral" value={field.value} onChangeText={field.onChange} />
          )} />
          <Controller control={control} name="remarks" render={({ field }) => (
            <AppInput label="Remarks" value={field.value} onChangeText={field.onChange} multiline numberOfLines={3} />
          )} />
        </View>

        <AppButton
          title="Save Registration"
          onPress={handleSubmit(onSubmit as any)}
          loading={submitting}
          size="lg"
          style={{ marginTop: spacing.lg, marginBottom: spacing.xxxl }}
        />

      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg },
  section: {
    backgroundColor: colors.surface,
    borderRadius: radii.lg,
    padding: spacing.lg,
    marginBottom: spacing.sm,
    ...shadows.sm,
  },
  subLabel: {
    fontSize: 13, fontWeight: '600', color: colors.primary,
    marginTop: spacing.md, marginBottom: spacing.sm,
  },
  checkRow: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.lg },
  checkGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
  formError: { fontSize: 12, color: colors.danger, marginBottom: spacing.sm },
});
