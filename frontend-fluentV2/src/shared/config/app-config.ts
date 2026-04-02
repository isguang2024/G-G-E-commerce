export const appConfig = {
  appName: 'G&G ERP Shell',
  appSubtitle: '业务工作台',
  defaultRoute: '/welcome',
  defaultSpaceKey: 'default',
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '',
  devProxyTarget: import.meta.env.VITE_API_PROXY_TARGET || 'http://127.0.0.1:8080',
} as const
