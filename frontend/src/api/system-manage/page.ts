import {
  request,
  PAGE_BASE,
  RUNTIME_BASE,
  normalizePageItem,
  normalizeMenuSpaceKey,
  normalizeRuntimeNavigationManifest,
  normalizePageUnregisteredItem,
  normalizePageBreadcrumbPreviewItem,
  normalizePageMenuOption,
  normalizePageAccessTraceResult
} from './_shared'

/** 获取页面列表 */
export function fetchGetPageList(params: Api.SystemManage.PageSearchParams) {
  const normalizedParams = {
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
  return request
    .get<Api.SystemManage.PageList>({
      url: PAGE_BASE,
      params: normalizedParams
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizePageItem)
    }))
}

export function fetchGetPageOptions(spaceKey?: string, appKey?: string) {
  return request
    .get<{ records: Api.SystemManage.PageItem[]; total: number }>({
      url: `${PAGE_BASE}/options`,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizePageItem),
      total: res?.total || 0
    }))
}

/**
 * 获取运行时导航清单。
 */
export function fetchGetRuntimeNavigation(spaceKey?: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RuntimeNavigationManifest>({
      url: `${RUNTIME_BASE}/navigation`,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res) => normalizeRuntimeNavigationManifest(res))
}

/** 获取运行时页面注册表 */
export function fetchGetRuntimePageList(spaceKey?: string, appKey?: string) {
  return request
    .get<{ records: Api.SystemManage.PageItem[]; total: number }>({
      url: `${PAGE_BASE}/runtime`,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizePageItem),
      total: res?.total || 0
    }))
}

/** 获取公开运行时页面注册表 */
export function fetchGetRuntimePublicPageList(spaceKey?: string, appKey?: string) {
  return request
    .get<{ records: Api.SystemManage.PageItem[]; total: number }>({
      url: `${PAGE_BASE}/runtime/public`,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizePageItem),
      total: res?.total || 0
    }))
}

/** 获取未注册页面 */
export function fetchGetPageUnregisteredList(appKey: string) {
  return request
    .get<{ records: Api.SystemManage.PageUnregisteredItem[]; total: number }>({
      url: `${PAGE_BASE}/unregistered`,
      params: { app_key: appKey }
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizePageUnregisteredItem),
      total: res?.total || 0
    }))
}

/** 同步页面注册表 */
export function fetchSyncPages(appKey: string) {
  return request
    .post<
      Api.SystemManage.PageSyncResult & {
        created_count?: number
        skipped_count?: number
        created_keys?: string[]
      }
    >({
      url: `${PAGE_BASE}/sync`,
      params: { app_key: appKey }
    })
    .then((res) => ({
      createdCount: res?.createdCount ?? res?.created_count ?? 0,
      skippedCount: res?.skippedCount ?? res?.skipped_count ?? 0,
      createdKeys: res?.createdKeys || res?.created_keys || []
    }))
}

/** 预览页面面包屑 */
export function fetchGetPageBreadcrumbPreview(id: string) {
  return request
    .get<{ items: Api.SystemManage.PageBreadcrumbPreviewItem[]; total: number }>({
      url: `${PAGE_BASE}/${id}/breadcrumb-preview`
    })
    .then((res) => ({
      items: (res?.items || []).map(normalizePageBreadcrumbPreviewItem),
      total: res?.total || 0
    }))
}

/** 获取页面详情 */
export function fetchGetPage(id: string) {
  return request
    .get<Api.SystemManage.PageItem>({
      url: `${PAGE_BASE}/${id}`
    })
    .then((res) => normalizePageItem(res))
}

/** 创建页面 */
export function fetchCreatePage(data: Api.SystemManage.PageSaveParams) {
  return request.post<Api.SystemManage.PageItem>({
    url: PAGE_BASE,
    data
  })
}

/** 更新页面 */
export function fetchUpdatePage(id: string, data: Api.SystemManage.PageSaveParams) {
  return request.put<Api.SystemManage.PageItem>({
    url: `${PAGE_BASE}/${id}`,
    data
  })
}

/** 删除页面 */
export function fetchDeletePage(id: string, appKey: string) {
  return request.del<void>({
    url: `${PAGE_BASE}/${id}`,
    params: { app_key: appKey }
  })
}

/** 获取页面上级菜单候选 */
export function fetchGetPageMenuOptions(spaceKey: string | undefined, appKey: string) {
  return request
    .get<{ records: Api.SystemManage.PageMenuOptionItem[]; total: number }>({
      url: `${PAGE_BASE}/menu-options`,
      params: {
        app_key: appKey,
        ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {})
      }
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizePageMenuOption),
      total: res?.total || 0
    }))
}

export function fetchGetPageAccessTrace(params: Api.SystemManage.PageAccessTraceParams) {
  return request
    .get<Api.SystemManage.PageAccessTraceResult>({
      url: `${PAGE_BASE}/access-trace`,
      params: {
        app_key: params.appKey,
        user_id: params.userId,
        collaboration_workspace_id: params.collaborationWorkspaceId,
        page_key: params.pageKey,
        page_keys: params.pageKeys,
        route_path: params.routePath,
        space_key: params.spaceKey
      }
    })
    .then((res) => normalizePageAccessTraceResult(res))
}
