import { expect, test } from '@playwright/test'

test('未登录访问后台页会跳转到登录页', async ({ page, context }) => {
  await context.clearCookies()
  await page.goto('/account/auth/login')
  await page.evaluate(() => {
    localStorage.clear()
    sessionStorage.clear()
  })

  await page.goto('/system/page')
  await page.waitForURL((url) => {
    const nextURL = new URL(url.toString())
    const redirect = nextURL.searchParams.get('redirect')
    const targetPath = nextURL.searchParams.get('target_path')
    return (
      nextURL.pathname === '/account/auth/login' &&
      (redirect === '/system/page' ||
        (targetPath === '/system/page' &&
          nextURL.searchParams.get('target_app_key') === 'platform-admin'))
    )
  })

  await expect(page.getByRole('button', { name: /登录/i })).toBeVisible()
})
