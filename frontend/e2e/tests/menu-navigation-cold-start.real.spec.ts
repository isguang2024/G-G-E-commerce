import { expect, type Page, test } from '@playwright/test'

const REAL_BACKEND_BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:5174'
const REAL_USERNAME = process.env.E2E_USERNAME || 'admin'
const REAL_PASSWORD = process.env.E2E_PASSWORD || 'admin123456'

function menuActionable(page: Page, label: string) {
  return page
    .locator('.layout-sidebar')
    .locator(`xpath=.//*[normalize-space(text())="${label}"]/ancestor::*[
      contains(concat(" ", normalize-space(@class), " "), " el-menu-item ")
      or contains(concat(" ", normalize-space(@class), " "), " el-sub-menu__title ")
    ][1]`)
    .first()
}

async function clickFirstVisibleMenu(page: Page, labels: string[]) {

  for (const label of labels) {
    const locator = menuActionable(page, label)
    if (await locator.count()) {
      await locator.click()
      return label
    }
  }

  throw new Error(`未找到菜单项: ${labels.join(' / ')}`)
}

async function expandSidebarGroup(page: Page, label: string) {
  const locator = menuActionable(page, label)
  if (!(await locator.count())) {
    throw new Error(`未找到分组菜单: ${label}`)
  }
  await locator.click()
}

async function expectNoRouteChunkErrors(page: Page, consoleErrors: string[]) {
  await expect
    .poll(
      () =>
        consoleErrors.filter(
          (message) =>
            message.includes('Outdated Optimize Dep') ||
            message.includes('Failed to fetch dynamically imported module') ||
            message.includes('Pre-transform error')
        ),
      { timeout: 5_000 }
    )
    .toEqual([])
}

test('冷启动后左侧菜单首次切换不会停留在旧页', async ({ page }) => {
  test.skip(!REAL_USERNAME || !REAL_PASSWORD, '缺少真实后端登录凭据')

  const consoleErrors: string[] = []

  page.on('console', (message) => {
    if (message.type() === 'error') {
      consoleErrors.push(message.text())
    }
  })

  await page.goto(`${REAL_BACKEND_BASE_URL}/account/auth/login?redirect=%2Fdashboard%2Fconsole`)
  await page.locator('input[type="text"]').first().fill(REAL_USERNAME)
  await page.locator('input[type="password"]').first().fill(REAL_PASSWORD)
  await page.getByRole('button', { name: /登录/i }).click()

  await page.waitForURL((url) => new URL(url.toString()).pathname === '/dashboard/console')
  await expect(page.getByRole('heading', { name: '后台工作台' })).toBeVisible()

  await expandSidebarGroup(page, '系统管理')
  await expandSidebarGroup(page, '导航与界面')
  await clickFirstVisibleMenu(page, ['应用管理'])
  await page.waitForURL((url) => new URL(url.toString()).pathname === '/system/app')
  await expect(page.getByRole('heading', { name: '应用管理' })).toBeVisible()
  await expect(page.locator('.app-manage-inline-alert')).toHaveCount(0)

  await clickFirstVisibleMenu(page, ['受管页面', '页面管理'])
  await page.waitForURL((url) => new URL(url.toString()).pathname === '/system/page')
  await expect(page.getByRole('heading', { name: '受管页面' })).toBeVisible()
  await expect(page.locator('.page-inline-alert')).toHaveCount(0)

  await expectNoRouteChunkErrors(page, consoleErrors)
})
