/**
 * 根据前端路由模块生成「路由 name -> 组件路径」映射
 * 用于后端菜单未配置 component 时，按 name 解析到正确组件
 *
 * 注意：仅从 routeModules 收集，不包含 staticRoutes。
 * 若菜单管理里配置了“个人中心”等仅存在于静态路由的页面，需在此补充映射，否则会被 filterEmptyMenus 过滤掉。
 */
import type { AppRouteRecord } from '@/types/router'
import { routeModules } from '../modules'

const nameToComponent: Record<string, string> = {}

function collect(routes: AppRouteRecord[]) {
  if (!routes?.length) return
  for (const r of routes) {
    const name = r.name as string | undefined
    const comp = typeof r.component === 'string' ? r.component : ''
    if (name && comp) nameToComponent[name] = comp
    if (r.children?.length) collect(r.children)
  }
}

collect(routeModules)

// 仅存在于静态路由中的页面：菜单管理里添加后需在此映射，否则会被 filterEmptyMenus 当作“无组件”过滤掉
const USER_CENTER_COMPONENT = '/system/user-center'
nameToComponent['UserCenter'] = USER_CENTER_COMPONENT
nameToComponent['user-center1'] = USER_CENTER_COMPONENT
nameToComponent['usercenter1'] = USER_CENTER_COMPONENT

export function getComponentPathByRouteName(name: string): string | undefined {
  if (!name) return undefined
  return nameToComponent[name] ?? nameToComponent[String(name).replace(/-/g, '_')]
}
