import { requestData } from '@/shared/api/client'
import type {
  FeaturePackageImpactSummary,
  FeaturePackageRecord,
  FeaturePackageSavePayload,
  PermissionActionConsumerDetail,
  PermissionActionEndpointBinding,
  PermissionActionRecord,
  PermissionActionSavePayload,
  PermissionGroupSummary,
  RelationSourceRecord,
  RoleRecord,
  RoleSavePayload,
  SelectionRelation,
  UserDiagnosisSummary,
  UserPermissionDiagnosis,
  UserRecord,
  UserRoleSummary,
  UserSavePayload,
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

function stringifyValue(value: unknown) {
  if (value === null || value === undefined || value === '') {
    return '-'
  }
  if (Array.isArray(value)) {
    return value.length ? value.map((item) => normalizeText(item)).filter(Boolean).join('、') : '-'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return normalizeText(value)
}

function normalizeDetailItems(
  input: Record<string, unknown> | null | undefined,
  labelMap: Record<string, string> = {},
) {
  if (!input) {
    return [] as Array<{ label: string; value: string }>
  }

  return Object.entries(input)
    .filter(([, value]) => value !== undefined)
    .map(([key, value]) => ({
      label: labelMap[key] || key,
      value: stringifyValue(value),
    }))
}

function normalizePermissionGroup(input?: Record<string, unknown> | null): PermissionGroupSummary | undefined {
  if (!input) {
    return undefined
  }

  return {
    id: normalizeText(input.id),
    code: normalizeText(input.code),
    name: normalizeText(input.name),
    groupType: normalizeText(input.group_type || input.groupType),
    description: normalizeText(input.description),
    status: normalizeText(input.status || 'normal'),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    isBuiltin: Boolean(input.is_builtin ?? input.isBuiltin),
  }
}

function normalizeFeaturePackage(input: Record<string, unknown>): FeaturePackageRecord {
  const packageKey = normalizeText(input.package_key || input.packageKey)
  return {
    id: normalizeText(input.id),
    packageKey,
    packageType: normalizeText(input.package_type || input.packageType || 'base'),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    contextType: normalizeText(
      input.context_type ||
        input.contextType ||
        (packageKey.startsWith('platform.') ? 'platform' : packageKey.startsWith('common.') ? 'common' : 'team'),
    ),
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

function normalizeRole(input: Record<string, unknown>): RoleRecord {
  return {
    id: normalizeText(input.roleId || input.role_id || input.id),
    code: normalizeText(input.roleCode || input.role_code || input.code),
    name: normalizeText(input.roleName || input.role_name || input.name),
    description: normalizeText(input.description),
    status: normalizeText(input.status || 'normal'),
    sortOrder: Number(input.sortOrder || input.sort_order || 0),
    priority: Number(input.priority || 0),
    canEditPermission: Boolean(input.canEditPermission ?? input.can_edit_permission ?? true),
    createdAt: normalizeText(input.createTime || input.created_at || input.createdAt),
  }
}

function normalizeUserRole(input: Record<string, unknown>): UserRoleSummary {
  return {
    id: normalizeText(input.id),
    code: normalizeText(input.code),
    name: normalizeText(input.name),
  }
}

function normalizeUser(input: Record<string, unknown>): UserRecord {
  return {
    id: normalizeText(input.id),
    userName: normalizeText(input.userName || input.user_name),
    nickName: normalizeText(input.nickName || input.nick_name),
    userEmail: normalizeText(input.userEmail || input.user_email),
    userPhone: normalizeText(input.userPhone || input.user_phone),
    status: normalizeText(input.status || 'active'),
    avatar: normalizeText(input.avatar),
    lastLoginTime: normalizeText(input.lastLoginTime || input.last_login_time),
    lastLoginIP: normalizeText(input.lastLoginIP || input.last_login_ip),
    registerSource: normalizeText(input.registerSource || input.register_source),
    invitedBy: normalizeText(input.invitedBy || input.invited_by),
    invitedByName: normalizeText(input.invitedByName || input.invited_by_name),
    systemRemark: normalizeText(input.systemRemark || input.system_remark),
    createdAt: normalizeText(input.createTime || input.created_at),
    updatedAt: normalizeText(input.updateTime || input.updated_at),
    userRoles: Array.isArray(input.userRoles || input.user_roles)
      ? ((input.userRoles || input.user_roles) as unknown[]).map(normalizeText).filter(Boolean)
      : [],
    roleDetails: Array.isArray(input.roleDetails || input.role_details)
      ? ((input.roleDetails || input.role_details) as unknown[]).map((item) => {
          const target = item as Record<string, unknown>
          return {
            code: normalizeText(target.code),
            name: normalizeText(target.name),
          }
        })
      : [],
    roles: Array.isArray(input.roles)
      ? (input.roles as unknown[]).map((item) => normalizeUserRole(item as Record<string, unknown>))
      : undefined,
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
    moduleGroupId: normalizeText(input.module_group_id || input.moduleGroupId),
    featureGroupId: normalizeText(input.feature_group_id || input.featureGroupId),
    moduleGroup: normalizePermissionGroup((input.module_group || input.moduleGroup) as Record<string, unknown>),
    featureGroup: normalizePermissionGroup((input.feature_group || input.featureGroup) as Record<string, unknown>),
    apiCount: Number(input.api_count || input.apiCount || 0),
    pageCount: Number(input.page_count || input.pageCount || 0),
    packageCount: Number(input.package_count || input.packageCount || 0),
    consumerTypes: Array.isArray(input.consumer_types || input.consumerTypes)
      ? ((input.consumer_types || input.consumerTypes) as unknown[]).map(normalizeText).filter(Boolean)
      : [],
    usagePattern: normalizeText(input.usage_pattern || input.usagePattern || 'unused'),
    usageNote: normalizeText(input.usage_note || input.usageNote),
    duplicatePattern: normalizeText(input.duplicate_pattern || input.duplicatePattern || 'none'),
    duplicateGroup: normalizeText(input.duplicate_group || input.duplicateGroup),
    duplicateKeys: Array.isArray(input.duplicate_keys || input.duplicateKeys)
      ? ((input.duplicate_keys || input.duplicateKeys) as unknown[]).map(normalizeText).filter(Boolean)
      : [],
    duplicateNote: normalizeText(input.duplicate_note || input.duplicateNote),
    status: normalizeText(input.status || 'normal'),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    isBuiltin: Boolean(input.is_builtin ?? input.isBuiltin),
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeRelationSources(
  input?: Array<Record<string, unknown>>,
  fieldKey: 'action_id' | 'menu_id' = 'action_id',
): RelationSourceRecord[] {
  if (!Array.isArray(input)) {
    return []
  }

  return input.map((item) => ({
    entityId: normalizeText(item[fieldKey]),
    packageIds: Array.isArray(item.package_ids)
      ? item.package_ids.map(normalizeText).filter(Boolean)
      : [],
  }))
}

function normalizePackageRelation(input: Record<string, unknown>): SelectionRelation<FeaturePackageRecord> {
  return {
    ids: Array.isArray(input.package_ids) ? input.package_ids.map(normalizeText).filter(Boolean) : [],
    items: Array.isArray(input.packages)
      ? input.packages.map((item) => normalizeFeaturePackage(item as Record<string, unknown>))
      : [],
    inherited: Boolean(input.inherited),
  }
}

export async function fetchRoleList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/roles',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeRole(item)),
  }
}

export async function fetchRoleDetail(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/roles/${roleId}`,
  })

  return normalizeRole(result)
}

export async function createRole(payload: RoleSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/roles',
    data: {
      name: payload.roleName,
      code: payload.roleCode,
      description: payload.description,
      sort_order: payload.sortOrder,
      priority: payload.priority,
      status: payload.status,
    },
  })

  return normalizeRole(result)
}

export async function updateRole(roleId: string, payload: RoleSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/roles/${roleId}`,
    data: {
      name: payload.roleName,
      code: payload.roleCode,
      description: payload.description,
      sort_order: payload.sortOrder,
      priority: payload.priority,
      status: payload.status,
    },
  })

  return normalizeRole(result)
}

export async function deleteRole(roleId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/roles/${roleId}`,
  })
}

export async function fetchRolePackages(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/roles/${roleId}/packages`,
  })

  return normalizePackageRelation(result)
}

export async function updateRolePackages(roleId: string, packageIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/roles/${roleId}/packages`,
    data: { package_ids: packageIds },
  })
}

export async function fetchRoleActions(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/roles/${roleId}/actions`,
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

export async function updateRoleActions(roleId: string, actionIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/roles/${roleId}/actions`,
    data: { action_ids: actionIds },
  })
}

export async function fetchRoleMenus(roleId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/roles/${roleId}/menus`,
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

export async function updateRoleMenus(roleId: string, menuIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/roles/${roleId}/menus`,
    data: { menu_ids: menuIds },
  })
}

export async function fetchUserList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/users',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeUser(item)),
  }
}

export async function fetchUserDetail(userId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/users/${userId}`,
  })

  return normalizeUser(result)
}

export async function createUser(payload: UserSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/users',
    data: {
      user_name: payload.userName,
      nick_name: payload.nickName,
      user_email: payload.userEmail,
      user_phone: payload.userPhone,
      password: payload.password,
      status: payload.status,
    },
  })

  return normalizeUser(result)
}

export async function updateUser(userId: string, payload: UserSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/users/${userId}`,
    data: {
      user_name: payload.userName,
      nick_name: payload.nickName,
      user_email: payload.userEmail,
      user_phone: payload.userPhone,
      status: payload.status,
    },
  })

  return normalizeUser(result)
}

export async function deleteUser(userId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/users/${userId}`,
  })
}

export async function fetchUserPackages(userId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/users/${userId}/packages`,
  })

  return normalizePackageRelation(result)
}

export async function updateUserPackages(userId: string, packageIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/users/${userId}/packages`,
    data: { package_ids: packageIds },
  })
}

export async function fetchUserMenus(userId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/users/${userId}/menus`,
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

export async function updateUserMenus(userId: string, menuIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/users/${userId}/menus`,
    data: { menu_ids: menuIds },
  })
}

export async function assignUserRoles(userId: string, roleIds: string[]) {
  await requestData<unknown>({
    method: 'POST',
    url: `/api/v1/users/${userId}/roles`,
    data: { role_ids: roleIds },
  })
}

export async function fetchUserPermissionDiagnosis(userId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/users/${userId}/permission-diagnosis`,
  })

  const raw = {
    context: (result.context as Record<string, unknown>) || {},
    diagnosis: (result.diagnosis as Record<string, unknown> | null) || null,
    roles: Array.isArray(result.roles) ? (result.roles as Record<string, unknown>[]) : [],
    snapshot: (result.snapshot as Record<string, unknown>) || {},
    user: (result.user as Record<string, unknown>) || {},
    teamMember:
      result.team_member && typeof result.team_member === 'object'
        ? (result.team_member as Record<string, unknown>)
        : undefined,
    teamPackages: Array.isArray(result.team_packages)
      ? (result.team_packages as Record<string, unknown>[])
      : undefined,
  } satisfies UserPermissionDiagnosis

  return {
    userItems: normalizeDetailItems(raw.user, {
      id: '用户 ID',
      user_name: '账号',
      nick_name: '昵称',
      status: '状态',
      is_super_admin: '超级管理员',
    }),
    contextItems: normalizeDetailItems(raw.context, {
      type: '上下文类型',
      tenant_id: '团队 ID',
      tenant_name: '团队名称',
    }),
    snapshotItems: normalizeDetailItems(raw.snapshot, {
      action_count: '动作数',
      package_count: '功能包数',
      role_count: '角色数',
      menu_count: '菜单数',
      team_id: '团队 ID',
    }),
    diagnosisItems: normalizeDetailItems(raw.diagnosis, {
      permission_key: '权限键',
      allowed: '是否允许',
      reason_text: '诊断结论',
      denial_stage: '拒绝阶段',
      denial_reason: '拒绝原因',
      boundary_state: '边界状态',
      blocked_by_team: '是否被团队阻断',
      matched_in_snapshot: '是否命中快照',
      member_status: '成员状态',
      member_matched: '成员命中',
      role_chain_matched: '角色链命中',
      role_chain_disabled: '角色链禁用',
      role_chain_available: '角色链可用',
    }),
    roleSummaries: raw.roles.map((item, index) => ({
      title: normalizeText(item.role_name || item.name || item.code) || `角色链 ${index + 1}`,
      items: normalizeDetailItems(item, {
        role_name: '角色名称',
        role_code: '角色编码',
        matched: '是否命中',
        disabled: '是否禁用',
        available: '是否可用',
        reason_text: '说明',
      }),
    })),
    sourcePackageItems: Array.isArray(raw.diagnosis?.source_packages)
      ? (raw.diagnosis?.source_packages as Record<string, unknown>[]).map((item) => ({
          label: normalizeText(item.name || item.package_key || item.id),
          value: normalizeText(item.package_key || item.id),
        }))
      : [],
    raw,
  } satisfies UserDiagnosisSummary
}

export async function refreshUserPermissionSnapshot(userId: string) {
  return requestData<Record<string, unknown>>({
    method: 'POST',
    url: `/api/v1/users/${userId}/permission-refresh`,
  })
}

export async function fetchPermissionActionList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/permission-actions',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizePermissionAction(item)),
  }
}

export async function fetchPermissionGroups() {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: '/api/v1/permission-actions/groups',
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records)
      ? result.records.map((item) => normalizePermissionGroup(item)).filter(Boolean) as PermissionGroupSummary[]
      : [],
  }
}

export async function createPermissionGroup(payload: Partial<PermissionGroupSummary>) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/permission-actions/groups',
    data: {
      code: payload.code,
      name: payload.name,
      description: payload.description,
      sort_order: payload.sortOrder,
      status: payload.status,
      group_type: payload.groupType,
    },
  })

  return normalizePermissionGroup(result)!
}

export async function updatePermissionGroup(groupId: string, payload: Partial<PermissionGroupSummary>) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/permission-actions/groups/${groupId}`,
    data: {
      code: payload.code,
      name: payload.name,
      description: payload.description,
      sort_order: payload.sortOrder,
      status: payload.status,
      group_type: payload.groupType,
    },
  })

  return normalizePermissionGroup(result)!
}

export async function fetchPermissionActionDetail(actionId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/permission-actions/${actionId}`,
  })

  return normalizePermissionAction(result)
}

export async function fetchPermissionActionEndpoints(actionId: string) {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: `/api/v1/permission-actions/${actionId}/endpoints`,
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records)
      ? result.records.map((item) => ({
          endpointCode: normalizeText(item.code || item.endpoint_code),
          method: normalizeText(item.method),
          path: normalizeText(item.path),
          summary: normalizeText(item.summary),
          authMode: normalizeText(item.auth_mode || item.authMode),
        }))
      : [],
  } satisfies { total: number; records: PermissionActionEndpointBinding[] }
}

export async function addPermissionActionEndpoint(actionId: string, endpointCode: string) {
  await requestData<unknown>({
    method: 'POST',
    url: `/api/v1/permission-actions/${actionId}/endpoints`,
    data: { endpoint_code: endpointCode },
  })
}

export async function removePermissionActionEndpoint(actionId: string, endpointCode: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/permission-actions/${actionId}/endpoints/${endpointCode}`,
  })
}

export async function fetchPermissionActionConsumers(actionId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/permission-actions/${actionId}/consumers`,
  })

  const records = Array.isArray(result.records)
    ? (result.records as Array<Record<string, unknown>>).map((item) => ({
        type: normalizeText(item.type),
        id: normalizeText(item.id),
        label: normalizeText(item.label || item.name),
        description: normalizeText(item.description),
      }))
    : []

  return {
    records,
    raw: result,
  } satisfies { records: PermissionActionConsumerDetail[]; raw: Record<string, unknown> }
}

export async function createPermissionAction(payload: PermissionActionSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/permission-actions',
    data: {
      permission_key: payload.permissionKey,
      name: payload.name,
      description: payload.description,
      module_group_id: payload.moduleGroupId || undefined,
      feature_group_id: payload.featureGroupId || undefined,
      context_type: payload.contextType,
      feature_kind: payload.featureKind,
      sort_order: payload.sortOrder,
      status: payload.status,
    },
  })

  return normalizePermissionAction(result)
}

export async function updatePermissionAction(actionId: string, payload: PermissionActionSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/permission-actions/${actionId}`,
    data: {
      permission_key: payload.permissionKey,
      name: payload.name,
      description: payload.description,
      module_group_id: payload.moduleGroupId || undefined,
      feature_group_id: payload.featureGroupId || undefined,
      context_type: payload.contextType,
      feature_kind: payload.featureKind,
      sort_order: payload.sortOrder,
      status: payload.status,
    },
  })

  return normalizePermissionAction(result)
}

export async function deletePermissionAction(actionId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/permission-actions/${actionId}`,
  })
}

export async function fetchFeaturePackageList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/feature-packages',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeFeaturePackage(item)),
  }
}

export async function fetchFeaturePackageDetail(packageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/feature-packages/${packageId}`,
  })

  return normalizeFeaturePackage(result)
}

export async function createFeaturePackage(payload: FeaturePackageSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/feature-packages',
    data: {
      package_key: payload.packageKey,
      package_type: payload.packageType,
      name: payload.name,
      description: payload.description,
      context_type: payload.contextType,
      sort_order: payload.sortOrder,
      status: payload.status,
    },
  })

  return normalizeFeaturePackage(result)
}

export async function updateFeaturePackage(packageId: string, payload: FeaturePackageSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/feature-packages/${packageId}`,
    data: {
      package_key: payload.packageKey,
      package_type: payload.packageType,
      name: payload.name,
      description: payload.description,
      context_type: payload.contextType,
      sort_order: payload.sortOrder,
      status: payload.status,
    },
  })

  return normalizeFeaturePackage(result)
}

export async function deleteFeaturePackage(packageId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/feature-packages/${packageId}`,
  })
}

export async function fetchFeaturePackageChildren(packageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/feature-packages/${packageId}/children`,
  })

  return {
    ids: Array.isArray(result.package_ids) ? result.package_ids.map(normalizeText).filter(Boolean) : [],
    items: Array.isArray(result.packages)
      ? result.packages.map((item) => normalizeFeaturePackage(item as Record<string, unknown>))
      : [],
    inherited: Boolean(result.inherited),
  } satisfies SelectionRelation<FeaturePackageRecord>
}

export async function updateFeaturePackageChildren(packageId: string, packageIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/feature-packages/${packageId}/children`,
    data: { package_ids: packageIds },
  })
}

export async function fetchFeaturePackageActions(packageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/feature-packages/${packageId}/actions`,
  })

  return {
    ids: Array.isArray(result.action_ids) ? result.action_ids.map(normalizeText).filter(Boolean) : [],
    items: Array.isArray(result.actions)
      ? result.actions.map((item) => normalizePermissionAction(item as Record<string, unknown>))
      : [],
  } satisfies SelectionRelation<PermissionActionRecord>
}

export async function updateFeaturePackageActions(packageId: string, actionIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/feature-packages/${packageId}/actions`,
    data: { action_ids: actionIds },
  })
}

export async function fetchFeaturePackageMenus(packageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/feature-packages/${packageId}/menus`,
  })

  return {
    ids: Array.isArray(result.menu_ids) ? result.menu_ids.map(normalizeText).filter(Boolean) : [],
    items: [],
  } satisfies SelectionRelation<never>
}

export async function updateFeaturePackageMenus(packageId: string, menuIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/feature-packages/${packageId}/menus`,
    data: { menu_ids: menuIds },
  })
}

export async function fetchFeaturePackageTeams(packageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/feature-packages/${packageId}/teams`,
  })

  return {
    ids: Array.isArray(result.team_ids) ? result.team_ids.map(normalizeText).filter(Boolean) : [],
    items: Array.isArray(result.teams)
      ? result.teams.map((item) => ({
          id: normalizeText((item as Record<string, unknown>).id),
          name: normalizeText((item as Record<string, unknown>).name),
          remark: normalizeText((item as Record<string, unknown>).remark),
          logoUrl: normalizeText((item as Record<string, unknown>).logo_url),
          plan: normalizeText((item as Record<string, unknown>).plan || 'free'),
          maxMembers: Number((item as Record<string, unknown>).max_members || 0),
          status: normalizeText((item as Record<string, unknown>).status || 'active'),
          ownerId: normalizeText((item as Record<string, unknown>).owner_id),
          createdAt: normalizeText((item as Record<string, unknown>).created_at),
          updatedAt: normalizeText((item as Record<string, unknown>).updated_at),
          adminUsers: [],
        }))
      : [],
  }
}

export async function updateFeaturePackageTeams(packageId: string, teamIds: string[]) {
  await requestData<unknown>({
    method: 'PUT',
    url: `/api/v1/feature-packages/${packageId}/teams`,
    data: { team_ids: teamIds },
  })
}

export async function fetchFeaturePackageRelationTree() {
  return requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/feature-packages/relationship-tree',
  })
}

export async function fetchFeaturePackageImpactPreview(packageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/feature-packages/${packageId}/impact-preview`,
  })

  const roleCount = Number(result.role_count || result.roleCount || 0)
  const teamCount = Number(result.team_count || result.teamCount || 0)
  const userCount = Number(result.user_count || result.userCount || 0)
  const menuCount = Number(result.menu_count || result.menuCount || 0)
  const actionCount = Number(result.action_count || result.actionCount || 0)

  return {
    packageId: normalizeText(result.package_id || result.packageId),
    roleCount,
    teamCount,
    userCount,
    menuCount,
    actionCount,
    metrics: [
      { id: 'roleCount', label: '角色影响', value: `${roleCount}`, tone: 'brand' },
      { id: 'teamCount', label: '团队影响', value: `${teamCount}`, tone: 'warning' },
      { id: 'userCount', label: '用户影响', value: `${userCount}`, tone: 'success' },
      { id: 'menuCount', label: '菜单数', value: `${menuCount}`, tone: 'neutral' },
      { id: 'actionCount', label: '动作数', value: `${actionCount}`, tone: 'neutral' },
    ],
  } satisfies FeaturePackageImpactSummary
}
