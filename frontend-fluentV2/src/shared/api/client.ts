import axios from 'axios'
import type { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import { ApiError, normalizeApiError } from '@/shared/api/errors'
import { appConfig } from '@/shared/config/app-config'
import type { ApiEnvelope } from '@/shared/types/api'
import type { AuthSession } from '@/shared/types/auth'
import { readStoredAuthSnapshot } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'

let unauthorizedHandler: ((error: unknown) => void) | null = null
let refreshSessionPromise: Promise<AuthSession | null> | null = null
export type RequestTenantMode = 'current' | 'none'

export interface RequestDataConfig extends AxiosRequestConfig {
  tenantMode?: RequestTenantMode
}

interface RetryableAxiosRequestConfig extends InternalAxiosRequestConfig {
  _retry?: boolean
  _skipAuthRefresh?: boolean
  tenantMode?: RequestTenantMode
}

export function setUnauthorizedHandler(handler: ((error: unknown) => void) | null) {
  unauthorizedHandler = handler
}

function resolveTenantId() {
  return useAuthStore.getState().tenantContext.currentTenantId
}

export const axiosInstance = axios.create({
  baseURL: appConfig.apiBaseUrl,
  timeout: 15_000,
})

const authRefreshClient = axios.create({
  baseURL: appConfig.apiBaseUrl,
  timeout: 15_000,
})

function isAuthLifecyclePath(url?: string) {
  const target = `${url || ''}`.trim()
  return target.includes('/api/v1/auth/login') || target.includes('/api/v1/auth/refresh')
}

async function performTokenRefresh() {
  const snapshot = readStoredAuthSnapshot()
  if (!snapshot?.session.refreshToken) {
    return null
  }

  const response = await authRefreshClient.request<ApiEnvelope<{
    access_token?: string
    refresh_token?: string
    expires_in?: number
  }>>({
    method: 'POST',
    url: '/api/v1/auth/refresh',
    data: {
      refresh_token: snapshot.session.refreshToken,
    },
  })

  const payload = response.data
  if (typeof payload?.code === 'number' && payload.code !== 0) {
    throw new ApiError({
      message: payload.message || '刷新会话失败',
      status: response.status,
      code: `${payload.code}`,
      businessCode: payload.code,
      details: payload.data,
    })
  }

  const nextSession: AuthSession = {
    accessToken: `${payload.data?.access_token || ''}`.trim(),
    refreshToken: `${payload.data?.refresh_token || snapshot.session.refreshToken}`.trim(),
    expiresIn: Number(payload.data?.expires_in || snapshot.session.expiresIn || 0),
    issuedAt: Date.now(),
  }

  if (!nextSession.accessToken) {
    throw new ApiError({
      message: '刷新访问令牌失败',
      status: response.status,
    })
  }

  useAuthStore.getState().updateSession(nextSession, snapshot.rememberMe)
  return nextSession
}

async function refreshSessionSingleFlight() {
  if (!refreshSessionPromise) {
    refreshSessionPromise = performTokenRefresh().finally(() => {
      refreshSessionPromise = null
    })
  }
  return refreshSessionPromise
}

axiosInstance.interceptors.request.use((config) => {
  const requestConfig = config as RetryableAxiosRequestConfig
  const snapshot = readStoredAuthSnapshot()
  const tenantId = resolveTenantId()

  config.headers = config.headers || {}
  config.headers.Accept = 'application/json'

  if (snapshot?.session.accessToken) {
    config.headers.Authorization = `Bearer ${snapshot.session.accessToken}`
  }

  if (requestConfig.tenantMode === 'none') {
    delete config.headers['X-Tenant-ID']
  } else if (tenantId) {
    config.headers['X-Tenant-ID'] = tenantId
  }

  return config
})

axiosInstance.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    const normalizedError = normalizeApiError(error)
    const originalRequest = (axios.isAxiosError(error) ? error.config : undefined) as RetryableAxiosRequestConfig | undefined

    if (
      normalizedError.status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      !originalRequest._skipAuthRefresh &&
      !isAuthLifecyclePath(originalRequest.url)
    ) {
      originalRequest._retry = true

      return refreshSessionSingleFlight()
        .then((nextSession) => {
          if (!nextSession?.accessToken) {
            unauthorizedHandler?.(normalizedError)
            return Promise.reject(normalizedError)
          }

          originalRequest.headers = originalRequest.headers || {}
          originalRequest.headers.Authorization = `Bearer ${nextSession.accessToken}`
          return axiosInstance.request(originalRequest)
        })
        .catch((refreshError) => {
          const finalError = normalizeApiError(refreshError)
          unauthorizedHandler?.(finalError)
          return Promise.reject(finalError)
        })
    }

    if (normalizedError.status === 401) {
      unauthorizedHandler?.(normalizedError)
    }

    return Promise.reject(normalizedError)
  },
)

export async function requestData<T>(config: RequestDataConfig) {
  const response = await axiosInstance.request<ApiEnvelope<T>>(config)
  const payload = response.data

  if (typeof payload?.code === 'number' && payload.code !== 0) {
    throw new ApiError({
      message: payload.message || '请求失败',
      status: response.status,
      code: `${payload.code}`,
      businessCode: payload.code,
      details: payload.data,
    })
  }

  return payload.data
}
