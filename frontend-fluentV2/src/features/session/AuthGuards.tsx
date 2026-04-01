import type { PropsWithChildren } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { appConfig } from '@/shared/config/app-config'
import { useAuthStore } from '@/features/session/auth.store'

export function RequireAuth({ children }: PropsWithChildren) {
  const authenticated = useAuthStore((state) => state.authenticated)
  const location = useLocation()

  if (!authenticated) {
    const from = `${location.pathname}${location.search}${location.hash}`
    return <Navigate replace state={{ from }} to="/login" />
  }

  return children ?? <Outlet />
}

export function RedirectIfAuthenticated({ children }: PropsWithChildren) {
  const authenticated = useAuthStore((state) => state.authenticated)

  if (authenticated) {
    return <Navigate replace to={appConfig.defaultRoute} />
  }

  return children ?? <Outlet />
}
