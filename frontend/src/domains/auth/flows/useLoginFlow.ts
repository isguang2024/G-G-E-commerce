import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchLogin, fetchSocialTokenExchange } from '@/domains/auth/api'
import { HttpError } from '@/utils/http/error'
import { logger } from '@/utils/logger'
import {
  buildLandingFromQuery,
  finalizeAuthenticatedSession,
  gotoAfterAuth,
  loadRememberedCredentials,
  normalizeRedirect,
  persistRememberedCredentials,
  showLoginSuccessNotice,
  type LoginFormState
} from './shared'

export function useLoginFlow() {
  const router = useRouter()
  const route = useRoute()
  const { t } = useI18n()
  const loading = ref(false)
  const submitError = ref('')

  async function consumeSocialToken(): Promise<boolean> {
    const socialToken = `${route.query.social_token || ''}`.trim()
    if (!socialToken) return false
    try {
      const result = await fetchSocialTokenExchange(socialToken)
      if (result.intent === 'login' && result.access_token) {
        await finalizeAuthenticatedSession(
          {
            access_token: result.access_token,
            refresh_token: result.refresh_token
          },
          { refreshUserContext: false }
        )
        const fallbackRedirect = normalizeRedirect(route.query.redirect as string)
        await gotoAfterAuth(
          { home_path: fallbackRedirect, navigation_space_key: `${route.query.navigation_space_key || ''}`.trim() },
          router
        )
        return true
      }
      const query: Record<string, string> = { social_token: socialToken }
      if (`${route.query.login_page_key || ''}`.trim()) {
        query.login_page_key = `${route.query.login_page_key}`.trim()
      }
      if (`${route.query.source_app_key || ''}`.trim()) {
        query.source_app_key = `${route.query.source_app_key}`.trim()
      }
      await router.replace({ path: '/account/auth/register', query })
      return true
    } catch (error) {
      const message =
        error instanceof HttpError ? error.message : error instanceof Error ? error.message : ''
      submitError.value = message || '社交登录凭证已失效，请重新发起'
      return false
    }
  }

  async function submit(formData: LoginFormState): Promise<void> {
    loading.value = true
    submitError.value = ''
    try {
      const response = await fetchLogin({
        username: formData.username,
        password: formData.password,
        target_app_key: `${route.query.target_app_key || ''}`.trim() || undefined,
        redirect_uri: `${route.query.redirect_uri || ''}`.trim() || undefined,
        target_path: `${route.query.target_path || ''}`.trim() || undefined,
        navigation_space_key: `${route.query.navigation_space_key || ''}`.trim() || undefined,
        state: `${route.query.state || ''}`.trim() || undefined,
        nonce: `${route.query.nonce || ''}`.trim() || undefined,
        auth_protocol_version: `${route.query.auth_protocol_version || ''}`.trim() || undefined
      })

      if (response.callback?.redirect_to) {
        persistRememberedCredentials(formData)
        window.location.assign(response.callback.redirect_to)
        return
      }

      await finalizeAuthenticatedSession(response, { refreshUserContext: false })
      persistRememberedCredentials(formData)

      const displayName =
        response.user?.nickname || response.user?.username || response.user?.email || ''
      showLoginSuccessNotice(displayName, t)

      // 构造 landing：优先 login response landing > URL query landing_* (pending register intent) > redirect
      const landing = response.landing || buildLandingFromQuery(route.query)
      const fallbackRedirect = normalizeRedirect(route.query.redirect as string)
      await gotoAfterAuth(landing, router, fallbackRedirect)
    } catch (error) {
      const message =
        error instanceof HttpError ? error.message : error instanceof Error ? error.message : ''
      if (message) {
        submitError.value = message
      } else {
        logger.error('auth.login_unexpected_error', { err: error })
      }
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    submitError,
    loadRememberedCredentials,
    submit,
    consumeSocialToken
  }
}
