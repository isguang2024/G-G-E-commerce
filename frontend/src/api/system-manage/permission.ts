import {
  request,
  ACTION_PERMISSION_BASE,
  FEATURE_PACKAGE_BASE,
  COLLABORATION_WORKSPACE_BASE,
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
export function fetchGetPermissionActionList(
  params: Api.SystemManage.PermissionActionSearchParams
) {
  const normalizedParams = {
    ...params,
    permission_key: params?.permissionKey,
    module_code: params?.moduleCode,
    module_group_id: params?.moduleGroupId,
    feature_group_id: params?.featureGroupId,
    context_type: params?.contextType,
    feature_kind: params?.featureKind,
    is_builtin: params?.isBuiltin,
    usage_pattern: params?.usagePattern,
    duplicate_pattern: params?.duplicatePattern,
    permissionKey: undefined,
    moduleCode: undefined,
    moduleGroupId: undefined,
    featureGroupId: undefined,
    contextType: undefined,
    featureKind: undefined,
    isBuiltin: undefined,
    usagePattern: undefined,
    duplicatePattern: undefined
  }
  return request
    .get<Api.SystemManage.PermissionActionList>({
      url: ACTION_PERMISSION_BASE,
      params: normalizedParams
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizePermissionAction),
      auditSummary: normalizePermissionAuditSummary(
        (res as any)?.audit_summary || res?.auditSummary || {}
      )
    }))
}

export function fetchGetPermissionActionOptions(
  params?: Api.SystemManage.PermissionActionSearchParams
) {
  const normalizedParams = {
    ...params,
    permission_key: params?.permissionKey,
    module_code: params?.moduleCode,
    module_group_id: params?.moduleGroupId,
    feature_group_id: params?.featureGroupId,
    context_type: params?.contextType,
    feature_kind: params?.featureKind,
    is_builtin: params?.isBuiltin,
    permissionKey: undefined,
    moduleCode: undefined,
    moduleGroupId: undefined,
    featureGroupId: undefined,
    contextType: undefined,
    featureKind: undefined,
    isBuiltin: undefined
  }
  return request
    .get<{ records: Api.SystemManage.PermissionActionItem[]; total: number }>({
      url: `${ACTION_PERMISSION_BASE}/options`,
      params: normalizedParams
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizePermissionAction),
      total: res?.total || 0
    }))
}

/** 获取功能权限详情 */
export function fetchGetPermissionAction(id: string) {
  return request
    .get<Api.SystemManage.PermissionActionItem>({
      url: `${ACTION_PERMISSION_BASE}/${id}`
    })
    .then((res) => normalizePermissionAction(res))
}

/** 获取功能权限关联接口 */
export function fetchGetPermissionActionEndpoints(id: string) {
  return request
    .get<Api.SystemManage.PermissionActionEndpointResponse>({
      url: `${ACTION_PERMISSION_BASE}/${id}/endpoints`
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeApiEndpoint),
      total: res?.total || 0
    }))
}

/** 获取功能权限消费明细（API/页面/功能包/角色） */
export function fetchGetPermissionActionConsumers(id: string) {
  return request
    .get<Api.SystemManage.PermissionActionConsumerDetails>({
      url: `${ACTION_PERMISSION_BASE}/${id}/consumers`
    })
    .then((res) => normalizePermissionActionConsumers(res))
}

/** 新增功能权限关联接口 */
export function fetchAddPermissionActionEndpoint(id: string, endpointCode: string) {
  return request.post<void>({
    url: `${ACTION_PERMISSION_BASE}/${id}/endpoints`,
    data: { endpoint_code: endpointCode }
  })
}

/** 删除功能权限关联接口 */
export function fetchDeletePermissionActionEndpoint(id: string, endpointCode: string) {
  return request.del<void>({
    url: `${ACTION_PERMISSION_BASE}/${id}/endpoints/${endpointCode}`
  })
}

export function fetchCleanupUnusedPermissionActions() {
  return request
    .post<
      Api.SystemManage.PermissionActionCleanupResult & {
        deleted_count?: number
        deleted_keys?: string[]
      }
    >({
      url: `${ACTION_PERMISSION_BASE}/cleanup-unused`
    })
    .then((res) => ({
      deletedCount: Number(res?.deletedCount ?? res?.deleted_count ?? 0),
      deletedKeys: Array.isArray(res?.deletedKeys || res?.deleted_keys || [])
        ? (res?.deletedKeys || res?.deleted_keys || [])
            .map((value: any) => `${value || ''}`.trim())
            .filter(Boolean)
        : []
    }))
}

/** 创建功能权限 */
export function fetchCreatePermissionAction(data: Api.SystemManage.PermissionActionCreateParams) {
  return request.post<{ id: string }>({
    url: ACTION_PERMISSION_BASE,
    data
  })
}

/** 更新功能权限 */
export function fetchUpdatePermissionAction(
  id: string,
  data: Api.SystemManage.PermissionActionUpdateParams
) {
  return request.put<void>({
    url: `${ACTION_PERMISSION_BASE}/${id}`,
    data
  })
}

/** 删除功能权限 */
export function fetchDeletePermissionAction(id: string) {
  return request.del<void>({
    url: `${ACTION_PERMISSION_BASE}/${id}`
  })
}

export function fetchGetPermissionActionImpactPreview(id: string) {
  return request
    .get<Api.SystemManage.PermissionImpactPreview>({
      url: `${ACTION_PERMISSION_BASE}/${id}/impact-preview`
    })
    .then((res: any) => ({
      permissionKey: res?.permissionKey || res?.permission_key || '',
      apiCount: Number(res?.apiCount ?? res?.api_count ?? 0),
      pageCount: Number(res?.pageCount ?? res?.page_count ?? 0),
      packageCount: Number(res?.packageCount ?? res?.package_count ?? 0),
      roleCount: Number(res?.roleCount ?? res?.role_count ?? 0),
      collaborationWorkspaceCount: Number(res?.collaborationWorkspaceCount ?? 0),
      userCount: Number(res?.userCount ?? res?.user_count ?? 0)
    }))
}

export function fetchBatchUpdatePermissionActions(
  data: Api.SystemManage.PermissionBatchUpdateParams
) {
  return request
    .post<Api.SystemManage.PermissionBatchUpdateResult>({
      url: `${ACTION_PERMISSION_BASE}/batch`,
      data: {
        ids: data.ids,
        status: data.status,
        module_group_id: data.moduleGroupId,
        feature_group_id: data.featureGroupId,
        template_name: data.templateName
      }
    })
    .then((res: any) => ({
      updatedCount: Number(res?.updatedCount ?? res?.updated_count ?? 0),
      skippedIds: Array.isArray(res?.skippedIds || res?.skipped_ids)
        ? res?.skippedIds || res?.skipped_ids
        : []
    }))
}

export function fetchSavePermissionBatchTemplate(
  data: Partial<Api.SystemManage.PermissionBatchTemplateItem>
) {
  return request
    .post<Api.SystemManage.PermissionBatchTemplateItem>({
      url: `${ACTION_PERMISSION_BASE}/templates`,
      data: {
        name: data.name,
        description: data.description,
        payload: data.payload
      }
    })
    .then((res: any) => ({
      id: res?.id || '',
      name: res?.name || '',
      description: res?.description || '',
      payload: res?.payload || {},
      createdBy: res?.created_by || res?.createdBy || '',
      createdAt: res?.created_at || res?.createdAt || '',
      updatedAt: res?.updated_at || res?.updatedAt || ''
    }))
}

export function fetchGetPermissionBatchTemplates() {
  return request
    .get<{ records: Api.SystemManage.PermissionBatchTemplateItem[]; total: number }>({
      url: `${ACTION_PERMISSION_BASE}/templates`
    })
    .then((res: any) => ({
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
    }))
}

export function fetchGetPermissionRiskAudits(params?: {
  current?: number
  size?: number
  objectId?: string
}) {
  return request
    .get<{ records: Api.SystemManage.RiskAuditItem[]; total: number }>({
      url: `${ACTION_PERMISSION_BASE}/risk-audits`,
      params: {
        current: params?.current,
        size: params?.size,
        object_id: params?.objectId
      }
    })
    .then((res: any) => ({
      records: (res?.records || []).map(normalizeRiskAudit),
      total: Number(res?.total || 0)
    }))
}

export function fetchGetPermissionGroupList(params: Api.SystemManage.PermissionGroupSearchParams) {
  const normalizedParams = {
    ...params,
    group_type: params?.groupType,
    groupType: undefined
  }
  return request
    .get<Api.SystemManage.PermissionGroupList>({
      url: `${ACTION_PERMISSION_BASE}/groups`,
      params: normalizedParams
    })
    .then((res) => ({
      ...res,
      records: (res?.records || [])
        .map((item: any) => normalizePermissionGroup(item))
        .filter((item): item is Api.SystemManage.PermissionGroupItem => Boolean(item))
    }))
}

export function fetchCreatePermissionGroup(data: Api.SystemManage.PermissionGroupSaveParams) {
  return request.post<{ id: string }>({
    url: `${ACTION_PERMISSION_BASE}/groups`,
    data
  })
}

export function fetchUpdatePermissionGroup(
  id: string,
  data: Api.SystemManage.PermissionGroupSaveParams
) {
  return request.put<void>({
    url: `${ACTION_PERMISSION_BASE}/groups/${id}`,
    data
  })
}

export function fetchDeletePermissionGroup(id: string) {
  return request.del<void>({
    url: `${ACTION_PERMISSION_BASE}/groups/${id}`
  })
}

/** 获取功能包列表 */
export function fetchGetFeaturePackageList(params: Api.SystemManage.FeaturePackageSearchParams) {
  const normalizedParams = {
    ...params,
    app_key: params?.appKey,
    package_key: params?.packageKey,
    package_type: params?.packageType,
    workspace_scope: params?.workspaceScope,
    appKey: undefined,
    packageKey: undefined,
    packageType: undefined,
    workspaceScope: undefined
  }
  return request
    .get<Api.SystemManage.FeaturePackageList>({
      url: FEATURE_PACKAGE_BASE,
      params: normalizedParams
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeFeaturePackage)
    }))
}

export function fetchGetFeaturePackageOptions(
  params?: Api.SystemManage.FeaturePackageSearchParams
) {
  const normalizedParams = {
    ...params,
    app_key: params?.appKey,
    package_key: params?.packageKey,
    package_type: params?.packageType,
    workspace_scope: params?.workspaceScope,
    appKey: undefined,
    packageKey: undefined,
    packageType: undefined,
    workspaceScope: undefined
  }
  return request
    .get<{ records: Api.SystemManage.FeaturePackageItem[]; total: number }>({
      url: `${FEATURE_PACKAGE_BASE}/options`,
      params: normalizedParams
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeFeaturePackage),
      total: res?.total || 0
    }))
}

export function fetchGetCollaborationWorkspaceOptions(
  params?: Partial<Api.SystemManage.CollaborationWorkspaceSearchParams>
) {
  return request
    .get<{ records: Api.SystemManage.CollaborationWorkspaceListItem[]; total: number }>({
      url: `${COLLABORATION_WORKSPACE_BASE}/options`,
      params
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeCollaborationWorkspace),
      total: res?.total || 0
    }))
}

/** 获取功能包详情 */
export function fetchGetFeaturePackage(id: string) {
  return request
    .get<Api.SystemManage.FeaturePackageItem>({
      url: `${FEATURE_PACKAGE_BASE}/${id}`
    })
    .then((res) => normalizeFeaturePackage(res))
}

/** 获取组合包基础包 */
export function fetchGetFeaturePackageChildren(id: string, appKey?: string) {
  return request
    .get<Api.SystemManage.FeaturePackageBundleResponse>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/children`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      child_package_ids: res?.child_package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置组合包基础包 */
export function fetchSetFeaturePackageChildren(
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
  return request
    .put<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/children`,
      data: payload
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

/** 获取功能包包含关系树 */
export function fetchGetFeaturePackageRelationTree(params?: {
  workspaceScope?: string
  keyword?: string
}) {
  return request
    .get<Api.SystemManage.FeaturePackageRelationTree>({
      url: `${FEATURE_PACKAGE_BASE}/relationship-tree`,
      params: {
        workspace_scope: params?.workspaceScope,
        keyword: params?.keyword
      }
    })
    .then((res) => normalizeFeaturePackageRelationTree(res))
}

/** 创建功能包 */
export function fetchCreateFeaturePackage(data: Api.SystemManage.FeaturePackageCreateParams) {
  return request.post<{ id: string }>({
    url: FEATURE_PACKAGE_BASE,
    data
  })
}

/** 更新功能包 */
export function fetchUpdateFeaturePackage(
  id: string,
  data: Api.SystemManage.FeaturePackageUpdateParams
) {
  return request
    .put<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}`,
      data
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

/** 删除功能包 */
export function fetchDeleteFeaturePackage(id: string) {
  return request
    .del<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}`
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

/** 获取功能包包含的功能权限 */
export function fetchGetFeaturePackageActions(id: string, appKey?: string) {
  return request
    .get<Api.SystemManage.FeaturePackageActionResponse>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/actions`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      action_ids: res?.action_ids || [],
      actions: (res?.actions || []).map(normalizePermissionAction)
    }))
}

/** 设置功能包包含的功能权限 */
export function fetchSetFeaturePackageActions(
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
  return request
    .put<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/actions`,
      data: payload
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

/** 获取功能包包含的菜单 */
export function fetchGetFeaturePackageMenus(id: string, appKey?: string) {
  return request
    .get<Api.SystemManage.FeaturePackageMenuResponse>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/menus`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      menu_ids: res?.menu_ids || [],
      menus: res?.menus || []
    }))
}

/** 设置功能包包含的菜单 */
export function fetchSetFeaturePackageMenus(id: string, menuIds: string[], appKey?: string) {
  return request
    .put<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/menus`,
      data: {
        menu_ids: menuIds,
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

/** 获取已开通当前功能包的协作空间 */
export function fetchGetFeaturePackageCollaborationWorkspaces(id: string) {
  return request.get<Api.SystemManage.FeaturePackageCollaborationWorkspaceBinding>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/collaboration-workspaces`
  })
}

/** 配置功能包开通协作空间 */
export function fetchSetFeaturePackageCollaborationWorkspaces(
  id: string,
  collaborationWorkspaceIds:
    | string[]
    | Api.SystemManage.FeaturePackageCollaborationWorkspaceSetParams
) {
  const payload = Array.isArray(collaborationWorkspaceIds)
    ? { collaboration_workspace_ids: collaborationWorkspaceIds }
    : collaborationWorkspaceIds
  return request
    .put<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/collaboration-workspaces`,
      data: payload
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

/** 获取协作空间已开通的功能包 */
export function fetchGetCollaborationWorkspaceFeaturePackages(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  return request
    .get<Api.SystemManage.CollaborationWorkspaceFeaturePackageResponse>({
      url: `${FEATURE_PACKAGE_BASE}/collaboration-workspaces/${collaborationWorkspaceId}`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置协作空间功能包 */
export function fetchSetCollaborationWorkspaceFeaturePackages(
  collaborationWorkspaceId: string,
  packageIds: string[] | Api.SystemManage.CollaborationWorkspaceFeaturePackageSetParams,
  appKey?: string
) {
  const payload = Array.isArray(packageIds) ? { package_ids: packageIds } : packageIds
  return request
    .put<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/collaboration-workspaces/${collaborationWorkspaceId}`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      },
      data: payload
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

export function fetchGetFeaturePackageImpactPreview(id: string) {
  return request
    .get<Api.SystemManage.FeaturePackageImpactPreview>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/impact-preview`
    })
    .then((res: any) => ({
      packageId: res?.package_id || res?.packageId || '',
      roleCount: Number(res?.role_count ?? res?.roleCount ?? 0),
      collaborationWorkspaceCount: Number(res?.collaborationWorkspaceCount ?? 0),
      userCount: Number(res?.user_count ?? res?.userCount ?? 0),
      menuCount: Number(res?.menu_count ?? res?.menuCount ?? 0),
      actionCount: Number(res?.action_count ?? res?.actionCount ?? 0)
    }))
}

export function fetchGetFeaturePackageVersions(id: string, current = 1, size = 20) {
  return request
    .get<{ records: Api.SystemManage.FeaturePackageVersionItem[]; total: number }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/versions`,
      params: { current, size }
    })
    .then((res: any) => ({
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
    }))
}

export function fetchRollbackFeaturePackage(id: string, versionId: string) {
  return request
    .post<{ refresh_stats?: Api.SystemManage.RefreshStats }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/rollback`,
      data: { version_id: versionId }
    })
    .then((res) => normalizeRefreshStats(res?.refresh_stats || (res as any)?.refreshStats || {}))
}

export function fetchGetFeaturePackageRiskAudits(id: string, current = 1, size = 20) {
  return request
    .get<{ records: Api.SystemManage.RiskAuditItem[]; total: number }>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/risk-audits`,
      params: { current, size }
    })
    .then((res: any) => ({
      records: (res?.records || []).map(normalizeRiskAudit),
      total: Number(res?.total || 0)
    }))
}
