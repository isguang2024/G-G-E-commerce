import type { AppRouteRecord, RouteMeta } from '@/types/router'
import { useMenuSpaceStore } from '@/store/modules/menu-space'
import { hasScopedActionPermission } from '@/utils/permission/action'
import { resolveManagedPageRoutePath } from '@/utils/navigation/managed-page'
import {
  getMenuActionRequirement,
  hasMenuActionAccess,
  shouldHideMenuWhenActionDenied
} from '@/utils/permission/menu'
import { isMenuSpaceVisible, normalizeMenuSpaceKey } from '@/utils/navigation/menu-space'

type RuntimePageItem = Api.SystemManage.PageItem
type UserInfo = Partial<Api.Auth.UserInfo> | null | undefined
type ActionMatchMode = 'any' | 'all'
type ActionVisibilityMode = 'hide' | 'show'

interface IndexedMenus {
  byId: Map<string, AppRouteRecord>
  byName: Map<string, AppRouteRecord>
  byPath: Map<string, AppRouteRecord>
  names: Set<string>
  paths: Set<string>
}

interface ResolvedPageConfig {
  allowed: boolean
  activePath: string
  customParent: string
  breadcrumbChain: Array<{ title: string; path?: string }>
  effectiveAccessMode: 'public' | 'jwt' | 'permission'
  spaceKey?: string
  spaceType?: string
  hostKey?: string
  requiredAction?: string
  requiredActions?: string[]
  actionMatchMode?: ActionMatchMode
  actionVisibilityMode?: ActionVisibilityMode
}

export class ManagedPageProcessor {
  private buildCache = new Map<string, AppRouteRecord[]>()

  /**
   * 清空 buildRoutes 缓存。在菜单/用户/manifest 刷新时调用。
   */
  invalidateCache(): void {
    this.buildCache.clear()
  }

  buildRoutes(
    menuList: AppRouteRecord[],
    pages: RuntimePageItem[],
    userInfo: UserInfo,
    options: { trustBackend?: boolean; cacheKey?: string } = {}
  ): AppRouteRecord[] {
    if (!Array.isArray(pages) || pages.length === 0) {
      return []
    }
    const cacheKey = options.cacheKey
    if (cacheKey) {
      const cached = this.buildCache.get(cacheKey)
      if (cached) {
        return cached
      }
    }

    const indexedMenus = this.indexMenus(menuList)
    const spaceStore = useMenuSpaceStore()
    spaceStore.syncRuntimeHost()
    const currentSpaceKey = spaceStore.currentSpaceKey
    const defaultSpaceKey = spaceStore.defaultSpaceKey
    const pageMap = new Map<string, RuntimePageItem>()
    pages.forEach((item) => {
      const pageKey = this.normalizeValue(item.pageKey)
      if (!pageKey) return
      pageMap.set(pageKey, item)
    })

    const resolvedCache = new Map<string, ResolvedPageConfig>()
    const resolving = new Set<string>()
    const routePathCache = new Map<string, string>()
    const parentChainCache = new Map<string, boolean>()
    const parentChainResolving = new Set<string>()
    const spaceCache = new Map<string, string>()
    const spaceResolving = new Set<string>()

    const hasAvailableParentChain = (pageKey: string): boolean => {
      const normalizedKey = this.normalizeValue(pageKey)
      if (!normalizedKey) {
        return false
      }
      const cached = parentChainCache.get(normalizedKey)
      if (cached !== undefined) {
        return cached
      }
      if (parentChainResolving.has(normalizedKey)) {
        return false
      }

      const page = pageMap.get(normalizedKey)
      if (!page) {
        parentChainCache.set(normalizedKey, false)
        return false
      }

      const parentPageKey = this.normalizeValue(page.parentPageKey)
      if (!parentPageKey) {
        parentChainCache.set(normalizedKey, true)
        return true
      }

      const parentPage = pageMap.get(parentPageKey)
      if (!parentPage) {
        parentChainCache.set(normalizedKey, false)
        return false
      }

      parentChainResolving.add(normalizedKey)
      const available = hasAvailableParentChain(parentPage.pageKey || parentPageKey)
      parentChainResolving.delete(normalizedKey)
      parentChainCache.set(normalizedKey, available)
      return available
    }

    const resolvePage = (pageKey: string): ResolvedPageConfig => {
      const normalizedKey = this.normalizeValue(pageKey)
      if (!normalizedKey) {
        return this.createDefaultResolvedConfig()
      }
      const cached = resolvedCache.get(normalizedKey)
      if (cached) {
        return cached
      }
      if (resolving.has(normalizedKey)) {
        return { ...this.createDefaultResolvedConfig(), allowed: false }
      }

      const page = pageMap.get(normalizedKey)
      if (!page) {
        return { ...this.createDefaultResolvedConfig(), allowed: false }
      }

      resolving.add(normalizedKey)
      const activePath = this.resolveActiveMenuPath(page, indexedMenus, pageMap, resolvePage)
      const breadcrumbChain = this.resolveBreadcrumbChain(page, indexedMenus, pageMap, resolvePage)
      const customParent = activePath
      const resolvedSpaceKey = resolvePageSpaceKey(pageKey)
      if (
        !options.trustBackend &&
        !isMenuSpaceVisible(resolvedSpaceKey, currentSpaceKey, defaultSpaceKey)
      ) {
        resolving.delete(normalizedKey)
        resolvedCache.set(normalizedKey, { ...this.createDefaultResolvedConfig(), allowed: false })
        return { ...this.createDefaultResolvedConfig(), allowed: false }
      }
      const permissionConfig = this.resolvePermissionConfig(
        page,
        indexedMenus,
        pageMap,
        userInfo,
        resolvePage,
        activePath,
        Boolean(options.trustBackend)
      )

      const resolved: ResolvedPageConfig = {
        ...permissionConfig,
        activePath,
        customParent,
        breadcrumbChain,
        spaceKey: resolvedSpaceKey,
        spaceType: this.normalizeValue(page.spaceType),
        hostKey: this.normalizeValue(page.hostKey)
      }
      resolving.delete(normalizedKey)
      resolvedCache.set(normalizedKey, resolved)
      return resolved
    }

    const resolvePageSpaceKey = (pageKey: string): string => {
      const normalizedKey = this.normalizeValue(pageKey)
      if (!normalizedKey) {
        return currentSpaceKey || defaultSpaceKey
      }
      const cached = spaceCache.get(normalizedKey)
      if (cached !== undefined) {
        return cached
      }
      if (spaceResolving.has(normalizedKey)) {
        return currentSpaceKey || defaultSpaceKey
      }

      const page = pageMap.get(normalizedKey)
      if (!page) {
        return currentSpaceKey || defaultSpaceKey
      }

      spaceResolving.add(normalizedKey)
      const visibilityScope = this.normalizeValue(page.visibilityScope || page.spaceScope)
      const rawSpaceKeys = Array.isArray(page.spaceKeys) ? page.spaceKeys : []
      const resolvedSpaceKeys = rawSpaceKeys
        .map((item: unknown) => normalizeMenuSpaceKey(`${item || ''}`))
        .filter(Boolean)

      let resolved = ''
      if (visibilityScope === 'app') {
        resolved = currentSpaceKey || defaultSpaceKey
      } else if (visibilityScope === 'spaces') {
        resolved =
          resolvedSpaceKeys.find((item: string) => item === currentSpaceKey) ||
          resolvedSpaceKeys[0] ||
          currentSpaceKey ||
          defaultSpaceKey
      } else {
        const parentPageKey = this.normalizeValue(page.parentPageKey)
        if (parentPageKey) {
          resolved = resolvePageSpaceKey(parentPageKey)
        }
        if (!resolved) {
          const parentMenuId = this.normalizeValue(page.parentMenuId)
          if (parentMenuId) {
            const parentMenu = indexedMenus.byId.get(parentMenuId)
            resolved = this.normalizeValue(
              parentMenu?.spaceKey || parentMenu?.meta?.spaceKey || parentMenu?.meta?.space_key
            )
          }
        }
      }
      if (!resolved) {
        resolved = currentSpaceKey || defaultSpaceKey
      }
      spaceResolving.delete(normalizedKey)
      spaceCache.set(normalizedKey, resolved)
      return resolved
    }

    const resolvePageRoutePath = (pageKey: string): string => {
      const normalizedKey = this.normalizeValue(pageKey)
      if (!normalizedKey) {
        return ''
      }
      const cached = routePathCache.get(normalizedKey)
      if (cached !== undefined) {
        return cached
      }
      const page = pageMap.get(normalizedKey)
      if (!page) {
        return ''
      }
      const resolvedPath = resolveManagedPageRoutePath(page, {
        getPageByKey: (key) => pageMap.get(key),
        getMenuPathById: (menuId) => {
          const menu = indexedMenus.byId.get(this.normalizeValue(menuId))
          return this.normalizePath(menu?.path)
        }
      })
      routePathCache.set(normalizedKey, resolvedPath)
      return resolvedPath
    }

    const runtimeRoutes: AppRouteRecord[] = []
    for (const page of pages) {
      if (!this.isRuntimePage(page)) {
        continue
      }

      const pageKey = this.normalizeValue(page.pageKey)
      const routeName = this.normalizeValue(page.routeName) || pageKey
      if (!hasAvailableParentChain(pageKey)) {
        continue
      }
      const routePath = resolvePageRoutePath(pageKey)
      const component = this.normalizeValue(page.component)
      if (!pageKey || !routeName || !routePath || !component) {
        continue
      }

      const resolved = resolvePage(pageKey)
      if (!resolved.allowed) {
        continue
      }

      if (
        indexedMenus.names.has(routeName) ||
        runtimeRoutes.some((item) => item.name === routeName)
      ) {
        if (this.isExpectedMenuBackedDuplicate(page, routeName, routePath, indexedMenus)) {
          continue
        }
        this.warnDuplicate('name', routeName, pageKey)
        continue
      }
      if (
        indexedMenus.paths.has(routePath) ||
        runtimeRoutes.some((item) => item.path === routePath)
      ) {
        if (this.isExpectedMenuBackedDuplicate(page, routeName, routePath, indexedMenus)) {
          continue
        }
        this.warnDuplicate('path', routePath, pageKey)
        continue
      }

      runtimeRoutes.push({
        path: routePath,
        name: routeName,
        component,
        sort_order: page.sortOrder ?? 0,
        meta: this.buildRouteMeta(page, resolved)
      })
    }

    if (cacheKey) {
      this.buildCache.set(cacheKey, runtimeRoutes)
    }
    return runtimeRoutes
  }

  private buildRouteMeta(page: RuntimePageItem, resolved: ResolvedPageConfig): RouteMeta {
    const pageMeta = (page.meta || {}) as Record<string, unknown>
    const meta: RouteMeta = {
      title: this.normalizeValue(page.name) || this.normalizeValue(page.pageKey) || '未命名页面',
      isHide: true,
      isInnerPage: true,
      keepAlive: Boolean(page.keepAlive),
      isFullPage: Boolean(page.isFullPage),
      isIframe: Boolean(pageMeta.isIframe),
      isHideTab: Boolean(pageMeta.isHideTab),
      link: this.normalizeValue(pageMeta.link),
      isEnable: page.status === 'normal',
      accessMode: resolved.effectiveAccessMode,
      appKey: this.normalizeValue(page.appKey)
    }

    if (resolved.spaceKey) {
      meta.spaceKey = resolved.spaceKey
    }
    if (resolved.spaceType) {
      meta.spaceType = resolved.spaceType
    }
    if (resolved.hostKey) {
      meta.hostKey = resolved.hostKey
    }
    if (resolved.activePath) {
      meta.activePath = resolved.activePath
    }
    if (resolved.customParent) {
      meta.customParent = resolved.customParent
    }
    if (resolved.breadcrumbChain.length) {
      meta.breadcrumbChain = resolved.breadcrumbChain
    }
    if (resolved.requiredAction) {
      meta.requiredAction = resolved.requiredAction
    }
    if (resolved.requiredActions?.length) {
      meta.requiredActions = resolved.requiredActions
    }
    if (resolved.actionMatchMode) {
      meta.actionMatchMode = resolved.actionMatchMode
    }
    if (resolved.actionVisibilityMode) {
      meta.actionVisibilityMode = resolved.actionVisibilityMode
    }

    return meta
  }

  private resolvePermissionConfig(
    page: RuntimePageItem,
    indexedMenus: IndexedMenus,
    pageMap: Map<string, RuntimePageItem>,
    userInfo: UserInfo,
    resolvePage: (pageKey: string) => ResolvedPageConfig,
    activePath: string,
    trustBackend: boolean
  ): ResolvedPageConfig {
    if (trustBackend) {
      // runtime/navigation 与 /pages/runtime 已经在后端完成 AccessGraph 编译，
      // 这里不再重复做菜单/页面权限裁剪，只负责把显式结果映射成路由 meta。
      return this.resolveCompiledPermissionConfig(
        page,
        indexedMenus,
        pageMap,
        resolvePage,
        activePath
      )
    }

    const isAuthenticated = this.isAuthenticated(userInfo)
    const accessMode = this.normalizeAccessMode(page.accessMode)
    if (accessMode === 'public') {
      return this.createDefaultResolvedConfig('public')
    }
    if (accessMode === 'jwt') {
      return isAuthenticated
        ? this.createDefaultResolvedConfig('jwt')
        : { ...this.createDefaultResolvedConfig('jwt'), allowed: false }
    }

    if (accessMode === 'permission') {
      const permissionKey = this.normalizeValue(page.permissionKey)
      const allowed = permissionKey ? hasScopedActionPermission(userInfo, permissionKey) : false
      return {
        allowed,
        activePath: '',
        breadcrumbChain: [],
        customParent: '',
        effectiveAccessMode: 'permission',
        requiredAction: permissionKey || undefined,
        actionMatchMode: 'any',
        actionVisibilityMode: 'hide'
      }
    }

    const parentPageKey = this.normalizeValue(page.parentPageKey)
    if (parentPageKey) {
      const parentPage = pageMap.get(parentPageKey)
      if (parentPage) {
        return {
          ...resolvePage(parentPage.pageKey || parentPageKey),
          activePath: '',
          breadcrumbChain: [],
          customParent: ''
        }
      }
    }

    if (!activePath) {
      return isAuthenticated
        ? this.createDefaultResolvedConfig('jwt')
        : { ...this.createDefaultResolvedConfig('jwt'), allowed: false }
    }

    const parentMenu = indexedMenus.byPath.get(activePath)
    if (!parentMenu) {
      return { ...this.createDefaultResolvedConfig(), allowed: false }
    }

    const parentMenuAccessMode = this.normalizeAccessMode(parentMenu.meta?.accessMode as string)
    if (parentMenuAccessMode === 'public') {
      return this.createDefaultResolvedConfig('public')
    }
    if (parentMenuAccessMode === 'jwt') {
      return isAuthenticated
        ? this.createDefaultResolvedConfig('jwt')
        : { ...this.createDefaultResolvedConfig('jwt'), allowed: false }
    }
    if (!isAuthenticated) {
      return { ...this.createDefaultResolvedConfig('permission'), allowed: false }
    }

    const requirement = getMenuActionRequirement(parentMenu.meta)
    if (!requirement.actions.length) {
      return this.createDefaultResolvedConfig('permission')
    }

    const requiredAction = requirement.actions.length === 1 ? requirement.actions[0] : undefined
    const requiredActions = requirement.actions.length > 1 ? requirement.actions : undefined

    return {
      allowed:
        hasMenuActionAccess(userInfo, parentMenu.meta) ||
        !shouldHideMenuWhenActionDenied(parentMenu.meta),
      activePath: '',
      breadcrumbChain: [],
      customParent: '',
      effectiveAccessMode: 'permission',
      requiredAction,
      requiredActions,
      actionMatchMode: requirement.matchMode,
      actionVisibilityMode: requirement.visibilityMode
    }
  }

  private resolveActiveMenuPath(
    page: RuntimePageItem,
    indexedMenus: IndexedMenus,
    pageMap: Map<string, RuntimePageItem>,
    resolvePage: (pageKey: string) => ResolvedPageConfig
  ): string {
    const explicitActivePath = this.normalizePath(page.activeMenuPath)
    if (explicitActivePath && indexedMenus.byPath.has(explicitActivePath)) {
      return explicitActivePath
    }

    const parentMenuId = this.normalizeValue(page.parentMenuId)
    if (parentMenuId) {
      const parentMenu = indexedMenus.byId.get(parentMenuId)
      if (parentMenu?.path) {
        return this.normalizePath(parentMenu.path)
      }
    }

    const parentPageKey = this.normalizeValue(page.parentPageKey)
    if (!parentPageKey) {
      return ''
    }

    const parentPage = pageMap.get(parentPageKey)
    if (!parentPage) {
      return ''
    }
    return resolvePage(parentPage.pageKey).activePath
  }

  private resolveBreadcrumbChain(
    page: RuntimePageItem,
    indexedMenus: IndexedMenus,
    pageMap: Map<string, RuntimePageItem>,
    resolvePage: (pageKey: string) => ResolvedPageConfig
  ): Array<{ title: string; path?: string }> {
    const breadcrumbMode = this.normalizeBreadcrumbMode(page.breadcrumbMode)
    if (breadcrumbMode === 'inherit_page') {
      const chain = this.resolveParentPageBreadcrumbChain(page, pageMap, resolvePage, indexedMenus)
      if (chain.length) {
        return chain
      }
    }
    return this.resolveMenuBreadcrumbChain(
      this.resolveActiveMenuPath(page, indexedMenus, pageMap, resolvePage),
      indexedMenus
    )
  }

  private resolveParentPageBreadcrumbChain(
    page: RuntimePageItem,
    pageMap: Map<string, RuntimePageItem>,
    resolvePage: (pageKey: string) => ResolvedPageConfig,
    indexedMenus: IndexedMenus
  ): Array<{ title: string; path?: string }> {
    const chain: Array<{ title: string; path?: string }> = []
    const seen = new Set<string>()
    let parentKey = this.normalizeValue(page.parentPageKey)

    while (parentKey) {
      if (seen.has(parentKey)) {
        return []
      }
      seen.add(parentKey)

      const parentPage = pageMap.get(parentKey)
      if (!parentPage) {
        return []
      }
      if (this.normalizePageType(parentPage.pageType) !== 'group') {
        chain.push({
          title: this.normalizeValue(parentPage.name) || parentKey,
          path: this.normalizePath(parentPage.routePath) || undefined
        })
      }

      parentKey = this.normalizeValue(parentPage.parentPageKey)
    }

    if (!chain.length) {
      return []
    }

    chain.reverse()
    const firstParentKey = this.normalizeValue(page.parentPageKey)
    const firstResolved = firstParentKey
      ? resolvePage(firstParentKey)
      : this.createDefaultResolvedConfig()
    const menuChain = this.resolveMenuBreadcrumbChain(firstResolved.activePath, indexedMenus)
    if (!menuChain.length) {
      return chain
    }
    return [...menuChain, ...chain]
  }

  private resolveMenuBreadcrumbChain(
    activePath: string,
    indexedMenus: IndexedMenus
  ): Array<{ title: string; path?: string }> {
    const targetPath = this.normalizePath(activePath)
    if (!targetPath || !indexedMenus.byPath.size) {
      return []
    }
    const targetMenu = indexedMenus.byPath.get(targetPath)
    if (!targetMenu) {
      return []
    }

    const chain: Array<{ title: string; path?: string }> = []
    let current: AppRouteRecord | undefined = targetMenu
    const safety = new Set<string>()

    while (current) {
      const title = this.normalizeValue(current.meta?.title) || this.normalizeValue(current.name)
      if (title) {
        chain.push({
          title,
          path: this.normalizePath(current.path) || undefined
        })
      }

      const currentId = this.normalizeValue(current.id)
      if (!currentId || safety.has(currentId)) {
        break
      }
      safety.add(currentId)

      const parentId = this.normalizeValue(current.parent_id)
      if (!parentId) {
        break
      }
      current = indexedMenus.byId.get(parentId)
    }
    chain.reverse()
    return chain
  }

  private indexMenus(menuList: AppRouteRecord[]): IndexedMenus {
    const byId = new Map<string, AppRouteRecord>()
    const byName = new Map<string, AppRouteRecord>()
    const byPath = new Map<string, AppRouteRecord>()
    const names = new Set<string>()
    const paths = new Set<string>()

    const walk = (items: AppRouteRecord[]) => {
      items.forEach((item) => {
        const id = this.normalizeValue(item.id)
        const path = this.normalizePath(item.path)
        const name = this.normalizeValue(item.name)

        if (id) {
          byId.set(id, item)
        }
        if (path) {
          byPath.set(path, item)
          paths.add(path)
        }
        if (name) {
          byName.set(name, item)
          names.add(name)
        }
        if (Array.isArray(item.children) && item.children.length > 0) {
          walk(item.children)
        }
      })
    }

    walk(menuList)
    return { byId, byName, byPath, names, paths }
  }

  private resolveCompiledPermissionConfig(
    page: RuntimePageItem,
    indexedMenus: IndexedMenus,
    pageMap: Map<string, RuntimePageItem>,
    resolvePage: (pageKey: string) => ResolvedPageConfig,
    activePath: string
  ): ResolvedPageConfig {
    const accessMode = this.normalizeAccessMode(page.accessMode)
    if (accessMode === 'public') {
      return this.createDefaultResolvedConfig('public')
    }
    if (accessMode === 'jwt') {
      return this.createDefaultResolvedConfig('jwt')
    }
    if (accessMode === 'permission') {
      const permissionKey = this.normalizeValue(page.permissionKey)
      return {
        allowed: true,
        activePath: '',
        breadcrumbChain: [],
        customParent: '',
        effectiveAccessMode: 'permission',
        requiredAction: permissionKey || undefined,
        actionMatchMode: 'any',
        actionVisibilityMode: 'show'
      }
    }

    const parentPageKey = this.normalizeValue(page.parentPageKey)
    if (parentPageKey) {
      const parentResolved = resolvePage(parentPageKey)
      return {
        ...parentResolved,
        allowed: true
      }
    }

    const menuNode =
      indexedMenus.byPath.get(this.normalizePath(activePath)) ||
      indexedMenus.byId.get(this.normalizeValue(page.parentMenuId))
    const accessModeFromMenu = this.normalizeMenuAccessMode(menuNode?.meta)
    const requirement = getMenuActionRequirement(menuNode?.meta)
    return {
      allowed: true,
      activePath: '',
      breadcrumbChain: [],
      customParent: '',
      effectiveAccessMode: accessModeFromMenu,
      requiredAction: requirement.actions[0],
      requiredActions: requirement.actions.length > 1 ? requirement.actions : undefined,
      actionMatchMode: requirement.actions.length ? requirement.matchMode : undefined,
      actionVisibilityMode: requirement.actions.length ? requirement.visibilityMode : undefined
    }
  }

  private isExpectedMenuBackedDuplicate(
    page: RuntimePageItem,
    routeName: string,
    routePath: string,
    indexedMenus: IndexedMenus
  ): boolean {
    const menuByPath = indexedMenus.byPath.get(routePath)
    if (menuByPath && this.normalizeValue(menuByPath.name) === routeName) {
      return true
    }

    const menuByName = indexedMenus.byName.get(routeName)
    if (menuByName && this.normalizePath(menuByName.path) === routePath) {
      return true
    }

    const parentMenuId = this.normalizeValue(page.parentMenuId)
    if (!parentMenuId) {
      return false
    }
    const parentMenu = indexedMenus.byId.get(parentMenuId)
    if (!parentMenu) {
      return false
    }
    return (
      this.normalizeValue(parentMenu.name) === routeName &&
      this.normalizePath(parentMenu.path) === routePath
    )
  }

  private isRuntimePage(page: RuntimePageItem): boolean {
    return page.status === 'normal' && this.normalizePageType(page.pageType) !== 'group'
  }

  private normalizePageType(value?: string): string {
    const target = this.normalizeValue(value)
    if (target === 'inner' || target === 'standalone' || target === 'group') {
      return target
    }
    return 'inner'
  }

  private normalizeAccessMode(value?: string): string {
    const target = this.normalizeValue(value)
    if (
      target === 'public' ||
      target === 'jwt' ||
      target === 'permission' ||
      target === 'inherit'
    ) {
      return target
    }
    return 'inherit'
  }

  private normalizeMenuAccessMode(meta?: AppRouteRecord['meta']): 'public' | 'jwt' | 'permission' {
    const target = this.normalizeValue(meta?.accessMode)
    if (target === 'public' || target === 'jwt' || target === 'permission') {
      return target
    }
    return 'permission'
  }

  private normalizeBreadcrumbMode(value?: string): string {
    const target = this.normalizeValue(value)
    if (target === 'inherit_menu' || target === 'inherit_page' || target === 'custom') {
      return target
    }
    return 'inherit_menu'
  }

  private normalizePath(path?: string): string {
    const target = this.normalizeValue(path)
    if (!target) {
      return ''
    }
    if (/^https?:\/\//i.test(target)) {
      return target
    }
    const normalized = `/${target.replace(/^\/+/, '')}`.replace(/\/+/g, '/')
    return normalized !== '/' ? normalized.replace(/\/$/, '') : normalized
  }

  private normalizeValue(value?: unknown): string {
    return `${value ?? ''}`.trim()
  }

  private isAuthenticated(userInfo: UserInfo): boolean {
    if (!userInfo) {
      return false
    }
    return Boolean(
      this.normalizeValue(userInfo.id) ||
      this.normalizeValue(userInfo.userId) ||
      this.normalizeValue(userInfo.email) ||
      this.normalizeValue(userInfo.username)
    )
  }

  private createDefaultResolvedConfig(
    effectiveAccessMode: 'public' | 'jwt' | 'permission' = 'jwt'
  ): ResolvedPageConfig {
    return {
      allowed: true,
      activePath: '',
      customParent: '',
      breadcrumbChain: [],
      effectiveAccessMode
    }
  }

  private warnDuplicate(field: 'name' | 'path', value: string, pageKey: string): void {
    if (!import.meta.env.DEV) {
      return
    }
    console.warn(
      `[ManagedPageProcessor] 页面 ${pageKey} 的${field}(${value}) 与现有动态路由重复，已跳过注册`
    )
  }
}
