// Phase 4: page 域迁移至 v5Client + openapi-fetch。
import {
  v5Client,
  unwrap,
  normalizePageItem,
  normalizeMenuSpaceKey,
  normalizeRuntimeNavigationManifest,
  normalizePageUnregisteredItem,
  normalizePageBreadcrumbPreviewItem,
  normalizePageMenuOption,
  normalizePageAccessTraceResult,
  type V5Query,
  type V5RequestBody
} from './_shared'

/** 获取页面列表 */
export async function fetchGetPageList(params: Api.SystemManage.PageSearchParams) {
  const query: V5Query<'/pages', 'get'> = {
    current: params?.current,
    size: params?.size,
    app_key: params?.appKey || '',
    keyword: params?.keyword,
    space_key: params?.spaceKey,
    status: params?.status
  }
  const res = await unwrap(v5Client.GET('/pages', { params: { query } }))
  return {
    ...res,
    records: (res.records || []).map(normalizePageItem)
  } as Api.SystemManage.PageList
}

export async function fetchGetPageOptions(spaceKey?: string, appKey?: string) {
  const query: V5Query<'/pages/options', 'get'> = { app_key: appKey || '' }
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  const res = await unwrap(v5Client.GET('/pages/options', { params: { query } }))
  return {
    records: (res.records || []).map(normalizePageItem),
    total: res.total || 0
  }
}

/**
 * 获取运行时导航清单。
 */
export async function fetchGetRuntimeNavigation(spaceKey?: string, appKey?: string) {
  const query: V5Query<'/runtime/navigation', 'get'> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(v5Client.GET('/runtime/navigation', { params: { query } }))
  return normalizeRuntimeNavigationManifest(res)
}

/** 获取运行时页面注册表 */
export async function fetchGetRuntimePageList(spaceKey?: string, appKey?: string) {
  const query: V5Query<'/pages/runtime', 'get'> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(v5Client.GET('/pages/runtime', { params: { query } }))
  return {
    records: (res.records || []).map(normalizePageItem),
    total: res.total || 0
  }
}

/** 获取公开运行时页面注册表 */
export async function fetchGetRuntimePublicPageList(spaceKey?: string, appKey?: string) {
  const query: V5Query<'/pages/runtime/public', 'get'> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(v5Client.GET('/pages/runtime/public', { params: { query } }))
  return {
    records: (res.records || []).map(normalizePageItem),
    total: res.total || 0
  }
}

/** 获取未注册页面 */
export async function fetchGetPageUnregisteredList(appKey: string) {
  const query: V5Query<'/pages/unregistered', 'get'> = { app_key: appKey }
  const res = await unwrap(v5Client.GET('/pages/unregistered', { params: { query } }))
  return {
    records: (res.records || []).map(normalizePageUnregisteredItem),
    total: res.total || 0
  }
}

/** 同步页面注册表 */
export async function fetchSyncPages(appKey: string) {
  const query: V5Query<'/pages/sync', 'post'> = { app_key: appKey }
  const res = await unwrap(v5Client.POST('/pages/sync', { params: { query } }))
  return {
    createdCount: res?.created_count ?? 0,
    skippedCount: res?.skipped_count ?? 0,
    createdKeys: res?.created_keys || []
  }
}

/** 预览页面面包屑 */
export async function fetchGetPageBreadcrumbPreview(id: string, appKey = '') {
  const res = await unwrap(
    v5Client.GET('/pages/{id}/breadcrumb-preview', {
      params: { path: { id }, query: { app_key: appKey } }
    })
  )
  return {
    items: (res.records || []).map(normalizePageBreadcrumbPreviewItem),
    total: res.total || 0
  }
}

/** 获取页面详情 */
export async function fetchGetPage(id: string, appKey = '') {
  const res = await unwrap(
    v5Client.GET('/pages/{id}', { params: { path: { id }, query: { app_key: appKey } } })
  )
  return normalizePageItem(res)
}

/** 创建页面 */
export async function fetchCreatePage(data: Api.SystemManage.PageSaveParams) {
  const query: V5Query<'/pages', 'post'> = { app_key: data.app_key }
  const body: V5RequestBody<'/pages', 'post'> = {
    page_key: data.page_key,
    name: data.name,
    route_name: data.route_name,
    route_path: data.route_path,
    component: data.component,
    page_type: data.page_type,
    source: data.source,
    module_key: data.module_key,
    sort_order: data.sort_order,
    parent_menu_id: data.parent_menu_id,
    parent_page_key: data.parent_page_key,
    display_group_key: data.display_group_key,
    active_menu_path: data.active_menu_path,
    breadcrumb_mode: data.breadcrumb_mode,
    access_mode: data.access_mode,
    permission_key: data.permission_key,
    inherit_permission: data.inherit_permission,
    keep_alive: data.keep_alive,
    is_full_page: data.is_full_page,
    space_keys: data.space_keys,
    visibility_scope: data.visibility_scope,
    remote_binding: data.remote_binding,
    status: data.status,
    meta: data.meta
  }
  const res = await unwrap(
    v5Client.POST('/pages', {
      params: { query },
      body
    })
  )
  return normalizePageItem(res)
}

/** 更新页面 */
export async function fetchUpdatePage(id: string, data: Api.SystemManage.PageSaveParams) {
  const query: V5Query<'/pages/{id}', 'put'> = { app_key: data.app_key }
  const body: V5RequestBody<'/pages/{id}', 'put'> = {
    page_key: data.page_key,
    name: data.name,
    route_name: data.route_name,
    route_path: data.route_path,
    component: data.component,
    page_type: data.page_type,
    source: data.source,
    module_key: data.module_key,
    sort_order: data.sort_order,
    parent_menu_id: data.parent_menu_id,
    parent_page_key: data.parent_page_key,
    display_group_key: data.display_group_key,
    active_menu_path: data.active_menu_path,
    breadcrumb_mode: data.breadcrumb_mode,
    access_mode: data.access_mode,
    permission_key: data.permission_key,
    inherit_permission: data.inherit_permission,
    keep_alive: data.keep_alive,
    is_full_page: data.is_full_page,
    space_keys: data.space_keys,
    visibility_scope: data.visibility_scope,
    remote_binding: data.remote_binding,
    status: data.status,
    meta: data.meta
  }
  const res = await unwrap(
    v5Client.PUT('/pages/{id}', {
      params: { path: { id }, query },
      body
    })
  )
  return normalizePageItem(res)
}

/** 删除页面 */
export async function fetchDeletePage(id: string, appKey: string) {
  const query: V5Query<'/pages/{id}', 'delete'> = { app_key: appKey }
  const { error } = await v5Client.DELETE('/pages/{id}', {
    params: { path: { id }, query }
  })
  if (error) throw error
}

/** 获取页面上级菜单候选 */
export async function fetchGetPageMenuOptions(spaceKey: string | undefined, appKey: string) {
  const query: V5Query<'/pages/menu-options', 'get'> = { app_key: appKey }
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  const res = await unwrap(v5Client.GET('/pages/menu-options', { params: { query } }))
  return {
    records: (res.records || []).map(normalizePageMenuOption),
    total: res.total || 0
  }
}

export async function fetchGetPageAccessTrace(params: Api.SystemManage.PageAccessTraceParams) {
  const query: V5Query<'/pages/access-trace', 'get'> = {
    app_key: params.appKey || '',
    user_id: params.userId,
    ...(params.collaborationWorkspaceId
      ? { collaboration_workspace_id: params.collaborationWorkspaceId }
      : {}),
    ...(params.pageKey ? { page_key: params.pageKey } : {}),
    ...(params.pageKeys ? { page_keys: params.pageKeys } : {}),
    ...(params.routePath ? { route_path: params.routePath } : {}),
    ...(params.spaceKey ? { space_key: params.spaceKey } : {})
  }
  const res = await unwrap(v5Client.GET('/pages/access-trace', { params: { query } }))
  return normalizePageAccessTraceResult(res)
}
