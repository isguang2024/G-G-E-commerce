/**
 * 路由类型定义模块
 *
 * 提供路由相关的类型定义
 *
 * ## 主要功能
 *
 * - 路由元数据类型（标题、图标、权限等）
 * - 应用路由记录类型
 * - 路由配置扩展
 *
 * ## 使用场景
 *
 * - 路由配置类型约束
 * - 路由元数据定义
 * - 菜单生成
 * - 权限控制
 *
 * @module types/router/index
 * @author Art Design Pro Team
 */

import { RouteRecordRaw } from 'vue-router'

/**
 * 路由元数据接口
 * 定义路由的各种配置属性
 */
export interface RouteMeta extends Record<string | number | symbol, unknown> {
  /** 路由标题 */
  title: string
  /** 路由图标 */
  icon?: string
  /** 是否显示徽章 */
  showBadge?: boolean
  /** 文本徽章 */
  showTextBadge?: string
  /** 是否在菜单中隐藏 */
  isHide?: boolean
  /** 是否为内页（默认不显示在侧栏，仅通过按钮等跳转） */
  isInnerPage?: boolean
  /** 是否在标签页中隐藏 */
  isHideTab?: boolean
  /** 外部链接 */
  link?: string
  /** 是否为iframe */
  isIframe?: boolean
  /** 是否缓存 */
  keepAlive?: boolean
  /** 是否为一级菜单 */
  isFirstLevel?: boolean
  /** 角色权限 */
  roles?: string[]
  /** 是否固定标签页 */
  fixedTab?: boolean
  /** 激活菜单路径 */
  activePath?: string
  /** 是否为全屏页面 */
  isFullPage?: boolean
  /** 父级路径 */
  parentPath?: string
  /** 是否启用 */
  isEnable?: boolean
  /** 进入页面所需的基础功能权限 */
  requiredAction?: string
  /** 进入页面所需的基础功能权限列表 */
  requiredActions?: string[]
  /** 多个功能权限的匹配方式 */
  actionMatchMode?: 'any' | 'all'
  /** 功能权限不满足时的菜单可见策略 */
  actionVisibilityMode?: 'hide' | 'show'
  /** 自定义上级菜单路径（用于面包屑显示） */
  customParent?: string
}

/**
 * 应用路由记录接口
 * 扩展 Vue Router 的路由记录类型
 */
export interface AppRouteRecord extends Omit<RouteRecordRaw, 'meta' | 'children' | 'component'> {
  id?: number
  meta: RouteMeta
  children?: AppRouteRecord[]
  component?: string | (() => Promise<any>)
  sort_order?: number
  parent_id?: number | string
  is_system?: boolean
}
