// Phase 4 slice 5: migrated to v5Client + openapi-fetch.
import {
  v5Client,
  unwrap,
  AppRouteRecord,
  normalizeMenuSpaceKey,
  normalizeRuntimeMenuTree,
  type V5Query,
  type V5RequestBody
} from './_shared'

/** 获取菜单树 */
export async function fetchGetMenuList(spaceKey?: string, appKey?: string) {
  const query: V5Query<'/menus/tree', 'get'> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(v5Client.GET('/menus/tree', { params: { query } }))
  return (res.records || []).map((item) => normalizeRuntimeMenuTree(item)) as AppRouteRecord[]
}

/** 获取完整菜单树 */
export async function fetchGetMenuTreeAll(spaceKey?: string, appKey?: string) {
  const query: V5Query<'/menus/tree', 'get'> = { all: '1' }
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(v5Client.GET('/menus/tree', { params: { query } }))
  return (res.records || []).map((item) => normalizeRuntimeMenuTree(item)) as AppRouteRecord[]
}

/** 创建菜单 */
export async function fetchCreateMenu(data: Api.SystemManage.MenuCreateParams, _config?: unknown) {
  const body: V5RequestBody<'/menus', 'post'> = {
    ...data,
    name: data.name || '',
    kind: data.kind || 'menu'
  }
  const res = await unwrap(v5Client.POST('/menus', { body }))
  return { success: Boolean(res.success) }
}

/** 更新菜单 */
export async function fetchUpdateMenu(
  id: string,
  data: Api.SystemManage.MenuUpdateParams,
  _config?: unknown
) {
  const body: V5RequestBody<'/menus/{id}', 'put'> = {
    ...data,
    name: data.name || '',
    kind: data.kind || 'menu'
  }
  const { error } = await v5Client.PUT('/menus/{id}', {
    params: { path: { id } },
    body
  })
  if (error) throw error
}

/** 删除菜单 */
export async function fetchDeleteMenu(id: string, params?: Api.SystemManage.MenuDeleteParams) {
  void params
  const { error } = await v5Client.DELETE('/menus/{id}', { params: { path: { id } } })
  if (error) throw error
}

export async function fetchGetMenuDeletePreview(
  id: string,
  params?: Api.SystemManage.MenuDeleteParams
) {
  void params
  const res = await unwrap(
    v5Client.GET('/menus/{id}/delete-preview', {
      params: { path: { id } }
    })
  )
  return {
    mode: `${res.mode || 'single'}`.trim(),
    menuCount: Number(res.menu_count ?? 0),
    childCount: Number(res.child_count ?? 0),
    affectedPageCount: Number(res.affected_page_count ?? 0),
    affectedRelationCount: Number(res.affected_relation_count ?? 0)
  }
}

