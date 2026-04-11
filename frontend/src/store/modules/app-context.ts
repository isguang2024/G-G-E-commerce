import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { normalizeManagedAppKey } from '@/hooks/business/managed-app-scope'
import { writeActiveAppScopeKey } from '@/utils/app-scope'

type AppCapabilities = Record<string, any>

export const useAppContextStore = defineStore(
  'appContextStore',
  () => {
    const runtimeAppKey = ref('')
    const managedAppKey = ref('')
    const runtimeFrontendEntryURL = ref('')
    const runtimeBackendEntryURL = ref('')
    const runtimeHealthCheckURL = ref('')
    const appCapabilitiesMap = ref<Record<string, AppCapabilities>>({})

    const currentRuntimeAppKey = computed(() => normalizeManagedAppKey(runtimeAppKey.value))
    const currentManagedAppKey = computed(() => normalizeManagedAppKey(managedAppKey.value))
    const currentRuntimeFrontendEntryURL = computed(() => `${runtimeFrontendEntryURL.value || ''}`.trim())
    const currentRuntimeBackendEntryURL = computed(() => `${runtimeBackendEntryURL.value || ''}`.trim())
    const currentRuntimeHealthCheckURL = computed(() => `${runtimeHealthCheckURL.value || ''}`.trim())
    const effectiveManagedAppKey = computed(
      () => currentManagedAppKey.value || currentRuntimeAppKey.value
    )

    const setRuntimeAppKey = (value?: string | null) => {
      runtimeAppKey.value = normalizeManagedAppKey(value)
      writeActiveAppScopeKey(runtimeAppKey.value || managedAppKey.value)
    }

    const setManagedAppKey = (value?: string | null) => {
      managedAppKey.value = normalizeManagedAppKey(value)
      writeActiveAppScopeKey(managedAppKey.value || runtimeAppKey.value)
    }

    const setActiveAppKey = (value?: string | null) => {
      const nextAppKey = normalizeManagedAppKey(value)
      runtimeAppKey.value = nextAppKey
      managedAppKey.value = nextAppKey
      writeActiveAppScopeKey(nextAppKey)
    }

    const setAppCapabilities = (appKey: string, capabilities?: AppCapabilities | null) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return
      appCapabilitiesMap.value = {
        ...appCapabilitiesMap.value,
        [normalizedAppKey]:
          capabilities && typeof capabilities === 'object' && !Array.isArray(capabilities)
            ? capabilities
            : {}
      }
    }

    const setRuntimeAppContext = (payload: {
      appKey?: string | null
      frontendEntryUrl?: string | null
      backendEntryUrl?: string | null
      healthCheckUrl?: string | null
      capabilities?: AppCapabilities | null
    }) => {
      const normalizedAppKey = normalizeManagedAppKey(payload?.appKey)
      if (!normalizedAppKey) return
      runtimeAppKey.value = normalizedAppKey
      runtimeFrontendEntryURL.value = `${payload?.frontendEntryUrl || ''}`.trim()
      runtimeBackendEntryURL.value = `${payload?.backendEntryUrl || ''}`.trim()
      runtimeHealthCheckURL.value = `${payload?.healthCheckUrl || ''}`.trim()
      setAppCapabilities(normalizedAppKey, payload?.capabilities)
      writeActiveAppScopeKey(normalizedAppKey)
    }

    const currentAppCapabilities = computed<AppCapabilities>(() => {
      const appKey = effectiveManagedAppKey.value
      if (!appKey) return {}
      return appCapabilitiesMap.value[appKey] || {}
    })

    const ensureManagedAppKey = () => {
      return currentManagedAppKey.value
    }

    const clearAppContext = () => {
      runtimeAppKey.value = ''
      managedAppKey.value = ''
      runtimeFrontendEntryURL.value = ''
      runtimeBackendEntryURL.value = ''
      runtimeHealthCheckURL.value = ''
      writeActiveAppScopeKey('')
    }

    watch(
      effectiveManagedAppKey,
      (value) => {
        writeActiveAppScopeKey(value || currentRuntimeAppKey.value)
      },
      { immediate: true }
    )

    return {
      // 暴露原始 ref 以便 pinia-plugin-persistedstate 可写入持久化
      runtimeAppKey,
      managedAppKey,
      runtimeFrontendEntryURL,
      runtimeBackendEntryURL,
      runtimeHealthCheckURL,
      // 规范化只读视图
      currentRuntimeAppKey,
      currentManagedAppKey,
      currentRuntimeFrontendEntryURL,
      currentRuntimeBackendEntryURL,
      currentRuntimeHealthCheckURL,
      effectiveManagedAppKey,
      currentAppCapabilities,
      setRuntimeAppKey,
      setManagedAppKey,
      setActiveAppKey,
      setAppCapabilities,
      setRuntimeAppContext,
      ensureManagedAppKey,
      clearAppContext
    }
  },
  {
    persist: {
      key: 'appContextStore',
      storage: localStorage,
      pick: [
        'runtimeAppKey',
        'managedAppKey',
        'runtimeFrontendEntryURL',
        'runtimeBackendEntryURL',
        'runtimeHealthCheckURL',
        'appCapabilitiesMap'
      ]
    }
  }
)
