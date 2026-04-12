import type { Router } from 'vue-router'

let navigationRouter: Router | null = null

export function registerNavigationRouter(router: Router): void {
  navigationRouter = router
}

export function getNavigationRouter(): Router {
  if (!navigationRouter) {
    throw new Error('[NavigationRuntime] router 尚未初始化')
  }
  return navigationRouter
}
