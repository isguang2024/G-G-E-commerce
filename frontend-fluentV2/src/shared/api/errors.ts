import axios from 'axios'
import type { ApiErrorShape } from '@/shared/types/api'

export class ApiError extends Error {
  public status?: number
  public code?: string

  constructor({ message, status, code }: ApiErrorShape) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export function normalizeApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error
  }

  if (axios.isAxiosError(error)) {
    return new ApiError({
      message: error.response?.data?.message || error.message || '请求失败',
      status: error.response?.status,
      code: error.code,
    })
  }

  if (error instanceof Error) {
    return new ApiError({ message: error.message })
  }

  return new ApiError({ message: '未知异常' })
}
