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
  const status = useAuthStore((state) => state.status)
  const session = useAuthStore((state) => state.session)

  useEffect(() => {
    if (startedRef.current || status !== 'idle') {
      return
    }

    startedRef.current = true

    if (!session?.accessToken) {
      useAuthStore.getState().completeBootstrap({ session: null, currentUser: null })
      return
    }

    useAuthStore.getState().beginBootstrap()
    fetchCurrentUser()
      .then((currentUser) => {
        useAuthStore.getState().completeBootstrap({
          session,
          currentUser,
          rememberMe: useAuthStore.getState().rememberMe,
        })
      })
      .catch(() => {
        useAuthStore.getState().clearAuth()
      })
  }, [session, status])

  if (status === 'idle' || status === 'bootstrapping') {
    return (
      <div className={styles.fallback}>
        <Spinner label="正在初始化登录态" />
      </div>
    )
  }

  return children
}
