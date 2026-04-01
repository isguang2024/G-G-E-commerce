import { create } from 'zustand'
import { buildMockSessionUser } from '@/shared/mocks/session.mock'
import type { SessionUser } from '@/shared/types/session'

const LOCAL_SESSION_KEY = 'frontend-fluentV2.auth.local'
const SESSION_STORAGE_KEY = 'frontend-fluentV2.auth.session'

type StoredAuthState = {
  authenticated: boolean
  currentUser: SessionUser | null
  rememberMe: boolean
}

type AuthState = StoredAuthState & {
  signIn: (payload: { account: string; rememberMe: boolean }) => SessionUser
  signOut: () => void
}

function getStorageSnapshot() {
  if (typeof window === 'undefined') {
    return null
  }

  const candidates = [window.localStorage.getItem(LOCAL_SESSION_KEY), window.sessionStorage.getItem(SESSION_STORAGE_KEY)]

  for (const rawValue of candidates) {
    if (!rawValue) {
      continue
    }

    try {
      const parsed = JSON.parse(rawValue) as StoredAuthState
      if (parsed?.authenticated && parsed.currentUser) {
        return parsed
      }
    } catch {
      continue
    }
  }

  return null
}

function writeStorageSnapshot(state: StoredAuthState) {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.removeItem(LOCAL_SESSION_KEY)
  window.sessionStorage.removeItem(SESSION_STORAGE_KEY)

  if (!state.authenticated || !state.currentUser) {
    return
  }

  const target = state.rememberMe ? window.localStorage : window.sessionStorage
  target.setItem(
    state.rememberMe ? LOCAL_SESSION_KEY : SESSION_STORAGE_KEY,
    JSON.stringify(state),
  )
}

const initialState = getStorageSnapshot() ?? {
  authenticated: false,
  currentUser: null,
  rememberMe: false,
}

export const useAuthStore = create<AuthState>((set) => ({
  ...initialState,
  signIn: ({ account, rememberMe }) => {
    const currentUser = buildMockSessionUser(account)
    const nextState: StoredAuthState = {
      authenticated: true,
      currentUser,
      rememberMe,
    }

    writeStorageSnapshot(nextState)
    set(nextState)
    return currentUser
  },
  signOut: () => {
    const nextState: StoredAuthState = {
      authenticated: false,
      currentUser: null,
      rememberMe: false,
    }

    writeStorageSnapshot(nextState)
    set(nextState)
  },
}))
