import {
  request,
  MENU_BASE,
  MENU_BACKUP_BASE,
  AppRouteRecord,
  normalizeMenuSpaceKey,
  normalizeRuntimeMenuTree,
  normalizeMenuManageGroup,
  normalizeMenuBackupItem
} from './_shared'

/** 获取菜单树 */
export function fetchGetMenuList(spaceKey?: string, appKey?: string) {
  return request
    .get<AppRouteRecord[]>({
      url: `${MENU_BASE}/tree`,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res) => (res || []).map((item: any) => normalizeRuntimeMenuTree(item)))
}

/** 获取完整菜单树 */
export function fetchGetMenuTreeAll(spaceKey?: string, appKey?: string) {
  return request
    .get<AppRouteRecord[]>({
      url: `${MENU_BASE}/tree`,
      params: {
        all: 1,
        ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => (res || []).map((item: any) => normalizeRuntimeMenuTree(item)))
}

/** 创建菜单 */
export function fetchCreateMenu(data: Api.SystemManage.MenuCreateParams, config?: any) {
  return request.post<{ id: string }>({
    url: MENU_BASE,
    data,
    ...config
  })
}

/** 更新菜单 */
export function fetchUpdateMenu(id: string, data: Api.SystemManage.MenuUpdateParams, config?: any) {
  return request.put<void>({
    url: `${MENU_BASE}/${id}`,
    data,
    ...config
  })
}

/** 获取菜单管理分组 */
export function fetchGetMenuManageGroups() {
  return request
    .get<Api.SystemManage.MenuManageGroupItem[]>({
      url: `${MENU_BASE}/groups`
    })
    .then((res) => (res || []).map(normalizeMenuManageGroup))
}

/** 创建菜单管理分组 */
export function fetchCreateMenuManageGroup(data: Api.SystemManage.MenuManageGroupSaveParams) {
  return request.post<{ id: string }>({
    url: `${MENU_BASE}/groups`,
    data
  })
}

/** 更新菜单管理分组 */
export function fetchUpdateMenuManageGroup(
  id: string,
  data: Api.SystemManage.MenuManageGroupSaveParams
) {
  return request.put<void>({
    url: `${MENU_BASE}/groups/${id}`,
    data
  })
}

/** 删除菜单管理分组 */
export function fetchDeleteMenuManageGroup(id: string) {
  return request.del<void>({
    url: `${MENU_BASE}/groups/${id}`
  })
}

/** 删除菜单 */
export function fetchDeleteMenu(id: string, params?: Api.SystemManage.MenuDeleteParams) {
  return request.del<void>({
    url: `${MENU_BASE}/${id}`,
    params: params
      ? {
          ...params,
          target_parent_id: params.target_parent_id || params.targetParentId || undefined
        }
      : undefined
  })
}

export function fetchGetMenuDeletePreview(id: string, params?: Api.SystemManage.MenuDeleteParams) {
  return request
    .get<Api.SystemManage.MenuDeletePreviewItem>({
      url: `${MENU_BASE}/${id}/delete-preview`,
      params: params
        ? {
            ...params,
            target_parent_id: params.target_parent_id || params.targetParentId || undefined
          }
        : undefined
    })
    .then((res) => ({
      mode: `${res?.mode || 'single'}`.trim(),
      menuCount: Number(res?.menuCount ?? res?.menu_count ?? 0),
      childCount: Number(res?.childCount ?? res?.child_count ?? 0),
      affectedPageCount: Number(res?.affectedPageCount ?? res?.affected_page_count ?? 0),
      affectedRelationCount: Number(res?.affectedRelationCount ?? res?.affected_relation_count ?? 0)
    }))
}

/** 更新菜单排序（全量重排） */
export function fetchUpdateMenuSort(data: { id: string; sort_order: number }[]) {
  return request.put<void>({
    url: `${MENU_BASE}/sort`,
    data
  })
}

/** 根据父级ID更新子节点排序（全量重排） */
export function fetchUpdateMenuSortByParent(parentId: string | null, menuIds: string[]) {
  return request.put<void>({
    url: `${MENU_BASE}/sort-by-parent`,
    data: {
      parent_id: parentId,
      menu_ids: menuIds
    }
  })
}

/** 创建菜单备份 */
export function fetchCreateMenuBackup(data: Api.SystemManage.MenuBackupCreateParams) {
  return request.post<void>({
    url: MENU_BACKUP_BASE,
    data
  })
}

/** 获取菜单备份列表 */
export function fetchGetMenuBackupList(spaceKey?: string, appKey?: string) {
  return request
    .get<Api.SystemManage.MenuBackupItem[]>({
      url: MENU_BACKUP_BASE,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res) => (res || []).map((item: any) => normalizeMenuBackupItem(item)))
}

/** 删除菜单备份 */
export function fetchDeleteMenuBackup(id: string, appKey: string) {
  return request.del<void>({
    url: `${MENU_BACKUP_BASE}/${id}`,
    params: { app_key: appKey }
  })
}

/** 恢复菜单备份 */
export function fetchRestoreMenuBackup(id: string, appKey: string) {
  return request.post<void>({
    url: `${MENU_BACKUP_BASE}/${id}/restore`,
    params: { app_key: appKey }
  })
}
