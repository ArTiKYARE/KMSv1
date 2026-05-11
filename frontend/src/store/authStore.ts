import { create } from 'zustand'
import type { User, AuthTokens } from '../types'
import { authAPI } from '../lib/api'

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, firstName: string, lastName: string) => Promise<void>
  logout: () => Promise<void>
  setTokens: (tokens: AuthTokens) => void
  clearAuth: () => void
  loadUserFromStorage: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: null,
  refreshToken: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,

  loadUserFromStorage: () => {
    const accessToken = localStorage.getItem('access_token')
    const refreshToken = localStorage.getItem('refresh_token')
    
    if (accessToken && refreshToken) {
      set({ 
        accessToken, 
        refreshToken, 
        isAuthenticated: true,
        isLoading: false 
      })
    } else {
      set({ isLoading: false })
    }
  },

  login: async (email: string, password: string) => {
    set({ isLoading: true, error: null })
    try {
      const tokens = await authAPI.login(email, password)
      get().setTokens(tokens)
    } catch (err: any) {
      set({ 
        error: err.response?.data?.error || 'Login failed', 
        isLoading: false 
      })
      throw err
    }
  },

  register: async (email: string, password: string, firstName: string, lastName: string) => {
    set({ isLoading: true, error: null })
    try {
      await authAPI.register(email, password, firstName, lastName)
      await get().login(email, password)
    } catch (err: any) {
      set({ 
        error: err.response?.data?.error || 'Registration failed', 
        isLoading: false 
      })
      throw err
    }
  },

  logout: async () => {
    try {
      await authAPI.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      get().clearAuth()
    }
  },

  setTokens: (tokens: AuthTokens) => {
    localStorage.setItem('access_token', tokens.access_token)
    localStorage.setItem('refresh_token', tokens.refresh_token)
    set({
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
      isAuthenticated: true,
      isLoading: false,
    })
  },

  clearAuth: () => {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    set({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    })
  },
}))
