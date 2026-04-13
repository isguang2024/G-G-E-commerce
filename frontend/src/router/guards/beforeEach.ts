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
import { nextTick } from 'vue'
import NProgress from 'nprogress'
import { useAppContextStore } from '@/domains/app-runtime/context'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { useMenuStore } from '@/domains/navigation/menu'
import { fetchGetUserInfo } from '@/domains/auth/api'
import {
  buildCentralizedLoginURL,
  createCentralizedAuthAttempt,
  persistCentralizedAuthAttempt
} from '@/domains/auth/centralized-login'
import { attemptSilentSSO } from '@/domains/auth/silent-sso'
import { useUserStore } from '@/domains/auth/store'
import { useSettingStore } from '@/store/modules/setting'
import { setWorktab } from '@/domains/navigation/utils/worktab'
import { setPageTitle } from '@/utils/router'
import { RoutesAlias } from '../routesAlias'
import { staticRoutes } from '../routes/staticRoutes'
import { loadingService } from '@/utils/ui'
import { fetchGetCurrentApp, fetchGetRuntimeNavigation } from '@/domains/governance/api'
import { ApiStatus } from '@/utils/http/status'
import { isHttpError } from '@/utils/http/error'
import { normalizeMenuSpaceKey } from '@/domains/navigation/utils/menu-space'
import {
  ensureAuthenticatedRoutes,
  ensurePublicRuntimeRoutes,
  getRouteRegistrationMode,
  hasDynamicRoute,
  initNavigationRuntime,
  isPublicRuntimeRoute,
  refreshUserMenus
} from '@/domains/navigation/runtime/navigation'
import {
  addRouteRefreshAttempted,
  clearRouteRefreshAttempted,
  getPendingLoading,
  getRouteInitFailed,
  getRouteInitInProgress,
  hasRouteRefreshAttempted,
  resetPendingLoading,
  resetRouteInitState,
  setPendingLoading,
  setRouteInitFailed,
  setRouteInitInProgress
} from '@/domains/navigation/runtime/guard-state'

/**
 * 设置路由全局前置守卫
 */
export function setupBeforeEachGuard(router: Router): void {
  initNavigationRuntime(router)

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
        void closeLoading()
        next({ name: 'Exception500' })
      }
    }
  )
}

/**
 * 关闭 loading 效果
 */
async function closeLoading(): Promise<void> {
  if (getPendingLoading()) {
    await nextTick()
    loadingService.hideLoading()
    resetPendingLoading()
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
  if (!userStore.isLogin && !isStaticRoute(to.path)) {
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
  if (!(await handleLoginStatus(to, userStore, next))) {
    return
  }

  // 2. 检查路由初始化是否已失败（防止死循环）
  if (getRouteInitFailed()) {
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
  if (userStore.isLogin && getRouteRegistrationMode() !== 'authenticated') {
    // 若已有进行中的初始化，等待其完成后再决定（避免 next(false) 造成死锁）
    const routeInitInProgress = getRouteInitInProgress()
    if (routeInitInProgress) {
      try {
        await routeInitInProgress
      } catch {
        // 忽略 — 下面重新判断状态
      }
      if (getRouteInitFailed()) {
        next({ name: 'Exception500', replace: true })
        return
      }
      if ((getRouteRegistrationMode() as string) === 'authenticated') {
        // 已由并发的初始化完成注册，重新解析当前目标路径
        next({
          path: to.path,
          query: to.query,
          hash: to.hash,
          replace: true
        })
        return
      }
      // 状态仍未就绪，继续走一次常规初始化
    }
    await handleDynamicRoutes(to, from, next, router)
    return
  }

  // 3.1 已登录场景支持通过 URL 显式切换菜单空间（如 /system/menu?space_key=ops）
  if (userStore.isLogin && preferredSpaceKey) {
    const menuSpaceStore = useMenuSpaceStore()
    if (normalizeMenuSpaceKey(menuSpaceStore.currentSpaceKey) !== preferredSpaceKey) {
      menuSpaceStore.setActiveSpaceKey(preferredSpaceKey)
      await menuSpaceStore.syncResolvedCurrentSpace(preferredSpaceKey)
      if (getRouteRegistrationMode() === 'authenticated') {
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
      clearRouteRefreshAttempted(to.fullPath)
    setWorktab(to)
    setPageTitle(to)
    next()
    return
  }

  if (await tryRefreshMissingDynamicRoute(to, next, router)) {
    return
  }

  const crossDomainTarget = buildCrossDomainAppRedirectTarget(to)
  if (crossDomainTarget) {
    window.location.assign(crossDomainTarget)
    next(false)
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
): Promise<boolean> {
  // 已登录或访问登录页或静态路由，直接放行
  if (
    userStore.isLogin ||
    to.path === RoutesAlias.Login ||
    isStaticRoute(to.path) ||
    isPublicRuntimeRoute(to)
  ) {
    return Promise.resolve(true)
  }
  return resolveLoginRedirectPolicy(to, userStore, next)
}

async function resolveLoginRedirectPolicy(
  to: RouteLocationNormalized,
  userStore: ReturnType<typeof useUserStore>,
  next: NavigationGuardNext
): Promise<boolean> {
  const appContextStore = useAppContextStore()
  const menuSpaceStore = useMenuSpaceStore()
  const targetAppKey = await resolveTargetAppKey(to, appContextStore)
  const preferredSpaceKey = resolvePreferredSpaceKeyFromRoute(to)
  const binding = menuSpaceStore.resolveHostBinding(preferredSpaceKey)
  const ssoMode = appContextStore.resolveAppSsoMode(targetAppKey)
  const shouldUseCentralizedLogin = await resolveCentralizedLoginPolicy(
    targetAppKey,
    appContextStore
  )

  // isolated 模式：不走认证中心，直接本地登录页
  if (ssoMode === 'isolated') {
    userStore.clearSessionState({ broadcast: false })
    next({ path: RoutesAlias.Login, query: { redirect: to.fullPath } })
    return false
  }

  // local_ui 模式：APP 自有登录页，不走认证中心
  const loginUiMode = appContextStore.resolveAppLoginUiMode(targetAppKey)
  const loginPageKey =
    loginUiMode === 'auth_center_custom' ? appContextStore.resolveAppLoginPageKey(targetAppKey) : ''
  if (loginUiMode === 'local_ui') {
    userStore.clearSessionState({ broadcast: false })
    next({ path: RoutesAlias.Login, query: { redirect: to.fullPath } })
    return false
  }

  if (shouldUseCentralizedLogin && targetAppKey && typeof window !== 'undefined') {
    const redirectUri = new URL(RoutesAlias.AuthCallback, window.location.origin).toString()
    const spaceKey = preferredSpaceKey || binding?.spaceKey || ''

    // participate 模式：先尝试 silent SSO（利用已有的中心 token 静默签发 callback）
    if (ssoMode === 'participate') {
      const attempt = createCentralizedAuthAttempt(
        targetAppKey,
        to.fullPath,
        redirectUri,
        spaceKey,
        loginPageKey
      )
      persistCentralizedAuthAttempt(attempt)
      const silentResult = await attemptSilentSSO({
        targetAppKey,
        redirectUri,
        state: attempt.state,
        nonce: attempt.nonce,
        targetPath: to.fullPath,
        navigationSpaceKey: spaceKey,
        loginPageKey
      })
      if (silentResult?.callback?.redirect_to) {
        window.location.assign(silentResult.callback.redirect_to)
        next(false)
        return false
      }
      // silent SSO 失败，落入下方正常跳转登录页逻辑
    }

    userStore.clearSessionState({ broadcast: false })
    const attempt = createCentralizedAuthAttempt(
      targetAppKey,
      to.fullPath,
      redirectUri,
      spaceKey,
      loginPageKey
    )
    persistCentralizedAuthAttempt(attempt)
    window.location.assign(
      buildCentralizedLoginURL({
        loginHost: binding?.loginHost || '',
        targetAppKey,
        targetPath: to.fullPath,
        redirectUri,
        navigationSpaceKey: spaceKey,
        state: attempt.state,
        nonce: attempt.nonce,
        loginPageKey: attempt.loginPageKey,
        // reauth 模式：带 prompt=login 强制重新认证
        prompt: ssoMode === 'reauth' ? 'login' : undefined
      })
    )
    next(false)
    return false
  }

  userStore.clearSessionState({ broadcast: false })
  next({
    path: RoutesAlias.Login,
    query: { redirect: to.fullPath }
  })
  return false
}

function resolveRouteMetaAppKey(to: RouteLocationNormalized): string {
  for (const record of [...to.matched].reverse()) {
    const appKey = `${record.meta?.appKey || ''}`.trim()
    if (appKey) {
      return appKey
    }
  }
  return ''
}

async function resolveTargetAppKey(
  to: RouteLocationNormalized,
  appContextStore: ReturnType<typeof useAppContextStore>
): Promise<string> {
  const routeMetaAppKey = resolveRouteMetaAppKey(to)
  if (routeMetaAppKey) {
    return routeMetaAppKey
  }

  const inferredPathAppKey = inferRuntimeAppKeyByPath(to.path, '')
  if (inferredPathAppKey) {
    return inferredPathAppKey
  }

  const managedAppKey = `${appContextStore.effectiveManagedAppKey || ''}`.trim()
  if (managedAppKey) {
    return managedAppKey
  }

  const runtimeAppKey = `${appContextStore.currentRuntimeAppKey || ''}`.trim()
  if (runtimeAppKey) {
    return runtimeAppKey
  }

  return inferRuntimeAppKeyByPath(to.path, runtimeAppKey)
}

async function resolveCentralizedLoginPolicy(
  targetAppKey: string,
  appContextStore: ReturnType<typeof useAppContextStore>
): Promise<boolean> {
  const normalizedAppKey = `${targetAppKey || ''}`.trim()
  if (!normalizedAppKey) {
    return false
  }

  if (
    !appContextStore.resolveAppAuthMode(normalizedAppKey) &&
    Object.keys(appContextStore.resolveAppCapabilities(normalizedAppKey)).length === 0
  ) {
    try {
      const response = await fetchGetCurrentApp(normalizedAppKey)
      appContextStore.setAppProfile({
        appKey: response?.app?.appKey || normalizedAppKey,
        authMode: response?.app?.authMode || '',
        capabilities: response?.app?.capabilities || {},
        meta: response?.app?.meta || {}
      })
    } catch (error) {
      console.warn('[RouteGuard] 预热 app 认证策略失败，回退本地上下文', error)
    }
  }

  if (appContextStore.shouldUseCentralizedLoginForApp(normalizedAppKey)) {
    return true
  }

  const authMode = `${appContextStore.resolveAppAuthMode(normalizedAppKey) || ''}`.trim()
  const capabilities = appContextStore.resolveAppCapabilities(normalizedAppKey)
  const authConfig =
    capabilities && typeof capabilities.auth === 'object' && !Array.isArray(capabilities.auth)
      ? capabilities.auth
      : null

  if (normalizedAppKey === 'account-portal') {
    return false
  }

  if (!authMode || authMode === 'inherit_host') {
    if (!authConfig || Object.keys(authConfig).length === 0) {
      return true
    }
  }

  return false
}

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
  const userStore = useUserStore()

  // 显示 loading
  setPendingLoading(true)
  loadingService.showLoading()

  const preferredSpaceKey = resolvePreferredSpaceKeyFromRoute(to)

  // 将整个初始化流程包裹在单个 Promise 中，作为并发哨兵
  const initWork = (async () => {
    // 并行预取 userInfo 与 navigation manifest，减少首屏串行等待
    const [userData] = await Promise.all([
      fetchGetUserInfo(),
      fetchGetRuntimeNavigation(normalizeMenuSpaceKey(preferredSpaceKey)).catch(() => null)
    ])

    await userStore.restoreSession({
      preferredSpaceKey,
      prefetchedUser: userData
    })

    // 构建注册路由（内部会再次调用 fetchGetRuntimeNavigation，应命中缓存）
    return await ensureAuthenticatedRoutes(
      to,
      preferredSpaceKey,
      async () => {
        const [userData] = await Promise.all([
          fetchGetUserInfo(),
          fetchGetRuntimeNavigation(normalizeMenuSpaceKey(preferredSpaceKey)).catch(() => null)
        ])

        await userStore.restoreSession({
          preferredSpaceKey,
          prefetchedUser: userData
        })
      },
      router
    )
  })()

  setRouteInitInProgress(initWork)

  try {
    const landingPath = await initWork

    if (isStaticRoute(to.path)) {
      next({
        path: to.path,
        query: to.query,
        hash: to.hash,
        replace: true
      })
      return
    }

    if (to.path === '/' && landingPath !== '/') {
      next({
        path: landingPath,
        replace: true
      })
      return
    }

    next({
      path: to.path,
      query: to.query,
      hash: to.hash,
      replace: true
    })
  } catch (error) {
    console.error('[RouteGuard] 动态路由注册失败:', error)

    void closeLoading()

    if (isUnauthorizedError(error)) {
      next(false)
      return
    }

    setRouteInitFailed(true)

    if (isHttpError(error)) {
      console.error(`[RouteGuard] 错误码: ${error.code}, 消息: ${error.message}`)
    }

    next({ name: 'Exception500', replace: true })
  } finally {
    setRouteInitInProgress(null)
  }
}

/**
 * 处理根路径重定向到首页
 * @returns true 表示已处理跳转，false 表示无需跳转
 */
function handleRootPathRedirect(to: RouteLocationNormalized, next: NavigationGuardNext): boolean {
  if (to.path !== '/') {
    return false
  }

  const homePath = useMenuStore().getHomePath()
  if (homePath && homePath !== '/') {
    next({ path: homePath, replace: true })
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
  const appContextStore = useAppContextStore()
  if (!userStore.isLogin || isStaticRoute(to.path) || isPublicRuntimeRoute(to)) {
    return false
  }
  if (hasRouteRefreshAttempted(to.fullPath)) {
    clearRouteRefreshAttempted(to.fullPath)
    return false
  }

  addRouteRefreshAttempted(to.fullPath)
  try {
    const inferredAppKey = inferRuntimeAppKeyByPath(to.path, appContextStore.currentRuntimeAppKey)
    if (inferredAppKey && inferredAppKey !== appContextStore.effectiveManagedAppKey) {
      appContextStore.setManagedAppKey(inferredAppKey)
    }
    // 仅刷新菜单，不再重新拉取用户信息（避免双重 user info fetch）
    await refreshUserMenus()
    // 刷新访问图后若目标路由仍不存在，说明它已经不在当前运行时导航清单里，
    // 常见于菜单被禁用、空间切换后失效或权限被收回，此时直接落到 404，
    // 避免再次强跳原路径造成一串 "No match found" 警告。
    if (!hasDynamicRoute(router, to.path)) {
      clearRouteRefreshAttempted(to.fullPath)
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
    clearRouteRefreshAttempted(to.fullPath)
    return false
  }
}

function buildCrossDomainAppRedirectTarget(to: RouteLocationNormalized): string {
  if (typeof window === 'undefined') {
    return ''
  }
  const appContextStore = useAppContextStore()
  const entry = `${appContextStore.currentRuntimeFrontendEntryURL || ''}`.trim()
  if (!/^https?:\/\//i.test(entry)) {
    return ''
  }
  try {
    const target = new URL(entry)
    if (target.origin === window.location.origin) {
      return ''
    }
    if (!target.searchParams.has('redirect')) {
      target.searchParams.set('redirect', to.fullPath || '/')
    }
    return target.toString()
  } catch {
    return ''
  }
}
