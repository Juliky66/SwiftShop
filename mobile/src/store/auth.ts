import { create } from 'zustand'
import AsyncStorage from '@react-native-async-storage/async-storage'
import { User } from '../types'
import * as authApi from '../api/auth'

interface AuthState {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, fullName: string) => Promise<void>
  logout: () => Promise<void>
  forceLogout: () => void
  loadFromStorage: () => Promise<void>
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoading: true,

  login: async (email, password) => {
    const data = await authApi.login({ email, password })
    await AsyncStorage.setItem('access_token', data.access_token)
    await AsyncStorage.setItem('refresh_token', data.refresh_token)
    set({ user: data.user })
  },

  register: async (email, password, fullName) => {
    const data = await authApi.register({ email, password, full_name: fullName })
    await AsyncStorage.setItem('access_token', data.access_token)
    await AsyncStorage.setItem('refresh_token', data.refresh_token)
    set({ user: data.user })
  },

  forceLogout: () => {
    set({ user: null })
  },

  logout: async () => {
    const refreshToken = await AsyncStorage.getItem('refresh_token')
    if (refreshToken) {
      try {
        await authApi.logout(refreshToken)
      } catch {
        // ignore errors during logout
      }
    }
    await AsyncStorage.multiRemove(['access_token', 'refresh_token'])
    set({ user: null })
  },

  loadFromStorage: async () => {
    try {
      const token = await AsyncStorage.getItem('access_token')
      if (token) {
        const user = await authApi.getMe()
        set({ user })
      }
    } catch {
      await AsyncStorage.multiRemove(['access_token', 'refresh_token'])
    } finally {
      set({ isLoading: false })
    }
  },
}))
