import request from '@/utils/http'
import { v5Client } from '@/api/v5/client'
import { AppRouteRecord } from '@/types/router'
import type { FastEnterApplication, FastEnterQuickLink } from '@/types/config'
import { normalizeMenuSpaceKey } from '@/utils/navigation/menu-space'

export { request, v5Client, normalizeMenuSpaceKey }
export type { AppRouteRecord, FastEnterApplication, FastEnterQuickLink }

// Phase 4 slice 5: 用户域 list/get/create/update/delete/assignRoles 走 v5Client。
// ogen handler 返回 snake_case schema 原型，这里负责 camelCase 归一，
// 保持 Api.SystemManage.UserListItem 等消费方契约不变。
export function normalizeUserSummary(item: any): Api.SystemManage.UserListItem {
  const roleDetails = Array.isArray(item?.role_details)
    ? item.role_details.map((r: any) => ({
        code: r?.code || '',
        name: r?.name || ''
      }))
    : Array.isArray(item?.roleDetails)
      ? item.roleDetails
      : []
  return {
    id: item?.id || '',
    avatar: item?.avatar || item?.avatar_url || '',
    status: item?.status || 'inactive',
    userName: item?.user_name || item?.userName || '',
    nickName: item?.nick_name || item?.nickName || '',
    userPhone: item?.user_phone || item?.userPhone || '',
    userEmail: item?.user_email || item?.userEmail || '',
    systemRemark: item?.system_remark || item?.systemRemark || '',
    lastLoginTime: item?.last_login_time || item?.lastLoginTime || '',
    lastLoginIP: item?.last_login_ip || item?.lastLoginIP || '',
    userRoles: Array.isArray(item?.user_roles)
      ? item.user_roles
      : Array.isArray(item?.userRoles)
        ? item.userRoles
        : [],
    roleDetails,
    registerSource: item?.register_source || item?.registerSource || '',
    invitedBy: item?.invited_by || item?.invitedBy || '',
    invitedByName: item?.invited_by_name || item?.invitedByName || '',
    createTime: item?.create_time || item?.createTime || '',
    updateTime: item?.update_time || item?.updateTime || ''
  }
}

export const USER_BASE = '/api/v1/users'
export const ROLE_BASE = '/api/v1/roles'
export const ACTION_PERMISSION_BASE = '/api/v1/permission-actions'
export const FEATURE_PACKAGE_BASE = '/api/v1/feature-packages'
export const COLLABORATION_WORKSPACE_BASE = '/api/v1/collaboration-workspaces'
export const SYSTEM_BASE = '/api/v1/system'
export const APP_BASE = '/api/v1/system/apps'
export const APP_HOST_BINDING_BASE = '/api/v1/system/app-host-bindings'
export const API_ENDPOINT_BASE = '/api/v1/api-endpoints'
export const PAGE_BASE = '/api/v1/pages'
export const RUNTIME_BASE = '/api/v1/runtime'
export const MENU_BASE = '/api/v1/menus'
export const MENU_BACKUP_BASE = '/api/v1/menus/backups'

export function normalizePermissionKey(value?: string) {
  const target = `${value || ''}`.trim()
  if (target) {
    if (target.includes(':')) {
      const [resource, action] = target.split(':', 2)
      return [resource, action].filter(Boolean).join('.')
    }
    return target
  }
  return ''
}

export function derivePermissionSegments(permissionKey?: string) {
  const normalized = normalizePermissionKey(permissionKey)
  const parts = normalized.split('.').filter(Boolean)
  if (parts.length <= 1) {
    return { resourceCode: '', actionCode: '' }
  }
  return {
    resourceCode: parts.slice(0, -1).join('_'),
    actionCode: parts[parts.length - 1]
  }
}

export function deriveContextType(permissionKey?: string, moduleCode?: string) {
  const key = `${permissionKey || ''}`.trim()
  const module = `${moduleCode || ''}`.trim()
  if (
    key === 'collaboration_workspace.manage' ||
    key.startsWith('collaboration_workspace.member.') ||
    key.startsWith('collaboration_workspace.boundary.') ||
    key.startsWith('collaboration_workspace.message.') ||
    module === 'collaboration_workspace' ||
    module === 'collaboration_workspace_member' ||
    module === 'collaboration_workspace_boundary' ||
    module === 'collaboration_workspace_message'
  ) {
    return 'collaboration'
  }
  if (
    key.startsWith('personal.') ||
    module === 'personal' ||
    module === 'personal_workspace'
  ) {
    return 'personal'
  }
  if (
    key.startsWith('system.') ||
    key.startsWith('feature_package.') ||
    key.startsWith('message.') ||
    module === 'role' ||
    module === 'user' ||
    module === 'menu' ||
    module === 'menu_backup' ||
    module === 'permission_action' ||
    module === 'permission_key' ||
    module === 'api_endpoint' ||
    module === 'feature_package' ||
    module === 'page' ||
    module === 'collaboration_workspace_member_admin'
  ) {
    return 'common'
  }
  return 'common'
}

export function normalizePermissionGroup(value: any): Api.SystemManage.PermissionGroupItem | undefined {
  if (!value) return undefined
  return {
    id: value?.id || '',
    groupType: value?.group_type || value?.groupType || '',
    code: value?.code || '',
    name: value?.name || '',
    nameEn: value?.name_en || value?.nameEn || '',
    description: value?.description || '',
    status: value?.status || 'normal',
    sortOrder: value?.sort_order ?? value?.sortOrder ?? 0,
    isBuiltin: Boolean(value?.is_builtin ?? value?.isBuiltin ?? false)
  }
}

export function normalizePermissionAction(item: any): Api.SystemManage.PermissionActionItem {
  const permissionKey = normalizePermissionKey(item?.permission_key || item?.permissionKey)
  const legacy = derivePermissionSegments(permissionKey)
  const moduleCode = item?.module_code || item?.moduleCode || legacy.resourceCode || ''
  const consumerTypes = item?.consumer_types || item?.consumerTypes || []
  const duplicateKeys = item?.duplicate_keys || item?.duplicateKeys || []
  const moduleGroup = normalizePermissionGroup(item?.module_group || item?.moduleGroup)
  const featureGroup = normalizePermissionGroup(item?.feature_group || item?.featureGroup)
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    resourceCode: legacy.resourceCode,
    actionCode: legacy.actionCode,
    moduleCode: moduleGroup?.code || moduleCode,
    moduleGroupId: item?.module_group_id || item?.moduleGroupId || moduleGroup?.id || '',
    featureGroupId: item?.feature_group_id || item?.featureGroupId || featureGroup?.id || '',
    moduleGroup,
    featureGroup,
    contextType:
      item?.context_type || item?.contextType || deriveContextType(permissionKey, moduleCode),
    permissionKey,
    featureKind: featureGroup?.code || item?.feature_kind || item?.featureKind || 'system',
    dataPolicy: item?.data_policy || item?.dataPolicy || '',
    allowedWorkspaceTypes: item?.allowed_workspace_types || item?.allowedWorkspaceTypes || '',
    name: item?.name || '',
    description: item?.description || '',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    apiCount: Number(item?.api_count ?? item?.apiCount ?? 0),
    pageCount: Number(item?.page_count ?? item?.pageCount ?? 0),
    packageCount: Number(item?.package_count ?? item?.packageCount ?? 0),
    consumerTypes: Array.isArray(consumerTypes)
      ? consumerTypes.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    usagePattern: item?.usage_pattern || item?.usagePattern || 'unused',
    usageNote: item?.usage_note || item?.usageNote || '',
    duplicatePattern: item?.duplicate_pattern || item?.duplicatePattern || 'none',
    duplicateGroup: item?.duplicate_group || item?.duplicateGroup || '',
    duplicateKeys: Array.isArray(duplicateKeys)
      ? duplicateKeys.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    duplicateNote: item?.duplicate_note || item?.duplicateNote || '',
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    isBuiltin: Boolean(item?.is_builtin ?? item?.isBuiltin ?? false),
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizePermissionAuditSummary(item: any): Api.SystemManage.PermissionActionAuditSummary {
  return {
    totalCount: Number(item?.total_count ?? item?.totalCount ?? 0),
    unusedCount: Number(item?.unused_count ?? item?.unusedCount ?? 0),
    apiOnlyCount: Number(item?.api_only_count ?? item?.apiOnlyCount ?? 0),
    pageOnlyCount: Number(item?.page_only_count ?? item?.pageOnlyCount ?? 0),
    packageOnlyCount: Number(item?.package_only_count ?? item?.packageOnlyCount ?? 0),
    multiConsumerCount: Number(item?.multi_consumer_count ?? item?.multiConsumerCount ?? 0),
    crossContextMirrorCount: Number(
      item?.cross_context_mirror_count ?? item?.crossContextMirrorCount ?? 0
    ),
    suspectedDuplicateCount: Number(
      item?.suspected_duplicate_count ?? item?.suspectedDuplicateCount ?? 0
    )
  }
}

export function normalizeApiEndpoint(item: any): Api.SystemManage.APIEndpointItem {
  const permissionKeysRaw = item?.permission_keys || item?.permissionKeys || []
  const permissionContextsRaw = item?.permission_contexts || item?.permissionContexts || []
  const permissionKeys = Array.isArray(permissionKeysRaw)
    ? permissionKeysRaw.map((v: any) => `${v || ''}`.trim()).filter(Boolean)
    : []
  const permissionKey =
    normalizePermissionKey(item?.permission_key || item?.permissionKey) || permissionKeys[0] || ''
  return {
    id: item?.id || '',
    code: item?.code || '',
    appKey: item?.app_key || item?.appKey || '',
    appScope: item?.app_scope || item?.appScope || 'shared',
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || '',
    featureKind: item?.feature_kind || item?.featureKind || 'system',
    handler: item?.handler || '',
    summary: item?.summary || '',
    permissionKey,
    permissionKeys,
    permissionContexts: Array.isArray(permissionContextsRaw)
      ? permissionContextsRaw.map((v: any) => `${v || ''}`.trim()).filter(Boolean)
      : [],
    permissionBindingMode:
      item?.permission_binding_mode ||
      item?.permissionBindingMode ||
      (permissionKeys.length > 1 ? 'shared' : permissionKeys.length === 1 ? 'single' : 'none'),
    sharedAcrossContexts: Boolean(item?.shared_across_contexts ?? item?.sharedAcrossContexts),
    permissionNote: item?.permission_note || item?.permissionNote || '',
    authMode: item?.auth_mode || item?.authMode || (permissionKey ? 'permission' : 'jwt'),
    categoryId: item?.category_id || item?.categoryId || item?.category?.id || '',
    category: item?.category
      ? {
          id: item.category.id || '',
          code: item.category.code || '',
          name: item.category.name || '',
          nameEn: item.category.name_en || item.category.nameEn || '',
          sortOrder: item.category.sort_order ?? item.category.sortOrder ?? 0,
          status: item.category.status || 'normal'
        }
      : undefined,
    contextScope: item?.context_scope || item?.contextScope || 'optional',
    source: item?.source || 'sync',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    runtimeExists: Boolean(item?.runtime_exists ?? item?.runtimeExists),
    stale: Boolean(item?.stale),
    staleReason: item?.stale_reason || item?.staleReason || '',
    status: item?.status || 'normal',
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizeFeaturePackage(item: any): Api.SystemManage.FeaturePackageItem {
  const packageKey = item?.package_key || item?.packageKey || ''
  const workspaceScope = item?.workspace_scope || item?.workspaceScope || 'all'
  const appKeysRaw = item?.app_keys || item?.appKeys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    appKeys,
    packageKey,
    packageType: item?.package_type || item?.packageType || 'base',
    name: item?.name || '',
    description: item?.description || '',
    workspaceScope,
    isBuiltin: Boolean(item?.is_builtin ?? item?.isBuiltin ?? false),
    actionCount: item?.action_count ?? item?.actionCount ?? 0,
    menuCount: item?.menu_count ?? item?.menuCount ?? 0,
    collaborationWorkspaceCount:
      item?.collaborationWorkspaceCount ?? item?.collaboration_workspace_count ?? 0,
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizeRole(item: any): Api.SystemManage.RoleListItem {
  const appKeysRaw = item?.appKeys || item?.app_keys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    roleId: item?.roleId || item?.role_id || item?.id || '',
    roleName: item?.roleName || item?.role_name || item?.name || '',
    roleCode: item?.roleCode || item?.role_code || item?.code || '',
    description: item?.description || '',
    appKeys,
    sortOrder: item?.sortOrder ?? item?.sort_order ?? 0,
    status: item?.status || 'normal',
    priority: item?.priority ?? 0,
    customParams: item?.customParams || item?.custom_params || {},
    createTime: item?.createTime || item?.create_time || item?.created_at || '',
    collaborationWorkspaceId:
      item?.collaborationWorkspaceId || item?.collaboration_workspace_id || null,
    isGlobal: Boolean(item?.isGlobal ?? item?.is_global ?? appKeys.length === 0),
    canEditPermission: Boolean(item?.canEditPermission ?? item?.can_edit_permission ?? true)
  }
}

export function normalizeCollaborationWorkspace(
  item: any
): Api.SystemManage.CollaborationWorkspaceListItem {
  return {
    id: item?.id || '',
    name: item?.name || '',
    remark: item?.remark || '',
    logoUrl: item?.logo_url || item?.logoUrl || '',
    plan: item?.plan || 'free',
    maxMembers: item?.max_members ?? item?.maxMembers ?? 0,
    status: item?.status || 'active',
    createTime: item?.created_at || item?.createTime || '',
    updateTime: item?.updated_at || item?.updateTime || '',
    ownerId: item?.owner_id || item?.ownerId || '',
    adminUsers: item?.admin_users || item?.adminUsers || [],
    adminUserIds: item?.admin_user_ids || item?.adminUserIds || [],
    currentRoleCode: item?.current_role_code || item?.currentRoleCode || '',
    memberStatus: item?.member_status || item?.memberStatus || ''
  }
}

export function normalizePageItem(item: any): Api.SystemManage.PageItem {
  const meta = item?.meta || {}
  const rawVisibilityScope =
    `${item?.visibility_scope || item?.visibilityScope || meta?.visibilityScope || ''}`.trim()
  const rawSpaceKeys = Array.isArray(item?.space_keys || item?.spaceKeys || meta?.spaceKeys)
    ? item?.space_keys || item?.spaceKeys || meta?.spaceKeys
    : []
  // spaceKeys 由后端把 page_space_bindings、菜单继承和父页继承统一编译后下发，前端不再自行猜测。
  const spaceKeys = rawSpaceKeys
    .map((value: any) => normalizeMenuSpaceKey(`${value || ''}`))
    .filter(Boolean)
  const spaceType = `${item?.space_type || item?.spaceType || meta?.spaceType || ''}`.trim()
  const hostKey = `${item?.host_key || item?.hostKey || meta?.hostKey || ''}`.trim()
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    pageKey: item?.page_key || item?.pageKey || '',
    name: item?.name || '',
    routeName: item?.route_name || item?.routeName || '',
    routePath: item?.route_path || item?.routePath || '',
    component: item?.component || '',
    pageType: item?.page_type || item?.pageType || 'inner',
    visibilityScope: rawVisibilityScope || undefined,
    source: item?.source || 'manual',
    moduleKey: item?.module_key || item?.moduleKey || '',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    parentMenuId: item?.parent_menu_id || item?.parentMenuId || '',
    parentMenuName: item?.parent_menu_name || item?.parentMenuName || '',
    parentPageKey: item?.parent_page_key || item?.parentPageKey || '',
    parentPageName: item?.parent_page_name || item?.parentPageName || '',
    displayGroupKey: item?.display_group_key || item?.displayGroupKey || '',
    displayGroupName: item?.display_group_name || item?.displayGroupName || '',
    activeMenuPath: item?.active_menu_path || item?.activeMenuPath || '',
    breadcrumbMode: item?.breadcrumb_mode || item?.breadcrumbMode || 'inherit_menu',
    accessMode: item?.access_mode || item?.accessMode || 'inherit',
    permissionKey: item?.permission_key || item?.permissionKey || '',
    inheritPermission: Boolean(item?.inherit_permission ?? item?.inheritPermission ?? true),
    keepAlive: Boolean(item?.keep_alive ?? item?.keepAlive ?? false),
    isFullPage: Boolean(item?.is_full_page ?? item?.isFullPage ?? false),
    isIframe: Boolean(meta?.isIframe ?? item?.is_iframe ?? item?.isIframe ?? false),
    isHideTab: Boolean(meta?.isHideTab ?? item?.is_hide_tab ?? item?.isHideTab ?? false),
    link: `${meta?.link || item?.link || ''}`.trim(),
    spaceKeys,
    spaceScope:
      `${item?.space_scope || item?.spaceScope || meta?.spaceScope || ''}`.trim() || undefined,
    spaceType,
    hostKey,
    status: item?.status || 'normal',
    meta: {
      ...meta,
      ...(spaceKeys.length ? { spaceKeys } : {}),
      ...(rawVisibilityScope ? { visibilityScope: rawVisibilityScope } : {}),
      ...(`${item?.space_scope || item?.spaceScope || meta?.spaceScope || ''}`.trim()
        ? {
            spaceScope: `${item?.space_scope || item?.spaceScope || meta?.spaceScope || ''}`.trim()
          }
        : {}),
      ...(spaceType ? { spaceType } : {}),
      ...(hostKey ? { hostKey } : {})
    },
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizePageMenuOption(item: any): Api.SystemManage.PageMenuOptionItem {
  return {
    id: item?.id || '',
    name: item?.name || '',
    title: item?.title || '',
    path: item?.path || '',
    children: Array.isArray(item?.children) ? item.children.map(normalizePageMenuOption) : []
  }
}

export function normalizePageUnregisteredItem(item: any): Api.SystemManage.PageUnregisteredItem {
  return {
    filePath: item?.file_path || item?.filePath || '',
    component: item?.component || '',
    pageKey: item?.page_key || item?.pageKey || '',
    name: item?.name || '',
    routeName: item?.route_name || item?.routeName || '',
    routePath: item?.route_path || item?.routePath || '',
    pageType: item?.page_type || item?.pageType || 'inner',
    moduleKey: item?.module_key || item?.moduleKey || '',
    parentMenuId: item?.parent_menu_id || item?.parentMenuId || '',
    parentMenuName: item?.parent_menu_name || item?.parentMenuName || '',
    activeMenuPath: item?.active_menu_path || item?.activeMenuPath || '',
    spaceKey: normalizeMenuSpaceKey(item?.space_key || item?.spaceKey),
    spaceType: `${item?.space_type || item?.spaceType || ''}`.trim(),
    hostKey: `${item?.host_key || item?.hostKey || ''}`.trim()
  }
}

export function normalizeMenuSpace(item: any): Api.SystemManage.MenuSpaceItem {
  const allowedRoleCodes = item?.allowed_role_codes ?? item?.allowedRoleCodes ?? []
  const rawAccessMode =
    item?.access_mode ||
    item?.accessMode ||
    item?.meta?.access_mode ||
    item?.meta?.accessMode ||
    'all'
  const accessMode = `${rawAccessMode}`.trim()
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    spaceKey: normalizeMenuSpaceKey(item?.space_key || item?.spaceKey),
    name: item?.name || '',
    description: item?.description || '',
    defaultHomePath: item?.default_home_path || item?.defaultHomePath || '',
    isDefault: Boolean(item?.is_default ?? item?.isDefault ?? false),
    status: item?.status || 'normal',
    hostCount: item?.host_count ?? item?.hostCount ?? 0,
    hosts: Array.isArray(item?.hosts)
      ? item.hosts.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    menuCount: Number(item?.menu_count ?? item?.menuCount ?? 0),
    pageCount: Number(item?.page_count ?? item?.pageCount ?? 0),
    accessMode,
    allowedRoleCodes: Array.isArray(allowedRoleCodes)
      ? allowedRoleCodes.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    meta: item?.meta || {},
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizeMenuSpaceHostBinding(item: any): Api.SystemManage.MenuSpaceHostBindingItem {
  const meta = item?.meta || {}
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    appName: item?.app_name || item?.appName || '',
    host: `${item?.host || ''}`.trim(),
    spaceKey: normalizeMenuSpaceKey(item?.space_key || item?.spaceKey),
    spaceName: item?.space_name || item?.spaceName || '',
    description: item?.description || '',
    isDefault: Boolean(item?.is_default ?? item?.isDefault ?? false),
    status: item?.status || 'normal',
    scheme: `${item?.scheme || item?.meta?.scheme || meta?.scheme || 'https'}`.trim() || 'https',
    routePrefix:
      `${item?.route_prefix || item?.routePrefix || meta?.route_prefix || meta?.routePrefix || ''}`.trim(),
    authMode:
      `${item?.auth_mode || item?.authMode || meta?.auth_mode || meta?.authMode || 'inherit_host'}`.trim() ||
      'inherit_host',
    loginHost:
      `${item?.login_host || item?.loginHost || meta?.login_host || meta?.loginHost || ''}`.trim(),
    callbackHost:
      `${item?.callback_host || item?.callbackHost || meta?.callback_host || meta?.callbackHost || ''}`.trim(),
    cookieScopeMode:
      `${item?.cookie_scope_mode || item?.cookieScopeMode || meta?.cookie_scope_mode || meta?.cookieScopeMode || 'inherit'}`.trim() ||
      'inherit',
    cookieDomain:
      `${item?.cookie_domain || item?.cookieDomain || meta?.cookie_domain || meta?.cookieDomain || ''}`.trim(),
    meta,
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizeApp(item: any): Api.SystemManage.AppItem {
  const primaryHostsRaw = item?.primary_hosts || item?.primaryHosts || []
  const primaryHosts = Array.isArray(primaryHostsRaw)
    ? primaryHostsRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    name: item?.name || '',
    description: item?.description || '',
    defaultSpaceKey: item?.default_space_key || item?.defaultSpaceKey || '',
    spaceMode: item?.space_mode || item?.spaceMode || 'single',
    isDefault: Boolean(item?.is_default ?? item?.isDefault ?? false),
    status: item?.status || 'normal',
    hostCount: Number(item?.host_count ?? item?.hostCount ?? 0),
    primaryHost: item?.primary_host || item?.primaryHost || primaryHosts[0] || '',
    menuSpaceCount: Number(
      item?.menu_space_count ?? item?.menuSpaceCount ?? item?.space_count ?? item?.spaceCount ?? 0
    ),
    menuCount: Number(item?.menu_count ?? item?.menuCount ?? 0),
    pageCount: Number(item?.page_count ?? item?.pageCount ?? 0),
    meta: item?.meta || {},
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizeAppHostBinding(item: any): Api.SystemManage.AppHostBindingItem {
  return {
    id: item?.id || '',
    appKey: item?.app_key || item?.appKey || '',
    appName: item?.app_name || item?.appName || '',
    host: `${item?.host || ''}`.trim(),
    defaultSpaceKey: item?.default_space_key || item?.defaultSpaceKey || '',
    description: item?.description || '',
    isPrimary: Boolean(item?.is_primary ?? item?.isPrimary ?? false),
    status: item?.status || 'normal',
    meta: item?.meta || {},
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizePageBreadcrumbPreviewItem(item: any): Api.SystemManage.PageBreadcrumbPreviewItem {
  return {
    type: item?.type || 'page',
    title: item?.title || '',
    path: item?.path || '',
    pageKey: item?.page_key || item?.pageKey || ''
  }
}

export function normalizeRefreshStats(item: any): Api.SystemManage.RefreshStats {
  return {
    requestedPackageCount: Number(
      item?.requestedPackageCount ?? item?.requested_package_count ?? 0
    ),
    impactedPackageCount: Number(item?.impactedPackageCount ?? item?.impacted_package_count ?? 0),
    roleCount: Number(item?.roleCount ?? item?.role_count ?? 0),
    collaborationWorkspaceCount: Number(item?.collaborationWorkspaceCount ?? 0),
    userCount: Number(item?.userCount ?? item?.user_count ?? 0),
    elapsedMilliseconds: Number(item?.elapsedMilliseconds ?? item?.elapsed_milliseconds ?? 0),
    finishedAt: item?.finishedAt || item?.finished_at || ''
  }
}

export function normalizeRiskAudit(item: any): Api.SystemManage.RiskAuditItem {
  return {
    id: item?.id || '',
    operatorId: item?.operator_id || item?.operatorId || '',
    objectType: item?.object_type || item?.objectType || '',
    objectId: item?.object_id || item?.objectId || '',
    operationType: item?.operation_type || item?.operationType || '',
    beforeSummary: item?.before_summary || item?.beforeSummary || {},
    afterSummary: item?.after_summary || item?.afterSummary || {},
    impactSummary: item?.impact_summary || item?.impactSummary || {},
    requestId: item?.request_id || item?.requestId || '',
    createdAt: item?.created_at || item?.createdAt || ''
  }
}

export function normalizeFeaturePackageRelationNode(
  item: any
): Api.SystemManage.FeaturePackageRelationNode {
  const packageKey = item?.package_key || item?.packageKey || ''
  const appKeysRaw = item?.app_keys || item?.appKeys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    packageKey,
    name: item?.name || '',
    packageType: item?.package_type || item?.packageType || 'base',
    workspaceScope: item?.workspace_scope || item?.workspaceScope || 'all',
    appKeys,
    status: item?.status || 'normal',
    referenceCount: Number(item?.reference_count ?? item?.referenceCount ?? 0),
    children: Array.isArray(item?.children)
      ? item.children.map((child: any) => normalizeFeaturePackageRelationNode(child))
      : []
  }
}

export function normalizeFeaturePackageRelationTree(
  item: any
): Api.SystemManage.FeaturePackageRelationTree {
  const cycleDependencies = item?.cycle_dependencies || item?.cycleDependencies || []
  const isolatedBaseKeys = item?.isolated_base_keys || item?.isolatedBaseKeys || []
  return {
    roots: Array.isArray(item?.roots)
      ? item.roots.map((node: any) => normalizeFeaturePackageRelationNode(node))
      : [],
    cycleDependencies: Array.isArray(cycleDependencies)
      ? cycleDependencies.map((cycle: any) =>
          Array.isArray(cycle)
            ? cycle.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
            : []
        )
      : [],
    isolatedBaseKeys: Array.isArray(isolatedBaseKeys)
      ? isolatedBaseKeys.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : []
  }
}

export function normalizePermissionActionConsumers(
  item: any
): Api.SystemManage.PermissionActionConsumerDetails {
  const featurePackages = item?.feature_packages || item?.featurePackages || []
  return {
    permissionKey: item?.permission_key || item?.permissionKey || '',
    apis: Array.isArray(item?.apis)
      ? item.apis.map((api: any) => ({
          code: api?.code || '',
          method: api?.method || '',
          path: api?.path || '',
          summary: api?.summary || ''
        }))
      : [],
    pages: Array.isArray(item?.pages)
      ? item.pages.map((page: any) => ({
          pageKey: page?.page_key || page?.pageKey || '',
          name: page?.name || '',
          routePath: page?.route_path || page?.routePath || '',
          accessMode: page?.access_mode || page?.accessMode || ''
        }))
      : [],
    featurePackages: Array.isArray(featurePackages)
      ? featurePackages.map((pkg: any) => ({
          id: pkg?.id || '',
          packageKey: pkg?.package_key || pkg?.packageKey || '',
          name: pkg?.name || '',
          packageType: pkg?.package_type || pkg?.packageType || '',
          contextType: pkg?.context_type || pkg?.contextType || 'common'
        }))
      : [],
    roles: Array.isArray(item?.roles)
      ? item.roles.map((role: any) => ({
          id: role?.id || '',
          code: role?.code || '',
          name: role?.name || '',
          contextType: role?.context_type || role?.contextType || ''
        }))
      : []
  }
}

export function normalizeUnregisteredApiScanConfig(item: any): Api.SystemManage.APIUnregisteredScanConfig {
  return {
    enabled: Boolean(item?.enabled),
    frequencyMinutes: Number(item?.frequency_minutes ?? item?.frequencyMinutes ?? 60),
    defaultCategoryId: item?.default_category_id || item?.defaultCategoryId || '',
    defaultPermissionKey: item?.default_permission_key || item?.defaultPermissionKey || '',
    markAsNoPermission: Boolean(item?.mark_as_no_permission ?? item?.markAsNoPermission)
  }
}

export function normalizePageAccessTraceResult(item: any): Api.SystemManage.PageAccessTraceResult {
  const visibleMenuIds = item?.visible_menu_ids || item?.visibleMenuIds || []
  return {
    userId: item?.user_id || item?.userId || '',
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || '',
    spaceKey: item?.space_key || item?.spaceKey || '',
    authenticated: Boolean(item?.authenticated),
    superAdmin: Boolean(item?.super_admin ?? item?.superAdmin),
    actionKeyCount: Number(item?.action_key_count ?? item?.actionKeyCount ?? 0),
    visibleMenuIds: Array.isArray(visibleMenuIds)
      ? visibleMenuIds.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    roles: Array.isArray(item?.roles)
      ? item.roles.map((role: any) => ({
          roleId: role?.role_id || role?.roleId || '',
          roleCode: role?.role_code || role?.roleCode || '',
          roleName: role?.role_name || role?.roleName || '',
          status: role?.status || ''
        }))
      : [],
    pages: Array.isArray(item?.pages)
      ? item.pages.map((page: any) => ({
          pageKey: page?.page_key || page?.pageKey || '',
          pageName: page?.page_name || page?.pageName || '',
          routePath: page?.route_path || page?.routePath || '',
          accessMode: page?.access_mode || page?.accessMode || '',
          permissionKey: page?.permission_key || page?.permissionKey || '',
          parentPageKey: page?.parent_page_key || page?.parentPageKey || '',
          parentMenuId: page?.parent_menu_id || page?.parentMenuId || '',
          activeMenuPath: page?.active_menu_path || page?.activeMenuPath || '',
          visible: Boolean(page?.visible),
          reason: page?.reason || '',
          matchedActionKey: page?.matched_action_key || page?.matchedActionKey || '',
          effectiveChain: Array.isArray(page?.effective_chain || page?.effectiveChain)
            ? page?.effective_chain || page?.effectiveChain
            : []
        }))
      : []
  }
}

export function normalizeMenuManageGroup(item: any): Api.SystemManage.MenuManageGroupItem {
  return {
    id: item?.id || '',
    name: item?.name || '',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    status: item?.status || 'normal',
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

export function normalizeMenuBackupItem(item: any): Api.SystemManage.MenuBackupItem {
  const scopeType = `${item?.scope_type || item?.scopeType || ''}`.trim()
  const scopeOrigin = `${item?.scope_origin || item?.scopeOrigin || ''}`.trim()
  const rawSpaceKey = `${item?.space_key || item?.spaceKey || ''}`.trim()
  return {
    id: item?.id || '',
    name: item?.name || '',
    description: item?.description || '',
    // 备份列表里空 space_key 仍有业务含义，不能像菜单/页面那样归一成 default。
    space_key: rawSpaceKey || undefined,
    space_name: item?.space_name || item?.spaceName || '',
    scope_type: scopeType || (rawSpaceKey ? 'space' : 'global'),
    // scope_origin 是后端显式返回的来源标签；缺省时退回旧语义，保证兼容旧接口数据。
    scope_origin: scopeOrigin || (rawSpaceKey ? 'space' : 'global'),
    created_at: item?.created_at || item?.createdAt || '',
    created_by: item?.created_by || item?.createdBy || ''
  }
}

export function normalizeRuntimeMenuTree(item: any): AppRouteRecord {
  const meta = item?.meta || {}
  const children = Array.isArray(item?.children)
    ? item.children.map((child: any) => normalizeRuntimeMenuTree(child))
    : []
  const spaceKey = normalizeMenuSpaceKey(item?.space_key || item?.spaceKey || meta?.spaceKey)
  const spaceType = `${item?.space_type || item?.spaceType || meta?.spaceType || ''}`.trim()
  const hostKey = `${item?.host_key || item?.hostKey || meta?.hostKey || ''}`.trim()
  return {
    id: item?.id || '',
    kind: item?.kind || '',
    path: item?.path || '',
    name: item?.name || '',
    component: item?.component || '',
    parent_id: item?.parent_id || item?.parentId || '',
    sort_order: item?.sort_order ?? item?.sortOrder ?? 0,
    redirect: item?.redirect || '',
    spaceKey,
    spaceType,
    hostKey,
    meta: {
      title: meta?.title || '',
      icon: meta?.icon || '',
      // accessMode 由后端运行时菜单显式下发；空值时交给前端/后续链路按默认 permission 处理。
      accessMode: `${meta?.accessMode || ''}`.trim() || undefined,
      // activePath 仅在需要高亮指定菜单时才返回，缺省表示沿用当前路由 path。
      activePath: `${meta?.activePath || ''}`.trim() || undefined,
      // link 仅外链菜单需要，未返回时保持 undefined，避免误判成空链接。
      link: `${meta?.link || ''}`.trim() || undefined,
      spaceKey: spaceKey || undefined,
      spaceType: spaceType || undefined,
      hostKey: hostKey || undefined,
      // 运行时菜单里未返回 isEnable 表示“沿用默认启用”，不能归一成 false，
      // 否则前端兜底过滤会把整棵菜单树误判为禁用。
      isEnable: meta?.isEnable === false ? false : undefined,
      // 后端运行时菜单对这些展示型布尔字段采用“只在为 true 时才返回”的约定，
      // 所以前端归一化时只认显式 true，缺省一律视为未开启该特性。
      isHide: meta?.isHide === true,
      isIframe: meta?.isIframe === true,
      isHideTab: meta?.isHideTab === true,
      keepAlive: meta?.keepAlive === true,
      fixedTab: meta?.fixedTab === true,
      isFullPage: meta?.isFullPage === true
    },
    children
  } as AppRouteRecord
}

export function normalizeRuntimeNavigationManifest(item: any): Api.SystemManage.RuntimeNavigationManifest {
  const currentApp = item?.current_app || item?.currentApp
  const currentSpace = item?.current_space || item?.currentSpace || {}
  const menuTree = item?.menu_tree || item?.menuTree || []
  const entryRoutes = item?.entry_routes || item?.entryRoutes || []
  const managedPages = item?.managed_pages || item?.managedPages || []
  const space = currentSpace?.space ? normalizeMenuSpace(currentSpace.space) : undefined
  const binding = currentSpace?.binding
    ? normalizeMenuSpaceHostBinding(currentSpace.binding)
    : undefined

  return {
    currentApp: currentApp
      ? {
          app: normalizeApp(currentApp?.app || {}),
          binding: currentApp?.binding ? normalizeAppHostBinding(currentApp.binding) : undefined,
          resolvedBy: currentApp?.resolved_by || currentApp?.resolvedBy || '',
          requestHost: currentApp?.request_host || currentApp?.requestHost || ''
        }
      : undefined,
    // currentSpace 是后端对 Host / 显式 space_key 解析后的最终上下文，前端菜单树和受管页面都必须跟随它。
    currentSpace: {
      space,
      binding,
      resolvedBy: currentSpace?.resolved_by || currentSpace?.resolvedBy || '',
      requestHost: currentSpace?.request_host || currentSpace?.requestHost || '',
      accessGranted: Boolean(currentSpace?.access_granted ?? currentSpace?.accessGranted ?? true)
    },
    context: {
      ...(item?.context || {}),
      app_key: item?.context?.app_key || item?.context?.appKey || '',
      space_key: item?.context?.space_key || item?.context?.spaceKey || '',
      requested_space_key:
        item?.context?.requested_space_key || item?.context?.requestedSpaceKey || '',
      request_host: item?.context?.request_host || item?.context?.requestHost || ''
    },
    // menuTree 已经完成启用态、空间和权限裁剪；前端这里只做归一化与动态注册。
    menuTree: Array.isArray(menuTree)
      ? menuTree.map((entry: any) => normalizeRuntimeMenuTree(entry))
      : [],
    entryRoutes: Array.isArray(entryRoutes)
      ? entryRoutes.map((entry: any) => normalizeRuntimeMenuTree(entry))
      : [],
    managedPages: Array.isArray(managedPages)
      ? managedPages.map((entry: any) => normalizePageItem(entry))
      : [],
    versionStamp: item?.version_stamp || item?.versionStamp || ''
  }
}

export function normalizeApiEndpointCategory(item: any): Api.SystemManage.APIEndpointCategoryItem {
  return {
    id: item?.id || '',
    code: item?.code || '',
    name: item?.name || '',
    nameEn: item?.name_en || item?.nameEn || '',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    status: item?.status || 'normal'
  }
}

export function normalizeUnregisteredApiRoute(item: any): Api.SystemManage.APIUnregisteredRouteItem {
  return {
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || `${item?.method || ''} ${item?.path || ''}`.trim(),
    handler: item?.handler || '',
    hasMeta: Boolean(item?.has_meta ?? item?.hasMeta),
    meta: item?.meta
      ? {
          summary: item.meta.summary || '',
          category_code: item.meta.category_code || '',
          context_scope: item.meta.context_scope || 'optional',
          source: item.meta.source || '',
          feature_kind: item.meta.feature_kind || 'system',
          permission_keys: Array.isArray(item.meta.permission_keys) ? item.meta.permission_keys : []
        }
      : undefined
  }
}

export function normalizeFastEnterConfig(item: any): Api.SystemManage.FastEnterConfig {
  const applications = Array.isArray(item?.applications) ? item.applications : []
  const quickLinks = Array.isArray(item?.quickLinks)
    ? item.quickLinks
    : Array.isArray(item?.quick_links)
      ? item.quick_links
      : []

  return {
    applications: applications.map(
      (entry: any): FastEnterApplication => ({
        id: entry?.id || '',
        name: entry?.name || '',
        description: entry?.description || '',
        icon: entry?.icon || 'ri:apps-2-line',
        iconColor: entry?.iconColor || entry?.icon_color || '#377dff',
        enabled: entry?.enabled !== false,
        order: entry?.order ?? 0,
        routeName: entry?.routeName || entry?.route_name || '',
        link: entry?.link || ''
      })
    ),
    quickLinks: quickLinks.map(
      (entry: any): FastEnterQuickLink => ({
        id: entry?.id || '',
        name: entry?.name || '',
        enabled: entry?.enabled !== false,
        order: entry?.order ?? 0,
        routeName: entry?.routeName || entry?.route_name || '',
        link: entry?.link || ''
      })
    ),
    minWidth: item?.minWidth ?? item?.min_width ?? 1450
  }
}

export function normalizeUserPermissionMenuTree(item: any): Api.SystemManage.UserPermissionMenuNode {
  return {
    id: item?.id || '',
    name: item?.name || '',
    title: item?.title || '',
    path: item?.path || '',
    component: item?.component || '',
    hidden: Boolean(item?.hidden),
    sort: item?.sort ?? 0,
    children: (item?.children || []).map((child: any) => normalizeUserPermissionMenuTree(child))
  }
}

export function normalizeUserCollaborationWorkspaceItem(
  item: any
): Api.SystemManage.CollaborationWorkspaceListItem {
  return {
    id: item?.id || '',
    name: item?.name || '',
    remark: item?.remark || '',
    logoUrl: item?.logo_url || item?.logoUrl || '',
    plan: item?.plan || 'free',
    maxMembers: item?.max_members ?? item?.maxMembers ?? 0,
    status: item?.status || 'active',
    createTime: item?.created_at || item?.createTime || '',
    updateTime: item?.updated_at || item?.updateTime || '',
    adminUsers: item?.admin_users || item?.adminUsers || [],
    adminUserIds: item?.admin_user_ids || item?.adminUserIds || [],
    currentRoleCode: item?.current_role_code || item?.currentRoleCode || '',
    memberStatus: item?.member_status || item?.memberStatus || ''
  }
}

export function normalizeUserPermissionDiagnosisResponse(
  item: any
): Api.SystemManage.UserPermissionDiagnosisResponse {
  const normalizePackages = (items: any[] | undefined) => (items || []).map(normalizeFeaturePackage)
  const normalizeGroup = (value: any): Api.SystemManage.PermissionGroupItem | undefined =>
    value
      ? {
          id: value?.id || '',
          groupType: value?.group_type || value?.groupType || '',
          code: value?.code || '',
          name: value?.name || '',
          nameEn: value?.name_en || value?.nameEn || '',
          description: value?.description || '',
          status: value?.status || 'normal',
          sortOrder: value?.sort_order ?? value?.sortOrder ?? 0,
          isBuiltin: Boolean(value?.is_builtin ?? value?.isBuiltin ?? false)
        }
      : undefined

  const diagnosis = item?.diagnosis
    ? {
        permissionKey: item.diagnosis?.permission_key || item.diagnosis?.permissionKey || '',
        allowed: Boolean(item.diagnosis?.allowed),
        reasonText: item.diagnosis?.reason_text || item.diagnosis?.reasonText || '',
        reasons: item.diagnosis?.reasons || [],
        matchedInSnapshot: Boolean(
          item.diagnosis?.matched_in_snapshot ?? item.diagnosis?.matchedInSnapshot
        ),
        bypassedBySuperAdmin: Boolean(
          item.diagnosis?.bypassed_by_super_admin ?? item.diagnosis?.bypassedBySuperAdmin
        ),
        blockedByCollaborationWorkspace: Boolean(
          item.diagnosis?.blocked_by_collaboration_workspace ??
          item.diagnosis?.blockedByCollaborationWorkspace
        ),
        denialStage: item.diagnosis?.denial_stage || item.diagnosis?.denialStage || '',
        denialReason: item.diagnosis?.denial_reason || item.diagnosis?.denialReason || '',
        memberStatus: item.diagnosis?.member_status || item.diagnosis?.memberStatus || '',
        memberMatched: Boolean(item.diagnosis?.member_matched ?? item.diagnosis?.memberMatched),
        boundaryState: item.diagnosis?.boundary_state || item.diagnosis?.boundaryState || '',
        boundaryConfigured: Boolean(
          item.diagnosis?.boundary_configured ?? item.diagnosis?.boundaryConfigured
        ),
        roleChainMatched: Boolean(
          item.diagnosis?.role_chain_matched ?? item.diagnosis?.roleChainMatched
        ),
        roleChainDisabled: Boolean(
          item.diagnosis?.role_chain_disabled ?? item.diagnosis?.roleChainDisabled
        ),
        roleChainAvailable: Boolean(
          item.diagnosis?.role_chain_available ?? item.diagnosis?.roleChainAvailable
        ),
        action: item.diagnosis?.action
          ? {
              id: item.diagnosis.action?.id || '',
              permissionKey:
                item.diagnosis.action?.permission_key || item.diagnosis.action?.permissionKey || '',
              name: item.diagnosis.action?.name || '',
              description: item.diagnosis.action?.description || '',
              status: item.diagnosis.action?.status || '',
              selfStatus:
                item.diagnosis.action?.self_status || item.diagnosis.action?.selfStatus || '',
              contextType:
                item.diagnosis.action?.context_type || item.diagnosis.action?.contextType || '',
              featureKind:
                item.diagnosis.action?.feature_kind || item.diagnosis.action?.featureKind || '',
              moduleCode:
                item.diagnosis.action?.module_code || item.diagnosis.action?.moduleCode || '',
              moduleGroupStatus:
                item.diagnosis.action?.module_group_status ||
                item.diagnosis.action?.moduleGroupStatus ||
                '',
              featureGroupStatus:
                item.diagnosis.action?.feature_group_status ||
                item.diagnosis.action?.featureGroupStatus ||
                '',
              moduleGroup: normalizeGroup(
                item.diagnosis.action?.module_group || item.diagnosis.action?.moduleGroup
              ),
              featureGroup: normalizeGroup(
                item.diagnosis.action?.feature_group || item.diagnosis.action?.featureGroup
              )
            }
          : null,
        sourcePackages: normalizePackages(
          item.diagnosis?.source_packages || item.diagnosis?.sourcePackages
        ),
        roleResults: (item.diagnosis?.role_results || item.diagnosis?.roleResults || []).map(
          (role: any) => ({
            roleId: role?.role_id || role?.roleId || '',
            roleCode: role?.role_code || role?.roleCode || '',
            roleName: role?.role_name || role?.roleName || '',
            inherited: Boolean(role?.inherited),
            refreshedAt: role?.refreshed_at || role?.refreshedAt || '',
            availableActionCount: role?.available_action_count ?? role?.availableActionCount ?? 0,
            disabledActionCount: role?.disabled_action_count ?? role?.disabledActionCount ?? 0,
            effectiveActionCount: role?.effective_action_count ?? role?.effectiveActionCount ?? 0,
            matched: Boolean(role?.matched),
            disabled: Boolean(role?.disabled),
            available: Boolean(role?.available),
            sourcePackages: normalizePackages(role?.source_packages || role?.sourcePackages)
          })
        )
      }
    : null

  const context = {
    type: item?.context?.type || 'personal',
    collaborationWorkspaceId:
      item?.context?.collaboration_workspace_id || item?.context?.collaborationWorkspaceId || '',
    collaborationWorkspaceName:
      item?.context?.collaboration_workspace_name || item?.context?.collaborationWorkspaceName || ''
  } as Api.SystemManage.UserPermissionContext & {
    collaborationWorkspaceName?: string
  }

  return {
    user: {
      id: item?.user?.id || '',
      userName: item?.user?.user_name || item?.user?.userName || '',
      nickName: item?.user?.nick_name || item?.user?.nickName || '',
      status: item?.user?.status || 'inactive',
      isSuperAdmin: Boolean(item?.user?.is_super_admin ?? item?.user?.isSuperAdmin)
    },
    context,
    snapshot: {
      refreshedAt: item?.snapshot?.refreshed_at || item?.snapshot?.refreshedAt || '',
      updatedAt: item?.snapshot?.updated_at || item?.snapshot?.updatedAt || '',
      roleCount: item?.snapshot?.role_count ?? item?.snapshot?.roleCount ?? 0,
      directPackageCount:
        item?.snapshot?.direct_package_count ?? item?.snapshot?.directPackageCount ?? 0,
      expandedPackageCount:
        item?.snapshot?.expanded_package_count ?? item?.snapshot?.expandedPackageCount ?? 0,
      actionCount: item?.snapshot?.action_count ?? item?.snapshot?.actionCount ?? 0,
      disabledActionCount:
        item?.snapshot?.disabled_action_count ?? item?.snapshot?.disabledActionCount ?? 0,
      menuCount: item?.snapshot?.menu_count ?? item?.snapshot?.menuCount ?? 0,
      hasPackageConfig: Boolean(
        item?.snapshot?.has_package_config ?? item?.snapshot?.hasPackageConfig
      ),
      derivedActionCount:
        item?.snapshot?.derived_action_count ?? item?.snapshot?.derivedActionCount ?? 0,
      blockedActionCount:
        item?.snapshot?.blocked_action_count ?? item?.snapshot?.blockedActionCount ?? 0,
      effectiveActionCount:
        item?.snapshot?.effective_action_count ?? item?.snapshot?.effectiveActionCount ?? 0
    },
    roles: (item?.roles || []).map((role: any) => ({
      roleId: role?.role_id || role?.roleId || '',
      roleCode: role?.role_code || role?.roleCode || '',
      roleName: role?.role_name || role?.roleName || '',
      inherited: Boolean(role?.inherited),
      refreshedAt: role?.refreshed_at || role?.refreshedAt || '',
      availableActionCount: role?.available_action_count ?? role?.availableActionCount ?? 0,
      disabledActionCount: role?.disabled_action_count ?? role?.disabledActionCount ?? 0,
      effectiveActionCount: role?.effective_action_count ?? role?.effectiveActionCount ?? 0,
      matched: Boolean(role?.matched),
      disabled: Boolean(role?.disabled),
      available: Boolean(role?.available),
      sourcePackages: normalizePackages(role?.source_packages || role?.sourcePackages)
    })),
    collaborationWorkspaceMember: item?.collaboration_workspace_member
      ? {
          id: item?.collaboration_workspace_member?.id || '',
          collaborationWorkspaceId:
            item?.collaboration_workspace_member?.collaboration_workspace_id || '',
          userId: item?.collaboration_workspace_member?.user_id || '',
          roleCode: item?.collaboration_workspace_member?.role_code || '',
          status: item?.collaboration_workspace_member?.status || '',
          matched: Boolean(item?.collaboration_workspace_member?.matched)
        }
      : null,
    collaborationWorkspacePackages: normalizePackages(item?.collaboration_workspace_packages),
    diagnosis,
    menus: (item?.menus || []).map((menu: any) => normalizeUserPermissionMenuTree(menu))
  }
}
