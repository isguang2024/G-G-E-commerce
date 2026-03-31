import request from '@/utils/http'

const TENANT_BASE = '/api/v1/tenants'
const MY_TEAM_BOUNDARY_ROLE_BASE = `${TENANT_BASE}/my-team/boundary/roles`

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
    module === 'permission_key' ||
    module === 'api_endpoint' ||
    module === 'feature_package'
  ) {
    return 'platform'
  }
  return 'team'
}

function normalizeTeam(item: any): Api.SystemManage.TeamListItem {
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

function normalizeRoleLabel(roleCode?: string) {
  return roleCode === 'team_admin' ? '团队管理员' : '团队成员'
}

function normalizeTeamMember(item: any): Api.SystemManage.TeamMemberItem {
  const roleCode = item?.role_code || item?.roleCode || ''
  return {
    id: item?.id || '',
    tenantId: item?.tenant_id || item?.tenantId || '',
    userId: item?.user_id || item?.userId || '',
    roleCode,
    role: normalizeRoleLabel(roleCode),
    status: item?.status || 'active',
    joinedAt: item?.joined_at || item?.joinedAt || '',
    userName: item?.user_name || item?.userName || '',
    nickName: item?.nick_name || item?.nickName || '',
    userEmail: item?.user_email || item?.userEmail || '',
    avatar: item?.avatar || ''
  }
}

export async function fetchGetTeamList(params: Api.SystemManage.TeamSearchParams) {
  const res = await request.get<Api.SystemManage.TeamList>({
    url: TENANT_BASE,
    params
  })

  return {
    ...res,
    records: (res?.records || []).map(normalizeTeam)
  }
}

export async function fetchGetTeamOptions(params?: Partial<Api.SystemManage.TeamSearchParams>) {
  const res = await request.get<{ records: Api.SystemManage.TeamListItem[]; total: number }>({
    url: `${TENANT_BASE}/options`,
    params
  })

  return {
    records: (res?.records || []).map(normalizeTeam),
    total: res?.total || 0
  }
}

export async function fetchGetTeam(id: string) {
  const res = await request.get<Api.SystemManage.TeamListItem>({
    url: `${TENANT_BASE}/${id}`
  })
  return normalizeTeam(res)
}

export function fetchCreateTeam(data: Api.SystemManage.TeamCreateParams) {
  return request.post<{ id: string }>({
    url: TENANT_BASE,
    data
  })
}

export function fetchUpdateTeam(id: string, data: Api.SystemManage.TeamUpdateParams) {
  return request.put<void>({
    url: `${TENANT_BASE}/${id}`,
    data
  })
}

export function fetchDeleteTeam(id: string) {
  return request.del<void>({
    url: `${TENANT_BASE}/${id}`
  })
}

export async function fetchGetTeamMembers(
  teamId: string,
  params?: { user_id?: string; user_name?: string; role_code?: string }
) {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/${teamId}/members`,
    params
  })
  return (res || []).map(normalizeTeamMember)
}

export function fetchAddTeamMember(teamId: string, data: { user_id: string; role_code?: string }) {
  return request.post<void>({
    url: `${TENANT_BASE}/${teamId}/members`,
    data: { user_id: data.user_id, role_code: data.role_code || 'team_member' }
  })
}

export function fetchRemoveTeamMember(teamId: string, userId: string) {
  return request.del<void>({
    url: `${TENANT_BASE}/${teamId}/members/${userId}`
  })
}

export function fetchUpdateTeamMemberRole(teamId: string, userId: string, roleCode: string) {
  return request.put<void>({
    url: `${TENANT_BASE}/${teamId}/members/${userId}/role`,
    data: { role_code: roleCode }
  })
}

export async function fetchGetMyTeam() {
  const res = await request.get<Api.SystemManage.TeamListItem>({
    url: `${TENANT_BASE}/my-team`
  })
  return normalizeTeam(res)
}

export async function fetchGetMyTeams() {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/my-teams`,
    skipTenantHeader: true,
    showErrorMessage: false
  })
  return (res || []).map(normalizeTeam)
}

export async function fetchGetMyTeamMembers() {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/my-team/members`
  })
  return (res || []).map(normalizeTeamMember)
}

export function fetchAddMyTeamMember(data: { user_id: string; role_code?: string }) {
  return request.post<void>({
    url: `${TENANT_BASE}/my-team/members`,
    data: { user_id: data.user_id, role_code: data.role_code || 'team_member' }
  })
}

export function fetchRemoveMyTeamMember(userId: string) {
  return request.del<void>({
    url: `${TENANT_BASE}/my-team/members/${userId}`
  })
}

export function fetchUpdateMyTeamMemberRole(userId: string, roleCode: string) {
  return request.put<void>({
    url: `${TENANT_BASE}/my-team/members/${userId}/role`,
    data: { role_code: roleCode }
  })
}

export function fetchGetMyTeamMemberRoles(userId: string) {
  return request.get<{
    role_ids: string[]
    global_role_ids?: string[]
    team_role_ids?: string[]
  }>({
    url: `${TENANT_BASE}/my-team/members/${userId}/roles`
  })
}

export function fetchSetMyTeamMemberRoles(userId: string, roleIds: string[]) {
  return request.put<void>({
    url: `${TENANT_BASE}/my-team/members/${userId}/roles`,
    data: { role_ids: roleIds }
  })
}

export async function fetchGetMyTeamRoles() {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/my-team/roles`
  })

  return (res || []).map((item: any) => ({
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    tenantId: item?.tenant_id || item?.tenantId || '',
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? false),
    createTime: item?.create_time || item?.created_at || ''
  }))
}

export async function fetchGetTeamRoles(teamId: string) {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/${teamId}/roles`
  })

  return (res || []).map((item: any) => ({
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    tenantId: item?.tenant_id || item?.tenantId || '',
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? false),
    createTime: item?.create_time || item?.created_at || ''
  }))
}

export async function fetchGetMyTeamBoundaryRoles() {
  const res = await request.get<any[]>({
    url: MY_TEAM_BOUNDARY_ROLE_BASE
  })
  return (res || []).map((item: any) => ({
    roleId: item?.id || '',
    roleCode: item?.code || '',
    roleName: item?.name || '',
    description: item?.description || '',
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    tenantId: item?.tenant_id || item?.tenantId || '',
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? false),
    createTime: item?.create_time || item?.created_at || ''
  }))
}

export function fetchCreateMyTeamRole(data: Api.SystemManage.RoleCreateParams) {
  return request.post<{ roleId: string }>({
    url: `${TENANT_BASE}/my-team/roles`,
    data
  })
}

export function fetchUpdateMyTeamRole(roleId: string, data: Api.SystemManage.RoleUpdateParams) {
  return request.put<void>({
    url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}`,
    data
  })
}

export function fetchDeleteMyTeamRole(roleId: string) {
  return request.del<void>({
    url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}`
  })
}

export function fetchGetMyTeamBoundaryRoleMenus(roleId: string) {
  return request
    .get<Api.SystemManage.RoleMenuBoundaryResponse>({
      url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}/menus`
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

export function fetchSetMyTeamBoundaryRoleMenus(roleId: string, menuIds: string[]) {
  return request.put<void>({
    url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}/menus`,
    data: { menu_ids: menuIds }
  })
}

export function fetchGetMyTeamBoundaryRoleActions(roleId: string) {
  return request
    .get<Api.SystemManage.RoleActionBoundaryResponse>({
      url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}/actions`
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

export function fetchSetMyTeamBoundaryRoleActions(roleId: string, actionIds: string[]) {
  return request.put<void>({
    url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}/actions`,
    data: { action_ids: actionIds }
  })
}

export function fetchGetMyTeamBoundaryRolePackages(roleId: string) {
  return request
    .get<Api.SystemManage.RoleFeaturePackageResponse>({
      url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}/packages`
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage),
      inherited: Boolean(res?.inherited)
    }))
}

export function fetchSetMyTeamBoundaryRolePackages(roleId: string, packageIds: string[]) {
  return request.put<void>({
    url: `${MY_TEAM_BOUNDARY_ROLE_BASE}/${roleId}/packages`,
    data: { package_ids: packageIds }
  })
}

export function fetchGetMyTeamBoundaryPackages() {
  return request
    .get<{ package_ids: string[]; packages: any[] }>({
      url: `${TENANT_BASE}/my-team/boundary/packages`
    })
    .then((res) => ({
      package_ids: res?.package_ids || [],
      packages: (res?.packages || []).map(normalizeFeaturePackage)
    }))
}

export async function fetchGetTeamActions(teamId: string) {
  const res = await request.get<{ action_ids: string[]; actions: any[] }>({
    url: `${TENANT_BASE}/${teamId}/actions`
  })
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export async function fetchGetTeamMenus(teamId: string) {
  const res = await request.get<{ menu_ids: string[] }>({
    url: `${TENANT_BASE}/${teamId}/menus`
  })
  return {
    menu_ids: res?.menu_ids || []
  }
}

export function fetchGetTeamActionOrigins(teamId: string) {
  return request
    .get<Api.SystemManage.TeamActionOriginsResponse>({
      url: `${TENANT_BASE}/${teamId}/action-origins`
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

export function fetchGetTeamMenuOrigins(teamId: string) {
  return request
    .get<{
      derived_menu_ids: string[]
      derived_sources?: Array<{ menu_id: string; package_ids: string[] }>
      blocked_menu_ids: string[]
    }>({
      url: `${TENANT_BASE}/${teamId}/menu-origins`
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

export function fetchSetTeamActions(teamId: string, actionIds: string[]) {
  return request.put<void>({
    url: `${TENANT_BASE}/${teamId}/actions`,
    data: { action_ids: actionIds }
  })
}

export function fetchSetTeamMenus(teamId: string, menuIds: string[]) {
  return request.put<void>({
    url: `${TENANT_BASE}/${teamId}/menus`,
    data: { menu_ids: menuIds }
  })
}

export async function fetchGetMyTeamActions() {
  const res = await request.get<{ action_ids: string[]; actions: any[] }>({
    url: `${TENANT_BASE}/my-team/actions`
  })
  return {
    action_ids: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export function fetchGetMyTeamActionOrigins() {
  return request
    .get<Api.SystemManage.TeamActionOriginsResponse>({
      url: `${TENANT_BASE}/my-team/action-origins`
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
