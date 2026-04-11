const AUTH_ATTEMPT_PREFIX = 'gge-auth-attempt:'
export const AUTH_PROTOCOL_VERSION = 'callback-v1'

export interface CentralizedAuthAttempt {
  state: string
  nonce: string
  targetAppKey: string
  targetPath: string
  redirectUri: string
  navigationSpaceKey?: string
}

function randomToken(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID().replace(/-/g, '')
  }
  return `${Date.now()}${Math.random().toString(16).slice(2)}`
}

export function createCentralizedAuthAttempt(
  targetAppKey: string,
  targetPath: string,
  redirectUri: string,
  navigationSpaceKey = ''
): CentralizedAuthAttempt {
  return {
    state: randomToken(),
    nonce: randomToken(),
    targetAppKey: `${targetAppKey || ''}`.trim(),
    targetPath: `${targetPath || ''}`.trim() || '/',
    redirectUri: `${redirectUri || ''}`.trim(),
    navigationSpaceKey: `${navigationSpaceKey || ''}`.trim()
  }
}

export function persistCentralizedAuthAttempt(attempt: CentralizedAuthAttempt): void {
  if (typeof window === 'undefined') return
  sessionStorage.setItem(`${AUTH_ATTEMPT_PREFIX}${attempt.state}`, JSON.stringify(attempt))
}

export function consumeCentralizedAuthAttempt(state: string): CentralizedAuthAttempt | null {
  if (typeof window === 'undefined') return null
  const key = `${AUTH_ATTEMPT_PREFIX}${`${state || ''}`.trim()}`
  const raw = sessionStorage.getItem(key)
  sessionStorage.removeItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as CentralizedAuthAttempt
  } catch {
    return null
  }
}

export function buildCentralizedLoginURL(input: {
  loginHost?: string
  targetAppKey: string
  targetPath: string
  redirectUri: string
  navigationSpaceKey?: string
  state: string
  nonce: string
}): string {
  const url = new URL('/account/auth/login', window.location.origin)
  const loginHost = `${input.loginHost || ''}`.trim()
  if (loginHost) {
    url.protocol = window.location.protocol
    url.host = loginHost
  }
  url.searchParams.set('target_app_key', input.targetAppKey)
  url.searchParams.set('redirect_uri', input.redirectUri)
  url.searchParams.set('target_path', input.targetPath)
  url.searchParams.set('state', input.state)
  url.searchParams.set('nonce', input.nonce)
  url.searchParams.set('auth_protocol_version', AUTH_PROTOCOL_VERSION)
  if (`${input.navigationSpaceKey || ''}`.trim()) {
    url.searchParams.set('navigation_space_key', `${input.navigationSpaceKey || ''}`.trim())
  }
  return url.toString()
}

export function resolveCentralizedTargetPath(landingPath?: string, targetPath?: string): string {
  const preferred = `${targetPath || ''}`.trim()
  if (preferred.startsWith('/')) return preferred
  const fallback = `${landingPath || ''}`.trim()
  if (fallback.startsWith('/')) return fallback
  return '/'
}
