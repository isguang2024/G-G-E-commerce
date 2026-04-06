/**
 * 路由全局前置守卫模块
 *
 * 提供完整的路由导航守卫功能
 *
 * ## 主要功能
 *
 * - 登录状态验证和重定向
 * - 动态路由注册和权限控制
 * - 菜单数据获取和处理（前端/后端模式）
 * - 用户信息获取和缓存
 * - 页面标题设置
 * - 工作标签页管理
 * - 进度条和加载动画控制
 * - 静态路由识别和处理
 * - 错误处理和异常跳转
 *
 * ## 使用场景
 *
 * - 路由跳转前的权限验证
 * - 动态菜单加载和路由注册
 * - 用户登录状态管理
 * - 页面访问控制
 * - 路由级别的加载状态管理
 *
 * ## 工作流程
 *
 * 1. 检查登录状态，未登录跳转到登录页
 * 2. 首次访问时获取用户信息和菜单数据
 * 3. 根据权限动态注册路由
 * 4. 设置页面标题和工作标签页
 * 5. 处理根路径重定向到首页
 * 6. 未匹配路由跳转到 404 页面
 *
 * @module router/guards/beforeEach
 * @author Art Design Pro Team
 */
import type { Router, RouteLocationNormalized, NavigationGuardNext } from 'vue-router'
import type { AppRouteRecord } from '@/types/router'
import { nextTick } from 'vue'
import NProgress from 'nprogress'
import { useSettingStore } from '@/store/modules/setting'
import { useUserStore } from '@/store/modules/user'
import { useMenuStore } from '@/store/modules/menu'
import { setWorktab } from '@/utils/navigation'
import { hasRegisteredRoutePath, setPageTitle } from '@/utils/router'
import { RoutesAlias } from '../routesAlias'
import { staticRoutes } from '../routes/staticRoutes'
import { loadingService } from '@/utils/ui'
import { useCommon } from '@/hooks/core/useCommon'
import { useWorktabStore } from '@/store/modules/worktab'
import {
  hasPlatformAccessByUserInfo,
  useCollaborationWorkspaceStore
} from '@/store/modules/collaboration-workspace'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { useMenuSpaceStore } from '@/store/modules/menu-space'
import { useAppContextStore } from '@/store/modules/app-context'
import { fetchGetUserInfo } from '@/api/auth'
import { fetchGetRuntimeNavigation, fetchGetRuntimePublicPageList } from '@/api/system-manage'
import { ApiStatus } from '@/utils/http/status'
import { isHttpError } from '@/utils/http/error'
import { RouteRegistry, MenuProcessor, IframeRouteManager, ManagedPageProcessor } from '../core'
import { normalizeMenuSpaceKey } from '@/utils/navigation/menu-space'

// 路由注册器实例
let routeRegistry: RouteRegistry | null = null
let routeRegistrationMode: 'none' | 'public' | 'authenticated' = 'none'

// 菜单处理器实例
const menuProcessor = new MenuProcessor()
const managedPageProcessor = new ManagedPageProcessor()

// 跟踪是否需要关闭 loading
let pendingLoading = false

// 路由初始化失败标记，防止死循环
// 一旦设置为 true，只有刷新页面或重新登录才能重置
let routeInitFailed = false

// 路由初始化进行中标记，防止并发请求
let routeInitInProgress = false
const routeRefreshAttempted = new Set<string>()

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
  const manifest = await fetchGetRuntimeNavigation(requestedSpaceKey)
  const resolvedAppKey =
    `${manifest.currentApp?.app?.appKey || manifest.context?.app_key || ''}`.trim()
  if (resolvedAppKey) {
    appContextStore.setRuntimeAppKey(resolvedAppKey)
  }
  const resolvedSpaceKey = normalizeMenuSpaceKey(
    manifest.currentSpace?.space?.spaceKey || manifest.context?.space_key || requestedSpaceKey
  )

  if (
    resolvedSpaceKey &&
    resolvedSpaceKey !== normalizeMenuSpaceKey(menuSpaceStore.currentSpaceKey)
  ) {
    menuSpaceStore.setActiveSpaceKey(resolvedSpaceKey)
  }

  const menuList = menuProcessor.normalizeMenuList(manifest.menuTree || [])
  const runtimeRoutes = managedPageProcessor.buildRoutes(
    menuList,
    manifest.managedPages || [],
    userStore.getUserInfo as Api.Auth.UserInfo,
    { trustBackend: true }
  )

  return {
    menuList,
    registeredRoutes: [...menuList, ...runtimeRoutes],
    manifest
  }
}

/**
 * 获取 pendingLoading 状态
 */
export function getPendingLoading(): boolean {
  return pendingLoading
}

/**
 * 重置 pendingLoading 状态
 */
export function resetPendingLoading(): void {
  pendingLoading = false
}

/**
 * 获取路由初始化失败状态
 */
export function getRouteInitFailed(): boolean {
  return routeInitFailed
}

/**
 * 重置路由初始化状态（用于重新登录场景）
 */
export function resetRouteInitState(): void {
  routeInitFailed = false
  routeInitInProgress = false
  routeRegistrationMode = 'none'
  routeRefreshAttempted.clear()
}

/**
 * 重新拉取当前用户菜单并更新侧栏与动态路由（角色菜单权限保存后调用，使新勾选的菜单立即生效）
 */
export async function refreshUserMenus(): Promise<void> {
  if (!routeRegistry) return
  const menuStore = useMenuStore()
  try {
    const { menuList, registeredRoutes } = await buildRegisteredRoutesFromManifest()
    routeRegistry.unregister()
    routeRegistry.register(registeredRoutes)
    routeRegistrationMode = 'authenticated'
    syncHomePathWithRegisteredRoutes(menuList, registeredRoutes)
    menuStore.clearRemoveRouteFns()
    menuStore.addRemoveRouteFns(routeRegistry.getRemoveRouteFns())
    IframeRouteManager.getInstance().save()
    routeRefreshAttempted.clear()
  } catch (e) {
    console.error('[RouteGuard] refreshUserMenus failed', e)
  }
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

export async function refreshUserAccessAndMenus(): Promise<void> {
  await refreshCurrentUserInfoContext()
  await refreshUserMenus()
}

function buildFrontendUserInfo(data: Api.Auth.UserInfo): Api.Auth.UserInfo {
  const roles = mapBackendRolesToFrontend(data)
  return {
    ...data,
    userId: data.id,
    userName: data.username || data.email,
    avatar: data.avatar_url,
    roles,
    buttons: data.buttons || [],
    actions: data.actions || []
  }
}

export async function refreshCurrentUserInfoContext(): Promise<void> {
  const userStore = useUserStore()
  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const workspaceStore = useWorkspaceStore()
  const menuSpaceStore = useMenuSpaceStore()
  const data = await fetchGetUserInfo()
  const mergedInfo: Api.Auth.UserInfo = {
    ...(userStore.getUserInfo as Api.Auth.UserInfo),
    ...buildFrontendUserInfo(data)
  }
  userStore.setUserInfo(mergedInfo)
  collaborationWorkspaceStore.setPlatformAccess(hasPlatformAccessByUserInfo(mergedInfo))
  await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
    preferredCollaborationWorkspaceId: data.current_collaboration_workspace_id || '',
    preferredLegacyCollaborationWorkspaceId:
      data.collaboration_workspace_id || data.current_collaboration_workspace_id || '',
    preferredWorkspaceId:
      workspaceStore.currentAuthWorkspaceId || data.current_auth_workspace_id || '',
    preferredWorkspaceType:
      workspaceStore.currentAuthWorkspaceType || data.current_auth_workspace_type || '',
    preferPlatform: hasPlatformAccessByUserInfo(mergedInfo)
  })
  menuSpaceStore.syncRuntimeHost()
  await menuSpaceStore.refreshRuntimeConfig(true)
  await menuSpaceStore.syncResolvedCurrentSpace()
}

/**
 * 设置路由全局前置守卫
 */
export function setupBeforeEachGuard(router: Router): void {
  // 初始化路由注册器
  routeRegistry = new RouteRegistry(router)

  router.beforeEach(
    async (
      to: RouteLocationNormalized,
      from: RouteLocationNormalized,
      next: NavigationGuardNext
    ) => {
      try {
        await handleRouteGuard(to, from, next, router)
      } catch (error) {
        console.error('[RouteGuard] 路由守卫处理失败:', error)
        closeLoading()
        next({ name: 'Exception500' })
      }
    }
  )
}

/**
 * 关闭 loading 效果
 */
function closeLoading(): void {
  if (pendingLoading) {
    nextTick(() => {
      loadingService.hideLoading()
      pendingLoading = false
    })
  }
}

/**
 * 处理路由守卫逻辑
 */
async function handleRouteGuard(
  to: RouteLocationNormalized,
  from: RouteLocationNormalized,
  next: NavigationGuardNext,
  router: Router
): Promise<void> {
  const settingStore = useSettingStore()
  const userStore = useUserStore()
  const preferredSpaceKey = resolvePreferredSpaceKeyFromRoute(to)

  // 启动进度条
  if (settingStore.showNprogress) {
    NProgress.start()
  }

  // 0. 未登录时，先尝试注册公开运行时页面
  if (!userStore.isLogin && !isStaticRoute(to.path) && to.path !== RoutesAlias.Login) {
    const alreadyMatchedPublicRoute = isPublicRuntimeRoute(to)
    const matchedPublicRoute = await ensurePublicRuntimeRoutes(to, router)
    if (matchedPublicRoute && !alreadyMatchedPublicRoute) {
      next({
        path: to.path,
        query: to.query,
        hash: to.hash,
        replace: true
      })
      return
    }
  }

  // 1. 检查登录状态
  if (!handleLoginStatus(to, userStore, next)) {
    return
  }

  // 2. 检查路由初始化是否已失败（防止死循环）
  if (routeInitFailed) {
    // 已经失败过，直接放行到错误页面，不再重试
    if (to.matched.length > 0) {
      next()
    } else {
      // 未匹配到路由，跳转到 500 页面
      next({ name: 'Exception500', replace: true })
    }
    return
  }

  // 3. 处理动态路由注册
  if (userStore.isLogin && routeRegistrationMode !== 'authenticated') {
    // 防止并发请求（快速连续导航场景）
    if (routeInitInProgress) {
      // 正在初始化中，等待完成后重新导航
      next(false)
      return
    }
    await handleDynamicRoutes(to, from, next, router)
    return
  }

  // 3.1 已登录场景支持通过 URL 显式切换菜单空间（如 #/system/menu?space_key=ops）
  if (userStore.isLogin && preferredSpaceKey) {
    const menuSpaceStore = useMenuSpaceStore()
    if (normalizeMenuSpaceKey(menuSpaceStore.currentSpaceKey) !== preferredSpaceKey) {
      menuSpaceStore.setActiveSpaceKey(preferredSpaceKey)
      await menuSpaceStore.syncResolvedCurrentSpace(preferredSpaceKey)
      if (routeRegistrationMode === 'authenticated') {
        await refreshUserMenus()
      }
      next({
        path: to.path,
        query: to.query,
        hash: to.hash,
        replace: true
      })
      return
    }
  }

  // 4. 处理根路径重定向
  if (handleRootPathRedirect(to, next)) {
    return
  }

  // 5. 处理已匹配的路由
  if (to.matched.length > 0) {
    routeRefreshAttempted.delete(to.fullPath)
    setWorktab(to)
    setPageTitle(to)
    next()
    return
  }

  if (await tryRefreshMissingDynamicRoute(to, next, router)) {
    return
  }

  // 6. 未匹配到路由，跳转到 404
  next({ name: 'Exception404', replace: true })
}

/**
 * 处理登录状态
 * @returns true 表示可以继续，false 表示已处理跳转
 */
function handleLoginStatus(
  to: RouteLocationNormalized,
  userStore: ReturnType<typeof useUserStore>,
  next: NavigationGuardNext
): boolean {
  // 已登录或访问登录页或静态路由，直接放行
  if (
    userStore.isLogin ||
    to.path === RoutesAlias.Login ||
    isStaticRoute(to.path) ||
    isPublicRuntimeRoute(to)
  ) {
    return true
  }

  // 未登录且访问需要权限的页面，跳转到登录页并携带 redirect 参数
  userStore.logOut()
  next({
    name: 'Login',
    query: { redirect: to.fullPath }
  })
  return false
}

async function ensurePublicRuntimeRoutes(
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
    const runtimeRes = await fetchGetRuntimePublicPageList(menuSpaceStore.currentSpaceKey)
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
    console.error('[RouteGuard] 注册公开运行时页面失败:', error)
    routeRegistry?.unregister()
    routeRegistrationMode = 'none'
    return false
  }
}

function isPublicRuntimeRoute(to: RouteLocationNormalized): boolean {
  return to.matched.some(
    (record) =>
      Boolean(record.meta?.isInnerPage) && `${record.meta?.accessMode || ''}`.trim() === 'public'
  )
}

/**
 * 检查路由是否为静态路由
 */
function isStaticRoute(path: string): boolean {
  const checkRoute = (routes: any[], targetPath: string): boolean => {
    return routes.some((route) => {
      // 处理动态路由参数匹配
      const routePath = route.path
      const pattern = routePath.replace(/:[^/]+/g, '[^/]+').replace(/\*/g, '.*')
      const regex = new RegExp(`^${pattern}$`)

      if (regex.test(targetPath)) {
        return true
      }
      if (route.children && route.children.length > 0) {
        return checkRoute(route.children, targetPath)
      }
      return false
    })
  }

  return checkRoute(staticRoutes, path)
}

/**
 * 处理动态路由注册
 */
async function handleDynamicRoutes(
  to: RouteLocationNormalized,
  from: RouteLocationNormalized,
  next: NavigationGuardNext,
  router: Router
): Promise<void> {
  // 标记初始化进行中
  routeInitInProgress = true

  // 显示 loading
  pendingLoading = true
  loadingService.showLoading()

  try {
    // 1. 获取用户信息
    await fetchUserInfo(resolvePreferredSpaceKeyFromRoute(to))

    // 2. 后端一次编译 navigation manifest，前端只做轻量归一化与路由注册
    const { menuList, registeredRoutes } = await buildRegisteredRoutesFromManifest(
      resolvePreferredSpaceKeyFromRoute(to)
    )

    // 3. 验证菜单数据
    // 4. 注册动态路由
    routeRegistry?.unregister()
    routeRegistry?.register(registeredRoutes)
    routeRegistrationMode = 'authenticated'

    // 5. 保存菜单数据到 store
    const menuStore = useMenuStore()
    syncHomePathWithRegisteredRoutes(menuList, registeredRoutes)
    menuStore.addRemoveRouteFns(routeRegistry?.getRemoveRouteFns() || [])
    const landingPath = menuStore.getHomePath() || '/'

    // 6. 保存 iframe 路由
    IframeRouteManager.getInstance().save()

    // 7. 验证工作标签页
    useWorktabStore().validateWorktabs(router)

    if (isStaticRoute(to.path)) {
      routeInitInProgress = false
      next({
        path: to.path,
        query: to.query,
        hash: to.hash,
        replace: true
      })
      return
    }

    // 初始化成功，重置进行中标记
    routeInitInProgress = false

    if (to.path === '/' && landingPath !== '/') {
      next({
        path: landingPath,
        replace: true
      })
      return
    }

    // 8. 重新导航到目标路由。目标路径是否可访问由后端 manifest 产出的动态路由决定，
    // 当前守卫不再重复做菜单/页面权限推导，只负责补注册后的二次命中。
    next({
      path: to.path,
      query: to.query,
      hash: to.hash,
      replace: true
    })
  } catch (error) {
    console.error('[RouteGuard] 动态路由注册失败:', error)

    // 关闭 loading
    closeLoading()

    // 401 错误：axios 拦截器已处理退出登录，取消当前导航
    if (isUnauthorizedError(error)) {
      // 重置状态，允许重新登录后再次初始化
      routeInitInProgress = false
      next(false)
      return
    }

    // 标记初始化失败，防止死循环
    routeInitFailed = true
    routeInitInProgress = false

    // 输出详细错误信息，便于排查
    if (isHttpError(error)) {
      console.error(`[RouteGuard] 错误码: ${error.code}, 消息: ${error.message}`)
    }

    // 跳转到 500 页面，使用 replace 避免产生历史记录
    next({ name: 'Exception500', replace: true })
  }
}

/**
 * 将后端角色 code 映射为前端菜单权限标识
 */
function mapBackendRolesToFrontend(data: {
  roles?: Array<{ code?: string }> | string[]
  is_super_admin?: boolean
}): string[] {
  if (data.is_super_admin) return ['R_SUPER']
  return ['R_USER']
}

/**
 * 获取用户信息
 */
async function fetchUserInfo(preferredSpaceKey = ''): Promise<void> {
  const userStore = useUserStore()
  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const workspaceStore = useWorkspaceStore()
  const menuSpaceStore = useMenuSpaceStore()
  const data = await fetchGetUserInfo()
  const frontendUserInfo = buildFrontendUserInfo(data)
  userStore.syncLoginUserIdentity(`${frontendUserInfo.userId || frontendUserInfo.id || ''}`.trim())
  userStore.setUserInfo(frontendUserInfo)
  userStore.checkAndClearWorktabs()
  collaborationWorkspaceStore.setPlatformAccess(hasPlatformAccessByUserInfo(frontendUserInfo))
  menuSpaceStore.syncRuntimeHost()
  await menuSpaceStore.refreshRuntimeConfig(true)
  await menuSpaceStore.syncResolvedCurrentSpace(preferredSpaceKey)
  await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
    preferredCollaborationWorkspaceId: data.current_collaboration_workspace_id || '',
    preferredLegacyCollaborationWorkspaceId:
      data.collaboration_workspace_id || data.current_collaboration_workspace_id || '',
    preferredWorkspaceId: data.current_auth_workspace_id || '',
    preferredWorkspaceType: data.current_auth_workspace_type || '',
    preferPlatform: hasPlatformAccessByUserInfo(frontendUserInfo)
  })
  if (
    workspaceStore.currentAuthWorkspaceType !== (data.current_auth_workspace_type || 'personal') ||
    workspaceStore.currentAuthWorkspaceId !== (data.current_auth_workspace_id || '')
  ) {
    await refreshCurrentUserInfoContext()
  }
}

/**
 * 重置路由相关状态
 */
export function resetRouterState(delay: number): void {
  setTimeout(() => {
    routeRegistry?.unregister()
    IframeRouteManager.getInstance().clear()

    const menuStore = useMenuStore()
    menuStore.removeAllDynamicRoutes()
    menuStore.setMenuList([])

    // 重置路由初始化状态，允许重新登录后再次初始化
    resetRouteInitState()
  }, delay)
}

/**
 * 处理根路径重定向到首页
 * @returns true 表示已处理跳转，false 表示无需跳转
 */
function handleRootPathRedirect(to: RouteLocationNormalized, next: NavigationGuardNext): boolean {
  if (to.path !== '/') {
    return false
  }

  const { homePath } = useCommon()
  if (homePath.value && homePath.value !== '/') {
    next({ path: homePath.value, replace: true })
    return true
  }

  return false
}

/**
 * 判断是否为未授权错误（401）
 */
function isUnauthorizedError(error: unknown): boolean {
  return isHttpError(error) && error.code === ApiStatus.unauthorized
}

function resolvePreferredSpaceKeyFromRoute(to: RouteLocationNormalized): string {
  if (!to) {
    return ''
  }
  return normalizeMenuSpaceKey(to.query?.space_key || to.query?.spaceKey)
}

async function tryRefreshMissingDynamicRoute(
  to: RouteLocationNormalized,
  next: NavigationGuardNext,
  router: Router
): Promise<boolean> {
  const userStore = useUserStore()
  if (!userStore.isLogin || isStaticRoute(to.path) || isPublicRuntimeRoute(to)) {
    return false
  }
  if (routeRefreshAttempted.has(to.fullPath)) {
    routeRefreshAttempted.delete(to.fullPath)
    return false
  }

  routeRefreshAttempted.add(to.fullPath)
  try {
    await refreshUserAccessAndMenus()
    // 刷新访问图后若目标路由仍不存在，说明它已经不在当前运行时导航清单里，
    // 常见于菜单被禁用、空间切换后失效或权限被收回，此时直接落到 404，
    // 避免再次强跳原路径造成一串 "No match found" 警告。
    if (!hasRegisteredRoutePath(router, to.path)) {
      routeRefreshAttempted.delete(to.fullPath)
      next({ name: 'Exception404', replace: true })
      return true
    }
    next({
      path: to.path,
      query: to.query,
      hash: to.hash,
      replace: true
    })
    return true
  } catch (error) {
    console.error('[RouteGuard] 动态路由缺失自动刷新失败:', error)
    routeRefreshAttempted.delete(to.fullPath)
    return false
  }
}
