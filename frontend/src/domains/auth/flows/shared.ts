import { ElNotification } from 'element-plus'
import { useAppContextStore } from '@/domains/app-runtime/context'
import { useMenuStore } from '@/domains/navigation/menu'
import { useWorktabStore } from '@/domains/navigation/worktab'
import AppConfig from '@/config'
import { useUserStore } from '@/domains/auth/store'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { resetRouterState, resetRouterStateNow } from '@/domains/navigation/runtime/reset-handlers'
import { RoutesAlias } from '@/router/routesAlias'
import { refreshUserMenus as refreshRuntimeUserMenus } from '@/domains/navigation/runtime/navigation'
import { ACTIVE_APP_SCOPE_STORAGE_KEY } from '@/domains/app-runtime/app-scope'
import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { logger } from '@/utils/logger'

export const LOGIN_REMEMBER_KEY = 'gg-login-remember'

export interface LoginFormState {
  username: string
  password: string
  rememberPassword: boolean
}

const RUNTIME_STORE_IDS = [
  'userStore',
  'appContextStore',
  'menuSpaceStore',
  'workspaceStore',
  'collaborationWorkspaceStore',
  'menuStore',
  'worktabStore'
]

function shouldRemoveRuntimeStorageKey(key: string): boolean {
  const normalizedKey = `${key || ''}`.trim()
  if (!normalizedKey) return false
  if (
    normalizedKey === ACTIVE_APP_SCOPE_STORAGE_KEY ||
    normalizedKey === 'gge:session-sync' ||
    normalizedKey === 'user' ||
    normalizedKey === 'appContextStore'
  ) {
    return true
  }

  return RUNTIME_STORE_IDS.some(
    (storeId) =>
      normalizedKey === storeId ||
      normalizedKey.endsWith(`-${storeId}`) ||
      normalizedKey.endsWith(`:${storeId}`)
  )
}

function resetPersistedRuntimeState(): void {
  if (typeof window === 'undefined') return

  for (const key of Object.keys(window.localStorage)) {
    if (shouldRemoveRuntimeStorageKey(key)) {
      window.localStorage.removeItem(key)
    }
  }

  for (const key of Object.keys(window.sessionStorage)) {
    if (shouldRemoveRuntimeStorageKey(key) || key === 'iframeRoutes') {
      window.sessionStorage.removeItem(key)
    }
  }
}

function resetInMemoryRuntimeState(): void {
  useMenuStore().setMenuList([])
  useMenuStore().setHomePath('')
  useWorktabStore().clearAll()
  useWorkspaceStore().clearWorkspaceContext()
  useCollaborationWorkspaceStore().clearCollaborationWorkspaceContext()
  useMenuSpaceStore().clearActiveSpaceKey()
  useAppContextStore().clearAppContext()
}

function rebuildClientRuntimeStateForLogin(): void {
  const userStore = useUserStore()
  userStore.clearSessionState({ broadcast: false, resetRouterDelay: null })
  resetPersistedRuntimeState()
  resetInMemoryRuntimeState()
}

export function loadRememberedCredentials(target: LoginFormState): void {
  try {
    const raw = localStorage.getItem(LOGIN_REMEMBER_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw) as Partial<LoginFormState>
    target.username = parsed.username || ''
    target.password = parsed.password || ''
    target.rememberPassword = !!parsed.rememberPassword
  } catch (error) {
    logger.warn('auth.remember_load_failed', { err: error })
    localStorage.removeItem(LOGIN_REMEMBER_KEY)
  }
}

export function persistRememberedCredentials(target: LoginFormState): void {
  if (target.rememberPassword) {
    localStorage.setItem(
      LOGIN_REMEMBER_KEY,
      JSON.stringify({
        username: target.username,
        password: target.password,
        rememberPassword: true
      })
    )
    return
  }
  localStorage.removeItem(LOGIN_REMEMBER_KEY)
}

export function normalizeRedirect(raw?: string): string {
  const cleanPath = `${raw || ''}`.trim()
  if (!cleanPath) return '/'

  let current = cleanPath
  const decodeOnce = (value: string) => {
    try {
      return decodeURIComponent(value)
    } catch {
      return value
    }
  }

  let safeIterations = 0
  while (current.includes('redirect=') && safeIterations < 5) {
    const redirectIndex = current.indexOf('redirect=')
    current = decodeOnce(current.slice(redirectIndex + 'redirect='.length))
    safeIterations += 1
  }

  const normalized = decodeOnce(current).trim()
  if (normalized.startsWith('#/')) return normalized.slice(1)
  if (normalized.startsWith('/#/')) return normalized.slice(2)
  if (!normalized || !normalized.startsWith('/')) return '/'
  if (normalized.startsWith('/auth/login') || normalized.startsWith(RoutesAlias.Login)) {
    return '/'
  }
  return normalized
}

export interface PostAuthLanding {
  url?: string
  app_key?: string
  navigation_space_key?: string
  home_path?: string
}

/**
 * 从 URL query 参数中恢复 pending register 保留的 landing 信息。
 * query 参数名：landing_url, landing_app_key, landing_space, landing_path
 */
export function buildLandingFromQuery(
  query: Record<string, string | (string | null)[] | null | undefined>
): PostAuthLanding | undefined {
  const url = `${query.landing_url || ''}`.trim()
  const appKey = `${query.landing_app_key || ''}`.trim()
  const space = `${query.landing_space || ''}`.trim()
  const path = `${query.landing_path || ''}`.trim()
  if (!url && !appKey && !path) return undefined
  return {
    ...(url ? { url } : {}),
    ...(appKey ? { app_key: appKey } : {}),
    ...(space ? { navigation_space_key: space } : {}),
    ...(path ? { home_path: path } : {})
  }
}

function isSafeRedirectURL(url: string): boolean {
  const trimmed = url.trim()
  if (!trimmed) return false
  if (trimmed.startsWith('/')) return true
  const lower = trimmed.toLowerCase()
  return lower.startsWith('http://') || lower.startsWith('https://')
}

/**
 * 统一认证后跳转方法。
 * 优先级：landing.url → landing.home_path + navigation_space_key → fallbackRedirect → '/'
 */
export async function gotoAfterAuth(
  landing: PostAuthLanding | null | undefined,
  router: ReturnType<typeof useRouter>,
  fallbackRedirect?: string
): Promise<void> {
  if (landing?.url && isSafeRedirectURL(landing.url)) {
    window.location.assign(landing.url)
    return
  }
  const landingPath = landing?.home_path || fallbackRedirect || '/'
  const navigationSpaceKey = landing?.navigation_space_key || ''
  await gotoAfterLogin(landingPath, router, navigationSpaceKey)
}

export async function gotoAfterLogin(
  landingPath: string,
  router: ReturnType<typeof useRouter>,
  navigationSpaceKey = ''
): Promise<void> {
  const menuSpaceStore = useMenuSpaceStore()
  const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(landingPath, navigationSpaceKey)
  if (nextTarget.mode === 'location') {
    window.location.assign(nextTarget.target)
    return
  }

  const fallbackUrl = new URL(router.resolve(landingPath).href, window.location.origin).toString()
  const hasJumpedOut = () => router.currentRoute.value.path !== RoutesAlias.Login

  try {
    await router.replace(landingPath)
    await nextTick()
    if (!hasJumpedOut()) {
      setTimeout(() => {
        if (!hasJumpedOut()) {
          window.location.assign(fallbackUrl)
        }
      }, 900)
    }
  } catch (error) {
    logger.warn('auth.post_login_nav_failed', { err: error, fallbackUrl })
    window.location.assign(fallbackUrl)
  }
}

export async function initializeLoginContext(user?: Api.Auth.UserInfo): Promise<void> {
  const userStore = useUserStore()

  try {
    await userStore.restoreSession({
      prefetchedUser: user,
      forceRefresh: true
    })
  } catch (error) {
    logger.warn('auth.init_context_failed', { err: error })
  }
}

export async function finalizeAuthenticatedSession(
  response: {
    access_token?: string | null
    refresh_token?: string | null
    user?: Api.Auth.UserInfo | null
  },
  options: {
    refreshUserContext?: boolean
    refreshMenus?: boolean
  } = {}
): Promise<void> {
  const userStore = useUserStore()
  if (!response.access_token) {
    throw new Error('登录失败，未收到访问令牌')
  }

  rebuildClientRuntimeStateForLogin()
  resetRouterStateNow()
  userStore.applySession({
    accessToken: response.access_token,
    refreshToken: response.refresh_token || undefined,
    isLogin: true
  })

  if (options.refreshUserContext !== false) {
    await userStore.restoreSession({
      prefetchedUser: response.user || undefined,
      forceRefresh: true
    })
  } else if (response.user) {
    await initializeLoginContext(response.user)
  }

  if (options.refreshMenus) {
    await refreshRuntimeUserMenus()
  }
}

export function showLoginSuccessNotice(displayName: string, t: (key: string) => string): void {
  const systemName = AppConfig.systemInfo.name
  setTimeout(() => {
    ElNotification({
      title: t('login.success.title'),
      type: 'success',
      duration: 2500,
      zIndex: 10000,
      message: `${t('login.success.message')}, ${displayName || systemName}!`
    })
  }, 1000)
}
