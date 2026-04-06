const fs = require('fs');
const path = require('path');
const { chromium } = require('playwright');

const BASE = 'http://localhost:5173';
const API = 'http://localhost:8080';
const VERSION = '3.0.1';

const routes = [
  { key: 'dashboard_console', path: '/dashboard/console' },
  { key: 'workspace_inbox', path: '/workspace/inbox' },
  { key: 'system_user', path: '/system/user' },
  { key: 'system_role', path: '/system/role' },
  { key: 'system_action_permission', path: '/system/action-permission' },
  { key: 'system_feature_package', path: '/system/feature-package' },
  { key: 'system_page', path: '/system/page' },
  { key: 'system_menu', path: '/system/menu' },
  { key: 'system_menu_space', path: '/system/menu-space' },
  { key: 'system_api_endpoint', path: '/system/api-endpoint' },
  { key: 'system_message', path: '/system/message' },
  { key: 'team_team', path: '/team/team' },
  { key: 'team_roles_permissions', path: '/system/team-roles-permissions' },
  { key: 'collaboration_workspace_message', path: '/team/message' }
];

async function login() {
  const creds = [
    { username: 'admin', password: 'admin123456' },
    { username: 'platform_admin_demo', password: 'Demo123456' },
    { username: 'admin@a.com', password: 'admin123456' }
  ];
  let lastErr = null;
  for (const c of creds) {
    try {
      const res = await fetch(`${API}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(c)
      });
      const text = await res.text();
      let json;
      try { json = JSON.parse(text); } catch (_) { json = { raw: text }; }
      if (res.ok && json && (json.access_token || json.data?.access_token)) {
        const payload = json.access_token ? json : json.data;
        return { ...payload, account: c.username };
      }
      lastErr = { account: c.username, status: res.status, body: json };
    } catch (err) {
      lastErr = { account: c.username, err: String(err) };
    }
  }
  throw new Error(`login failed: ${JSON.stringify(lastErr)}`);
}

(async () => {
  const auth = await login();
  const user = auth.user || {};

  const userStore = {
    language: 'zh',
    isLogin: true,
    isLock: false,
    lockPassword: '',
    info: {
      ...user,
      userId: user.id,
      userName: user.username || user.email || '',
      avatar: user.avatar_url || '',
      roles: user.is_super_admin ? ['R_SUPER'] : ['R_USER'],
      buttons: [],
      actions: user.actions || []
    },
    searchHistory: [],
    accessToken: auth.access_token,
    refreshToken: auth.refresh_token || ''
  };

  const settingStore = {
    menuType: 'left',
    menuOpenWidth: 230,
    menuOpen: true,
    uniqueOpened: false,
    dualMenuShowText: false,
    showNprogress: false
  };

  const context = await chromium.launchPersistentContext(path.resolve('.codex-tmp/pw-profile'), {
    headless: true,
    viewport: { width: 1440, height: 920 }
  });
  const page = context.pages()[0] || await context.newPage();

  const consoleErrors = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      consoleErrors.push({ url: page.url(), text: msg.text() });
    }
  });

  await page.goto(BASE, { waitUntil: 'domcontentloaded' });
  await page.evaluate(({ userStore, settingStore, version }) => {
    localStorage.setItem('sys-version', version);
    localStorage.setItem(`sys-v${version}-user`, JSON.stringify(userStore));
    localStorage.setItem(`sys-v${version}-setting`, JSON.stringify(settingStore));
  }, { userStore, settingStore, version: VERSION });

  const menuTree = await page.evaluate(async ({ token }) => {
    const res = await fetch('/api/v1/menus/tree?space_key=default', {
      headers: { Authorization: token.startsWith('Bearer ') ? token : `Bearer ${token}` }
    });
    const json = await res.json();
    return json;
  }, { token: auth.access_token });

  const runtimePages = await page.evaluate(async ({ token }) => {
    const res = await fetch('/api/v1/pages/runtime?space_key=default', {
      headers: { Authorization: token.startsWith('Bearer ') ? token : `Bearer ${token}` }
    });
    const json = await res.json();
    return json;
  }, { token: auth.access_token });

  const inspections = [];

  for (const r of routes) {
    const target = `${BASE}/#${r.path}`;
    await page.goto(target, { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(1200);

    const info = await page.evaluate(() => {
      const heading = (document.querySelector('h1, h2, h3')?.textContent || '').trim();
      const breadcrumbs = Array.from(document.querySelectorAll('nav[aria-label="breadcrumb"] li')).map((el) => (el.textContent || '').trim()).filter(Boolean);
      const leftMenus = Array.from(document.querySelectorAll('.layout-sidebar .el-menu > li')).map((el) => (el.textContent || '').trim().split(/\s+/).join(' ')).filter(Boolean);
      const buttons = Array.from(document.querySelectorAll('button')).map((el) => (el.textContent || '').trim()).filter(Boolean).slice(0, 12);
      const tabs = Array.from(document.querySelectorAll('.el-tabs__item')).map((el) => (el.textContent || '').trim()).filter(Boolean).slice(0, 12);
      const hasTable = !!document.querySelector('.el-table, table');
      const hasForm = !!document.querySelector('form, .el-form');
      const hasTree = !!document.querySelector('.el-tree');
      const hasDrawer = !!document.querySelector('.el-drawer');
      const hasDialog = !!document.querySelector('.el-dialog');
      const mainText = (document.querySelector('#app-content')?.textContent || '').replace(/\s+/g, ' ').trim().slice(0, 200);
      const title = document.title;
      return { heading, breadcrumbs, leftMenus, buttons, tabs, hasTable, hasForm, hasTree, hasDrawer, hasDialog, mainText, title };
    });

    inspections.push({
      key: r.key,
      route: r.path,
      landed: page.url().replace(BASE, ''),
      ...info
    });
  }

  const output = {
    account: auth.account,
    menuTree,
    runtimePagesCount: Array.isArray(runtimePages?.data?.records) ? runtimePages.data.records.length : (Array.isArray(runtimePages?.data) ? runtimePages.data.length : null),
    inspections,
    consoleErrors: consoleErrors.slice(0, 100)
  };

  const outPath = path.resolve('.codex-tmp/product-audit.json');
  fs.writeFileSync(outPath, JSON.stringify(output, null, 2), 'utf8');
  console.log(outPath);

  await context.close();
})();

