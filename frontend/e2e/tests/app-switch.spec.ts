import { expect, test } from '@playwright/test'

function json(body: unknown) {
  return {
    status: 200,
    contentType: 'application/json; charset=utf-8',
    body: JSON.stringify(body)
  }
}

function resolveAppDefinition(appKey: string) {
  if (appKey === 'demo-app') {
    return {
      id: 'app-demo-app',
      app_key: 'demo-app',
      name: 'Demo 应用',
      description: 'Demo 应用壳',
      auth_mode: 'shared_cookie',
      frontend_entry_url: '/demo/lab',
      backend_entry_url: '',
      health_check_url: '',
      manifest_url: '',
      runtime_version: '5.0.0',
      capabilities: {
        integration: {
          supports_app_switch: true
        }
      },
      status: 'active',
      is_default: false,
      meta: {}
    }
  }

  return {
    id: 'app-platform-admin',
    app_key: 'platform-admin',
    name: '平台管理',
    description: '平台管理后台',
    auth_mode: 'shared_cookie',
    frontend_entry_url: '/dashboard/console',
    backend_entry_url: '',
    health_check_url: '',
    manifest_url: '',
    runtime_version: '5.0.0',
    capabilities: {
      integration: {
        supports_app_switch: true
      }
    },
    status: 'active',
    is_default: true,
    meta: {}
  }
}

function resolveSpaceDefinition(appKey: string) {
  if (appKey === 'demo-app') {
    return {
      space_key: 'demo-default',
      menu_space_key: 'demo-default',
      menuSpaceKey: 'demo-default',
      name: 'Demo 空间',
      status: 'normal',
      is_default: true,
      default_home_path: '/demo/lab',
      meta: {
        space_type: 'default'
      }
    }
  }

  return {
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
}

function resolveNavigationManifest(appKey: string) {
  if (appKey === 'demo-app') {
    return {
      current_app: {
        app: resolveAppDefinition(appKey),
        resolved_by: 'query',
        request_host: '127.0.0.1:4174'
      },
      current_space: {
        space: resolveSpaceDefinition(appKey),
        resolved_by: 'default',
        request_host: '127.0.0.1:4174',
        access_granted: true
      },
      context: {
        app_key: 'demo-app',
        space_key: 'demo-default',
        menu_space_key: 'demo-default',
        menuSpaceKey: 'demo-default'
      },
      menu_tree: [
        {
          id: 'menu-demo-root',
          name: 'DemoRoot',
          path: '/demo',
          sort_order: 1,
          meta: {
            title: 'Demo 应用',
            isEnable: true
          },
          children: [
            {
              id: 'menu-demo-lab',
              name: 'DemoLab',
              path: 'lab',
              component: 'demo/lab/index',
              sort_order: 1,
              meta: {
                title: 'Demo 实验室',
                isEnable: true
              }
            }
          ]
        }
      ],
      entry_routes: [],
      managed_pages: [],
      version_stamp: `e2e-app-switch-${appKey}`
    }
  }

  return {
    current_app: {
      app: resolveAppDefinition(appKey),
      resolved_by: 'query',
      request_host: '127.0.0.1:4174'
    },
    current_space: {
      space: resolveSpaceDefinition(appKey),
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
    version_stamp: `e2e-app-switch-${appKey}`
  }
}

test('已登录后切换 App 会刷新菜单并更新上下文', async ({ page }) => {
  await page.route('**/api/v1/**', async (route) => {
    const url = new URL(route.request().url())
    const appKey = `${url.searchParams.get('app_key') || ''}`.trim() || 'platform-admin'

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

    if (url.pathname === '/api/v1/system/apps') {
      await route.fulfill(
        json({
          records: [resolveAppDefinition('platform-admin'), resolveAppDefinition('demo-app')],
          total: 2
        })
      )
      return
    }

    if (url.pathname === '/api/v1/system/apps/current') {
      await route.fulfill(
        json({
          app: resolveAppDefinition(appKey),
          resolved_by: 'query',
          request_host: '127.0.0.1:4174'
        })
      )
      return
    }

    if (url.pathname === '/api/v1/system/menu-spaces') {
      await route.fulfill(
        json({
          records: [resolveSpaceDefinition(appKey)],
          total: 1
        })
      )
      return
    }

    if (url.pathname === '/api/v1/system/menu-space-host-bindings') {
      await route.fulfill(json({ records: [], total: 0 }))
      return
    }

    if (url.pathname === '/api/v1/system/menu-spaces/current') {
      await route.fulfill(
        json({
          space: resolveSpaceDefinition(appKey),
          resolved_by: 'default',
          request_host: '127.0.0.1:4174',
          access_granted: true
        })
      )
      return
    }

    if (url.pathname === '/api/v1/runtime/navigation') {
      await route.fulfill(json(resolveNavigationManifest(appKey)))
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

    if (url.pathname === '/api/v1/collaboration-workspaces/mine') {
      await route.fulfill(json({ records: [], total: 0 }))
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

    if (url.pathname === '/api/v1/auth/logout') {
      await route.fulfill(json({ success: true }))
      return
    }

    await route.continue()
  })

  await page.goto('/account/auth/login?redirect=%2Fdashboard%2Fconsole')
  await page.locator('input[type="text"]').first().fill('tester')
  await page.locator('input[type="password"]').first().fill('correct-password')
  await page.getByRole('button', { name: /登录/i }).click()

  await page.waitForURL((url) => new URL(url.toString()).pathname === '/dashboard/console')
  await expect(page.getByRole('heading', { name: '后台工作台' })).toBeVisible()
  await expect(page.locator('.app-switcher')).toBeVisible()

  await page.locator('.app-switcher .el-select__wrapper').click()
  await page.getByRole('option', { name: /Demo 应用/ }).click()

  await page.waitForURL((url) => new URL(url.toString()).pathname === '/demo/lab')
  await expect(page.locator('#app-sidebar').getByText('Demo 应用', { exact: true })).toBeVisible()
  await expect(page.locator('#app-sidebar').getByText('Demo 实验室', { exact: true })).toBeVisible()
  await expect(page.getByText('Demo App 验证页', { exact: true })).toBeVisible()
  await expect(page.getByText('当前 APP：demo-app', { exact: true })).toBeVisible()

  const persistedAppContextEntry = await page.evaluate(() => {
    const entries = Object.entries(localStorage)
    const matchedEntry = entries.find(([key]) => key.endsWith('-appContextStore'))
    return matchedEntry ? { key: matchedEntry[0], value: matchedEntry[1] } : null
  })

  expect(persistedAppContextEntry?.key).toMatch(/^sys-v.+-appContextStore$/)
  expect(persistedAppContextEntry?.value).toContain('demo-app')
})
