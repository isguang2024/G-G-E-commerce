import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

const USER_BASE = '/api/v1/users'
const ROLE_BASE = '/api/v1/roles'
const ACTION_PERMISSION_BASE = '/api/v1/permission-actions'
const SCOPE_BASE = '/api/v1/scopes'
const TENANT_BASE = '/api/v1/tenants'
const SYSTEM_BASE = '/api/v1/system'
const API_ENDPOINT_BASE = '/api/v1/api-endpoints'

function normalizePermissionAction(item: any): Api.SystemManage.PermissionActionItem {
  return {
    id: item?.id || '',
    resourceCode: item?.resource_code || item?.resourceCode || '',
    actionCode: item?.action_code || item?.actionCode || '',
    moduleCode: item?.module_code || item?.moduleCode || item?.category || item?.resource_code || item?.resourceCode || '',
    permissionKey:
      item?.permission_key ||
      item?.permissionKey ||
      `${item?.resource_code || item?.resourceCode || ''}:${item?.action_code || item?.actionCode || ''}`,
    category: item?.category || '',
    source: item?.source || 'business',
    featureKind: item?.feature_kind || item?.featureKind || 'system',
    name: item?.name || '',
    description: item?.description || '',
    scopeId: item?.scope_id || item?.scopeId || '',
    scopeCode: item?.scope_code || item?.scopeCode || item?.scope || '',
    scopeName: item?.scope_name || item?.scopeName || '',
    scopeContextKind: item?.scope_context_kind || item?.scopeContextKind || '',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    scope: item?.scope || item?.scope_code || item?.scopeCode || '',
    requiresTenantContext: Boolean(
      item?.requires_tenant_context ?? item?.requiresTenantContext ?? false
    ),
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

function normalizeApiEndpoint(item: any): Api.SystemManage.APIEndpointItem {
  return {
    id: item?.id || '',
    method: item?.method || '',
    path: item?.path || '',
    module: item?.module || '',
    featureKind: item?.feature_kind || item?.featureKind || 'system',
    handler: item?.handler || '',
    summary: item?.summary || '',
    resourceCode: item?.resource_code || item?.resourceCode || '',
    actionCode: item?.action_code || item?.actionCode || '',
    scopeId: item?.scope_id || item?.scopeId || '',
    scopeCode: item?.scope_code || item?.scopeCode || '',
    scopeName: item?.scope_name || item?.scopeName || '',
    scopeContextKind: item?.scope_context_kind || item?.scopeContextKind || '',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    requiresTenantContext: Boolean(
      item?.requires_tenant_context ?? item?.requiresTenantContext ?? false
    ),
    status: item?.status || 'normal',
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

function normalizeScope(item: any): Api.SystemManage.ScopeListItem {
  return {
    scopeId: item?.scope_id || item?.scopeId || '',
    scopeCode: item?.scope_code || item?.scopeCode || '',
    scopeName: item?.scope_name || item?.scopeName || '',
    description: item?.description || '',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    contextKind: item?.context_kind || item?.contextKind || 'global',
    dataPermissionCode: item?.data_permission_code || item?.dataPermissionCode || '',
    dataPermissionName: item?.data_permission_name || item?.dataPermissionName || '',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    createTime: item?.create_time || item?.createTime || ''
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

/** 获取用户平台级功能权限 */
export async function fetchGetUserActions(userId: string) {
  const res = await request.get<{ actions: any[] }>({
    url: `${USER_BASE}/${userId}/actions`
  })
  return (res?.actions || []).map((item: any) => ({
    actionId: item?.action_id || item?.actionId || '',
    effect: item?.effect || 'allow',
    action: item?.action ? normalizePermissionAction(item.action) : undefined
  })) as Api.SystemManage.UserActionPermissionItem[]
}

/** 设置用户平台级功能权限 */
export function fetchSetUserActions(
  userId: string,
  actions: Array<{ action_id: string; effect: 'allow' | 'deny' }>
) {
  return request.put<void>({
    url: `${USER_BASE}/${userId}/actions`,
    data: { actions }
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
  return request.get<{ menu_ids: string[] }>({
    url: `${ROLE_BASE}/${roleId}/menus`
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
  return request.get<{ actions: Array<{ action_id: string; effect: 'allow' | 'deny' }> }>({
    url: `${ROLE_BASE}/${roleId}/actions`
  })
}

/** 设置角色功能权限 */
export function fetchSetRoleActions(
  roleId: string,
  actions: Array<{ action_id: string; effect: 'allow' | 'deny' }>
) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/actions`,
    data: { actions }
  })
}

/** 获取角色数据权限 */
export async function fetchGetRoleDataPermissions(roleId: string) {
  return request.get<{
    permissions: Array<{ resource_code: string; scope_code: string }>
    resources: Array<{ resource_code: string; resource_name: string }>
    available_scopes: Array<{ scope_code: string; scope_name: string }>
  }>({
    url: `${ROLE_BASE}/${roleId}/data-permissions`
  })
}

/** 设置角色数据权限 */
export function fetchSetRoleDataPermissions(
  roleId: string,
  permissions: Array<{ resource_code: string; scope_code: string }>
) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/data-permissions`,
    data: { permissions }
  })
}

/** 获取功能权限列表 */
export function fetchGetPermissionActionList(params: Api.SystemManage.PermissionActionSearchParams) {
  const normalizedParams = {
    ...params,
    resource_code: params?.resourceCode,
    action_code: params?.actionCode,
    module_code: params?.moduleCode,
    scope_id: params?.scopeId,
    scope_code: params?.scopeCode,
    feature_kind: params?.featureKind,
    resourceCode: undefined,
    actionCode: undefined,
    moduleCode: undefined,
    scopeId: undefined,
    scopeCode: undefined,
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

/** 获取 API 注册表 */
export function fetchGetApiEndpointList(params: Api.SystemManage.APIEndpointSearchParams) {
  const normalizedParams = {
    ...params,
    resource_code: params?.resourceCode,
    action_code: params?.actionCode,
    scope_code: params?.scopeCode,
    feature_kind: params?.featureKind,
    resourceCode: undefined,
    actionCode: undefined,
    scopeCode: undefined,
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

// ========== 作用域管理 ==========
/** 获取作用域列表 */
export function fetchGetScopeList(params: Api.SystemManage.ScopeSearchParams) {
  return request
    .get<Api.SystemManage.ScopeList>({
      url: SCOPE_BASE,
      params
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeScope)
    }))
}

/** 获取所有作用域（用于下拉选择） */
export function fetchGetAllScopes() {
  return request
    .get<Api.SystemManage.ScopeListItem[] | { records?: Api.SystemManage.ScopeListItem[] }>({
      url: `${SCOPE_BASE}/all`
    })
    .then((res) => {
      if (Array.isArray(res)) return res.map(normalizeScope)
      return (res?.records || []).map(normalizeScope)
    })
}

/** 获取作用域详情 */
export function fetchGetScope(id: string) {
  return request
    .get<Api.SystemManage.ScopeListItem>({
      url: `${SCOPE_BASE}/${id}`
    })
    .then((res) => normalizeScope(res))
}

/** 创建作用域 */
export function fetchCreateScope(data: Api.SystemManage.ScopeCreateParams) {
  return request.post<{ scopeId: string }>({
    url: SCOPE_BASE,
    data
  })
}

/** 更新作用域 */
export function fetchUpdateScope(id: string, data: Api.SystemManage.ScopeUpdateParams) {
  return request.put<void>({
    url: `${SCOPE_BASE}/${id}`,
    data
  })
}

/** 删除作用域 */
export function fetchDeleteScope(id: string) {
  return request.del<void>({
    url: `${SCOPE_BASE}/${id}`,
    showErrorMessage: false // 禁用自动错误消息显示，由业务代码处理
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
