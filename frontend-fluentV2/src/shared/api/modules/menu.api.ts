import { requestData } from '@/shared/api/client'
import type { MenuManageGroup, MenuNode, MenuPageBinding } from '@/shared/types/menu'

function normalizeManageGroup(input?: Record<string, unknown>): MenuManageGroup | undefined {
  if (!input) {
    return undefined
  }

  return {
    id: `${input.id || ''}`.trim(),
    name: `${input.name || ''}`.trim(),
    sortOrder: Number(input.sort_order ?? input.sortOrder ?? 0),
    status: `${input.status || 'normal'}`.trim(),
  }
}

function normalizeMenuNode(input: Record<string, unknown>): MenuNode {
  const meta = (input.meta || {}) as Record<string, unknown>
  return {
    id: `${input.id || ''}`.trim(),
    parentId: `${input.parent_id || input.parentId || ''}`.trim() || undefined,
    manageGroupId: `${input.manage_group_id || input.manageGroupId || ''}`.trim() || undefined,
    manageGroup: normalizeManageGroup(input.manage_group as Record<string, unknown> | undefined),
    spaceKey: `${input.space_key || input.spaceKey || meta.spaceKey || 'default'}`.trim(),
    kind: `${input.kind || ''}`.trim(),
    path: `${input.path || ''}`.trim(),
    name: `${input.name || ''}`.trim(),
    title: `${meta.title || input.title || input.name || ''}`.trim(),
    component: `${input.component || ''}`.trim(),
    icon: `${meta.icon || input.icon || ''}`.trim(),
    sortOrder: Number(input.sort_order ?? input.sortOrder ?? 0),
    hidden: Boolean(input.hidden ?? meta.isHide ?? false),
    meta,
    children: Array.isArray(input.children)
      ? input.children.map((item) => normalizeMenuNode(item as Record<string, unknown>))
      : [],
  }
}

function normalizeRuntimePage(input: Record<string, unknown>): MenuPageBinding {
  return {
    pageKey: `${input.page_key || input.pageKey || ''}`.trim(),
    name: `${input.name || ''}`.trim(),
    routePath: `${input.route_path || input.routePath || ''}`.trim(),
    routeName: `${input.route_name || input.routeName || ''}`.trim() || undefined,
    component: `${input.component || ''}`.trim(),
    parentMenuId: `${input.parent_menu_id || input.parentMenuId || ''}`.trim() || undefined,
    accessMode: `${input.access_mode || input.accessMode || ''}`.trim() || undefined,
    permissionKey: `${input.permission_key || input.permissionKey || ''}`.trim() || undefined,
    parentPageKey: `${input.parent_page_key || input.parentPageKey || ''}`.trim() || undefined,
    pageType: `${input.page_type || input.pageType || 'inner'}`.trim(),
    spaceKey: `${input.space_key || input.spaceKey || ''}`.trim() || undefined,
    spaceKeys: Array.isArray(input.space_keys || input.spaceKeys)
      ? ((input.space_keys || input.spaceKeys) as unknown[])
          .map((item: unknown) => `${item || ''}`.trim())
          .filter(Boolean)
      : [],
  }
}

export async function fetchMenuTree(spaceKey: string) {
  const result = await requestData<Array<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/menus/tree',
    params: {
      all: 1,
      space_key: spaceKey,
    },
  })

  return result.map(normalizeMenuNode)
}

export async function fetchMenuManageGroups() {
  const result = await requestData<Array<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/menus/groups',
  })

  return result.map((item) => normalizeManageGroup(item)!).filter(Boolean)
}

export async function fetchRuntimePages(spaceKey: string) {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
  }>({
    method: 'GET',
    url: '/api/v1/pages/runtime',
    params: {
      space_key: spaceKey,
    },
  })

  return (result.records || []).map(normalizeRuntimePage)
}
