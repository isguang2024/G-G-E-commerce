// 回归：前端 logger 的字段契约与脱敏行为
//
// 目标：不引入浏览器 / Playwright，只用 node:test + tsx 验证 Logger 类的：
//   1. flush 时 transport 被调用，payload 形如 { entries: [...] }
//   2. entry 字段使用 snake_case（session_id / user_agent / viewport）
//   3. Error 对象被序列化进 entry.error
//   4. 敏感字段（password / token）被替换为 [REDACTED]
//   5. level < minReportLevel 的日志不会上报
//
// 实现要点：logger 引用 navigator / window / document，Node 环境下需要先挂 stub 再
// import。同时我们 stub 掉 fetch / sendBeacon 把"实际发送"转成可断言的本地变量。

import test from 'node:test'
import assert from 'node:assert/strict'

type IngestPayload = { entries: Array<Record<string, unknown>> }

// ── 环境 stub ──────────────────────────────────────────────────────────────
// Node 22+ 把 navigator 变成 getter，必须用 defineProperty 覆盖。
const captured: IngestPayload[] = []

function defineGlobal(name: string, value: unknown) {
  Object.defineProperty(globalThis, name, {
    configurable: true,
    writable: true,
    value,
  })
}

defineGlobal('navigator', {
  userAgent: 'node-test/1.0',
  sendBeacon: () => false, // 强制走 fetch 路径
})
defineGlobal('window', {
  innerWidth: 1920,
  innerHeight: 1080,
  addEventListener: () => {},
})
defineGlobal('document', { visibilityState: 'visible' })
defineGlobal('fetch', (_url: string | URL, init?: RequestInit) => {
  try {
    const body = typeof init?.body === 'string' ? init!.body! : ''
    captured.push(JSON.parse(body) as IngestPayload)
  } catch {
    captured.push({ entries: [] })
  }
  return Promise.resolve({
    ok: true,
    status: 200,
    json: async () => ({ accepted: 1, dropped: 0 }),
  } as unknown as Response)
})

// import.meta.env 通过 Vite 注入；Node 下手工伪造（避免 logger.console 里 DEV 分支报错）
// tsx 允许 import.meta，但不会注入 env 字段。我们直接给 Logger 显式配置绕过 DEV 判定：
//   - minReportLevel: 'debug'  让 debug 也能上报（方便验证 level 过滤的边界）
//   - disableRemote: false
// 并在下面单独测试 level 过滤。

import { Logger } from '../src/utils/logger/index'

test('logger.flush 上报 payload 包含 snake_case 字段 + 脱敏 + error 快照', () => {
  captured.length = 0
  const l = new Logger({
    endpoint: '/api/v1/telemetry/logs',
    batchSize: 10,
    flushIntervalMs: 10_000, // 避免 timer 干扰，用 flush() 主动送出
    minReportLevel: 'debug',
    disableRemote: false,
  })

  l.error('http.error', {
    url: '/api/v1/demo',
    status: 500,
    password: 'secret123',
    token: 'bearer-xxx',
    nested: { authorization: 'Basic AAA' },
    err: new Error('mock failure'),
  })
  l.flush()

  assert.equal(captured.length, 1, '应该送出一次批次')
  const batch = captured[0]
  assert.ok(Array.isArray(batch.entries))
  assert.equal(batch.entries.length, 1)

  const entry = batch.entries[0]
  // snake_case 字段
  assert.equal(entry.level, 'error')
  assert.equal(entry.event, 'http.error')
  assert.equal(typeof entry.session_id, 'string')
  assert.ok((entry.session_id as string).length > 0)
  assert.equal(typeof entry.user_agent, 'string')
  assert.deepEqual(entry.viewport, { w: 1920, h: 1080 })
  assert.equal(typeof entry.timestamp, 'string')
  assert.ok(!Number.isNaN(Date.parse(entry.timestamp as string)))

  // error 快照
  const snap = entry.error as { name?: string; message?: string }
  assert.equal(snap?.message, 'mock failure')
  assert.ok((snap?.name || '').length > 0)

  // 脱敏（顶层 + 嵌套）
  const ctx = entry.context as Record<string, unknown>
  assert.equal(ctx.password, '[REDACTED]')
  assert.equal(ctx.token, '[REDACTED]')
  const nested = ctx.nested as Record<string, unknown>
  assert.equal(nested.authorization, '[REDACTED]')
  // 非敏感字段保留
  assert.equal(ctx.url, '/api/v1/demo')
  assert.equal(ctx.status, 500)
})

test('logger 在 level < minReportLevel 时仅打 console、不上报', () => {
  captured.length = 0
  const l = new Logger({
    endpoint: '/api/v1/telemetry/logs',
    batchSize: 10,
    flushIntervalMs: 10_000,
    minReportLevel: 'warn',
    disableRemote: false,
  })
  l.info('page.view', { route: '/home' })
  l.debug('noise', { foo: 1 })
  l.flush()
  assert.equal(captured.length, 0, 'info/debug 不应触发上报')

  l.warn('http.retry', { url: '/x', attempt: 2 })
  l.flush()
  assert.equal(captured.length, 1)
  const entry = captured[0].entries[0]
  assert.equal(entry.level, 'warn')
  assert.equal(entry.event, 'http.retry')
})

test('logger.setUser / setRoute 注入到后续 entry', () => {
  captured.length = 0
  const l = new Logger({
    endpoint: '/api/v1/telemetry/logs',
    batchSize: 10,
    flushIntervalMs: 10_000,
    minReportLevel: 'info',
    disableRemote: false,
  })
  l.setUser('u-42')
  l.setRoute('/dashboard')
  l.info('dashboard.mounted', { widgets: 3 })
  l.flush()

  assert.equal(captured.length, 1)
  const entry = captured[0].entries[0]
  assert.equal(entry.user_id, 'u-42')
  assert.equal(entry.route, '/dashboard')

  // 退出登录
  l.setUser('')
  l.info('logout', {})
  l.flush()
  const second = captured[1].entries[0]
  assert.equal(second.user_id, '')
})

test('disableRemote=true 时不调用 fetch', () => {
  captured.length = 0
  const l = new Logger({
    endpoint: '/api/v1/telemetry/logs',
    disableRemote: true,
  })
  l.error('will.not.send', { err: new Error('no transport') })
  l.flush()
  assert.equal(captured.length, 0)
})
