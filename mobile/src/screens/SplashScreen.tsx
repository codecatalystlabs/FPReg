import React from 'react';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, spacing, typography } from '../theme';

export function SplashScreen() {
  return (
    <View style={styles.container}>
      <Ionicons name="heart-half" size={64} color={colors.textInverse} />
      <Text style={styles.title}>FP Register</Text>
      <Text style={styles.subtitle}>HMIS MCH 007</Text>
      <ActivityIndicator size="small" color={colors.textInverse} style={styles.spinner} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1e293b',
    alignItems: 'center',
    justifyContent: 'center',
    gap: spacing.sm,
  },
  title: { ...typography.h1, color: colors.textInverse },
  subtitle: { ...typography.bodySmall, color: 'rgba(255,255,255,0.6)' },
  spinner: { marginTop: spacing.xxl },
});
