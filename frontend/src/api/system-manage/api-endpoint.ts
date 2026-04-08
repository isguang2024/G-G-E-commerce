import {
  request,
  API_ENDPOINT_BASE,
  normalizeApiEndpoint,
  normalizeApiEndpointCategory,
  normalizeUnregisteredApiRoute,
  normalizeUnregisteredApiScanConfig
} from './_shared'

/** 获取 API 注册表 */
export function fetchGetApiEndpointList(params: Api.SystemManage.APIEndpointSearchParams) {
  const normalizedParams = {
    app_key: params?.appKey,
    app_scope: params?.appScope,
    permission_key: params?.permissionKey,
    permission_pattern: params?.permissionPattern,
    keyword: params?.keyword,
    method: params?.method,
    path: params?.path,
    status: params?.status,
    current: params?.current,
    size: params?.size,
    feature_kind: params?.featureKind,
    category_id: params?.categoryId,
    context_scope: params?.contextScope,
    source: params?.source,
    has_permission_key: params?.hasPermissionKey,
    has_category: params?.hasCategory
  }
  return request
    .get<Api.SystemManage.APIEndpointList>({
      url: API_ENDPOINT_BASE,
      params: normalizedParams
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeApiEndpoint)
    }))
}

export function fetchGetApiEndpointOverview(appKey?: string) {
  return request
    .get<
      Api.SystemManage.APIEndpointOverview & {
        total_count?: number
        uncategorized_count?: number
        stale_count?: number
        no_permission_count?: number
        shared_permission_count?: number
        cross_context_shared_count?: number
        category_counts?: any[]
      }
    >({
      url: `${API_ENDPOINT_BASE}/overview`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      totalCount: res?.totalCount ?? res?.total_count ?? 0,
      uncategorizedCount: res?.uncategorizedCount ?? res?.uncategorized_count ?? 0,
      staleCount: res?.staleCount ?? res?.stale_count ?? 0,
      noPermissionCount: res?.noPermissionCount ?? res?.no_permission_count ?? 0,
      sharedPermissionCount: res?.sharedPermissionCount ?? res?.shared_permission_count ?? 0,
      crossContextSharedCount: res?.crossContextSharedCount ?? res?.cross_context_shared_count ?? 0,
      categoryCounts: (res?.categoryCounts || res?.category_counts || []).map((item: any) => ({
        categoryId: item?.categoryId || item?.category_id || '',
        count: item?.count || 0
      }))
    }))
}

export function fetchGetStaleApiEndpointList(params: { current?: number; size?: number }) {
  return request
    .get<Api.SystemManage.APIEndpointList>({
      url: `${API_ENDPOINT_BASE}/stale`,
      params: {
        current: params?.current,
        size: params?.size
      }
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeApiEndpoint)
    }))
}

/** 同步 API 注册表 */
export function fetchSyncApiEndpoints() {
  return request.post<void>({
    url: `${API_ENDPOINT_BASE}/sync`
  })
}

export function fetchCleanupStaleApiEndpoints(ids: string[]) {
  return request
    .post<{ deleted_count?: number; deletedCount?: number }>({
      url: `${API_ENDPOINT_BASE}/cleanup-stale`,
      data: { ids }
    })
    .then((res) => ({
      deletedCount: res?.deletedCount ?? res?.deleted_count ?? 0
    }))
}

export function fetchCreateApiEndpoint(data: Partial<Api.SystemManage.APIEndpointItem>) {
  return request.post<Api.SystemManage.APIEndpointItem>({
    url: API_ENDPOINT_BASE,
    data
  })
}

export function fetchUpdateApiEndpoint(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointItem>
) {
  return request.put<Api.SystemManage.APIEndpointItem>({
    url: `${API_ENDPOINT_BASE}/${id}`,
    data
  })
}

export function fetchUpdateApiEndpointContextScope(id: string, contextScope: string) {
  return request.put<Api.SystemManage.APIEndpointItem>({
    url: `${API_ENDPOINT_BASE}/${id}/context-scope`,
    data: { context_scope: contextScope }
  })
}

export function fetchGetApiEndpointCategories() {
  return request
    .get<{ records: Api.SystemManage.APIEndpointCategoryItem[]; total: number }>({
      url: `${API_ENDPOINT_BASE}/categories`
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeApiEndpointCategory),
      total: res?.total || 0
    }))
}

export function fetchGetUnregisteredApiRouteList(params: {
  current?: number
  size?: number
  method?: string
  path?: string
  keyword?: string
  only_no_meta?: boolean
}) {
  return request
    .get<Api.SystemManage.APIUnregisteredRouteList>({
      url: `${API_ENDPOINT_BASE}/unregistered`,
      params: {
        current: params?.current,
        size: params?.size,
        method: params?.method,
        path: params?.path,
        keyword: params?.keyword,
        only_no_meta: params?.only_no_meta
      }
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeUnregisteredApiRoute)
    }))
}

export function fetchGetUnregisteredApiScanConfig() {
  return request
    .get<Api.SystemManage.APIUnregisteredScanConfig>({
      url: `${API_ENDPOINT_BASE}/unregistered/scan-config`
    })
    .then((res) => normalizeUnregisteredApiScanConfig(res))
}

export function fetchSaveUnregisteredApiScanConfig(
  data: Partial<Api.SystemManage.APIUnregisteredScanConfig>
) {
  return request
    .put<Api.SystemManage.APIUnregisteredScanConfig>({
      url: `${API_ENDPOINT_BASE}/unregistered/scan-config`,
      data: {
        enabled: data.enabled,
        frequency_minutes: data.frequencyMinutes,
        default_category_id: data.defaultCategoryId,
        default_permission_key: data.defaultPermissionKey,
        mark_as_no_permission: data.markAsNoPermission
      }
    })
    .then((res) => normalizeUnregisteredApiScanConfig(res))
}

export function fetchCreateApiEndpointCategory(
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  return request
    .post<Api.SystemManage.APIEndpointCategoryItem>({
      url: `${API_ENDPOINT_BASE}/categories`,
      data
    })
    .then((res) => normalizeApiEndpointCategory(res))
}

export function fetchUpdateApiEndpointCategory(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  return request
    .put<Api.SystemManage.APIEndpointCategoryItem>({
      url: `${API_ENDPOINT_BASE}/categories/${id}`,
      data
    })
    .then((res) => normalizeApiEndpointCategory(res))
}
