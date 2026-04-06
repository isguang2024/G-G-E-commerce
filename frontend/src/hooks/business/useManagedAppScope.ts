import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAppContextStore } from '@/store/modules/app-context'
import {
  normalizeManagedAppKey,
  resolveManagedAppKey,
  resolveManagedAppStorageKey
} from '@/hooks/business/managed-app-scope'

export function useManagedAppScope(options?: { syncRoute?: boolean; storageKey?: string }) {
  const route = useRoute()
  const appContextStore = useAppContextStore()
  const storageKey = resolveManagedAppStorageKey(options?.storageKey, route.name, route.path)
  const initialAppKey = resolveManagedAppKey(
    typeof window !== 'undefined' ? window.localStorage.getItem(storageKey) : '',
    appContextStore.runtimeAppKey
  )
  const localManagedAppKey = ref(initialAppKey)

  const targetAppKey = computed(() => resolveManagedAppKey(localManagedAppKey.value))

  const ensureManagedAppRoute = async () => {
    const fallbackAppKey = resolveManagedAppKey(
      localManagedAppKey.value,
      appContextStore.runtimeAppKey
    )
    if (!fallbackAppKey) {
      return ''
    }
    localManagedAppKey.value = fallbackAppKey
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(storageKey, fallbackAppKey)
    }
    return fallbackAppKey
  }

  const setManagedAppKey = async (value?: string | null) => {
    const nextAppKey = normalizeManagedAppKey(value)
    localManagedAppKey.value = nextAppKey
    if (typeof window !== 'undefined') {
      if (nextAppKey) {
        window.localStorage.setItem(storageKey, nextAppKey)
      } else {
        window.localStorage.removeItem(storageKey)
      }
    }
    return nextAppKey
  }

  return {
    targetAppKey,
    ensureManagedAppRoute,
    setManagedAppKey
  }
}
