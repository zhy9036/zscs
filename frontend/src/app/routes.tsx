import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import { AppLayout } from '../components/layout/AppLayout'

export function RequireAuth() {
  const { authenticated, hydrating } = useAuth()
  const location = useLocation()

  if (hydrating) {
    return (
      <div className="flex h-full items-center justify-center text-gray-500">
        Loading…
      </div>
    )
  }

  if (!authenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  )
}
