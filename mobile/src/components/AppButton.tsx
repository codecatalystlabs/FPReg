import React from 'react';
import {
  TouchableOpacity,
  Text,
  ActivityIndicator,
  StyleSheet,
  ViewStyle,
  TextStyle,
} from 'react-native';
import { colors, radii, spacing, typography } from '../theme';

interface Props {
  title: string;
  onPress: () => void;
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  loading?: boolean;
  disabled?: boolean;
  icon?: React.ReactNode;
  style?: ViewStyle;
}

export function AppButton({
  title,
  onPress,
  variant = 'primary',
  size = 'md',
  loading,
  disabled,
  icon,
  style,
}: Props) {
  const bg = variantStyles[variant].bg;
  const textColor = variantStyles[variant].text;
  const height = size === 'sm' ? 36 : size === 'lg' ? 52 : 44;

  return (
    <TouchableOpacity
      style={[
        styles.base,
        { backgroundColor: bg, height },
        variant === 'ghost' && styles.ghost,
        (disabled || loading) && styles.disabled,
        style,
      ]}
      onPress={onPress}
      disabled={disabled || loading}
      activeOpacity={0.7}
    >
      {loading ? (
        <ActivityIndicator size="small" color={textColor} />
      ) : (
        <>
          {icon}
          <Text style={[styles.text, { color: textColor, fontSize: size === 'sm' ? 13 : 15 }]}>
            {title}
          </Text>
        </>
      )}
    </TouchableOpacity>
  );
}

const variantStyles = {
  primary: { bg: colors.primary, text: colors.textInverse },
  secondary: { bg: colors.border, text: colors.text },
  danger: { bg: colors.danger, text: colors.textInverse },
  ghost: { bg: 'transparent', text: colors.primary },
};

const styles = StyleSheet.create({
  base: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: radii.md,
    paddingHorizontal: spacing.xl,
    gap: spacing.sm,
  },
  ghost: {
    borderWidth: 1,
    borderColor: colors.border,
  },
  disabled: {
    opacity: 0.5,
  },
  text: {
    ...typography.button,
  },
});
