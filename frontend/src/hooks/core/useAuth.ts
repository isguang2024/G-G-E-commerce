import { storeToRefs } from 'pinia'
import { useUserStore } from '@/store/modules/user'

const userStore = useUserStore()

export const useAuth = () => {
  const { info } = storeToRefs(userStore)

  const hasAction = (action: string): boolean => {
    if (info.value?.is_super_admin) {
      return true
    }
    return (info.value?.actions ?? []).includes(action)
  }

  return {
    hasAction
  }
}
