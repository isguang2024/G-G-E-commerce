import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/store/modules/user'
import { resolveActionKey } from '@/utils/permission/action'

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
