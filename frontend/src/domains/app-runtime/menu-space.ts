import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { normalizeManagedAppKey } from '@/domains/app-runtime/managed-app-scope'
import { useAppContextStore } from '@/domains/app-runtime/context'
import {
  fetchGetCurrentMenuSpace,
  fetchGetMenuSpaces,
  fetchGetMenuSpaceHostBindings
} from '@/domains/governance/api'
import type { MenuSpaceConfig } from '@/types/config'
import {
  buildMenuSpaceTargetUrl,
  DEFAULT_MENU_SPACE_KEY,
  createFallbackMenuSpaceConfig,
  normalizeMenuHost,
  normalizeMenuSpaceKey,
  SHARED_MENU_SPACE_KEY,
  resolveMenuSpaceDefaultKey,
  resolveMenuSpaceHostBinding,
  resolveMenuSpaceDefinition,
  resolveMenuSpaceKeyByHost,
  shouldUseFullMenuSpaceNavigation
} from '@/domains/navigation/utils/menu-space'
import { logger } from '@/utils/logger'

const warnDev = (event: string, context?: Record<string, unknown>) => {
  // DEV-only 提示：logger.debug 生产端不上报。
  if (import.meta.env.DEV) {
    logger.debug(`menu-space.${event}`, context)
  }
}

function buildRuntimeMenuSpaceConfig(
  spaces: Api.SystemManage.MenuSpaceItem[] = [],
  hostBindings: Api.SystemManage.MenuSpaceHostBindingItem[] = [],
  fallbackConfig: MenuSpaceConfig = createFallbackMenuSpaceConfig()
): MenuSpaceConfig {
  const normalizedSpaces = (spaces || [])
    .filter((item) => `${item.menuSpaceKey || ''}`.trim())
    .map((item) => ({
      menuSpaceKey: item.menuSpaceKey,
      spaceName: item.name,
      spaceType: `${item.meta?.spaceType || item.meta?.space_type || 'default'}`,
      description: item.description || '',
      enabled: item.status !== 'disabled',
      isDefault: Boolean(item.isDefault),
      defaultLandingRoute: item.defaultHomePath || '/'
    }))

  const normalizedHostBindings = (hostBindings || [])
    .filter((item) => `${item.host || ''}`.trim() && `${item.menuSpaceKey || ''}`.trim())
    .map((item) => ({
      host: item.host,
      menuSpaceKey: item.menuSpaceKey,
      enabled: item.status !== 'disabled',
      isPrimary: Boolean(item.meta?.isPrimary || item.meta?.is_primary),
      scheme: `${item.scheme || item.meta?.scheme || ''}`,
      routePrefix: `${item.routePrefix || item.meta?.routePrefix || item.meta?.route_prefix || ''}`,
      authMode: `${item.authMode || item.meta?.authMode || item.meta?.auth_mode || 'inherit_host'}`,
      loginHost: `${item.loginHost || item.meta?.loginHost || item.meta?.login_host || ''}`,
      callbackHost: `${item.callbackHost || item.meta?.callbackHost || item.meta?.callback_host || ''}`,
      cookieScopeMode: `${item.cookieScopeMode || item.meta?.cookieScopeMode || item.meta?.cookie_scope_mode || 'inherit'}`,
      cookieDomain: `${item.cookieDomain || item.meta?.cookieDomain || item.meta?.cookie_domain || ''}`
    }))

  if (!normalizedSpaces.length) {
    return fallbackConfig || createFallbackMenuSpaceConfig()
  }

    const defaultSpace =
      normalizedSpaces.find((item) => item.isDefault) ||
      normalizedSpaces.find(
        (item) => normalizeMenuSpaceKey(item.menuSpaceKey) === DEFAULT_MENU_SPACE_KEY
      ) ||
      normalizedSpaces[0]

    return {
      defaultMenuSpaceKey:
        normalizeMenuSpaceKey(defaultSpace?.menuSpaceKey) || DEFAULT_MENU_SPACE_KEY,
      spaces: normalizedSpaces,
      hostBindings: normalizedHostBindings
    }
  }

function appendSpaceKeyToTargetPath(targetPath: string, spaceKey: string): string {
  const normalizedPath = `${targetPath || ''}`.trim() || '/'
  const normalizedSpaceKey = normalizeMenuSpaceKey(spaceKey)
  if (!normalizedSpaceKey || normalizedSpaceKey === SHARED_MENU_SPACE_KEY) {
    return normalizedPath
  }

  try {
    const url = new URL(normalizedPath, 'http://gge.local')
    url.searchParams.set('menu_space_key', normalizedSpaceKey)
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return normalizedPath
  }
}

export const useMenuSpaceStore = defineStore(
  'menuSpaceStore',
  () => {
    const appContextStore = useAppContextStore()
    const menuSpaceConfigMap = ref<Record<string, MenuSpaceConfig>>({})
    const overrideSpaceKeyMap = ref<Record<string, string>>({})
    const runtimeHost = ref('')
    const loadingAppKeys = ref<Record<string, boolean>>({})
    const loadedAppKeys = ref<Record<string, boolean>>({})

    const currentHost = computed(() => {
      const host = runtimeHost.value || (typeof window !== 'undefined' ? window.location.host : '')
      return normalizeMenuHost(host)
    })

    const currentAppKey = computed(() =>
      normalizeManagedAppKey(
        appContextStore.effectiveManagedAppKey || appContextStore.runtimeAppKey
      )
    )
    const currentMenuSpaceConfig = computed(() => {
      const appKey = currentAppKey.value
      if (!appKey) {
        return createFallbackMenuSpaceConfig() || createFallbackMenuSpaceConfig()
      }
      return menuSpaceConfigMap.value[appKey] || createFallbackMenuSpaceConfig()
    })
    const resolvedDefaultMenuSpaceKey = computed(() =>
      resolveMenuSpaceDefaultKey(currentMenuSpaceConfig.value, DEFAULT_MENU_SPACE_KEY)
    )
    const currentOverrideSpaceKey = computed(() => {
      const appKey = currentAppKey.value
      if (!appKey) {
        return ''
      }
      return normalizeMenuSpaceKey(overrideSpaceKeyMap.value[appKey])
    })
    const loading = computed(() => Boolean(loadingAppKeys.value[currentAppKey.value]))
    const loaded = computed(() => Boolean(loadedAppKeys.value[currentAppKey.value]))

    const currentSpaceKey = computed(() => {
      const forcedKey = currentOverrideSpaceKey.value
      if (forcedKey) {
        return forcedKey
      }
      return resolveMenuSpaceKeyByHost(
        currentHost.value,
        currentMenuSpaceConfig.value,
        resolvedDefaultMenuSpaceKey.value
      )
    })

    const currentSpace = computed(
      () =>
        resolveMenuSpaceDefinition(currentSpaceKey.value, currentMenuSpaceConfig.value) ||
        resolveMenuSpaceDefinition(
          resolvedDefaultMenuSpaceKey.value,
          currentMenuSpaceConfig.value
        ) ||
        currentMenuSpaceConfig.value.spaces[0] ||
        null
    )

    const hasMultiSpace = computed(() => (currentMenuSpaceConfig.value.spaces || []).length > 1)
    const hasHostBinding = computed(() =>
      (currentMenuSpaceConfig.value.hostBindings || []).some((item) =>
        Boolean(item?.enabled ?? true)
      )
    )

    const isDefaultSpace = computed(
      () => currentSpaceKey.value === resolvedDefaultMenuSpaceKey.value
    )

    const setMenuSpaceConfig = (config: MenuSpaceConfig, appKey = currentAppKey.value) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) {
        return
      }
      menuSpaceConfigMap.value = {
        ...menuSpaceConfigMap.value,
        [normalizedAppKey]: config || createFallbackMenuSpaceConfig()
      }
      loadedAppKeys.value = {
        ...loadedAppKeys.value,
        [normalizedAppKey]: true
      }
    }

    const setOverrideSpaceKey = (spaceKey: string, appKey = currentAppKey.value) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) {
        return
      }
      overrideSpaceKeyMap.value = {
        ...overrideSpaceKeyMap.value,
        [normalizedAppKey]: normalizeMenuSpaceKey(spaceKey)
      }
    }

    const clearOverrideSpaceKey = (appKey = currentAppKey.value) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) {
        return
      }
      const nextMap = { ...overrideSpaceKeyMap.value }
      delete nextMap[normalizedAppKey]
      overrideSpaceKeyMap.value = nextMap
    }

    const setLoadingState = (appKey: string, value: boolean) => {
      const normalizedAppKey = normalizeManagedAppKey(appKey)
      if (!normalizedAppKey) {
        return
      }
      loadingAppKeys.value = {
        ...loadingAppKeys.value,
        [normalizedAppKey]: value
      }
    }

    const ensureRuntimeAppKey = async () => {
      return appContextStore.ensureRuntimeAppKey()
    }

    const refreshRuntimeConfig = async (force = false) => {
      const appKey = await ensureRuntimeAppKey()
      if (loadingAppKeys.value[appKey]) {
        return (
          menuSpaceConfigMap.value[appKey] ||
          createFallbackMenuSpaceConfig() ||
          createFallbackMenuSpaceConfig()
        )
      }
      if (loadedAppKeys.value[appKey] && !force) {
        return (
          menuSpaceConfigMap.value[appKey] ||
          createFallbackMenuSpaceConfig() ||
          createFallbackMenuSpaceConfig()
        )
      }
      setLoadingState(appKey, true)
      try {
        const [spacesRes, hostBindingsRes] = await Promise.all([
          fetchGetMenuSpaces(appKey),
          fetchGetMenuSpaceHostBindings(appKey)
        ])
        setMenuSpaceConfig(
          buildRuntimeMenuSpaceConfig(
            spacesRes.records || [],
            hostBindingsRes.records || [],
            createFallbackMenuSpaceConfig()
          ),
          appKey
        )
      } catch (error) {
        warnDev('sync_failed_fallback_static', { err: error })
        setMenuSpaceConfig(
          createFallbackMenuSpaceConfig() || createFallbackMenuSpaceConfig(),
          appKey
        )
      } finally {
        setLoadingState(appKey, false)
      }
      return menuSpaceConfigMap.value[appKey] || createFallbackMenuSpaceConfig()
    }

    const syncResolvedCurrentSpace = async (preferredSpaceKey = '') => {
      const appKey = await ensureRuntimeAppKey()
      const requestedSpaceKey = normalizeMenuSpaceKey(
        preferredSpaceKey || currentOverrideSpaceKey.value || currentSpaceKey.value
      )
      const hostResolvedSpaceKey = resolveMenuSpaceKeyByHost(
        currentHost.value,
        currentMenuSpaceConfig.value,
        resolvedDefaultMenuSpaceKey.value
      )
      try {
        const response = await fetchGetCurrentMenuSpace(requestedSpaceKey || undefined, appKey)
        const resolvedSpaceKey = normalizeMenuSpaceKey(response?.space?.menuSpaceKey)
        if (!resolvedSpaceKey) {
          clearOverrideSpaceKey(appKey)
          return null
        }
        if (resolvedSpaceKey === hostResolvedSpaceKey) {
          clearOverrideSpaceKey(appKey)
        } else {
          setOverrideSpaceKey(resolvedSpaceKey, appKey)
        }
        return response
      } catch (error) {
        warnDev('sync_current_failed_local_kept', { err: error })
        return null
      }
    }

    const setActiveSpaceKey = (spaceKey: string) => {
      setOverrideSpaceKey(spaceKey)
    }

    const clearActiveSpaceKey = () => {
      clearOverrideSpaceKey()
    }

    const syncRuntimeHost = () => {
      runtimeHost.value = typeof window !== 'undefined' ? window.location.host : ''
    }

    const shouldShowSpaceBadge = computed(
      () => hasMultiSpace.value || hasHostBinding.value || !isDefaultSpace.value
    )

    const resolveSpaceLandingPath = (availablePaths: string[] = []) => {
      const normalizedAvailablePaths = (availablePaths || [])
        .map((item) => `${item || ''}`.trim())
        .filter(Boolean)
      const hasWorkspaceInbox = normalizedAvailablePaths.includes('/workspace/inbox')
      const nonConsolePaths = normalizedAvailablePaths.filter(
        (item) => item !== '/dashboard/console' && !item.startsWith('/dashboard/console/')
      )
      const onlyWorkspaceFallbackPaths =
        hasWorkspaceInbox &&
        nonConsolePaths.length > 0 &&
        nonConsolePaths.every((item) =>
          ['/workspace/inbox', '/dashboard/console/user-center'].includes(item)
        )
      if (onlyWorkspaceFallbackPaths) {
        return '/workspace/inbox'
      }
      const configuredLandingPath = `${currentSpace.value?.defaultLandingRoute || ''}`.trim()
      if (configuredLandingPath) {
        if (
          !normalizedAvailablePaths.length ||
          normalizedAvailablePaths.includes(configuredLandingPath)
        ) {
          return configuredLandingPath
        }
      }
      const defaultSpaceDefinition = resolveMenuSpaceDefinition(
        resolvedDefaultMenuSpaceKey.value,
        currentMenuSpaceConfig.value
      )
      const defaultSpaceLandingPath = `${defaultSpaceDefinition?.defaultLandingRoute || ''}`.trim()
      if (
        defaultSpaceLandingPath &&
        (!normalizedAvailablePaths.length ||
          normalizedAvailablePaths.includes(defaultSpaceLandingPath))
      ) {
        return defaultSpaceLandingPath
      }
      if (
        !normalizedAvailablePaths.length ||
        normalizedAvailablePaths.includes('/workspace/inbox')
      ) {
        return '/workspace/inbox'
      }
      if (
        !normalizedAvailablePaths.length ||
        normalizedAvailablePaths.includes('/dashboard/console')
      ) {
        return '/dashboard/console'
      }
      return normalizedAvailablePaths[0] || '/'
    }

    const matchesSpace = (spaceKey?: unknown) => {
      const targetSpaceKey = normalizeMenuSpaceKey(spaceKey)
      if (!targetSpaceKey) {
        return isDefaultSpace.value
      }
      return (
        normalizeMenuSpaceKey(targetSpaceKey) === currentSpaceKey.value ||
        normalizeMenuSpaceKey(targetSpaceKey) === 'shared'
      )
    }

    const resolveHostBinding = (spaceKey?: string) => {
      return resolveMenuSpaceHostBinding(
        normalizeMenuSpaceKey(spaceKey) || currentSpaceKey.value,
        currentMenuSpaceConfig.value,
        typeof window !== 'undefined' ? window.location.host : runtimeHost.value
      )
    }

    const buildSpaceTargetUrl = (targetPath: string, spaceKey?: string) => {
      return buildMenuSpaceTargetUrl(
        resolveHostBinding(spaceKey),
        targetPath,
        typeof window !== 'undefined' ? window.location.host : runtimeHost.value
      )
    }

    const resolveSpaceNavigationTarget = (targetPath: string, spaceKey?: string) => {
      const normalizedPath = `${targetPath || ''}`.trim() || '/'
      const normalizedTargetSpaceKey = normalizeMenuSpaceKey(spaceKey) || currentSpaceKey.value
      const hostResolvedSpaceKey = resolveMenuSpaceKeyByHost(
        typeof window !== 'undefined' ? window.location.host : runtimeHost.value,
        currentMenuSpaceConfig.value,
        resolvedDefaultMenuSpaceKey.value
      )
      const routerTarget =
        normalizedTargetSpaceKey && normalizedTargetSpaceKey !== hostResolvedSpaceKey
          ? appendSpaceKeyToTargetPath(normalizedPath, normalizedTargetSpaceKey)
          : normalizedPath
      const binding = resolveHostBinding(spaceKey)
      if (!binding?.host) {
        return {
          mode: 'router' as const,
          target: routerTarget
        }
      }
      const currentHost = window.location.host
      const targetUrl = buildMenuSpaceTargetUrl(binding, normalizedPath, currentHost)
      if (typeof window === 'undefined') {
        return {
          mode: 'location' as const,
          target: targetUrl
        }
      }
      const shouldUseLocationNavigation = shouldUseFullMenuSpaceNavigation(
        binding,
        currentHost,
        window.location.protocol,
        window.location.pathname
      )
      if (!shouldUseLocationNavigation) {
        return {
          mode: 'router' as const,
          target: routerTarget
        }
      }
      warnDev('full_page_nav_switch', {
        targetPath: normalizedPath,
        targetUrl,
        host: binding.host,
        routePrefix: `${binding.routePrefix || ''}`.trim()
      })
      return {
        mode: 'location' as const,
        target: targetUrl
      }
    }

    syncRuntimeHost()

    return {
      menuSpaceConfig: currentMenuSpaceConfig,
      runtimeHost,
      loading,
      loaded,
      currentHost,
      currentSpaceKey,
      defaultMenuSpaceKey: resolvedDefaultMenuSpaceKey,
      currentSpace,
      hasMultiSpace,
      hasHostBinding,
      isDefaultSpace,
      shouldShowSpaceBadge,
      resolveSpaceLandingPath,
      setMenuSpaceConfig,
      refreshRuntimeConfig,
      syncResolvedCurrentSpace,
      setActiveSpaceKey,
      clearActiveSpaceKey,
      syncRuntimeHost,
      matchesSpace,
      resolveHostBinding,
      buildSpaceTargetUrl,
      resolveSpaceNavigationTarget
    }
  },
  {
    persist: {
      storage: localStorage
    }
  }
)
