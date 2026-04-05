import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppContextStore } from '@/store/modules/app-context'
import {
  normalizeManagedAppKey,
  resolveManagedAppKey
} from '@/hooks/business/managed-app-scope'

export function useManagedAppScope(options?: { syncRoute?: boolean }) {
  const route = useRoute()
  const router = useRouter()
  const appContextStore = useAppContextStore()
  const syncRoute = options?.syncRoute === true

  const targetAppKey = computed(() => {
    return resolveManagedAppKey(
      route.query.app_key as string | undefined,
      appContextStore.managedAppKey || appContextStore.runtimeAppKey
    )
  })

  const ensureManagedAppRoute = async () => {
    const routeAppKey = normalizeManagedAppKey(route.query.app_key as string | undefined)
    if (routeAppKey) {
      appContextStore.setManagedAppKey(routeAppKey)
      return routeAppKey
    }
    const fallbackAppKey = resolveManagedAppKey(
      '',
      appContextStore.managedAppKey || appContextStore.runtimeAppKey
    )
    if (!fallbackAppKey) {
      return ''
    }
    appContextStore.setManagedAppKey(fallbackAppKey)
    return fallbackAppKey
  }

  const setManagedAppKey = async (value?: string | null) => {
    const nextAppKey = normalizeManagedAppKey(value)
    appContextStore.setManagedAppKey(nextAppKey)
    if (!syncRoute || !nextAppKey || route.query.app_key === nextAppKey) {
      return nextAppKey
    }
    await router.replace({
      path: route.path,
      query: {
        ...route.query,
        app_key: nextAppKey
      },
      hash: route.hash
    })
    return nextAppKey
  }

  watch(
    () => `${route.query.app_key || ''}`,
    (value) => {
      const normalized = normalizeManagedAppKey(value)
      if (normalized) {
        appContextStore.setManagedAppKey(normalized)
        return
      }
      const fallbackAppKey = resolveManagedAppKey(
        '',
        appContextStore.managedAppKey || appContextStore.runtimeAppKey
      )
      if (fallbackAppKey) {
        appContextStore.setManagedAppKey(fallbackAppKey)
      }
    },
    { immediate: true }
  )

  return {
    targetAppKey,
    ensureManagedAppRoute,
    setManagedAppKey
  }
}
