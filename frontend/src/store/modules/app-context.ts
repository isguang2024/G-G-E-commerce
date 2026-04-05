import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

function normalizeAppKey(value?: string | null) {
  return `${value || ''}`.trim().toLowerCase()
}

export const useAppContextStore = defineStore(
  'appContextStore',
  () => {
    const runtimeAppKey = ref('')
    const managedAppKey = ref('')

    const currentRuntimeAppKey = computed(() => normalizeAppKey(runtimeAppKey.value))
    const currentManagedAppKey = computed(() => normalizeAppKey(managedAppKey.value))

    const setRuntimeAppKey = (value?: string | null) => {
      runtimeAppKey.value = normalizeAppKey(value)
      if (!currentManagedAppKey.value && runtimeAppKey.value) {
        managedAppKey.value = runtimeAppKey.value
      }
    }

    const setManagedAppKey = (value?: string | null) => {
      managedAppKey.value = normalizeAppKey(value)
    }

    const ensureManagedAppKey = (fallback?: string | null) => {
      const nextValue =
        currentManagedAppKey.value || normalizeAppKey(fallback) || currentRuntimeAppKey.value
      if (nextValue && nextValue !== currentManagedAppKey.value) {
        managedAppKey.value = nextValue
      }
      return nextValue
    }

    const clearAppContext = () => {
      runtimeAppKey.value = ''
      managedAppKey.value = ''
    }

    return {
      runtimeAppKey: currentRuntimeAppKey,
      managedAppKey: currentManagedAppKey,
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
