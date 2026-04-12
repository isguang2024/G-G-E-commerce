/**
 * page 视图：纯函数辅助工具。
 *
 * 抽离自 index.vue，所有函数不依赖任何 reactive 状态，
 * 仅基于入参产出结果，便于单元测试与跨文件复用。
 */
import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
import { formatMenuTitle } from '@/utils/router'

type PageItem = Api.SystemManage.PageItem
type PageMenuOptionItem = Api.SystemManage.PageMenuOptionItem

export function toTreeSelectNode(item: PageMenuOptionItem): any {
  const title = `${item.title || ''}`.trim()
  const formattedTitle = formatMenuTitle(title)
  const menuName = `${item.name || ''}`.trim()
  const labelSource = formattedTitle || menuName || `${item.path || item.id}`.trim()
  return {
    label: labelSource,
    value: item.id,
    children: Array.isArray(item.children) ? item.children.map(toTreeSelectNode) : []
  }
}

export function normalizeMenuId(value: unknown): string {
  if (Array.isArray(value)) {
    for (let i = value.length - 1; i >= 0; i -= 1) {
      const item = `${value[i] ?? ''}`.trim()
      if (item) return item
    }
    return ''
  }
  return `${value ?? ''}`.trim()
}

export type TreePageItem = PageItem & { children?: TreePageItem[] }

export const normalizeKeyword = (value?: string) => `${value || ''}`.trim().toLowerCase()

export const comparePages = (left: PageItem, right: PageItem) => {
  const sortDiff = Number(left.sortOrder || 0) - Number(right.sortOrder || 0)
  if (sortDiff !== 0) return sortDiff
  return `${left.name || ''}${left.pageKey || ''}`.localeCompare(
    `${right.name || ''}${right.pageKey || ''}`,
    'zh-Hans-CN'
  )
}

export function buildPageTree(items: PageItem[]): TreePageItem[] {
  const logicNodeMap = new Map<string, TreePageItem>()
  const displayGroupMap = new Map<string, TreePageItem>()
  const childrenMap = new Map<string, TreePageItem[]>()
  const roots: TreePageItem[] = []

  items.forEach((item) => {
    const node = { ...item, children: [] }
    if (item.pageType === 'display_group') {
      displayGroupMap.set(item.pageKey, node)
      return
    }
    logicNodeMap.set(item.pageKey, node)
  })

  Array.from(logicNodeMap.values()).forEach((item) => {
    const parentKey = `${item.parentPageKey || ''}`.trim()
    if (!parentKey || !logicNodeMap.has(parentKey)) {
      roots.push(item)
      return
    }
    const children = childrenMap.get(parentKey) || []
    children.push(item)
    childrenMap.set(parentKey, children)
  })

  const attachChildren = (node: TreePageItem) => {
    const children = (childrenMap.get(node.pageKey) || []).sort(comparePages)
    node.children = children.map((child) => attachChildren(child))
    return node
  }

  const resolvedRoots = roots.sort(comparePages).map((item) => attachChildren(item))
  const ungroupedRoots: TreePageItem[] = []
  resolvedRoots.forEach((item) => {
    const displayGroupKey = `${item.displayGroupKey || ''}`.trim()
    const displayGroup = displayGroupKey ? displayGroupMap.get(displayGroupKey) : undefined
    if (!displayGroup) {
      ungroupedRoots.push(item)
      return
    }
    const groupChildren = displayGroup.children || []
    groupChildren.push(item)
    displayGroup.children = groupChildren.sort(comparePages)
  })

  const groupedRoots = Array.from(displayGroupMap.values()).sort(comparePages)
  return [...groupedRoots, ...ungroupedRoots].sort(comparePages)
}

export function countTreeNodes(items: TreePageItem[]): number {
  return items.reduce((total, item) => total + 1 + countTreeNodes(item.children || []), 0)
}

export function buildMenuPathMap(
  items: Api.SystemManage.PageMenuOptionItem[],
  joinPath: (parent: string, segment?: string) => string
) {
  const nextMap = new Map<string, string>()
  const walk = (nodes: Api.SystemManage.PageMenuOptionItem[], parentPath = '') => {
    nodes.forEach((item) => {
      const fullPath = joinPath(parentPath, item.path)
      if (item.id) {
        nextMap.set(item.id, fullPath)
      }
      if (Array.isArray(item.children) && item.children.length) {
        walk(item.children, fullPath)
      }
    })
  }
  walk(items)
  return nextMap
}

export function buildCopyPageData(row: PageItem): Partial<PageItem> {
  const routeNameBase = `${row.routeName || row.pageKey || ''}`.trim()
  const routePathBase = `${row.routePath || ''}`.trim()
  return {
    ...row,
    id: '',
    name: `${row.name || '页面'} 副本`,
    pageKey: row.pageKey ? `${row.pageKey}.copy` : '',
    routeName: routeNameBase ? `${routeNameBase}Copy` : '',
    routePath: routePathBase,
    source: 'manual'
  }
}

export function getPageTypeText(row: PageItem) {
  if (row.pageType === 'group') return '逻辑分组'
  if (row.pageType === 'display_group') return '普通分组'
  if (row.pageType === 'standalone') return '独立页'
  return '内页'
}

export function getPageTypeTag(row: PageItem) {
  if (row.pageType === 'group') return 'info'
  if (row.pageType === 'display_group') return 'success'
  if (row.pageType === 'standalone') return 'primary'
  return 'warning'
}

export function getAccessModeText(accessMode?: string) {
  const accessModeTextMap: Record<string, string> = {
    inherit: '继承',
    public: '公开',
    jwt: '登录',
    permission: '权限'
  }
  return accessModeTextMap[accessMode || 'inherit'] || accessMode || '-'
}

export function getAccessModeTag(accessMode?: string) {
  const tagMap: Record<string, 'primary' | 'success' | 'info' | 'warning' | 'danger'> = {
    inherit: 'info',
    public: 'success',
    jwt: 'warning',
    permission: 'danger'
  }
  return tagMap[accessMode || 'inherit'] || 'info'
}

export function getMountModeText(row: PageItem) {
  if (row.pageType === 'standalone') {
    return row.visibilityScope === 'spaces' || row.spaceScope === 'spaces'
      ? '独立页（指定空间）'
      : '独立页'
  }
  if (row.parentPageKey) return '挂到页面'
  if (row.parentMenuId) return '挂到菜单'
  return '未挂载内页'
}

export function getMountTargetText(row: PageItem) {
  if (row.pageType === 'display_group') return '普通分组'
  if (row.pageType === 'standalone') {
    return row.visibilityScope === 'spaces' || row.spaceScope === 'spaces'
      ? '独立页 · 指定空间'
      : '独立页'
  }
  if (row.pageType === 'group') {
    return row.displayGroupName ? `逻辑分组 · ${row.displayGroupName}` : '逻辑分组'
  }
  if (row.parentMenuName) return `挂到菜单 · ${row.parentMenuName}`
  if (row.parentPageName) return `挂到页面 · ${row.parentPageName}`
  if (row.displayGroupName) return `列表分组 · ${row.displayGroupName}`
  return '未挂载内页'
}

export function getRelationDisplayText(row: PageItem) {
  if (row.pageType === 'display_group') {
    return '仅列表归类'
  }
  if (row.pageType === 'standalone') {
    return row.visibilityScope === 'spaces' || row.spaceScope === 'spaces'
      ? '独立页 · 指定空间'
      : '独立页 · App 全局'
  }
  const parentPageName = `${row.parentPageName || ''}`.trim()
  if (parentPageName) {
    return `挂到页面 · ${parentPageName}`
  }
  const parentMenuName = `${row.parentMenuName || ''}`.trim()
  if (parentMenuName) {
    return `挂到菜单 · ${parentMenuName}`
  }
  const displayGroupName = `${row.displayGroupName || ''}`.trim()
  if (displayGroupName) {
    return `普通分组：${displayGroupName}`
  }
  return '未挂载内页'
}

function hasRemoteMeta(meta?: Record<string, any>) {
  if (!meta || typeof meta !== 'object') return false
  return [
    meta.remoteAppKey,
    meta.remotePageKey,
    meta.remoteEntryUrl,
    meta.remoteModule,
    meta.remoteModuleName,
    meta.remoteRoutePath,
    meta.remoteUrl
  ].some((item) => `${item || ''}`.trim())
}

function hasRemoteBinding(row: PageItem) {
  const binding = row.remoteBinding
  if (!binding || typeof binding !== 'object') return false
  return [
    binding.manifestUrl,
    binding.remoteAppKey,
    binding.remotePageKey,
    binding.remoteEntryUrl,
    binding.remoteRoutePath,
    binding.remoteModule,
    binding.remoteModuleName,
    binding.remoteUrl,
    binding.runtimeVersion,
    binding.healthCheckUrl
  ].some((item) => `${item || ''}`.trim())
}

export function getPageSourceKind(row: PageItem) {
  const link = `${row.link || ''}`.trim()
  if (/^https?:\/\//i.test(link) || hasRemoteBinding(row) || hasRemoteMeta(row.meta)) {
    return 'remote'
  }
  if (row.source === 'sync') return 'sync'
  if (row.source === 'seed') return 'seed'
  return 'manual'
}

export function getPageSourceText(row: PageItem) {
  const sourceKind = getPageSourceKind(row)
  if (sourceKind === 'remote') return '远端页'
  if (sourceKind === 'sync') return '扫描同步'
  if (sourceKind === 'seed') return 'Seed'
  return '本地配置'
}

export function getPageSourceTag(row: PageItem) {
  const sourceKind = getPageSourceKind(row)
  if (sourceKind === 'remote') return 'primary'
  if (sourceKind === 'sync') return 'success'
  if (sourceKind === 'seed') return 'info'
  return 'warning'
}

export function getPageGovernanceText(row: PageItem) {
  const sourceKind = getPageSourceKind(row)
  if (sourceKind === 'remote') {
    const binding = row.remoteBinding
    const manifest = `${binding?.manifestUrl || ''}`.trim()
    const version = `${binding?.runtimeVersion || ''}`.trim()
    if (manifest || version) {
      return `远端接入页已声明${manifest ? ' manifest' : ''}${manifest && version ? ' / ' : ''}${version ? '版本' : ''}契约，应优先以治理后端下发字段为准，不再回退到 link/meta 猜测。`
    }
    return '远端接入页应以 link 或 remote contract 作为唯一入口，避免再为同一路由补本地组件。'
  }
  if (sourceKind === 'sync') {
    return '扫描同步页应优先回到扫描源修正，避免人工编辑后再次被覆盖。'
  }
  if (sourceKind === 'seed') {
    return 'Seed 页适合做基线能力，后续差异优先走增量配置，不回写初始化脚本。'
  }
  return '本地配置页由当前后台直接治理，注意不要和远端入口或扫描结果重复占用同一路由。'
}

export function formatUpdatedAt(value?: string) {
  const target = `${value || ''}`.trim()
  if (!target) {
    return '-'
  }
  const date = new Date(target)
  if (Number.isNaN(date.getTime())) {
    return target
  }
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  const hour = `${date.getHours()}`.padStart(2, '0')
  const minute = `${date.getMinutes()}`.padStart(2, '0')
  return `${year}-${month}-${day} ${hour}:${minute}`
}

export function toPageSaveParams(
  row: PageItem,
  nextSortOrder: number,
  appKey: string
): Api.SystemManage.PageSaveParams {
  return {
    app_key: appKey,
    page_key: row.pageKey,
    name: row.name,
    route_name: row.routeName || row.pageKey,
    route_path: row.routePath || '',
    component: row.component || '',
    page_type: row.pageType,
    source: row.source || 'manual',
    module_key: row.moduleKey || '',
    sort_order: nextSortOrder,
    parent_menu_id: row.parentMenuId || '',
    parent_page_key: row.parentPageKey || '',
    display_group_key: row.displayGroupKey || '',
    active_menu_path: row.activeMenuPath || '',
    breadcrumb_mode: row.breadcrumbMode || 'inherit_menu',
    access_mode: row.accessMode || 'inherit',
    permission_key: row.permissionKey || '',
    keep_alive: Boolean(row.keepAlive),
    is_full_page: Boolean(row.isFullPage),
    remote_binding: row.remoteBinding
      ? {
          manifest_url: row.remoteBinding.manifestUrl || '',
          remote_app_key: row.remoteBinding.remoteAppKey || '',
          remote_page_key: row.remoteBinding.remotePageKey || '',
          remote_entry_url: row.remoteBinding.remoteEntryUrl || '',
          remote_route_path: row.remoteBinding.remoteRoutePath || '',
          remote_module: row.remoteBinding.remoteModule || '',
          remote_module_name: row.remoteBinding.remoteModuleName || '',
          remote_url: row.remoteBinding.remoteUrl || '',
          runtime_version: row.remoteBinding.runtimeVersion || '',
          health_check_url: row.remoteBinding.healthCheckUrl || ''
        }
      : undefined,
    status: row.status || 'normal',
    meta: {
      ...(row.meta || {}),
      isIframe: Boolean(row.isIframe),
      isHideTab: Boolean(row.isHideTab),
      link: row.link || ''
    }
  }
}

export function getOperationList(row: PageItem): ButtonMoreItem[] {
  if (row.pageType === 'display_group') {
    return [
      {
        key: 'add-group',
        label: '新增组内逻辑分组',
        icon: 'ri:folder-add-line',
        auth: 'system.page.manage'
      },
      {
        key: 'add-page',
        label: '新增组内页面',
        icon: 'ri:file-add-line',
        auth: 'system.page.manage'
      },
      { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'system.page.manage' },
      {
        key: 'delete',
        label: '删除',
        icon: 'ri:delete-bin-4-line',
        auth: 'system.page.manage',
        color: '#f56c6c'
      }
    ]
  }
  const list: ButtonMoreItem[] = [
    {
      key: 'add-group',
      label: '新增子逻辑分组',
      icon: 'ri:folder-add-line',
      auth: 'system.page.manage'
    },
    {
      key: 'add-page',
      label: '新增子页面',
      icon: 'ri:file-add-line',
      auth: 'system.page.manage'
    },
    { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'system.page.manage' },
    {
      key: 'delete',
      label: '删除',
      icon: 'ri:delete-bin-4-line',
      auth: 'system.page.manage',
      color: '#f56c6c'
    }
  ]
  if (row.pageType === 'inner' || row.pageType === 'standalone') {
    list.splice(3, 0, {
      key: 'copy',
      label: '复制页面',
      icon: 'ri:file-copy-line',
      auth: 'system.page.manage'
    })
    list.splice(3, 0, { key: 'visit', label: '访问', icon: 'ri:external-link-line' })
  }
  return list
}
