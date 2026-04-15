/**
 * 前端统一日志器。
 *
 * 设计动机：error-handle.ts / http/error.ts / 业务页面到处散落 console.error /
 * console.warn。我们需要一个单一入口，把日志：
 *  1. 按 level 打到浏览器 console（开发体验）；
 *  2. 结构化、批量地上报到 /api/v1/telemetry/logs（运维观测）；
 *  3. 自动携带 request_id / route / user / session / viewport 等上下文，
 *     和后端 audit / access log 打通成同一条链路。
 *
 * 规范文档：docs/guides/logging-spec.md §4（前端日志契约）。
 *
 * 用法：
 *
 *   import { logger } from '@/utils/logger'
 *   logger.info('dashboard.mounted', { route: '/dashboard' })
 *   logger.warn('http.retry', { url, status })
 *   logger.error('payment.submit_failed', { err, orderId })
 *
 * 关键约定：
 *  - 第一参数是稳定的"事件名"（dot-case），不是面向人的句子 —— 方便后端按 event 聚合；
 *  - 第二参数是结构化字段 map，不要拼字符串；
 *  - 永远异步落盘：logger.xxx 不 return Promise，调用方无需 await。
 */
/** 日志级别；顺序与后端 zap 保持一致。 */
export type LogLevel = 'debug' | 'info' | 'warn' | 'error'

/**
 * 一条待上报的日志条目。
 *
 * 字段命名用 snake_case 以便直接对齐 OpenAPI 契约（TelemetryLogEntry，见
 * backend/api/openapi/domains/telemetry/schemas.yaml），后端 ogen 生成的
 * Go 结构体也是同一批字段 —— 这里保持 snake_case 可以省掉一层序列化映射，
 * 也避免 request_id / session_id 被某层拦截器误转 camelCase。
 */
export interface LogEntry {
  level: LogLevel
  /** 稳定事件名（dot-case），后端按此聚合。 */
  event: string
  /** 发生时间（ISO8601，UTC）。 */
  timestamp: string
  /** 请求级/页面级 request_id，和后端 X-Request-Id 对齐。 */
  request_id?: string
  /** 结构化字段（已过滤敏感 key）。 */
  context?: Record<string, unknown>
  /** 错误对象序列化快照（仅 error 级生效）。 */
  error?: {
    name?: string
    message?: string
    stack?: string
  }
  /** 页面路由路径，路由变更时会刷新。 */
  route?: string
  /** 当前登录用户 ID，未登录为空字符串。 */
  user_id?: string
  /** 会话 ID，浏览器关闭前保持一致。 */
  session_id: string
  /** 浏览器 UA，用于终端侧分布。 */
  user_agent: string
  /** 视口大小，排查样式/布局问题用。 */
  viewport: { w: number; h: number }
}

/** Logger 可调参数。生产默认即可，测试 / 特殊场景可覆写。 */
export interface LoggerOptions {
  /** 上报端点，默认 `/api/v1/telemetry/logs`。 */
  endpoint?: string
  /** 单批最大条数；超过后立即 flush。默认 20。 */
  batchSize?: number
  /** 批处理时间窗口（毫秒），到点无论是否够数都 flush。默认 3000。 */
  flushIntervalMs?: number
  /** 最小上报级别；低于此级别只打 console 不上报。默认 'info'。 */
  minReportLevel?: LogLevel
  /** 禁用上报（SSR / 测试 / 隐私模式下可关）。 */
  disableRemote?: boolean
}

const LEVEL_ORDER: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3,
}

/**
 * 敏感字段名黑名单。和后端 DefaultRedactFields 对齐；命中的 key 在序列化前
 * 会被替换为 '[REDACTED]'，防止 password / token / 身份证 等明文上报。
 */
const REDACT_KEYS = new Set([
  'password',
  'pwd',
  'secret',
  'token',
  'access_token',
  'refresh_token',
  'api_key',
  'authorization',
  'cookie',
  'credit_card',
  'card_no',
  'id_card',
  'phone',
  'mobile',
])

/** session_id 在整个标签页生命周期内保持一致。 */
function generateSessionId(): string {
  // 不依赖 crypto.randomUUID 的老浏览器兜底。
  const rand = Math.random().toString(36).slice(2, 10)
  return `${Date.now().toString(36)}-${rand}`
}

/** 递归脱敏：深拷贝 + 替换命中的 key。 */
function redact(value: unknown, depth = 0): unknown {
  if (depth > 6 || value == null) return value
  if (Array.isArray(value)) return value.map((v) => redact(v, depth + 1))
  if (typeof value === 'object') {
    const out: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      out[k] = REDACT_KEYS.has(k.toLowerCase()) ? '[REDACTED]' : redact(v, depth + 1)
    }
    return out
  }
  return value
}

/** 错误对象抽取：Error / HttpError 等统一成 {name, message, stack}。 */
function serializeError(err: unknown): LogEntry['error'] | undefined {
  if (err == null) return undefined
  if (err instanceof Error) {
    return { name: err.name, message: err.message, stack: err.stack }
  }
  // 兜底：保留可枚举字段
  try {
    return { message: JSON.stringify(err) }
  } catch {
    return { message: String(err) }
  }
}

/**
 * 内部 Logger 实现。调用方通过 `logger` 单例使用。
 */
export class Logger {
  private buffer: LogEntry[] = []
  private flushTimer: ReturnType<typeof setTimeout> | null = null
  private readonly sessionId: string
  private readonly opts: Required<LoggerOptions>
  private userId = ''
  private currentRoute = ''

  constructor(opts: LoggerOptions = {}) {
    this.sessionId = generateSessionId()
    this.opts = {
      endpoint: opts.endpoint ?? '/api/v1/telemetry/logs',
      batchSize: opts.batchSize ?? 20,
      flushIntervalMs: opts.flushIntervalMs ?? 3000,
      minReportLevel: opts.minReportLevel ?? 'info',
      disableRemote: opts.disableRemote ?? false,
    }
    this.installUnloadFlush()
  }

  /**
   * setUser 在登录成功 / 切换用户 / 退出时调用，更新后续日志的 user_id。
   * 传空串表示匿名。
   */
  setUser(userId: string): void {
    this.userId = userId || ''
  }

  /** setRoute 在 vue-router afterEach 里调用，跟踪当前路由。 */
  setRoute(path: string): void {
    this.currentRoute = path || ''
  }

  debug(event: string, context?: Record<string, unknown>): void {
    this.log('debug', event, context)
  }

  info(event: string, context?: Record<string, unknown>): void {
    this.log('info', event, context)
  }

  warn(event: string, context?: Record<string, unknown>): void {
    this.log('warn', event, context)
  }

  error(event: string, contextOrError?: Record<string, unknown> | unknown): void {
    // 允许直接传 Error：logger.error('render.crash', err)
    if (contextOrError instanceof Error) {
      this.log('error', event, { err: contextOrError })
      return
    }
    this.log('error', event, contextOrError as Record<string, unknown> | undefined)
  }

  /** 手动刷盘；生产很少用，调试测试会用。 */
  flush(): void {
    if (this.buffer.length === 0) return
    const batch = this.buffer
    this.buffer = []
    this.clearTimer()
    this.transport(batch)
  }

  private log(level: LogLevel, event: string, context?: Record<string, unknown>): void {
    // 1. 控制台输出（开发体验）：error/warn 永远打，debug/info 仅在 DEV。
    this.console(level, event, context)

    if (this.opts.disableRemote) return
    if (LEVEL_ORDER[level] < LEVEL_ORDER[this.opts.minReportLevel]) return

    // 2. 构造 entry + 脱敏
    const redactedContext = context ? (redact(context) as Record<string, unknown>) : undefined
    const errorSnapshot =
      level === 'error' && context && 'err' in context ? serializeError(context.err) : undefined

    const entry: LogEntry = {
      level,
      event,
      timestamp: new Date().toISOString(),
      request_id: getRequestIdSafely(),
      context: redactedContext,
      error: errorSnapshot,
      route: this.currentRoute,
      user_id: this.userId,
      session_id: this.sessionId,
      user_agent: typeof navigator !== 'undefined' ? navigator.userAgent : '',
      viewport: {
        w: typeof window !== 'undefined' ? window.innerWidth : 0,
        h: typeof window !== 'undefined' ? window.innerHeight : 0,
      },
    }

    this.buffer.push(entry)
    if (this.buffer.length >= this.opts.batchSize) {
      this.flush()
    } else {
      this.scheduleFlush()
    }
  }

  private console(level: LogLevel, event: string, context?: Record<string, unknown>): void {
    const isDev = typeof import.meta !== 'undefined' && (import.meta as any).env?.DEV
    if (!isDev && (level === 'debug' || level === 'info')) return
    const tag = `[${level.toUpperCase()}] ${event}`
    switch (level) {
      case 'debug':
      case 'info':
        // eslint-disable-next-line no-console
        console.log(tag, context ?? '')
        break
      case 'warn':
        // eslint-disable-next-line no-console
        console.warn(tag, context ?? '')
        break
      case 'error':
        // eslint-disable-next-line no-console
        console.error(tag, context ?? '')
        break
    }
  }

  private scheduleFlush(): void {
    if (this.flushTimer) return
    this.flushTimer = setTimeout(() => {
      this.flushTimer = null
      this.flush()
    }, this.opts.flushIntervalMs)
  }

  private clearTimer(): void {
    if (this.flushTimer) {
      clearTimeout(this.flushTimer)
      this.flushTimer = null
    }
  }

  /**
   * transport 优先 sendBeacon（页面卸载也能送达），失败 fallback 到 fetch。
   * 无论如何不能抛出：业务不感知日志上报失败。
   */
  private transport(batch: LogEntry[]): void {
    if (batch.length === 0) return
    const payload = JSON.stringify({ entries: batch })
    try {
      if (typeof navigator !== 'undefined' && typeof navigator.sendBeacon === 'function') {
        const blob = new Blob([payload], { type: 'application/json' })
        const ok = navigator.sendBeacon(this.opts.endpoint, blob)
        if (ok) return
      }
    } catch {
      /* fall through to fetch */
    }
    try {
      fetch(this.opts.endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payload,
        keepalive: true,
        credentials: 'same-origin',
      }).catch(() => {
        /* silent: 日志丢失不影响业务 */
      })
    } catch {
      /* silent */
    }
  }

  /** 页面关闭前最后一次 flush —— pagehide 比 beforeunload 在移动端更可靠。 */
  private installUnloadFlush(): void {
    if (typeof window === 'undefined') return
    const handler = () => this.flush()
    window.addEventListener('pagehide', handler)
    window.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') this.flush()
    })
  }
}

/**
 * lastServerRequestId 由 axios response 拦截器写入，logger 拉取做 join key。
 * 它是"最近一次 HTTP 响应带回的 X-Request-Id"，足以让前端日志在大多数
 * 场景下找到对应的后端 audit/access 日志。SSR / 非请求上下文为空串。
 */
let lastServerRequestId = ''

/** setLastRequestId 由 HTTP 拦截器在 response 到达时调用。 */
export function setLastRequestId(id: string): void {
  lastServerRequestId = id || ''
}

function getRequestIdSafely(): string | undefined {
  return lastServerRequestId || undefined
}

/** 单例：业务侧直接 `import { logger } from '@/utils/logger'`。 */
export const logger = new Logger()
