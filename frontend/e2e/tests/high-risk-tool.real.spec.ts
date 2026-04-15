import { expect, test, type Page } from '@playwright/test'
import { waitForDashboardReady, collectAbortEvents, filterApiAborts } from '../support/high-risk'

/**
 * 高风险工具页回归（P2 组）——对应 docs/guides/high-risk-remediation-matrix.md
 *   B1 /system/access-trace 访问链路测试
 *   B2 /system/api-endpoint  API 管理（同步/清理）
 *
 * 断言目标：
 *   - 访问链路测试：trace-summary / trace-node / trace-node-status / trace-node-reason 等结构化节点存在；
 *     即使"未触发测试"也应看到空态；触发测试后 trace-summary 与 trace-node 列表同时出现。
 *   - API 管理：sync / unregistered / cleanup-stale 按钮带有 data-testid，
 *     点击 cleanup-stale 后：
 *       * 若存在失效注册项 → 弹出 api-endpoint-stale-dialog + 含 api-endpoint-stale-table；
 *       * 若不存在 → ElMessage 信息条显示并不阻塞（"stale 注册项场景不再出现" 的正向验证）。
 *   - 同时采样页面的 `/api/v1/` ERR_ABORTED，要求无批量 abort。
 */

const REAL_BACKEND_BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:5174'
const REAL_USERNAME = process.env.E2E_USERNAME || ''
const REAL_PASSWORD = process.env.E2E_PASSWORD || ''

async function login(page: Page) {
  await page.goto(`${REAL_BACKEND_BASE_URL}/account/auth/login?redirect=%2Fdashboard%2Fconsole`)
  await page.locator('input[type="text"]').first().fill(REAL_USERNAME)
  await page.locator('input[type="password"]').first().fill(REAL_PASSWORD)
  await page.getByRole('button', { name: /登录/i }).click()
  await waitForDashboardReady(page)
}

test.describe('高风险工具页观测基线 - 结果区结构化节点', () => {
  test.beforeEach(async () => {
    test.skip(!REAL_USERNAME || !REAL_PASSWORD, '缺少真实后端登录凭据')
  })

  test('B1 /system/access-trace 触发测试后生成 trace-summary + trace-node 列表', async ({
    page
  }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/access-trace`)
    await expect(page.getByRole('heading', { name: /访问链路测试/ })).toBeVisible()

    // 空态
    await expect(page.getByText(/点击「测试链路」查看结果/)).toBeVisible()

    // 选择第一个 App + 第一个用户后触发测试
    await page.locator('.trace-field').first().click()
    await page.getByRole('option').first().click({ trial: false })

    // 选到用户下拉（顺序为 App → 空间 → 协作空间 → 角色 → 用户）
    const userSelect = page.locator('.trace-field').nth(4)
    await userSelect.click()
    await page.getByRole('option').first().click()

    await page.getByRole('button', { name: /测试链路/ }).click()

    // 结构化节点稳定：trace-summary 作为 hidden meta 节点必定出现
    await expect(page.locator('[data-testid="trace-summary"]')).toHaveCount(1, {
      timeout: 15_000
    })
    // trace-node 至少出现一个（菜单 or 页面）
    await expect(page.locator('[data-testid="trace-node"]').first()).toBeVisible()
    // 每个节点上的 status 标识可读
    const statuses = page.locator('[data-testid="trace-node-status"]')
    expect(await statuses.count()).toBeGreaterThan(0)

    expect(filterApiAborts(abortEvents)).toEqual([])
  })

  test('B2 /system/api-endpoint 同步/清理按钮带 data-testid，stale 清理弹层正确切换', async ({
    page
  }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/api-endpoint`)
    await expect(
      page.getByRole('heading', { name: /API 管理|接口注册表|API 注册表/ })
    ).toBeVisible()

    // 三个关键操作入口必须带 testid
    await expect(page.locator('[data-testid="api-endpoint-sync-button"]')).toBeVisible()
    await expect(page.locator('[data-testid="api-endpoint-unregistered-button"]')).toBeVisible()
    const cleanupButton = page.locator('[data-testid="api-endpoint-cleanup-stale-button"]')
    await expect(cleanupButton).toBeVisible()

    // 点击清理入口：可能弹出 stale 弹层，也可能因为无 stale 而只有 ElMessage
    await cleanupButton.click()

    const dialog = page.locator('[data-testid="api-endpoint-stale-dialog"]')
    const infoToast = page.locator('.el-message--info')

    // Promise.race 语义：两个里有且仅有一个会出现
    await Promise.race([
      dialog.waitFor({ state: 'visible', timeout: 8_000 }).catch(() => null),
      infoToast
        .filter({ hasText: /当前没有可清理的失效 API/ })
        .waitFor({ state: 'visible', timeout: 8_000 })
        .catch(() => null)
    ])

    if (await dialog.isVisible().catch(() => false)) {
      // 正向：弹层结构必须含 stale 表格 + 提交按钮
      await expect(page.locator('[data-testid="api-endpoint-stale-table"]')).toBeVisible()
      await expect(page.locator('[data-testid="api-endpoint-stale-confirm-button"]')).toBeVisible()
      await page.keyboard.press('Escape')
    } else {
      // 反向：没有 stale 的正向验证（"stale 注册项场景不再出现"）
      await expect(infoToast.filter({ hasText: /当前没有可清理的失效 API/ })).toBeVisible()
    }

    expect(filterApiAborts(abortEvents)).toEqual([])
  })
})
