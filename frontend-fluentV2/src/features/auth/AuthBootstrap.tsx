import type { PropsWithChildren } from 'react'
import { useEffect, useRef } from 'react'
import { Spinner, makeStyles } from '@fluentui/react-components'
import { fetchCurrentUser } from '@/shared/api/modules/auth.api'
import { useAuthStore } from '@/features/auth/auth.store'

const useStyles = makeStyles({
  fallback: {
    minHeight: '100vh',
    display: 'grid',
    placeItems: 'center',
  },
})

export function AuthBootstrap({ children }: PropsWithChildren) {
  const styles = useStyles()
  const startedRef = useRef(false)
  const { status, session, beginRestore, completeRestore } = useAuthStore((state) => ({
    status: state.status,
    session: state.session,
    beginRestore: state.beginRestore,
    completeRestore: state.completeRestore,
  }))

  useEffect(() => {
    if (startedRef.current || status !== 'idle') {
      return
    }

    startedRef.current = true

    if (!session?.accessToken) {
      completeRestore({ session: null, currentUser: null })
      return
    }

    beginRestore()
    fetchCurrentUser()
      .then((currentUser) => {
        completeRestore({
          session,
          currentUser,
          rememberMe: useAuthStore.getState().rememberMe,
        })
      })
      .catch(() => {
        useAuthStore.getState().clearAuth()
      })
  }, [beginRestore, completeRestore, session, status])

  if (status === 'idle' || status === 'restoring') {
    return (
      <div className={styles.fallback}>
        <Spinner label="正在恢复会话" />
      </div>
    )
  }

  return children
}
