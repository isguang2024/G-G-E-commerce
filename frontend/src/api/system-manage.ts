import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

const USER_BASE = '/api/v1/users'
const ROLE_BASE = '/api/v1/roles'
const ACTION_PERMISSION_BASE = '/api/v1/permission-actions'
const FEATURE_PACKAGE_BASE = '/api/v1/feature-packages'
const TENANT_BASE = '/api/v1/tenants'
const SYSTEM_BASE = '/api/v1/system'
const API_ENDPOINT_BASE = '/api/v1/api-endpoints'

function normalizePermissionKey(value?: string, resourceCode?: string, actionCode?: string) {
  const target = `${value || ''}`.trim()
  if (target) {
    if (target.includes(':')) {
      const [resource, action] = target.split(':', 2)
      return [resource, action].filter(Boolean).join('.')
    }
    return target
  }
  const fallbackResource = `${resourceCode || ''}`.trim()
  const fallbackAction = `${actionCode || ''}`.trim()
  return fallbackResource && fallbackAction ? `${fallbackResource}.${fallbackAction}` : ''
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
  const resourceCode = item?.resource_code || item?.resourceCode || ''
  const actionCode = item?.action_code || item?.actionCode || ''
  const permissionKey = normalizePermissionKey(
    item?.permission_key || item?.permissionKey,
    resourceCode,
    actionCode
  )
  const moduleCode = item?.module_code || item?.moduleCode || resourceCode || ''
  return {
    id: item?.id || '',
    resourceCode,
    actionCode,
    moduleCode,
    contextType:
      item?.context_type || item?.contextType || deriveContextType(permissionKey, moduleCode),
    permissionKey,
    source: item?.source || 'business',
    featureKind: item?.feature_kind || item?.featureKind || 'system',
    name: item?.name || '',
    description: item?.description || '',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

function normalizeApiEndpoint(item: any): Api.SystemManage.APIEndpointItem {
  const resourceCode = item?.resource_code || item?.resourceCode || ''
  const actionCode = item?.action_code || item?.actionCode || ''
  const permissionKey = normalizePermissionKey(
    item?.permission_key || item?.permissionKey,
    resourceCode,
    actionCode
  )
  return {
    id: item?.id || '',
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || '',
    module: item?.module || '',
    featureKind: item?.feature_kind || item?.featureKind || 'system',
    handler: item?.handler || '',
    summary: item?.summary || '',
    permissionKey,
    authMode: item?.auth_mode || item?.authMode || (permissionKey ? 'permission' : 'jwt'),
    resourceCode,
    actionCode,
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
    (packageKey.startsWith('platform.') ? 'platform' : 'team')
  return {
    id: item?.id || '',
    packageKey,
    name: item?.name || '',
    description: item?.description || '',
    contextType,
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

/** 获取用户权限（计算后的菜单权限） */
export function fetchGetUserPermissions(userId: string) {
  return request.get<any[]>({
    url: `${USER_BASE}/${userId}/permissions`
  })
}

/** 获取用户平台功能包 */
export function fetchGetUserPackages(userId: string) {
  return request
    .get<Api.SystemManage.UserFeaturePackageResponse>({
      url: `${USER_BASE}/${userId}/packages`
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
    data: { package_ids: packageIds }
  })
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

/** 获取平台用户例外权限覆盖（兼容入口） */
export async function fetchGetUserActions(userId: string) {
  const res = await request.get<Api.SystemManage.UserActionPermissionResponse>({
    url: `${USER_BASE}/${userId}/actions`
  })
  const actions = (res?.actions || []).map((item: any) => ({
      action_id: item?.action_id || '',
      effect: item?.effect || 'allow',
      action: item?.action ? normalizePermissionAction(item.action) : undefined
    })) as Api.SystemManage.UserActionPermissionItem[]
  return {
    actions,
    available_action_ids: res?.available_action_ids || [],
    available_actions: (res?.available_actions || []).map(normalizePermissionAction),
    expanded_package_ids: res?.expanded_package_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      action_id: item?.action_id || '',
      package_ids: item?.package_ids || []
    })),
    has_package_config: Boolean(res?.has_package_config)
  }
}

export const fetchGetUserActionOverrides = fetchGetUserActions

/** 获取平台用户菜单裁剪 */
export async function fetchGetUserMenus(userId: string) {
  const res = await request.get<Api.SystemManage.UserMenuBoundaryResponse>({
    url: `${USER_BASE}/${userId}/menus`
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
    data: { menu_ids: menuIds }
  })
}

/** 设置平台用户例外权限覆盖（兼容入口） */
export function fetchSetUserActions(
  userId: string,
  actions: Array<{ action_id: string; effect: 'allow' | 'deny' }>
) {
  return request.put<void>({
    url: `${USER_BASE}/${userId}/actions`,
    data: { actions }
  })
}

export const fetchSetUserActionOverrides = fetchSetUserActions

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
      package_ids: res?.package_ids || [],
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
      package_ids: res?.package_ids || [],
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
    resource_code: params?.resourceCode,
    action_code: params?.actionCode,
    module_code: params?.moduleCode,
    context_type: params?.contextType,
    feature_kind: params?.featureKind,
    permissionKey: undefined,
    resourceCode: undefined,
    actionCode: undefined,
    moduleCode: undefined,
    contextType: undefined,
    featureKind: undefined
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
    context_type: params?.contextType,
    packageKey: undefined,
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
  return request.get<Api.SystemManage.FeaturePackageMenuResponse>({
    url: `${FEATURE_PACKAGE_BASE}/${id}/menus`
  })
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
    ...params,
    resource_code: params?.resourceCode,
    action_code: params?.actionCode,
    feature_kind: params?.featureKind,
    resourceCode: undefined,
    actionCode: undefined,
    featureKind: undefined
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
