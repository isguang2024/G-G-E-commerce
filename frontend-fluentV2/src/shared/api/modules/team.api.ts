import { requestData } from '@/shared/api/client'
import type {
  FeaturePackageRecord,
  PermissionActionRecord,
  RelationSourceRecord,
  SelectionRelation,
  TeamBoundaryOriginSummary,
  TeamMemberDetail,
  TeamMemberRecord,
  TeamRecord,
  TeamSavePayload,
  UserRoleSummary,
} from '@/shared/types/admin'

interface PaginationEnvelope<T> {
  current?: number
  total?: number
  size?: number
  records?: T[]
}

export interface PaginatedResult<T> {
  current: number
  total: number
  size: number
  records: T[]
}

function toPaginatedResult<T>(input: PaginationEnvelope<T>): PaginatedResult<T> {
  return {
    current: Number(input.current || 1),
    total: Number(input.total || 0),
    size: Number(input.size || input.records?.length || 0),
    records: Array.isArray(input.records) ? input.records : [],
  }
}

function normalizeText(value: unknown) {
  return `${value || ''}`.trim()
}

function normalizeTeam(input: Record<string, unknown>): TeamRecord {
  return {
    id: normalizeText(input.id),
    name: normalizeText(input.name),
    remark: normalizeText(input.remark),
    logoUrl: normalizeText(input.logo_url || input.logoUrl),
    plan: normalizeText(input.plan || 'free'),
    maxMembers: Number(input.max_members || input.maxMembers || 0),
    status: normalizeText(input.status || 'active'),
    ownerId: normalizeText(input.owner_id || input.ownerId),
    createdAt: normalizeText(input.created_at || input.createTime),
    updatedAt: normalizeText(input.updated_at || input.updateTime),
    adminUsers: Array.isArray(input.admin_users || input.adminUsers)
      ? ((input.admin_users || input.adminUsers) as unknown[]).map((item) => {
          const target = item as Record<string, unknown>
          return {
            userId: normalizeText(target.user_id || target.userId),
            userName: normalizeText(target.user_name || target.userName),
            nickName: normalizeText(target.nick_name || target.nickName),
          }
        })
      : [],
  }
}

function normalizeTeamMember(input: Record<string, unknown>): TeamMemberRecord {
  const roleCode = normalizeText(input.role_code || input.roleCode)
  return {
    id: normalizeText(input.id),
    tenantId: normalizeText(input.tenant_id || input.tenantId),
    userId: normalizeText(input.user_id || input.userId),
    roleCode,
    role: roleCode === 'team_admin' ? '团队管理员' : '团队成员',
    status: normalizeText(input.status || 'active'),
    joinedAt: normalizeText(input.joined_at || input.joinedAt),
    userName: normalizeText(input.user_name || input.userName),
    nickName: normalizeText(input.nick_name || input.nickName),
    userEmail: normalizeText(input.user_email || input.userEmail),
    userPhone: normalizeText(input.user_phone || input.userPhone),
    avatar: normalizeText(input.avatar),
  }
}

function normalizeTeamMemberDetail(input: Record<string, unknown>): TeamMemberDetail {
  const record = normalizeTeamMember(input)
  return {
    ...record,
    displayName: record.nickName || record.userName || record.userId,
    contactItems: [
      { label: '账号', value: record.userName || '-' },
      { label: '邮箱', value: record.userEmail || '-' },
      { label: '手机号', value: record.userPhone || '-' },
      { label: '加入时间', value: record.joinedAt || '-' },
    ],
    roleItems: [
      { label: '团队角色', value: record.role || record.roleCode || '-' },
      { label: '状态', value: record.status || '-' },
      { label: '团队 ID', value: record.tenantId || '-' },
    ],
  }
}

function normalizePermissionAction(input: Record<string, unknown>): PermissionActionRecord {
  return {
    id: normalizeText(input.id),
    permissionKey: normalizeText(input.permission_key || input.permissionKey),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    contextType: normalizeText(input.context_type || input.contextType),
    featureKind: normalizeText(input.feature_kind || input.featureKind || 'system'),
    moduleCode: normalizeText(input.module_code || input.moduleCode),
    moduleGroupId: '',
    featureGroupId: '',
    apiCount: 0,
    pageCount: 0,
    packageCount: 0,
    consumerTypes: [],
    usagePattern: '',
    usageNote: '',
    duplicatePattern: 'none',
    duplicateGroup: '',
    duplicateKeys: [],
    duplicateNote: '',
    status: normalizeText(input.status || 'normal'),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    isBuiltin: Boolean(input.is_builtin ?? input.isBuiltin),
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeFeaturePackage(input: Record<string, unknown>): FeaturePackageRecord {
  return {
    id: normalizeText(input.id),
    packageKey: normalizeText(input.package_key || input.packageKey),
    packageType: normalizeText(input.package_type || input.packageType || 'base'),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    contextType: normalizeText(input.context_type || input.contextType || 'team'),
    isBuiltin: Boolean(input.is_builtin ?? input.isBuiltin),
    actionCount: Number(input.action_count || input.actionCount || 0),
    menuCount: Number(input.menu_count || input.menuCount || 0),
    teamCount: Number(input.team_count || input.teamCount || 0),
    status: normalizeText(input.status || 'normal'),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeRelationSources(input?: Array<Record<string, unknown>>, fieldKey: 'action_id' | 'menu_id' = 'action_id') {
  if (!Array.isArray(input)) {
    return [] as RelationSourceRecord[]
  }

  return input.map((item) => ({
    entityId: normalizeText(item[fieldKey]),
    packageIds: Array.isArray(item.package_ids) ? item.package_ids.map(normalizeText).filter(Boolean) : [],
  }))
}

function normalizeOriginSummary(
  input: Record<string, unknown>,
  fieldKey: 'action_id' | 'menu_id',
): TeamBoundaryOriginSummary {
  return {
    derivedIds: Array.isArray(input.derived_action_ids || input.derived_menu_ids)
      ? ((input.derived_action_ids || input.derived_menu_ids) as unknown[]).map(normalizeText).filter(Boolean)
      : [],
    blockedIds: Array.isArray(input.blocked_action_ids || input.blocked_menu_ids)
      ? ((input.blocked_action_ids || input.blocked_menu_ids) as unknown[]).map(normalizeText).filter(Boolean)
      : [],
    derivedSources: normalizeRelationSources(
      Array.isArray(input.derived_sources) ? (input.derived_sources as Array<Record<string, unknown>>) : undefined,
      fieldKey,
    ),
  }
}

export async function fetchTeamList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/tenants',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeTeam(item)),
  }
}

export async function fetchTeamDetail(teamId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/tenants/${teamId}`,
  })

  return normalizeTeam(result)
}

export async function createTeam(payload: TeamSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/tenants',
    data: {
      name: payload.name,
      remark: payload.remark,
      plan: payload.plan,
      max_members: payload.maxMembers,
      status: payload.status,
    },
  })

  return normalizeTeam(result)
}

export async function updateTeam(teamId: string, payload: TeamSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/tenants/${teamId}`,
    data: {
      name: payload.name,
      remark: payload.remark,
      plan: payload.plan,
      max_members: payload.maxMembers,
      status: payload.status,
    },
  })

  return normalizeTeam(result)
}

export async function deleteTeam(teamId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/tenants/${teamId}`,
  })
}

export async function fetchTeamMembers(teamId: string, params?: Record<string, unknown>) {
  const result = await requestData<Array<Record<string, unknown>>>({
    method: 'GET',
    url: `/api/v1/tenants/${teamId}/members`,
    params,
  })

  return result.map((item) => normalizeTeamMemberDetail(item))
}

export async function addTeamMember(teamId: string, userId: string, roleCode = 'team_member') {
  await requestData<unknown>({
    method: 'POST',
    url: `/api/v1/tenants/${teamId}/members`,
    data: {
      user_id: userId,
      role_code: roleCode,
    },
  })
}

export async function removeTeamMember(teamId: string, userId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/tenants/${teamId}/members/${userId}`,
  })
}

export async function updateTeamMemberRole(teamId: string, userId: string, roleCode: string) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/${teamId}/members/${userId}/role`,
    data: {
      role_code: roleCode,
    },
  })
}

export async function fetchMyTeamMembers() {
  const result = await requestData<Array<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/tenants/my-team/members',
  })

  return result.map((item) => normalizeTeamMemberDetail(item))
}

export async function addMyTeamMember(userId: string, roleCode = 'team_member') {
  await requestData<unknown>({
    method: 'POST',
    url: '/api/v1/tenants/my-team/members',
    data: {
      user_id: userId,
      role_code: roleCode,
    },
  })
}

export async function removeMyTeamMember(userId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/tenants/my-team/members/${userId}`,
  })
}

export async function updateMyTeamMemberRole(userId: string, roleCode: string) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/my-team/members/${userId}/role`,
    data: {
      role_code: roleCode,
    },
  })
}

export async function fetchMyTeamMemberRoles(userId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/tenants/my-team/members/${userId}/roles`,
  })

  return Array.isArray(result.role_ids) ? result.role_ids.map(normalizeText).filter(Boolean) : []
}

export async function setMyTeamMemberRoles(userId: string, roleIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/my-team/members/${userId}/roles`,
    data: {
      role_ids: roleIds,
    },
  })
}

export async function fetchMyTeamBoundaryRoles() {
  const result = await requestData<Array<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/tenants/my-team/boundary/roles',
  })

  return result.map((item) => ({
    id: normalizeText(item.id),
    code: normalizeText(item.code),
    name: normalizeText(item.name),
    description: normalizeText(item.description),
  })) as UserRoleSummary[]
}

export async function createMyTeamBoundaryRole(payload: { roleName: string; roleCode: string; description: string }) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/tenants/my-team/boundary/roles',
    data: {
      name: payload.roleName,
      code: payload.roleCode,
      description: payload.description,
    },
  })

  return {
    id: normalizeText(result.id),
    code: normalizeText(result.code),
    name: normalizeText(result.name),
    description: normalizeText(result.description),
  } satisfies UserRoleSummary
}

export async function updateMyTeamBoundaryRole(roleId: string, payload: { roleName: string; roleCode: string; description: string }) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}`,
    data: {
      name: payload.roleName,
      code: payload.roleCode,
      description: payload.description,
    },
  })
}

export async function deleteMyTeamBoundaryRole(roleId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}`,
  })
}

export async function fetchMyTeamBoundaryRoleActions(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}/actions`,
  })

  return {
    ids: Array.isArray(result.action_ids) ? result.action_ids.map(normalizeText).filter(Boolean) : [],
    items: Array.isArray(result.actions)
      ? result.actions.map((item) => normalizePermissionAction(item as Record<string, unknown>))
      : [],
    availableIds: Array.isArray(result.available_action_ids)
      ? result.available_action_ids.map(normalizeText).filter(Boolean)
      : undefined,
    disabledIds: Array.isArray(result.disabled_action_ids)
      ? result.disabled_action_ids.map(normalizeText).filter(Boolean)
      : undefined,
    expandedPackageIds: Array.isArray(result.expanded_package_ids)
      ? result.expanded_package_ids.map(normalizeText).filter(Boolean)
      : undefined,
    derivedSources: normalizeRelationSources(
      Array.isArray(result.derived_sources) ? (result.derived_sources as Array<Record<string, unknown>>) : undefined,
    ),
  } satisfies SelectionRelation<PermissionActionRecord>
}

export async function setMyTeamBoundaryRoleActions(roleId: string, actionIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}/actions`,
    data: {
      action_ids: actionIds,
    },
  })
}

export async function fetchMyTeamBoundaryRoleMenus(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}/menus`,
  })

  return {
    ids: Array.isArray(result.menu_ids) ? result.menu_ids.map(normalizeText).filter(Boolean) : [],
    items: [],
    availableIds: Array.isArray(result.available_menu_ids)
      ? result.available_menu_ids.map(normalizeText).filter(Boolean)
      : undefined,
    hiddenIds: Array.isArray(result.hidden_menu_ids)
      ? result.hidden_menu_ids.map(normalizeText).filter(Boolean)
      : undefined,
    expandedPackageIds: Array.isArray(result.expanded_package_ids)
      ? result.expanded_package_ids.map(normalizeText).filter(Boolean)
      : undefined,
    derivedSources: normalizeRelationSources(
      Array.isArray(result.derived_sources) ? (result.derived_sources as Array<Record<string, unknown>>) : undefined,
      'menu_id',
    ),
  } satisfies SelectionRelation<never>
}

export async function setMyTeamBoundaryRoleMenus(roleId: string, menuIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}/menus`,
    data: {
      menu_ids: menuIds,
    },
  })
}

export async function fetchMyTeamBoundaryRolePackages(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}/packages`,
  })

  return {
    ids: Array.isArray(result.package_ids) ? result.package_ids.map(normalizeText).filter(Boolean) : [],
    items: Array.isArray(result.packages)
      ? result.packages.map((item) => normalizeFeaturePackage(item as Record<string, unknown>))
      : [],
    inherited: Boolean(result.inherited),
  } satisfies SelectionRelation<FeaturePackageRecord>
}

export async function fetchMyTeamActionOrigins() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/tenants/my-team/action-origins',
  })

  return normalizeOriginSummary(result, 'action_id')
}

export async function fetchMyTeamMenuOrigins() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/tenants/my-team/menu-origins',
  })

  return normalizeOriginSummary(result, 'menu_id')
}

export async function setMyTeamBoundaryRolePackages(roleId: string, packageIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/tenants/my-team/boundary/roles/${roleId}/packages`,
    data: {
      package_ids: packageIds,
    },
  })
}
