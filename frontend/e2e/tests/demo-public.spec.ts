import { expect, test } from '@playwright/test'

function json(body: unknown) {
  return {
    status: 200,
    contentType: 'application/json; charset=utf-8',
    body: JSON.stringify(body)
  }
}

test('未登录访问 Demo 公开页不会跳转到登录页', async ({ page }) => {
  await page.route('**/api/v1/pages/runtime/public**', async (route) => {
    await route.fulfill(
      json({
        records: [
          {
            id: 'page-demo-lab',
            app_key: 'demo-app',
            page_key: 'demo.lab',
            name: 'Demo 实验室',
            route_name: 'DemoLab',
            route_path: '/demo/lab',
            component: 'demo/lab/index',
            page_type: 'inner',
            source: 'manual',
            sort_order: 1,
            parent_menu_id: '',
            parent_page_key: '',
            active_menu_path: '',
            breadcrumb_mode: 'inherit_menu',
            access_mode: 'public',
            keep_alive: false,
            is_full_page: false,
            status: 'normal',
            meta: {}
          }
        ],
        total: 1
      })
    )
  })

  await page.goto('/demo/lab')

  await page.waitForURL((url) => new URL(url.toString()).pathname === '/demo/lab')
  await expect(page.getByText('Demo App 验证页', { exact: true })).toBeVisible()
  await expect(page.getByText('当前 APP：demo-app', { exact: true })).toBeVisible()
  await expect(page).not.toHaveURL(/\/account\/auth\/login/)
})
