import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, spacing, typography } from '../theme';

interface Props {
  icon?: keyof typeof Ionicons.glyphMap;
  title: string;
  message?: string;
  action?: React.ReactNode;
}

export function EmptyState({ icon = 'document-text-outline', title, message, action }: Props) {
  return (
    <View style={styles.container}>
      <Ionicons name={icon} size={56} color={colors.textMuted} />
      <Text style={styles.title}>{title}</Text>
      {message && <Text style={styles.message}>{message}</Text>}
      {action && <View style={styles.action}>{action}</View>}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { alignItems: 'center', justifyContent: 'center', padding: spacing.xxxl, gap: spacing.sm },
  title: { ...typography.h4, color: colors.textSecondary, textAlign: 'center' },
  message: { ...typography.bodySmall, color: colors.textMuted, textAlign: 'center', maxWidth: 260 },
  action: { marginTop: spacing.lg },
});
