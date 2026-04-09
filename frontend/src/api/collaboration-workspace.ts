import {
  v5Client,
  unwrap,
  toV5Body,
  toV5Record,
  type V5Query,
  type V5RequestBody
} from '@/api/system-manage/_shared'

type V5PermissionGroupLike = {
  id?: string
  group_type?: string
  code?: string
  name?: string
  name_en?: string
  description?: string
  status?: string
  sort_order?: number
  is_builtin?: boolean
}

type V5CollaborationWorkspaceLike = {
  id?: string
  collaboration_workspace_id?: string
  name?: string
  remark?: string
  logo_url?: string
  plan?: string
  max_members?: number
  status?: string
  created_at?: string
  updated_at?: string
  admin_users?: Array<{ user_id: string; user_name?: string; nick_name?: string }>
  admin_user_ids?: string[]
  workspace_id?: string
  workspace_type?: string
  current_role_code?: string
  member_status?: string
}

type V5PermissionActionLike = {
  id?: string
  permission_key?: string
  module_code?: string
  module_group_id?: string
  feature_group_id?: string
  module_group?: V5PermissionGroupLike
  feature_group?: V5PermissionGroupLike
  context_type?: string
  feature_kind?: string
  name?: string
  description?: string | null
  status?: string | null
  data_permission_code?: string
  data_permission_name?: string
  sort_order?: number
  is_builtin?: boolean
  created_at?: string
  updated_at?: string
}

type V5FeaturePackageLike = {
  id?: string
  package_key?: string
  package_type?: string | null
  name?: string
  description?: string | null
  workspace_scope?: string
  app_key?: string
  app_keys?: string[]
  is_builtin?: boolean
  action_count?: number
  menu_count?: number
  collaborationWorkspaceCount?: number
  status?: string
  sort_order?: number
  created_at?: string
  updated_at?: string
}

type V5CollaborationWorkspaceMemberLike = {
  id?: string
  collaboration_workspace_id?: string
  workspace_id?: string
  workspace_type?: string
  user_id?: string
  role_code?: string
  member_type?: string
  status?: string
  joined_at?: string
  user_name?: string
  nick_name?: string
  user_email?: string
  avatar?: string
}

type V5RoleSummaryLike = {
  id?: string
  code?: string
  name?: string
  description?: string
  status?: string
  is_system?: boolean
  collaboration_workspace_id?: string | null
  is_global?: boolean
  created_at?: string
}

type V5MenuDerivedSourceLike = {
  menu_id?: string
  package_ids?: string[]
}

type V5ActionDerivedSourceLike = {
  action_id?: string
  package_ids?: string[]
}

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
  item: V5CollaborationWorkspaceLike | undefined
): Api.SystemManage.CollaborationWorkspaceListItem {
  // V5：协作空间 ID 单一语义，CW 接口仅认 collaboration_workspace_id（spec 字段名）。
  // 旧 axios 返回 id / workspace_id 的兜底已废弃。
  const collaborationWorkspaceId = item?.collaboration_workspace_id || ''
  return {
    id: item?.id || '',
    name: item?.name || '',
    remark: item?.remark || '',
    logoUrl: item?.logo_url || '',
    plan: item?.plan || 'free',
    maxMembers: item?.max_members ?? 0,
    status: item?.status || 'active',
    createTime: item?.created_at || '',
    updateTime: item?.updated_at || '',
    adminUsers: item?.admin_users || [],
    adminUserIds: item?.admin_user_ids || [],
    collaborationWorkspaceId,
    workspaceId: item?.workspace_id || '',
    workspaceType: item?.workspace_type || 'collaboration',
    currentRoleCode: item?.current_role_code || '',
    memberStatus: item?.member_status || ''
  }
}

function normalizeAction(item: V5PermissionActionLike | undefined): Api.SystemManage.PermissionActionItem {
  const permissionKey = normalizePermissionKey(item?.permission_key)
  const legacy = derivePermissionSegments(permissionKey)
  const moduleCode = item?.module_code || legacy.resourceCode || ''
  const normalizeGroup = (
    value: V5PermissionGroupLike | undefined
  ): Api.SystemManage.PermissionGroupItem | undefined =>
    value
      ? {
          id: value?.id || '',
          groupType: value?.group_type || '',
          code: value?.code || '',
          name: value?.name || '',
          nameEn: value?.name_en || '',
          description: value?.description || '',
          status: value?.status || 'normal',
          sortOrder: value?.sort_order ?? 0,
          isBuiltin: Boolean(value?.is_builtin ?? false)
        }
      : undefined
  const moduleGroup = normalizeGroup(item?.module_group)
  const featureGroup = normalizeGroup(item?.feature_group)
  return {
    id: item?.id || '',
    resourceCode: legacy.resourceCode,
    actionCode: legacy.actionCode,
    moduleCode: moduleGroup?.code || moduleCode,
    moduleGroupId: item?.module_group_id || moduleGroup?.id || '',
    featureGroupId: item?.feature_group_id || featureGroup?.id || '',
    moduleGroup,
    featureGroup,
    contextType: item?.context_type || deriveContextType(permissionKey, moduleCode),
    permissionKey,
    featureKind: featureGroup?.code || item?.feature_kind || 'business',
    name: item?.name || '',
    description: item?.description || '',
    dataPermissionCode: item?.data_permission_code || '',
    dataPermissionName: item?.data_permission_name || '',
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? 0,
    isBuiltin: Boolean(item?.is_builtin ?? false),
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

function normalizeFeaturePackage(
  item: V5FeaturePackageLike | undefined
): Api.SystemManage.FeaturePackageItem {
  const packageKey = item?.package_key || ''
  const appKeysRaw = item?.app_keys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    packageKey,
    packageType: item?.package_type || 'base',
    name: item?.name || '',
    description: item?.description || '',
    workspaceScope: item?.workspace_scope || 'all',
    appKey: item?.app_key || '',
    appKeys,
    isBuiltin: Boolean(item?.is_builtin ?? false),
    actionCount: item?.action_count ?? 0,
    menuCount: item?.menu_count ?? 0,
    collaborationWorkspaceCount: item?.collaborationWorkspaceCount ?? 0,
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? 0,
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

function normalizeRoleLabel(roleCode?: string) {
  return roleCode === 'collaboration_workspace_admin' ? '协作空间管理员' : '协作空间成员'
}

function normalizeRoleSummary(item: V5RoleSummaryLike | undefined): Api.SystemManage.RoleListItem {
  return {
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? false),
    collaborationWorkspaceId: item?.collaboration_workspace_id || '',
    isGlobal: Boolean(item?.is_global ?? false),
    createTime: item?.created_at || ''
  }
}

function normalizeMenuDerivedSource(
  item: V5MenuDerivedSourceLike | undefined
): { menu_id: string; package_ids: string[] } {
  return {
    menu_id: item?.menu_id || '',
    package_ids: Array.isArray(item?.package_ids) ? item.package_ids : []
  }
}

function normalizeActionDerivedSource(
  item: V5ActionDerivedSourceLike | undefined
): { action_id: string; package_ids: string[] } {
  return {
    action_id: item?.action_id || '',
    package_ids: Array.isArray(item?.package_ids) ? item.package_ids : []
  }
}

function normalizeCollaborationWorkspaceMember(
  item: V5CollaborationWorkspaceMemberLike | undefined
): Api.SystemManage.CollaborationWorkspaceMemberItem {
  const roleCode = item?.role_code || ''
  return {
    id: item?.id || '',
    collaborationWorkspaceId: item?.collaboration_workspace_id || '',
    workspaceId: item?.workspace_id || '',
    workspaceType: item?.workspace_type || 'collaboration',
    userId: item?.user_id || '',
    roleCode,
    role: normalizeRoleLabel(roleCode),
    memberType: item?.member_type || '',
    status: item?.status || 'active',
    joinedAt: item?.joined_at || '',
    userName: item?.user_name || '',
    nickName: item?.nick_name || '',
    userEmail: item?.user_email || '',
    avatar: item?.avatar || ''
  }
}

export async function fetchGetCollaborationWorkspaceList(
  params: Api.SystemManage.CollaborationWorkspaceSearchParams
) {
  const query: V5Query<'/collaboration-workspaces', 'get'> = params
  const res = await unwrap(v5Client.GET('/collaboration-workspaces', { params: { query } }))
  return {
    ...res,
    records: (res.records || []).map(normalizeCollaborationWorkspace)
  } as Api.SystemManage.CollaborationWorkspaceList
}

export async function fetchGetCollaborationWorkspaceOptions(
  params?: Partial<Api.SystemManage.CollaborationWorkspaceSearchParams>
) {
  void params
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/options'))
  return {
    records: (res.records || []).map(normalizeCollaborationWorkspace),
    total: res.total || 0
  }
}

export async function fetchGetCollaborationWorkspace(id: string) {
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}', { params: { path: { id } } })
  )
  return normalizeCollaborationWorkspace(res)
}

export async function fetchCreateCollaborationWorkspace(
  data: Api.SystemManage.CollaborationWorkspaceCreateParams
) {
  const body: V5RequestBody<'/collaboration-workspaces', 'post'> = {
    name: data.name,
    remark: data.remark,
    logo_url: data.logo_url,
    plan: data.plan,
    max_members: data.max_members,
    status: data.status,
    admin_user_ids: data.admin_user_ids
  }
  const res = await unwrap(v5Client.POST('/collaboration-workspaces', { body }))
  return { id: res.id }
}

export async function fetchUpdateCollaborationWorkspace(
  id: string,
  data: Api.SystemManage.CollaborationWorkspaceUpdateParams
) {
  const body: V5RequestBody<'/collaboration-workspaces/{id}', 'put'> = {
    name: data.name || '',
    remark: data.remark,
    logo_url: data.logo_url,
    plan: data.plan,
    max_members: data.max_members,
    status: data.status,
    admin_user_ids: data.admin_user_ids
  }
  const { error } = await v5Client.PUT('/collaboration-workspaces/{id}', {
    params: { path: { id } },
    body
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
  void params
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/members', {
      params: { path: { id: collaborationWorkspaceId } }
    })
  )
  return (res.records || []).map(normalizeCollaborationWorkspaceMember)
}

export async function fetchAddCollaborationWorkspaceMember(
  collaborationWorkspaceId: string,
  data: { user_id: string; role_code?: string }
) {
  const body: V5RequestBody<'/collaboration-workspaces/{id}/members', 'post'> = {
    user_id: data.user_id,
    role_code: data.role_code || 'collaboration_workspace_member'
  }
  const { error } = await v5Client.POST('/collaboration-workspaces/{id}/members', {
    params: { path: { id: collaborationWorkspaceId } },
    body
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
  const body: V5RequestBody<'/collaboration-workspaces/{id}/members/{userId}/role', 'put'> = {
    role_code: roleCode
  }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/{id}/members/{userId}/role',
    {
      params: { path: { id: collaborationWorkspaceId, userId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspace() {
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current'))
  return normalizeCollaborationWorkspace(res)
}

export async function fetchGetMyCollaborationWorkspaces() {
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/mine'))
  return (res.records || []).map(normalizeCollaborationWorkspace)
}

export async function fetchGetMyCollaborationWorkspaceMembers() {
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current/members'))
  return (res.records || []).map(normalizeCollaborationWorkspaceMember)
}

export async function fetchAddMyCollaborationWorkspaceMember(data: {
  user_id: string
  role_code?: string
}) {
  const body: V5RequestBody<'/collaboration-workspaces/current/members', 'post'> = {
    user_id: data.user_id,
    role_code: data.role_code || 'collaboration_workspace_member'
  }
  const { error } = await v5Client.POST('/collaboration-workspaces/current/members', {
    body
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
  const body: V5RequestBody<'/collaboration-workspaces/current/members/{userId}/role', 'put'> = {
    role_code: roleCode
  }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/members/{userId}/role',
    {
      params: { path: { userId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceMemberRoles(userId: string) {
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/members/{userId}/roles', {
      params: { path: { userId } }
    })
  )
  return {
    role_ids: Array.isArray(res.role_ids) ? res.role_ids : [],
    roles: (Array.isArray(res.roles) ? res.roles : []).map((item) => ({
      id: item?.id || '',
      code: item?.code || '',
      name: item?.name || ''
    })),
    global_role_ids: [],
    collaboration_workspace_role_ids: [],
    bindingWorkspaceId: '',
    collaborationWorkspaceId: '',
    bindingWorkspaceType: 'collaboration',
    memberType: ''
  } satisfies Api.SystemManage.CollaborationWorkspaceMemberRoleBindingResponse
}

export async function fetchSetMyCollaborationWorkspaceMemberRoles(
  userId: string,
  roleIds: string[]
) {
  const body: V5RequestBody<'/collaboration-workspaces/current/members/{userId}/roles', 'put'> = {
    ids: roleIds
  }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/members/{userId}/roles',
    {
      params: { path: { userId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceRoles() {
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current/roles'))
  return (res.records || []).map(normalizeRoleSummary)
}

export async function fetchGetCollaborationWorkspaceRoles(collaborationWorkspaceId: string) {
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/roles', {
      params: { path: { id: collaborationWorkspaceId } }
    })
  )
  return (res.records || []).map(normalizeRoleSummary)
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRoles(appKey?: string) {
  void appKey
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current/boundary/roles'))
  return (res.records || []).map(normalizeRoleSummary)
}

export async function fetchCreateMyCollaborationWorkspaceRole(
  data: Api.SystemManage.RoleCreateParams
) {
  const body: V5RequestBody<'/collaboration-workspaces/current/roles', 'post'> = {
    code: data.code,
    name: data.name,
    description: data.description,
    sort_order: data.sort_order,
    priority: data.priority,
    status: data.status
  }
  return unwrap(v5Client.POST('/collaboration-workspaces/current/roles', { body }))
}

export async function fetchUpdateMyCollaborationWorkspaceRole(
  roleId: string,
  data: Api.SystemManage.RoleUpdateParams
) {
  const body: V5RequestBody<'/collaboration-workspaces/current/boundary/roles/{roleId}', 'put'> =
    {
      code: data.code || '',
      name: data.name || '',
      description: data.description,
      sort_order: data.sort_order,
      priority: data.priority,
      status: data.status
    }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}',
    {
      params: { path: { roleId } },
      body
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
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles/{roleId}/menus', {
      params: { path: { roleId } }
    })
  )
  return {
    menu_ids: Array.isArray(res.menu_ids) ? res.menu_ids : [],
    available_menu_ids: Array.isArray(res.available_menu_ids) ? res.available_menu_ids : [],
    hidden_menu_ids: Array.isArray(res.hidden_menu_ids) ? res.hidden_menu_ids : [],
    expanded_package_ids: Array.isArray(res.expanded_package_ids) ? res.expanded_package_ids : [],
    derived_sources: (Array.isArray(res.derived_sources) ? res.derived_sources : []).map(
      normalizeMenuDerivedSource
    )
  }
}

export async function fetchSetMyCollaborationWorkspaceBoundaryRoleMenus(
  roleId: string,
  menuIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration-workspaces/current/boundary/roles/{roleId}/menus', 'put'> =
    { ids: menuIds }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}/menus',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRoleActions(
  roleId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles/{roleId}/actions', {
      params: { path: { roleId } }
    })
  )
  return {
    action_ids: Array.isArray(res.action_ids) ? res.action_ids : [],
    available_action_ids: Array.isArray(res.available_action_ids) ? res.available_action_ids : [],
    disabled_action_ids: Array.isArray(res.disabled_action_ids) ? res.disabled_action_ids : [],
    actions: (Array.isArray(res.actions) ? res.actions : []).map(normalizeAction),
    expanded_package_ids: Array.isArray(res.expanded_package_ids) ? res.expanded_package_ids : [],
    derived_sources: (Array.isArray(res.derived_sources) ? res.derived_sources : []).map(
      normalizeActionDerivedSource
    )
  }
}

export async function fetchSetMyCollaborationWorkspaceBoundaryRoleActions(
  roleId: string,
  actionIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration-workspaces/current/boundary/roles/{roleId}/actions', 'put'> =
    { ids: actionIds }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}/actions',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryRolePackages(
  roleId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/current/boundary/roles/{roleId}/packages', {
      params: { path: { roleId } }
    })
  )
  return {
    package_ids: Array.isArray(res.package_ids) ? res.package_ids : [],
    packages: (Array.isArray(res.packages) ? res.packages : []).map(normalizeFeaturePackage),
    inherited: Boolean(res.inherited)
  }
}

export async function fetchSetMyCollaborationWorkspaceBoundaryRolePackages(
  roleId: string,
  packageIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration-workspaces/current/boundary/roles/{roleId}/packages', 'put'> =
    { ids: packageIds }
  const { error } = await v5Client.PUT(
    '/collaboration-workspaces/current/boundary/roles/{roleId}/packages',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceBoundaryPackages(appKey?: string) {
  void appKey
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current/boundary/packages'))
  return {
    package_ids: Array.isArray(res.package_ids) ? res.package_ids : [],
    packages: (Array.isArray(res.packages) ? res.packages : []).map(normalizeFeaturePackage)
  }
}

export async function fetchGetCollaborationWorkspaceActions(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/actions', {
      params: {
        path: { id: collaborationWorkspaceId }
      }
    })
  )
  return {
    action_ids: Array.isArray(res.action_ids) ? res.action_ids : [],
    actions: (Array.isArray(res.actions) ? res.actions : []).map(normalizeAction)
  }
}

export async function fetchGetCollaborationWorkspaceMenus(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/menus', {
      params: {
        path: { id: collaborationWorkspaceId }
      }
    })
  )
  return { menu_ids: Array.isArray(res.menu_ids) ? res.menu_ids : [] }
}

export async function fetchGetCollaborationWorkspaceActionOrigins(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/action-origins', {
      params: {
        path: { id: collaborationWorkspaceId }
      }
    })
  )
  return {
    derived_action_ids: Array.isArray(res.derived_action_ids) ? res.derived_action_ids : [],
    derived_sources: (Array.isArray(res.derived_sources) ? res.derived_sources : []).map(
      normalizeActionDerivedSource
    ),
    blocked_action_ids: Array.isArray(res.blocked_action_ids) ? res.blocked_action_ids : []
  }
}

export async function fetchGetCollaborationWorkspaceMenuOrigins(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration-workspaces/{id}/menu-origins', {
      params: {
        path: { id: collaborationWorkspaceId }
      }
    })
  )
  return {
    derived_menu_ids: Array.isArray(res.derived_menu_ids) ? res.derived_menu_ids : [],
    derived_sources: (Array.isArray(res.derived_sources) ? res.derived_sources : []).map(
      normalizeMenuDerivedSource
    ),
    blocked_menu_ids: Array.isArray(res.blocked_menu_ids) ? res.blocked_menu_ids : []
  }
}

export async function fetchSetCollaborationWorkspaceActions(
  collaborationWorkspaceId: string,
  actionIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration-workspaces/{id}/actions', 'put'> = {
    ids: actionIds
  }
  const { error } = await v5Client.PUT('/collaboration-workspaces/{id}/actions', {
    params: {
      path: { id: collaborationWorkspaceId }
    },
    body
  })
  if (error) throw error
}

export async function fetchSetCollaborationWorkspaceMenus(
  collaborationWorkspaceId: string,
  menuIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration-workspaces/{id}/menus', 'put'> = {
    ids: menuIds
  }
  const { error } = await v5Client.PUT('/collaboration-workspaces/{id}/menus', {
    params: {
      path: { id: collaborationWorkspaceId }
    },
    body
  })
  if (error) throw error
}

export async function fetchGetMyCollaborationWorkspaceActions() {
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current/actions'))
  return {
    action_ids: Array.isArray(res.action_ids) ? res.action_ids : [],
    actions: (Array.isArray(res.actions) ? res.actions : []).map(normalizeAction)
  }
}

export async function fetchGetMyCollaborationWorkspaceActionOrigins() {
  const res = await unwrap(v5Client.GET('/collaboration-workspaces/current/action-origins'))
  return {
    derived_action_ids: Array.isArray(res.derived_action_ids) ? res.derived_action_ids : [],
    derived_sources: (Array.isArray(res.derived_sources) ? res.derived_sources : []).map(
      normalizeActionDerivedSource
    ),
    blocked_action_ids: Array.isArray(res.blocked_action_ids) ? res.blocked_action_ids : []
  }
}
