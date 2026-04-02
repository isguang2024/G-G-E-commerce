import { Route, Routes } from 'react-router-dom'
import { AppShell } from '@/features/shell/AppShell'

export function AppRouter() {
  return (
    <Routes>
      <Route path="*" element={<AppShell />} />
    </Routes>
  )
}
