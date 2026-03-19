import request from '@/utils/http'

const TENANT_BASE = '/api/v1/tenants'

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
  return {
    id: item?.id || '',
    resourceCode: item?.resource_code || item?.resourceCode || '',
    actionCode: item?.action_code || item?.actionCode || '',
    name: item?.name || '',
    description: item?.description || '',
    scopeId: item?.scope_id || item?.scopeId || '',
    scopeCode: item?.scope_code || item?.scopeCode || item?.scope || '',
    scopeName: item?.scope_name || item?.scopeName || '',
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
  params?: { user_id?: string; user_name?: string; role?: string }
) {
  const res = await request.get<any[]>({
    url: `${TENANT_BASE}/${teamId}/members`,
    params
  })
  return (res || []).map(normalizeTeamMember)
}

export function fetchAddTeamMember(teamId: string, data: { user_id: string; role?: string }) {
  return request.post<void>({
    url: `${TENANT_BASE}/${teamId}/members`,
    data: { user_id: data.user_id, role: data.role || 'team_member' }
  })
}

export function fetchRemoveTeamMember(teamId: string, userId: string) {
  return request.del<void>({
    url: `${TENANT_BASE}/${teamId}/members/${userId}`
  })
}

export function fetchUpdateTeamMemberRole(teamId: string, userId: string, role: string) {
  return request.put<void>({
    url: `${TENANT_BASE}/${teamId}/members/${userId}/role`,
    data: { role }
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

export function fetchAddMyTeamMember(data: { user_id: string; role?: string }) {
  return request.post<void>({
    url: `${TENANT_BASE}/my-team/members`,
    data: { user_id: data.user_id, role_code: data.role || 'team_member' }
  })
}

export function fetchRemoveMyTeamMember(userId: string) {
  return request.del<void>({
    url: `${TENANT_BASE}/my-team/members/${userId}`
  })
}

export function fetchUpdateMyTeamMemberRole(userId: string, role: string) {
  return request.put<void>({
    url: `${TENANT_BASE}/my-team/members/${userId}/role`,
    data: { role }
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
    scope: item?.scope || '',
    scopes: Array.isArray(item?.scopes)
      ? item.scopes.map((scope: any) => ({
          scopeId: scope?.scopeId || scope?.scope_id || '',
          scopeCode: scope?.scopeCode || scope?.scope_code || '',
          scopeName: scope?.scopeName || scope?.scope_name || '',
          dataPermissionCode:
            scope?.dataPermissionCode || scope?.data_permission_code || '',
          dataPermissionName:
            scope?.dataPermissionName || scope?.data_permission_name || ''
        }))
      : [],
    status: item?.status || 'normal',
    isSystem: Boolean(item?.is_system ?? item?.isSystem ?? false),
    createTime: item?.created_at || ''
  }))
}

export function fetchGetMyTeamRoleActions(roleId: string) {
  return request.get<{ action_ids: string[] }>({
    url: `${TENANT_BASE}/my-team/roles/${roleId}/actions`
  })
}

export async function fetchGetTeamActions(teamId: string) {
  const res = await request.get<{ action_ids: string[]; actions: any[] }>({
    url: `${TENANT_BASE}/${teamId}/actions`
  })
  return {
    actionIds: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export function fetchSetTeamActions(teamId: string, actionIds: string[]) {
  return request.put<void>({
    url: `${TENANT_BASE}/${teamId}/actions`,
    data: { action_ids: actionIds }
  })
}

export async function fetchGetMyTeamActions() {
  const res = await request.get<{ action_ids: string[]; actions: any[] }>({
    url: `${TENANT_BASE}/my-team/actions`
  })
  return {
    actionIds: res?.action_ids || [],
    actions: (res?.actions || []).map(normalizeAction)
  }
}

export async function fetchGetMyTeamMemberActions(userId: string) {
  const res = await request.get<{ actions: any[] }>({
    url: `${TENANT_BASE}/my-team/members/${userId}/actions`
  })
  return (res?.actions || []).map((item: any) => ({
    actionId: item?.action_id || item?.actionId || '',
    effect: item?.effect || 'allow',
    action: item?.action ? normalizeAction(item.action) : undefined
  })) as Api.SystemManage.TeamMemberActionPermissionItem[]
}

export function fetchSetMyTeamMemberActions(
  userId: string,
  actions: Array<{ action_id: string; effect: 'allow' | 'deny' }>
) {
  return request.put<void>({
    url: `${TENANT_BASE}/my-team/members/${userId}/actions`,
    data: { actions }
  })
}
