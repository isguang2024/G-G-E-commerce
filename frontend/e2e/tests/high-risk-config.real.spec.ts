import { expect, test, type Page } from '@playwright/test'
import { waitForDashboardReady, collectAbortEvents, filterApiAborts } from '../support/high-risk'

/**
 * 高风险配置页回归（P2 组）——对应 docs/guides/high-risk-remediation-matrix.md
 * A1/A2/A3/A4：/system/register-entry、/system/login-page-template、/system/app、/system/menu-space
 *
 * 目的：
 *   - 验证每个页面都渲染了 `data-testid` 观测基线节点（卡片 / 行 / field-error 容器）。
 *   - 打开"新建"入口后提交空表单，确认必填 rules 触发对应的 field-error 节点；
 *     这样 E2E 与 `el-form-item :error` 回显契约绑死，后续后端错误键变更会立刻暴露回归。
 *   - 顺带统计首屏 ERR_ABORTED 不再成批出现（acceptance：回归中首屏不再批量报 abort）。
 *
 * 真实后端用例，需要 E2E_USERNAME / E2E_PASSWORD。
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

test.describe('高风险配置页观测基线 - el-form rules + field-error', () => {
  test.beforeEach(async () => {
    test.skip(!REAL_USERNAME || !REAL_PASSWORD, '缺少真实后端登录凭据')
  })

  test('A1 /system/register-entry 打开新建后必填字段触发 field-error', async ({ page }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/register-entry`)
    await expect(page.getByRole('heading', { name: /注册入口/ })).toBeVisible()

    // 已有数据则 register-entry-row 存在；否则仍然能打开新建
    await expect(page.locator('[data-testid="register-entry-row"]').first()).toBeVisible({
      timeout: 15_000
    })

    await page
      .getByRole('button', { name: /新建入口/ })
      .first()
      .click()
    await page
      .getByRole('button', { name: /(保存|确定|提交)/ })
      .first()
      .click()

    // 期望三个必填字段 error 节点都渲染
    const fieldErrors = page.locator('[data-testid="register-entry-field-error"]')
    await expect(fieldErrors).toHaveCount(3)
    await expect(fieldErrors.filter({ has: page.locator('.el-form-item__error') })).not.toHaveCount(
      0
    )

    // 关闭抽屉避免影响后续 case
    await page.keyboard.press('Escape')

    expect(filterApiAborts(abortEvents)).toEqual([])
  })

  test('A2 /system/login-page-template 打开新建后 template_key/name 触发 field-error', async ({
    page
  }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/login-page-template`)
    await expect(page.getByRole('heading', { name: /登录页模板/ })).toBeVisible()
    await expect(page.locator('[data-testid="login-template-row"]').first()).toBeVisible({
      timeout: 15_000
    })

    await page
      .getByRole('button', { name: /新建模板/ })
      .first()
      .click()
    await page
      .getByRole('button', { name: /(保存|确定|提交)/ })
      .first()
      .click()

    const fieldErrors = page.locator('[data-testid="login-template-field-error"]')
    await expect(fieldErrors).toHaveCount(2)

    await page.keyboard.press('Escape')
    expect(filterApiAborts(abortEvents)).toEqual([])
  })

  test('A3 /system/app 打开新建后 app_key/name 触发 field-error', async ({ page }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/app`)
    await expect(page.getByRole('heading', { name: /应用管理/ })).toBeVisible()
    await expect(page.locator('[data-testid="app-card"]').first()).toBeVisible({ timeout: 15_000 })

    await page
      .getByRole('button', { name: /(新建|创建)/ })
      .first()
      .click()
    await page
      .getByRole('button', { name: /(保存|确定|提交)/ })
      .first()
      .click()

    const fieldErrors = page.locator('[data-testid="app-field-error"]')
    await expect(fieldErrors.first()).toBeVisible()
    // 至少有 app_key + name 两个必填标记（页面还可能带 host/url 校验节点）
    expect(await fieldErrors.count()).toBeGreaterThanOrEqual(2)

    await page.keyboard.press('Escape')
    expect(filterApiAborts(abortEvents)).toEqual([])
  })

  test('A4 /system/menu-space 打开新建后 space_key/name 触发 field-error', async ({ page }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/menu-space`)
    await expect(page.getByRole('heading', { name: /菜单空间|导航空间|空间管理/ })).toBeVisible()
    await expect(page.locator('[data-testid="menu-space-card"]').first()).toBeVisible({
      timeout: 15_000
    })

    await page
      .getByRole('button', { name: /新建空间|新建|创建空间/ })
      .first()
      .click()
    await page
      .getByRole('button', { name: /(保存|确定|提交)/ })
      .first()
      .click()

    const fieldErrors = page.locator('[data-testid="menu-space-field-error"]')
    await expect(fieldErrors.first()).toBeVisible()
    expect(await fieldErrors.count()).toBeGreaterThanOrEqual(2)

    await page.keyboard.press('Escape')
    expect(filterApiAborts(abortEvents)).toEqual([])
  })
})
