import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/store/modules/user'
import { hasScopedActionPermission } from '@/utils/permission/action'

const userStore = useUserStore()

export const useAuth = () => {
  const { info } = storeToRefs(userStore)
  const actionSet = computed(() => new Set(info.value?.actions ?? []))
  const scopedActionSet = computed(
    () => new Set(info.value?.scoped_actions || info.value?.scopedActions || [])
  )

  const hasAction = (action: string, scopeCode?: string): boolean => {
    if (info.value?.is_super_admin) {
      return true
    }
    if (scopeCode || action.includes('@')) {
      return hasScopedActionPermission(
        {
          ...info.value,
          actions: Array.from(actionSet.value),
          scoped_actions: Array.from(scopedActionSet.value)
        },
        action,
        scopeCode
      )
    }
    return actionSet.value.has(action)
  }

  return {
    hasAction
  }
}
