import { create } from 'zustand'
import type { AuthSession, CurrentUser, TenantContext } from '@/shared/types/auth'
import { clearStoredAuthSnapshot, persistAuthSnapshot, readStoredAuthSnapshot } from '@/features/auth/auth.storage'

export type AuthStatus = 'idle' | 'bootstrapping' | 'authenticated' | 'guest'

type AuthState = {
  status: AuthStatus
  session: AuthSession | null
  currentUser: CurrentUser | null
  tenantContext: TenantContext
  rememberMe: boolean
  beginBootstrap: () => void
  completeBootstrap: (payload: {
    session: AuthSession | null
    currentUser: CurrentUser | null
    rememberMe?: boolean
  }) => void
  setAuthenticated: (payload: {
    session: AuthSession
    currentUser: CurrentUser
    rememberMe: boolean
  }) => void
  updateSession: (session: AuthSession, rememberMe?: boolean) => void
  updateCurrentUser: (currentUser: CurrentUser) => void
  clearAuth: () => void
}

const storedSnapshot = readStoredAuthSnapshot()

export const useAuthStore = create<AuthState>((set) => ({
  status: storedSnapshot?.session ? 'idle' : 'guest',
  session: storedSnapshot?.session || null,
  currentUser: null,
  tenantContext: {
    currentTenantId: null,
  },
  rememberMe: storedSnapshot?.rememberMe ?? true,
  beginBootstrap: () => set({ status: 'bootstrapping' }),
  completeBootstrap: ({ session, currentUser, rememberMe }) =>
    set({
      status: session && currentUser ? 'authenticated' : 'guest',
      session,
      currentUser,
      rememberMe: rememberMe ?? false,
      tenantContext: {
        currentTenantId: currentUser?.currentTenantId || null,
      },
    }),
  setAuthenticated: ({ session, currentUser, rememberMe }) => {
    persistAuthSnapshot(session, rememberMe)
    set({
      status: 'authenticated',
      session,
      currentUser,
      rememberMe,
      tenantContext: {
        currentTenantId: currentUser.currentTenantId,
      },
    })
  },
  updateSession: (session, rememberMe) => {
    const nextRememberMe = rememberMe ?? useAuthStore.getState().rememberMe
    persistAuthSnapshot(session, nextRememberMe)
    set((state) => ({
      ...state,
      session,
      rememberMe: nextRememberMe,
      status: state.currentUser ? 'authenticated' : state.status,
    }))
  },
  updateCurrentUser: (currentUser) =>
    set((state) => ({
      ...state,
      currentUser,
      status: state.session ? 'authenticated' : state.status,
      tenantContext: {
        currentTenantId: currentUser.currentTenantId,
      },
    })),
  clearAuth: () => {
    clearStoredAuthSnapshot()
    set({
      status: 'guest',
      session: null,
      currentUser: null,
      rememberMe: false,
      tenantContext: {
        currentTenantId: null,
      },
    })
  },
}))
