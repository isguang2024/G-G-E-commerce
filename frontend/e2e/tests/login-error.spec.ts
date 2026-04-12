import { expect, test } from '@playwright/test'

test('错误凭据登录后显示错误提示且不跳转', async ({ page }) => {
  await page.route('**/api/v1/auth/login', async (route) => {
    await route.fulfill({
      status: 401,
      contentType: 'application/json; charset=utf-8',
      body: JSON.stringify({
        code: 40101,
        message: '用户名或密码错误'
      })
    })
  })

  await page.goto('/account/auth/login')
  const usernameInput = page.locator('input[type="text"]').first()
  const passwordInput = page.locator('input[type="password"]').first()

  await expect(usernameInput).toBeVisible()
  await expect(passwordInput).toBeVisible()
  await usernameInput.fill('wrong-user')
  await passwordInput.fill('wrong-password')
  await page.getByRole('button', { name: /登录/i }).click()

  await expect(page).toHaveURL(/\/account\/auth\/login/)
  await expect(page.getByText('用户名或密码错误')).toBeVisible()
  await expect(usernameInput).toHaveValue('wrong-user')
  await expect(passwordInput).toHaveValue('wrong-password')
})
