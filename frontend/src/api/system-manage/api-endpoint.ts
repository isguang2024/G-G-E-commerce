import {
  v5Client,
  unwrap,
  normalizeApiEndpoint,
  normalizeApiEndpointCategory,
  normalizeUnregisteredApiRoute,
  toV5Body,
  type V5Query,
  type V5RequestBody
} from './_shared'

/** 获取 API 注册表 */
export async function fetchGetApiEndpointList(params: Api.SystemManage.APIEndpointSearchParams) {
  const normalizedParams: V5Query<'/api-endpoints', 'get'> = {
    permission_key: params?.permissionKey,
    permission_pattern: params?.permissionPattern,
    keyword: params?.keyword,
    method: params?.method,
    path: params?.path,
    status: params?.status,
    current: params?.current,
    size: params?.size,
    category_id: params?.categoryId,
    has_permission_key: params?.hasPermissionKey,
    has_category: params?.hasCategory
  }
  const res = await unwrap(v5Client.GET('/api-endpoints', { params: { query: normalizedParams } }))
  return {
    total: Number(res.total || 0),
    current: Number(res.current || 1),
    size: Number(res.size || 20),
    records: (res.records || []).map(normalizeApiEndpoint)
  } as Api.SystemManage.APIEndpointList
}

export async function fetchGetApiEndpointOverview(_appKey?: string) {
  const res = await unwrap(v5Client.GET('/api-endpoints/overview'))
  return {
    totalCount: Number(res.total_count ?? 0),
    uncategorizedCount: Number(res.uncategorized_count ?? 0),
    staleCount: Number(res.stale_count ?? 0),
    noPermissionCount: Number(res.no_permission_count ?? 0),
    sharedPermissionCount: Number(res.shared_permission_count ?? 0),
    crossContextSharedCount: Number(res.cross_context_shared_count ?? 0),
    categoryCounts: (res.category_counts || []).map((item) => ({
      categoryId: item.category_id || '',
      count: Number(item.count || 0)
    }))
  }
}

export async function fetchGetStaleApiEndpointList(params: { current?: number; size?: number }) {
  const query: V5Query<'/api-endpoints/stale', 'get'> = {
    current: params?.current,
    size: params?.size
  }
  const res = await unwrap(
    v5Client.GET('/api-endpoints/stale', {
      params: { query }
    })
  )
  return {
    total: Number(res.total || 0),
    current: Number(res.current || 1),
    size: Number(res.size || 20),
    records: (res?.records || []).map(normalizeApiEndpoint)
  } as Api.SystemManage.APIEndpointList
}

/** 同步 API 注册表 */
export async function fetchSyncApiEndpoints() {
  const { error } = await v5Client.POST('/api-endpoints/sync', {})
  if (error) throw error
}

export async function fetchCleanupStaleApiEndpoints(ids: string[]) {
  const body: V5RequestBody<'/api-endpoints/cleanup-stale', 'post'> = { ids }
  const res = await unwrap(v5Client.POST('/api-endpoints/cleanup-stale', { body }))
  return { deletedCount: Number(res.deleted_count ?? 0) }
}

export async function fetchUpdateApiEndpoint(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointItem>
) {
  const body: V5RequestBody<'/api-endpoints/{id}', 'put'> = {
    code: data.code || '',
    method: data.method || '',
    path: data.path || '',
    summary: data.summary || '',
    category_id: data.categoryId || null,
    status: data.status || 'normal',
    handler: data.handler || '',
    permission_keys: Array.isArray(data.permissionKeys) ? data.permissionKeys : []
  }
  const res = await unwrap(
    v5Client.PUT('/api-endpoints/{id}', {
      params: { path: { id } },
      body
    })
  )
  return normalizeApiEndpoint(res)
}

export async function fetchGetApiEndpointCategories() {
  const res = await unwrap(v5Client.GET('/api-endpoints/categories'))
  return {
    records: (res.records || []).map(normalizeApiEndpointCategory),
    total: Number(res.total || 0)
  }
}

export async function fetchGetUnregisteredApiRouteList(params: {
  current?: number
  size?: number
  method?: string
  path?: string
  keyword?: string
  only_no_meta?: boolean
}) {
  const query: V5Query<'/api-endpoints/unregistered', 'get'> = {
    current: params?.current,
    size: params?.size,
    method: params?.method,
    path: params?.path,
    keyword: params?.keyword,
    only_no_meta: params?.only_no_meta
  }
  const res = await unwrap(
    v5Client.GET('/api-endpoints/unregistered', {
      params: { query }
    })
  )
  return {
    total: Number(res.total || 0),
    current: Number(res.current || 1),
    size: Number(res.size || 20),
    records: (res?.records || []).map(normalizeUnregisteredApiRoute)
  } as Api.SystemManage.APIUnregisteredRouteList
}

export async function fetchCreateApiEndpointCategory(
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  const body: V5RequestBody<'/api-endpoints/categories', 'post'> = {
    code: data.code || '',
    name: data.name || '',
    name_en: data.nameEn || '',
    sort_order: data.sortOrder ?? 0,
    status: data.status || 'normal'
  }
  const res = await unwrap(v5Client.POST('/api-endpoints/categories', { body }))
  return normalizeApiEndpointCategory(res)
}

export async function fetchUpdateApiEndpointCategory(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  const body: V5RequestBody<'/api-endpoints/categories/{id}', 'put'> = {
    code: data.code || '',
    name: data.name || '',
    name_en: data.nameEn || '',
    sort_order: data.sortOrder ?? 0,
    status: data.status || 'normal'
  }
  const res = await unwrap(
    v5Client.PUT('/api-endpoints/categories/{id}', {
      params: { path: { id } },
      body
    })
  )
  return normalizeApiEndpointCategory(res)
}
