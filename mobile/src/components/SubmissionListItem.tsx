import React from 'react';
import { TouchableOpacity, View, Text, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { StatusBadge } from './StatusBadge';
import { colors, radii, spacing, typography, shadows } from '../theme';
import type { FPRegistration } from '../types';

interface Props {
  item: FPRegistration;
  onPress: () => void;
  /** Show which facility owns the row (multi-facility roles). */
  showFacility?: boolean;
}

export function SubmissionListItem({ item, onPress, showFacility }: Props) {
  return (
    <TouchableOpacity style={styles.card} onPress={onPress} activeOpacity={0.7}>
      <View style={styles.row}>
        <View style={styles.left}>
          <Text style={styles.name}>{item.surname} {item.given_name}</Text>
          <Text style={styles.clientNum}>{item.client_number || 'Visitor'}</Text>
          {showFacility && item.facility && (
            <Text style={styles.facility}>{item.facility.name}</Text>
          )}
        </View>
        <View style={styles.right}>
          <StatusBadge
            label={item.is_new_user ? 'New' : 'Revisit'}
            variant={item.is_new_user ? 'success' : 'warning'}
          />
        </View>
      </View>
      <View style={styles.meta}>
        <View style={styles.metaItem}>
          <Ionicons name="calendar-outline" size={13} color={colors.textMuted} />
          <Text style={styles.metaText}>{item.visit_date}</Text>
        </View>
        <View style={styles.metaItem}>
          <Ionicons name="person-outline" size={13} color={colors.textMuted} />
          <Text style={styles.metaText}>{item.sex} / {item.age}y</Text>
        </View>
        {item.hts_code && (
          <View style={styles.metaItem}>
            <Text style={styles.metaText}>HTS: {item.hts_code}</Text>
          </View>
        )}
        <View style={{ flex: 1 }} />
        <Ionicons name="chevron-forward" size={16} color={colors.textMuted} />
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.card,
    borderRadius: radii.md,
    padding: spacing.lg,
    marginBottom: spacing.sm,
    ...shadows.sm,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: spacing.sm,
  },
  left: { flex: 1 },
  right: { marginLeft: spacing.sm },
  name: { ...typography.h4, color: colors.text },
  clientNum: { ...typography.caption, color: colors.textMuted, marginTop: 2 },
  facility: { ...typography.caption, color: colors.primary, marginTop: 4, fontWeight: '600' },
  meta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    borderTopWidth: StyleSheet.hairlineWidth,
    borderTopColor: colors.divider,
    paddingTop: spacing.sm,
  },
  metaItem: { flexDirection: 'row', alignItems: 'center', gap: 4 },
  metaText: { ...typography.caption, color: colors.textMuted },
});
