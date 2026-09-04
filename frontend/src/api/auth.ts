import { apiClient } from './client'
import type { LoginResponse, User } from '../types/auth'

export const authApi = {
  login: (username: string, password: string) =>
    apiClient.post<LoginResponse>('/auth/login', { username, password }),
  me: () => apiClient.get<User>('/auth/me'),
  register: (username: string, password: string) =>
    apiClient.post<User>('/auth/register', { username, password }),
}
