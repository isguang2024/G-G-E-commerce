import request from '@/utils/http'

const COLLABORATION_WORKSPACE_BASE = '/api/v1/collaboration-workspaces'
const CURRENT_COLLABORATION_BASE = `${COLLABORATION_WORKSPACE_BASE}/current`
const CURRENT_BOUNDARY_ROLE_BASE = `${CURRENT_COLLABORATION_BASE}/boundary/roles`
const TENANT_BASE = COLLABORATION_WORKSPACE_BASE

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
    key.startsWith('collaboration_workspace.') ||
    key.startsWith('platform.') ||
    key === 'collaboration_workspace.manage' ||
    module === 'role' ||
    module === 'user' ||
    module === 'menu' ||
    module === 'menu_backup' ||
    module === 'permission_action' ||
    module === 'permission_key' ||
    module === 'api_endpoint' ||
    module === 'feature_package'
  ) {
    return 'platform'
  }
  return 'collaboration'
}

function normalizeCollaborationWorkspace(
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
    collaborationWorkspaceId:
      item?.collaboration_workspace_id ||
      item?.collaborationWorkspaceId ||
      item?.workspace_id ||
      item?.workspaceId ||
      '',
    workspaceId: item?.workspace_id || item?.workspaceId || '',
    workspaceType: item?.workspace_type || item?.workspaceType || 'collaboration',
    legacyCollaborationWorkspaceId:
      item?.legacy_collaboration_workspace_id ||
      item?.legacyCollaborationWorkspaceId ||
      item?.collaboration_workspace_id ||
      item?.id ||
      '',
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
  const contextType =
    item?.context_type ||
    item?.contextType ||
    (packageKey.startsWith('platform.')
      ? 'platform'
      : packageKey.startsWith('common.')
        ? 'common'
        : 'collaboration')
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
    collaborationWorkspaceCount: item?.team_count ?? item?.collaborationWorkspaceCount ?? 0,
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
    legacyCollaborationWorkspaceId:
      item?.legacy_collaboration_workspace_id ||
      item?.legacyCollaborationWorkspaceId ||
      item?.id ||
      '',
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
  const res = await request.get<Api.SystemManage.CollaborationWorkspaceList>({
    url: TENANT_BASE,
    params
  })

  return {
    ...res,
    records: (res?.records || []).map(normalizeCollaborationWorkspace)
  }
}

export async function fetchGetCollaborationWorkspaceOptions(
  params?: Partial<Api.SystemManage.CollaborationWorkspaceSearchParams>
) {
  const res = await request.get<{
    records: Api.SystemManage.CollaborationWorkspaceListItem[]
    total: number
  }>({
    url: `${TENANT_BASE}/options`,
    params
  })

  return {
    records: (res?.records || []).map(normalizeCollaborationWorkspace),
    total: res?.total || 0
  }
}

export async function fetchGetCollaborationWorkspace(id: string) {
  const res = await request.get<Api.SystemManage.CollaborationWorkspaceListItem>({
    url: `${TENANT_BASE}/${id}`
  })
  return normalizeCollaborationWorkspace(res)
}

export function fetchCreateCollaborationWorkspace(
  data: Api.SystemManage.CollaborationWorkspaceCreateParams
) {
  return request.post<{ id: string }>({
    url: TENANT_BASE,
    data
  })
}

export function fetchUpdateCollaborationWorkspace(
  id: string,
  data: Api.SystemManage.CollaborationWorkspaceUpdateParams
) {
  return request.put<void>({
    url: `${TENANT_BASE}/${id}`,
    data
  })
}

export function fetchDeleteCollaborationWorkspace(id: string) {
  return request.del<void>({
    url: `${TENANT_BASE}/${id}`
  })
}

export async function fetchGetCollaborationWorkspaceMembers(
  collaborationWorkspaceId: string,
  params?: { user_id?: string; user_name?: string; role_code?: string }
) {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/members`,
    params
  })
  return (res || []).map(normalizeCollaborationWorkspaceMember)
}

export function fetchAddCollaborationWorkspaceMember(
  collaborationWorkspaceId: string,
  data: { user_id: string; role_code?: string }
) {
  return request.post<void>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/members`,
    data: { user_id: data.user_id, role_code: data.role_code || 'collaboration_workspace_member' }
  })
}

export function fetchRemoveCollaborationWorkspaceMember(
  collaborationWorkspaceId: string,
  userId: string
) {
  return request.del<void>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/members/${userId}`
  })
}

export function fetchUpdateCollaborationWorkspaceMemberRole(
  collaborationWorkspaceId: string,
  userId: string,
  roleCode: string
) {
  return request.put<void>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/members/${userId}/role`,
    data: { role_code: roleCode }
  })
}

export async function fetchGetMyCollaborationWorkspace() {
  const res = await request.get<Api.SystemManage.CollaborationWorkspaceListItem>({
    url: CURRENT_COLLABORATION_BASE
  })
  return normalizeCollaborationWorkspace(res)
}

export async function fetchGetMyCollaborationWorkspaces() {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/mine`,
    skipCollaborationWorkspaceHeader: true,
    showErrorMessage: false
  })
  return (res || []).map(normalizeCollaborationWorkspace)
}

export async function fetchGetMyCollaborationWorkspaceMembers() {
  const res = await request.get<any[]>({
    url: `${CURRENT_COLLABORATION_BASE}/members`
  })
  return (res || []).map(normalizeCollaborationWorkspaceMember)
}

export function fetchAddMyCollaborationWorkspaceMember(data: {
  user_id: string
  role_code?: string
}) {
  return request.post<void>({
    url: `${CURRENT_COLLABORATION_BASE}/members`,
    data: { user_id: data.user_id, role_code: data.role_code || 'collaboration_workspace_member' }
  })
}

export function fetchRemoveMyCollaborationWorkspaceMember(userId: string) {
  return request.del<void>({
    url: `${CURRENT_COLLABORATION_BASE}/members/${userId}`
  })
}

export function fetchUpdateMyCollaborationWorkspaceMemberRole(userId: string, roleCode: string) {
  return request.put<void>({
    url: `${CURRENT_COLLABORATION_BASE}/members/${userId}/role`,
    data: { role_code: roleCode }
  })
}

export function fetchGetMyCollaborationWorkspaceMemberRoles(userId: string) {
  return request
    .get<any>({
      url: `${CURRENT_COLLABORATION_BASE}/members/${userId}/roles`
    })
    .then(
      (res) =>
        ({
          role_ids: res?.role_ids || [],
          roles: (res?.roles || []).map((item: any) => ({
            id: item?.id || '',
            code: item?.code || '',
            name: item?.name || ''
          })),
          global_role_ids: res?.global_role_ids || [],
          team_role_ids: res?.team_role_ids || [],
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
        }) satisfies Api.SystemManage.CollaborationWorkspaceMemberRoleBindingResponse
    )
}

export function fetchSetMyCollaborationWorkspaceMemberRoles(userId: string, roleIds: string[]) {
  return request.put<void>({
    url: `${CURRENT_COLLABORATION_BASE}/members/${userId}/roles`,
    data: { role_ids: roleIds }
  })
}

export async function fetchGetMyTeamRoles() {
  const res = await request.get<any[]>({
    url: `${CURRENT_COLLABORATION_BASE}/roles`
  })

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

export async function fetchGetTeamRoles(collaborationWorkspaceId: string) {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/roles`
  })

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

export async function fetchGetMyTeamBoundaryRoles(appKey?: string) {
  const res = await request.get<any[]>({
    url: CURRENT_BOUNDARY_ROLE_BASE,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    }
  })
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

export function fetchCreateMyTeamRole(data: Api.SystemManage.RoleCreateParams) {
  return request.post<{ roleId: string }>({
    url: `${CURRENT_COLLABORATION_BASE}/roles`,
    data
  })
}

export function fetchUpdateMyTeamRole(roleId: string, data: Api.SystemManage.RoleUpdateParams) {
  return request.put<void>({
    url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}`,
    data
  })
}

export function fetchDeleteMyTeamRole(roleId: string) {
  return request.del<void>({
    url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}`
  })
}

export function fetchGetMyTeamBoundaryRoleMenus(roleId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RoleMenuBoundaryResponse>({
      url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}/menus`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      menu_ids: res?.menu_ids || [],
      available_menu_ids: res?.available_menu_ids || [],
      hidden_menu_ids: res?.hidden_menu_ids || [],
      expanded_package_ids: res?.expanded_package_ids || [],
      derived_sources: (res?.derived_sources || []).map((item) => ({
        menu_id: item?.menu_id || '',
        package_ids: item?.package_ids || []
      }))
    }))
}

export function fetchSetMyTeamBoundaryRoleMenus(
  roleId: string,
  menuIds: string[],
  appKey?: string
) {
  return request.put<void>({
    url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}/menus`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { menu_ids: menuIds }
  })
}

export function fetchGetMyTeamBoundaryRoleActions(roleId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RoleActionBoundaryResponse>({
      url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}/actions`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      action_ids: res?.action_ids || [],
      available_action_ids: res?.available_action_ids || [],
      disabled_action_ids: res?.disabled_action_ids || [],
      actions: (res?.actions || []).map(normalizeAction),
      expanded_package_ids: res?.expanded_package_ids || [],
      derived_sources: (res?.derived_sources || []).map((item) => ({
        action_id: item?.action_id || '',
        package_ids: item?.package_ids || []
      }))
    }))
}

export function fetchSetMyTeamBoundaryRoleActions(
  roleId: string,
  actionIds: string[],
  appKey?: string
) {
  return request.put<void>({
    url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}/actions`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { action_ids: actionIds }
  })
}

export function fetchGetMyTeamBoundaryRolePackages(roleId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.RoleFeaturePackageResponse>({
      url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}/packages`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage),
      inherited: Boolean(res?.inherited)
    }))
}

export function fetchSetMyTeamBoundaryRolePackages(
  roleId: string,
  packageIds: string[],
  appKey?: string
) {
  return request.put<void>({
    url: `${CURRENT_BOUNDARY_ROLE_BASE}/${roleId}/packages`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { package_ids: packageIds }
  })
}

export function fetchGetMyTeamBoundaryPackages(appKey?: string) {
  return request
    .get<{ package_ids: string[]; packages: any[] }>({
      url: `${CURRENT_COLLABORATION_BASE}/boundary/packages`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

export async function fetchGetTeamActions(collaborationWorkspaceId: string, appKey?: string) {
  const res = await request.get<{ action_ids: string[]; actions: any[] }>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/actions`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    }
  })
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export async function fetchGetTeamMenus(collaborationWorkspaceId: string, appKey?: string) {
  const res = await request.get<{ menu_ids: string[] }>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/menus`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    }
  })
  return {
    menu_ids: res?.menu_ids || []
  }
}

export function fetchGetTeamActionOrigins(collaborationWorkspaceId: string, appKey?: string) {
  return request
    .get<Api.SystemManage.TeamActionOriginsResponse>({
      url: `${TENANT_BASE}/${collaborationWorkspaceId}/action-origins`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      derived_action_ids: res?.derived_action_ids || [],
      derived_sources: (res?.derived_sources || []).map((item) => ({
        action_id: item?.action_id || '',
        package_ids: item?.package_ids || []
      })),
      blocked_action_ids: res?.blocked_action_ids || []
    }))
}

export function fetchGetTeamMenuOrigins(collaborationWorkspaceId: string, appKey?: string) {
  return request
    .get<{
      derived_menu_ids: string[]
      derived_sources?: Array<{ menu_id: string; package_ids: string[] }>
      blocked_menu_ids: string[]
    }>({
      url: `${TENANT_BASE}/${collaborationWorkspaceId}/menu-origins`,
      params: {
        ...(appKey ? { app_key: appKey } : {})
      }
    })
    .then((res) => ({
      derived_menu_ids: res?.derived_menu_ids || [],
      derived_sources: (res?.derived_sources || []).map((item) => ({
        menu_id: item?.menu_id || '',
        package_ids: item?.package_ids || []
      })),
      blocked_menu_ids: res?.blocked_menu_ids || []
    }))
}

export function fetchSetTeamActions(
  collaborationWorkspaceId: string,
  actionIds: string[],
  appKey?: string
) {
  return request.put<void>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/actions`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { action_ids: actionIds }
  })
}

export function fetchSetTeamMenus(
  collaborationWorkspaceId: string,
  menuIds: string[],
  appKey?: string
) {
  return request.put<void>({
    url: `${TENANT_BASE}/${collaborationWorkspaceId}/menus`,
    params: {
      ...(appKey ? { app_key: appKey } : {})
    },
    data: { menu_ids: menuIds }
  })
}

export async function fetchGetMyTeamActions() {
  const res = await request.get<{ action_ids: string[]; actions: any[] }>({
    url: `${CURRENT_COLLABORATION_BASE}/actions`
  })
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export function fetchGetMyTeamActionOrigins() {
  return request
    .get<Api.SystemManage.TeamActionOriginsResponse>({
      url: `${CURRENT_COLLABORATION_BASE}/action-origins`
    })
    .then((res) => ({
      derived_action_ids: res?.derived_action_ids || [],
      derived_sources: (res?.derived_sources || []).map((item) => ({
        action_id: item?.action_id || '',
        package_ids: item?.package_ids || []
      })),
      blocked_action_ids: res?.blocked_action_ids || []
    }))
}

export const fetchGetTeamList = fetchGetCollaborationWorkspaceList
export const fetchGetTeamOptions = fetchGetCollaborationWorkspaceOptions
export const fetchGetTeam = fetchGetCollaborationWorkspace
export const fetchCreateTeam = fetchCreateCollaborationWorkspace
export const fetchUpdateTeam = fetchUpdateCollaborationWorkspace
export const fetchDeleteTeam = fetchDeleteCollaborationWorkspace
export const fetchGetMyTeam = fetchGetMyCollaborationWorkspace
export const fetchGetMyTeams = fetchGetMyCollaborationWorkspaces
