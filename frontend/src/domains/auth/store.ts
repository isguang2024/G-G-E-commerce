/**
 * 用户状态管理模块
 *
 * 提供用户相关的状态管理
 *
 * ## 主要功能
 *
 * - 用户登录状态管理
 * - 用户信息存储
 * - 访问令牌和刷新令牌管理
 * - 语言设置
 * - 搜索历史记录
 * - 锁屏状态和密码管理
 * - 登出清理逻辑
 *
 * ## 使用场景
 *
 * - 用户登录和认证
 * - 权限验证
 * - 个人信息展示
 * - 多语言切换
 * - 锁屏功能
 * - 搜索历史管理
 *
 * ## 持久化
 *
 * - 使用 localStorage 存储
 * - 存储键：sys-v{version}-user
 * - 登出时自动清理
 *
 * @module domains/auth/store
 * @author Art Design Pro Team
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  buildCentralizedLoginURL,
  createCentralizedAuthAttempt,
  persistCentralizedAuthAttempt
} from '@/domains/auth/centralized-login'
import { useAppContextStore } from '@/domains/app-runtime/context'
import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { restoreSessionViaHandler } from '@/domains/auth/runtime/session-handlers'
import { useMenuStore } from '@/domains/navigation/menu'
import { resetRouterState } from '@/domains/navigation/runtime/reset-handlers'
import { useWorktabStore } from '@/domains/navigation/worktab'
import { getNavigationRouter } from '@/domains/navigation/runtime/router-instance'
import { LanguageEnum } from '@/enums/appEnum'
import { AppRouteRecord } from '@/types/router'
import { registerHttpAuthContext } from '@/utils/http/request-context'
import { logger } from '@/utils/logger'
import { setPageTitle } from '@/utils/router'
import { StorageConfig } from '@/utils/storage/storage-config'
import { RoutesAlias } from '@/router/routesAlias'
import { useCollaborationStore } from '@/store/modules/collaboration'
import { useSettingStore } from '@/store/modules/setting'

const SESSION_SYNC_EVENT_KEY = 'gge:session-sync'

interface RestoreSessionOptions {
  preferredSpaceKey?: string
  prefetchedUser?: Api.Auth.UserInfo
  forceRefresh?: boolean
  skipWorkspaceReconcile?: boolean
}

export const useUserStore = defineStore(
  'userStore',
  () => {
    const language = ref(LanguageEnum.ZH)
    const isLogin = ref(false)
    const isLock = ref(false)
    const lockPassword = ref('')
    const info = ref<Partial<Api.Auth.UserInfo>>({})
    const emptyUserInfo: Partial<Api.Auth.UserInfo> = { actions: [], roles: [] }
    const searchHistory = ref<AppRouteRecord[]>([])
    const accessToken = ref('')
    const refreshToken = ref('')

    const broadcastSessionEvent = (payload: Record<string, any>) => {
      if (typeof window === 'undefined') return
      localStorage.setItem(
        SESSION_SYNC_EVENT_KEY,
        JSON.stringify({
          ...payload,
          emittedAt: Date.now()
        })
      )
    }

    const getUserInfo = computed(() => info.value)
    const getSettingState = computed(() => useSettingStore().$state)
    const getWorktabState = computed(() => useWorktabStore().$state)

    const setUserInfo = (newInfo: Api.Auth.UserInfo) => {
      info.value = newInfo
      // 把 userId 同步给 logger —— 后续上报的日志都带用户身份，
      // 方便运维侧按人聚合异常。
      logger.setUser(`${newInfo?.userId || newInfo?.id || ''}`)
    }

    const setLoginStatus = (status: boolean) => {
      isLogin.value = status
    }

    const setLanguage = (lang: LanguageEnum) => {
      setPageTitle(getNavigationRouter().currentRoute.value)
      language.value = lang
    }

    const setSearchHistory = (list: AppRouteRecord[]) => {
      searchHistory.value = list
    }

    const setLockStatus = (status: boolean) => {
      isLock.value = status
    }

    const setLockPassword = (password: string) => {
      lockPassword.value = password
    }

    const setToken = (newAccessToken: string, newRefreshToken?: string) => {
      accessToken.value = newAccessToken
      if (newRefreshToken) {
        refreshToken.value = newRefreshToken
      }
    }

    registerHttpAuthContext({
      getAccessToken: () => accessToken.value,
      getRefreshToken: () => refreshToken.value,
      applySession: (payload) => applySession(payload),
      logOut: () => logOut()
    })

    const applySession = (
      payload: {
        accessToken: string
        refreshToken?: string
        isLogin?: boolean
      },
      options: { broadcast?: boolean } = {}
    ) => {
      setToken(payload.accessToken, payload.refreshToken)
      setLoginStatus(payload.isLogin ?? true)
      if (options.broadcast !== false) {
        broadcastSessionEvent({
          type: 'session:update',
          accessToken: payload.accessToken,
          refreshToken: payload.refreshToken,
          isLogin: payload.isLogin ?? true
        })
      }
    }

    const resolveCurrentUserId = (): string => {
      const userId = `${info.value.userId || info.value.id || ''}`.trim()
      return userId
    }

    const syncLoginUserIdentity = (nextUserId?: string): void => {
      const normalizedNextUserId = `${nextUserId || ''}`.trim()
      if (!normalizedNextUserId) return

      const currentUserId = resolveCurrentUserId()
      const lastUserId = `${localStorage.getItem(StorageConfig.LAST_USER_ID_KEY) || ''}`.trim()
      const shouldClearWorktabs =
        (currentUserId && currentUserId !== normalizedNextUserId) ||
        (lastUserId && lastUserId !== normalizedNextUserId)

      if (shouldClearWorktabs) {
        useWorktabStore().clearAll()
      }

      localStorage.removeItem(StorageConfig.LAST_USER_ID_KEY)
    }

    const clearSessionState = (
      options: {
        preserveLastUserId?: boolean
        broadcast?: boolean
        resetRouterDelay?: number | null
      } = {}
    ) => {
      if (options.preserveLastUserId) {
        const currentUserId = info.value.userId
        if (currentUserId) {
          localStorage.setItem(StorageConfig.LAST_USER_ID_KEY, String(currentUserId))
        }
      }
      info.value = { ...emptyUserInfo }
      isLogin.value = false
      isLock.value = false
      lockPassword.value = ''
      accessToken.value = ''
      refreshToken.value = ''
      // 登出/清空会话后 logger 也要断开 userId，避免后续日志错误归属。
      logger.setUser('')
      sessionStorage.removeItem('iframeRoutes')
      useMenuStore().setHomePath('')
      useCollaborationStore().clearCollaborationContext()
      useAppContextStore().clearAppContext()
      if (options.resetRouterDelay !== null) {
        resetRouterState(options.resetRouterDelay ?? 500)
      }
      if (options.broadcast !== false) {
        broadcastSessionEvent({ type: 'session:clear' })
      }
    }

    const logOut = async () => {
      const router = getNavigationRouter()
      const menuSpaceStore = useMenuSpaceStore()
      const appContextStore = useAppContextStore()
      const currentRoute = router.currentRoute.value
      const redirectTarget = currentRoute.path !== RoutesAlias.Login ? currentRoute.fullPath : '/'
      const targetAppKey =
        appContextStore.effectiveManagedAppKey || appContextStore.currentRuntimeAppKey
      const preferredSpaceKey = menuSpaceStore.currentSpaceKey
      const binding = menuSpaceStore.resolveHostBinding(preferredSpaceKey)

      if (accessToken.value || refreshToken.value || isLogin.value) {
        try {
          const { fetchLogout } = await import('@/domains/auth/api')
          await fetchLogout()
        } catch {
          // stateless JWT 登出允许前端本地兜底
        }
      }
      clearSessionState({ preserveLastUserId: true })

      if (
        typeof window !== 'undefined' &&
        targetAppKey &&
        appContextStore.shouldUseCentralizedLoginForApp(targetAppKey)
      ) {
        const redirectUri = new URL(RoutesAlias.AuthCallback, window.location.origin).toString()
        const loginUiMode = appContextStore.resolveAppLoginUiMode(targetAppKey)
        const loginPageKey =
          loginUiMode === 'auth_center_custom'
            ? appContextStore.resolveAppLoginPageKey(targetAppKey)
            : ''
        const attempt = createCentralizedAuthAttempt(
          targetAppKey,
          redirectTarget,
          redirectUri,
          preferredSpaceKey || binding?.menuSpaceKey || '',
          loginPageKey
        )
        persistCentralizedAuthAttempt(attempt)
        window.location.assign(
          buildCentralizedLoginURL({
            loginHost: binding?.loginHost || '',
            targetAppKey,
            targetPath: redirectTarget,
            redirectUri,
            navigationSpaceKey: preferredSpaceKey || binding?.menuSpaceKey || '',
            state: attempt.state,
            nonce: attempt.nonce,
            loginPageKey: attempt.loginPageKey
          })
        )
        return
      }

      void router.push({
        path: RoutesAlias.Login,
        query: redirectTarget && redirectTarget !== '/' ? { redirect: redirectTarget } : undefined
      })
    }

    const restoreSession = async (options: RestoreSessionOptions = {}) => {
      return restoreSessionViaHandler(options)
    }

    // @compat-status: keep 当前仍需依赖 last-user-id 与 worktab 清理逻辑，避免跨账号残留标签页污染。
    const checkAndClearWorktabs = () => {
      const lastUserId = localStorage.getItem(StorageConfig.LAST_USER_ID_KEY)
      const currentUserId = resolveCurrentUserId()

      if (!currentUserId) return
      if (!lastUserId) return

      if (String(currentUserId) !== lastUserId) {
        useWorktabStore().clearAll()
      }

      localStorage.removeItem(StorageConfig.LAST_USER_ID_KEY)
    }

    if (typeof window !== 'undefined') {
      window.addEventListener('storage', (event) => {
        if (event.key !== SESSION_SYNC_EVENT_KEY || !event.newValue) return
        try {
          const payload = JSON.parse(event.newValue) as {
            type?: string
            accessToken?: string
            refreshToken?: string
            isLogin?: boolean
          }
          if (payload.type === 'session:update' && payload.accessToken) {
            applySession(
              {
                accessToken: payload.accessToken,
                refreshToken: payload.refreshToken,
                isLogin: payload.isLogin ?? true
              },
              { broadcast: false }
            )
            return
          }
          if (payload.type === 'session:clear') {
            clearSessionState({ broadcast: false })
          }
        } catch {
          // ignore malformed sync payload
        }
      })
    }

    return {
      language,
      isLogin,
      isLock,
      lockPassword,
      info,
      searchHistory,
      accessToken,
      refreshToken,
      getUserInfo,
      getSettingState,
      getWorktabState,
      setUserInfo,
      setLoginStatus,
      setLanguage,
      setSearchHistory,
      setLockStatus,
      setLockPassword,
      setToken,
      applySession,
      syncLoginUserIdentity,
      clearSessionState,
      logOut,
      restoreSession,
      checkAndClearWorktabs
    }
  },
  {
    persist: {
      key: 'user',
      storage: localStorage
    }
  }
)
