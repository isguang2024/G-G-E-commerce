// Phase 4 slice 5: migrated to v5Client + openapi-fetch.
import {
  v5Client,
  unwrap,
  normalizeRole,
  normalizeFeaturePackage,
  normalizePermissionAction
} from './_shared'

// 获取角色列表
export async function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  const res = await unwrap(v5Client.GET('/roles', { params: { query: params as any } }))
  return {
    ...(res as any),
    records: ((res as any)?.records || []).map(normalizeRole)
  } as Api.SystemManage.RoleList
}

export async function fetchGetRoleOptions() {
  const res: any = await unwrap(v5Client.GET('/roles/options', { params: { query: {} as any } }))
  return {
    records: (res?.records || []).map(normalizeRole),
    total: Number(res?.total || 0)
  }
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
export async function fetchGetRole(id: string) {
  const res = await unwrap(v5Client.GET('/roles/{id}', { params: { path: { id } } }))
  return normalizeRole(res)
}

// 创建角色
export async function fetchCreateRole(data: Api.SystemManage.RoleCreateParams) {
  const res: any = await unwrap(v5Client.POST('/roles', { body: data as any }))
  return res as { roleId: string }
}

// 更新角色
export async function fetchUpdateRole(id: string, data: Api.SystemManage.RoleUpdateParams) {
  const { error } = await v5Client.PUT('/roles/{id}', {
    params: { path: { id } },
    body: data as any
  })
  if (error) throw error
}

// 删除角色
export async function fetchDeleteRole(id: string) {
  const { error } = await v5Client.DELETE('/roles/{id}', { params: { path: { id } } })
  if (error) throw error
}

/** 获取角色已分配的菜单 ID 列表（用于菜单权限配置） */
export async function fetchGetRoleMenus(roleId: string, appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/roles/{id}/menus', {
      params: { path: { id: roleId }, query: { app_key: appKey as any } }
    })
  )
  return {
    menu_ids: res?.menu_ids || [],
    available_menu_ids: res?.available_menu_ids || [],
    hidden_menu_ids: res?.hidden_menu_ids || [],
    expanded_package_ids: res?.expanded_package_ids || [],
    derived_sources: res?.derived_sources || []
  }
}

/** 获取角色功能包 */
export async function fetchGetRolePackages(roleId: string, appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/roles/{id}/packages', {
      params: { path: { id: roleId }, query: { app_key: appKey as any } }
    })
  )
  return {
    package_ids: res?.package_ids || [],
    packages: (res?.packages || []).map(normalizeFeaturePackage)
  }
}

/** 设置角色功能包 */
export async function fetchSetRolePackages(roleId: string, packageIds: string[], appKey?: string) {
  const { error } = await v5Client.PUT('/roles/{id}/packages', {
    params: { path: { id: roleId }, query: { app_key: appKey as any } },
    body: { package_ids: packageIds } as any
  })
  if (error) throw error
}

/** 设置角色菜单权限 */
export async function fetchSetRoleMenus(roleId: string, menuIds: string[], appKey?: string) {
  const { error } = await v5Client.PUT('/roles/{id}/menus', {
    params: { path: { id: roleId }, query: { app_key: appKey as any } },
    body: { menu_ids: menuIds } as any
  })
  if (error) throw error
}

/** 获取角色功能权限 */
export async function fetchGetRoleActions(roleId: string, appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/roles/{id}/actions', {
      params: { path: { id: roleId }, query: { app_key: appKey as any } }
    })
  )
  return {
    action_ids: res?.action_ids || [],
    available_action_ids: res?.available_action_ids || [],
    disabled_action_ids: res?.disabled_action_ids || [],
    actions: (res?.actions || []).map(normalizePermissionAction),
    expanded_package_ids: res?.expanded_package_ids || [],
    derived_sources: res?.derived_sources || []
  }
}

/** 设置角色功能权限 */
export async function fetchSetRoleActions(roleId: string, actionIds: string[], appKey?: string) {
  const { error } = await v5Client.PUT('/roles/{id}/actions', {
    params: { path: { id: roleId }, query: { app_key: appKey as any } },
    body: { action_ids: actionIds, ...(appKey ? { app_key: appKey } : {}) } as any
  })
  if (error) throw error
}

/** 获取角色数据权限 */
export async function fetchGetRoleDataPermissions(roleId: string) {
  const res: any = await unwrap(
    v5Client.GET('/roles/{id}/data-permissions', { params: { path: { id: roleId } } })
  )
  return {
    permissions: res?.permissions || [],
    resources: res?.resources || [],
    available_data_scopes: res?.available_data_scopes || res?.data_scopes || []
  } as {
    permissions: Array<{ resource_code: string; data_scope: string }>
    resources: Array<{ resource_code: string; resource_name: string }>
    available_data_scopes: Array<{ data_scope: string; label: string }>
  }
}

/** 设置角色数据权限 */
export async function fetchSetRoleDataPermissions(
  roleId: string,
  permissions: Array<{ resource_code: string; data_scope: string }>
) {
  const { error } = await v5Client.PUT('/roles/{id}/data-permissions', {
    params: { path: { id: roleId } },
    body: { permissions } as any
  })
  if (error) throw error
}
