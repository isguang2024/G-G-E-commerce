/**
 * 路由工具函数
 *
 * 提供路由相关的工具函数
 *
 * @module utils/router
 */
import type {
  RouteLocationNormalized,
  RouteRecordRaw,
  RouteRecordNormalized,
  Router
} from 'vue-router'
import AppConfig from '@/config'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'
import { $t } from '@/locales'

/** 扩展的路由配置类型 */
export type AppRouteRecordRaw = RouteRecordRaw & {
  hidden?: boolean
}

/** 顶部进度条配置 */
export const configureNProgress = () => {
  NProgress.configure({
    easing: 'ease',
    speed: 600,
    showSpinner: false,
    parent: 'body'
  })
}

/**
 * 设置页面标题，根据路由元信息和系统信息拼接标题
 * @param to 当前路由对象
 */
export const setPageTitle = (to: RouteLocationNormalized): void => {
  const { title } = to.meta
  if (title) {
    setTimeout(() => {
      document.title = `${formatMenuTitle(String(title))} - ${AppConfig.systemInfo.name}`
    }, 150)
  }
}

/**
 * 格式化菜单标题
 * @param title 菜单标题，可以是 i18n 的 key，也可以是字符串
 * @returns 格式化后的菜单标题
 */
export const formatMenuTitle = (title: string): string => {
  const normalizedTitle = `${title || ''}`.trim()
  if (normalizedTitle) {
    if (normalizedTitle.startsWith('menus.')) {
      const translated = `${$t(normalizedTitle) || ''}`.trim()
      if (translated && translated !== normalizedTitle) {
        return translated
      }
      return normalizedTitle.split('.').pop() || normalizedTitle
    }
    return normalizedTitle
  }
  return ''
}

/**
 * 统一比较路由路径时的归一化处理。
 * 这里只做字符串标准化，不触发 router.resolve，避免在动态路由尚未注册时额外制造 warning。
 */
export const normalizeComparableRoutePath = (path: string): string => {
  const target = `${path || ''}`.trim()
  if (!target || /^https?:\/\//i.test(target)) {
    return ''
  }
  const [pathname] = target.split(/[?#]/, 1)
  const normalized = `/${pathname.replace(/^\/+/, '')}`.replace(/\/+/g, '/')
  return normalized !== '/' ? normalized.replace(/\/$/, '') : normalized
}

export const routePathMatches = (routePath: string, targetPath: string): boolean => {
  const normalizedRoute = normalizeComparableRoutePath(routePath)
  const normalizedTarget = normalizeComparableRoutePath(targetPath)
  if (!normalizedRoute || !normalizedTarget) {
    return false
  }
  if (normalizedRoute === normalizedTarget) {
    return true
  }

  const routeSegments = normalizedRoute.split('/')
  const targetSegments = normalizedTarget.split('/')
  if (routeSegments.length !== targetSegments.length) {
    return false
  }

  for (let index = 0; index < routeSegments.length; index += 1) {
    const routeSegment = routeSegments[index]
    const targetSegment = targetSegments[index]
    if (routeSegment === targetSegment) {
      continue
    }
    if (routeSegment.startsWith(':')) {
      continue
    }
    if (routeSegment.includes('*')) {
      return true
    }
    return false
  }

  return true
}

/**
 * runtime/navigation 已经是后端编译后的唯一真源。
 * 当前端只想判断“这条路由是否已注册”时，必须静默读取 router.getRoutes()，
 * 不能再用 router.resolve('/foo') 探测，否则菜单被禁用或权限撤回时会刷一串 No match warning。
 */
export const findRegisteredRouteByPath = (
  router: Router,
  targetPath: string
): RouteRecordNormalized | undefined => {
  const normalizedTarget = normalizeComparableRoutePath(targetPath)
  if (!normalizedTarget) {
    return undefined
  }
  return router.getRoutes().find((route) => routePathMatches(route.path, normalizedTarget))
}

export const hasRegisteredRoutePath = (router: Router, targetPath: string): boolean => {
  return Boolean(findRegisteredRouteByPath(router, targetPath))
}
