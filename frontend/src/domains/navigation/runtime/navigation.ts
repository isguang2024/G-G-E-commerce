import type { Router, RouteLocationNormalized } from 'vue-router'
import type { AppRouteRecord } from '@/types/router'
import { fetchGetRuntimeNavigation, fetchGetRuntimePublicPageList } from '@/domains/governance/api'
import { useAppContextStore } from '@/domains/app-runtime/context'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { useMenuStore } from '@/domains/navigation/menu'
import { useUserStore } from '@/domains/auth/store'
import { useWorktabStore } from '@/domains/navigation/worktab'
import { RouteRegistry } from '@/domains/navigation/router-core/RouteRegistry'
import { MenuProcessor } from '@/domains/navigation/router-core/MenuProcessor'
import { IframeRouteManager } from '@/domains/navigation/router-core/IframeRouteManager'
import { ManagedPageProcessor } from '@/domains/navigation/router-core/ManagedPageProcessor'
import {
  DEFAULT_MENU_SPACE_KEY,
  normalizeMenuSpaceKey
} from '@/domains/navigation/utils/menu-space'
import { registerNavigationRuntimeResetHandler } from '@/domains/navigation/runtime/reset-handlers'
import { hasRegisteredRoutePath } from '@/utils/router'
import { logger } from '@/utils/logger'

export class RuntimeManifestInvalidError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'RuntimeManifestInvalidError'
  }
}

let routeRegistry: RouteRegistry | null = null
let routeRegistrationMode: 'none' | 'public' | 'authenticated' = 'none'
const menuProcessor = new MenuProcessor()
const managedPageProcessor = new ManagedPageProcessor()

export function initNavigationRuntime(router: Router): void {
  routeRegistry = new RouteRegistry(router)
}

export function getRouteRegistrationMode(): 'none' | 'public' | 'authenticated' {
  return routeRegistrationMode
}

export function resetNavigationRuntime(): void {
  routeRegistry?.unregister()
  managedPageProcessor.invalidateCache()
  IframeRouteManager.getInstance().clear()

  const menuStore = useMenuStore()
  menuStore.removeAllDynamicRoutes()
  menuStore.setMenuList([])
  routeRegistrationMode = 'none'
}

registerNavigationRuntimeResetHandler(resetNavigationRuntime)

function inferRuntimeAppKeyByPath(path: string, runtimeAppKey?: string): string {
  const normalizedPath = `${path || ''}`.trim()
  if (normalizedPath.startsWith('/account/')) {
    return 'account-portal'
  }
  if (
    normalizedPath.startsWith('/dashboard/') ||
    normalizedPath.startsWith('/system/') ||
    normalizedPath.startsWith('/workspace/') ||
    normalizedPath.startsWith('/collaboration-workspace/')
  ) {
    return 'platform-admin'
  }
  if (normalizedPath.startsWith('/demo/')) {
    return 'demo-app'
  }
  return `${runtimeAppKey || ''}`.trim()
}

function collectRegisteredEntryPaths(list: AppRouteRecord[]): string[] {
  const visiblePaths: string[] = []
  const fallbackPaths: string[] = []
  const walk = (items: AppRouteRecord[]) => {
    ;(items || []).forEach((item) => {
      const path = `${item.path || ''}`.trim()
      if (path && !item.children?.length) {
        if (item.meta?.isHide) {
          fallbackPaths.push(path)
        } else {
          visiblePaths.push(path)
        }
      }
      if (item.children?.length) {
        walk(item.children)
      }
    })
  }
  walk(list)
  return visiblePaths.length ? [...visiblePaths, ...fallbackPaths] : fallbackPaths
}

function syncHomePathWithRegisteredRoutes(
  menuList: AppRouteRecord[],
  registeredRoutes: AppRouteRecord[]
): void {
  const menuStore = useMenuStore()
  const menuSpaceStore = useMenuSpaceStore()
  menuStore.setMenuList(menuList)

  const availablePaths = collectRegisteredEntryPaths(registeredRoutes)
  const preferredHomePath = menuSpaceStore.resolveSpaceLandingPath(availablePaths)
  if (preferredHomePath) {
    menuStore.setHomePath(preferredHomePath)
  }
}

async function buildRegisteredRoutesFromManifest(preferredSpaceKey = ''): Promise<{
  menuList: AppRouteRecord[]
  registeredRoutes: AppRouteRecord[]
  manifest: Api.SystemManage.RuntimeNavigationManifest
}> {
  const menuSpaceStore = useMenuSpaceStore()
  const appContextStore = useAppContextStore()
  const userStore = useUserStore()
  const requestedSpaceKey = normalizeMenuSpaceKey(
    preferredSpaceKey || menuSpaceStore.currentSpaceKey
  )
  let requestedAppKey = `${appContextStore.effectiveManagedAppKey || ''}`.trim()
  if (!requestedAppKey) {
    try {
      await menuSpaceStore.refreshRuntimeConfig(false)
      requestedAppKey = `${appContextStore.effectiveManagedAppKey || ''}`.trim()
    } catch (error) {
      logger.warn('navigation.runtime.prewarm_app_context_failed', { err: error })
    }
  }
  const fallbackSpaceKey = normalizeMenuSpaceKey(menuSpaceStore.defaultSpaceKey)
  const candidateSpaceKeys = Array.from(
    new Set(
      [requestedSpaceKey, fallbackSpaceKey, '']
        .map((item) => normalizeMenuSpaceKey(item))
        .filter((item, index, source) => index === source.indexOf(item))
    )
  )
  let manifest: Api.SystemManage.RuntimeNavigationManifest | null = null
  let menuTree: any[] = []
  let managedPages: any[] = []
  let usedSpaceKey = requestedSpaceKey

  for (const candidate of candidateSpaceKeys) {
    const current = await fetchGetRuntimeNavigation(
      candidate || undefined,
      requestedAppKey || undefined
    )
    const currentMenuTree = Array.isArray(current?.menuTree) ? current.menuTree : []
    const currentManagedPages = Array.isArray(current?.managedPages) ? current.managedPages : []
    if (current && (currentMenuTree.length > 0 || currentManagedPages.length > 0)) {
      manifest = current
      menuTree = currentMenuTree
      managedPages = currentManagedPages
      usedSpaceKey = candidate
      break
    }
  }

  if (!manifest || (menuTree.length === 0 && managedPages.length === 0)) {
    throw new RuntimeManifestInvalidError(
      `[NavigationRuntime] runtime navigation manifest 缺失或无可注册路由 (space=${requestedSpaceKey || 'default'})`
    )
  }
  const resolvedAppKey =
    `${manifest.currentApp?.app?.appKey || manifest.context?.app_key || ''}`.trim()
  if (resolvedAppKey) {
    appContextStore.setRuntimeAppContext({
      appKey: resolvedAppKey,
      frontendEntryUrl: manifest.currentApp?.app?.frontendEntryUrl || '',
      backendEntryUrl: manifest.currentApp?.app?.backendEntryUrl || '',
      healthCheckUrl: manifest.currentApp?.app?.healthCheckUrl || '',
      authMode: manifest.currentApp?.app?.authMode || '',
      capabilities: manifest.currentApp?.app?.capabilities,
      meta: manifest.currentApp?.app?.meta || {}
    })
  }
  const resolvedSpaceKey = normalizeMenuSpaceKey(
    manifest.currentSpace?.space?.spaceKey || manifest.context?.space_key || usedSpaceKey
  )

  if (
    resolvedSpaceKey &&
    resolvedSpaceKey !== normalizeMenuSpaceKey(menuSpaceStore.currentSpaceKey)
  ) {
    menuSpaceStore.setActiveSpaceKey(resolvedSpaceKey)
  }

  const menuList = menuProcessor.normalizeMenuList(menuTree)
  const currentUser = userStore.getUserInfo as Api.Auth.UserInfo
  const manifestVersion = `${
    (manifest as { version?: string | number }).version ||
    (manifest.context as { version?: string | number } | undefined)?.version ||
    'v0'
  }`
  const cacheKey = `${resolvedSpaceKey || usedSpaceKey || DEFAULT_MENU_SPACE_KEY}:${currentUser?.id || currentUser?.userId || 'anon'}:${manifestVersion}`
  const runtimeRoutes = managedPageProcessor.buildRoutes(menuList, managedPages, currentUser, {
    trustBackend: true,
    cacheKey
  })

  return {
    menuList,
    registeredRoutes: [...menuList, ...runtimeRoutes],
    manifest
  }
}

export async function refreshUserMenus(preferredSpaceKey = ''): Promise<void> {
  if (!routeRegistry) return
  const menuStore = useMenuStore()
  managedPageProcessor.invalidateCache()
  const { menuList, registeredRoutes } = await buildRegisteredRoutesFromManifest(preferredSpaceKey)
  routeRegistry.unregister()
  routeRegistry.register(registeredRoutes)
  routeRegistrationMode = 'authenticated'
  syncHomePathWithRegisteredRoutes(menuList, registeredRoutes)
  menuStore.clearRemoveRouteFns()
  menuStore.addRemoveRouteFns(routeRegistry.getRemoveRouteFns())
  IframeRouteManager.getInstance().save()
}

export async function ensureAuthenticatedRoutes(
  to: RouteLocationNormalized,
  preferredSpaceKey: string,
  ensureSession: () => Promise<void>,
  router: Router
): Promise<string> {
  if (!routeRegistry) {
    throw new Error('[NavigationRuntime] routeRegistry 未初始化')
  }

  const appContextStore = useAppContextStore()
  const inferredAppKey = inferRuntimeAppKeyByPath(to.path, appContextStore.currentRuntimeAppKey)
  if (inferredAppKey && inferredAppKey !== appContextStore.effectiveManagedAppKey) {
    appContextStore.setManagedAppKey(inferredAppKey)
  }

  await ensureSession()
  const { menuList, registeredRoutes } = await buildRegisteredRoutesFromManifest(preferredSpaceKey)
  routeRegistry.register(registeredRoutes)
  routeRegistrationMode = 'authenticated'

  const menuStore = useMenuStore()
  syncHomePathWithRegisteredRoutes(menuList, registeredRoutes)
  menuStore.addRemoveRouteFns(routeRegistry.getRemoveRouteFns())

  IframeRouteManager.getInstance().save()
  useWorktabStore().validateWorktabs(router)
  return menuStore.getHomePath() || '/'
}

export function isPublicRuntimeRoute(to: RouteLocationNormalized): boolean {
  return to.matched.some(
    (record) =>
      Boolean(record.meta?.isInnerPage) && `${record.meta?.accessMode || ''}`.trim() === 'public'
  )
}

export async function ensurePublicRuntimeRoutes(
  to: RouteLocationNormalized,
  router: Router
): Promise<boolean> {
  if (!routeRegistry) {
    return false
  }
  if (routeRegistrationMode === 'authenticated') {
    return isPublicRuntimeRoute(to)
  }
  if (routeRegistrationMode === 'public' && isPublicRuntimeRoute(to)) {
    return true
  }
  if (routeRegistrationMode === 'public') {
    routeRegistry.unregister()
    routeRegistrationMode = 'none'
  }
  try {
    const menuSpaceStore = useMenuSpaceStore()
    const appContextStore = useAppContextStore()
    const inferredPublicAppKey = inferRuntimeAppKeyByPath(
      to.path,
      appContextStore.currentRuntimeAppKey
    )
    const runtimeRes = await fetchGetRuntimePublicPageList(
      menuSpaceStore.currentSpaceKey,
      `${appContextStore.effectiveManagedAppKey || inferredPublicAppKey || ''}`.trim() || undefined
    )
    const publicRoutes = managedPageProcessor.buildRoutes([], runtimeRes.records || [], null, {
      trustBackend: true
    })
    routeRegistry.register(publicRoutes)
    routeRegistrationMode = 'public'
    const resolved = router.resolve({
      path: to.path,
      query: to.query,
      hash: to.hash
    })
    return resolved.matched.some(
      (record) =>
        Boolean(record.meta?.isInnerPage) && `${record.meta?.accessMode || ''}`.trim() === 'public'
    )
  } catch (error) {
    logger.error('navigation.runtime.register_public_pages_failed', { err: error })
    routeRegistry?.unregister()
    routeRegistrationMode = 'none'
    return false
  }
}

export function hasDynamicRoute(router: Router, path: string): boolean {
  return hasRegisteredRoutePath(router, path)
}
