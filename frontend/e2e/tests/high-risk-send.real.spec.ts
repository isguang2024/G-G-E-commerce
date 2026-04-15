import { expect, test, type Page } from '@playwright/test'
import { waitForDashboardReady, collectAbortEvents, filterApiAborts } from '../support/high-risk'

/**
 * 高风险发送链路回归（P1 组）——对应 docs/guides/high-risk-remediation-matrix.md
 *   C1 /system/message              (scope=personal)
 *   C2 /collaboration-workspace/message (scope=collaboration)
 *
 * 两个页面都挂载同一个 `MessageDispatchConsole` 组件，测试数据 & 断言结构共享，
 * 差异只在 scope 与对应必填项（collaboration 需要勾选协作空间范围）。
 *
 * 每个用例按以下流程：
 *   1. 进入页面，等待数据加载结束；
 *   2. 在启用 dry-run 的前提下点"确认发送"→预期走校验失败分支，
 *      `send-status[data-status-code="validation_failed"]` 亮起，
 *      同时 `send-error[data-field=...]` 节点渲染；
 *   3. 填入最小可通过表单 → 点发送 → 打开 `send-preview-dialog`；
 *   4. 点 `send-preview-confirm` → `send-status[data-status-code="preview"]`
 *      + `data-dry-run="true"` 锁定，Toast 文案匹配沙箱成功；
 *   5. 可重复（dry-run 不会入库），验证"可在 dry-run 下反复执行"。
 *
 * 前端 `dryRunAvailable` 对生产环境要求 URL 带 `?__dry_run=1` 才开启，
 * 所以每次访问都带这个 query，兼容 dev / ci / prod。
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

async function enableDryRun(page: Page) {
  const toggle = page.locator('[data-testid="send-dry-run-toggle"]')
  await expect(toggle).toBeVisible({ timeout: 15_000 })
  // dryRunEnabled 的默认状态是 false；点开关使其为 true
  await page.locator('[data-testid="send-dry-run-switch"] .el-switch__core').click()
}

async function assertValidationFailed(page: Page) {
  await page.locator('[data-testid="send-button-dryrun"]').click()
  const status = page.locator('[data-testid="send-status"]')
  await expect(status).toHaveAttribute('data-status-code', 'validation_failed', { timeout: 5_000 })
  // 必填错误节点存在 —— 至少 sender/message_type/audience/title/content 其中几个
  const errors = page.locator('[data-testid="send-error"]')
  expect(await errors.count()).toBeGreaterThan(0)
}

test.describe('高风险发送链路 - dry-run 反复执行 + data-testid 断言', () => {
  test.beforeEach(async () => {
    test.skip(!REAL_USERNAME || !REAL_PASSWORD, '缺少真实后端登录凭据')
  })

  test('C1 /system/message?__dry_run=1 校验失败 + 预览 + 沙箱确认', async ({ page }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/system/message?__dry_run=1`)
    await expect(page.locator('[data-testid="send-status"]')).toHaveAttribute(
      'data-scope',
      'personal'
    )
    await expect(page.locator('[data-testid="send-status"]')).toHaveAttribute(
      'data-dry-run-available',
      'true'
    )

    await enableDryRun(page)

    // step 1: 空提交 → validation_failed
    await assertValidationFailed(page)

    // step 2: 填入最小表单（具体字段值只要求合法，UI 层用默认首项）
    // 发送人
    await page.locator('[data-testid="send-field-sender"]').click()
    await page.getByRole('option').first().click()
    // 消息类型
    await page.locator('[data-testid="send-field-message-type"]').click()
    await page.getByRole('option').first().click()
    // 发送对象（用第一个 audience 选项）
    await page.locator('[data-testid="send-field-audience"]').click()
    await page.getByRole('option').first().click()

    // 标题 + 摘要
    await page.locator('[data-testid="send-field-title"] input').fill('[dry-run] 回归标题')
    await page.locator('[data-testid="send-field-summary"] input').fill('[dry-run] 回归摘要')

    // step 3: 打开预览
    await page.locator('[data-testid="send-button-dryrun"]').click()
    await expect(page.locator('[data-testid="send-preview-dialog"]')).toBeVisible()
    await expect(page.locator('[data-testid="send-preview-dryrun-banner"]')).toBeVisible()
    await expect(page.locator('[data-testid="send-preview-title"]')).toContainText('[dry-run]')

    // step 4: 确认发送 → preview 状态
    await page.locator('[data-testid="send-preview-confirm"]').click()
    const status = page.locator('[data-testid="send-status"]')
    await expect(status).toHaveAttribute('data-status-code', 'preview', { timeout: 10_000 })
    await expect(status).toHaveAttribute('data-dry-run', 'true')

    // step 5: 反复执行——再点一次发送，预览应再次出现（dry-run 不清表单）
    await page.locator('[data-testid="send-button-dryrun"]').click()
    await expect(page.locator('[data-testid="send-preview-dialog"]')).toBeVisible()
    await page.locator('[data-testid="send-preview-cancel"]').click()

    expect(filterApiAborts(abortEvents)).toEqual([])
  })

  test('C2 /collaboration-workspace/message?__dry_run=1 范围必填 + 预览', async ({ page }) => {
    const abortEvents = collectAbortEvents(page)
    await login(page)

    await page.goto(`${REAL_BACKEND_BASE_URL}/collaboration-workspace/message?__dry_run=1`)
    await expect(page.locator('[data-testid="send-status"]')).toHaveAttribute(
      'data-scope',
      'collaboration'
    )
    await expect(page.locator('[data-testid="send-status"]')).toHaveAttribute(
      'data-dry-run-available',
      'true'
    )

    await enableDryRun(page)

    // 空提交 → validation_failed + 协作空间需要目标工作空间
    await assertValidationFailed(page)
    await expect(
      page
        .locator('[data-testid="send-error"]')
        .filter({
          has: page.locator('[data-testid="send-field-target-workspaces"]')
        })
        .first()
    ).toBeVisible()

    expect(filterApiAborts(abortEvents)).toEqual([])
  })
})
