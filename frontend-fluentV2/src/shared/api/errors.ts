import axios from 'axios'
import type { ApiErrorPayload, ApiErrorShape } from '@/shared/types/api'

export class ApiError extends Error {
  public status?: number
  public code?: string
  public businessCode?: number | string
  public details?: unknown

  constructor({ message, status, code, businessCode, details }: ApiErrorShape) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.businessCode = businessCode
    this.details = details
  }
}

function resolveBusinessPayload(payload: unknown): ApiErrorPayload {
  if (!payload || typeof payload !== 'object') {
    return {}
  }

  return payload as ApiErrorPayload
}

export function normalizeApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error
  }

  if (axios.isAxiosError(error)) {
    const payload = resolveBusinessPayload(error.response?.data)
    return new ApiError({
      message: payload.message || error.message || '请求失败',
      status: error.response?.status,
      code: error.code,
      businessCode: payload.code,
      details: payload.data,
    })
  }

  if (error instanceof Error) {
    return new ApiError({ message: error.message })
  }

  return new ApiError({ message: '未知异常' })
}

export function isUnauthorizedError(error: unknown) {
  return normalizeApiError(error).status === 401
}
