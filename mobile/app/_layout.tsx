import { useEffect } from 'react'
import { Stack, useRouter, useSegments } from 'expo-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { PaperProvider, MD3LightTheme } from 'react-native-paper'
import { GestureHandlerRootView } from 'react-native-gesture-handler'
import { useAuthStore } from '../src/store/auth'
import { setSessionExpiredCallback } from '../src/api/authEvents'

const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: 1, staleTime: 30_000 } },
})

const theme = {
  ...MD3LightTheme,
  colors: { ...MD3LightTheme.colors, primary: '#1E90FF', secondary: '#63B3ED' },
}

export default function RootLayout() {
  const { user, isLoading, loadFromStorage, forceLogout } = useAuthStore()
  const segments = useSegments()
  const router = useRouter()

  useEffect(() => {
    loadFromStorage()
    setSessionExpiredCallback(forceLogout)
  }, [])

  useEffect(() => {
    if (isLoading) return
    const inAuth = segments[0] === '(auth)'
    if (!user && !inAuth) router.replace('/(auth)/login')
    else if (user && inAuth) router.replace('/(tabs)')
  }, [user, isLoading])

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <QueryClientProvider client={queryClient}>
        <PaperProvider theme={theme}>
          <Stack screenOptions={{ headerShown: false }} />
        </PaperProvider>
      </QueryClientProvider>
    </GestureHandlerRootView>
  )
}
