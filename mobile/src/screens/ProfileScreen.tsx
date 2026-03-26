import React from 'react';
import { View, Text, ScrollView, StyleSheet, Alert } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { Ionicons } from '@expo/vector-icons';
import { AppCard } from '../components/AppCard';
import { AppButton } from '../components/AppButton';
import { useAuthStore } from '../store/authStore';
import { canManageUsers } from '../utils/permissions';
import { colors, spacing, typography, radii } from '../theme';
import type { MainStackParamList } from '../navigation/MainNavigator';

type NavProp = NativeStackNavigationProp<MainStackParamList>;

export function ProfileScreen() {
  const { user, logout } = useAuthStore();
  const nav = useNavigation<NavProp>();

  const handleLogout = () => {
    Alert.alert('Sign Out', 'Are you sure you want to sign out?', [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Sign Out', style: 'destructive', onPress: logout },
    ]);
  };

  if (!user) return null;

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      {/* Avatar & Name */}
      <View style={styles.header}>
        <View style={styles.avatar}>
          <Ionicons name="person" size={36} color={colors.textInverse} />
        </View>
        <Text style={styles.name}>{user.full_name}</Text>
        <Text style={styles.email}>{user.email}</Text>
      </View>

      {/* Details */}
      <AppCard>
        <DetailRow icon="shield-checkmark" label="Role" value={user.role.replace('_', ' ')} />
        {user.role === 'district_biostatistician' && user.district ? (
          <DetailRow icon="map" label="Assigned district" value={user.district} />
        ) : (
          <DetailRow icon="business" label="Facility" value={user.facility?.name || 'All Facilities'} />
        )}
        {user.facility && (
          <>
            <DetailRow icon="code" label="Facility Code" value={user.facility.code} />
            <DetailRow icon="location" label="District" value={user.facility.district || '–'} />
            <DetailRow icon="layers" label="Level" value={user.facility.level || '–'} />
          </>
        )}
      </AppCard>

      {canManageUsers(user.role) && (
        <AppButton
          title="Manage users"
          variant="secondary"
          onPress={() => nav.navigate('Users')}
          style={{ marginTop: spacing.lg }}
          icon={<Ionicons name="people-outline" size={20} color={colors.primary} />}
        />
      )}

      <AppCard style={{ marginTop: spacing.lg }}>
        <DetailRow icon="information-circle" label="App Version" value="1.0.0" />
        <DetailRow icon="server" label="API" value="HMIS MCH 007 v1" />
      </AppCard>

      <AppButton
        title="Sign Out"
        variant="danger"
        onPress={handleLogout}
        size="lg"
        icon={<Ionicons name="log-out-outline" size={20} color={colors.textInverse} />}
        style={{ marginTop: spacing.xxl }}
      />
    </ScrollView>
  );
}

function DetailRow({
  icon,
  label,
  value,
}: {
  icon: keyof typeof Ionicons.glyphMap;
  label: string;
  value: string;
}) {
  return (
    <View style={rowStyles.row}>
      <View style={rowStyles.left}>
        <Ionicons name={icon} size={18} color={colors.textMuted} />
        <Text style={rowStyles.label}>{label}</Text>
      </View>
      <Text style={rowStyles.value}>{value}</Text>
    </View>
  );
}

const rowStyles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.divider,
  },
  left: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  label: { ...typography.body, color: colors.textSecondary },
  value: { ...typography.body, color: colors.text, fontWeight: '500', textTransform: 'capitalize' },
});

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg },
  header: { alignItems: 'center', marginBottom: spacing.xxl, paddingTop: spacing.lg },
  avatar: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.md,
  },
  name: { ...typography.h2, color: colors.text },
  email: { ...typography.bodySmall, color: colors.textMuted, marginTop: 2 },
});
