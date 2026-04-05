import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppContextStore } from '@/store/modules/app-context'

function normalizeAppKey(value?: string | null) {
  return `${value || ''}`.trim().toLowerCase()
}

export function useManagedAppScope(options?: { fallbackAppKey?: string; syncRoute?: boolean }) {
  const route = useRoute()
  const router = useRouter()
  const appContextStore = useAppContextStore()
  const syncRoute = options?.syncRoute !== false

  const targetAppKey = computed(() => {
    const routeAppKey = normalizeAppKey(`${route.query.app_key || ''}`)
    if (routeAppKey) return routeAppKey
    return appContextStore.ensureManagedAppKey(options?.fallbackAppKey)
  })

  const ensureManagedAppRoute = async () => {
    const routeAppKey = normalizeAppKey(`${route.query.app_key || ''}`)
    if (routeAppKey) {
      appContextStore.setManagedAppKey(routeAppKey)
      return routeAppKey
    }
    const nextAppKey = appContextStore.ensureManagedAppKey(options?.fallbackAppKey)
    if (!nextAppKey) return ''
    if (syncRoute) {
      await router.replace({
        path: route.path,
        query: {
          ...route.query,
          app_key: nextAppKey
        },
        hash: route.hash
      })
    }
    return nextAppKey
  }

  const setManagedAppKey = async (value?: string | null) => {
    const nextAppKey = normalizeAppKey(value)
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
      const normalized = normalizeAppKey(value)
      if (normalized) {
        appContextStore.setManagedAppKey(normalized)
        return
      }
      if (syncRoute) {
        void ensureManagedAppRoute()
      }
    },
    { immediate: true }
  )

  onMounted(() => {
    if (syncRoute) {
      void ensureManagedAppRoute()
    }
  })

  return {
    targetAppKey,
    ensureManagedAppRoute,
    setManagedAppKey
  }
}
