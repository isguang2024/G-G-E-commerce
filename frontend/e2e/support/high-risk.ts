import type { Page } from '@playwright/test'

/**
 * 高风险页面回归共用辅助函数。
 *
 * 背景：动态路由注册完成后 beforeEach 会执行 next({ path, replace: true })，
 * 导致 Layout 重挂载、部分首屏请求被浏览器兜底取消（ERR_ABORTED）。
 * 对前端逻辑无影响，但测试时会产生"随机 abort 日志"。
 *
 * 详见 docs/guides/dashboard-request-abort-rootcause.md。
 *
 * 本模块提供：
 * 1. `waitForDashboardReady`：等到动态路由注册后的再导航完成 + network idle，
 *    再开始断言，避免抓到过渡期间的不稳定状态。
 * 2. `collectAbortEvents`：收集 `requestfailed` 事件中与当前页面实际路径相关
 *    的 ERR_ABORTED，用于验证"回归中首屏不再批量报 abort"。
 */
export interface AbortEvent {
  url: string
  failure: string
}

/** 等待 dashboard 首屏稳定：路由跳到 /dashboard/console + network idle。 */
export async function waitForDashboardReady(page: Page): Promise<void> {
  await page.waitForURL((url) => new URL(url.toString()).pathname === '/dashboard/console')
  // networkidle 等待 500ms 无请求，以规避路由再导航导致的二次 mount。
  await page.waitForLoadState('networkidle')
}

/** 订阅当前 page 的 requestfailed，聚合为 AbortEvent 列表。 */
export function collectAbortEvents(page: Page): AbortEvent[] {
  const events: AbortEvent[] = []
  page.on('requestfailed', (request) => {
    const failure = request.failure()?.errorText || ''
    if (!failure.includes('ERR_ABORTED') && !failure.includes('net::ERR_ABORTED')) return
    events.push({ url: request.url(), failure })
  })
  return events
}

/** 仅保留 /api/v1/ 路径的 abort，滤掉 HMR / 静态资源的抖动。 */
export function filterApiAborts(events: AbortEvent[]): AbortEvent[] {
  return events.filter((e) => e.url.includes('/api/v1/'))
}
