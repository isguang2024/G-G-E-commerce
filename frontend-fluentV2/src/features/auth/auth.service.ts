import { useMutation, useQuery } from '@tanstack/react-query'
import { fetchCurrentUser, loginWithPassword, registerWithPassword } from '@/shared/api/modules/auth.api'
import { queryKeys } from '@/shared/api/query-keys'
import { clearStoredAuthSnapshot, persistAuthSnapshot } from '@/features/auth/auth.storage'

export function useCurrentUserQuery(enabled: boolean) {
  return useQuery({
    queryKey: queryKeys.auth.currentUser,
    queryFn: fetchCurrentUser,
    enabled,
    retry: false,
  })
}

export function useLoginMutation() {
  return useMutation({
    mutationFn: async (payload: { username: string; password: string; rememberMe: boolean }) => {
      const loginResult = await loginWithPassword(payload)
      persistAuthSnapshot(loginResult.session, payload.rememberMe)
      try {
        const currentUser = await fetchCurrentUser()
        return {
          session: loginResult.session,
          loginUser: currentUser,
        }
      } catch (error) {
        clearStoredAuthSnapshot()
        throw error
      }
    },
  })
}

export function useRegisterMutation() {
  return useMutation({
    mutationFn: async (payload: {
      username: string
      password: string
      email?: string
      nickname?: string
      rememberMe: boolean
    }) => {
      const registerResult = await registerWithPassword(payload)
      persistAuthSnapshot(registerResult.session, payload.rememberMe)
      try {
        const currentUser = await fetchCurrentUser()
        return {
          session: registerResult.session,
          loginUser: currentUser,
        }
      } catch (error) {
        clearStoredAuthSnapshot()
        throw error
      }
    },
  })
}
