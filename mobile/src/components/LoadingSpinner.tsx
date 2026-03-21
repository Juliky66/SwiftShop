import React from 'react'
import { View, ActivityIndicator, StyleSheet } from 'react-native'
import { Text } from 'react-native-paper'

interface LoadingSpinnerProps {
  message?: string
}

export default function LoadingSpinner({ message }: LoadingSpinnerProps) {
  return (
    <View style={styles.container}>
      <ActivityIndicator size="large" color="#1E90FF" />
      {message ? <Text style={styles.message}>{message}</Text> : null}
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 12,
  },
  message: {
    color: '#6B7280',
    fontSize: 14,
    marginTop: 8,
  },
})
