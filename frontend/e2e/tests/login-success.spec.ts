import { expect, test } from '@playwright/test'

function json(body: unknown) {
  return {
    status: 200,
    contentType: 'application/json; charset=utf-8',
    body: JSON.stringify(body)
  }
}

test('正确凭据登录后进入系统主页并写入会话', async ({ page }) => {
  page.on('console', (message) => {
    if (message.type() === 'error' || message.type() === 'warning') {
      console.log(`[login-success][console:${message.type()}]`, message.text())
    }
  })
  page.on('pageerror', (error) => {
    console.log('[login-success][pageerror]', error.message)
  })

  await page.route('**/api/v1/**', async (route) => {
    const url = new URL(route.request().url())

    if (url.pathname === '/api/v1/auth/login') {
      await route.fulfill(
        json({
          access_token: 'mock-access-token',
          refresh_token: 'mock-refresh-token',
          user: {
            id: 'user-1',
            username: 'tester',
            nickname: '测试用户',
            email: 'tester@example.com',
            is_super_admin: true,
            actions: ['system.page.view', 'system.role.manage'],
            roles: ['admin'],
            current_auth_workspace_id: 'ws-personal',
            current_auth_workspace_type: 'personal',
            current_collaboration_workspace_id: '',
            collaboration_workspace_id: ''
          }
        })
      )
      return
    }

    if (url.pathname === '/api/v1/system/menu-spaces') {
      await route.fulfill(
        json({
          records: [
            {
              space_key: 'default',
              menu_space_key: 'default',
              menuSpaceKey: 'default',
              name: '默认空间',
              status: 'normal',
              is_default: true,
              default_home_path: '/dashboard/console',
              meta: {
                space_type: 'default'
              }
            }
          ],
          total: 1
        })
      )
      return
    }

    if (url.pathname === '/api/v1/system/apps/current') {
      await route.fulfill(
        json({
          app: {
            app_key: 'platform-admin',
            name: '平台管理',
            auth_mode: 'shared_cookie',
            frontend_entry_url: '',
            backend_entry_url: '',
            health_check_url: '',
            capabilities: {},
            meta: {}
          },
          resolved_by: 'query',
          request_host: '127.0.0.1:4174'
        })
      )
      return
    }

    if (url.pathname === '/api/v1/system/apps') {
      await route.fulfill(
        json({
          records: [
            {
              id: 'app-platform-admin',
              app_key: 'platform-admin',
              name: '平台管理',
              description: '平台管理后台',
              auth_mode: 'shared_cookie',
              frontend_entry_url: '',
              backend_entry_url: '',
              health_check_url: '',
              manifest_url: '',
              runtime_version: '5.0.0',
              capabilities: {
                centralized_login: false
              },
              status: 'active',
              is_default: true,
              meta: {}
            }
          ],
          total: 1
        })
      )
      return
    }

    if (url.pathname === '/api/v1/auth/me') {
      await route.fulfill(
        json({
          id: 'user-1',
          username: 'tester',
          nickname: '测试用户',
          email: 'tester@example.com',
          is_super_admin: true,
          actions: ['system.page.view', 'system.role.manage'],
          roles: [{ code: 'admin' }],
          current_auth_workspace_id: 'ws-personal',
          current_auth_workspace_type: 'personal',
          current_collaboration_workspace_id: '',
          collaboration_workspace_id: ''
        })
      )
      return
    }

    if (url.pathname === '/api/v1/auth/logout') {
      await route.fulfill(json({ success: true }))
      return
    }

    if (url.pathname === '/api/v1/system/menu-space-host-bindings') {
      await route.fulfill(json({ records: [], total: 0 }))
      return
    }

    if (url.pathname === '/api/v1/system/menu-spaces/current') {
      await route.fulfill(
        json({
          space: {
            space_key: 'default',
            name: '默认空间',
            status: 'normal',
            is_default: true,
            default_home_path: '/dashboard/console',
            meta: {
              space_type: 'default'
            }
          },
          resolved_by: 'default',
          request_host: '127.0.0.1:4174',
          access_granted: true
        })
      )
      return
    }

    if (url.pathname === '/api/v1/collaboration-workspaces/mine') {
      await route.fulfill(json({ records: [], total: 0 }))
      return
    }

    if (url.pathname === '/api/v1/workspaces/my') {
      await route.fulfill(
        json({
          records: [
            {
              id: 'ws-personal',
              workspace_type: 'personal',
              name: '个人空间',
              code: 'personal',
              status: 'active'
            }
          ],
          total: 1
        })
      )
      return
    }

    if (url.pathname === '/api/v1/messages/inbox/summary') {
      await route.fulfill(
        json({
          unread_total: 0,
          notice_count: 0,
          message_count: 0,
          todo_count: 0
        })
      )
      return
    }

    if (url.pathname === '/api/v1/runtime/navigation') {
      await route.fulfill(
        json({
          current_app: {
            app: {
              app_key: 'platform-admin',
              name: '平台管理',
              auth_mode: 'shared_cookie',
              capabilities: {},
              meta: {}
            },
            resolved_by: 'query',
            request_host: '127.0.0.1:4174'
          },
          current_space: {
            space: {
              space_key: 'default',
              menu_space_key: 'default',
              menuSpaceKey: 'default',
              name: '默认空间',
              status: 'normal',
              is_default: true,
              default_home_path: '/dashboard/console',
              meta: {
                space_type: 'default'
              }
            },
            resolved_by: 'default',
            request_host: '127.0.0.1:4174',
            access_granted: true
          },
          context: {
            app_key: 'platform-admin',
            space_key: 'default',
            menu_space_key: 'default',
            menuSpaceKey: 'default'
          },
          menu_tree: [
            {
              id: 'menu-dashboard',
              name: 'DashboardRoot',
              path: '/dashboard',
              sort_order: 1,
              meta: {
                title: '工作台',
                isEnable: true
              },
              children: [
                {
                  id: 'menu-console',
                  name: 'Console',
                  path: 'console',
                  component: 'dashboard/console/index',
                  sort_order: 1,
                  meta: {
                    title: '后台工作台',
                    isEnable: true
                  }
                }
              ]
            }
          ],
          entry_routes: [],
          managed_pages: [],
          version_stamp: 'e2e-login-success'
        })
      )
      return
    }

    await route.continue()
  })

  await page.goto('/account/auth/login?redirect=%2Fdashboard%2Fconsole')

  const usernameInput = page.locator('input[type="text"]').first()
  const passwordInput = page.locator('input[type="password"]').first()
  await usernameInput.fill('tester')
  await passwordInput.fill('correct-password')
  await page.getByRole('button', { name: /登录/i }).click()

  await page.waitForURL((url) => new URL(url.toString()).pathname === '/dashboard/console')
  await expect(page.locator('#app-sidebar').getByText('工作台', { exact: true })).toBeVisible()
  await expect(page.getByRole('heading', { name: '后台工作台' })).toBeVisible()

  await expect
    .poll(async () => {
      return page.evaluate(() => {
        const entries = Object.entries(localStorage)
        const matchedEntry = entries.find(([key]) => key.endsWith('-user'))
        return matchedEntry ? matchedEntry[1] : null
      })
    })
    .not.toBeNull()

  const persistedUserEntry = await page.evaluate(() => {
    const entries = Object.entries(localStorage)
    const matchedEntry = entries.find(([key]) => key.endsWith('-user'))
    return matchedEntry ? { key: matchedEntry[0], value: matchedEntry[1] } : null
  })

  expect(persistedUserEntry?.key).toMatch(/^sys-v.+-user$/)
  expect(persistedUserEntry?.value).toContain('mock-access-token')
  expect(persistedUserEntry?.value).toContain('tester@example.com')

  await expect(page.locator('.el-loading-mask.is-fullscreen')).toHaveCount(0)
  await page.locator('img[alt="avatar"]').click()
  await expect(page.getByText('tester', { exact: true })).toBeVisible()
  await expect(page.getByText('tester@example.com')).toBeVisible()
})
