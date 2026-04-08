// Phase 4 slice 5: migrated to v5Client + openapi-fetch.
import {
  v5Client,
  unwrap,
  AppRouteRecord,
  normalizeMenuSpaceKey,
  normalizeRuntimeMenuTree
} from './_shared'

/** 获取菜单树 */
export async function fetchGetMenuList(spaceKey?: string, appKey?: string) {
  const query: any = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(v5Client.GET('/menus/tree', { params: { query } }))
  return (res?.records || []).map((item: any) => normalizeRuntimeMenuTree(item)) as AppRouteRecord[]
}

/** 获取完整菜单树 */
export async function fetchGetMenuTreeAll(spaceKey?: string, appKey?: string) {
  const query: any = { all: 1 }
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(v5Client.GET('/menus/tree', { params: { query } }))
  return (res?.records || []).map((item: any) => normalizeRuntimeMenuTree(item)) as AppRouteRecord[]
}

/** 创建菜单 */
export async function fetchCreateMenu(data: Api.SystemManage.MenuCreateParams, _config?: any) {
  const res: any = await unwrap(v5Client.POST('/menus', { body: data as any }))
  return { id: res?.id || '' }
}

/** 更新菜单 */
export async function fetchUpdateMenu(
  id: string,
  data: Api.SystemManage.MenuUpdateParams,
  _config?: any
) {
  const { error } = await v5Client.PUT('/menus/{id}', {
    params: { path: { id } },
    body: data as any
  })
  if (error) throw error
}

/** 删除菜单 */
export async function fetchDeleteMenu(id: string, params?: Api.SystemManage.MenuDeleteParams) {
  const query: any = params ? { ...params } : {}
  const { error } = await v5Client.DELETE('/menus/{id}', {
    params: { path: { id }, query }
  })
  if (error) throw error
}

export async function fetchGetMenuDeletePreview(
  id: string,
  params?: Api.SystemManage.MenuDeleteParams
) {
  const query: any = params ? { ...params } : {}
  const res: any = await unwrap(
    v5Client.GET('/menus/{id}/delete-preview', {
      params: { path: { id }, query }
    })
  )
  return {
    mode: `${res?.mode || 'single'}`.trim(),
    menuCount: Number(res?.menuCount ?? res?.menu_count ?? 0),
    childCount: Number(res?.childCount ?? res?.child_count ?? 0),
    affectedPageCount: Number(res?.affectedPageCount ?? res?.affected_page_count ?? 0),
    affectedRelationCount: Number(res?.affectedRelationCount ?? res?.affected_relation_count ?? 0)
  }
}

