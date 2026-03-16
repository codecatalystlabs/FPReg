import React, { useEffect, useState, useMemo } from 'react';
import {
  ScrollView,
  View,
  Text,
  Alert,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { RouteProp, useRoute, useNavigation } from '@react-navigation/native';
import { useForm, Controller } from 'react-hook-form';
import { AppInput } from '../components/AppInput';
import { AppSelect } from '../components/AppSelect';
import { AppCheckbox } from '../components/AppCheckbox';
import { AppRadioGroup } from '../components/AppRadioGroup';
import { AppButton } from '../components/AppButton';
import { SectionHeader } from '../components/SectionHeader';
import { LoadingState } from '../components/LoadingState';
import { ErrorState } from '../components/ErrorState';
import { useOptionSetStore } from '../store/optionSetStore';
import { registrationsApi } from '../api/registrations';
import { logger } from '../utils/logger';
import { colors, spacing, radii, shadows } from '../theme';
import type { FPRegistration, RegistrationInput, OptionSetItem } from '../types';
import type { MainStackParamList } from '../navigation/MainNavigator';
import { AxiosError } from 'axios';

type RouteP = RouteProp<MainStackParamList, 'EditRegistration'>;

function toOptions(items: OptionSetItem[]) {
  if (!Array.isArray(items)) return [];
  return items
    .filter((i) => i != null && (i.code != null && i.code !== ''))
    .map((i) => ({
      value: String(i.code),
      label: `${i.code} – ${i.label ?? i.code}`,
    }));
}

function regToInput(reg: FPRegistration): RegistrationInput {
  return {
    visit_date: reg.visit_date, is_visitor: reg.is_visitor, nin: reg.nin || '',
    surname: reg.surname, given_name: reg.given_name, phone_number: reg.phone_number || '',
    village: reg.village || '', parish: reg.parish || '', subcounty: reg.subcounty || '',
    district: reg.district || '', sex: reg.sex, age: reg.age,
    is_new_user: reg.is_new_user, is_revisit: reg.is_revisit,
    previous_method: reg.previous_method || '', hts_code: reg.hts_code || '',
    counseling_individual: reg.counseling_individual, counseling_as_couple: reg.counseling_as_couple,
    counseling_om: reg.counseling_om, counseling_se: reg.counseling_se,
    counseling_wd: reg.counseling_wd, counseling_ms: reg.counseling_ms,
    is_switching: reg.is_switching, switching_reason: reg.switching_reason || '',
    pills_coc_cycles: reg.pills_coc_cycles, pills_pop_cycles: reg.pills_pop_cycles,
    pills_ecp_pieces: reg.pills_ecp_pieces, condoms_male_units: reg.condoms_male_units,
    condoms_female_units: reg.condoms_female_units,
    injectable_dmpa_im_doses: reg.injectable_dmpa_im_doses,
    injectable_dmpa_sc_pa_doses: reg.injectable_dmpa_sc_pa_doses,
    injectable_dmpa_sc_si_doses: reg.injectable_dmpa_sc_si_doses,
    implant_3_years: reg.implant_3_years, implant_5_years: reg.implant_5_years,
    iud_copper_t: reg.iud_copper_t, iud_hormonal_3_years: reg.iud_hormonal_3_years,
    iud_hormonal_5_years: reg.iud_hormonal_5_years,
    tubal_ligation: reg.tubal_ligation, vasectomy: reg.vasectomy,
    fam_standard_days: reg.fam_standard_days, fam_lam: reg.fam_lam, fam_two_day: reg.fam_two_day,
    postpartum_fp_timing: reg.postpartum_fp_timing || '',
    post_abortion_fp_timing: reg.post_abortion_fp_timing || '',
    implant_removal_reason: reg.implant_removal_reason || '',
    implant_removal_timing: reg.implant_removal_timing || '',
    iud_removal_reason: reg.iud_removal_reason || '',
    iud_removal_timing: reg.iud_removal_timing || '',
    side_effects: reg.side_effects || '',
    cervical_screening_method: reg.cervical_screening_method || '',
    cervical_cancer_status: reg.cervical_cancer_status || '',
    cervical_cancer_treatment: reg.cervical_cancer_treatment || '',
    breast_cancer_screening: reg.breast_cancer_screening || '',
    screened_for_sti: reg.screened_for_sti ?? null,
    referral_number: reg.referral_number || '', referral_reason: reg.referral_reason || '',
    remarks: reg.remarks || '',
  };
}

export function EditRegistrationScreen() {
  const { params } = useRoute<RouteP>();
  const nav = useNavigation();
  const sets = useOptionSetStore((s) => s.sets);
  const [reg, setReg] = useState<FPRegistration | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const data = await registrationsApi.getById(params.id);
        setReg(data);
      } catch {
        setError(true);
      } finally {
        setLoading(false);
      }
    })();
  }, [params.id]);

  if (loading) return <LoadingState />;
  if (error || !reg) return <ErrorState message="Failed to load registration" />;

  return <EditForm reg={reg} sets={sets} submitting={submitting} setSubmitting={setSubmitting} nav={nav} />;
}

function EditForm({
  reg,
  sets,
  submitting,
  setSubmitting,
  nav,
}: {
  reg: FPRegistration;
  sets: any;
  submitting: boolean;
  setSubmitting: (v: boolean) => void;
  nav: any;
}) {
  const { control, handleSubmit, watch, setValue, formState: { errors } } = useForm<RegistrationInput>({
    defaultValues: regToInput(reg),
  });

  const watchSex = watch('sex');
  const watchRevisit = watch('is_revisit');
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
      await registrationsApi.update(reg.id, data);
      Alert.alert('Updated', 'Registration has been updated', [
        { text: 'OK', onPress: () => nav.goBack() },
      ]);
    } catch (e) {
      const msg = e instanceof AxiosError ? e.response?.data?.message || 'Update failed' : 'Update failed';
      Alert.alert('Error', msg);
      logger.error('Registration', 'Update failed', e);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView style={styles.container} contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">
        <SectionHeader title="Client Information" icon="person" />
        <View style={styles.section}>
          <Controller control={control} name="surname" render={({ field }) => (
            <AppInput label="Surname" value={field.value} onChangeText={field.onChange} required />
          )} />
          <Controller control={control} name="given_name" render={({ field }) => (
            <AppInput label="Given Name" value={field.value} onChangeText={field.onChange} required />
          )} />
          <Controller control={control} name="sex" render={({ field }) => (
            <AppSelect label="Sex" options={[{ value: 'F', label: 'Female' }, { value: 'M', label: 'Male' }]} value={field.value} onChange={field.onChange} required />
          )} />
          <Controller control={control} name="age" render={({ field }) => (
            <AppInput label="Age" value={String(field.value)} onChangeText={(v) => field.onChange(parseInt(v) || 0)} keyboardType="number-pad" required />
          )} />
          <Controller control={control} name="phone_number" render={({ field }) => (
            <AppInput label="Phone" value={field.value} onChangeText={field.onChange} keyboardType="phone-pad" />
          )} />
        </View>

        <SectionHeader title="Visit Type" icon="clipboard" />
        <View style={styles.section}>
          <Controller control={control} name="is_new_user" render={({ field }) => (
            <AppCheckbox label="New User" value={field.value} onChange={(v) => { field.onChange(v); if (v) setValue('is_revisit', false); }} />
          )} />
          <Controller control={control} name="is_revisit" render={({ field }) => (
            <AppCheckbox label="Revisit" value={field.value} onChange={(v) => { field.onChange(v); if (v) setValue('is_new_user', false); }} />
          )} />
          {watchRevisit && (
            <Controller control={control} name="previous_method" render={({ field }) => (
              <AppSelect label="Previous Method" options={opts.fpMethod} value={field.value} onChange={field.onChange} />
            )} />
          )}
          <Controller control={control} name="hts_code" render={({ field }) => (
            <AppSelect label="HTS Code" options={opts.hts} value={field.value} onChange={field.onChange} />
          )} />
        </View>

        <SectionHeader title="Counseling" icon="chatbubbles" />
        <View style={styles.section}>
          <Controller
            control={control}
            name="counseling_individual"
            render={({ field: ind }) => (
              <Controller
                control={control}
                name="counseling_as_couple"
                render={({ field: cou }) => (
                  <AppRadioGroup
                    label="Counseled"
                    options={[{ value: 'individual', label: 'Individual' }, { value: 'couple', label: 'As Couple' }]}
                    value={ind.value ? 'individual' : cou.value ? 'couple' : ''}
                    onChange={(v) => {
                      ind.onChange(v === 'individual');
                      cou.onChange(v === 'couple');
                    }}
                  />
                )}
              />
            )}
          />
          <Controller control={control} name="is_switching" render={({ field }) => (
            <AppCheckbox label="Switching" value={field.value} onChange={field.onChange} />
          )} />
          {watchSwitching && (
            <Controller control={control} name="switching_reason" render={({ field }) => (
              <AppSelect label="Switch Reason" options={opts.switchReason} value={field.value} onChange={field.onChange} />
            )} />
          )}
        </View>

        <SectionHeader title="Remarks" icon="document-text" />
        <View style={styles.section}>
          <Controller control={control} name="referral_number" render={({ field }) => (
            <AppInput label="Referral #" value={field.value} onChangeText={field.onChange} />
          )} />
          <Controller control={control} name="remarks" render={({ field }) => (
            <AppInput label="Remarks" value={field.value} onChangeText={field.onChange} multiline numberOfLines={3} />
          )} />
        </View>

        <AppButton
          title="Update Registration"
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
});
