import { getHttpAccessToken } from '@/utils/http/request-context'
import { logger } from '@/utils/logger'

export interface SilentSSOParams {
  targetAppKey: string
  redirectUri: string
  state: string
  nonce: string
  maxAge?: number
  targetPath?: string
  navigationSpaceKey?: string
  loginPageKey?: string
}

export interface SilentSSOResult {
  callback?: {
    mode: string
    code: string
    state: string
    target_app_key: string
    redirect_uri: string
    redirect_to: string
    target_path?: string
    navigation_space_key?: string
    auth_protocol_version?: string
  }
}

export async function attemptSilentSSO(params: SilentSSOParams): Promise<SilentSSOResult | null> {
  const accessToken = getHttpAccessToken()
  if (!accessToken) return null

  try {
    const body: Record<string, unknown> = {
      target_app_key: params.targetAppKey,
      redirect_uri: params.redirectUri,
      state: params.state,
      nonce: params.nonce
    }
    if (params.maxAge != null) body.max_age = params.maxAge
    if (params.targetPath) body.target_path = params.targetPath
    if (params.navigationSpaceKey) body.navigation_space_key = params.navigationSpaceKey
    if (params.loginPageKey) body.login_page_key = params.loginPageKey

    const res = await fetch('/api/v1/auth/callback/silent', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`
      },
      body: JSON.stringify(body)
    })
    if (!res.ok) return null
    return await res.json()
  } catch (err) {
    logger.warn('auth.silent_sso_failed', { err })
    return null
  }
}
