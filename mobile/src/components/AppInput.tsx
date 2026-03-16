import React from 'react';
import { View, Text, TextInput, StyleSheet, TextInputProps } from 'react-native';
import { colors, radii, spacing, typography } from '../theme';

interface Props extends TextInputProps {
  label: string;
  error?: string;
  helpText?: string;
  required?: boolean;
  disabled?: boolean;
}

export function AppInput({ label, error, helpText, required, disabled, style, editable, ...props }: Props) {
  return (
    <View style={[styles.container, disabled && styles.containerDisabled]}>
      <Text style={styles.label}>
        {label}
        {required && <Text style={styles.required}> *</Text>}
      </Text>
      <TextInput
        style={[styles.input, error && styles.inputError, disabled && styles.inputDisabled, style]}
        placeholderTextColor={colors.textMuted}
        editable={editable ?? !disabled}
        {...props}
      />
      {error ? (
        <Text style={styles.error}>{error}</Text>
      ) : helpText ? (
        <Text style={styles.help}>{helpText}</Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { marginBottom: spacing.md },
  containerDisabled: { opacity: 0.7 },
  label: { ...typography.label, color: colors.textSecondary, marginBottom: spacing.xs },
  required: { color: colors.danger },
  input: {
    backgroundColor: colors.inputBg,
    borderWidth: 1,
    borderColor: colors.inputBorder,
    borderRadius: radii.sm,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm + 2,
    fontSize: 15,
    color: colors.text,
  },
  inputError: { borderColor: colors.danger },
  inputDisabled: { backgroundColor: colors.divider },
  error: { ...typography.caption, color: colors.danger, marginTop: spacing.xs },
  help: { ...typography.caption, color: colors.textMuted, marginTop: spacing.xs },
});
