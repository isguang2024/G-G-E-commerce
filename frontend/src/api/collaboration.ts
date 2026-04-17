import {
  v5Client,
  unwrap,
  toV5Body,
  type V5Query,
  type V5RequestBody
} from '@/domains/governance/api/_shared'

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

type V5CollaborationLike = {
  id?: string
  collaboration_workspace_id?: string
  workspace_id?: string
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
  workspace_count?: number
  collaborationWorkspaceCount?: number
  status?: string
  sort_order?: number
  created_at?: string
  updated_at?: string
}

type V5CollaborationMemberLike = {
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
  scope_id?: string | null
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
  if (
    key === 'workspace.manage' ||
    module === 'workspace'
  ) {
    return 'personal'
  }
  if (
    key.startsWith('collaboration.member.') ||
    key.startsWith('collaboration.boundary.') ||
    key.startsWith('collaboration.message.') ||
    module === 'collaboration_member' ||
    module === 'collaboration_boundary' ||
    module === 'collaboration_message'
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
    module === 'workspace_member_admin'
  ) {
    return 'personal'
  }
  return 'common'
}

function normalizeCollaboration(
  item: V5CollaborationLike | undefined
): Api.SystemManage.CollaborationWorkspaceListItem {
  const collaborationWorkspaceId = item?.workspace_id || item?.collaboration_workspace_id || ''
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
    collaborationWorkspaceCount: item?.workspace_count ?? item?.collaborationWorkspaceCount ?? 0,
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? 0,
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

function normalizeRoleLabel(roleCode?: string) {
  const normalized = `${roleCode || ''}`.trim()
  return normalized === 'collaboration_admin' ? '协作空间管理员' : '协作空间成员'
}

function normalizeRoleSummary(item: V5RoleSummaryLike | undefined): Api.SystemManage.RoleListItem {
  return {
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? false),
    collaborationWorkspaceId: item?.scope_id || item?.collaboration_workspace_id || '',
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

function normalizeCollaborationMember(
  item: V5CollaborationMemberLike | undefined
): Api.SystemManage.CollaborationWorkspaceMemberItem {
  const roleCode = item?.role_code || ''
  return {
    id: item?.id || '',
    collaborationWorkspaceId: item?.workspace_id || item?.collaboration_workspace_id || '',
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

export async function fetchGetCollaborationList(
  params: Api.SystemManage.CollaborationSearchParams
) {
  const query: V5Query<'/workspaces/collaboration', 'get'> = params
  const res = await unwrap(v5Client.GET('/workspaces/collaboration', { params: { query } }))
  return {
    ...res,
    records: (res.records || []).map(normalizeCollaboration)
  } as Api.SystemManage.CollaborationWorkspaceList
}

export async function fetchGetCollaborationOptions(
  params?: Partial<Api.SystemManage.CollaborationSearchParams>
) {
  void params
  const res = await unwrap(v5Client.GET('/workspaces/collaboration/options'))
  return {
    records: (res.records || []).map(normalizeCollaboration),
    total: res.total || 0
  }
}

export async function fetchGetCollaboration(id: string) {
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}', { params: { path: { id } } })
  )
  return normalizeCollaboration(res)
}

export async function fetchCreateCollaboration(
  data: Api.SystemManage.CollaborationWorkspaceCreateParams
) {
  const body: V5RequestBody<'/workspaces/collaboration', 'post'> = {
    name: data.name,
    remark: data.remark,
    logo_url: data.logo_url,
    plan: data.plan,
    max_members: data.max_members,
    status: data.status,
    admin_user_ids: data.admin_user_ids
  }
  const res = await unwrap(v5Client.POST('/workspaces/collaboration', { body }))
  return { id: res.id }
}

export async function fetchUpdateCollaboration(
  id: string,
  data: Api.SystemManage.CollaborationWorkspaceUpdateParams
) {
  const body: V5RequestBody<'/workspaces/collaboration/{id}', 'put'> = {
    name: data.name || '',
    remark: data.remark,
    logo_url: data.logo_url,
    plan: data.plan,
    max_members: data.max_members,
    status: data.status,
    admin_user_ids: data.admin_user_ids
  }
  const { error } = await v5Client.PUT('/workspaces/collaboration/{id}', {
    params: { path: { id } },
    body
  })
  if (error) throw error
}

export async function fetchDeleteCollaboration(id: string) {
  const { error } = await v5Client.DELETE('/workspaces/collaboration/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

export async function fetchGetCollaborationMembers(
  collaborationWorkspaceId: string,
  params?: { user_id?: string; user_name?: string; role_code?: string }
) {
  void params
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}/members', {
      params: { path: { id: collaborationWorkspaceId } }
    })
  )
  return (res.records || []).map(normalizeCollaborationMember)
}

export async function fetchAddCollaborationMember(
  collaborationWorkspaceId: string,
  data: { user_id: string; role_code?: string }
) {
  const body: V5RequestBody<'/workspaces/collaboration/{id}/members', 'post'> = {
    user_id: data.user_id,
    role_code: data.role_code || 'collaboration_member'
  }
  const { error } = await v5Client.POST('/workspaces/collaboration/{id}/members', {
    params: { path: { id: collaborationWorkspaceId } },
    body
  })
  if (error) throw error
}

export async function fetchRemoveCollaborationMember(
  collaborationWorkspaceId: string,
  userId: string
) {
  const { error } = await v5Client.DELETE('/workspaces/collaboration/{id}/members/{userId}', {
    params: { path: { id: collaborationWorkspaceId, userId } }
  })
  if (error) throw error
}

export async function fetchUpdateCollaborationMemberRole(
  collaborationWorkspaceId: string,
  userId: string,
  roleCode: string
) {
  const body: V5RequestBody<'/workspaces/collaboration/{id}/members/{userId}/role', 'put'> = {
    role_code: roleCode
  }
  const { error } = await v5Client.PUT(
    '/workspaces/collaboration/{id}/members/{userId}/role',
    {
      params: { path: { id: collaborationWorkspaceId, userId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaboration() {
  const res = await unwrap(v5Client.GET('/collaboration/current'))
  return normalizeCollaboration(res)
}

export async function fetchGetMyCollaborations() {
  const res = await unwrap(v5Client.GET('/workspaces/collaboration/mine'))
  return (res.records || []).map(normalizeCollaboration)
}

export async function fetchGetMyCollaborationMembers() {
  const res = await unwrap(v5Client.GET('/collaboration/current/members'))
  return (res.records || []).map(normalizeCollaborationMember)
}

export async function fetchAddMyCollaborationMember(data: {
  user_id: string
  role_code?: string
}) {
  const body: V5RequestBody<'/collaboration/current/members', 'post'> = {
    user_id: data.user_id,
    role_code: data.role_code || 'collaboration_member'
  }
  const { error } = await v5Client.POST('/collaboration/current/members', {
    body
  })
  if (error) throw error
}

export async function fetchRemoveMyCollaborationMember(userId: string) {
  const { error } = await v5Client.DELETE(
    '/collaboration/current/members/{userId}',
    { params: { path: { userId } } }
  )
  if (error) throw error
}

export async function fetchUpdateMyCollaborationMemberRole(
  userId: string,
  roleCode: string
) {
  const body: V5RequestBody<'/collaboration/current/members/{userId}/role', 'put'> = {
    role_code: roleCode
  }
  const { error } = await v5Client.PUT(
    '/collaboration/current/members/{userId}/role',
    {
      params: { path: { userId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationMemberRoles(userId: string) {
  const res = await unwrap(
    v5Client.GET('/collaboration/current/members/{userId}/roles', {
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

export async function fetchSetMyCollaborationMemberRoles(
  userId: string,
  roleIds: string[]
) {
  const body: V5RequestBody<'/collaboration/current/members/{userId}/roles', 'put'> = {
    ids: roleIds
  }
  const { error } = await v5Client.PUT(
    '/collaboration/current/members/{userId}/roles',
    {
      params: { path: { userId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationRoles() {
  const res = await unwrap(v5Client.GET('/collaboration/current/roles'))
  return (res.records || []).map(normalizeRoleSummary)
}

export async function fetchGetCollaborationRoles(collaborationWorkspaceId: string) {
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}/roles', {
      params: { path: { id: collaborationWorkspaceId } }
    })
  )
  return (res.records || []).map(normalizeRoleSummary)
}

export async function fetchGetMyCollaborationBoundaryRoles(appKey?: string) {
  void appKey
  const res = await unwrap(v5Client.GET('/collaboration/current/boundary/roles'))
  return (res.records || []).map(normalizeRoleSummary)
}

export async function fetchCreateMyCollaborationRole(
  data: Api.SystemManage.RoleCreateParams
) {
  const body: V5RequestBody<'/collaboration/current/roles', 'post'> = {
    code: data.code,
    name: data.name,
    description: data.description,
    sort_order: data.sort_order,
    status: data.status
  }
  return unwrap(v5Client.POST('/collaboration/current/roles', { body }))
}

export async function fetchUpdateMyCollaborationRole(
  roleId: string,
  data: Api.SystemManage.RoleUpdateParams
) {
  const body: V5RequestBody<'/collaboration/current/boundary/roles/{roleId}', 'put'> =
    {
      code: data.code || '',
      name: data.name || '',
      description: data.description,
      sort_order: data.sort_order,
      status: data.status
    }
  const { error } = await v5Client.PUT(
    '/collaboration/current/boundary/roles/{roleId}',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchDeleteMyCollaborationRole(roleId: string) {
  const { error } = await v5Client.DELETE(
    '/collaboration/current/boundary/roles/{roleId}',
    { params: { path: { roleId } } }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationBoundaryRoleMenus(
  roleId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration/current/boundary/roles/{roleId}/menus', {
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

export async function fetchSetMyCollaborationBoundaryRoleMenus(
  roleId: string,
  menuIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration/current/boundary/roles/{roleId}/menus', 'put'> =
    { ids: menuIds }
  const { error } = await v5Client.PUT(
    '/collaboration/current/boundary/roles/{roleId}/menus',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationBoundaryRoleActions(
  roleId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration/current/boundary/roles/{roleId}/actions', {
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

export async function fetchSetMyCollaborationBoundaryRoleActions(
  roleId: string,
  actionIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration/current/boundary/roles/{roleId}/actions', 'put'> =
    { ids: actionIds }
  const { error } = await v5Client.PUT(
    '/collaboration/current/boundary/roles/{roleId}/actions',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationBoundaryRolePackages(
  roleId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/collaboration/current/boundary/roles/{roleId}/packages', {
      params: { path: { roleId } }
    })
  )
  return {
    package_ids: Array.isArray(res.package_ids) ? res.package_ids : [],
    packages: (Array.isArray(res.packages) ? res.packages : []).map(normalizeFeaturePackage),
    inherited: Boolean(res.inherited)
  }
}

export async function fetchSetMyCollaborationBoundaryRolePackages(
  roleId: string,
  packageIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/collaboration/current/boundary/roles/{roleId}/packages', 'put'> =
    { ids: packageIds }
  const { error } = await v5Client.PUT(
    '/collaboration/current/boundary/roles/{roleId}/packages',
    {
      params: { path: { roleId } },
      body
    }
  )
  if (error) throw error
}

export async function fetchGetMyCollaborationBoundaryPackages(appKey?: string) {
  void appKey
  const res = await unwrap(v5Client.GET('/collaboration/current/boundary/packages'))
  return {
    package_ids: Array.isArray(res.package_ids) ? res.package_ids : [],
    packages: (Array.isArray(res.packages) ? res.packages : []).map(normalizeFeaturePackage)
  }
}

export async function fetchGetCollaborationActions(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}/actions', {
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

export async function fetchGetCollaborationMenus(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}/menus', {
      params: {
        path: { id: collaborationWorkspaceId }
      }
    })
  )
  return { menu_ids: Array.isArray(res.menu_ids) ? res.menu_ids : [] }
}

export async function fetchGetCollaborationActionOrigins(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}/action-origins', {
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

export async function fetchGetCollaborationMenuOrigins(
  collaborationWorkspaceId: string,
  appKey?: string
) {
  void appKey
  const res = await unwrap(
    v5Client.GET('/workspaces/collaboration/{id}/menu-origins', {
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

export async function fetchSetCollaborationActions(
  collaborationWorkspaceId: string,
  actionIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/workspaces/collaboration/{id}/actions', 'put'> = {
    ids: actionIds
  }
  const { error } = await v5Client.PUT('/workspaces/collaboration/{id}/actions', {
    params: {
      path: { id: collaborationWorkspaceId }
    },
    body
  })
  if (error) throw error
}

export async function fetchSetCollaborationMenus(
  collaborationWorkspaceId: string,
  menuIds: string[],
  appKey?: string
) {
  void appKey
  const body: V5RequestBody<'/workspaces/collaboration/{id}/menus', 'put'> = {
    ids: menuIds
  }
  const { error } = await v5Client.PUT('/workspaces/collaboration/{id}/menus', {
    params: {
      path: { id: collaborationWorkspaceId }
    },
    body
  })
  if (error) throw error
}

export async function fetchGetMyCollaborationActions() {
  const res = await unwrap(v5Client.GET('/collaboration/current/actions'))
  return {
    action_ids: Array.isArray(res.action_ids) ? res.action_ids : [],
    actions: (Array.isArray(res.actions) ? res.actions : []).map(normalizeAction)
  }
}

export async function fetchGetMyCollaborationActionOrigins() {
  const res = await unwrap(v5Client.GET('/collaboration/current/action-origins'))
  return {
    derived_action_ids: Array.isArray(res.derived_action_ids) ? res.derived_action_ids : [],
    derived_sources: (Array.isArray(res.derived_sources) ? res.derived_sources : []).map(
      normalizeActionDerivedSource
    ),
    blocked_action_ids: Array.isArray(res.blocked_action_ids) ? res.blocked_action_ids : []
  }
}
