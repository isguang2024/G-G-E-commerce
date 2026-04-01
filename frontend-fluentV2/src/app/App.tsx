import { AppProviders } from '@/app/providers/AppProviders'
import { AppRouter } from '@/app/router/AppRouter'
import '@/app/global.css'

export function App() {
  return (
    <AppProviders>
      <AppRouter />
    </AppProviders>
  )
}
