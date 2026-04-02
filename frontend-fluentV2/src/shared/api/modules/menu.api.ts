import { requestData } from '@/shared/api/client'
import type {
  MenuDeletePreview,
  MenuManageGroup,
  MenuMutationDraft,
  MenuNode,
  MenuNodeKind,
  MenuNodeMeta,
  MenuPageBinding,
} from '@/shared/types/menu'

function normalizeMenuKind(value: unknown, component: unknown, meta: MenuNodeMeta): MenuNodeKind {
  const target = `${value || ''}`.trim().toLowerCase()
  if (target === 'directory' || target === 'entry' || target === 'external') {
    return target
  }

  if (`${meta.link || ''}`.trim()) {
    return 'external'
  }

  if (`${component || ''}`.trim()) {
    return 'entry'
  }

  return 'directory'
}

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

function normalizeMenuMeta(input?: Record<string, unknown>): MenuNodeMeta {
  const meta = (input || {}) as MenuNodeMeta
  return {
    ...meta,
    title: `${meta.title || ''}`.trim() || undefined,
    icon: `${meta.icon || ''}`.trim() || undefined,
    link: `${meta.link || ''}`.trim() || undefined,
    activePath: `${meta.activePath || ''}`.trim() || undefined,
    accessMode: `${meta.accessMode || 'permission'}`.trim() || 'permission',
    isHide: Boolean(meta.isHide),
    isIframe: Boolean(meta.isIframe),
    isHideTab: Boolean(meta.isHideTab),
    keepAlive: Boolean(meta.keepAlive),
    fixedTab: Boolean(meta.fixedTab),
    isFullPage: Boolean(meta.isFullPage),
  }
}

function normalizeMenuNode(input: Record<string, unknown>): MenuNode {
  const meta = normalizeMenuMeta((input.meta || {}) as Record<string, unknown>)
  return {
    id: `${input.id || ''}`.trim(),
    parentId: `${input.parent_id || input.parentId || ''}`.trim() || undefined,
    manageGroupId: `${input.manage_group_id || input.manageGroupId || ''}`.trim() || undefined,
    manageGroup: normalizeManageGroup(input.manage_group as Record<string, unknown> | undefined),
    spaceKey: `${input.space_key || input.spaceKey || meta.spaceKey || 'default'}`.trim(),
    kind: normalizeMenuKind(input.kind, input.component, meta),
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

function buildMenuMetaPayload(input: MenuMutationDraft) {
  const meta: MenuNodeMeta = {
    title: input.title.trim(),
    icon: input.icon.trim() || undefined,
    accessMode: input.accessMode.trim() || 'permission',
    activePath: input.kind === 'entry' ? input.activePath.trim() || undefined : undefined,
    link: input.kind === 'external' ? input.externalLink.trim() || undefined : undefined,
    keepAlive: input.kind === 'entry' ? input.keepAlive : false,
    fixedTab: input.kind === 'entry' ? input.fixedTab : false,
    isFullPage: input.kind === 'entry' ? input.isFullPage : false,
    isHide: input.hidden,
  }

  return Object.fromEntries(
    Object.entries(meta).filter(([, value]) => value !== undefined && value !== ''),
  )
}

function buildMenuMutationPayload(input: MenuMutationDraft) {
  return {
    parent_id: input.parentId?.trim() || '',
    manage_group_id: input.manageGroupId?.trim() || '',
    space_key: input.spaceKey.trim() || 'default',
    kind: input.kind,
    path: input.path.trim(),
    name: input.name.trim(),
    component: input.kind === 'entry' ? input.component.trim() : '',
    title: input.title.trim(),
    icon: input.icon.trim(),
    sort_order: Number.isFinite(input.sortOrder) ? input.sortOrder : 0,
    hidden: input.hidden,
    meta: buildMenuMetaPayload(input),
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

export async function createMenu(input: MenuMutationDraft) {
  const result = await requestData<{ id?: string }>({
    method: 'POST',
    url: '/api/v1/menus',
    data: buildMenuMutationPayload(input),
  })

  return {
    id: `${result.id || ''}`.trim(),
  }
}

export async function updateMenu(menuId: string, input: MenuMutationDraft) {
  await requestData({
    method: 'PUT',
    url: `/api/v1/menus/${menuId}`,
    data: buildMenuMutationPayload(input),
  })
}

export async function fetchMenuDeletePreview(
  menuId: string,
  payload: { mode: MenuDeletePreview['mode']; targetParentId?: string | null },
) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/menus/${menuId}/delete-preview`,
    params: {
      mode: payload.mode,
      target_parent_id: payload.targetParentId?.trim() || undefined,
    },
  })

  return {
    mode: `${result.mode || 'single'}`.trim() as MenuDeletePreview['mode'],
    menuCount: Number(result.menu_count ?? result.menuCount ?? 0),
    childCount: Number(result.child_count ?? result.childCount ?? 0),
    affectedPageCount: Number(result.affected_page_count ?? result.affectedPageCount ?? 0),
    affectedRelationCount: Number(result.affected_relation_count ?? result.affectedRelationCount ?? 0),
  } satisfies MenuDeletePreview
}

export async function deleteMenu(
  menuId: string,
  payload: { mode: MenuDeletePreview['mode']; targetParentId?: string | null },
) {
  await requestData({
    method: 'DELETE',
    url: `/api/v1/menus/${menuId}`,
    params: {
      mode: payload.mode,
      target_parent_id: payload.targetParentId?.trim() || undefined,
    },
  })
}
