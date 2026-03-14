import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppButton } from './AppButton';
import { colors, spacing, typography } from '../theme';

interface Props {
  message?: string;
  onRetry?: () => void;
}

export function ErrorState({ message = 'Something went wrong', onRetry }: Props) {
  return (
    <View style={styles.container}>
      <Ionicons name="alert-circle-outline" size={48} color={colors.danger} />
      <Text style={styles.text}>{message}</Text>
      {onRetry && <AppButton title="Retry" variant="ghost" size="sm" onPress={onRetry} />}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: spacing.xxxl, gap: spacing.md },
  text: { ...typography.body, color: colors.textSecondary, textAlign: 'center', maxWidth: 280 },
});
