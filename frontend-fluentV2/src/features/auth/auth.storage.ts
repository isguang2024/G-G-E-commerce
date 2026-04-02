import type { AuthSession } from '@/shared/types/auth'
import { readJsonStorage, removeStorageKeys, writeJsonStorage } from '@/shared/lib/storage'

const AUTH_LOCAL_KEY = 'frontend-fluentV2.auth.local'
const AUTH_SESSION_KEY = 'frontend-fluentV2.auth.session'

interface StoredAuthSnapshot {
  session: AuthSession
  rememberMe: boolean
}

function canUseStorage() {
  return typeof window !== 'undefined'
}

export function readStoredAuthSnapshot(): StoredAuthSnapshot | null {
  if (!canUseStorage()) {
    return null
  }

  return (
    readJsonStorage<StoredAuthSnapshot>(window.localStorage, AUTH_LOCAL_KEY) ||
    readJsonStorage<StoredAuthSnapshot>(window.sessionStorage, AUTH_SESSION_KEY)
  )
}

export function persistAuthSnapshot(session: AuthSession, rememberMe: boolean) {
  if (!canUseStorage()) {
    return
  }

  clearStoredAuthSnapshot()
  const payload: StoredAuthSnapshot = { session, rememberMe }
  if (rememberMe) {
    writeJsonStorage(window.localStorage, AUTH_LOCAL_KEY, payload)
    return
  }

  writeJsonStorage(window.sessionStorage, AUTH_SESSION_KEY, payload)
}

export function clearStoredAuthSnapshot() {
  if (!canUseStorage()) {
    return
  }

  removeStorageKeys(window.localStorage, [AUTH_LOCAL_KEY])
  removeStorageKeys(window.sessionStorage, [AUTH_SESSION_KEY])
}
