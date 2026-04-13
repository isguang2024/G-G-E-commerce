import { computed, ref, watch } from 'vue'
import { defineStore } from 'pinia'
import { normalizeManagedAppKey } from '@/domains/app-runtime/managed-app-scope'
import {
  ensureRuntimeAppKeyViaHandler,
  switchAppViaHandler,
  type SwitchAppPayload
} from '@/domains/app-runtime/runtime/context-handlers'
import { writeActiveAppScopeKey } from '@/domains/app-runtime/app-scope'
import { registerHttpAppContext } from '@/utils/http/request-context'

type AppCapabilities = Record<string, any>
type AppMeta = Record<string, any>

const APP_CONTEXT_STORE_SUFFIX = '-appContextStore'

function normalizePlainObject(value?: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, any>)
    : {}
}

function isLoopbackHost(hostname: string): boolean {
  const normalized = `${hostname || ''}`.toLowerCase()
  return normalized === 'localhost' || normalized === '127.0.0.1' || normalized === '::1'
}

function sanitizeRuntimeEntryURL(value?: string | null): string {
  const raw = `${value || ''}`.trim()
  if (!raw) return ''

  if (/^https?:\/\//i.test(raw)) {
    try {
      const target = new URL(raw)
      if (typeof window !== 'undefined') {
        const currentHost = `${window.location.hostname || ''}`.toLowerCase()
        // 防止历史缓存把线上请求重写到本机 localhost，导致登录后接口全部失效。
        if (isLoopbackHost(target.hostname) && !isLoopbackHost(currentHost)) {
          return ''
        }
      }
      return raw
    } catch {
      return ''
    }
  }

  if (raw.startsWith('/')) {
    return raw
  }

  return ''
}

function prunePersistedRuntimeEntryURLs(): void {
  if (typeof window === 'undefined') return

  for (const key of Object.keys(window.localStorage)) {
    if (!key.endsWith(APP_CONTEXT_STORE_SUFFIX)) {
      continue
    }

    const raw = window.localStorage.getItem(key)
    if (!raw) {
      continue
    }

    try {
      const parsed = JSON.parse(raw) as Record<string, unknown>
      let changed = false

      for (const field of [
        'runtimeFrontendEntryURL',
        'runtimeBackendEntryURL',
        'runtimeHealthCheckURL'
      ]) {
        if (field in parsed) {
          delete parsed[field]
          changed = true
        }
      }

      if (changed) {
        window.localStorage.setItem(key, JSON.stringify(parsed))
      }
    } catch {
      // ignore malformed persisted payload
    }
  }
}

prunePersistedRuntimeEntryURLs()

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

    const currentRuntimeAppKey = computed(() => normalizeManagedAppKey(runtimeAppKey.value))
    const currentManagedAppKey = computed(() => normalizeManagedAppKey(managedAppKey.value))
    const effectiveManagedAppKey = computed(
      () => currentManagedAppKey.value || currentRuntimeAppKey.value
    )
    const currentRuntimeFrontendEntryURL = computed(() => {
      return sanitizeRuntimeEntryURL(runtimeFrontendEntryURL.value)
    })
    const currentRuntimeBackendEntryURL = computed(() => {
      return sanitizeRuntimeEntryURL(runtimeBackendEntryURL.value)
    })
    const currentRuntimeHealthCheckURL = computed(() => {
      return sanitizeRuntimeEntryURL(runtimeHealthCheckURL.value)
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
      runtimeFrontendEntryURL.value = sanitizeRuntimeEntryURL(payload?.frontendEntryUrl)
      runtimeBackendEntryURL.value = sanitizeRuntimeEntryURL(payload?.backendEntryUrl)
      runtimeHealthCheckURL.value = sanitizeRuntimeEntryURL(payload?.healthCheckUrl)
      setAppProfile({
        appKey: normalizedAppKey,
        authMode: payload?.authMode,
        capabilities: payload?.capabilities,
        meta: payload?.meta
      })
      writeActiveAppScopeKey(normalizedAppKey)
    }

    registerHttpAppContext({
      getCurrentRuntimeBackendEntryURL: () => currentRuntimeBackendEntryURL.value,
      getEffectiveManagedAppKey: () => effectiveManagedAppKey.value,
      getCurrentRuntimeAppKey: () => currentRuntimeAppKey.value,
      resolveAppAuthMode: (appKey) => resolveAppAuthMode(appKey),
      resolveAppSsoMode: (appKey) => resolveAppSsoMode(appKey),
      resolveAppLoginPageKey: (appKey) => resolveAppLoginPageKey(appKey)
    })

    const ensureRuntimeAppKey = async () => ensureRuntimeAppKeyViaHandler()

    const switchApp = async (payload: SwitchAppPayload) => switchAppViaHandler(payload)

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
      return supportsCapabilityForApp(appKey, 'integration', 'supports_app_switch', true)
    }

    const supportsDynamicRoutesForApp = (appKey?: string | null) => {
      return supportsCapabilityForApp(appKey, 'runtime', 'supports_dynamic_routes', true)
    }

    const resolveAppSsoMode = (
      appKey?: string | null
    ): 'participate' | 'reauth' | 'isolated' => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return 'participate'
      const capabilities = resolveAppCapabilities(normalizedAppKey)
      const authConfig =
        capabilities && typeof capabilities.auth === 'object' && !Array.isArray(capabilities.auth)
          ? capabilities.auth
          : {}
      const mode = `${authConfig?.sso_mode || authConfig?.ssoMode || ''}`.trim()
      if (mode === 'participate' || mode === 'reauth' || mode === 'isolated') return mode
      return 'participate'
    }

    const resolveAppLoginUiMode = (
      appKey?: string | null
    ): 'auth_center_ui' | 'auth_center_custom' | 'local_ui' => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return 'auth_center_ui'
      const capabilities = resolveAppCapabilities(normalizedAppKey)
      const authConfig =
        capabilities && typeof capabilities.auth === 'object' && !Array.isArray(capabilities.auth)
          ? capabilities.auth
          : {}
      const mode = `${authConfig?.login_ui_mode || authConfig?.loginUiMode || ''}`.trim()
      if (mode === 'local_ui') return 'local_ui'
      if (mode === 'auth_center_custom') return 'auth_center_custom'
      return 'auth_center_ui'
    }

    const resolveAppLoginPageKey = (appKey?: string | null): string => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) return ''
      const capabilities = resolveAppCapabilities(normalizedAppKey)
      const authConfig =
        capabilities && typeof capabilities.auth === 'object' && !Array.isArray(capabilities.auth)
          ? capabilities.auth
          : {}
      return `${authConfig?.login_page_key || authConfig?.loginPageKey || ''}`.trim()
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
      const loginStrategy =
        `${authConfig?.login_strategy || authConfig?.loginStrategy || ''}`.trim()
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
      runtimeAppKey,
      managedAppKey,
      runtimeFrontendEntryURL,
      runtimeBackendEntryURL,
      runtimeHealthCheckURL,
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
      ensureRuntimeAppKey,
      switchApp,
      resolveAppAuthMode,
      resolveAppCapabilities,
      resolveAppMeta,
      supportsAppSwitchForApp,
      supportsDynamicRoutesForApp,
      resolveAppSsoMode,
      resolveAppLoginUiMode,
      resolveAppLoginPageKey,
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
        'appAuthModeMap',
        'appCapabilitiesMap',
        'appMetaMap'
      ]
    }
  }
)
