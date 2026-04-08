import {
  request,
  ROLE_BASE,
  normalizeRole,
  normalizeFeaturePackage,
  normalizePermissionAction
} from './_shared'

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request
    .get<Api.SystemManage.RoleList>({
      url: ROLE_BASE,
      params
    })
    .then((res) => ({
      ...res,
      records: (res?.records || []).map(normalizeRole)
    }))
}

export function fetchGetRoleOptions() {
  return request
    .get<{ records: Api.SystemManage.RoleListItem[]; total: number }>({
      url: `${ROLE_BASE}/options`
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeRole),
      total: Number(res?.total || 0)
    }))
}

// 获取角色列表（简单列表，用于下拉等）
export function fetchGetRoleListSimple() {
  return fetchGetRoleOptions().then((res) => ({
    records: res?.records || [],
    total: res?.total || 0,
    current: 1,
    size: Math.max(res?.total || 0, 20)
  }))
}

// 获取角色详情
export function fetchGetRole(id: string) {
  return request
    .get<Api.SystemManage.RoleListItem>({
      url: `${ROLE_BASE}/${id}`
    })
    .then((res) => normalizeRole(res))
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
export function fetchGetRoleMenus(roleId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RoleMenuBoundaryResponse>({
      url: `${ROLE_BASE}/${roleId}/menus`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
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
export function fetchGetRolePackages(roleId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RoleFeaturePackageResponse>({
      url: `${ROLE_BASE}/${roleId}/packages`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

/** 设置角色功能包 */
export function fetchSetRolePackages(roleId: string, packageIds: string[], appKey?: string) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/packages`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { package_ids: packageIds }
  })
}

/** 设置角色菜单权限 */
export function fetchSetRoleMenus(roleId: string, menuIds: string[], appKey?: string) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/menus`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { menu_ids: menuIds }
  })
}

/** 获取角色功能权限 */
export function fetchGetRoleActions(roleId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RoleActionBoundaryResponse>({
      url: `${ROLE_BASE}/${roleId}/actions`,
      params: appKey ? { app_key: appKey } : undefined
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
export function fetchSetRoleActions(roleId: string, actionIds: string[], appKey?: string) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/actions`,
    data: {
      action_ids: actionIds,
      ...(appKey ? { app_key: appKey } : {})
    }
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
