/**
 * HTTP 请求封装模块
 * 基于 Axios 封装的 HTTP 请求工具，提供统一的请求/响应处理
 *
 * ## 主要功能
 *
 * - 请求/响应拦截器（自动添加 Token、统一错误处理）
 * - 401 未授权自动登出（带防抖机制）
 * - 请求失败自动重试（可配置）
 * - 统一的成功/错误消息提示
 * - 支持 GET/POST/PUT/DELETE 等常用方法
 *
 * @module utils/http
 * @author Art Design Pro Team
 */

import axios, { AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { useUserStore } from '@/store/modules/user'
import { useCollaborationWorkspaceStore } from '@/store/modules/collaboration-workspace'
import { useWorkspaceStore } from '@/store/modules/workspace'
import { ApiStatus } from './status'
import { HttpError, handleError, showError, showSuccess } from './error'
import { $t } from '@/locales'
import { BaseResponse } from '@/types'

/** 请求配置常量 */
const REQUEST_TIMEOUT = 15000
const LOGOUT_DELAY = 500
const MAX_RETRIES = 0
const RETRY_DELAY = 1000
const UNAUTHORIZED_RESET_DELAY = 3000

/**
 * 401 防抖：使用 Promise 哨兵，确保多并发 401 仅触发一次登出/提示。
 * 旧实现使用全局布尔位 + setTimeout，存在以下竞态：
 *   1. 同一 tick 内多个 401 同时进入，可能都看到 false 而都执行登出；
 *   2. setTimeout 复位与下一次 401 之间无 happens-before 关系。
 */
let unauthorizedHandling: Promise<void> | null = null

/** 扩展 AxiosRequestConfig */
interface ExtendedAxiosRequestConfig extends AxiosRequestConfig {
  showErrorMessage?: boolean
  showSuccessMessage?: boolean
  skipAuthWorkspaceHeader?: boolean
  skipCollaborationWorkspaceHeader?: boolean
  skipWorkspaceHeader?: boolean
  /** GET 缓存 TTL（毫秒）。仅 GET 生效；> 0 时启用 */
  cache?: number
  /** 关闭 in-flight 去重（默认 GET 自动去重） */
  dedupe?: boolean
}

/**
 * GET 请求层优化：
 *   1. in-flight 去重：同 key 的并发 GET 只发一次，复用 Promise；
 *   2. 可选短 TTL 缓存：调用方显式传入 `cache: ms` 时启用；
 *
 * key = METHOD + URL + JSON(params)
 * 注意：仅对 GET 启用，避免改变写操作语义。
 */
interface CacheEntry {
  expires: number
  value: unknown
}
const inflightMap = new Map<string, Promise<unknown>>()
const responseCache = new Map<string, CacheEntry>()

function buildCacheKey(config: ExtendedAxiosRequestConfig): string {
  const method = (config.method || 'GET').toUpperCase()
  const url = config.url || ''
  let params = ''
  try {
    params = config.params ? JSON.stringify(config.params) : ''
  } catch {
    params = ''
  }
  return `${method}|${url}|${params}`
}

const { VITE_API_URL, VITE_WITH_CREDENTIALS } = import.meta.env

/** Axios实例 */
const axiosInstance = axios.create({
  timeout: REQUEST_TIMEOUT,
  baseURL: VITE_API_URL,
  withCredentials: VITE_WITH_CREDENTIALS === 'true',
  validateStatus: (status) => status >= 200 && status < 300,
  transformResponse: [
    (data, headers) => {
      const contentType = headers['content-type']
      if (contentType?.includes('application/json')) {
        try {
          return JSON.parse(data)
        } catch {
          return data
        }
      }
      return data
    }
  ]
})

/** 请求拦截器 */
axiosInstance.interceptors.request.use(
  (
    request: InternalAxiosRequestConfig & {
      skipAuthWorkspaceHeader?: boolean
      skipCollaborationWorkspaceHeader?: boolean
      skipWorkspaceHeader?: boolean
    }
  ) => {
    const { accessToken } = useUserStore()
    const { currentCollaborationWorkspaceId, currentContextMode } = useCollaborationWorkspaceStore()
    const { currentAuthWorkspaceId } = useWorkspaceStore()
    if (accessToken) {
      // 添加 Bearer 前缀（如果还没有）
      const token = accessToken.startsWith('Bearer ') ? accessToken : `Bearer ${accessToken}`
      request.headers.set('Authorization', token)
    }

    const skipAuthWorkspaceHeader = Boolean(request.skipAuthWorkspaceHeader)
    if (!request.skipWorkspaceHeader && !skipAuthWorkspaceHeader && currentAuthWorkspaceId) {
      request.headers.set('X-Auth-Workspace-Id', currentAuthWorkspaceId)
    }

    if (!request.skipCollaborationWorkspaceHeader && currentContextMode === 'collaboration') {
      if (currentCollaborationWorkspaceId) {
        request.headers.set('X-Collaboration-Workspace-Id', currentCollaborationWorkspaceId)
      }
    }

    if (request.data && !(request.data instanceof FormData) && !request.headers['Content-Type']) {
      request.headers.set('Content-Type', 'application/json')
      request.data = JSON.stringify(request.data)
    }

    return request
  },
  (error) => {
    showError(createHttpError($t('httpMsg.requestConfigError'), ApiStatus.error))
    return Promise.reject(error)
  }
)

/** 响应拦截器 */
axiosInstance.interceptors.response.use(
  (response: AxiosResponse<BaseResponse>) => {
    const raw = response.data as any
    // V5 兼容：ogen handler 直接返回裸 schema（无 {code,data,message} 信封），
    // legacy request<T> 仍走老的 res.data.data 解包路径，这里把裸响应包成
    // 兼容信封，让两边接口能在过渡期共存。判定：response.data 不是带 code
    // 字段的对象就视为 v5 裸响应。
    if (
      response.status >= 200 &&
      response.status < 300 &&
      (raw == null || typeof raw !== 'object' || raw.code === undefined)
    ) {
      ;(response as any).data = { code: 0, data: raw }
      return response
    }

    const { code, message, msg } = response.data
    // 后端返回 code: 0 表示成功，其他值表示错误
    if (code === 0) return response
    // 401 未授权错误
    const errorMsg = message || msg || $t('httpMsg.requestFailed')
    if (code === 401 || response.status === ApiStatus.unauthorized) {
      handleUnauthorizedError(errorMsg)
    }
    // 传递完整的响应数据，包括data字段（可能包含角色列表等信息）
    throw createHttpError(errorMsg, code, { data: response.data.data })
  },
  (error) => {
    // HTTP 状态码错误处理
    if (error.response?.status === ApiStatus.unauthorized) {
      handleUnauthorizedError()
    }
    return Promise.reject(handleError(error))
  }
)

/** 统一创建HttpError */
function createHttpError(message: string, code: number, options?: { data?: unknown }) {
  return new HttpError(message, code, options)
}

/** 处理401错误（Promise 哨兵，全局只触发一次登出/提示） */
function handleUnauthorizedError(message?: string): never {
  const error = createHttpError(message || $t('httpMsg.unauthorized'), ApiStatus.unauthorized)

  if (!unauthorizedHandling) {
    unauthorizedHandling = (async () => {
      try {
        showError(error, true)
        logOut()
      } finally {
        // 留出窗口期吞掉同一波 401，再放开下一次
        setTimeout(() => {
          unauthorizedHandling = null
        }, UNAUTHORIZED_RESET_DELAY)
      }
    })()
  }

  throw error
}

/** 退出登录函数 */
function logOut() {
  setTimeout(() => {
    useUserStore().logOut()
  }, LOGOUT_DELAY)
}

/** 是否需要重试 */
function shouldRetry(statusCode: number) {
  return [
    ApiStatus.requestTimeout,
    ApiStatus.internalServerError,
    ApiStatus.badGateway,
    ApiStatus.serviceUnavailable,
    ApiStatus.gatewayTimeout
  ].includes(statusCode)
}

/** 请求重试逻辑 */
async function retryRequest<T>(
  config: ExtendedAxiosRequestConfig,
  retries: number = MAX_RETRIES
): Promise<T> {
  try {
    return await request<T>(config)
  } catch (error) {
    if (retries > 0 && error instanceof HttpError && shouldRetry(error.code)) {
      await delay(RETRY_DELAY)
      return retryRequest<T>(config, retries - 1)
    }
    throw error
  }
}

/** 延迟函数 */
function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/** 请求函数 */
async function request<T = any>(config: ExtendedAxiosRequestConfig): Promise<T> {
  const method = (config.method || 'GET').toUpperCase()

  // POST | PUT 参数自动填充
  if (['POST', 'PUT'].includes(method) && config.params && !config.data) {
    config.data = config.params
    config.params = undefined
  }

  // GET 缓存 / 去重
  const isGet = method === 'GET'
  const dedupeEnabled = isGet && config.dedupe !== false
  const cacheTtl = isGet && typeof config.cache === 'number' ? config.cache : 0
  const cacheKey = dedupeEnabled || cacheTtl > 0 ? buildCacheKey(config) : ''

  if (cacheTtl > 0 && cacheKey) {
    const hit = responseCache.get(cacheKey)
    if (hit && hit.expires > Date.now()) {
      return hit.value as T
    }
  }
  if (dedupeEnabled && cacheKey && inflightMap.has(cacheKey)) {
    return inflightMap.get(cacheKey) as Promise<T>
  }

  const exec = doRequest<T>(config).then((value) => {
    if (cacheTtl > 0 && cacheKey) {
      responseCache.set(cacheKey, { expires: Date.now() + cacheTtl, value })
    }
    return value
  })

  if (dedupeEnabled && cacheKey) {
    inflightMap.set(cacheKey, exec)
    exec.finally(() => inflightMap.delete(cacheKey))
  }

  return exec
}

/** 实际下发请求 */
async function doRequest<T = any>(config: ExtendedAxiosRequestConfig): Promise<T> {
  try {
    const res = await axiosInstance.request<BaseResponse<T>>(config)

    // 显示成功消息
    const successMsg = res.data.message || res.data.msg
    if (config.showSuccessMessage && successMsg) {
      showSuccess(successMsg)
    }

    return res.data.data as T
  } catch (error) {
    if (error instanceof HttpError && error.code !== ApiStatus.unauthorized) {
      const showMsg = config.showErrorMessage !== false
      showError(error, showMsg)
    }
    return Promise.reject(error)
  }
}

/** API方法集合 */
const api = {
  get<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'GET' })
  },
  post<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'POST' })
  },
  put<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'PUT' })
  },
  del<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>({ ...config, method: 'DELETE' })
  },
  request<T>(config: ExtendedAxiosRequestConfig) {
    return retryRequest<T>(config)
  }
}

export default api
