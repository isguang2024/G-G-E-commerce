import { ApiStatus } from './status'
import { HttpError } from './error'
import {
  isUnauthorizedBusinessCode,
  shouldBypassUnauthorizedLogout,
  triggerUnauthorizedLogout
} from './auth-session'

function normalizeV5BusinessCode(error: any): number {
  const code = Number(error?.code || 0)
  return Number.isFinite(code) && code > 0 ? code : 0
}

function normalizeV5StatusCode(status?: number): number {
  const responseStatus = Number(status || 0)
  return Number.isFinite(responseStatus) && responseStatus > 0 ? responseStatus : 0
}

function normalizeV5ErrorMessage(error: any, statusCode: number): string {
  const backendMessage = `${error?.message || error?.msg || ''}`.trim()
  if (backendMessage) {
    return backendMessage
  }
  if (statusCode === ApiStatus.unauthorized) {
    return '未授权访问，请重新登录'
  }
  if (statusCode >= 500) {
    return '服务器开小差了，请稍后重试'
  }
  return '请求失败'
}

export function createUnifiedV5HttpError(error: any, response?: Response): HttpError {
  const statusCode = normalizeV5StatusCode(response?.status)
  const businessCode = normalizeV5BusinessCode(error)
  const normalizedCode = businessCode || statusCode || ApiStatus.error
  const httpError = new HttpError(normalizeV5ErrorMessage(error, statusCode), normalizedCode, {
    data: {
      ...(error && typeof error === 'object' ? error : {}),
      status: statusCode
    },
    url: response?.url,
    method: undefined
  })

  if (
    (statusCode === ApiStatus.unauthorized || isUnauthorizedBusinessCode(normalizedCode)) &&
    !shouldBypassUnauthorizedLogout(response?.url)
  ) {
    triggerUnauthorizedLogout(httpError)
  }

  return httpError
}

export async function unwrapV5Response<T>(
  promise: Promise<{ data?: T; error?: any; response: Response }>
): Promise<T> {
  const { data, error, response } = await promise
  if (error) {
    throw createUnifiedV5HttpError(error, response)
  }
  if (data === undefined) {
    throw new Error('v5Client: empty response')
  }
  return data as T
}
