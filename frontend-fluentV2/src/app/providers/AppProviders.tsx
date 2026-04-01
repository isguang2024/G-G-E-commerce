import type { PropsWithChildren } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { HashRouter } from 'react-router-dom'
import { AppErrorBoundary } from '@/app/ErrorBoundary'
import { FluentThemeProvider } from '@/app/providers/FluentThemeProvider'
import { queryClient } from '@/shared/api/query-client'

export function AppProviders({ children }: PropsWithChildren) {
  return (
    <QueryClientProvider client={queryClient}>
      <HashRouter>
        <FluentThemeProvider>
          <AppErrorBoundary>{children}</AppErrorBoundary>
        </FluentThemeProvider>
      </HashRouter>
    </QueryClientProvider>
  )
}
