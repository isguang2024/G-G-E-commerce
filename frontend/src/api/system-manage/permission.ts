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
  normalizeCollaborationWorkspace
} from './_shared'

/** 获取功能权限列表 */
export async function fetchGetPermissionActionList(
  params: Api.SystemManage.PermissionActionSearchParams
) {
  const query: Record<string, any> = {
    ...params,
    permission_key: params?.permissionKey,
    module_code: params?.moduleCode,
    module_group_id: params?.moduleGroupId,
    feature_group_id: params?.featureGroupId,
    context_type: params?.contextType,
    feature_kind: params?.featureKind,
    is_builtin: params?.isBuiltin,
    usage_pattern: params?.usagePattern,
    duplicate_pattern: params?.duplicatePattern
  }
  delete query.permissionKey
  delete query.moduleCode
  delete query.moduleGroupId
  delete query.featureGroupId
  delete query.contextType
  delete query.featureKind
  delete query.isBuiltin
  delete query.usagePattern
  delete query.duplicatePattern
  const res: any = await unwrap(
    v5Client.GET('/permission-actions', { params: { query } as any })
  )
  return {
    ...res,
    records: (res?.records || []).map(normalizePermissionAction),
    auditSummary: normalizePermissionAuditSummary(
      res?.audit_summary || res?.auditSummary || {}
    )
  } as Api.SystemManage.PermissionActionList
}

export async function fetchGetPermissionActionOptions(
  params?: Api.SystemManage.PermissionActionSearchParams
) {
  const query: Record<string, any> = {
    ...(params || {}),
    permission_key: params?.permissionKey,
    module_code: params?.moduleCode,
    module_group_id: params?.moduleGroupId,
    feature_group_id: params?.featureGroupId,
    context_type: params?.contextType,
    feature_kind: params?.featureKind,
    is_builtin: params?.isBuiltin
  }
  delete query.permissionKey
  delete query.moduleCode
  delete query.moduleGroupId
  delete query.featureGroupId
  delete query.contextType
  delete query.featureKind
  delete query.isBuiltin
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/options', { params: { query } as any })
  )
  return {
    records: (res?.records || []).map(normalizePermissionAction),
    total: res?.total || 0
  }
}

/** 获取功能权限详情 */
export async function fetchGetPermissionAction(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/{id}', { params: { path: { id } } })
  )
  return normalizePermissionAction(res)
}

/** 获取功能权限关联接口 */
export async function fetchGetPermissionActionEndpoints(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/{id}/endpoints', { params: { path: { id } } })
  )
  return {
    records: (res?.records || []).map(normalizeApiEndpoint),
    total: res?.total || 0
  }
}

/** 获取功能权限消费明细（API/页面/功能包/角色） */
export async function fetchGetPermissionActionConsumers(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/{id}/consumers', { params: { path: { id } } })
  )
  return normalizePermissionActionConsumers(res)
}

/** 新增功能权限关联接口 */
export async function fetchAddPermissionActionEndpoint(id: string, endpointCode: string) {
  const { error } = await v5Client.POST('/permission-actions/{id}/endpoints', {
    params: { path: { id } },
    body: { endpoint_code: endpointCode } as any
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
  const res: any = await unwrap(
    v5Client.POST('/permission-actions/cleanup-unused', {} as any)
  )
  return {
    deletedCount: Number(res?.deletedCount ?? res?.deleted_count ?? 0),
    deletedKeys: Array.isArray(res?.deletedKeys || res?.deleted_keys || [])
      ? (res?.deletedKeys || res?.deleted_keys || [])
          .map((value: any) => `${value || ''}`.trim())
          .filter(Boolean)
      : []
  }
}

/** 创建功能权限 */
export async function fetchCreatePermissionAction(
  data: Api.SystemManage.PermissionActionCreateParams
) {
  const res: any = await unwrap(
    v5Client.POST('/permission-actions', { body: data as any })
  )
  return res as { id: string }
}

/** 更新功能权限 */
export async function fetchUpdatePermissionAction(
  id: string,
  data: Api.SystemManage.PermissionActionUpdateParams
) {
  const { error } = await v5Client.PUT('/permission-actions/{id}', {
    params: { path: { id } },
    body: data as any
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
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/{id}/impact-preview', { params: { path: { id } } })
  )
  return {
    permissionKey: res?.permissionKey || res?.permission_key || '',
    apiCount: Number(res?.apiCount ?? res?.api_count ?? 0),
    pageCount: Number(res?.pageCount ?? res?.page_count ?? 0),
    packageCount: Number(res?.packageCount ?? res?.package_count ?? 0),
    roleCount: Number(res?.roleCount ?? res?.role_count ?? 0),
    collaborationWorkspaceCount: Number(res?.collaborationWorkspaceCount ?? 0),
    userCount: Number(res?.userCount ?? res?.user_count ?? 0)
  }
}

export async function fetchBatchUpdatePermissionActions(
  data: Api.SystemManage.PermissionBatchUpdateParams
) {
  const res: any = await unwrap(
    v5Client.POST('/permission-actions/batch', {
      body: {
        ids: data.ids,
        status: data.status,
        module_group_id: data.moduleGroupId,
        feature_group_id: data.featureGroupId,
        template_name: data.templateName
      } as any
    })
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
  const res: any = await unwrap(
    v5Client.POST('/permission-actions/templates', {
      body: {
        name: data.name,
        description: data.description,
        payload: data.payload
      } as any
    })
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
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/templates', { params: { query: {} as any } })
  )
  return {
    records: (res?.records || []).map((item: any) => ({
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
  const query: Record<string, any> = {
    current: params?.current,
    size: params?.size,
    object_id: params?.objectId
  }
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/risk-audits', { params: { query } as any })
  )
  return {
    records: (res?.records || []).map(normalizeRiskAudit),
    total: Number(res?.total || 0)
  }
}

export async function fetchGetPermissionGroupList(
  params: Api.SystemManage.PermissionGroupSearchParams
) {
  const query: Record<string, any> = {
    ...params,
    group_type: params?.groupType
  }
  delete query.groupType
  const res: any = await unwrap(
    v5Client.GET('/permission-actions/groups', { params: { query } as any })
  )
  return {
    ...res,
    records: (res?.records || [])
      .map((item: any) => normalizePermissionGroup(item))
      .filter((item: any): item is Api.SystemManage.PermissionGroupItem => Boolean(item))
  } as Api.SystemManage.PermissionGroupList
}

export async function fetchCreatePermissionGroup(
  data: Api.SystemManage.PermissionGroupSaveParams
) {
  const res: any = await unwrap(
    v5Client.POST('/permission-actions/groups', { body: data as any })
  )
  return res as { id: string }
}

export async function fetchUpdatePermissionGroup(
  id: string,
  data: Api.SystemManage.PermissionGroupSaveParams
) {
  const { error } = await v5Client.PUT('/permission-actions/groups/{id}', {
    params: { path: { id } },
    body: data as any
  })
  if (error) throw error
}

export async function fetchDeletePermissionGroup(id: string) {
  // TODO(v5): DELETE /permission-actions/groups/{id} 未在 openapi.yaml 暴露，临时走 legacy
  const { error } = await (v5Client as any).DELETE('/permission-actions/groups/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

/** 获取功能包列表 */
export async function fetchGetFeaturePackageList(
  params: Api.SystemManage.FeaturePackageSearchParams
) {
  const query: Record<string, any> = {
    ...params,
    app_key: params?.appKey,
    package_key: params?.packageKey,
    package_type: params?.packageType,
    workspace_scope: params?.workspaceScope
  }
  delete query.appKey
  delete query.packageKey
  delete query.packageType
  delete query.workspaceScope
  const res: any = await unwrap(
    v5Client.GET('/feature-packages', { params: { query } as any })
  )
  return {
    ...res,
    records: (res?.records || []).map(normalizeFeaturePackage)
  } as Api.SystemManage.FeaturePackageList
}

export async function fetchGetFeaturePackageOptions(
  params?: Api.SystemManage.FeaturePackageSearchParams
) {
  const query: Record<string, any> = {
    ...(params || {}),
    app_key: params?.appKey,
    package_key: params?.packageKey,
    package_type: params?.packageType,
    workspace_scope: params?.workspaceScope
  }
  delete query.appKey
  delete query.packageKey
  delete query.packageType
  delete query.workspaceScope
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/options', { params: { query } as any })
  )
  return {
    records: (res?.records || []).map(normalizeFeaturePackage),
    total: res?.total || 0
  }
}

export async function fetchGetCollaborationWorkspaceOptions(
  params?: Partial<Api.SystemManage.CollaborationWorkspaceSearchParams>
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/options', {
      params: { query: (params || {}) as any }
    })
  )
  return {
    records: (res?.records || []).map(normalizeCollaborationWorkspace),
    total: res?.total || 0
  }
}

/** 获取功能包详情 */
export async function fetchGetFeaturePackage(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}', { params: { path: { id } } })
  )
  return normalizeFeaturePackage(res)
}

/** 获取组合包基础包 */
export async function fetchGetFeaturePackageChildren(id: string, appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/children', {
      params: { path: { id }, query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    child_package_ids: res?.child_package_ids || [],
    packages: (res?.packages || []).map(normalizeFeaturePackage)
  }
}

/** 设置组合包基础包 */
export async function fetchSetFeaturePackageChildren(
  id: string,
  childPackageIds: string[] | Api.SystemManage.FeaturePackageChildSetParams,
  appKey?: string
) {
  const payload = Array.isArray(childPackageIds)
    ? { child_package_ids: childPackageIds, ...(appKey ? { app_key: appKey } : {}) }
    : {
        ...childPackageIds,
        ...(appKey && !childPackageIds.app_key ? { app_key: appKey } : {})
      }
  const res: any = await unwrap(
    v5Client.PUT('/feature-packages/{id}/children', {
      params: { path: { id } },
      body: payload as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取功能包包含关系树 */
export async function fetchGetFeaturePackageRelationTree(params?: {
  workspaceScope?: string
  keyword?: string
}) {
  const query: Record<string, any> = {
    workspace_scope: params?.workspaceScope,
    keyword: params?.keyword
  }
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/relationship-tree', { params: { query } as any })
  )
  return normalizeFeaturePackageRelationTree(res)
}

/** 创建功能包 */
export async function fetchCreateFeaturePackage(
  data: Api.SystemManage.FeaturePackageCreateParams
) {
  const res: any = await unwrap(
    v5Client.POST('/feature-packages', { body: data as any })
  )
  return res as { id: string }
}

/** 更新功能包 */
export async function fetchUpdateFeaturePackage(
  id: string,
  data: Api.SystemManage.FeaturePackageUpdateParams
) {
  const res: any = await unwrap(
    v5Client.PUT('/feature-packages/{id}', {
      params: { path: { id } },
      body: data as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 删除功能包 */
export async function fetchDeleteFeaturePackage(id: string) {
  const res: any = await unwrap(
    v5Client.DELETE('/feature-packages/{id}', { params: { path: { id } } })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取功能包包含的功能权限 */
export async function fetchGetFeaturePackageActions(id: string, appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/actions', {
      params: { path: { id }, query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizePermissionAction)
  }
}

/** 设置功能包包含的功能权限 */
export async function fetchSetFeaturePackageActions(
  id: string,
  actionIds: string[] | Api.SystemManage.FeaturePackageActionSetParams,
  appKey?: string
) {
  const payload = Array.isArray(actionIds)
    ? { action_ids: actionIds, ...(appKey ? { app_key: appKey } : {}) }
    : {
        ...actionIds,
        ...(appKey && !actionIds.app_key ? { app_key: appKey } : {})
      }
  const res: any = await unwrap(
    v5Client.PUT('/feature-packages/{id}/actions', {
      params: { path: { id } },
      body: payload as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取功能包包含的菜单 */
export async function fetchGetFeaturePackageMenus(id: string, appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/menus', {
      params: { path: { id }, query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    menu_ids: res?.menu_ids || [],
    menus: res?.menus || []
  }
}

/** 设置功能包包含的菜单 */
export async function fetchSetFeaturePackageMenus(
  id: string,
  menuIds: string[],
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.PUT('/feature-packages/{id}/menus', {
      params: { path: { id } },
      body: { menu_ids: menuIds, ...(appKey ? { app_key: appKey } : {}) } as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取已开通当前功能包的协作空间 */
export async function fetchGetFeaturePackageCollaborationWorkspaces(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/collaboration-workspaces', {
      params: { path: { id } }
    })
  )
  return res as Api.SystemManage.FeaturePackageCollaborationWorkspaceBinding
}

/** 配置功能包开通协作空间 */
export async function fetchSetFeaturePackageCollaborationWorkspaces(
  id: string,
  collaborationWorkspaceIds:
    | string[]
    | Api.SystemManage.FeaturePackageCollaborationWorkspaceSetParams
) {
  const payload = Array.isArray(collaborationWorkspaceIds)
    ? { collaboration_workspace_ids: collaborationWorkspaceIds }
    : collaborationWorkspaceIds
  const res: any = await unwrap(
    v5Client.PUT('/feature-packages/{id}/collaboration-workspaces', {
      params: { path: { id } },
      body: payload as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

/** 获取协作空间已开通的功能包 */
export async function fetchGetCollaborationWorkspaceFeaturePackages(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', {
      params: {
        path: { collaborationWorkspaceId },
        query: (appKey ? { app_key: appKey } : {}) as any
      }
    })
  )
  return {
    package_ids: res?.package_ids || [],
    packages: (res?.packages || []).map(normalizeFeaturePackage)
  }
}

/** 设置协作空间功能包 */
export async function fetchSetCollaborationWorkspaceFeaturePackages(
  collaborationWorkspaceId: string,
  packageIds: string[] | Api.SystemManage.CollaborationWorkspaceFeaturePackageSetParams,
  appKey?: string
) {
  const payload = Array.isArray(packageIds) ? { package_ids: packageIds } : packageIds
  const res: any = await unwrap(
    v5Client.PUT('/feature-packages/collaboration-workspaces/{collaborationWorkspaceId}', {
      params: {
        path: { collaborationWorkspaceId },
        query: (appKey ? { app_key: appKey } : {}) as any
      },
      body: payload as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

export async function fetchGetFeaturePackageImpactPreview(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/impact-preview', { params: { path: { id } } })
  )
  return {
    packageId: res?.package_id || res?.packageId || '',
    roleCount: Number(res?.role_count ?? res?.roleCount ?? 0),
    collaborationWorkspaceCount: Number(res?.collaborationWorkspaceCount ?? 0),
    userCount: Number(res?.user_count ?? res?.userCount ?? 0),
    menuCount: Number(res?.menu_count ?? res?.menuCount ?? 0),
    actionCount: Number(res?.action_count ?? res?.actionCount ?? 0)
  }
}

export async function fetchGetFeaturePackageVersions(id: string, current = 1, size = 20) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/versions', {
      params: { path: { id }, query: { current, size } as any }
    })
  )
  return {
    records: (res?.records || []).map((item: any) => ({
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
  const res: any = await unwrap(
    v5Client.POST('/feature-packages/{id}/rollback', {
      params: { path: { id } },
      body: { version_id: versionId } as any
    })
  )
  return normalizeRefreshStats(res?.refresh_stats || res?.refreshStats || {})
}

export async function fetchGetFeaturePackageRiskAudits(id: string, current = 1, size = 20) {
  const res: any = await unwrap(
    v5Client.GET('/feature-packages/{id}/risk-audits', {
      params: { path: { id }, query: { current, size } as any }
    })
  )
  return {
    records: (res?.records || []).map(normalizeRiskAudit),
    total: Number(res?.total || 0)
  }
}
