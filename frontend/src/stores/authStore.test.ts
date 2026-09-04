import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

describe('authStore', () => {
  beforeEach(() => {
    localStorage.clear()
    useAuthStore.setState({
      user: null,
      accessToken: null,
      authenticated: false,
      hydrating: true,
    })
  })

  it('sets auth and persists token', () => {
    useAuthStore.getState().setAuth({ id: 'u1', username: 'alice' }, 'tok-123')
    const s = useAuthStore.getState()
    expect(s.authenticated).toBe(true)
    expect(s.accessToken).toBe('tok-123')
    expect(s.user?.username).toBe('alice')
    expect(localStorage.getItem('zmp_access_token')).toBe('tok-123')
  })

  it('logout clears auth and token', () => {
    useAuthStore.getState().setAuth({ id: 'u1', username: 'alice' }, 'tok-123')
    useAuthStore.getState().logout()
    const s = useAuthStore.getState()
    expect(s.authenticated).toBe(false)
    expect(s.accessToken).toBeNull()
    expect(localStorage.getItem('zmp_access_token')).toBeNull()
  })
})
