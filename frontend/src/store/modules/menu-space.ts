import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import AppConfig from '@/config'
import { useUserStore } from '@/store/modules/user'
import { hasPlatformAccessByUserInfo } from '@/store/modules/tenant'
import {
  fetchGetCurrentMenuSpace,
  fetchGetMenuSpaces,
  fetchGetMenuSpaceHostBindings
} from '@/api/system-manage'
import type { MenuSpaceConfig } from '@/types/config'
import {
  buildMenuSpaceTargetUrl,
  DEFAULT_MENU_SPACE_KEY,
  createFallbackMenuSpaceConfig,
  normalizeMenuHost,
  normalizeMenuSpaceKey,
  resolveMenuSpaceHostBinding,
  resolveMenuSpaceDefinition,
  resolveMenuSpaceKeyByHost
} from '@/utils/navigation/menu-space'

const runtimeMenuSpaceConfig = AppConfig.menuSpace || createFallbackMenuSpaceConfig()

const warnDev = (...args: any[]) => {
  if (import.meta.env.DEV) {
    console.warn(...args)
  }
}

function buildRuntimeMenuSpaceConfig(
  spaces: Api.SystemManage.MenuSpaceItem[] = [],
  hostBindings: Api.SystemManage.MenuSpaceHostBindingItem[] = [],
  fallbackConfig: MenuSpaceConfig = runtimeMenuSpaceConfig
): MenuSpaceConfig {
  const normalizedSpaces = (spaces || [])
    .filter((item) => `${item.spaceKey || ''}`.trim())
    .map((item) => ({
      spaceKey: item.spaceKey,
      spaceName: item.name,
      spaceType: `${item.meta?.spaceType || item.meta?.space_type || 'default'}`,
      description: item.description || '',
      enabled: item.status !== 'disabled',
      isDefault: Boolean(item.isDefault),
      defaultLandingRoute: item.defaultHomePath || '/'
    }))

  const normalizedHostBindings = (hostBindings || [])
    .filter((item) => `${item.host || ''}`.trim() && `${item.spaceKey || ''}`.trim())
    .map((item) => ({
      host: item.host,
      spaceKey: item.spaceKey,
      enabled: item.status !== 'disabled',
      isPrimary: Boolean(item.meta?.isPrimary || item.meta?.is_primary),
      scheme: `${item.scheme || item.meta?.scheme || 'https'}`,
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
    normalizedSpaces.find((item) => normalizeMenuSpaceKey(item.spaceKey) === DEFAULT_MENU_SPACE_KEY) ||
    normalizedSpaces[0]

  return {
    defaultSpaceKey: normalizeMenuSpaceKey(defaultSpace?.spaceKey) || DEFAULT_MENU_SPACE_KEY,
    spaces: normalizedSpaces,
    hostBindings: normalizedHostBindings
  }
}

export const useMenuSpaceStore = defineStore(
  'menuSpaceStore',
  () => {
    const menuSpaceConfig = ref(runtimeMenuSpaceConfig)
    const overrideSpaceKey = ref('')
    const runtimeHost = ref('')
    const loading = ref(false)
    const loaded = ref(false)

    const currentHost = computed(() => {
      const host = runtimeHost.value || (typeof window !== 'undefined' ? window.location.hostname : '')
      return normalizeMenuHost(host)
    })

    const defaultSpaceKey = computed(() => {
      const key = normalizeMenuSpaceKey(menuSpaceConfig.value.defaultSpaceKey)
      return key || DEFAULT_MENU_SPACE_KEY
    })

    const currentSpaceKey = computed(() => {
      const forcedKey = normalizeMenuSpaceKey(overrideSpaceKey.value)
      if (forcedKey) {
        return forcedKey
      }
      return resolveMenuSpaceKeyByHost(currentHost.value, menuSpaceConfig.value, defaultSpaceKey.value)
    })

    const currentSpace = computed(() =>
      resolveMenuSpaceDefinition(currentSpaceKey.value, menuSpaceConfig.value) ||
      resolveMenuSpaceDefinition(defaultSpaceKey.value, menuSpaceConfig.value) ||
      menuSpaceConfig.value.spaces[0] ||
      null
    )

    const hasMultiSpace = computed(() => (menuSpaceConfig.value.spaces || []).length > 1)
    const hasHostBinding = computed(() => (menuSpaceConfig.value.hostBindings || []).some((item) => Boolean(item?.enabled ?? true)))

    const isDefaultSpace = computed(() => currentSpaceKey.value === defaultSpaceKey.value)

    const setMenuSpaceConfig = (config: typeof menuSpaceConfig.value) => {
      menuSpaceConfig.value = config || createFallbackMenuSpaceConfig()
    }

    const refreshRuntimeConfig = async (force = false) => {
      if (loading.value) {
        return menuSpaceConfig.value
      }
      if (loaded.value && !force) {
        return menuSpaceConfig.value
      }
      const userStore = useUserStore()
      const currentUserInfo = userStore.getUserInfo as Api.Auth.UserInfo
      if (!hasPlatformAccessByUserInfo(currentUserInfo)) {
        menuSpaceConfig.value = runtimeMenuSpaceConfig || createFallbackMenuSpaceConfig()
        loaded.value = true
        return menuSpaceConfig.value
      }
      loading.value = true
      try {
        const [spacesRes, hostBindingsRes] = await Promise.all([
          fetchGetMenuSpaces(),
          fetchGetMenuSpaceHostBindings()
        ])
        menuSpaceConfig.value = buildRuntimeMenuSpaceConfig(
          spacesRes.records || [],
          hostBindingsRes.records || [],
          runtimeMenuSpaceConfig
        )
        loaded.value = true
      } catch (error) {
        warnDev('[menu-space] 同步后端菜单空间配置失败，已回退静态配置', error)
        menuSpaceConfig.value = runtimeMenuSpaceConfig || createFallbackMenuSpaceConfig()
      } finally {
        loading.value = false
      }
      return menuSpaceConfig.value
    }

    const syncResolvedCurrentSpace = async (preferredSpaceKey = '') => {
      const requestedSpaceKey = normalizeMenuSpaceKey(preferredSpaceKey || overrideSpaceKey.value || currentSpaceKey.value)
      const hostResolvedSpaceKey = resolveMenuSpaceKeyByHost(currentHost.value, menuSpaceConfig.value, defaultSpaceKey.value)
      try {
        const response = await fetchGetCurrentMenuSpace(requestedSpaceKey || undefined)
        const resolvedSpaceKey = normalizeMenuSpaceKey(response?.space?.spaceKey)
        if (!resolvedSpaceKey) {
          overrideSpaceKey.value = ''
          return null
        }
        overrideSpaceKey.value = resolvedSpaceKey === hostResolvedSpaceKey ? '' : resolvedSpaceKey
        return response
      } catch (error) {
        warnDev('[menu-space] 同步当前空间解析失败，已保留本地结果', error)
        return null
      }
    }

    const setActiveSpaceKey = (spaceKey: string) => {
      overrideSpaceKey.value = normalizeMenuSpaceKey(spaceKey)
    }

    const clearActiveSpaceKey = () => {
      overrideSpaceKey.value = ''
    }

    const syncRuntimeHost = () => {
      runtimeHost.value = typeof window !== 'undefined' ? window.location.hostname : ''
    }

    const shouldShowSpaceBadge = computed(() => hasMultiSpace.value || hasHostBinding.value || !isDefaultSpace.value)

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
      const defaultSpaceDefinition = resolveMenuSpaceDefinition(defaultSpaceKey.value, menuSpaceConfig.value)
      const defaultSpaceLandingPath = `${defaultSpaceDefinition?.defaultLandingRoute || ''}`.trim()
      if (
        defaultSpaceLandingPath &&
        (!normalizedAvailablePaths.length || normalizedAvailablePaths.includes(defaultSpaceLandingPath))
      ) {
        return defaultSpaceLandingPath
      }
      if (
        !normalizedAvailablePaths.length ||
        normalizedAvailablePaths.includes('/workspace/inbox')
      ) {
        return '/workspace/inbox'
      }
      if (!normalizedAvailablePaths.length || normalizedAvailablePaths.includes('/dashboard/console')) {
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
        menuSpaceConfig.value
      )
    }

    const buildSpaceTargetUrl = (targetPath: string, spaceKey?: string) => {
      return buildMenuSpaceTargetUrl(resolveHostBinding(spaceKey), targetPath)
    }

    const resolveSpaceNavigationTarget = (targetPath: string, spaceKey?: string) => {
      const normalizedPath = `${targetPath || ''}`.trim() || '/'
      const binding = resolveHostBinding(spaceKey)
      if (!binding?.host) {
        return {
          mode: 'router' as const,
          target: normalizedPath
        }
      }
      const targetUrl = buildMenuSpaceTargetUrl(binding, normalizedPath)
      if (typeof window === 'undefined') {
        return {
          mode: 'location' as const,
          target: targetUrl
        }
      }
      const currentProtocol = window.location.protocol.replace(':', '')
      const currentOriginHost = normalizeMenuHost(window.location.hostname)
      const targetHost = normalizeMenuHost(binding.host)
      const routePrefix = `${binding.routePrefix || ''}`.trim()
      if (
        targetHost === currentOriginHost &&
        (!routePrefix || routePrefix === '/') &&
        `${binding.scheme || 'https'}`.trim().toLowerCase() === currentProtocol
      ) {
        return {
          mode: 'router' as const,
          target: normalizedPath
        }
      }
      return {
        mode: 'location' as const,
        target: targetUrl
      }
    }

    syncRuntimeHost()

    return {
      menuSpaceConfig,
      runtimeHost,
      loading,
      loaded,
      currentHost,
      currentSpaceKey,
      currentSpace,
      defaultSpaceKey,
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
      key: 'menu-space',
      storage: localStorage
    }
  }
)
