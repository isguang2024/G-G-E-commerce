import { ElNotification } from 'element-plus'
import AppConfig from '@/config'
import { useUserStore } from '@/domains/auth/store'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { resetRouterState } from '@/domains/navigation/runtime/reset-handlers'
import { RoutesAlias } from '@/router/routesAlias'
import { refreshUserMenus as refreshRuntimeUserMenus } from '@/domains/navigation/runtime/navigation'

export const LOGIN_REMEMBER_KEY = 'gg-login-remember'

export interface LoginFormState {
  username: string
  password: string
  rememberPassword: boolean
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
    console.warn('[Login] 读取记住密码失败，已忽略:', error)
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
    console.warn('[AuthFlow] 登录导航失败，尝试兜底跳转:', error)
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
    console.warn('[AuthFlow] 登录初始化上下文失败，仍允许进入应用:', error)
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

  resetRouterState(0)
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
