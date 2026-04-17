/**
 * user-permission-test-drawer 视图：纯函数辅助工具。
 *
 * 抽离自 user-permission-test-drawer.vue，所有函数不依赖任何 reactive 状态，
 * 仅基于入参产出结果，便于单元测试与跨文件复用。
 */
import type { CascaderOption } from 'element-plus'
import { formatMenuTitle } from '@/utils/router'

export interface MenuOption extends CascaderOption {
  path?: string
  hidden?: boolean
  isIframe?: boolean
  isEnable?: boolean
  totalLeafCount?: number
  visibleLeafCount?: number
}

export function formatPermissionStatus(status?: string) {
  if (!status) return '-'
  return status === 'normal' || status === 'active' ? '正常' : '停用'
}

export function formatMemberStatus(status?: string) {
  switch (status) {
    case 'active':
      return '有效成员'
    case 'inactive':
      return '成员停用'
    case 'missing':
      return '未加入协作空间'
    default:
      return '-'
  }
}

export function formatRoleCode(roleCode?: string) {
  switch (roleCode) {
    case 'collaboration_admin':
      return '协作空间管理员'
    case 'collaboration_member':
      return '协作空间成员'
    default:
      return roleCode || '-'
  }
}

export function getMemberStatusTagType(status?: string) {
  switch (status) {
    case 'active':
      return 'success'
    case 'inactive':
      return 'danger'
    case 'missing':
      return 'warning'
    default:
      return 'info'
  }
}

export function formatBoundaryState(state?: string) {
  switch (state) {
    case '命中':
      return '命中'
    case '拦截':
      return '拦截'
    case '未配置':
      return '未配置'
    case '未命中':
      return '未命中'
    case '超级管理员直通':
      return '超级管理员直通'
    default:
      return '-'
  }
}

export function getBoundaryStateTagType(state?: string) {
  switch (state) {
    case '命中':
      return 'success'
    case '拦截':
      return 'danger'
    case '未配置':
      return 'warning'
    case '超级管理员直通':
      return 'warning'
    default:
      return 'info'
  }
}

export function countPermissionMenuLeaves(node: Api.SystemManage.UserPermissionMenuNode): number {
  if (!(node.children || []).length) return 1
  return (node.children || []).reduce((sum, child) => sum + countPermissionMenuLeaves(child), 0)
}

export function countVisibleMenuLeaves(items: MenuOption[]): number {
  return items.reduce((sum, item) => {
    if (!(item.children || []).length) return sum + 1
    return sum + countVisibleMenuLeaves((item.children || []) as MenuOption[])
  }, 0)
}

export function normalizePermissionMenuOptions(
  items: Api.SystemManage.UserPermissionMenuNode[]
): MenuOption[] {
  return items.map((item) => {
    const children = normalizePermissionMenuOptions(item.children || [])
    return {
      value: item.id,
      label: formatMenuTitle(item.title || item.name || ''),
      path: item.path || '',
      hidden: Boolean(item.hidden),
      isIframe: Boolean(item.path && /^https?:\/\//.test(item.path)),
      isEnable: true,
      leaf: !(item.children || []).length,
      totalLeafCount: countPermissionMenuLeaves(item),
      visibleLeafCount: countVisibleMenuLeaves(children),
      children
    }
  })
}

export function filterNestedOptions<T extends CascaderOption>(
  items: T[],
  predicate: (node: T) => boolean
): T[] {
  return items
    .map((item) => {
      const children = filterNestedOptions(((item.children || []) as T[]) || [], predicate)
      const passed = predicate(item)
      if (!passed && !children.length) return null
      return {
        ...item,
        children
      } as T
    })
    .filter((item): item is T => Boolean(item))
}

export function ensureExpandedMenus(panel: any, selectedValues: string[]) {
  const rootMenus = panel?.menus?.[0]
  if (!panel || !rootMenus?.length) return
  const firstValue = selectedValues?.[selectedValues.length - 1]
  let rootNode = rootMenus[0]
  let childNode = rootNode?.children?.[0]
  if (firstValue) {
    const matchedNode = panel
      .getFlattedNodes?.(false)
      ?.find((node: any) => `${node?.value}` === `${firstValue}`)
    const pathNodes = matchedNode?.pathNodes || []
    if (pathNodes[0]) rootNode = pathNodes[0]
    if (pathNodes[1]) childNode = pathNodes[1]
  }
  const nextMenus = [rootMenus]
  if (rootNode?.children?.length) nextMenus.push(rootNode.children)
  if (childNode?.children?.length) nextMenus.push(childNode.children)
  panel.menus = nextMenus
}
