import { AxiosHeaders, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { AUTH_PROTOCOL_VERSION } from '@/domains/auth/centralized-login'
import { ErrorCodes } from '@/api/v5/error-codes'
import { HttpError, showError } from './error'
import { ApiStatus } from './status'
import {
  applyHttpSession,
  getCurrentRuntimeAppKey,
  getHttpAccessToken,
  getEffectiveManagedAppKey,
  getHttpRefreshToken,
  isHttpAppFeatureEnabled,
  logoutHttpSession,
  resolveHttpAppAuthMode
} from './request-context'

export const SKIP_WORKSPACE_CONTEXT_HEADER = 'X-Skip-Workspace-Context'
export const AUTH_RETRY_HEADER = 'X-Auth-Retry'

const UNAUTHORIZED_RESET_DELAY = 3000
const AUTH_FLOW_ENDPOINTS = [
  '/auth/login',
  '/auth/logout',
  '/auth/refresh',
  '/auth/register',
  '/auth/register-context',
  '/auth/callback/exchange'
]

let unauthorizedHandling: Promise<void> | null = null
let refreshPromise: Promise<boolean> | null = null

function normalizeRequestPath(url?: string): string {
  const raw = `${url || ''}`.trim()
  if (!raw) return ''
  try {
    return new URL(raw, window.location.origin).pathname
  } catch {
    return raw
  }
}

export function shouldBypassUnauthorizedLogout(url?: string): boolean {
  const path = normalizeRequestPath(url)
  return AUTH_FLOW_ENDPOINTS.some((endpoint) => path.endsWith(endpoint))
}

export function isUnauthorizedBusinessCode(code?: number): boolean {
  return new Set<number>([
    ApiStatus.unauthorized,
    ErrorCodes.Unauthorized,
    ErrorCodes.TokenExpired
  ]).has(Number(code || 0))
}

export function shouldUseSharedSessionMode(): boolean {
  const appKey = getEffectiveManagedAppKey() || getCurrentRuntimeAppKey()
  const authMode = resolveHttpAppAuthMode(appKey)
  if (authMode === 'shared_cookie') {
    return true
  }
  return isHttpAppFeatureEnabled(appKey, 'shared_cookie')
}

export function shouldSkipRefreshAttempt(request: Request): boolean {
  const pathname = new URL(request.url, window.location.origin).pathname
  if (request.headers.get(AUTH_RETRY_HEADER) === '1') {
    return true
  }
  return AUTH_FLOW_ENDPOINTS.some((endpoint) => pathname.endsWith(endpoint))
}

export function shouldSkipAxiosRefreshAttempt(
  request: Pick<InternalAxiosRequestConfig, 'url' | 'headers'>
): boolean {
  const retryHeader =
    typeof request.headers?.get === 'function'
      ? request.headers.get(AUTH_RETRY_HEADER)
      : (request.headers as Record<string, string | undefined>)?.[AUTH_RETRY_HEADER]
  if (retryHeader === '1') {
    return true
  }
  return AUTH_FLOW_ENDPOINTS.some((endpoint) =>
    normalizeRequestPath(request.url).endsWith(endpoint)
  )
}

export async function refreshSessionIfNeeded(): Promise<boolean> {
  if (refreshPromise) {
    return refreshPromise
  }
  refreshPromise = (async () => {
    const rawRefreshToken = getHttpRefreshToken()
    if (!rawRefreshToken) {
      return false
    }
    const response = await fetch('/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        [SKIP_WORKSPACE_CONTEXT_HEADER]: 'true'
      },
      body: JSON.stringify({
        refresh_token: rawRefreshToken,
        client_app_key: getEffectiveManagedAppKey() || getCurrentRuntimeAppKey(),
        auth_protocol_version: AUTH_PROTOCOL_VERSION
      })
    }).catch(() => null)
    if (!response?.ok) {
      return false
    }
    const payload = await response.json().catch(() => null)
    const accessToken = `${payload?.access_token || ''}`.trim()
    const refreshToken = `${payload?.refresh_token || ''}`.trim()
    if (!accessToken || !refreshToken) {
      return false
    }
    applyHttpSession({
      accessToken,
      refreshToken,
      isLogin: true
    })
    return true
  })().finally(() => {
    refreshPromise = null
  })
  return refreshPromise
}

export function triggerUnauthorizedLogout(error: HttpError): void {
  if (!unauthorizedHandling) {
    unauthorizedHandling = (async () => {
      try {
        showError(error, true)
        await logoutHttpSession()
      } finally {
        setTimeout(() => {
          unauthorizedHandling = null
        }, UNAUTHORIZED_RESET_DELAY)
      }
    })()
  }
}

export async function retryAxiosRequestWithRefresh<T = unknown>(
  response: AxiosResponse<T>,
  request: InternalAxiosRequestConfig,
  resend: (config: InternalAxiosRequestConfig) => Promise<AxiosResponse<T>>
): Promise<AxiosResponse<T> | null> {
  void response
  if (
    !shouldUseSharedSessionMode() ||
    shouldSkipAxiosRefreshAttempt(request) ||
    shouldBypassUnauthorizedLogout(request.url)
  ) {
    return null
  }
  const refreshed = await refreshSessionIfNeeded()
  if (!refreshed) {
    return null
  }
  const accessToken = getHttpAccessToken()
  if (!accessToken) {
    return null
  }
  const headers = AxiosHeaders.from(request.headers)
  headers.set(
    'Authorization',
    accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`
  )
  headers.set(AUTH_RETRY_HEADER, '1')
  return resend({
    ...request,
    headers
  })
}
