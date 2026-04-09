import {
  v5Client,
  unwrap,
  normalizeApiEndpoint,
  normalizeApiEndpointCategory,
  normalizeUnregisteredApiRoute,
  toV5Body,
  toV5ListResponse,
  toV5Record,
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
  const res = toV5ListResponse(
    await unwrap(v5Client.GET('/api-endpoints', { params: { query: normalizedParams } }))
  )
  return {
    ...res,
    records: (res.records || []).map(normalizeApiEndpoint)
  } as Api.SystemManage.APIEndpointList
}

export async function fetchGetApiEndpointOverview(_appKey?: string) {
  const res = toV5Record(await unwrap(v5Client.GET('/api-endpoints/overview')))
  return {
    totalCount: Number(res?.totalCount ?? res?.total_count ?? 0),
    uncategorizedCount: Number(res?.uncategorizedCount ?? res?.uncategorized_count ?? 0),
    staleCount: Number(res?.staleCount ?? res?.stale_count ?? 0),
    noPermissionCount: Number(res?.noPermissionCount ?? res?.no_permission_count ?? 0),
    sharedPermissionCount: Number(
      res?.sharedPermissionCount ?? res?.shared_permission_count ?? 0
    ),
    crossContextSharedCount: Number(
      res?.crossContextSharedCount ?? res?.cross_context_shared_count ?? 0
    ),
    categoryCounts: ((res?.categoryCounts || res?.category_counts || []) as unknown[]).map((item: any) => ({
      categoryId: item?.categoryId || item?.category_id || '',
      count: item?.count || 0
    }))
  }
}

export async function fetchGetStaleApiEndpointList(params: { current?: number; size?: number }) {
  const query: V5Query<'/api-endpoints/stale', 'get'> = {
    current: params?.current,
    size: params?.size
  }
  const res = toV5ListResponse(
    await unwrap(
      v5Client.GET('/api-endpoints/stale', {
        params: { query }
      })
    )
  )
  return {
    ...res,
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
  const res = toV5Record(await unwrap(v5Client.POST('/api-endpoints/cleanup-stale', { body })))
  return { deletedCount: Number(res?.deletedCount ?? res?.deleted_count ?? 0) }
}

export function fetchUpdateApiEndpoint(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointItem>
) {
  const body: V5RequestBody<'/api-endpoints/{id}', 'put'> = toV5Body(data)
  return unwrap(
    v5Client.PUT('/api-endpoints/{id}', {
      params: { path: { id } },
      body
    })
  ) as unknown as Promise<Api.SystemManage.APIEndpointItem>
}

export async function fetchGetApiEndpointCategories() {
  const res = toV5ListResponse(await unwrap(v5Client.GET('/api-endpoints/categories')))
  return {
    records: (res.records || []).map(normalizeApiEndpointCategory),
    total: res.total || 0
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
  const res = toV5ListResponse(
    await unwrap(
      v5Client.GET('/api-endpoints/unregistered', {
        params: { query }
      })
    )
  )
  return {
    ...res,
    records: (res?.records || []).map(normalizeUnregisteredApiRoute)
  } as Api.SystemManage.APIUnregisteredRouteList
}

export async function fetchCreateApiEndpointCategory(
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  const body: V5RequestBody<'/api-endpoints/categories', 'post'> = toV5Body(data)
  const res = await unwrap(v5Client.POST('/api-endpoints/categories', { body }))
  return normalizeApiEndpointCategory(res)
}

export async function fetchUpdateApiEndpointCategory(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  const body: V5RequestBody<'/api-endpoints/categories/{id}', 'put'> = toV5Body(data)
  const res = await unwrap(
    v5Client.PUT('/api-endpoints/categories/{id}', {
      params: { path: { id } },
      body
    })
  )
  return normalizeApiEndpointCategory(res)
}
