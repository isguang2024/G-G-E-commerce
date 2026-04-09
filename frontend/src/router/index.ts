import type { App } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { staticRoutes } from './routes/staticRoutes'
import { configureNProgress } from '@/utils/router'
import {
  setupBeforeEachGuard,
  refreshCurrentUserInfoContext,
  refreshUserMenus,
  refreshUserAccessAndMenus
} from './guards/beforeEach'
import { setupAfterEachGuard } from './guards/afterEach'

/** 角色/用户菜单权限变更后调用，使侧栏与动态路由立即更新 */
export { refreshUserMenus }
export { refreshCurrentUserInfoContext }
export { refreshUserAccessAndMenus }

// 创建路由实例
export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: staticRoutes // 静态路由
})

// 初始化路由
export function initRouter(app: App<Element>): void {
  configureNProgress() // 顶部进度条
  setupBeforeEachGuard(router) // 路由前置守卫
  setupAfterEachGuard(router) // 路由后置守卫
  app.use(router)
}

// 主页路径，默认使用菜单第一个有效路径，配置后使用此路径
export const HOME_PAGE_PATH = ''
