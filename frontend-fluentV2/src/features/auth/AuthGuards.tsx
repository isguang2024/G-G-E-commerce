import type { PropsWithChildren } from 'react'
import { Spinner, makeStyles } from '@fluentui/react-components'
import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { appConfig } from '@/shared/config/app-config'
import { useAuthStore } from '@/features/auth/auth.store'

const useStyles = makeStyles({
  loading: {
    minHeight: '100vh',
    display: 'grid',
    placeItems: 'center',
  },
})

export function RequireAuth({ children }: PropsWithChildren) {
  const styles = useStyles()
  const status = useAuthStore((state) => state.status)
  const location = useLocation()

  if (status === 'idle' || status === 'bootstrapping') {
    return (
      <div className={styles.loading}>
        <Spinner label="正在校验登录状态" />
      </div>
    )
  }

  if (status !== 'authenticated') {
    const from = `${location.pathname}${location.search}${location.hash}`
    return <Navigate replace state={{ from }} to={`/login?redirect=${encodeURIComponent(from)}`} />
  }

  return children ?? <Outlet />
}

export function RedirectIfAuthenticated({ children }: PropsWithChildren) {
  const styles = useStyles()
  const status = useAuthStore((state) => state.status)

  if (status === 'idle' || status === 'bootstrapping') {
    return (
      <div className={styles.loading}>
        <Spinner label="正在恢复登录状态" />
      </div>
    )
  }

  if (status === 'authenticated') {
    return <Navigate replace to={appConfig.defaultRoute} />
  }

  return children ?? <Outlet />
}
