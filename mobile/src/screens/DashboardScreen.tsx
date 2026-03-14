import React, { useCallback, useState } from 'react';
import { View, Text, ScrollView, RefreshControl, TouchableOpacity, StyleSheet } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { Ionicons } from '@expo/vector-icons';
import { AppCard } from '../components/AppCard';
import { AppButton } from '../components/AppButton';
import { SubmissionListItem } from '../components/SubmissionListItem';
import { EmptyState } from '../components/EmptyState';
import { useAuthStore } from '../store/authStore';
import { useOptionSetStore } from '../store/optionSetStore';
import { registrationsApi } from '../api/registrations';
import { canCreateRegistration } from '../utils/permissions';
import { todayISO } from '../utils/format';
import { colors, spacing, typography, radii, shadows } from '../theme';
import type { FPRegistration } from '../types';
import type { MainStackParamList } from '../navigation/MainNavigator';

type NavProp = NativeStackNavigationProp<MainStackParamList>;

export function DashboardScreen() {
  const nav = useNavigation<NavProp>();
  const user = useAuthStore((s) => s.user);
  const fetchOptionSets = useOptionSetStore((s) => s.fetchAll);

  const [recent, setRecent] = useState<FPRegistration[]>([]);
  const [todayTotal, setTodayTotal] = useState(0);
  const [todayNew, setTodayNew] = useState(0);
  const [todayRevisit, setTodayRevisit] = useState(0);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async () => {
    try {
      fetchOptionSets();
      const today = todayISO();
      const res = await registrationsApi.list({ visit_date: today, per_page: 100 });
      setRecent(res.items.slice(0, 8));
      setTodayTotal(res.meta.total);
      setTodayNew(res.items.filter((r) => r.is_new_user).length);
      setTodayRevisit(res.items.filter((r) => r.is_revisit).length);
    } catch {}
  }, []);

  useFocusEffect(
    useCallback(() => {
      load();
    }, [load]),
  );

  const onRefresh = async () => {
    setRefreshing(true);
    await load();
    setRefreshing(false);
  };

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={onRefresh} />}
    >
      {/* Welcome */}
      <View style={styles.welcomeCard}>
        <View>
          <Text style={styles.welcomeText}>Welcome back,</Text>
          <Text style={styles.userName}>{user?.full_name || user?.email}</Text>
          <Text style={styles.facility}>
            {user?.facility?.name || 'All Facilities'}
          </Text>
        </View>
        <View style={styles.roleTag}>
          <Text style={styles.roleText}>{user?.role?.replace('_', ' ')}</Text>
        </View>
      </View>

      {/* Stats */}
      <View style={styles.statsRow}>
        <StatCard value={todayTotal} label="Today" color={colors.primary} />
        <StatCard value={todayNew} label="New Users" color={colors.accent} />
        <StatCard value={todayRevisit} label="Revisits" color={colors.warning} />
      </View>

      {/* Quick Actions */}
      {user && canCreateRegistration(user.role) && (
        <AppButton
          title="New Registration"
          onPress={() => nav.navigate('NewRegistration')}
          size="lg"
          icon={<Ionicons name="add-circle" size={20} color={colors.textInverse} />}
          style={{ marginBottom: spacing.xl }}
        />
      )}

      {/* Recent */}
      <Text style={styles.sectionTitle}>Today's Entries</Text>
      {recent.length === 0 ? (
        <EmptyState
          icon="document-text-outline"
          title="No entries today"
          message="Start by creating a new registration"
        />
      ) : (
        recent.map((item) => (
          <SubmissionListItem
            key={item.id}
            item={item}
            onPress={() => nav.navigate('SubmissionDetail', { id: item.id })}
          />
        ))
      )}
    </ScrollView>
  );
}

function StatCard({ value, label, color }: { value: number; label: string; color: string }) {
  return (
    <View style={[statStyles.card, { borderLeftColor: color }]}>
      <Text style={statStyles.value}>{value}</Text>
      <Text style={statStyles.label}>{label}</Text>
    </View>
  );
}

const statStyles = StyleSheet.create({
  card: {
    flex: 1,
    backgroundColor: colors.card,
    borderRadius: radii.md,
    padding: spacing.md,
    borderLeftWidth: 4,
    ...shadows.sm,
  },
  value: { fontSize: 24, fontWeight: '700', color: colors.text },
  label: { ...typography.caption, color: colors.textMuted, marginTop: 2, textTransform: 'uppercase', letterSpacing: 0.5 },
});

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg },
  welcomeCard: {
    backgroundColor: '#1e293b',
    borderRadius: radii.lg,
    padding: spacing.xl,
    marginBottom: spacing.lg,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
  },
  welcomeText: { ...typography.bodySmall, color: 'rgba(255,255,255,0.6)' },
  userName: { ...typography.h3, color: colors.textInverse, marginTop: 2 },
  facility: { ...typography.caption, color: 'rgba(255,255,255,0.5)', marginTop: 4 },
  roleTag: {
    backgroundColor: 'rgba(255,255,255,0.15)',
    paddingHorizontal: spacing.sm + 2,
    paddingVertical: spacing.xs,
    borderRadius: radii.sm,
  },
  roleText: { ...typography.caption, color: colors.textInverse, textTransform: 'capitalize' },
  statsRow: {
    flexDirection: 'row',
    gap: spacing.sm,
    marginBottom: spacing.xl,
  },
  sectionTitle: { ...typography.h4, color: colors.text, marginBottom: spacing.md },
});
