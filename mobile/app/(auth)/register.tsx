import React, { useState } from 'react'
import { View, StyleSheet, KeyboardAvoidingView, Platform, ScrollView } from 'react-native'
import { Text, TextInput, Button, Surface, HelperText } from 'react-native-paper'
import { Link } from 'expo-router'
import { useAuthStore } from '../../src/store/auth'

export default function RegisterScreen() {
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [phone, setPhone] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [passwordVisible, setPasswordVisible] = useState(false)
  const { register } = useAuthStore()

  const validate = (): string | null => {
    if (!fullName.trim()) return 'Full name is required.'
    if (!email.trim()) return 'Email is required.'
    if (password.length < 8) return 'Password must be at least 8 characters.'
    return null
  }

  const handleRegister = async () => {
    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }
    setError(null)
    setLoading(true)
    try {
      await register(email.trim(), password, fullName.trim())
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { message?: string } } })?.response?.data?.message ??
        'Registration failed. Please try again.'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scroll} keyboardShouldPersistTaps="handled">
        <View style={styles.header}>
          <Text style={styles.appTitle}>SwiftShop</Text>
          <Text style={styles.subtitle}>Create your account</Text>
        </View>

        <Surface style={styles.card} elevation={3}>
          <TextInput
            label="Full Name"
            value={fullName}
            onChangeText={setFullName}
            autoCapitalize="words"
            style={styles.input}
            mode="outlined"
            outlineColor="#C7DCF8"
            activeOutlineColor="#1E90FF"
          />
          <TextInput
            label="Email"
            value={email}
            onChangeText={setEmail}
            keyboardType="email-address"
            autoCapitalize="none"
            autoCorrect={false}
            style={styles.input}
            mode="outlined"
            outlineColor="#C7DCF8"
            activeOutlineColor="#1E90FF"
          />
          <TextInput
            label="Password (min 8 characters)"
            value={password}
            onChangeText={setPassword}
            secureTextEntry={!passwordVisible}
            style={styles.input}
            mode="outlined"
            outlineColor="#C7DCF8"
            activeOutlineColor="#1E90FF"
            right={
              <TextInput.Icon
                icon={passwordVisible ? 'eye-off' : 'eye'}
                onPress={() => setPasswordVisible(!passwordVisible)}
              />
            }
          />
          <HelperText type="info" visible={password.length > 0 && password.length < 8}>
            Password must be at least 8 characters
          </HelperText>
          <TextInput
            label="Phone (optional)"
            value={phone}
            onChangeText={setPhone}
            keyboardType="phone-pad"
            style={styles.input}
            mode="outlined"
            outlineColor="#C7DCF8"
            activeOutlineColor="#1E90FF"
          />

          {error && <Text style={styles.error}>{error}</Text>}

          <Button
            mode="contained"
            onPress={handleRegister}
            loading={loading}
            disabled={loading}
            style={styles.button}
            buttonColor="#1E90FF"
            contentStyle={styles.buttonContent}
            labelStyle={styles.buttonLabel}
          >
            Create Account
          </Button>

          <View style={styles.linkRow}>
            <Text style={styles.linkText}>Already have an account? </Text>
            <Link href="/(auth)/login" asChild>
              <Button mode="text" compact textColor="#1E90FF" style={styles.linkButton}>
                Sign In
              </Button>
            </Link>
          </View>
        </Surface>
      </ScrollView>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#E6F0FF',
  },
  scroll: {
    flexGrow: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
    backgroundColor: '#E6F0FF',
  },
  header: {
    alignItems: 'center',
    marginBottom: 32,
  },
  appTitle: {
    fontSize: 38,
    fontWeight: '800',
    color: '#1E90FF',
    letterSpacing: -0.5,
    textShadowColor: 'rgba(30,144,255,0.25)',
    textShadowOffset: { width: 0, height: 2 },
    textShadowRadius: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#4B7BB5',
    marginTop: 6,
  },
  card: {
    width: '100%',
    maxWidth: 400,
    padding: 24,
    borderRadius: 20,
    backgroundColor: '#FFFFFF',
    shadowColor: '#1E90FF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.12,
    shadowRadius: 16,
  },
  input: {
    marginBottom: 14,
    backgroundColor: '#F8FAFF',
  },
  error: {
    color: '#EF4444',
    fontSize: 13,
    marginBottom: 12,
    textAlign: 'center',
  },
  button: {
    borderRadius: 12,
    marginTop: 4,
    shadowColor: '#1E90FF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 4,
  },
  buttonContent: {
    height: 50,
  },
  buttonLabel: {
    fontSize: 16,
    fontWeight: '700',
    letterSpacing: 0.3,
  },
  linkRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 16,
  },
  linkText: {
    color: '#6B7280',
    fontSize: 14,
  },
  linkButton: {
    marginLeft: -4,
  },
})
