import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/domains/auth/store'
import { resolveActionKey } from '@/domains/governance/utils/action'

const userStore = useUserStore()

export const useAuth = () => {
  const { info } = storeToRefs(userStore)
  const actionSet = computed(() => new Set(info.value?.actions ?? []))

  const hasAction = (action: string): boolean => {
    const userInfo = info.value
    if (userInfo?.is_super_admin) {
      return true
    }
    return actionSet.value.has(resolveActionKey(action).key)
  }

  return {
    hasAction
  }
}
