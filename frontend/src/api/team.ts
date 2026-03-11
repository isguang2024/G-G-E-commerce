import request from '@/utils/http'

const TENANT_BASE = '/api/v1/tenants'

// ========== 团队管理 (管理员用) ==========

export function fetchGetTeamList(params: Api.SystemManage.TeamSearchParams) {
  return request.get<Api.SystemManage.TeamList>({
    url: TENANT_BASE,
    params
  })
}

export function fetchGetTeam(id: string) {
  return request.get<Api.SystemManage.TeamListItem>({
    url: `${TENANT_BASE}/${id}`
  })
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

export function fetchGetTeamMembers(teamId: string, params?: { user_id?: string; user_name?: string; role?: string }) {
  return request.get<{ records: Api.SystemManage.TeamMemberItem[] }>({
    url: `${TENANT_BASE}/${teamId}/members`,
    params
  })
}

export function fetchAddTeamMember(teamId: string, data: { user_id: string; role?: string }) {
  return request.post<void>({
    url: `${TENANT_BASE}/${teamId}/members`,
    data: { user_id: data.user_id, role: data.role || 'editor' }
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

// ========== 我的团队 (普通管理员/成员用) ==========

export function fetchGetMyTeam() {
  return request.get<Api.SystemManage.TeamListItem>({
    url: `${TENANT_BASE}/my-team`
  })
}

export function fetchGetMyTeamMembers() {
  return request.get<{ records: Api.SystemManage.TeamMemberItem[] }>({
    url: `${TENANT_BASE}/my-team/members`
  })
}

export function fetchAddMyTeamMember(data: { user_id: string; role?: string }) {
  return request.post<void>({
    url: `${TENANT_BASE}/my-team/members`,
    data: { user_id: data.user_id, role: data.role || 'editor' }
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

/** 我的团队 - 成员在本团队内的角色 */
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

/** 我的团队 - 角色列表（仅全局 scope=team 角色） */
export function fetchGetMyTeamRoles() {
  return request.get<{ records: Api.SystemManage.RoleListItem[] }>({
    url: `${TENANT_BASE}/my-team/roles`
  })
}
