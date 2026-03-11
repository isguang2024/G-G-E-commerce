/**
 * HTTP 错误处理模块
 *
 * 提供统一的 HTTP 请求错误处理机制
 *
 * ## 主要功能
 *
 * - 自定义 HttpError 错误类，封装错误信息、状态码、时间戳等
 * - 错误拦截和转换，将 Axios 错误转换为标准的 HttpError
 * - 错误消息国际化处理，根据状态码返回对应的多语言错误提示
 * - 错误日志记录，便于问题追踪和调试
 * - 错误和成功消息的统一展示
 * - 类型守卫函数，用于判断错误类型
 *
 * ## 使用场景
 *
 * - HTTP 请求拦截器中统一处理错误
 * - 业务代码中捕获和处理特定错误
 * - 错误日志收集和上报
 *
 * @module utils/http/error
 * @author Art Design Pro Team
 */
import { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import { ApiStatus } from './status'
import { $t } from '@/locales'

// 错误响应接口
export interface ErrorResponse {
  /** 错误状态码 */
  code: number
  /** 错误消息 */
  msg: string
  /** 错误附加数据 */
  data?: unknown
}

// 错误日志数据接口
export interface ErrorLogData {
  /** 错误状态码 */
  code: number
  /** 错误消息 */
  message: string
  /** 错误附加数据 */
  data?: unknown
  /** 错误发生时间戳 */
  timestamp: string
  /** 请求 URL */
  url?: string
  /** 请求方法 */
  method?: string
  /** 错误堆栈信息 */
  stack?: string
}

// 自定义 HttpError 类
export class HttpError extends Error {
  public readonly code: number
  public readonly data?: unknown
  public readonly timestamp: string
  public readonly url?: string
  public readonly method?: string

  constructor(
    message: string,
    code: number,
    options?: {
      data?: unknown
      url?: string
      method?: string
    }
  ) {
    super(message)
    this.name = 'HttpError'
    this.code = code
    this.data = options?.data
    this.timestamp = new Date().toISOString()
    this.url = options?.url
    this.method = options?.method
  }

  public toLogData(): ErrorLogData {
    return {
      code: this.code,
      message: this.message,
      data: this.data,
      timestamp: this.timestamp,
      url: this.url,
      method: this.method,
      stack: this.stack
    }
  }
}

/**
 * 获取错误消息
 * @param status 错误状态码
 * @returns 错误消息
 */
const getErrorMessage = (status: number): string => {
  const errorMap: Record<number, string> = {
    [ApiStatus.unauthorized]: 'httpMsg.unauthorized',
    [ApiStatus.forbidden]: 'httpMsg.forbidden',
    [ApiStatus.notFound]: 'httpMsg.notFound',
    [ApiStatus.methodNotAllowed]: 'httpMsg.methodNotAllowed',
    [ApiStatus.requestTimeout]: 'httpMsg.requestTimeout',
    [ApiStatus.internalServerError]: 'httpMsg.internalServerError',
    [ApiStatus.badGateway]: 'httpMsg.badGateway',
    [ApiStatus.serviceUnavailable]: 'httpMsg.serviceUnavailable',
    [ApiStatus.gatewayTimeout]: 'httpMsg.gatewayTimeout'
  }

  return $t(errorMap[status] || 'httpMsg.internalServerError')
}

/**
 * 处理错误
 * @param error 错误对象
 * @returns 错误对象
 */
export function handleError(error: AxiosError<ErrorResponse>): never {
  // 处理取消的请求
  if (error.code === 'ERR_CANCELED') {
    console.warn('Request cancelled:', error.message)
    throw new HttpError($t('httpMsg.requestCancelled'), ApiStatus.error)
  }

  const statusCode = error.response?.status
  const errorMessage = error.response?.data?.msg || error.message
  const requestConfig = error.config

  // 处理网络错误
  if (!error.response) {
    throw new HttpError($t('httpMsg.networkError'), ApiStatus.error, {
      url: requestConfig?.url,
      method: requestConfig?.method?.toUpperCase()
    })
  }

  // 处理 HTTP 状态码错误
  // 优先使用后端返回的错误消息（message 或 msg），如果没有则使用通用错误消息
  const backendMessage = error.response?.data?.message || error.response?.data?.msg
  const message = backendMessage || (statusCode ? getErrorMessage(statusCode) : errorMessage || $t('httpMsg.requestFailed'))
  
  // 提取响应数据：如果响应体有 data 字段，则传递 data；否则传递整个响应体
  // 这样可以在业务代码中通过 error.data.data 访问嵌套的数据（如 roles, roleCount）
  const responseData = error.response.data
  throw new HttpError(message, statusCode || ApiStatus.error, {
    data: responseData,
    url: requestConfig?.url,
    method: requestConfig?.method?.toUpperCase()
  })
}

/**
 * 显示错误消息
 * @param error 错误对象
 * @param showMessage 是否显示错误消息
 */
export function showError(error: HttpError, showMessage: boolean = true): void {
  if (!showMessage) return

  // 根据错误码决定是否显示全局消息
  const code = error.code
  if (!shouldShowErrorMessage(code)) {
    return
  }

  ElMessage.error(error.message)
  // 记录错误日志
  console.error('[HTTP Error]', error.toLogData())
}

/**
 * 根据错误码判断是否显示全局消息提示
 * @param code 错误码
 * @returns 是否显示消息
 */
function shouldShowErrorMessage(code: number): boolean {
  // 错误码规则：0=成功；1xxxx=参数/请求；2xxxx=认证/授权；3xxxx=业务/资源；5xxxx=服务端

  // 需要显示全局提示的错误码
  const showCodes = [
    // 1xxxx 参数/请求错误 - 全部显示
    1001, // 参数错误
    1002, // 参数缺失
    1003, // 参数格式错误
    1004, // 无效的 ID

    // 2xxxx 认证/授权错误 - 全部显示
    2001, // 未登录或 token 无效
    2002, // token 已过期
    2003, // 无权限
    2004, // 缺少 API Key
    2005, // Token 格式错误

    // 3xxxx 业务/资源错误 - 选择性显示
    3006, // 您暂无管理的团队
    3007, // 角色编码已存在
    3008, // 该用户已在团队中
    3011, // 系统默认菜单不可删除
    3012, // 无效的上级
    3013, // 业务冲突
    3014, // 用户名已存在
    3015, // 系统角色不可删除

    // 5xxxx 服务端错误 - 全部显示
    5001, // 内部错误
    5002, // 数据库错误
    5003  // 外部服务错误
  ]

  // 不需要显示全局提示的错误码（静默处理，由业务代码自行处理）
  const hideCodes = [
    3001, // 资源不存在（通常业务代码会自行处理）
    3002, // 用户不存在
    3003, // 团队不存在
    3004, // 菜单不存在
    3005, // 角色不存在
    3009, // 成员不在团队中
    3010  // 团队角色不存在或无权操作
  ]

  if (showCodes.includes(code)) {
    return true
  }
  if (hideCodes.includes(code)) {
    return false
  }

  // 未定义的错误码，默认显示（安全起见）
  return code >= 5000
}

/**
 * 显示成功消息
 * @param message 成功消息
 * @param showMessage 是否显示消息
 */
export function showSuccess(message: string, showMessage: boolean = true): void {
  if (showMessage) {
    ElMessage.success(message)
  }
}

/**
 * 判断是否为 HttpError 类型
 * @param error 错误对象
 * @returns 是否为 HttpError 类型
 */
export const isHttpError = (error: unknown): error is HttpError => {
  return error instanceof HttpError
}
