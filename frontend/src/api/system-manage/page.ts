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
  normalizePageAccessTraceResult
} from './_shared'

/** 获取页面列表 */
export async function fetchGetPageList(params: Api.SystemManage.PageSearchParams) {
  const query: Record<string, any> = {
    current: params?.current,
    size: params?.size,
    app_key: params?.appKey,
    keyword: params?.keyword,
    page_type: params?.pageType,
    module_key: params?.moduleKey,
    parent_menu_id: params?.parentMenuId,
    space_key: params?.spaceKey,
    access_mode: params?.accessMode,
    source: params?.source,
    status: params?.status
  }
  const res: any = await unwrap(v5Client.GET('/pages', { params: { query } as any }))
  return {
    ...res,
    records: (res?.records || []).map(normalizePageItem)
  } as Api.SystemManage.PageList
}

export async function fetchGetPageOptions(spaceKey?: string, appKey?: string) {
  const query: Record<string, any> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(v5Client.GET('/pages/options', { params: { query } as any }))
  return {
    records: (res?.records || []).map(normalizePageItem),
    total: res?.total || 0
  }
}

/**
 * 获取运行时导航清单。
 */
export async function fetchGetRuntimeNavigation(spaceKey?: string, appKey?: string) {
  const query: Record<string, any> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(
    v5Client.GET('/runtime/navigation', { params: { query } as any })
  )
  return normalizeRuntimeNavigationManifest(res)
}

/** 获取运行时页面注册表 */
export async function fetchGetRuntimePageList(spaceKey?: string, appKey?: string) {
  const query: Record<string, any> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(v5Client.GET('/pages/runtime', { params: { query } as any }))
  return {
    records: (res?.records || []).map(normalizePageItem),
    total: res?.total || 0
  }
}

/** 获取公开运行时页面注册表 */
export async function fetchGetRuntimePublicPageList(spaceKey?: string, appKey?: string) {
  const query: Record<string, any> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(
    v5Client.GET('/pages/runtime/public', { params: { query } as any })
  )
  return {
    records: (res?.records || []).map(normalizePageItem),
    total: res?.total || 0
  }
}

/** 获取未注册页面 */
export async function fetchGetPageUnregisteredList(appKey: string) {
  const res: any = await unwrap(
    v5Client.GET('/pages/unregistered', { params: { query: { app_key: appKey } as any } })
  )
  return {
    records: (res?.records || []).map(normalizePageUnregisteredItem),
    total: res?.total || 0
  }
}

/** 同步页面注册表 */
export async function fetchSyncPages(appKey: string) {
  const res: any = await unwrap(
    v5Client.POST('/pages/sync', { params: { query: { app_key: appKey } as any } } as any)
  )
  return {
    createdCount: res?.createdCount ?? res?.created_count ?? 0,
    skippedCount: res?.skippedCount ?? res?.skipped_count ?? 0,
    createdKeys: res?.createdKeys || res?.created_keys || []
  }
}

/** 预览页面面包屑 */
export async function fetchGetPageBreadcrumbPreview(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/pages/{id}/breadcrumb-preview', {
      params: { path: { id }, query: {} as any }
    })
  )
  return {
    items: (res?.items || []).map(normalizePageBreadcrumbPreviewItem),
    total: res?.total || 0
  }
}

/** 获取页面详情 */
export async function fetchGetPage(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/pages/{id}', { params: { path: { id }, query: {} as any } })
  )
  return normalizePageItem(res)
}

/** 创建页面 */
export async function fetchCreatePage(data: Api.SystemManage.PageSaveParams) {
  const appKey = (data as any)?.app_key || (data as any)?.appKey || ''
  const res: any = await unwrap(
    v5Client.POST('/pages', {
      params: { query: { app_key: appKey } as any },
      body: data as any
    })
  )
  return res as Api.SystemManage.PageItem
}

/** 更新页面 */
export async function fetchUpdatePage(id: string, data: Api.SystemManage.PageSaveParams) {
  const appKey = (data as any)?.app_key || (data as any)?.appKey || ''
  const res: any = await unwrap(
    v5Client.PUT('/pages/{id}', {
      params: { path: { id }, query: { app_key: appKey } as any },
      body: data as any
    })
  )
  return res as Api.SystemManage.PageItem
}

/** 删除页面 */
export async function fetchDeletePage(id: string, appKey: string) {
  const { error } = await v5Client.DELETE('/pages/{id}', {
    params: { path: { id }, query: { app_key: appKey } as any }
  })
  if (error) throw error
}

/** 获取页面上级菜单候选 */
export async function fetchGetPageMenuOptions(spaceKey: string | undefined, appKey: string) {
  const query: Record<string, any> = { app_key: appKey }
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  const res: any = await unwrap(
    v5Client.GET('/pages/menu-options', { params: { query } as any })
  )
  return {
    records: (res?.records || []).map(normalizePageMenuOption),
    total: res?.total || 0
  }
}

export async function fetchGetPageAccessTrace(params: Api.SystemManage.PageAccessTraceParams) {
  const raw: Record<string, any> = {
    app_key: params.appKey,
    user_id: params.userId,
    collaboration_workspace_id: params.collaborationWorkspaceId,
    page_key: params.pageKey,
    page_keys: params.pageKeys,
    route_path: params.routePath,
    space_key: params.spaceKey
  }
  // 空串会让 ogen 的 uuid 解析报 400，直接剔除空值
  const query: Record<string, any> = {}
  for (const [k, v] of Object.entries(raw)) {
    if (v !== undefined && v !== null && `${v}`.trim() !== '') query[k] = v
  }
  const res: any = await unwrap(
    v5Client.GET('/pages/access-trace', { params: { query } as any })
  )
  return normalizePageAccessTraceResult(res)
}
