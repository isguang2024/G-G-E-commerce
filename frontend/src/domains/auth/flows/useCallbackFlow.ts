import { ref } from 'vue'
import { fetchExchangeAuthCallback } from '@/domains/auth/api'
import {
  consumeCentralizedAuthAttempt,
  resolveCentralizedTargetPath
} from '@/domains/auth/centralized-login'
import { finalizeAuthenticatedSession, gotoAfterLogin } from './shared'

export function useCallbackFlow() {
  const router = useRouter()
  const route = useRoute()
  const message = ref('正在校验回调参数并交换登录令牌...')

  async function run(): Promise<void> {
    const code = `${route.query.code || ''}`.trim()
    const state = `${route.query.state || ''}`.trim()
    const targetAppKey = `${route.query.target_app_key || ''}`.trim()
    const redirectUri = `${route.query.redirect_uri || window.location.href}`.trim()
    if (!code || !state || !targetAppKey || !redirectUri) {
      throw new Error('缺少 callback 参数')
    }

    const attempt = consumeCentralizedAuthAttempt(state)
    if (!attempt?.nonce) {
      throw new Error('登录回调上下文已失效，请重新登录')
    }

    const result = await fetchExchangeAuthCallback({
      code,
      state,
      nonce: attempt.nonce,
      target_app_key: targetAppKey,
      redirect_uri: redirectUri
    })

    await finalizeAuthenticatedSession(result, {
      refreshUserContext: true,
      refreshMenus: true
    })

    const landingPath = resolveCentralizedTargetPath(
      result.landing?.home_path,
      attempt.targetPath || `${route.query.target_path || ''}`.trim()
    )
    await gotoAfterLogin(
      landingPath,
      router,
      result.landing?.navigation_space_key || attempt.navigationSpaceKey || ''
    )
  }

  return {
    message,
    run
  }
}
