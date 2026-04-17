import { fetchGetCurrentApp } from '@/domains/governance/api'
import { useMenuStore } from '@/domains/navigation/menu'
import { useWorktabStore } from '@/domains/navigation/worktab'
import { registerAppContextRuntimeHandlers } from '@/domains/app-runtime/runtime/context-handlers'
import { refreshUserMenus } from '@/domains/navigation/runtime/navigation'
import { getNavigationRouter } from '@/domains/navigation/runtime/router-instance'
import { IframeRouteManager } from '@/domains/navigation/router-core/IframeRouteManager'
import { useAppContextStore } from '@/domains/app-runtime/context'
import { normalizeManagedAppKey } from '@/domains/app-runtime/managed-app-scope'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { findRegisteredRouteByPath } from '@/utils/router'

let runtimeAppLoading: Promise<string> | null = null

export interface SwitchAppPayload {
  appKey?: string | null
  name?: string | null
  frontendEntryUrl?: string | null
  backendEntryUrl?: string | null
  healthCheckUrl?: string | null
  authMode?: string | null
  capabilities?: Record<string, unknown> | null
  meta?: Record<string, unknown> | null
  defaultMenuSpaceKey?: string | null
}

function normalizeInternalPath(value?: string | null) {
  const raw = `${value || ''}`.trim()
  if (!raw || /^https?:\/\//i.test(raw)) {
    return raw
  }
  return raw.startsWith('/') ? raw : `/${raw}`
}

function resolveAppEntryTarget(targetApp: SwitchAppPayload) {
  const router = getNavigationRouter()
  const menuStore = useMenuStore()
  const menuSpaceStore = useMenuSpaceStore()
  const candidates = [
    menuStore.getHomePath(),
    menuSpaceStore.resolveSpaceLandingPath(),
    targetApp.frontendEntryUrl,
    '/'
  ]
    .map((item) => normalizeInternalPath(item))
    .filter(Boolean)

  for (const candidate of candidates) {
    if (/^https?:\/\//i.test(candidate)) {
      return {
        entryTarget: candidate,
        targetPath: candidate,
        menuSpaceKey: undefined
      }
    }

    const resolvedRoute = findRegisteredRouteByPath(router, candidate)
    if (resolvedRoute) {
      return {
        entryTarget: candidate,
        targetPath: candidate,
        menuSpaceKey: `${resolvedRoute.meta?.menuSpaceKey || ''}`.trim() || undefined
      }
    }
  }

  const fallbackPath = normalizeInternalPath(
    menuStore.getHomePath() || menuSpaceStore.resolveSpaceLandingPath() || '/'
  )
  return {
    entryTarget: fallbackPath,
    targetPath: fallbackPath,
    menuSpaceKey: undefined
  }
}

export async function ensureRuntimeAppKey(): Promise<string> {
  const appContextStore = useAppContextStore()
  const existingAppKey = normalizeManagedAppKey(
    appContextStore.effectiveManagedAppKey || appContextStore.runtimeAppKey
  )
  if (existingAppKey) {
    return existingAppKey
  }

  if (runtimeAppLoading) {
    return runtimeAppLoading
  }

  const pending = fetchGetCurrentApp(appContextStore.currentManagedAppKey || undefined)
    .then((response) => {
      const resolvedAppKey = normalizeManagedAppKey(response?.app?.appKey)
      if (!resolvedAppKey) {
        throw new Error('缺少运行时 app 上下文')
      }
      appContextStore.setRuntimeAppContext({
        appKey: resolvedAppKey,
        frontendEntryUrl: response?.app?.frontendEntryUrl || '',
        backendEntryUrl: response?.app?.backendEntryUrl || '',
        healthCheckUrl: response?.app?.healthCheckUrl || '',
        authMode: response?.app?.authMode || '',
        capabilities: response?.app?.capabilities,
        meta: response?.app?.meta || {}
      })
      return resolvedAppKey
    })
    .finally(() => {
      runtimeAppLoading = null
    })

  runtimeAppLoading = pending
  return pending
}

export async function switchApp(targetApp: SwitchAppPayload): Promise<void> {
  const router = getNavigationRouter()
  const appContextStore = useAppContextStore()
  const menuStore = useMenuStore()
  const menuSpaceStore = useMenuSpaceStore()
  const worktabStore = useWorktabStore()
  const nextAppKey = normalizeManagedAppKey(targetApp.appKey)
  if (!nextAppKey) {
    throw new Error('缺少目标应用 appKey')
  }

  appContextStore.setRuntimeAppContext({
    appKey: nextAppKey,
    frontendEntryUrl: targetApp.frontendEntryUrl || '',
    backendEntryUrl: targetApp.backendEntryUrl || '',
    healthCheckUrl: targetApp.healthCheckUrl || '',
    authMode: targetApp.authMode || '',
    capabilities: targetApp.capabilities || {},
    meta: targetApp.meta || {}
  })
  appContextStore.setManagedAppKey(nextAppKey)

  worktabStore.clearAll()
  IframeRouteManager.getInstance().clear()
  menuStore.removeAllDynamicRoutes()
  menuStore.setMenuList([])
  menuSpaceStore.clearActiveSpaceKey()

  await menuSpaceStore.refreshRuntimeConfig(true)
  const preferredSpaceKey = `${targetApp.defaultMenuSpaceKey || ''}`.trim()
  if (preferredSpaceKey) {
    menuSpaceStore.setActiveSpaceKey(preferredSpaceKey)
  }
  await menuSpaceStore.syncResolvedCurrentSpace(preferredSpaceKey)
  await refreshUserMenus()

  const { entryTarget, targetPath, menuSpaceKey } = resolveAppEntryTarget(targetApp)
  if (/^https?:\/\//i.test(entryTarget)) {
    window.location.assign(entryTarget)
    return
  }

  const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(targetPath, menuSpaceKey)
  if (nextTarget.mode === 'router') {
    await router.replace(nextTarget.target)
    return
  }

  window.location.assign(nextTarget.target)
}

registerAppContextRuntimeHandlers({
  ensureRuntimeAppKey,
  switchApp
})
