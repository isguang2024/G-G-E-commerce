import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { normalizeManagedAppKey } from '@/hooks/business/managed-app-scope'

export const useAppContextStore = defineStore(
  'appContextStore',
  () => {
    const runtimeAppKey = ref('')
    const managedAppKey = ref('')

    const currentRuntimeAppKey = computed(() => normalizeManagedAppKey(runtimeAppKey.value))
    const currentManagedAppKey = computed(() => normalizeManagedAppKey(managedAppKey.value))
    const effectiveManagedAppKey = computed(() => currentManagedAppKey.value)

    const setRuntimeAppKey = (value?: string | null) => {
      runtimeAppKey.value = normalizeManagedAppKey(value)
    }

    const setManagedAppKey = (value?: string | null) => {
      managedAppKey.value = normalizeManagedAppKey(value)
    }

    const ensureManagedAppKey = () => {
      return currentManagedAppKey.value
    }

    const clearAppContext = () => {
      runtimeAppKey.value = ''
      managedAppKey.value = ''
    }

    return {
      runtimeAppKey: currentRuntimeAppKey,
      managedAppKey: currentManagedAppKey,
      effectiveManagedAppKey,
      setRuntimeAppKey,
      setManagedAppKey,
      ensureManagedAppKey,
      clearAppContext
    }
  },
  {
    persist: true
  }
)
