import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

const USER_BASE = '/api/v1/users'
const ROLE_BASE = '/api/v1/roles'
const ACTION_PERMISSION_BASE = '/api/v1/permission-actions'
const FEATURE_PACKAGE_BASE = '/api/v1/feature-packages'
const TENANT_BASE = '/api/v1/tenants'
const SYSTEM_BASE = '/api/v1/system'
const API_ENDPOINT_BASE = '/api/v1/api-endpoints'

function normalizePermissionKey(value?: string) {
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

function derivePermissionSegments(permissionKey?: string) {
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

function deriveContextType(permissionKey?: string, moduleCode?: string) {
  const key = `${permissionKey || ''}`.trim()
  const module = `${moduleCode || ''}`.trim()
  if (
    key.startsWith('system.') ||
    key.startsWith('tenant.') ||
    key.startsWith('platform.') ||
    key === 'tenant.manage' ||
    module === 'role' ||
    module === 'user' ||
    module === 'menu' ||
    module === 'menu_backup' ||
    module === 'permission_action' ||
    module === 'api_endpoint' ||
    module === 'feature_package'
  ) {
    return 'platform'
  }
  return 'team'
}

function normalizePermissionAction(item: any): Api.SystemManage.PermissionActionItem {
  const permissionKey = normalizePermissionKey(item?.permission_key || item?.permissionKey)
  const legacy = derivePermissionSegments(permissionKey)
  const moduleCode = item?.module_code || item?.moduleCode || legacy.resourceCode || ''
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
  const moduleGroup = normalizeGroup(item?.module_group || item?.moduleGroup)
  const featureGroup = normalizeGroup(item?.feature_group || item?.featureGroup)
  return {
    id: item?.id || '',
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
    name: item?.name || '',
    description: item?.description || '',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    isBuiltin: Boolean(item?.is_builtin ?? item?.isBuiltin ?? false),
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

function normalizeApiEndpoint(item: any): Api.SystemManage.APIEndpointItem {
  const permissionKeysRaw = item?.permission_keys || item?.permissionKeys || []
  const permissionKeys = Array.isArray(permissionKeysRaw)
    ? permissionKeysRaw.map((v: any) => `${v || ''}`.trim()).filter(Boolean)
    : []
  const permissionKey = normalizePermissionKey(item?.permission_key || item?.permissionKey) || permissionKeys[0] || ''
  return {
    id: item?.id || '',
    code: item?.code || '',
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || '',
    module: item?.module || '',
    featureKind: item?.feature_kind || item?.featureKind || 'system',
    handler: item?.handler || '',
    summary: item?.summary || '',
    permissionKey,
    permissionKeys,
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
    status: item?.status || 'normal',
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

function normalizeFeaturePackage(item: any): Api.SystemManage.FeaturePackageItem {
  const packageKey = item?.package_key || item?.packageKey || ''
  const contextType =
    item?.context_type ||
    item?.contextType ||
    (packageKey.startsWith('platform.')
      ? 'platform'
      : packageKey.startsWith('common.')
        ? 'common'
        : 'team')
  return {
    id: item?.id || '',
    packageKey,
    packageType: item?.package_type || item?.packageType || 'base',
    name: item?.name || '',
    description: item?.description || '',
    contextType,
    isBuiltin: Boolean(item?.is_builtin ?? item?.isBuiltin ?? false),
    actionCount: item?.action_count ?? item?.actionCount ?? 0,
    menuCount: item?.menu_count ?? item?.menuCount ?? 0,
    teamCount: item?.team_count ?? item?.teamCount ?? 0,
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

// 获取用户列表
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: USER_BASE,
    params
  })
}

/** 获取用户平台功能包 */
export function fetchGetUserPackages(userId: string) {
  return request
    .get<Api.SystemManage.UserFeaturePackageResponse>({
      url: `${USER_BASE}/${userId}/packages`,
      skipTenantHeader: true
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置用户平台功能包 */
export function fetchSetUserPackages(userId: string, packageIds: string[]) {
  return request.put<void>({
    url: `${USER_BASE}/${userId}/packages`,
    skipTenantHeader: true,
    data: { package_ids: packageIds }
  })
}

function normalizeApiEndpointCategory(item: any): Api.SystemManage.APIEndpointCategoryItem {
  return {
    id: item?.id || '',
    code: item?.code || '',
    name: item?.name || '',
    nameEn: item?.name_en || item?.nameEn || '',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    status: item?.status || 'normal'
  }
}

function normalizeUnregisteredApiRoute(item: any): Api.SystemManage.APIUnregisteredRouteItem {
  return {
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || `${item?.method || ''} ${item?.path || ''}`.trim(),
    handler: item?.handler || '',
    module: item?.module || '',
    hasMeta: Boolean(item?.has_meta ?? item?.hasMeta),
    meta: item?.meta
      ? {
          summary: item.meta.summary || '',
          module: item.meta.module || '',
          category_code: item.meta.category_code || '',
          context_scope: item.meta.context_scope || 'optional',
          source: item.meta.source || '',
          feature_kind: item.meta.feature_kind || 'system',
          permission_keys: Array.isArray(item.meta.permission_keys)
            ? item.meta.permission_keys
            : []
        }
      : undefined
  }
}

// 获取用户详情
export function fetchGetUser(id: string) {
  return request.get<Api.SystemManage.UserListItem>({
    url: `${USER_BASE}/${id}`
  })
}

// 创建用户
export function fetchCreateUser(data: Api.SystemManage.UserCreateParams) {
  return request.post<{ id: string }>({
    url: USER_BASE,
    data
  })
}

// 更新用户
export function fetchUpdateUser(id: string, data: Api.SystemManage.UserUpdateParams) {
  return request.put<void>({
    url: `${USER_BASE}/${id}`,
    data
  })
}

// 删除用户
export function fetchDeleteUser(id: string) {
  return request.del<void>({
    url: `${USER_BASE}/${id}`
  })
}

// 分配用户角色
export function fetchAssignUserRoles(id: string, roleIds: string[]) {
  return request.post<void>({
    url: `${USER_BASE}/${id}/roles`,
    data: { roleIds }
  })
}

/** 获取平台用户菜单裁剪 */
export async function fetchGetUserMenus(userId: string) {
  const res = await request.get<Api.SystemManage.UserMenuBoundaryResponse>({
    url: `${USER_BASE}/${userId}/menus`,
    skipTenantHeader: true
  })
  return {
    menu_ids: res?.menu_ids || [],
    available_menu_ids: res?.available_menu_ids || [],
    hidden_menu_ids: res?.hidden_menu_ids || [],
    expanded_package_ids: res?.expanded_package_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      menu_id: item?.menu_id || '',
      package_ids: item?.package_ids || []
    })),
    has_package_config: Boolean(res?.has_package_config)
  }
}

/** 设置平台用户菜单裁剪 */
export function fetchSetUserMenus(userId: string, menuIds: string[]) {
  return request.put<void>({
    url: `${USER_BASE}/${userId}/menus`,
    skipTenantHeader: true,
    data: { menu_ids: menuIds }
  })
}

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: ROLE_BASE,
    params
  })
}

// 获取角色列表（简单列表，用于下拉等）
export function fetchGetRoleListSimple() {
  return request.get<Api.SystemManage.RoleList>({
    url: ROLE_BASE,
    params: { current: 1, size: 100 }
  })
}

// 获取角色详情
export function fetchGetRole(id: string) {
  return request.get<Api.SystemManage.RoleListItem>({
    url: `${ROLE_BASE}/${id}`
  })
}

// 创建角色
export function fetchCreateRole(data: Api.SystemManage.RoleCreateParams) {
  return request.post<{ roleId: string }>({
    url: ROLE_BASE,
    data
  })
}

// 更新角色
export function fetchUpdateRole(id: string, data: Api.SystemManage.RoleUpdateParams) {
  return request.put<void>({
    url: `${ROLE_BASE}/${id}`,
    data
  })
}

// 删除角色
export function fetchDeleteRole(id: string) {
  return request.del<void>({
    url: `${ROLE_BASE}/${id}`
  })
}

/** 获取角色已分配的菜单 ID 列表（用于菜单权限配置） */
export function fetchGetRoleMenus(roleId: string) {
  return request
    .get<Api.SystemManage.RoleMenuBoundaryResponse>({
      url: `${ROLE_BASE}/${roleId}/menus`
    })
    .then((res) => ({
      menu_ids: res?.menu_ids || [],
      available_menu_ids: res?.available_menu_ids || [],
      hidden_menu_ids: res?.hidden_menu_ids || [],
      expanded_package_ids: res?.expanded_package_ids || [],
      derived_sources: res?.derived_sources || []
    }))
}

/** 获取角色功能包 */
export function fetchGetRolePackages(roleId: string) {
  return request
    .get<Api.SystemManage.RoleFeaturePackageResponse>({
      url: `${ROLE_BASE}/${roleId}/packages`
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置角色功能包 */
export function fetchSetRolePackages(roleId: string, packageIds: string[]) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/packages`,
    data: { package_ids: packageIds }
  })
}

/** 设置角色菜单权限 */
export function fetchSetRoleMenus(roleId: string, menuIds: string[]) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/menus`,
    data: { menu_ids: menuIds }
  })
}

/** 获取角色功能权限 */
export function fetchGetRoleActions(roleId: string) {
  return request
    .get<Api.SystemManage.RoleActionBoundaryResponse>({
      url: `${ROLE_BASE}/${roleId}/actions`
    })
    .then((res) => ({
      action_ids: res?.action_ids || [],
      available_action_ids: res?.available_action_ids || [],
      disabled_action_ids: res?.disabled_action_ids || [],
      actions: (res?.actions || []).map(normalizePermissionAction),
      expanded_package_ids: res?.expanded_package_ids || [],
      derived_sources: res?.derived_sources || []
    }))
}

/** 设置角色功能权限 */
export function fetchSetRoleActions(roleId: string, actionIds: string[]) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/actions`,
    data: { action_ids: actionIds }
  })
}

/** 获取角色数据权限 */
export async function fetchGetRoleDataPermissions(roleId: string) {
  return request.get<{
    permissions: Array<{ resource_code: string; data_scope: string }>
    resources: Array<{ resource_code: string; resource_name: string }>
    available_data_scopes: Array<{ data_scope: string; label: string }>
  }>({
    url: `${ROLE_BASE}/${roleId}/data-permissions`
  })
}

/** 设置角色数据权限 */
export function fetchSetRoleDataPermissions(
  roleId: string,
  permissions: Array<{ resource_code: string; data_scope: string }>
) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/data-permissions`,
    data: { permissions }
  })
}

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
    permissionKey: undefined,
    moduleCode: undefined,
    moduleGroupId: undefined,
    featureGroupId: undefined,
    contextType: undefined,
    featureKind: undefined,
    isBuiltin: undefined
  }
  return request
    .get<Api.SystemManage.PermissionActionList>({
      url: ACTION_PERMISSION_BASE,
      params: normalizedParams
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizePermissionAction)
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

/** 获取功能包列表 */
export function fetchGetFeaturePackageList(params: Api.SystemManage.FeaturePackageSearchParams) {
  const normalizedParams = {
    ...params,
    package_key: params?.packageKey,
    package_type: params?.packageType,
    context_type: params?.contextType,
    packageKey: undefined,
    packageType: undefined,
    contextType: undefined
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

/** 获取功能包详情 */
export function fetchGetFeaturePackage(id: string) {
  return request
    .get<Api.SystemManage.FeaturePackageItem>({
      url: `${FEATURE_PACKAGE_BASE}/${id}`
    })
    .then((res) => normalizeFeaturePackage(res))
}

export function fetchGetPermissionGroupList(params: Api.SystemManage.PermissionGroupSearchParams) {
  const normalizedParams = {
    ...params,
    group_type: params?.groupType,
    groupType: undefined
  }
  return request.get<Api.SystemManage.PermissionGroupList>({
    url: `${ACTION_PERMISSION_BASE}/groups`,
    params: normalizedParams
  })
}

export function fetchCreatePermissionGroup(data: Api.SystemManage.PermissionGroupSaveParams) {
  return request.post<{ id: string }>({
    url: `${ACTION_PERMISSION_BASE}/groups`,
    data
  })
}

export function fetchUpdatePermissionGroup(id: string, data: Api.SystemManage.PermissionGroupSaveParams) {
  return request.put<void>({
    url: `${ACTION_PERMISSION_BASE}/groups/${id}`,
    data
  })
}

/** 获取组合包基础包 */
export function fetchGetFeaturePackageChildren(id: string) {
  return request
    .get<Api.SystemManage.FeaturePackageBundleResponse>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/children`
    })
    .then((res) => ({
      child_package_ids: res?.child_package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置组合包基础包 */
export function fetchSetFeaturePackageChildren(
  id: string,
  childPackageIds: string[] | Api.SystemManage.FeaturePackageChildSetParams
) {
  const payload = Array.isArray(childPackageIds)
    ? { child_package_ids: childPackageIds }
    : childPackageIds
  return request.put<void>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/children`,
    data: payload
  })
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
  return request.put<void>({
    url: `${FEATURE_PACKAGE_BASE}/${id}`,
    data
  })
}

/** 删除功能包 */
export function fetchDeleteFeaturePackage(id: string) {
  return request.del<void>({
    url: `${FEATURE_PACKAGE_BASE}/${id}`
  })
}

/** 获取功能包包含的功能权限 */
export function fetchGetFeaturePackageActions(id: string) {
  return request
    .get<Api.SystemManage.FeaturePackageActionResponse>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/actions`
    })
    .then((res) => ({
      action_ids: res?.action_ids || [],
      actions: (res?.actions || []).map(normalizePermissionAction)
    }))
}

/** 设置功能包包含的功能权限 */
export function fetchSetFeaturePackageActions(
  id: string,
  actionIds: string[] | Api.SystemManage.FeaturePackageActionSetParams
) {
  const payload = Array.isArray(actionIds) ? { action_ids: actionIds } : actionIds
  return request.put<void>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/actions`,
    data: payload
  })
}

/** 获取功能包包含的菜单 */
export function fetchGetFeaturePackageMenus(id: string) {
  return request
    .get<Api.SystemManage.FeaturePackageMenuResponse>({
      url: `${FEATURE_PACKAGE_BASE}/${id}/menus`
    })
    .then((res) => ({
      menu_ids: res?.menu_ids || [],
      menus: res?.menus || []
    }))
}

/** 设置功能包包含的菜单 */
export function fetchSetFeaturePackageMenus(id: string, menuIds: string[]) {
  return request.put<void>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/menus`,
    data: { menu_ids: menuIds }
  })
}

/** 获取已开通当前功能包的团队 */
export function fetchGetFeaturePackageTeams(id: string) {
  return request.get<Api.SystemManage.FeaturePackageTeamBinding>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/teams`
  })
}

/** 配置功能包开通团队 */
export function fetchSetFeaturePackageTeams(
  id: string,
  teamIds: string[] | Api.SystemManage.FeaturePackageTeamSetParams
) {
  const payload = Array.isArray(teamIds) ? { team_ids: teamIds } : teamIds
  return request.put<void>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/teams`,
    data: payload
  })
}

/** 获取团队已开通的功能包 */
export function fetchGetTeamFeaturePackages(teamId: string) {
  return request
    .get<Api.SystemManage.TeamFeaturePackageResponse>({
      url: `${FEATURE_PACKAGE_BASE}/teams/${teamId}`
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置团队功能包 */
export function fetchSetTeamFeaturePackages(
  teamId: string,
  packageIds: string[] | Api.SystemManage.TeamFeaturePackageSetParams
) {
  const payload = Array.isArray(packageIds) ? { package_ids: packageIds } : packageIds
  return request.put<void>({
    url: `${FEATURE_PACKAGE_BASE}/teams/${teamId}`,
    data: payload
  })
}

/** 获取 API 注册表 */
export function fetchGetApiEndpointList(params: Api.SystemManage.APIEndpointSearchParams) {
  const normalizedParams = {
    permission_key: params?.permissionKey,
    keyword: params?.keyword,
    method: params?.method,
    path: params?.path,
    module: params?.module,
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

/** 同步 API 注册表 */
export function fetchSyncApiEndpoints() {
  return request.post<void>({
    url: `${API_ENDPOINT_BASE}/sync`
  })
}

export function fetchCreateApiEndpoint(data: Partial<Api.SystemManage.APIEndpointItem>) {
  return request.post<Api.SystemManage.APIEndpointItem>({
    url: API_ENDPOINT_BASE,
    data
  })
}

export function fetchUpdateApiEndpoint(id: string, data: Partial<Api.SystemManage.APIEndpointItem>) {
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
  module?: string
  keyword?: string
  only_no_meta?: boolean
}) {
  return request
    .get<Api.SystemManage.APIUnregisteredRouteList>({
      url: `${API_ENDPOINT_BASE}/unregistered`,
      params
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeUnregisteredApiRoute)
    }))
}

export function fetchCreateApiEndpointCategory(data: Partial<Api.SystemManage.APIEndpointCategoryItem>) {
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

/** 重建 API/权限/功能包基础数据（保留菜单、默认管理员与默认角色） */
const MENU_BASE = '/api/v1/menus'

/** 获取菜单树（按当前用户角色过滤，用于侧栏；后端菜单模式时使用） */
export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: `${MENU_BASE}/tree`
  })
}

/** 获取完整菜单树（不限角色，用于菜单管理页；需管理员） */
export function fetchGetMenuTreeAll() {
  return request.get<AppRouteRecord[]>({
    url: `${MENU_BASE}/tree`,
    params: { all: 1 }
  })
}

/** 枚举 views 页面文件（后端 Redis 缓存，支持强制刷新） */
export function fetchGetViewPages(force = false) {
  return request.get<{
    pages: Array<{ filePath: string; componentPath: string }>
    refreshed: boolean
    refreshedAt: string
  }>({
    url: `${SYSTEM_BASE}/view-pages`,
    params: force ? { force: 1 } : undefined
  })
}

/** 创建菜单 */
export function fetchCreateMenu(data: Api.SystemManage.MenuCreateParams, config?: any) {
  return request.post<{ id: string }>({
    url: MENU_BASE,
    data,
    ...config
  })
}

/** 更新菜单 */
export function fetchUpdateMenu(id: string, data: Api.SystemManage.MenuUpdateParams, config?: any) {
  return request.put<void>({
    url: `${MENU_BASE}/${id}`,
    data,
    ...config
  })
}

/** 删除菜单 */
export function fetchDeleteMenu(id: string) {
  return request.del<void>({
    url: `${MENU_BASE}/${id}`
  })
}

/** 更新菜单排序（全量重排） */
export function fetchUpdateMenuSort(data: { id: string; sort_order: number }[]) {
  return request.put<void>({
    url: `${MENU_BASE}/sort`,
    data
  })
}

/** 根据父级ID更新子节点排序（全量重排） */
export function fetchUpdateMenuSortByParent(parentId: string | null, menuIds: string[]) {
  return request.put<void>({
    url: `${MENU_BASE}/sort-by-parent`,
    data: {
      parent_id: parentId,
      menu_ids: menuIds
    }
  })
}

// 菜单备份相关API
const MENU_BACKUP_BASE = '/api/v1/menus/backups'

/** 创建菜单备份 */
export function fetchCreateMenuBackup(data: { name: string; description?: string }) {
  return request.post<void>({
    url: MENU_BACKUP_BASE,
    data
  })
}

/** 获取菜单备份列表 */
export function fetchGetMenuBackupList() {
  return request.get<
    {
      id: string
      name: string
      description: string
      created_at: string
      created_by: string
    }[]
  >({
    url: MENU_BACKUP_BASE
  })
}

/** 删除菜单备份 */
export function fetchDeleteMenuBackup(id: string) {
  return request.del<void>({
    url: `${MENU_BACKUP_BASE}/${id}`
  })
}

/** 恢复菜单备份 */
export function fetchRestoreMenuBackup(id: string) {
  return request.post<void>({
    url: `${MENU_BACKUP_BASE}/${id}/restore`
  })
}
