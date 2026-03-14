import React, { useEffect, useState } from 'react';
import { ScrollView, View, Text, StyleSheet, Alert } from 'react-native';
import { RouteProp, useRoute, useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { Ionicons } from '@expo/vector-icons';
import { AppCard } from '../components/AppCard';
import { AppButton } from '../components/AppButton';
import { StatusBadge } from '../components/StatusBadge';
import { LoadingState } from '../components/LoadingState';
import { ErrorState } from '../components/ErrorState';
import { PermissionGuard } from '../components/PermissionGuard';
import { registrationsApi } from '../api/registrations';
import { useOptionSetStore } from '../store/optionSetStore';
import { useAuthStore } from '../store/authStore';
import { canEditRegistration, canDeleteRegistration } from '../utils/permissions';
import { formatDate } from '../utils/format';
import { colors, spacing, typography, radii, shadows } from '../theme';
import type { FPRegistration } from '../types';
import type { MainStackParamList } from '../navigation/MainNavigator';

type RouteP = RouteProp<MainStackParamList, 'SubmissionDetail'>;
type NavProp = NativeStackNavigationProp<MainStackParamList>;

export function SubmissionDetailScreen() {
  const { params } = useRoute<RouteP>();
  const nav = useNavigation<NavProp>();
  const user = useAuthStore((s) => s.user);
  const getLabel = useOptionSetStore((s) => s.getLabelByCode);
  const [reg, setReg] = useState<FPRegistration | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  const load = async () => {
    setLoading(true);
    setError(false);
    try {
      const data = await registrationsApi.getById(params.id);
      setReg(data);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [params.id]);

  if (loading) return <LoadingState />;
  if (error || !reg) return <ErrorState onRetry={load} />;

  const check = (v: boolean) => v ? '✓' : '–';
  const val = (v?: string | number | null) => (v != null && v !== '') ? String(v) : '–';

  const handleDelete = () => {
    Alert.alert('Delete', 'Are you sure you want to delete this record?', [
      { text: 'Cancel', style: 'cancel' },
      {
        text: 'Delete',
        style: 'destructive',
        onPress: async () => {
          try {
            await registrationsApi.remove(reg.id);
            nav.goBack();
          } catch {
            Alert.alert('Error', 'Failed to delete registration');
          }
        },
      },
    ]);
  };

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* Header cards */}
      <View style={styles.headerRow}>
        <View style={[styles.headerCard, { borderLeftColor: colors.primary }]}>
          <Text style={styles.headerLabel}>CLIENT #</Text>
          <Text style={styles.headerValue}>{reg.client_number || 'Visitor'}</Text>
        </View>
        <View style={[styles.headerCard, { borderLeftColor: colors.accent }]}>
          <Text style={styles.headerLabel}>S/N</Text>
          <Text style={styles.headerValue}>{reg.serial_number}</Text>
        </View>
        <View style={[styles.headerCard, { borderLeftColor: colors.warning }]}>
          <Text style={styles.headerLabel}>DATE</Text>
          <Text style={styles.headerValue}>{reg.visit_date}</Text>
        </View>
      </View>

      <StatusBadge
        label={reg.is_new_user ? 'New User' : 'Revisit'}
        variant={reg.is_new_user ? 'success' : 'warning'}
      />

      {/* Client Info */}
      <SectionTitle title="Client Information" />
      <AppCard>
        <Row label="Name" value={`${reg.surname} ${reg.given_name}`} />
        <Row label="Sex" value={reg.sex} />
        <Row label="Age" value={`${reg.age} years`} />
        <Row label="NIN" value={val(reg.nin)} />
        <Row label="Phone" value={val(reg.phone_number)} />
        <Row label="Address" value={[reg.village, reg.parish, reg.subcounty, reg.district].filter(Boolean).join(', ') || '–'} />
      </AppCard>

      {/* Visit Type */}
      <SectionTitle title="Visit & Counseling" />
      <AppCard>
        <Row label="HTS" value={reg.hts_code ? getLabel('hts_code', reg.hts_code) : '–'} />
        <Row label="Previous Method" value={reg.previous_method ? getLabel('fp_method', reg.previous_method) : '–'} />
        <Row label="Counseling" value={[
          reg.counseling_individual && 'Individual',
          reg.counseling_as_couple && 'Couple',
        ].filter(Boolean).join(', ') || '–'} />
        <Row label="Topics" value={[
          reg.counseling_om && 'OM',
          reg.counseling_se && 'SE',
          reg.counseling_wd && 'WD',
          reg.counseling_ms && 'MS',
        ].filter(Boolean).join(', ') || '–'} />
        <Row label="Switching" value={reg.is_switching ? `Yes (${getLabel('switching_reason', reg.switching_reason || '')})` : 'No'} />
      </AppCard>

      {/* Contraceptives */}
      <SectionTitle title="Contraceptives Dispensed" />
      <AppCard>
        <Row label="CoC Cycles" value={val(reg.pills_coc_cycles)} />
        <Row label="POP Cycles" value={val(reg.pills_pop_cycles)} />
        <Row label="ECP Pieces" value={val(reg.pills_ecp_pieces)} />
        <Row label="Male Condoms" value={val(reg.condoms_male_units)} />
        <Row label="Female Condoms" value={val(reg.condoms_female_units)} />
        <Row label="DMPA-IM" value={val(reg.injectable_dmpa_im_doses)} />
        <Row label="DMPA-SC PA" value={val(reg.injectable_dmpa_sc_pa_doses)} />
        <Row label="DMPA-SC SI" value={val(reg.injectable_dmpa_sc_si_doses)} />
        <Row label="Implant 3yr" value={check(reg.implant_3_years)} />
        <Row label="Implant 5yr" value={check(reg.implant_5_years)} />
        <Row label="IUD Copper-T" value={check(reg.iud_copper_t)} />
        <Row label="IUD Hormonal 3yr" value={check(reg.iud_hormonal_3_years)} />
        <Row label="IUD Hormonal 5yr" value={check(reg.iud_hormonal_5_years)} />
        <Row label="Tubal Ligation" value={check(reg.tubal_ligation)} />
        <Row label="Vasectomy" value={check(reg.vasectomy)} />
        <Row label="FAM SDM" value={check(reg.fam_standard_days)} />
        <Row label="FAM LAM" value={check(reg.fam_lam)} />
        <Row label="FAM Two Day" value={check(reg.fam_two_day)} />
      </AppCard>

      {/* Other services */}
      <SectionTitle title="Other Services" />
      <AppCard>
        <Row label="Postpartum FP" value={reg.postpartum_fp_timing ? getLabel('postpartum_fp_timing', reg.postpartum_fp_timing) : '–'} />
        <Row label="Post-Abortion FP" value={reg.post_abortion_fp_timing ? getLabel('post_abortion_fp_timing', reg.post_abortion_fp_timing) : '–'} />
        <Row label="Side Effects" value={val(reg.side_effects)} />
        <Row label="Cervical Screen" value={reg.cervical_screening_method ? getLabel('cervical_screening_method', reg.cervical_screening_method) : '–'} />
        <Row label="Cervical Status" value={reg.cervical_cancer_status ? getLabel('cervical_cancer_status', reg.cervical_cancer_status) : '–'} />
        <Row label="Breast Screen" value={reg.breast_cancer_screening ? getLabel('breast_cancer_screening', reg.breast_cancer_screening) : '–'} />
        <Row label="STI Screened" value={reg.screened_for_sti === true ? 'Yes' : reg.screened_for_sti === false ? 'No' : '–'} />
        <Row label="Referral #" value={val(reg.referral_number)} />
        <Row label="Referral Reason" value={val(reg.referral_reason)} />
        <Row label="Remarks" value={val(reg.remarks)} />
      </AppCard>

      {/* Meta */}
      <Text style={styles.metaText}>
        Created: {formatDate(reg.created_at)} · Facility: {reg.facility?.name || '–'}
      </Text>

      {/* Actions */}
      <View style={styles.actions}>
        {user && canEditRegistration(user.role) && (
          <AppButton
            title="Edit"
            variant="ghost"
            onPress={() => nav.navigate('EditRegistration', { id: reg.id })}
            icon={<Ionicons name="create-outline" size={18} color={colors.primary} />}
            style={{ flex: 1 }}
          />
        )}
        {user && canDeleteRegistration(user.role) && (
          <AppButton
            title="Delete"
            variant="danger"
            onPress={handleDelete}
            icon={<Ionicons name="trash-outline" size={18} color={colors.textInverse} />}
            style={{ flex: 1 }}
          />
        )}
      </View>
    </ScrollView>
  );
}

function SectionTitle({ title }: { title: string }) {
  return <Text style={styles.sectionTitle}>{title}</Text>;
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={styles.rowValue}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  headerRow: { flexDirection: 'row', gap: spacing.sm, marginBottom: spacing.md },
  headerCard: {
    flex: 1,
    backgroundColor: colors.card,
    borderRadius: radii.sm,
    padding: spacing.md,
    borderLeftWidth: 4,
    ...shadows.sm,
  },
  headerLabel: { ...typography.caption, color: colors.textMuted, textTransform: 'uppercase', letterSpacing: 0.5 },
  headerValue: { ...typography.h4, color: colors.text, marginTop: 2 },
  sectionTitle: { ...typography.h4, color: colors.text, marginTop: spacing.xl, marginBottom: spacing.sm },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: spacing.sm,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.divider,
  },
  rowLabel: { ...typography.bodySmall, color: colors.textSecondary, flex: 1 },
  rowValue: { ...typography.bodySmall, color: colors.text, fontWeight: '500', flex: 1, textAlign: 'right' },
  metaText: { ...typography.caption, color: colors.textMuted, marginTop: spacing.xl, textAlign: 'center' },
  actions: { flexDirection: 'row', gap: spacing.sm, marginTop: spacing.xl },
});
