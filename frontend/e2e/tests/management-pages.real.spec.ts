import { expect, test } from '@playwright/test'

const REAL_BACKEND_BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:5174'
const REAL_USERNAME = process.env.E2E_USERNAME || 'admin'
const REAL_PASSWORD = process.env.E2E_PASSWORD || 'admin123456'

test('应用管理页和受管页面可正常加载', async ({ page }) => {
  test.skip(!REAL_USERNAME || !REAL_PASSWORD, '缺少真实后端登录凭据')

  const consoleErrors: string[] = []
  const relevantApiFailures: Array<{ url: string; status: number }> = []

  page.on('console', (message) => {
    if (message.type() === 'error') {
      consoleErrors.push(message.text())
    }
  })

  page.on('response', (response) => {
    const url = response.url()
    if (!url.includes('/api/v1/')) return
    if (
      url.includes('/api/v1/system/apps') ||
      url.includes('/api/v1/system/apps/current') ||
      url.includes('/api/v1/system/app-host-bindings') ||
      url.includes('/api/v1/system/menu-spaces') ||
      url.includes('/api/v1/pages?') ||
      url.includes('/api/v1/pages/menu-options')
    ) {
      if (response.status() >= 400) {
        relevantApiFailures.push({ url, status: response.status() })
      }
    }
  })

  await page.goto(`${REAL_BACKEND_BASE_URL}/account/auth/login?redirect=%2Fdashboard%2Fconsole`)
  await page.locator('input[type="text"]').first().fill(REAL_USERNAME)
  await page.locator('input[type="password"]').first().fill(REAL_PASSWORD)
  await page.getByRole('button', { name: /登录/i }).click()

  await page.waitForURL((url) => new URL(url.toString()).pathname === '/dashboard/console')
  await expect(page.getByRole('heading', { name: '后台工作台' })).toBeVisible()

  const appListResponsePromise = page.waitForResponse((response) => {
    const url = response.url()
    return url.includes('/api/v1/system/apps') && !url.includes('/current')
  })
  const currentAppResponsePromise = page.waitForResponse((response) =>
    response.url().includes('/api/v1/system/apps/current')
  )

  await page.goto(`${REAL_BACKEND_BASE_URL}/system/app`)
  await expect(page.getByRole('heading', { name: '应用管理' })).toBeVisible()
  await expect(page.locator('.app-manage-inline-alert')).toHaveCount(0)
  expect((await appListResponsePromise).status()).toBe(200)
  expect((await currentAppResponsePromise).status()).toBe(200)
  await expect(page.getByText('App 列表', { exact: true })).toBeVisible()

  const pageListResponsePromise = page.waitForResponse((response) =>
    response.url().includes('/api/v1/pages?')
  )
  const pageMenuOptionsResponsePromise = page.waitForResponse((response) =>
    response.url().includes('/api/v1/pages/menu-options')
  )

  await page.goto(`${REAL_BACKEND_BASE_URL}/system/page`)
  await expect(page.getByRole('heading', { name: '受管页面' })).toBeVisible()
  await expect(page.locator('.page-inline-alert')).toHaveCount(0)
  expect((await pageListResponsePromise).status()).toBe(200)
  expect((await pageMenuOptionsResponsePromise).status()).toBe(200)
  await expect(page.getByText('页面挂载关系、访问方式和父链路在这里集中治理', { exact: false })).toBeVisible()

  expect(relevantApiFailures).toEqual([])
  expect(consoleErrors).toEqual([])
})
