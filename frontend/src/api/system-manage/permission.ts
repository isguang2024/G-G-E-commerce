// Phase 4: permission/feature-package 域迁移至 v5Client + openapi-fetch。
import {
  v5Client,
  unwrap,
  normalizePermissionAction,
  normalizePermissionAuditSummary,
  normalizeApiEndpoint,
  normalizePermissionActionConsumers,
  normalizeRiskAudit,
  normalizePermissionGroup,
  normalizeFeaturePackage,
  normalizeFeaturePackageRelationTree,
  normalizeRefreshStats,
  normalizeCollaborationWorkspace,
  toV5Body,
  toV5Record,
  type V5Query,
  type V5RequestBody
} from './_shared'

type PermissionBatchTemplateRecord = {
  id?: string
  name?: string
  description?: string
  payload?: unknown
  created_by?: string
  createdBy?: string
  created_at?: string
  createdAt?: string
  updated_at?: string
  updatedAt?: string
}

type FeaturePackageVersionRecord = {
  id?: string
  package_id?: string
  packageId?: string
  version_no?: number
  versionNo?: number
  change_type?: string
  changeType?: string
  snapshot?: unknown
  operator_id?: string
  operatorId?: string
  request_id?: string
  requestId?: string
  created_at?: string
  createdAt?: string
}

/** 获取功能权限列表 */
export async function fetchGetPermissionActionList(
  params: Api.SystemManage.PermissionActionSearchParams
) {
  const query: V5Query<'/permission-actions', 'get'> = {
    current: params?.current,
    size: params?.size,
    keyword: params?.keyword,
    group_id: params?.featureGroupId || params?.moduleGroupId,
    status: params?.status
  }
  const res = await unwrap(v5Client.GET('/permission-actions', { params: { query } }))
  return {
    ...res,
    records: (res.records || []).map(normalizePermissionAction),
    auditSummary: normalizePermissionAuditSummary(res.audit_summary || {})
  } as Api.SystemManage.PermissionActionList
}

export async function fetchGetPermissionActionOptions(
  params?: Api.SystemManage.PermissionActionSearchParams
) {
  const query: V5Query<'/permission-actions/options', 'get'> = {
    keyword: params?.keyword,
    group_id: params?.featureGroupId || params?.moduleGroupId
  }
  const res = await unwrap(v5Client.GET('/permission-actions/options', { params: { query } }))
  return {
    records: (res.records || []).map(normalizePermissionAction),
    total: res.total || 0
  }
}

/** 获取功能权限详情 */
export async function fetchGetPermissionAction(id: string) {
  const res = await unwrap(v5Client.GET('/permission-actions/{id}', { params: { path: { id } } }))
  return normalizePermissionAction(res)
}

/** 获取功能权限关联接口 */
export async function fetchGetPermissionActionEndpoints(id: string) {
  const res = await unwrap(
    v5Client.GET('/permission-actions/{id}/endpoints', { params: { path: { id } } })
  )
  return {
    records: (res.records || []).map(normalizeApiEndpoint),
    total: res.total || 0
  }
}

/** 获取功能权限消费明细（API/页面/功能包/角色） */
export async function fetchGetPermissionActionConsumers(id: string) {
  const res = await unwrap(
    v5Client.GET('/permission-actions/{id}/consumers', { params: { path: { id } } })
  )
  return normalizePermissionActionConsumers(res)
}

/** 新增功能权限关联接口 */
export async function fetchAddPermissionActionEndpoint(id: string, endpointCode: string) {
  const body: V5RequestBody<'/permission-actions/{id}/endpoints', 'post'> = {
    endpoint_code: endpointCode
  }
  const { error } = await v5Client.POST('/permission-actions/{id}/endpoints', {
    params: { path: { id } },
    body
  })
  if (error) throw error
}

/** 删除功能权限关联接口 */
export async function fetchDeletePermissionActionEndpoint(id: string, endpointCode: string) {
  const { error } = await v5Client.DELETE('/permission-actions/{id}/endpoints/{endpointCode}', {
    params: { path: { id, endpointCode } }
  })
  if (error) throw error
}

export async function fetchCleanupUnusedPermissionActions() {
  const res = toV5Record(await unwrap(v5Client.POST('/permission-actions/cleanup-unused')))
  const deletedKeys = Array.isArray(res?.deletedKeys)
    ? res.deletedKeys
    : Array.isArray(res?.deleted_keys)
      ? res.deleted_keys
      : []
  return {
    deletedCount: Number(res?.deletedCount ?? res?.deleted_count ?? 0),
    deletedKeys: deletedKeys.map((value) => `${value || ''}`.trim()).filter(Boolean)
  }
}

/** 创建功能权限 */
export async function fetchCreatePermissionAction(
  data: Api.SystemManage.PermissionActionCreateParams
) {
  const body: V5RequestBody<'/permission-actions', 'post'> = {
    action_key: data.permission_key,
    name: data.name,
    description: data.description,
    status: data.status,
    group_id: data.feature_group_id || data.module_group_id || null
  }
  const res = toV5Record(await unwrap(v5Client.POST('/permission-actions', { body })))
  return { id: `${res.id || ''}` }
}

/** 更新功能权限 */
export async function fetchUpdatePermissionAction(
  id: string,
  data: Api.SystemManage.PermissionActionUpdateParams
) {
  const body: V5RequestBody<'/permission-actions/{id}', 'put'> = {
    action_key: data.permission_key || '',
    name: data.name || '',
    description: data.description,
    status: data.status,
    group_id: data.feature_group_id || data.module_group_id || null
  }
  const { error } = await v5Client.PUT('/permission-actions/{id}', {
    params: { path: { id } },
    body
  })
  if (error) throw error
}

/** 删除功能权限 */
export async function fetchDeletePermissionAction(id: string) {
  const { error } = await v5Client.DELETE('/permission-actions/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

export async function fetchGetPermissionActionImpactPreview(id: string) {
  const res = await unwrap(
    v5Client.GET('/permission-actions/{id}/impact-preview', { params: { path: { id } } })
  )
  return {
    permissionKey: res.permission_key || '',
    apiCount: Number(res.api_count ?? 0),
    pageCount: Number(res.page_count ?? 0),
    packageCount: Number(res.package_count ?? 0),
    roleCount: Number(res.role_count ?? 0),
    collaborationWorkspaceCount: Number(res.collaboration_workspace_count ?? 0),
    userCount: Number(res.user_count ?? 0)
  }
}

export async function fetchBatchUpdatePermissionActions(
  data: Api.SystemManage.PermissionBatchUpdateParams
) {
  const body: V5RequestBody<'/permission-actions/batch', 'post'> = {
    ids: data.ids,
    status: data.status,
    module_group_id: data.moduleGroupId,
    feature_group_id: data.featureGroupId,
    template_name: data.templateName
  }
  const res = toV5Record(
    await unwrap(
    v5Client.POST('/permission-actions/batch', {
      body
    })
  )
  )
  return {
    updatedCount: Number(res?.updatedCount ?? res?.updated_count ?? 0),
    skippedIds: Array.isArray(res?.skippedIds || res?.skipped_ids)
      ? res?.skippedIds || res?.skipped_ids
      : []
  }
}

export async function fetchSavePermissionBatchTemplate(
  data: Partial<Api.SystemManage.PermissionBatchTemplateItem>
) {
  const body: V5RequestBody<'/permission-actions/templates', 'post'> = {
    name: data.name,
    description: data.description,
    payload: data.payload
  }
  const res = toV5Record(
    await unwrap(
    v5Client.POST('/permission-actions/templates', {
      body
    })
  )
  )
  return {
    id: res?.id || '',
    name: res?.name || '',
    description: res?.description || '',
    payload: res?.payload || {},
    createdBy: res?.created_by || res?.createdBy || '',
    createdAt: res?.created_at || res?.createdAt || '',
    updatedAt: res?.updated_at || res?.updatedAt || ''
  }
}

export async function fetchGetPermissionBatchTemplates() {
  const res = await unwrap(v5Client.GET('/permission-actions/templates'))
  const records = Array.isArray(res?.records) ? (res.records as PermissionBatchTemplateRecord[]) : []
  return {
    records: records.map((item) => ({
      id: item?.id || '',
      name: item?.name || '',
      description: item?.description || '',
      payload: item?.payload || {},
      createdBy: item?.created_by || item?.createdBy || '',
      createdAt: item?.created_at || item?.createdAt || '',
      updatedAt: item?.updated_at || item?.updatedAt || ''
    })),
    total: Number(res?.total || 0)
  }
}

export async function fetchGetPermissionRiskAudits(params?: {
  current?: number
  size?: number
  objectId?: string
}) {
  const query: V5Query<'/permission-actions/risk-audits', 'get'> = {
    current: params?.current,
    size: params?.size
  }
  const res = await unwrap(v5Client.GET('/permission-actions/risk-audits', { params: { query } }))
  return {
    records: (res.records || []).map(normalizeRiskAudit),
    total: Number(res.total || 0)
  }
}

export async function fetchGetPermissionGroupList(
  params: Api.SystemManage.PermissionGroupSearchParams
) {
  void params
  const res = await unwrap(v5Client.GET('/permission-actions/groups'))
  return {
    ...res,
    records: (res.records || [])
      .map((item) => normalizePermissionGroup(item))
      .filter((item): item is Api.SystemManage.PermissionGroupItem => Boolean(item))
  } as Api.SystemManage.PermissionGroupList
}

export async function fetchCreatePermissionGroup(
  data: Api.SystemManage.PermissionGroupSaveParams
) {
  const body: V5RequestBody<'/permission-actions/groups', 'post'> = toV5Body(data)
  const res = toV5Record(await unwrap(v5Client.POST('/permission-actions/groups', { body })))
  return { id: `${res.id || ''}` }
}

export async function fetchUpdatePermissionGroup(
  id: string,
  data: Api.SystemManage.PermissionGroupSaveParams
) {
  const body: V5RequestBody<'/permission-actions/groups/{id}', 'put'> = toV5Body(data)
  const { error } = await v5Client.PUT('/permission-actions/groups/{id}', {
    params: { path: { id } },
    body
  })
  if (error) throw error
}

export async function fetchDeletePermissionGroup(id: string) {
  const { error } = await v5Client.DELETE('/permission-actions/groups/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

/** 获取功能包列表 */
export async function fetchGetFeaturePackageList(
  params: Api.SystemManage.FeaturePackageSearchParams
) {
  const query: V5Query<'/feature-packages', 'get'> = {
    current: params?.current,
    size: params?.size,
    keyword: params?.keyword,
    package_type: params?.packageType,
    status: params?.status
  }
  const res = await unwrap(v5Client.GET('/feature-packages', { params: { query } }))
  return {
    ...res,
    records: (res.records || []).map(normalizeFeaturePackage)
  } as Api.SystemManage.FeaturePackageList
}

export async function fetchGetFeaturePackageOptions(
  params?: Api.SystemManage.FeaturePackageSearchParams
) {
  const query: V5Query<'/feature-packages/options', 'get'> = {
    app_key: params?.appKey,
    package_type: params?.packageType
  }
  const res = await unwrap(v5Client.GET('/feature-packages/options', { params: { query } }))
  return {
    records: (res.records || []).map(normalizeFeaturePackage),
    total: res.total || 0
  }
}

export async function fetchGetCollaborationWorkspaceOptions(
  params?: Partial<Api.SystemManage.CollaborationWorkspaceSearchParams>
) {
  void params
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/options'))
  return {
    records: (res.records || []).map(normalizeCollaborationWorkspace),
    total: res.total || 0
  }
}

/** 获取功能包详情 */
export async function fetchGetFeaturePackage(id: string) {
  const res = await unwrap(v5Client.GET('/feature-packages/{id}', { params: { path: { id } } }))
  return normalizeFeaturePackage(res)
}

/** 获取组合包基础包 */
export async function fetchGetFeaturePackageChildren(id: string, appKey?: string) {
  const query: V5Query<'/feature-packages/{id}/children', 'get'> = appKey
    ? { app_key: appKey }
    : {}
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/children', {
      params: { path: { id }, query }
    })
  )
  return {
    child_package_ids: res.package_ids || [],
    packages: (res.packages || []).map(normalizeFeaturePackage)
  }
}

/** 设置组合包基础包 */
export async function fetchSetFeaturePackageChildren(
  id: string,
  childPackageIds: string[] | Api.SystemManage.FeaturePackageChildSetParams,
  appKey?: string
) {
  const query = appKey
    ? { app_key: appKey }
    : !Array.isArray(childPackageIds) && childPackageIds.app_key
      ? { app_key: childPackageIds.app_key }
      : undefined
  const body: V5RequestBody<'/feature-packages/{id}/children', 'put'> = Array.isArray(childPackageIds)
    ? { ids: childPackageIds }
    : {
        ids: childPackageIds.child_package_ids
      }
  const res = toV5Record(
    await unwrap(
    v5Client.PUT('/feature-packages/{id}/children', {
      params: query ? { path: { id }, query } : { path: { id } },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取功能包包含关系树 */
export async function fetchGetFeaturePackageRelationTree(params?: {
  workspaceScope?: string
  keyword?: string
}) {
  const query: V5Query<'/feature-packages/relationship-tree', 'get'> = {
    workspace_scope: params?.workspaceScope,
    keyword: params?.keyword
  }
  const res = await unwrap(v5Client.GET('/feature-packages/relationship-tree', { params: { query } }))
  return normalizeFeaturePackageRelationTree(res)
}

/** 创建功能包 */
export async function fetchCreateFeaturePackage(
  data: Api.SystemManage.FeaturePackageCreateParams
) {
  const body: V5RequestBody<'/feature-packages', 'post'> = data
  const res = toV5Record(await unwrap(v5Client.POST('/feature-packages', { body })))
  return { id: `${res.id || ''}` }
}

/** 更新功能包 */
export async function fetchUpdateFeaturePackage(
  id: string,
  data: Api.SystemManage.FeaturePackageUpdateParams
) {
  const body: V5RequestBody<'/feature-packages/{id}', 'put'> = {
    package_key: data.package_key || '',
    name: data.name || '',
    description: data.description,
    package_type: data.package_type,
    context_type: data.workspace_scope,
    status: data.status,
    sort_order: data.sort_order,
    app_keys: data.app_keys
  }
  const res = toV5Record(
    await unwrap(
    v5Client.PUT('/feature-packages/{id}', {
      params: { path: { id } },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 删除功能包 */
export async function fetchDeleteFeaturePackage(id: string) {
  const res = toV5Record(
    await unwrap(
    v5Client.DELETE('/feature-packages/{id}', { params: { path: { id } } })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取功能包包含的功能权限 */
export async function fetchGetFeaturePackageActions(id: string, appKey?: string) {
  const query: V5Query<'/feature-packages/{id}/actions', 'get'> = appKey
    ? { app_key: appKey }
    : {}
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/actions', {
      params: { path: { id }, query }
    })
  )
  return {
    action_ids: res.action_ids || [],
    actions: (res.actions || []).map(normalizePermissionAction)
  }
}

/** 设置功能包包含的功能权限 */
export async function fetchSetFeaturePackageActions(
  id: string,
  actionIds: string[] | Api.SystemManage.FeaturePackageActionSetParams,
  appKey?: string
) {
  const query =
    appKey
      ? { app_key: appKey }
      : !Array.isArray(actionIds) && actionIds.app_key
        ? { app_key: actionIds.app_key }
        : undefined
  const body: V5RequestBody<'/feature-packages/{id}/actions', 'put'> = Array.isArray(actionIds)
    ? { ids: actionIds }
    : {
        ids: actionIds.action_ids
      }
  const res = toV5Record(
    await unwrap(
    v5Client.PUT('/feature-packages/{id}/actions', {
      params: query ? { path: { id }, query } : { path: { id } },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取功能包包含的菜单 */
export async function fetchGetFeaturePackageMenus(id: string, appKey?: string) {
  const query: V5Query<'/feature-packages/{id}/menus', 'get'> = appKey ? { app_key: appKey } : {}
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/menus', {
      params: { path: { id }, query }
    })
  )
  return {
    menu_ids: res.menu_ids || [],
    menus: res.menus || []
  }
}

/** 设置功能包包含的菜单 */
export async function fetchSetFeaturePackageMenus(
  id: string,
  menuIds: string[],
  appKey?: string
) {
  const query = appKey ? { app_key: appKey } : undefined
  const body: V5RequestBody<'/feature-packages/{id}/menus', 'put'> = {
    ids: menuIds
  }
  const res = toV5Record(
    await unwrap(
    v5Client.PUT('/feature-packages/{id}/menus', {
      params: query ? { path: { id }, query } : { path: { id } },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取已开通当前功能包的协作空间 */
export async function fetchGetFeaturePackageCollaborationWorkspaces(id: string) {
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/collaboration-workspaces', {
      params: { path: { id } }
    })
  )
  return {
    collaboration_workspace_ids: (res.records || [])
      .map((item: { id?: string; collaboration_workspace_id?: string }) =>
        `${item.id || item.collaboration_workspace_id || ''}`.trim()
      )
      .filter(Boolean)
  } satisfies Api.SystemManage.FeaturePackageCollaborationWorkspaceBinding
}

/** 配置功能包开通协作空间 */
export async function fetchSetFeaturePackageCollaborationWorkspaces(
  id: string,
  collaborationWorkspaceIds:
    | string[]
    | Api.SystemManage.FeaturePackageCollaborationWorkspaceSetParams
) {
  const body: V5RequestBody<'/feature-packages/{id}/collaboration-workspaces', 'put'> = Array.isArray(collaborationWorkspaceIds)
    ? { ids: collaborationWorkspaceIds }
    : { ids: collaborationWorkspaceIds.collaboration_workspace_ids }
  const res = toV5Record(
    await unwrap(
    v5Client.PUT('/feature-packages/{id}/collaboration-workspaces', {
      params: { path: { id } },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取协作空间已开通的功能包 */
export async function fetchGetCollaborationWorkspaceFeaturePackages(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  const query: V5Query<'/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', 'get'> =
    appKey ? { app_key: appKey } : {}
  const res = await unwrap(
    v5Client.GET('/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', {
      params: {
        path: { collaborationWorkspaceId },
        query
      }
    })
  )
  return {
    package_ids: res.package_ids || [],
    packages: (res.packages || []).map(normalizeFeaturePackage)
  }
}

/** 设置协作空间功能包 */
export async function fetchSetCollaborationWorkspaceFeaturePackages(
  collaborationWorkspaceId: string,
  packageIds: string[] | Api.SystemManage.CollaborationWorkspaceFeaturePackageSetParams,
  appKey?: string
) {
  const query: V5Query<'/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', 'put'> =
    appKey ? { app_key: appKey } : {}
  const body: V5RequestBody<'/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', 'put'> =
    Array.isArray(packageIds) ? { ids: packageIds } : { ids: packageIds.package_ids }
  const res = toV5Record(
    await unwrap(
    v5Client.PUT('/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', {
      params: {
        path: { collaborationWorkspaceId },
        query
      },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

export async function fetchGetFeaturePackageImpactPreview(id: string) {
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/impact-preview', { params: { path: { id } } })
  )
  return {
    packageId: res?.package_id || '',
    roleCount: Number(res?.role_count ?? 0),
    collaborationWorkspaceCount: Number(res?.collaboration_workspace_count ?? 0),
    userCount: Number(res?.user_count ?? 0),
    menuCount: Number(res?.menu_count ?? 0),
    actionCount: Number(res?.action_count ?? 0)
  }
}

export async function fetchGetFeaturePackageVersions(id: string, current = 1, size = 20) {
  const query: V5Query<'/feature-packages/{id}/versions', 'get'> = { current, size }
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/versions', {
      params: { path: { id }, query }
    })
  )
  const records = Array.isArray(res?.records) ? (res.records as FeaturePackageVersionRecord[]) : []
  return {
    records: records.map((item) => ({
      id: item?.id || '',
      packageId: item?.package_id || item?.packageId || '',
      versionNo: Number(item?.version_no ?? item?.versionNo ?? 0),
      changeType: item?.change_type || item?.changeType || '',
      snapshot: item?.snapshot || {},
      operatorId: item?.operator_id || item?.operatorId || '',
      requestId: item?.request_id || item?.requestId || '',
      createdAt: item?.created_at || item?.createdAt || ''
    })),
    total: Number(res?.total || 0)
  }
}

export async function fetchRollbackFeaturePackage(id: string, versionId: string) {
  const body: V5RequestBody<'/feature-packages/{id}/rollback', 'post'> = { version_id: versionId }
  const res = toV5Record(
    await unwrap(
    v5Client.POST('/feature-packages/{id}/rollback', {
      params: { path: { id } },
      body
    })
  )
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

export async function fetchGetFeaturePackageRiskAudits(id: string, current = 1, size = 20) {
  const query: V5Query<'/feature-packages/{id}/risk-audits', 'get'> = { current, size }
  const res = await unwrap(
    v5Client.GET('/feature-packages/{id}/risk-audits', {
      params: { path: { id }, query }
    })
  )
  return {
    records: (res.records || []).map(normalizeRiskAudit),
    total: Number(res.total || 0)
  }
}
