import React from 'react';
import { TouchableOpacity, Text, View, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, radii, spacing, typography } from '../theme';

interface Props {
  label: string;
  value: boolean;
  onChange: (value: boolean) => void;
  helpText?: string;
  disabled?: boolean;
}

export function AppCheckbox({ label, value, onChange, helpText, disabled }: Props) {
  return (
    <TouchableOpacity
      style={[styles.container, disabled && styles.disabled]}
      onPress={() => !disabled && onChange(!value)}
      activeOpacity={0.7}
    >
      <View style={[styles.box, value && styles.boxChecked]}>
        {value && <Ionicons name="checkmark" size={14} color={colors.textInverse} />}
      </View>
      <View style={styles.textContainer}>
        <Text style={styles.label}>{label}</Text>
        {helpText && <Text style={styles.help}>{helpText}</Text>}
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: spacing.sm + 2,
    paddingVertical: spacing.xs + 2,
  },
  disabled: { opacity: 0.4 },
  box: {
    width: 22,
    height: 22,
    borderRadius: radii.sm - 2,
    borderWidth: 2,
    borderColor: colors.inputBorder,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 1,
  },
  boxChecked: {
    backgroundColor: colors.primary,
    borderColor: colors.primary,
  },
  textContainer: { flex: 1 },
  label: { ...typography.body, color: colors.text },
  help: { ...typography.caption, color: colors.textMuted, marginTop: 2 },
});
