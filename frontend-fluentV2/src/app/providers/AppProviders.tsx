import type { PropsWithChildren } from 'react'
import { useEffect } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { HashRouter } from 'react-router-dom'
import { AppErrorBoundary } from '@/app/ErrorBoundary'
import { FluentThemeProvider } from '@/app/providers/FluentThemeProvider'
import { AuthBootstrap } from '@/features/auth/AuthBootstrap'
import { useAuthStore } from '@/features/auth/auth.store'
import { queryClient } from '@/shared/api/query-client'
import { setUnauthorizedHandler } from '@/shared/api/client'

function buildLoginRedirectHash() {
  if (typeof window === 'undefined') {
    return '#/login'
  }

  const currentHash = window.location.hash.replace(/^#/, '') || '/'
  const redirect = encodeURIComponent(currentHash)
  return `#/login?redirect=${redirect}`
}

function AuthRuntimeBridge() {
  useEffect(() => {
    setUnauthorizedHandler(() => {
      useAuthStore.getState().clearAuth()
      queryClient.clear()
      if (typeof window !== 'undefined' && !window.location.hash.startsWith('#/login')) {
        window.location.hash = buildLoginRedirectHash()
      }
    })

    return () => setUnauthorizedHandler(null)
  }, [])

  return null
}

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <QueryClientProvider client={queryClient}>
      <HashRouter>
        <FluentThemeProvider>
          <AppErrorBoundary>
            <AuthRuntimeBridge />
            <AuthBootstrap>{children}</AuthBootstrap>
          </AppErrorBoundary>
        </FluentThemeProvider>
      </HashRouter>
    </QueryClientProvider>
  )
}
