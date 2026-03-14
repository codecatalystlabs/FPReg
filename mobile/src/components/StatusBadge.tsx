import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { colors, radii, spacing, typography } from '../theme';

type Variant = 'success' | 'warning' | 'danger' | 'info' | 'neutral';

interface Props {
  label: string;
  variant?: Variant;
}

const variantMap: Record<Variant, { bg: string; text: string }> = {
  success: { bg: colors.accentLight, text: colors.accent },
  warning: { bg: colors.warningLight, text: '#92400e' },
  danger: { bg: colors.dangerLight, text: colors.danger },
  info: { bg: colors.infoLight, text: colors.info },
  neutral: { bg: colors.divider, text: colors.textSecondary },
};

export function StatusBadge({ label, variant = 'neutral' }: Props) {
  const v = variantMap[variant];
  return (
    <View style={[styles.badge, { backgroundColor: v.bg }]}>
      <Text style={[styles.text, { color: v.text }]}>{label}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    paddingHorizontal: spacing.sm + 2,
    paddingVertical: spacing.xs,
    borderRadius: radii.sm,
    alignSelf: 'flex-start',
  },
  text: {
    ...typography.caption,
    fontWeight: '600',
  },
});
