import axios, { AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import AsyncStorage from '@react-native-async-storage/async-storage'
import { notifySessionExpired } from './authEvents'

const AUTH_URL = process.env.EXPO_PUBLIC_AUTH_URL ?? 'http://localhost:8081'
const CATALOG_URL = process.env.EXPO_PUBLIC_CATALOG_URL ?? 'http://localhost:8082'
const ORDERS_URL = process.env.EXPO_PUBLIC_ORDERS_URL ?? 'http://localhost:8083'
const REVIEWS_URL = process.env.EXPO_PUBLIC_REVIEWS_URL ?? 'http://localhost:8085'

// Bare client for token refresh (no interceptors to avoid loops)
const bareAuthClient = axios.create({ baseURL: AUTH_URL, timeout: 10_000 })

function createClient(baseURL: string): AxiosInstance {
  const client = axios.create({ baseURL, timeout: 10_000 })

  client.interceptors.request.use(async (config: InternalAxiosRequestConfig) => {
    const token = await AsyncStorage.getItem('access_token')
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  })

  client.interceptors.response.use(
    (res) => res,
    async (error) => {
      const original = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
      if (error.response?.status === 401 && !original._retry) {
        original._retry = true
        try {
          const refreshToken = await AsyncStorage.getItem('refresh_token')
          if (!refreshToken) throw new Error('no refresh token')
          const { data } = await bareAuthClient.post('/api/v1/auth/refresh', {
            refresh_token: refreshToken,
          })
          await AsyncStorage.setItem('access_token', data.access_token)
          await AsyncStorage.setItem('refresh_token', data.refresh_token)
          original.headers.Authorization = `Bearer ${data.access_token}`
          return client(original)
        } catch {
          await AsyncStorage.multiRemove(['access_token', 'refresh_token'])
          notifySessionExpired()
        }
      }
      return Promise.reject(error)
    },
  )

  return client
}

export const authClient = createClient(AUTH_URL)
export const catalogClient = createClient(CATALOG_URL)
export const ordersClient = createClient(ORDERS_URL)
export const reviewsClient = createClient(REVIEWS_URL)
