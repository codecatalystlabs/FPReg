import React, { useCallback, useState } from 'react';
import { View, FlatList, TextInput, StyleSheet } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { Ionicons } from '@expo/vector-icons';
import { SubmissionListItem } from '../components/SubmissionListItem';
import { LoadingState } from '../components/LoadingState';
import { EmptyState } from '../components/EmptyState';
import { ErrorState } from '../components/ErrorState';
import { registrationsApi, RegistrationListParams } from '../api/registrations';
import { colors, spacing, radii, typography } from '../theme';
import type { FPRegistration, PaginationMeta } from '../types';
import type { MainStackParamList } from '../navigation/MainNavigator';

type NavProp = NativeStackNavigationProp<MainStackParamList>;

export function SubmissionsScreen() {
  const nav = useNavigation<NavProp>();
  const [items, setItems] = useState<FPRegistration[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async (p = 1, append = false) => {
    if (!append) setLoading(true);
    setError(false);
    try {
      const params: RegistrationListParams = { page: p, per_page: 25 };
      if (search.trim()) params.search = search.trim();
      const res = await registrationsApi.list(params);
      setItems(append ? [...items, ...res.items] : res.items);
      setMeta(res.meta);
      setPage(p);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [search]);

  useFocusEffect(
    useCallback(() => {
      load(1);
    }, [load]),
  );

  const onRefresh = () => {
    setRefreshing(true);
    load(1);
  };

  const onEndReached = () => {
    if (meta && page < meta.total_pages) {
      load(page + 1, true);
    }
  };

  if (loading && !refreshing) return <LoadingState />;
  if (error) return <ErrorState onRetry={() => load(1)} />;

  return (
    <View style={styles.container}>
      {/* Search bar */}
      <View style={styles.searchBar}>
        <Ionicons name="search" size={18} color={colors.textMuted} />
        <TextInput
          style={styles.searchInput}
          placeholder="Search name, NIN, client #..."
          placeholderTextColor={colors.textMuted}
          value={search}
          onChangeText={setSearch}
          onSubmitEditing={() => load(1)}
          returnKeyType="search"
        />
        {search.length > 0 && (
          <Ionicons
            name="close-circle"
            size={18}
            color={colors.textMuted}
            onPress={() => { setSearch(''); setTimeout(() => load(1), 100); }}
          />
        )}
      </View>

      <FlatList
        data={items}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <SubmissionListItem
            item={item}
            onPress={() => nav.navigate('SubmissionDetail', { id: item.id })}
          />
        )}
        contentContainerStyle={styles.list}
        refreshing={refreshing}
        onRefresh={onRefresh}
        onEndReached={onEndReached}
        onEndReachedThreshold={0.3}
        ListEmptyComponent={
          <EmptyState
            title="No submissions found"
            message={search ? 'Try a different search term' : 'Create your first registration'}
          />
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  searchBar: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surface,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
    paddingHorizontal: spacing.lg,
    paddingVertical: spacing.sm,
    gap: spacing.sm,
  },
  searchInput: {
    flex: 1,
    fontSize: 15,
    color: colors.text,
    paddingVertical: spacing.xs,
  },
  list: { padding: spacing.lg, paddingTop: spacing.sm },
});
