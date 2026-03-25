import React, { useState } from 'react';
import {
  View,
  Text,
  Image,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppInput } from '../components/AppInput';
import { AppButton } from '../components/AppButton';
import { useAuthStore } from '../store/authStore';
import { colors, spacing, typography, radii, shadows } from '../theme';
import { AxiosError } from 'axios';

export function LoginScreen() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login, isLoading } = useAuthStore();

  const handleLogin = async () => {
    setError('');
    if (!email.trim() || !password) {
      setError('Please enter email and password');
      return;
    }
    try {
      await login(email.trim(), password);
    } catch (e) {
      if (e instanceof AxiosError && e.response?.data?.message) {
        setError(e.response.data.message);
      } else {
        setError('Unable to sign in. Please check your credentials.');
      }
    }
  };

  return (
    <View style={styles.bg}>
      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
        style={styles.flex}
      >
        <ScrollView
          contentContainerStyle={styles.scroll}
          keyboardShouldPersistTaps="handled"
        >
          <Image
            source={require('../../assets/branding/moh_header.png')}
            resizeMode="contain"
            style={styles.mohHeader}
          />
          <View style={styles.header}>
            <Ionicons name="heart-half" size={48} color={colors.textInverse} />
            <Text style={styles.brandTitle}>FP Register</Text>
            <Text style={styles.brandSub}>HMIS MCH 007 — Integrated Family Planning</Text>
          </View>

          <View style={styles.card}>
            <Text style={styles.cardTitle}>Sign In</Text>

            {!!error && (
              <View style={styles.errorBox}>
                <Ionicons name="alert-circle" size={16} color={colors.danger} />
                <Text style={styles.errorText}>{error}</Text>
              </View>
            )}

            <AppInput
              label="Email Address"
              placeholder="user@moh.go.ug"
              value={email}
              onChangeText={setEmail}
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              required
            />

            <AppInput
              label="Password"
              placeholder="Enter password"
              value={password}
              onChangeText={setPassword}
              secureTextEntry
              required
            />

            <AppButton
              title="Sign In"
              onPress={handleLogin}
              loading={isLoading}
              size="lg"
              style={{ marginTop: spacing.sm }}
            />
          </View>

          <Text style={styles.footer}>Ministry of Health — Uganda</Text>
        </ScrollView>
      </KeyboardAvoidingView>
    </View>
  );
}

const styles = StyleSheet.create({
  bg: { flex: 1, backgroundColor: '#1e293b' },
  flex: { flex: 1 },
  scroll: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: spacing.xxl,
  },
  mohHeader: {
    width: '100%',
    height: 34,
    marginBottom: spacing.xl,
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing.xxxl,
    gap: spacing.xs,
  },
  brandTitle: { ...typography.h1, color: colors.textInverse },
  brandSub: { ...typography.bodySmall, color: 'rgba(255,255,255,0.55)', textAlign: 'center' },
  card: {
    backgroundColor: colors.surface,
    borderRadius: radii.lg,
    padding: spacing.xxl,
    ...shadows.lg,
  },
  cardTitle: { ...typography.h3, color: colors.text, marginBottom: spacing.xl, textAlign: 'center' },
  errorBox: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    backgroundColor: colors.dangerLight,
    padding: spacing.md,
    borderRadius: radii.sm,
    marginBottom: spacing.lg,
  },
  errorText: { ...typography.bodySmall, color: colors.danger, flex: 1 },
  footer: {
    ...typography.caption,
    color: 'rgba(255,255,255,0.35)',
    textAlign: 'center',
    marginTop: spacing.xxl,
  },
});
