import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchLogin } from '@/domains/auth/api'
import { HttpError } from '@/utils/http/error'
import {
  finalizeAuthenticatedSession,
  gotoAfterLogin,
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

      await gotoAfterLogin(normalizeRedirect(route.query.redirect as string), router)
    } catch (error) {
      const message =
        error instanceof HttpError ? error.message : error instanceof Error ? error.message : ''
      if (message) {
        submitError.value = message
      } else {
        console.error('[Login] Unexpected error:', error)
      }
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    submitError,
    loadRememberedCredentials,
    submit
  }
}
