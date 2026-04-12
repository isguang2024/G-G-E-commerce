import { useUserStore } from '@/domains/auth/store'

export function useLogoutFlow() {
  const userStore = useUserStore()

  return {
    logout: () => userStore.logOut()
  }
}
