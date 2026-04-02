import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/features/shell/AppShell'
import { RedirectIfAuthenticated, RequireAuth } from '@/features/auth/AuthGuards'
import { RuntimePageOutlet } from '@/features/navigation/route-registry'
import { ForgotPasswordPage } from '@/pages/auth/ForgotPasswordPage'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'

export function AppRouter() {
  return (
    <Routes>
      <Route element={<RedirectIfAuthenticated />}>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/auth" element={<Navigate replace to="/login" />} />
        <Route path="/auth/login" element={<Navigate replace to="/login" />} />
        <Route path="/auth/register" element={<Navigate replace to="/register" />} />
        <Route path="/auth/forgot-password" element={<Navigate replace to="/forgot-password" />} />
      </Route>
      <Route
        element={
          <RequireAuth>
            <AppShell />
          </RequireAuth>
        }
      >
        <Route path="*" element={<RuntimePageOutlet />} />
      </Route>
    </Routes>
  )
}
