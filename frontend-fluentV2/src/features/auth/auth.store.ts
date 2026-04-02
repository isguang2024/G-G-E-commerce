import { create } from 'zustand'
import type { AuthSession, CurrentUser, TenantContext } from '@/shared/types/auth'
import { clearStoredAuthSnapshot, persistAuthSnapshot, readStoredAuthSnapshot } from '@/features/auth/auth.storage'

export type AuthStatus = 'idle' | 'restoring' | 'authenticated' | 'anonymous'

type AuthState = {
  status: AuthStatus
  session: AuthSession | null
  currentUser: CurrentUser | null
  tenantContext: TenantContext
  rememberMe: boolean
  beginRestore: () => void
  completeRestore: (payload: {
    session: AuthSession | null
    currentUser: CurrentUser | null
    rememberMe?: boolean
  }) => void
  setAuthenticated: (payload: {
    session: AuthSession
    currentUser: CurrentUser
    rememberMe: boolean
  }) => void
  updateCurrentUser: (currentUser: CurrentUser) => void
  clearAuth: () => void
}

const storedSnapshot = readStoredAuthSnapshot()

export const useAuthStore = create<AuthState>((set) => ({
  status: storedSnapshot?.session ? 'idle' : 'anonymous',
  session: storedSnapshot?.session || null,
  currentUser: null,
  tenantContext: {
    currentTenantId: null,
  },
  rememberMe: storedSnapshot?.rememberMe ?? true,
  beginRestore: () => set({ status: 'restoring' }),
  completeRestore: ({ session, currentUser, rememberMe }) =>
    set({
      status: session && currentUser ? 'authenticated' : 'anonymous',
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
      status: 'anonymous',
      session: null,
      currentUser: null,
      rememberMe: false,
      tenantContext: {
        currentTenantId: null,
      },
    })
  },
}))
