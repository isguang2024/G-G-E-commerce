import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { normalizeManagedAppKey } from '@/hooks/business/managed-app-scope'
import { writeActiveAppScopeKey } from '@/utils/app-scope'

type AppCapabilities = Record<string, any>
type AppMeta = Record<string, any>

function normalizePlainObject(value?: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, any>)
    : {}
}

function readStringCandidate(record: Record<string, any>, ...keys: string[]): string {
  for (const key of keys) {
    const value = `${record?.[key] ?? ''}`.trim()
    if (value) return value
  }
  return ''
}

export const useAppContextStore = defineStore(
  'appContextStore',
  () => {
    const runtimeAppKey = ref('')
    const managedAppKey = ref('')
    const runtimeFrontendEntryURL = ref('')
    const runtimeBackendEntryURL = ref('')
    const runtimeHealthCheckURL = ref('')
    const appAuthModeMap = ref<Record<string, string>>({})
    const appCapabilitiesMap = ref<Record<string, AppCapabilities>>({})
    const appMetaMap = ref<Record<string, AppMeta>>({})

    const resolveRuntimeEnvProfile = () => {
      if (typeof window === 'undefined') {
        return `${import.meta.env.MODE || ''}`.trim() || 'default'
      }
      const host = `${window.location.hostname || ''}`.toLowerCase()
      const mode = `${import.meta.env.MODE || ''}`.trim()
      if (mode) return mode
      if (host === '127.0.0.1' || host === 'localhost') return 'local'
      if (host.includes('test')) return 'test'
      if (host.includes('staging') || host.includes('pre')) return 'staging'
      return 'default'
    }

    const resolveProfileObject = (appKey?: string | null): Record<string, any> => {
      const meta = resolveAppMeta(appKey)
      const envProfiles = normalizePlainObject(meta.env_profiles)
      if (!Object.keys(envProfiles).length) return {}
      const profileKey = resolveRuntimeEnvProfile()
      const defaultProfile = normalizePlainObject(envProfiles.default)
      const namedProfile = normalizePlainObject(
        envProfiles[profileKey] || envProfiles[profileKey.toLowerCase()]
      )
      return {
        ...defaultProfile,
        ...namedProfile
      }
    }

    const currentRuntimeAppKey = computed(() => normalizeManagedAppKey(runtimeAppKey.value))
    const currentManagedAppKey = computed(() => normalizeManagedAppKey(managedAppKey.value))
    const effectiveManagedAppKey = computed(
      () => currentManagedAppKey.value || currentRuntimeAppKey.value
    )
    const currentRuntimeFrontendEntryURL = computed(() => {
      const profile = resolveProfileObject(effectiveManagedAppKey.value || currentRuntimeAppKey.value)
      return (
        `${runtimeFrontendEntryURL.value || ''}`.trim() ||
        readStringCandidate(profile, 'frontend_entry_url', 'frontendEntryUrl')
      )
    })
    const currentRuntimeBackendEntryURL = computed(() => {
      const profile = resolveProfileObject(effectiveManagedAppKey.value || currentRuntimeAppKey.value)
      return (
        `${runtimeBackendEntryURL.value || ''}`.trim() ||
        readStringCandidate(profile, 'backend_entry_url', 'backendEntryUrl')
      )
    })
    const currentRuntimeHealthCheckURL = computed(() => {
      const profile = resolveProfileObject(effectiveManagedAppKey.value || currentRuntimeAppKey.value)
      return (
        `${runtimeHealthCheckURL.value || ''}`.trim() ||
        readStringCandidate(profile, 'health_check_url', 'healthCheckUrl')
      )
    })

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

    const setAppMeta = (appKey: string, meta?: AppMeta | null) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return
      appMetaMap.value = {
        ...appMetaMap.value,
        [normalizedAppKey]: normalizePlainObject(meta)
      }
    }

    const setAppAuthMode = (appKey: string, authMode?: string | null) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return
      appAuthModeMap.value = {
        ...appAuthModeMap.value,
        [normalizedAppKey]: `${authMode || ''}`.trim() || 'inherit_host'
      }
    }

    const setAppProfile = (payload: {
      appKey?: string | null
      authMode?: string | null
      capabilities?: AppCapabilities | null
      meta?: AppMeta | null
    }) => {
      const normalizedAppKey = normalizeManagedAppKey(payload?.appKey)
      if (!normalizedAppKey) return
      setAppAuthMode(normalizedAppKey, payload?.authMode)
      setAppCapabilities(normalizedAppKey, payload?.capabilities)
      setAppMeta(normalizedAppKey, payload?.meta)
    }

    const setRuntimeAppContext = (payload: {
      appKey?: string | null
      frontendEntryUrl?: string | null
      backendEntryUrl?: string | null
      healthCheckUrl?: string | null
      authMode?: string | null
      capabilities?: AppCapabilities | null
      meta?: AppMeta | null
    }) => {
      const normalizedAppKey = normalizeManagedAppKey(payload?.appKey)
      if (!normalizedAppKey) return
      runtimeAppKey.value = normalizedAppKey
      runtimeFrontendEntryURL.value = `${payload?.frontendEntryUrl || ''}`.trim()
      runtimeBackendEntryURL.value = `${payload?.backendEntryUrl || ''}`.trim()
      runtimeHealthCheckURL.value = `${payload?.healthCheckUrl || ''}`.trim()
      setAppProfile({
        appKey: normalizedAppKey,
        authMode: payload?.authMode,
        capabilities: payload?.capabilities,
        meta: payload?.meta
      })
      writeActiveAppScopeKey(normalizedAppKey)
    }

    const resolveAppAuthMode = (appKey?: string | null) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return ''
      return `${appAuthModeMap.value[normalizedAppKey] || ''}`.trim()
    }

    const resolveAppCapabilities = (appKey?: string | null): AppCapabilities => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return {}
      return appCapabilitiesMap.value[normalizedAppKey] || {}
    }

    const resolveAppMeta = (appKey?: string | null): AppMeta => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return {}
      return appMetaMap.value[normalizedAppKey] || {}
    }

    const isFeatureEnabledForApp = (appKey: string | null | undefined, flagKey: string) => {
      const normalizedFlagKey = `${flagKey || ''}`.trim()
      if (!normalizedFlagKey) return false
      const featureFlags = normalizePlainObject(resolveAppMeta(appKey).feature_flags)
      const rawValue = featureFlags[normalizedFlagKey]
      if (typeof rawValue === 'boolean') {
        return rawValue
      }
      const profileOverrides = normalizePlainObject(rawValue)
      const profileKey = resolveRuntimeEnvProfile()
      if (typeof profileOverrides[profileKey] === 'boolean') {
        return profileOverrides[profileKey]
      }
      if (typeof profileOverrides.default === 'boolean') {
        return profileOverrides.default
      }
      return false
    }

    const supportsCapabilityForApp = (
      appKey: string | null | undefined,
      groupKey: string,
      capabilityKey: string,
      fallback = true
    ) => {
      const capabilities = normalizePlainObject(resolveAppCapabilities(appKey))
      const group = normalizePlainObject(capabilities[groupKey])
      const raw = group[capabilityKey]
      if (typeof raw === 'boolean') {
        return raw
      }
      return fallback
    }

    const supportsAppSwitchForApp = (appKey?: string | null) => {
      if (isFeatureEnabledForApp(appKey, 'app_switcher')) {
        return true
      }
      return supportsCapabilityForApp(appKey, 'integration', 'supports_app_switch', true)
    }

    const supportsDynamicRoutesForApp = (appKey?: string | null) => {
      if (isFeatureEnabledForApp(appKey, 'disable_dynamic_routes')) {
        return false
      }
      return supportsCapabilityForApp(appKey, 'runtime', 'supports_dynamic_routes', true)
    }

    const shouldUseCentralizedLoginForApp = (appKey?: string | null) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return false

      const authMode = resolveAppAuthMode(normalizedAppKey)
      if (authMode === 'centralized_login') {
        return true
      }
      if (authMode === 'shared_cookie') {
        return false
      }

      const capabilities = resolveAppCapabilities(normalizedAppKey)
      const authConfig =
        capabilities && typeof capabilities.auth === 'object' && !Array.isArray(capabilities.auth)
          ? capabilities.auth
          : {}
      const loginStrategy = `${authConfig?.login_strategy || authConfig?.loginStrategy || ''}`.trim()
      if (loginStrategy === 'centralized_login') {
        return true
      }
      if (loginStrategy === 'local' || loginStrategy === 'shared_cookie') {
        return false
      }
      if (authConfig?.is_auth_center === true || authConfig?.isAuthCenter === true) {
        return false
      }
      return false
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
      appMetaMap.value = {}
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
      appAuthModeMap,
      appMetaMap,
      setRuntimeAppKey,
      setManagedAppKey,
      setActiveAppKey,
      setAppAuthMode,
      setAppProfile,
      setAppCapabilities,
      setAppMeta,
      setRuntimeAppContext,
      resolveAppAuthMode,
      resolveAppCapabilities,
      resolveAppMeta,
      isFeatureEnabledForApp,
      supportsAppSwitchForApp,
      supportsDynamicRoutesForApp,
      shouldUseCentralizedLoginForApp,
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
        'appAuthModeMap',
        'appCapabilitiesMap',
        'appMetaMap'
      ]
    }
  }
)
