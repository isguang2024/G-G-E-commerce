import {
  v5Client,
  unwrap,
  normalizeApiEndpoint,
  normalizeApiEndpointCategory,
  normalizeUnregisteredApiRoute
} from './_shared'

/** 获取 API 注册表 */
export async function fetchGetApiEndpointList(params: Api.SystemManage.APIEndpointSearchParams) {
  const normalizedParams: any = {
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
  const res: any = await unwrap(
    v5Client.GET('/api-endpoints', { params: { query: normalizedParams } })
  )
  return {
    ...res,
    records: (res?.records || []).map(normalizeApiEndpoint)
  } as Api.SystemManage.APIEndpointList
}

export async function fetchGetApiEndpointOverview(_appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/api-endpoints/overview', {
      params: { query: {} as any }
    })
  )
  return {
    totalCount: res?.totalCount ?? res?.total_count ?? 0,
    uncategorizedCount: res?.uncategorizedCount ?? res?.uncategorized_count ?? 0,
    staleCount: res?.staleCount ?? res?.stale_count ?? 0,
    noPermissionCount: res?.noPermissionCount ?? res?.no_permission_count ?? 0,
    sharedPermissionCount: res?.sharedPermissionCount ?? res?.shared_permission_count ?? 0,
    crossContextSharedCount:
      res?.crossContextSharedCount ?? res?.cross_context_shared_count ?? 0,
    categoryCounts: (res?.categoryCounts || res?.category_counts || []).map((item: any) => ({
      categoryId: item?.categoryId || item?.category_id || '',
      count: item?.count || 0
    }))
  }
}

export async function fetchGetStaleApiEndpointList(params: { current?: number; size?: number }) {
  const res: any = await unwrap(
    v5Client.GET('/api-endpoints/stale', {
      params: { query: { current: params?.current, size: params?.size } as any }
    })
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
  const res: any = await unwrap(
    v5Client.POST('/api-endpoints/cleanup-stale', { body: { ids } as any })
  )
  return { deletedCount: res?.deletedCount ?? res?.deleted_count ?? 0 }
}

export function fetchUpdateApiEndpoint(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointItem>
) {
  return unwrap(
    v5Client.PUT('/api-endpoints/{id}', {
      params: { path: { id } },
      body: data as any
    })
  ) as unknown as Promise<Api.SystemManage.APIEndpointItem>
}

export async function fetchGetApiEndpointCategories() {
  const res: any = await unwrap(
    v5Client.GET('/api-endpoints/categories', { params: { query: {} as any } })
  )
  return {
    records: (res?.records || []).map(normalizeApiEndpointCategory),
    total: res?.total || 0
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
  const res: any = await unwrap(
    v5Client.GET('/api-endpoints/unregistered', {
      params: {
        query: {
          current: params?.current,
          size: params?.size,
          method: params?.method,
          path: params?.path,
          keyword: params?.keyword,
          only_no_meta: params?.only_no_meta
        } as any
      }
    })
  )
  return {
    ...res,
    records: (res?.records || []).map(normalizeUnregisteredApiRoute)
  } as Api.SystemManage.APIUnregisteredRouteList
}

export async function fetchCreateApiEndpointCategory(
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  const res: any = await unwrap(
    v5Client.POST('/api-endpoints/categories', { body: data as any })
  )
  return normalizeApiEndpointCategory(res)
}

export async function fetchUpdateApiEndpointCategory(
  id: string,
  data: Partial<Api.SystemManage.APIEndpointCategoryItem>
) {
  const res: any = await unwrap(
    v5Client.PUT('/api-endpoints/categories/{id}', {
      params: { path: { id } },
      body: data as any
    })
  )
  return normalizeApiEndpointCategory(res)
}
