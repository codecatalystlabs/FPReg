import React, { useEffect, useMemo, useState } from 'react';
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
import { AppInput } from '../components/AppInput';
import { AppSelect } from '../components/AppSelect';
import { AppButton } from '../components/AppButton';
import { SectionHeader } from '../components/SectionHeader';
import { useAuthStore } from '../store/authStore';
import { usersApi } from '../api/users';
import { facilitiesApi } from '../api/facilities';
import { colors, spacing, typography, radii } from '../theme';
import type { Facility, Role } from '../types';
import { AxiosError } from 'axios';

function rolesForCreator(meRole: Role | undefined): { value: string; label: string }[] {
  if (meRole === 'superadmin') {
    return [
      { value: 'facility_user', label: 'Facility user' },
      { value: 'reviewer', label: 'Reviewer' },
      { value: 'facility_admin', label: 'Facility admin' },
      { value: 'district_biostatistician', label: 'District biostatistician' },
    ];
  }
  return [
    { value: 'facility_user', label: 'Facility user' },
    { value: 'reviewer', label: 'Reviewer' },
  ];
}

function needsFacility(role: string): boolean {
  return role === 'facility_user' || role === 'reviewer' || role === 'facility_admin';
}

export function CreateUserScreen() {
  const nav = useNavigation();
  const me = useAuthStore((s) => s.user);

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [fullName, setFullName] = useState('');
  const [role, setRole] = useState('facility_user');
  const [facilityId, setFacilityId] = useState('');
  const [district, setDistrict] = useState('');
  const [facilities, setFacilities] = useState<Facility[]>([]);
  const [districts, setDistricts] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);

  const roleOptions = useMemo(() => rolesForCreator(me?.role), [me?.role]);

  useEffect(() => {
    setRole(roleOptions[0]?.value || 'facility_user');
  }, [roleOptions]);

  useEffect(() => {
    (async () => {
      try {
        if (me?.role === 'superadmin') {
          const [f, d] = await Promise.all([facilitiesApi.listAll(), facilitiesApi.listDistricts()]);
          setFacilities(f);
          setDistricts(d);
        } else if (me?.role === 'district_biostatistician' || me?.role === 'facility_admin') {
          const f = await facilitiesApi.listAll();
          setFacilities(f);
        }
      } catch {
        Alert.alert('Error', 'Could not load facilities');
      }
    })();
  }, [me?.role]);

  useEffect(() => {
    if (role === 'district_biostatistician') {
      setFacilityId('');
    } else {
      setDistrict('');
      if (needsFacility(role) && me?.role === 'facility_admin' && me.facility_id) {
        setFacilityId(me.facility_id);
      }
    }
  }, [role, me?.role, me?.facility_id]);

  const showFacilityPicker =
    needsFacility(role) &&
    me?.role !== 'facility_admin' &&
    (me?.role === 'superadmin' || me?.role === 'district_biostatistician');

  const showDistrictPicker = role === 'district_biostatistician' && me?.role === 'superadmin';

  const onSubmit = async () => {
    const err =
      !email.trim() || !password || !fullName.trim()
        ? 'Email, password and full name are required'
        : password.length < 8
          ? 'Password must be at least 8 characters'
          : needsFacility(role) && !facilityId
            ? 'Select a facility'
            : role === 'district_biostatistician' && !district.trim()
              ? 'Select or enter a district'
              : '';
    if (err) {
      Alert.alert('Validation', err);
      return;
    }

    setSubmitting(true);
    try {
      const payload: Parameters<typeof usersApi.create>[0] = {
        email: email.trim(),
        password,
        full_name: fullName.trim(),
        role,
      };
      if (needsFacility(role)) {
        payload.facility_id = facilityId;
      }
      if (role === 'district_biostatistician') {
        payload.district = district.trim();
        payload.facility_id = null;
      }
      await usersApi.create(payload);
      Alert.alert('Success', 'User created', [{ text: 'OK', onPress: () => nav.goBack() }]);
    } catch (e) {
      const ax = e as AxiosError<{ message?: string; errors?: { message: string }[] }>;
      const msg =
        ax.response?.data?.errors?.map((x) => x.message).join('\n') ||
        ax.response?.data?.message ||
        'Could not create user';
      Alert.alert('Error', msg);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <KeyboardAvoidingView
      style={styles.flex}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <ScrollView contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">
        <SectionHeader title="Account" icon="person-add" subtitle="Creates an account in your allowed scope" />

        <AppInput label="Full name" value={fullName} onChangeText={setFullName} autoCapitalize="words" />
        <AppInput label="Email" value={email} onChangeText={setEmail} keyboardType="email-address" autoCapitalize="none" />
        <AppInput label="Password (min 8)" value={password} onChangeText={setPassword} secureTextEntry />

        <AppSelect
          label="Role"
          value={role}
          options={roleOptions.map((o) => ({ value: o.value, label: o.label }))}
          onChange={setRole}
          placeholder="Role"
        />

        {showDistrictPicker && (
          <AppSelect
            label="District"
            value={district}
            options={districts.map((d) => ({ value: d, label: d }))}
            onChange={setDistrict}
            placeholder="Select district…"
          />
        )}

        {showFacilityPicker && (
          <AppSelect
            label="Facility"
            value={facilityId}
            options={facilities.map((f) => ({ value: f.id, label: `${f.name} (${f.code})` }))}
            onChange={setFacilityId}
            placeholder="Select facility…"
          />
        )}

        {me?.role === 'facility_admin' && needsFacility(role) && me.facility && (
          <View style={styles.fixedFac}>
            <Text style={styles.fixedFacLabel}>Facility</Text>
            <Text style={styles.fixedFacVal}>{me.facility.name} ({me.facility.code})</Text>
          </View>
        )}

        <AppButton title="Create user" onPress={onSubmit} loading={submitting} style={{ marginTop: spacing.xl }} />
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxxl },
  fixedFac: {
    backgroundColor: colors.surface,
    borderRadius: radii.md,
    padding: spacing.md,
    marginBottom: spacing.md,
    borderWidth: 1,
    borderColor: colors.border,
  },
  fixedFacLabel: { ...typography.caption, color: colors.textMuted, marginBottom: 4 },
  fixedFacVal: { ...typography.body, color: colors.text, fontWeight: '600' },
});
