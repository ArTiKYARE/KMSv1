import axios from 'axios'
import type { AuthTokens, User } from '../types'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      const refreshToken = localStorage.getItem('refresh_token')
      if (!refreshToken) {
        throw error
      }

      try {
        const response = await axios.post(`${API_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        })

        const { access_token, refresh_token } = response.data as AuthTokens
        localStorage.setItem('access_token', access_token)
        localStorage.setItem('refresh_token', refresh_token)

        originalRequest.headers.Authorization = `Bearer ${access_token}`
        return apiClient(originalRequest)
      } catch (refreshError) {
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        window.location.href = '/login'
        throw refreshError
      }
    }

    throw error
  }
)

export const authAPI = {
  login: async (email: string, password: string): Promise<AuthTokens> => {
    const response = await apiClient.post('/auth/login', { email, password })
    return response.data as AuthTokens
  },

  register: async (
    email: string,
    password: string,
    firstName: string,
    lastName: string
  ): Promise<{ user: User }> => {
    const response = await apiClient.post('/auth/register', {
      email,
      password,
      first_name: firstName,
      last_name: lastName,
    })
    return response.data as { user: User }
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout')
  },

  getMe: async (): Promise<User> => {
    const response = await apiClient.get('/auth/me')
    return response.data as User
  },
}

export default apiClient
