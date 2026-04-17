import { expect, test, type Page } from '@playwright/test'
import { waitForDashboardReady } from '../support/high-risk'

const REAL_BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:5174'
const REAL_USERNAME = process.env.E2E_USERNAME || 'admin'
const REAL_PASSWORD = process.env.E2E_PASSWORD || 'admin123456'

type PageCheck = {
  path: string
  title: RegExp
  ready?: string
  inlineAlertSelector?: string
  verify?: (page: Page) => Promise<void>
}

const pageChecks: PageCheck[] = [
  {
    path: '/system/message',
    title: /消息发送/,
    ready: '[data-testid="send-status"]'
  },
  {
    path: '/system/message-template',
    title: /消息模板/,
    ready: '.message-template-shell'
  },
  {
    path: '/system/message-sender',
    title: /全局发送人|发送人/,
    ready: '.message-sender-shell'
  },
  {
    path: '/system/message-recipient-group',
    title: /接收组管理|接收组/,
    ready: '.message-group-shell'
  },
  {
    path: '/system/message-record',
    title: /消息发送记录/,
    ready: '.message-record-shell'
  },
  {
    path: '/collaboration/message',
    title: /协作空间消息发送/,
    ready: '[data-testid="send-status"][data-scope="collaboration"]'
  },
  {
    path: '/collaboration/workspaces',
    title: /协作空间管理/,
    ready: '.collaboration-toolbar-tip'
  },
  {
    path: '/collaboration/members',
    title: /协作空间成员/,
    verify: async (page: Page) => {
      await expect
        .poll(
          async () => {
            const tableVisible = await page
              .locator('.collaborationWorkspace-members-table')
              .first()
              .isVisible()
              .catch(() => false)
            const emptyVisible = await page
              .getByText(/您当前还未加入协作空间|请先加入协作空间/)
              .first()
              .isVisible()
              .catch(() => false)
            return tableVisible || emptyVisible
          },
          { timeout: 15_000 }
        )
        .toBe(true)
    }
  },
  {
    path: '/collaboration/roles',
    title: /协作空间角色与权限/,
    ready: 'text=当前协作空间角色管理'
  }
]

async function login(page: Page) {
  await page.goto(`${REAL_BASE_URL}/account/auth/login?redirect=%2Fdashboard%2Fconsole`)
  await page.locator('input[type="text"]').first().fill(REAL_USERNAME)
  await page.locator('input[type="password"]').first().fill(REAL_PASSWORD)
  await page.getByRole('button', { name: /登录/i }).click()
  await waitForDashboardReady(page)
}

test('workspace/collaboration 主路径真实可用', async ({ page }) => {
  test.skip(!REAL_USERNAME || !REAL_PASSWORD, '缺少真实后端登录凭据')

  const consoleErrors: string[] = []
  const apiFailures: Array<{ url: string; status: number }> = []

  page.on('console', (message) => {
    if (message.type() !== 'error') return
    const text = message.text()
    if (text.includes('ERR_ABORTED') || text.includes('net::ERR_ABORTED')) return
    consoleErrors.push(text)
  })

  page.on('response', (response) => {
    const url = response.url()
    if (!url.includes('/api/v1/')) return
    if (response.status() >= 400) {
      apiFailures.push({ url, status: response.status() })
    }
  })

  await login(page)
  consoleErrors.length = 0
  apiFailures.length = 0

  for (const pageCheck of pageChecks) {
    await page.goto(`${REAL_BASE_URL}${pageCheck.path}`)
    await page.waitForLoadState('networkidle')
    await expect(page.getByText(pageCheck.title).first()).toBeVisible({ timeout: 15_000 })
    if (pageCheck.verify) {
      await pageCheck.verify(page)
    } else if (pageCheck.ready) {
      await expect(page.locator(pageCheck.ready).first()).toBeVisible({ timeout: 15_000 })
    }
    if (pageCheck.inlineAlertSelector) {
      await expect(page.locator(pageCheck.inlineAlertSelector)).toHaveCount(0)
    }
  }

  expect(apiFailures).toEqual([])
  expect(consoleErrors).toEqual([])
})
