import { Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/features/shell/AppShell'
import { routeDefinitions, renderRouteElement } from '@/features/navigation/route-registry'
import { RedirectIfAuthenticated, RequireAuth } from '@/features/session/AuthGuards'
import { ForgotPasswordPage } from '@/pages/auth/ForgotPasswordPage'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'
import { NotFoundPage } from '@/pages/not-found/NotFoundPage'
import { appConfig } from '@/shared/config/app-config'

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
        <Route index element={<Navigate replace to={appConfig.defaultRoute} />} />
        {routeDefinitions.map((definition) => (
          <Route key={definition.id} path={definition.path} element={renderRouteElement(definition.id)} />
        ))}
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}
