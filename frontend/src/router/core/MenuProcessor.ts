/**
 * 菜单处理器
 *
 * 负责菜单数据的获取、过滤和处理
 *
 * @module router/core/MenuProcessor
 * @author Art Design Pro Team
 */

import type { AppRouteRecord } from '@/types/router'
import { fetchGetMenuList } from '@/api/system-manage'
import { useMenuSpaceStore } from '@/store/modules/menu-space'
import { useUserStore } from '@/store/modules/user'
import { RoutesAlias } from '../routesAlias'
import { formatMenuTitle } from '@/utils'
import { hasMenuActionAccess, shouldHideMenuWhenActionDenied } from '@/utils/permission/menu'
import { isMenuSpaceVisible, normalizeMenuSpaceKey } from '@/utils/navigation/menu-space'

export class MenuProcessor {
  /**
   * 获取菜单数据（仅从后端实时获取，不再使用前端默认菜单）
   */
  async getMenuList(): Promise<AppRouteRecord[]> {
    const spaceStore = useMenuSpaceStore()
    const menuList = await fetchGetMenuList(spaceStore.currentSpaceKey)
    const filteredByEnabled = this.filterDisabledMenus(menuList)
    const filteredByAction = this.filterActionRequirementMenus(filteredByEnabled)
    spaceStore.syncRuntimeHost()
    const filteredBySpace = this.filterMenusBySpace(
      filteredByAction,
      spaceStore.currentSpaceKey,
      spaceStore.defaultSpaceKey
    )

    // 在规范化路径之前，验证原始路径配置
    this.validateMenuPaths(filteredBySpace)

    // 规范化路径（将相对路径转换为完整路径）
    return this.normalizeMenuPaths(this.filterEmptyMenus(filteredBySpace))
  }

  /**
   * 过滤运行时被标记为未启用的菜单，避免后端遗漏时仍出现在导航里
   */
  private filterDisabledMenus(menuList: AppRouteRecord[]): AppRouteRecord[] {
    return menuList.reduce<AppRouteRecord[]>((result, item) => {
      if (item.meta?.isEnable === false) {
        return result
      }

      const children = item.children?.length ? this.filterDisabledMenus(item.children) : item.children
      result.push({
        ...item,
        children
      })
      return result
    }, [])
  }

  /**
   * 递归过滤空菜单项
   */
  private filterEmptyMenus(menuList: AppRouteRecord[]): AppRouteRecord[] {
    return menuList
      .map((item) => {
        // 如果有子菜单，先递归过滤子菜单
        if (item.children && item.children.length > 0) {
          const filteredChildren = this.filterEmptyMenus(item.children)
          return {
            ...item,
            children: filteredChildren
          }
        }
        return item
      })
      .filter((item) => {
        // 有有效子菜单的目录菜单，保留
        if (item.children && item.children.length > 0) {
          return true
        }

        // 如果有外链或 iframe，保留
        if (item.meta?.isIframe === true || item.meta?.link) {
          return true
        }

        // 如果有有效的 component，保留
        if (item.component && item.component !== '' && item.component !== RoutesAlias.Layout) {
          return true
        }

        // 其他情况过滤掉
        return false
      })
  }

  /**
   * 验证菜单列表是否有效
   */
  validateMenuList(menuList: AppRouteRecord[]): boolean {
    return Array.isArray(menuList)
  }

  /**
   * 过滤绑定了基础功能门槛但当前用户不满足的菜单
   */
  private filterActionRequirementMenus(menuList: AppRouteRecord[]): AppRouteRecord[] {
    const userStore = useUserStore()
    const userInfo = userStore.getUserInfo as Api.Auth.UserInfo | undefined

    if (userInfo?.is_super_admin) {
      return menuList
    }

    return menuList.reduce<AppRouteRecord[]>((result, item) => {
        const children = item.children?.length
          ? this.filterActionRequirementMenus(item.children)
          : item.children
        const hasActionAccess = hasMenuActionAccess(userInfo, item.meta)
        const shouldHide = shouldHideMenuWhenActionDenied(item.meta)

        if (!hasActionAccess && shouldHide && !children?.length) {
          return result
        }

        result.push({
          ...item,
          children
        })
        return result
      }, [])
  }

  /**
   * 按菜单空间过滤菜单
   */
  private filterMenusBySpace(
    menuList: AppRouteRecord[],
    currentSpaceKey: string,
    defaultSpaceKey: string,
    inheritedSpaceKey = ''
  ): AppRouteRecord[] {
    return menuList.reduce<AppRouteRecord[]>((result, item) => {
      const ownSpaceKey = normalizeMenuSpaceKey(
        item.spaceKey || item.meta?.spaceKey || item.meta?.space_key
      )
      const effectiveSpaceKey = ownSpaceKey || normalizeMenuSpaceKey(inheritedSpaceKey) || defaultSpaceKey
      const children = item.children?.length
        ? this.filterMenusBySpace(item.children, currentSpaceKey, defaultSpaceKey, effectiveSpaceKey)
        : item.children

      const clone: AppRouteRecord = {
        ...item,
        spaceKey: effectiveSpaceKey,
        meta: {
          ...(item.meta || {}),
          spaceKey: effectiveSpaceKey
        },
        children
      }

      if (!isMenuSpaceVisible(effectiveSpaceKey, currentSpaceKey, defaultSpaceKey) && !children?.length) {
        return result
      }

      result.push(clone)
      return result
    }, [])
  }

  /**
   * 规范化菜单路径
   * 将相对路径转换为完整路径，确保菜单跳转正确
   */
  private normalizeMenuPaths(menuList: AppRouteRecord[], parentPath = ''): AppRouteRecord[] {
    return menuList.map((item) => {
      // 构建完整路径
      const fullPath = this.buildFullPath(item.path || '', parentPath)

      // 递归处理子菜单
      const children = item.children?.length
        ? this.normalizeMenuPaths(item.children, fullPath)
        : item.children

      return {
        ...item,
        path: fullPath,
        children
      }
    })
  }

  /**
   * 验证菜单路径配置
   * 检测非法的绝对路径配置
   */
  private validateMenuPaths(menuList: AppRouteRecord[], level = 1): void {
    menuList.forEach((route) => {
      if (!route.children?.length) return

      const parentName = String(route.name || route.path || '未知路由')

      route.children.forEach((child) => {
        const childPath = child.path || ''

        // 跳过合法的绝对路径：外部链接和 iframe 路由
        if (this.isValidAbsolutePath(childPath)) return

        // 检测非法的绝对路径
        if (childPath.startsWith('/')) {
          this.logPathError(child, childPath, parentName, level)
        }
      })

      // 递归检查更深层级的子路由
      this.validateMenuPaths(route.children, level + 1)
    })
  }

  /**
   * 判断是否为合法的绝对路径
   */
  private isValidAbsolutePath(path: string): boolean {
    return (
      path.startsWith('/') ||
      path.startsWith('http://') ||
      path.startsWith('https://') ||
      path.startsWith('/outside/iframe/')
    )
  }

  /**
   * 输出路径配置错误日志
   */
  private logPathError(
    route: AppRouteRecord,
    path: string,
    parentName: string,
    level: number
  ): void {
    const routeName = String(route.name || path || '未知路由')
    const menuTitle = route.meta?.title || routeName
    const suggestedPath = path.split('/').pop() || path.slice(1)

    console.error(
      `[路由配置错误] 菜单 "${formatMenuTitle(menuTitle)}" (name: ${routeName}, path: ${path}) 配置错误\n` +
        `  位置: ${parentName} > ${routeName}\n` +
        `  问题: 当前绝对路径不在允许范围内\n` +
        `  当前配置: path: '${path}'\n` +
        `  建议检查是否应改为相对路径: '${suggestedPath}'`
    )
  }

  /**
   * 构建完整路径
   */
  private buildFullPath(path: string, parentPath: string): string {
    if (!path) return ''

    // 外部链接直接返回
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path
    }

    // 如果已经是绝对路径，直接返回
    if (path.startsWith('/')) {
      return path
    }

    // 拼接父路径和当前路径
    if (parentPath) {
      // 移除父路径末尾的斜杠，移除子路径开头的斜杠，然后拼接
      const cleanParent = parentPath.replace(/\/$/, '')
      const cleanChild = path.replace(/^\//, '')
      return `${cleanParent}/${cleanChild}`
    }

    // 没有父路径，添加前导斜杠
    return `/${path}`
  }
}
