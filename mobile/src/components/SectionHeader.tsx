import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, spacing, typography, radii } from '../theme';

interface Props {
  title: string;
  subtitle?: string;
  icon?: keyof typeof Ionicons.glyphMap;
}

export function SectionHeader({ title, subtitle, icon }: Props) {
  return (
    <View style={styles.container}>
      <View style={styles.iconContainer}>
        {icon && <Ionicons name={icon} size={16} color={colors.textInverse} />}
      </View>
      <View>
        <Text style={styles.title}>{title}</Text>
        {subtitle && <Text style={styles.subtitle}>{subtitle}</Text>}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm + 2,
    marginBottom: spacing.md,
    marginTop: spacing.lg,
  },
  iconContainer: {
    width: 28,
    height: 28,
    borderRadius: radii.sm,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  title: { ...typography.h4, color: colors.text },
  subtitle: { ...typography.caption, color: colors.textMuted },
});
