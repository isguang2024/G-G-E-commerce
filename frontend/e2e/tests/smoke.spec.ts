import { expect, test } from '@playwright/test'

test('公开登录页可以打开', async ({ page }) => {
  await page.goto('/account/auth/login')

  await expect(page).toHaveURL(/\/account\/auth\/login/)
  await expect(page.getByRole('button', { name: /登录/i })).toBeVisible()
})
