import {
  v5Client,
  unwrap,
  normalizeUserSummary,
  normalizeFeaturePackage,
  normalizeUserPermissionDiagnosisResponse,
  normalizeUserPermissionMenuTree,
  normalizeUserCollaborationWorkspaceItem,
  type V5Query,
  type V5RequestBody
} from './_shared'
import type { components } from '@/api/v5/schema'

type V5UserMenusResponse = components['schemas']['UserMenusResponse']

// 获取用户列表
export async function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  const query: V5Query<'/users', 'get'> = {}
  if (params?.current != null) query.current = params.current
  if (params?.size != null) query.size = params.size
  if (params?.id) query.id = params.id
  if (params?.userName) query.user_name = params.userName
  if (params?.userEmail) query.user_email = params.userEmail
  if (params?.userPhone) query.user_phone = params.userPhone
  if (params?.status) query.status = params.status
  if ('roleId' in params && params.roleId) query.role_id = params.roleId
  if (params?.registerSource) query.register_source = params.registerSource
  if (params?.invitedBy) query.invited_by = params.invitedBy
  try {
    const data = await unwrap(v5Client.GET('/users', { params: { query } }))
    return {
      records: (data.records || []).map(normalizeUserSummary),
      total: data.total || 0,
      current: data.current,
      size: data.size
    } as Api.SystemManage.UserList
  } catch {
    return { records: [], total: 0, current: params?.current ?? 1, size: params?.size ?? 20 }
  }
}

/** 获取用户个人空间功能包 */
export async function fetchGetUserPackages(userId: string, appKey?: string) {
  const query: V5Query<'/users/{id}/packages', 'get'> = appKey ? { app_key: appKey } : {}
  const res = await unwrap(
    v5Client.GET('/users/{id}/packages', {
      params: { path: { id: userId }, query }
    })
  )
  return {
    package_ids: res?.package_ids || [],
    packages: (res?.packages || []).map(normalizeFeaturePackage)
  }
}

/** 设置用户个人空间功能包 */
export async function fetchSetUserPackages(
  userId: string,
  packageIds: string[]
) {
  const body: V5RequestBody<'/users/{id}/packages', 'put'> = { ids: packageIds }
  const { error } = await v5Client.PUT('/users/{id}/packages', {
    params: { path: { id: userId } },
    body
  })
  if (error) throw error
}

// 获取用户详情（Phase 4 slice 5: v5Client）
export async function fetchGetUser(id: string) {
  try {
    const data = await unwrap(v5Client.GET('/users/{id}', { params: { path: { id } } }))
    return normalizeUserSummary(data)
  } catch {
    return normalizeUserSummary({})
  }
}

// 创建用户（Phase 4 slice 5: v5Client）
export async function fetchCreateUser(data: Api.SystemManage.UserCreateParams) {
  const out = await unwrap(
    v5Client.POST('/users', {
      body: {
        username: data.username,
        password: data.password,
        email: data.email,
        nickname: data.nickname,
        phone: data.phone,
        system_remark: data.systemRemark,
        status: data.status,
        role_ids: data.roleIds
      }
    })
  )
  return { id: out.id }
}

// 更新用户（Phase 4 slice 5: v5Client）
export async function fetchUpdateUser(id: string, data: Api.SystemManage.UserUpdateParams) {
  const { error } = await v5Client.PUT('/users/{id}', {
    params: { path: { id } },
    body: {
      email: data.email,
      nickname: data.nickname,
      phone: data.phone,
      system_remark: data.systemRemark,
      status: data.status,
      role_ids: data.roleIds
    }
  })
  if (error) throw error
}

// 删除用户（Phase 4 slice 5: v5Client）
export async function fetchDeleteUser(id: string) {
  const { error } = await v5Client.DELETE('/users/{id}', { params: { path: { id } } })
  if (error) throw error
}

// 分配个人空间角色（Phase 4 slice 5: v5Client）
export async function fetchAssignUserRoles(id: string, roleIds: string[]) {
  const { error } = await v5Client.POST('/users/{id}/roles', {
    params: { path: { id } },
    body: { role_ids: roleIds }
  })
  if (error) throw error
}

/** 获取用户个人空间菜单裁剪 */
export async function fetchGetUserMenus(userId: string, appKey?: string) {
  const query: V5Query<'/users/{id}/menus', 'get'> = appKey ? { app_key: appKey } : {}
  const res: V5UserMenusResponse = await unwrap(
    v5Client.GET('/users/{id}/menus', {
      params: { path: { id: userId }, query }
    })
  )
  return {
    menu_ids: res.menu_ids || [],
    available_menu_ids: res.available_menu_ids || [],
    hidden_menu_ids: res.hidden_menu_ids || [],
    expanded_package_ids: res.expanded_package_ids || [],
    derived_sources: (res.derived_sources || []).map((item) => ({
      menu_id: item.menu_id || '',
      package_ids: item.package_ids || []
    })),
    has_package_config: Boolean(res.has_package_config)
  }
}

/** 设置用户个人空间菜单裁剪 */
export async function fetchSetUserMenus(userId: string, menuIds: string[], appKey?: string) {
  void appKey
  const body: V5RequestBody<'/users/{id}/menus', 'put'> = {
    menu_ids: menuIds,
    available_menu_ids: [],
    hidden_menu_ids: [],
    expanded_package_ids: [],
    derived_sources: [],
    has_package_config: false
  }
  const { error } = await v5Client.PUT('/users/{id}/menus', {
    params: { path: { id: userId } },
    body
  })
  if (error) throw error
}

export async function fetchGetUserCollaborationWorkspaces(userId: string) {
  const res = await unwrap(
    v5Client.GET('/users/{id}/collaborations', {
      params: { path: { id: userId } }
    })
  )
  return (res.records || []).map((item) => normalizeUserCollaborationWorkspaceItem(item))
}

/** 获取用户权限诊断 */
export async function fetchGetUserPermissionDiagnosis(
  userId: string,
  params?: Api.SystemManage.UserPermissionDiagnosisParams
) {
  const query: V5Query<'/users/{id}/permission-diagnosis', 'get'> = {
    workspace_id: params?.collaborationWorkspaceId,
    permission_key: params?.permissionKey
  }
  const res = await unwrap(
    v5Client.GET('/users/{id}/permission-diagnosis', {
      params: { path: { id: userId }, query }
    })
  )
  return normalizeUserPermissionDiagnosisResponse(res)
}

/** 刷新用户权限快照 */
export async function fetchRefreshUserPermissionSnapshot(
  userId: string,
  collaborationWorkspaceId?: string
) {
  void collaborationWorkspaceId
  return unwrap(v5Client.POST('/users/{id}/permission-refresh', {
    params: { path: { id: userId } }
  }))
}

/** 获取用户当前上下文可见菜单 */
export async function fetchGetUserPermissionMenus(
  userId: string,
  collaborationWorkspaceId?: string,
  appKey?: string
) {
  const query: V5Query<'/users/{id}/permissions', 'get'> = {
    ...(appKey ? { app_key: appKey } : {})
  }
  const res = await unwrap(
    v5Client.GET('/users/{id}/permissions', {
      params: { path: { id: userId }, query }
    })
  )
  void collaborationWorkspaceId
  return (res.menu_tree || []).map((item) => normalizeUserPermissionMenuTree(item))
}
