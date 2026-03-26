import React, { useCallback, useState } from 'react';
import { View, Text, FlatList, StyleSheet, RefreshControl } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { Ionicons } from '@expo/vector-icons';
import { usersApi } from '../api/users';
import { LoadingState } from '../components/LoadingState';
import { EmptyState } from '../components/EmptyState';
import { AppButton } from '../components/AppButton';
import { useAuthStore } from '../store/authStore';
import { colors, spacing, typography, radii, shadows } from '../theme';
import type { User } from '../types';
import { navigateToMainStack } from '../navigation/navigationHelpers';

export function UsersScreen() {
  const nav = useNavigation();
  const me = useAuthStore((s) => s.user);
  const [items, setItems] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async () => {
    try {
      const res = await usersApi.list({ page: 1, per_page: 100 });
      setItems(res.items);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      setLoading(true);
      load();
    }, [load]),
  );

  if (loading) return <LoadingState />;

  return (
    <View style={styles.container}>
      <View style={styles.actions}>
        <AppButton
          title="Create user"
          onPress={() => navigateToMainStack(nav, 'CreateUser')}
          size="md"
        />
      </View>
      <FlatList
        data={items}
        keyExtractor={(u) => u.id}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => { setRefreshing(true); load(); }} />}
        ListEmptyComponent={<EmptyState title="No users" message="No accounts in your scope yet." />}
        contentContainerStyle={styles.list}
        renderItem={({ item: u }) => (
          <View style={styles.card}>
            <View style={styles.cardTop}>
              <Text style={styles.name}>{u.full_name}</Text>
              <Text style={styles.role}>{u.role.replace(/_/g, ' ')}</Text>
            </View>
            <Text style={styles.email}>{u.email}</Text>
            <Text style={styles.meta}>
              {u.role === 'district_biostatistician' && u.district
                ? `District: ${u.district}`
                : u.facility
                  ? `${u.facility.name} (${u.facility.code})`
                  : '—'}
            </Text>
            {!u.is_active && <Text style={styles.inactive}>Inactive</Text>}
          </View>
        )}
      />
      {me?.role === 'district_biostatistician' && (
        <Text style={styles.hint}>You only see users for facilities in {me.district || 'your district'}.</Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  actions: { padding: spacing.lg, paddingBottom: spacing.sm },
  list: { paddingHorizontal: spacing.lg, paddingBottom: spacing.xxl },
  card: {
    backgroundColor: colors.card,
    borderRadius: radii.md,
    padding: spacing.lg,
    marginBottom: spacing.sm,
    ...shadows.sm,
  },
  cardTop: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 },
  name: { ...typography.body, fontWeight: '700', color: colors.text, flex: 1 },
  role: { ...typography.caption, color: colors.primary, textTransform: 'capitalize' },
  email: { ...typography.bodySmall, color: colors.textSecondary },
  meta: { ...typography.caption, color: colors.textMuted, marginTop: 4 },
  inactive: { ...typography.caption, color: colors.danger, marginTop: 6 },
  hint: { ...typography.caption, color: colors.textMuted, paddingHorizontal: spacing.lg, paddingBottom: spacing.lg },
});
