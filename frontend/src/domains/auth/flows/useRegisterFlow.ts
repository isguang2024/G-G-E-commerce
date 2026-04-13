import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchRegister, fetchRegisterContext, fetchSocialTokenExchange } from '@/domains/auth/api'
import { RoutesAlias } from '@/router/routesAlias'
import { finalizeAuthenticatedSession } from './shared'

export interface RegisterFormState {
  username: string
  password: string
  confirmPassword: string
  email: string
  invitationCode: string
  captchaToken: string
  agreement: boolean
}

const REDIRECT_DELAY = 1000

export function useRegisterFlow() {
  const router = useRouter()
  const route = useRoute()
  const ctx = ref<Awaited<ReturnType<typeof fetchRegisterContext>> | null>(null)
  const socialContext = ref<Awaited<ReturnType<typeof fetchSocialTokenExchange>> | null>(null)
  const socialToken = ref('')
  const contextError = ref('')
  const loading = ref(false)
  const isPublicRegisterDisabled = computed(() => ctx.value?.allow_public_register === false)

  async function loadContext(): Promise<void> {
    try {
      ctx.value = await fetchRegisterContext(window.location.host, window.location.pathname)
      contextError.value = ''
      socialToken.value = `${route.query.social_token || ''}`.trim()
      if (socialToken.value) {
        socialContext.value = await fetchSocialTokenExchange(socialToken.value)
      } else {
        socialContext.value = null
      }
    } catch (error) {
      ctx.value = null
      contextError.value =
        '未读取到当前 URL 对应的注册入口，请先检查注册入口、策略和 account-portal 页面是否已完成配置。'
      console.warn('[Register] 读取注册上下文失败', error)
    }
  }

  async function register(formData: RegisterFormState): Promise<void> {
    loading.value = true
    try {
      if (isPublicRegisterDisabled.value) {
        ElMessage.error('当前未开启公开注册')
        return
      }

      const response = await fetchRegister({
        username: formData.username,
        password: formData.password,
        confirm_password: formData.confirmPassword,
        ...(formData.email ? { email: formData.email } : {}),
        ...(formData.invitationCode ? { invitation_code: formData.invitationCode } : {}),
        ...(formData.captchaToken ? { captcha_token: formData.captchaToken } : {}),
        agreement_version: 'v1',
        ...(socialToken.value ? { social_token: socialToken.value } : {})
      })
      ElMessage.success('注册成功')

      if (response.access_token) {
        await finalizeAuthenticatedSession(
          response as {
            access_token?: string | null
            refresh_token?: string | null
            user?: Api.Auth.UserInfo | null
          },
          {
            refreshUserContext: true,
            refreshMenus: true
          }
        )
        const homePath = response.landing?.home_path ?? '/dashboard/console'
        setTimeout(() => {
          void router.push(homePath)
        }, REDIRECT_DELAY)
        return
      }

      if (response.pending) {
        setTimeout(() => {
          void router.push({ path: RoutesAlias.Login, query: { registered: '1' } })
        }, REDIRECT_DELAY)
        return
      }

      if (response.landing?.home_path) {
        setTimeout(() => {
          void router.push(response.landing?.home_path || '/')
        }, REDIRECT_DELAY)
        return
      }

      setTimeout(() => {
        void router.push({ path: RoutesAlias.Login })
      }, REDIRECT_DELAY)
    } finally {
      loading.value = false
    }
  }

  return {
    ctx,
    contextError,
    loading,
    isPublicRegisterDisabled,
    loadContext,
    register,
    socialContext,
    socialToken
  }
}
