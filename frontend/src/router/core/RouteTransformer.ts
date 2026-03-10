/**
 * 路由转换器
 *
 * 负责将菜单数据转换为 Vue Router 路由配置
 *
 * @module router/core/RouteTransformer
 * @author Art Design Pro Team
 */

import type { RouteRecordRaw } from 'vue-router'
import type { AppRouteRecord } from '@/types/router'
import { ComponentLoader } from './ComponentLoader'
import { IframeRouteManager } from './IframeRouteManager'
import { getComponentPathByRouteName } from './routeNameComponentMap'

interface ConvertedRoute extends Omit<RouteRecordRaw, 'children'> {
  id?: number
  children?: ConvertedRoute[]
  component?: RouteRecordRaw['component'] | (() => Promise<any>)
}

export class RouteTransformer {
  private componentLoader: ComponentLoader
  private iframeManager: IframeRouteManager

  constructor(componentLoader: ComponentLoader) {
    this.componentLoader = componentLoader
    this.iframeManager = IframeRouteManager.getInstance()
  }

  /**
   * 转换路由配置
   * 当菜单项有 name 但无 component 时，按前端路由表用 name 解析组件路径，避免多级下误用父级组件
   * @param ancestorNames 祖先路由的 name 集合，用于避免子路由与祖先重名（Vue Router 要求 name 全局唯一）
   */
  transform(route: AppRouteRecord, depth = 0, ancestorNames: Set<string> = new Set()): ConvertedRoute {
    const { component, children, ...routeConfig } = route

    let componentPath = typeof component === 'string' ? component : ''
    if (route.name) {
      const resolved = getComponentPathByRouteName(String(route.name))
      const hasValidComponent = componentPath?.trim() && this.componentLoader.exists(componentPath)

      // 组件为空或后端给了错误路径时，按前端路由名映射回退到正确组件
      if ((!hasValidComponent || !componentPath?.trim()) && resolved && this.componentLoader.exists(resolved)) {
        if (componentPath?.trim() && componentPath !== resolved) {
          console.warn(
            `[RouteTransformer] 路由 "${route.name}" 的组件路径 "${componentPath}" 无效，已回退为 "${resolved}"`
          )
        }
        componentPath = resolved
      }
    }

    // 若当前 name 与某祖先重复，改为唯一名，避免 Vue Router 报错
    let uniqueName = route.name
    if (uniqueName && ancestorNames.has(String(uniqueName))) {
      const pathSeg = (route.path || '').replace(/^\//, '').replace(/\//g, '_') || 'child'
      uniqueName = `${uniqueName}_${pathSeg}`
    }
    const nextAncestor = new Set(ancestorNames)
    if (uniqueName) nextAncestor.add(String(uniqueName))

    const converted: ConvertedRoute = {
      ...routeConfig,
      name: uniqueName,
      component: undefined
    }

    // 处理不同类型的路由
    if (route.meta?.isIframe) {
      this.handleIframeRoute(converted, route, depth)
    } else if (this.isFirstLevelRoute(route, depth)) {
      this.handleFirstLevelRoute(converted, route, componentPath)
    } else {
      this.handleNormalRoute(converted, componentPath)
    }

    // 递归处理子路由，传入祖先 name 集合
    if (children?.length) {
      converted.children = children.map((child) => this.transform(child, depth + 1, nextAncestor))
    }

    return converted
  }

  /**
   * 判断是否为一级路由（需要 Layout 包裹）
   */
  private isFirstLevelRoute(route: AppRouteRecord, depth: number): boolean {
    return depth === 0 && (!route.children || route.children.length === 0)
  }

  /**
   * 处理 iframe 类型路由
   */
  private handleIframeRoute(
    targetRoute: ConvertedRoute,
    sourceRoute: AppRouteRecord,
    depth: number
  ): void {
    if (depth === 0) {
      // 顶级 iframe：用 Layout 包裹
      targetRoute.component = this.componentLoader.loadLayout()
      targetRoute.path = this.extractFirstSegment(sourceRoute.path || '')
      targetRoute.name = ''

      targetRoute.children = [
        {
          ...sourceRoute,
          component: this.componentLoader.loadIframe()
        } as ConvertedRoute
      ]
    } else {
      // 非顶级（嵌套）iframe：直接使用 Iframe.vue
      targetRoute.component = this.componentLoader.loadIframe()
    }

    // 记录 iframe 路由
    this.iframeManager.add(sourceRoute)
  }

  /**
   * 处理一级菜单路由
   */
  private handleFirstLevelRoute(
    converted: ConvertedRoute,
    route: AppRouteRecord,
    component: string | undefined
  ): void {
    converted.component = this.componentLoader.loadLayout()
    converted.path = this.extractFirstSegment(route.path || '')
    converted.name = ''
    route.meta.isFirstLevel = true

    converted.children = [
      {
        ...route,
        component: component ? this.componentLoader.load(component) : undefined
      } as ConvertedRoute
    ]
  }

  /**
   * 处理普通路由
   */
  private handleNormalRoute(converted: ConvertedRoute, component: string | undefined): void {
    if (component) {
      converted.component = this.componentLoader.load(component)
    }
  }

  /**
   * 提取路径的第一段
   */
  private extractFirstSegment(path: string): string {
    const segments = path.split('/').filter(Boolean)
    return segments.length > 0 ? `/${segments[0]}` : '/'
  }
}
