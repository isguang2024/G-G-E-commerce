import { v5Client, unwrap } from '@/api/system-manage/_shared'

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
  if (key === 'collaboration_workspace.manage' || module === 'collaboration_workspace') {
    return 'personal'
  }
  if (
    key.startsWith('collaboration_workspace.member.') ||
    key.startsWith('collaboration_workspace.boundary.') ||
    key.startsWith('collaboration_workspace.message.') ||
    module === 'collaboration_workspace_member' ||
    module === 'collaboration_workspace_boundary' ||
    module === 'collaboration_workspace_message'
  ) {
    return 'collaboration'
  }
  if (
    key.startsWith('system.') ||
    key.startsWith('feature_package.') ||
    key.startsWith('message.') ||
    key.startsWith('personal.') ||
    module === 'role' ||
    module === 'user' ||
    module === 'menu' ||
    module === 'menu_backup' ||
    module === 'permission_action' ||
    module === 'permission_key' ||
    module === 'api_endpoint' ||
    module === 'feature_package' ||
    module === 'collaboration_workspace_member_admin'
  ) {
    return 'personal'
  }
  return 'common'
}

function normalizeCollaborationWorkspace(
  item: any
): Api.SystemManage.CollaborationWorkspaceListItem {
  const collaborationWorkspaceId =
    item?.collaboration_workspace_id ||
    item?.collaborationWorkspaceId ||
    item?.id ||
    item?.workspace_id ||
    item?.workspaceId ||
    ''
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
    collaborationWorkspaceId,
    workspaceId: item?.workspace_id || item?.workspaceId || '',
    workspaceType: item?.workspace_type || item?.workspaceType || 'collaboration',
    currentRoleCode: item?.current_role_code || item?.currentRoleCode || '',
    memberStatus: item?.member_status || item?.memberStatus || ''
  }
}

function normalizeAction(item: any): Api.SystemManage.PermissionActionItem {
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
    featureKind: featureGroup?.code || item?.feature_kind || item?.featureKind || 'business',
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

function normalizeFeaturePackage(item: any): Api.SystemManage.FeaturePackageItem {
  const packageKey = item?.package_key || item?.packageKey || ''
  const appKeysRaw = item?.app_keys || item?.appKeys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    packageKey,
    packageType: item?.package_type || item?.packageType || 'base',
    name: item?.name || '',
    description: item?.description || '',
    workspaceScope: item?.workspace_scope || item?.workspaceScope || 'all',
    appKey: item?.app_key || item?.appKey || '',
    appKeys,
    isBuiltin: Boolean(item?.is_builtin ?? item?.isBuiltin ?? false),
    actionCount: item?.action_count ?? item?.actionCount ?? 0,
    menuCount: item?.menu_count ?? item?.menuCount ?? 0,
    collaborationWorkspaceCount: item?.collaborationWorkspaceCount ?? 0,
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    createdAt: item?.created_at || item?.createdAt || '',
    updatedAt: item?.updated_at || item?.updatedAt || ''
  }
}

function normalizeRoleLabel(roleCode?: string) {
  return roleCode === 'collaboration_workspace_admin' ? '协作空间管理员' : '协作空间成员'
}

function normalizeCollaborationWorkspaceMember(
  item: any
): Api.SystemManage.CollaborationWorkspaceMemberItem {
  const roleCode = item?.role_code || item?.roleCode || ''
  return {
    id: item?.id || '',
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || '',
    workspaceId: item?.workspace_id || item?.workspaceId || '',
    workspaceType: item?.workspace_type || item?.workspaceType || 'collaboration',
    userId: item?.user_id || item?.userId || '',
    roleCode,
    role: normalizeRoleLabel(roleCode),
    memberType: item?.member_type || item?.memberType || '',
    status: item?.status || 'active',
    joinedAt: item?.joined_at || item?.joinedAt || '',
    userName: item?.user_name || item?.userName || '',
    nickName: item?.nick_name || item?.nickName || '',
    userEmail: item?.user_email || item?.userEmail || '',
    avatar: item?.avatar || ''
  }
}

export async function fetchGetCollaborationWorkspaceList(
  params: Api.SystemManage.CollaborationWorkspaceSearchParams
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces', { params: { query: params as any } })
  )
  return {
    ...res,
    records: (res?.records || []).map(normalizeCollaborationWorkspace)
  } as Api.SystemManage.CollaborationWorkspaceList
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

export async function fetchGetCollaborationWorkspace(id: string) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}', { params: { path: { id } } })
  )
  return normalizeCollaborationWorkspace(res)
}

export function fetchCreateCollaborationWorkspace(
  data: Api.SystemManage.CollaborationWorkspaceCreateParams
) {
  return unwrap(
    v5Client.POST('/collaboration-workspaces', { body: data as any })
  ) as unknown as Promise<{ id: string }>
}

export async function fetchUpdateCollaborationWorkspace(
  id: string,
  data: Api.SystemManage.CollaborationWorkspaceUpdateParams
) {
  const { error } = await v5Client.PUT('/collaboration-workspaces/{id}', {
    params: { path: { id } },
    body: data as any
  })
  if (error) throw error
}

export async function fetchDeleteCollaborationWorkspace(id: string) {
  const { error } = await v5Client.DELETE('/collaboration-workspaces/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

export async function fetchGetCollaborationWorkspaceMembers(
  collaborationWorkspaceId: string,
  params?: { user_id?: string; user_name?: string; role_code?: string }
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/members', {
      params: { path: { id: collaborationWorkspaceId }, query: (params || {}) as any }
    })
  )
  return (res || []).map(normalizeCollaborationWorkspaceMember)
}

export async function fetchAddCollaborationWorkspaceMember(
  collaborationWorkspaceId: string,
  data: { user_id: string; role_code?: string }
) {
  const { error } = await v5Client.POST('/collaboration-workspaces/{id}/members', {
    params: { path: { id: collaborationWorkspaceId } },
    body: {
      user_id: data.user_id,
      role_code: data.role_code || 'collaboration_workspace_member'
    } as any
  })
  if (error) throw error
}

export async function fetchRemoveCollaborationWorkspaceMember(
  collaborationWorkspaceId: string,
  userId: string
) {
  const { error } = await v5Client.DELETE('/collaboration-workspaces/{id}/members/{userId}', {
    params: { path: { id: collaborationWorkspaceId, userId } }
  })
  if (error) throw error
}

export async function fetchUpdateCollaborationWorkspaceMemberRole(
  collaborationWorkspaceId: string,
  userId: string,
  roleCode: string
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/{id}/members/{userId}/role',
    {
      params: { path: { id: collaborationWorkspaceId, userId } },
      body: { role_code: roleCode } as any
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspace() {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current', { params: { query: {} as any } })
  )
  return normalizeCollaborationWorkspace(res)
}

export async function fetchGetMyCollaborationWorkspaces() {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/mine', { params: { query: {} as any } })
  )
  return (res || []).map(normalizeCollaborationWorkspace)
}

export async function fetchGetMyCollaborationWorkspaceMembers() {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/members', { params: { query: {} as any } })
  )
  return (res || []).map(normalizeCollaborationWorkspaceMember)
}

export async function fetchAddMyCollaborationWorkspaceMember(data: {
  user_id: string
  role_code?: string
}) {
  const { error } = await v5Client.POST('/collaboration-workspaces/current/members', {
    body: {
      user_id: data.user_id,
      role_code: data.role_code || 'collaboration_workspace_member'
    } as any
  })
  if (error) throw error
}

export async function fetchRemoveMyCollaborationWorkspaceMember(userId: string) {
  const { error } = await v5Client.DELETE(
    '/collaboration-workspaces/current/members/{userId}',
    { params: { path: { userId } } }
  )
  if (error) throw error
}

export async function fetchUpdateMyCollaborationWorkspaceMemberRole(
  userId: string,
  roleCode: string
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/members/{userId}/role',
    {
      params: { path: { userId } },
      body: { role_code: roleCode } as any
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceMemberRoles(userId: string) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/members/{userId}/roles', {
      params: { path: { userId } }
    })
  )
  return {
    role_ids: res?.role_ids || [],
    roles: (res?.roles || []).map((item: any) => ({
      id: item?.id || '',
      code: item?.code || '',
      name: item?.name || ''
    })),
    global_role_ids: res?.global_role_ids || [],
    collaboration_workspace_role_ids: res?.collaboration_workspace_role_ids || [],
    bindingWorkspaceId: res?.binding_workspace_id || res?.bindingWorkspaceId || '',
    collaborationWorkspaceId:
      res?.collaboration_workspace_id ||
      res?.collaborationWorkspaceId ||
      res?.binding_workspace_id ||
      res?.bindingWorkspaceId ||
      '',
    bindingWorkspaceType:
      res?.binding_workspace_type || res?.bindingWorkspaceType || 'collaboration',
    memberType: res?.member_type || res?.memberType || ''
  } satisfies Api.SystemManage.CollaborationWorkspaceMemberRoleBindingResponse
}

export async function fetchSetMyCollaborationWorkspaceMemberRoles(
  userId: string,
  roleIds: string[]
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/members/{userId}/roles',
    {
      params: { path: { userId } },
      body: { role_ids: roleIds } as any
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceRoles() {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/roles', { params: { query: {} as any } })
  )
  return (res || []).map((item: any) => ({
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || '',
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? false),
    createTime: item?.create_time || item?.created_at || ''
  }))
}

export async function fetchGetCollaborationWorkspaceRoles(collaborationWorkspaceId: string) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/roles', {
      params: { path: { id: collaborationWorkspaceId } }
    })
  )
  return (res || []).map((item: any) => ({
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || '',
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? false),
    createTime: item?.create_time || item?.created_at || ''
  }))
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRoles(appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles', {
      params: { query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return (res || []).map((item: any) => ({
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || '',
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? false),
    createTime: item?.create_time || item?.created_at || ''
  }))
}

export function fetchCreateMyCollaborationWorkspaceRole(data: Api.SystemManage.RoleCreateParams) {
  return unwrap(
    v5Client.POST('/collaboration-workspaces/current/roles', { body: data as any })
  ) as unknown as Promise<{ roleId: string }>
}

export async function fetchUpdateMyCollaborationWorkspaceRole(
  roleId: string,
  data: Api.SystemManage.RoleUpdateParams
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}',
    {
      params: { path: { roleId } },
      body: data as any
    }
  )
  if (error) throw error
}

export async function fetchDeleteMyCollaborationWorkspaceRole(roleId: string) {
  const { error } = await v5Client.DELETE(
    '/collaboration-workspaces/current/boundary/roles/{roleId}',
    { params: { path: { roleId } } }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRoleMenus(
  roleId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles/{roleId}/menus', {
      params: { path: { roleId }, query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    menu_ids: res?.menu_ids || [],
    available_menu_ids: res?.available_menu_ids || [],
    hidden_menu_ids: res?.hidden_menu_ids || [],
    expanded_package_ids: res?.expanded_package_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      menu_id: item?.menu_id || '',
      package_ids: item?.package_ids || []
    }))
  }
}

export async function fetchSetMyCollaborationWorkspaceBoundaryRoleMenus(
  roleId: string,
  menuIds: string[],
  appKey?: string
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}/menus',
    {
      params: { path: { roleId }, query: (appKey ? { app_key: appKey } : {}) as any },
      body: { menu_ids: menuIds } as any
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRoleActions(
  roleId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles/{roleId}/actions', {
      params: { path: { roleId }, query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    action_ids: res?.action_ids || [],
    available_action_ids: res?.available_action_ids || [],
    disabled_action_ids: res?.disabled_action_ids || [],
    actions: (res?.actions || []).map(normalizeAction),
    expanded_package_ids: res?.expanded_package_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      action_id: item?.action_id || '',
      package_ids: item?.package_ids || []
    }))
  }
}

export async function fetchSetMyCollaborationWorkspaceBoundaryRoleActions(
  roleId: string,
  actionIds: string[],
  appKey?: string
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}/actions',
    {
      params: { path: { roleId }, query: (appKey ? { app_key: appKey } : {}) as any },
      body: { action_ids: actionIds } as any
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRolePackages(
  roleId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles/{roleId}/packages', {
      params: { path: { roleId }, query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    package_ids: res?.package_ids || [],
    packages: (res?.packages || []).map(normalizeFeaturePackage),
    inherited: Boolean(res?.inherited)
  }
}

export async function fetchSetMyCollaborationWorkspaceBoundaryRolePackages(
  roleId: string,
  packageIds: string[],
  appKey?: string
) {
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}/packages',
    {
      params: { path: { roleId }, query: (appKey ? { app_key: appKey } : {}) as any },
      body: { package_ids: packageIds } as any
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryPackages(appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/packages', {
      params: { query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    package_ids: res?.package_ids || [],
    packages: (res?.packages || []).map(normalizeFeaturePackage)
  }
}

export async function fetchGetCollaborationWorkspaceActions(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/actions', {
      params: {
        path: { id: collaborationWorkspaceId },
        query: (appKey ? { app_key: appKey } : {}) as any
      }
    })
  )
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export async function fetchGetCollaborationWorkspaceMenus(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/menus', {
      params: {
        path: { id: collaborationWorkspaceId },
        query: (appKey ? { app_key: appKey } : {}) as any
      }
    })
  )
  return { menu_ids: res?.menu_ids || [] }
}

export async function fetchGetCollaborationWorkspaceActionOrigins(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/action-origins', {
      params: {
        path: { id: collaborationWorkspaceId },
        query: (appKey ? { app_key: appKey } : {}) as any
      }
    })
  )
  return {
    derived_action_ids: res?.derived_action_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      action_id: item?.action_id || '',
      package_ids: item?.package_ids || []
    })),
    blocked_action_ids: res?.blocked_action_ids || []
  }
}

export async function fetchGetCollaborationWorkspaceMenuOrigins(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/menu-origins', {
      params: {
        path: { id: collaborationWorkspaceId },
        query: (appKey ? { app_key: appKey } : {}) as any
      }
    })
  )
  return {
    derived_menu_ids: res?.derived_menu_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      menu_id: item?.menu_id || '',
      package_ids: item?.package_ids || []
    })),
    blocked_menu_ids: res?.blocked_menu_ids || []
  }
}

export async function fetchSetCollaborationWorkspaceActions(
  collaborationWorkspaceId: string,
  actionIds: string[],
  appKey?: string
) {
  const { error } = await v5Client.PUT('/collaboration-workspaces/{id}/actions', {
    params: {
      path: { id: collaborationWorkspaceId },
      query: (appKey ? { app_key: appKey } : {}) as any
    },
    body: { action_ids: actionIds } as any
  })
  if (error) throw error
}

export async function fetchSetCollaborationWorkspaceMenus(
  collaborationWorkspaceId: string,
  menuIds: string[],
  appKey?: string
) {
  const { error } = await v5Client.PUT('/collaboration-workspaces/{id}/menus', {
    params: {
      path: { id: collaborationWorkspaceId },
      query: (appKey ? { app_key: appKey } : {}) as any
    },
    body: { menu_ids: menuIds } as any
  })
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceActions() {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/actions', { params: { query: {} as any } })
  )
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export async function fetchGetMyCollaborationWorkspaceActionOrigins() {
  const res: any = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/action-origins', {
      params: { query: {} as any }
    })
  )
  return {
    derived_action_ids: res?.derived_action_ids || [],
    derived_sources: (res?.derived_sources || []).map((item: any) => ({
      action_id: item?.action_id || '',
      package_ids: item?.package_ids || []
    })),
    blocked_action_ids: res?.blocked_action_ids || []
  }
}
