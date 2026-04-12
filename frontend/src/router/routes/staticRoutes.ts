import { AppRouteRecordRaw } from '@/utils/router'

/**
 * 静态路由配置（不需要权限就能访问的路由）
 *
 * 属性说明：
 * isHideTab: true 表示不在标签页中显示
 *
 * 注意事项：
 * 1、path、name 不要和动态路由冲突，否则会导致路由冲突无法访问
 * 2、静态路由不管是否登录都可以访问
 */
export const staticRoutes: AppRouteRecordRaw[] = [
  // @compat-status: transition 旧认证路径仍需重定向到 account-portal 新入口，待外链与回跳全部切换后再移除。
  // 不需要登录就能访问的路由示例
  // {
  //   path: '/welcome',
  //   name: 'WelcomeStatic',
  //   component: () => import('@views/dashboard/console/index.vue'),
  //   meta: { title: 'menus.dashboard.title' }
  // },
  // 旧路径兼容：重定向到 account-portal 下的新入口
  {
    path: '/auth/login',
    redirect: '/account/auth/login',
    meta: { title: 'menus.login.title', isHideTab: true }
  },
  {
    path: '/auth/register',
    redirect: '/account/auth/register',
    meta: { title: 'menus.register.title', isHideTab: true }
  },
  {
    path: '/auth/forget-password',
    redirect: '/account/auth/forget-password',
    meta: { title: 'menus.forgetPassword.title', isHideTab: true }
  },
  {
    path: '/auth/callback',
    redirect: '/account/auth/callback',
    meta: { title: 'auth-callback', isHideTab: true }
  },
  {
    path: '/account/auth/login',
    name: 'AccountPortalLogin',
    component: () => import('@views/auth/login/index.vue'),
    meta: { title: 'menus.login.title', isHideTab: true, appKey: 'account-portal' }
  },
  {
    path: '/account/auth/register',
    name: 'AccountPortalRegister',
    component: () => import('@views/auth/register/index.vue'),
    meta: { title: 'menus.register.title', isHideTab: true, appKey: 'account-portal' }
  },
  {
    path: '/account/auth/forget-password',
    name: 'AccountPortalForgetPassword',
    component: () => import('@views/auth/forget-password/index.vue'),
    meta: { title: 'menus.forgetPassword.title', isHideTab: true, appKey: 'account-portal' }
  },
  {
    path: '/account/auth/callback',
    name: 'AuthCallback',
    component: () => import('@views/auth/callback/index.vue'),
    meta: { title: 'auth-callback', isHideTab: true, appKey: 'account-portal' }
  },
  {
    path: '/403',
    name: 'Exception403',
    component: () => import('@views/exception/403/index.vue'),
    meta: { title: '403', isHideTab: true }
  },
  {
    path: '/404',
    name: 'Exception404',
    component: () => import('@views/exception/404/index.vue'),
    meta: { title: '404', isHideTab: true }
  },
  {
    path: '/:pathMatch(.*)*',
    component: () => import('@views/exception/404/index.vue'),
    meta: { title: '404', isHideTab: true }
  },
  {
    path: '/500',
    name: 'Exception500',
    component: () => import('@views/exception/500/index.vue'),
    meta: { title: '500', isHideTab: true }
  },
  {
    path: '/outside',
    component: () => import('@views/index/index.vue'),
    name: 'Outside',
    meta: { title: 'menus.outside.title' },
    children: [
      // iframe 内嵌页面
      {
        path: '/outside/iframe/:path',
        name: 'Iframe',
        component: () => import('@/views/outside/Iframe.vue'),
        meta: { title: 'iframe' }
      }
    ]
  }
]
