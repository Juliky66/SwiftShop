import { authClient } from './client'
import { User } from '../types'

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
  expires_at: string
}

export interface LoginPayload {
  email: string
  password: string
}

export interface RegisterPayload {
  email: string
  password: string
  full_name: string
  phone?: string
}

export async function login(payload: LoginPayload): Promise<AuthResponse> {
  const { data } = await authClient.post<AuthResponse>('/api/v1/auth/login', payload)
  return data
}

export async function register(payload: RegisterPayload): Promise<AuthResponse> {
  const { data } = await authClient.post<AuthResponse>('/api/v1/auth/register', payload)
  return data
}

export async function logout(refreshToken: string): Promise<void> {
  await authClient.post('/api/v1/auth/logout', { refresh_token: refreshToken })
}

export async function getMe(): Promise<User> {
  const { data } = await authClient.get<User>('/api/v1/auth/me')
  return data
}

export async function refreshTokens(refreshToken: string): Promise<Omit<AuthResponse, 'user'>> {
  const { data } = await authClient.post<Omit<AuthResponse, 'user'>>('/api/v1/auth/refresh', {
    refresh_token: refreshToken,
  })
  return data
}
