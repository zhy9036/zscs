import { create } from 'zustand'
import type { User } from '../types/auth'

const TOKEN_KEY = 'zmp_access_token'

interface AuthState {
  user: User | null
  accessToken: string | null
  authenticated: boolean
  hydrating: boolean
  setAuth: (user: User, token: string) => void
  setHydrating: (v: boolean) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: localStorage.getItem(TOKEN_KEY),
  authenticated: false,
  hydrating: true,
  setAuth: (user, token) => {
    localStorage.setItem(TOKEN_KEY, token)
    set({ user, accessToken: token, authenticated: true, hydrating: false })
  },
  setHydrating: (v) => set({ hydrating: v }),
  logout: () => {
    localStorage.removeItem(TOKEN_KEY)
    set({ user: null, accessToken: null, authenticated: false, hydrating: false })
  },
}))
