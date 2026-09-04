import { useEffect } from 'react'
import { authApi } from '../api/auth'
import { useAuthStore } from '../stores/authStore'

export function useAuth() {
  const store = useAuthStore()

  useEffect(() => {
    if (!store.accessToken) {
      store.setHydrating(false)
      return
    }
    let cancelled = false
    authApi
      .me()
      .then((user) => {
        if (cancelled) return
        useAuthStore.setState({
          user,
          authenticated: true,
          hydrating: false,
        })
      })
      .catch(() => {
        if (cancelled) return
        store.logout()
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return store
}
