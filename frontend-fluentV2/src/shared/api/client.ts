import axios from 'axios'
import type { AxiosRequestConfig } from 'axios'
import { ApiError, normalizeApiError } from '@/shared/api/errors'
import { appConfig } from '@/shared/config/app-config'
import type { ApiEnvelope } from '@/shared/types/api'
import { readStoredAuthSnapshot } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'

let unauthorizedHandler: ((error: unknown) => void) | null = null

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

axiosInstance.interceptors.request.use((config) => {
  const snapshot = readStoredAuthSnapshot()
  const tenantId = resolveTenantId()

  config.headers = config.headers || {}
  config.headers.Accept = 'application/json'

  if (snapshot?.session.accessToken) {
    config.headers.Authorization = `Bearer ${snapshot.session.accessToken}`
  }

  if (tenantId) {
    config.headers['X-Tenant-ID'] = tenantId
  }

  return config
})

axiosInstance.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    const normalizedError = normalizeApiError(error)
    if (normalizedError.status === 401) {
      unauthorizedHandler?.(normalizedError)
    }
    return Promise.reject(normalizedError)
  },
)

export async function requestData<T>(config: AxiosRequestConfig) {
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
