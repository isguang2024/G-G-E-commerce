import type { App } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import { setupAfterEachGuard } from '@/domains/navigation/guards/afterEach'
import { setupBeforeEachGuard } from '@/domains/navigation/guards/beforeEach'
import { DEFAULT_HOME_PAGE_PATH } from '@/domains/navigation/constants'
import { registerNavigationRouter } from '@/domains/navigation/runtime/router-instance'
import { staticRoutes } from './routes/staticRoutes'
import { configureNProgress } from '@/utils/router'

// 创建路由实例
export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: staticRoutes // 静态路由
})
registerNavigationRouter(router)

// 初始化路由
export function initRouter(app: App<Element>): void {
  configureNProgress() // 顶部进度条
  setupBeforeEachGuard(router) // 路由前置守卫
  setupAfterEachGuard(router) // 路由后置守卫
  app.use(router)
}

// 主页路径，默认使用菜单第一个有效路径，配置后使用此路径
export const HOME_PAGE_PATH = DEFAULT_HOME_PAGE_PATH
