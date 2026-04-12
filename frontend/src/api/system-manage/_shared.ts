import { v5Client } from '@/api/v5/client'
import type { components } from '@/api/v5/schema'
import { AppRouteRecord } from '@/types/router'
import type { FastEnterApplication, FastEnterQuickLink } from '@/types/config'
import { normalizeMenuSpaceKey } from '@/utils/navigation/menu-space'
import { HttpError, showError } from '@/utils/http/error'
import { ApiStatus } from '@/utils/http/status'
import { $t } from '@/locales'
import { useUserStore } from '@/store/modules/user'
export type { V5Path, V5Method, V5PathParams, V5Query, V5RequestBody } from '@/api/v5/types'

export { v5Client, normalizeMenuSpaceKey }
export type { AppRouteRecord, FastEnterApplication, FastEnterQuickLink }

type V5UserPermissionDiagnosisResponse = components['schemas']['UserPermissionDiagnosisResponse']
type V5UserPermissionDiagnosisResult = components['schemas']['UserPermissionDiagnosisResult']
type V5UserPermissionDiagnosisAction = components['schemas']['UserPermissionDiagnosisAction']
type V5UserPermissionRoleResult = components['schemas']['UserPermissionRoleResult']
type V5UserPermissionMenuTreeItem = components['schemas']['UserPermissionMenuTreeItem']
type V5PermissionBatchTemplateItem = components['schemas']['PermissionActionBatchTemplateItem']
type V5RiskAuditItem = components['schemas']['RiskAuditItem']
type V5RiskAuditSummary = components['schemas']['RiskAuditSummary']
type V5PermissionBatchTemplatePayload = components['schemas']['PermissionActionBatchTemplatePayload']
type V5PageMenuOptionLike = components['schemas']['PageMenuOptionItem']
type V5PageUnregisteredLike = components['schemas']['PageUnregisteredItem']
type V5PageBreadcrumbPreviewLike = components['schemas']['PageBreadcrumbPreviewItem']
type V5PageAccessTraceResultLike = components['schemas']['PageAccessTraceResponse']
type V5RefreshStatsLike = components['schemas']['RefreshStats']
type V5PermissionActionConsumerDetailsLike = components['schemas']['PermissionActionConsumersResponse']
type V5PermissionAuditSummaryLike = {
  total_count?: number
  unused_count?: number
  api_only_count?: number
  page_only_count?: number
  package_only_count?: number
  multi_consumer_count?: number
  cross_context_mirror_count?: number
  suspected_duplicate_count?: number
}
type V5PageLike = {
  id?: string
  app_key?: string
  page_key?: string
  name?: string
  route_name?: string
  route_path?: string
  component?: string
  page_type?: string
  visibility_scope?: string
  source?: string
  module_key?: string
  sort_order?: number
  parent_menu_id?: string
  parent_menu_name?: string
  parent_page_key?: string
  parent_page_name?: string
  display_group_key?: string
  display_group_name?: string
  active_menu_path?: string
  breadcrumb_mode?: string
  access_mode?: string
  permission_key?: string
  inherit_permission?: boolean
  keep_alive?: boolean
  is_full_page?: boolean
  is_iframe?: boolean
  is_hide_tab?: boolean
  space_keys?: string[]
  page_space_bindings?: Array<{
    space_key?: string
    source?: string
  }>
  space_scope?: string
  space_type?: string
  host_key?: string
  status?: string
  remote_binding?: {
    manifest_url?: string
    remote_app_key?: string
    remote_page_key?: string
    remote_entry_url?: string
    remote_route_path?: string
    remote_module?: string
    remote_module_name?: string
    remote_url?: string
    runtime_version?: string
    health_check_url?: string
  }
  meta?: {
    spaceKeys?: string[]
    spaceScope?: string
    visibilityScope?: string
    link?: string
    isIframe?: boolean
    isHideTab?: boolean
    requiredAction?: string
    requiredActions?: string[]
    actionMatchMode?: string
    actionVisibilityMode?: string
    customParent?: string
    breadcrumbChain?: string[]
    hostKey?: string
    spaceType?: string
    manifest_url?: string
    manifestUrl?: string
    remote_app_key?: string
    remoteAppKey?: string
    remote_page_key?: string
    remotePageKey?: string
    remote_entry_url?: string
    remoteEntryUrl?: string
    remote_route_path?: string
    remoteRoutePath?: string
    remote_module?: string
    remoteModule?: string
    remote_module_name?: string
    remoteModuleName?: string
    remote_url?: string
    remoteUrl?: string
    runtime_version?: string
    runtimeVersion?: string
    version?: string
    health_check_url?: string
    healthCheckUrl?: string
  }
  created_at?: string
  updated_at?: string
}
type V5PermissionActionLike = {
  id?: string
  resource_code?: string
  action_code?: string
  module_code?: string
  action_key?: string
  permission_key?: string
  name?: string
  description?: string | null
  status?: string | null
  group_id?: string | null
  group_name?: string | null
  module_group_id?: string | null
  feature_group_id?: string | null
  module_group?: { id?: string; code?: string; name?: string; name_en?: string } | null
  feature_group?: { id?: string; code?: string; name?: string; name_en?: string } | null
  consumer_types?: string[]
  duplicate_keys?: string[]
  data_policy?: string
  data_permission_code?: string
  data_permission_name?: string
  api_count?: number
  page_count?: number
  package_count?: number
  usage_pattern?: string
  usage_note?: string
  duplicate_pattern?: string
  duplicate_group?: string
  duplicate_note?: string
  sort_order?: number
  is_builtin?: boolean
  created_at?: string | null
  updated_at?: string | null
}

let unauthorizedHandling: Promise<void> | null = null

// V5 真相源：HTTP status + spec error.code/error.message。
// 旧 axios 错误协议（statusCode、msg、error 字段）已废弃。
function normalizeV5StatusCode(status?: number, error?: any): number {
  const responseStatus = Number(status || 0)
  if (Number.isFinite(responseStatus) && responseStatus > 0) {
    return responseStatus
  }
  const specCode = Number(error?.code || 0)
  if (Number.isFinite(specCode) && specCode > 0) {
    return specCode
  }
  return ApiStatus.error
}

function normalizeV5ErrorMessage(error: any, statusCode: number): string {
  const backendMessage = `${error?.message || ''}`.trim()
  if (backendMessage) {
    return backendMessage
  }
  if (statusCode === ApiStatus.unauthorized) {
    return $t('httpMsg.unauthorized')
  }
  if (statusCode >= 500) {
    return $t('httpMsg.internalServerError')
  }
  return $t('httpMsg.requestFailed')
}

function handleV5Unauthorized(error: HttpError): void {
  if (!unauthorizedHandling) {
    unauthorizedHandling = (async () => {
      try {
        showError(error, true)
        useUserStore().logOut()
      } finally {
        setTimeout(() => {
          unauthorizedHandling = null
        }, 3000)
      }
    })()
  }
}

export function createV5HttpError(error: any, response?: Response): HttpError {
  const statusCode = normalizeV5StatusCode(response?.status, error)
  const httpError = new HttpError(normalizeV5ErrorMessage(error, statusCode), statusCode, {
    data: error,
    url: response?.url,
    method: undefined
  })

  if (statusCode === ApiStatus.unauthorized) {
    handleV5Unauthorized(httpError)
  }

  return httpError
}

/**
 * 解包 openapi-fetch 的 `{ data, error }` 返回，行为对齐 legacy `request<T>`：
 * - error 非空 → 抛出
 * - data 为 undefined → 抛出（HTTP 200 但响应体异常）
 * - 否则返回 data
 */
export async function unwrap<T>(
  promise: Promise<{ data?: T; error?: any; response: Response }>
): Promise<T> {
  const { data, error, response } = await promise
  if (error) throw createV5HttpError(error, response)
  if (data === undefined) throw new Error('v5Client: empty response')
  return data as T
}

export type V5AnyRecord = Record<string, unknown>

export type V5BodyShape<T extends object> = {
  [K in keyof T]: T[K]
}

export function toV5Body<T extends object>(value: T): V5BodyShape<T> {
  return { ...value }
}

export function toV5StringArray(value: unknown): string[] {
  if (!Array.isArray(value)) return []
  return value.map((item) => `${item || ''}`.trim()).filter(Boolean)
}

// Phase 4 slice 5: 用户域 list/get/create/update/delete/assignRoles 走 v5Client。
// ogen handler 返回 snake_case schema 原型，这里负责 camelCase 归一，
// 保持 Api.SystemManage.UserListItem 等消费方契约不变。
export function normalizeUserSummary(item: any): Api.SystemManage.UserListItem {
  const roleDetails = Array.isArray(item?.role_details)
    ? item.role_details.map((r: any) => ({
        code: r?.code || '',
        name: r?.name || ''
      }))
    : []
  return {
    id: item?.id || '',
    avatar: item?.avatar_url || '',
    status: item?.status || 'inactive',
    userName: item?.user_name || '',
    nickName: item?.nick_name || '',
    userPhone: item?.user_phone || '',
    userEmail: item?.user_email || '',
    systemRemark: item?.system_remark || '',
    lastLoginTime: item?.last_login_time || '',
    lastLoginIP: item?.last_login_ip || '',
    userRoles: Array.isArray(item?.user_roles)
      ? item.user_roles
      : Array.isArray(item?.userRoles)
        ? item.userRoles
        : [],
    roleDetails,
    registerSource: item?.register_source || '',
    invitedBy: item?.invited_by || '',
    invitedByName: item?.invited_by_name || '',
    createTime: item?.create_time || '',
    updateTime: item?.update_time || ''
  }
}


export function normalizePermissionKey(value?: string) {
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

export function derivePermissionSegments(permissionKey?: string) {
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

export function deriveContextType(permissionKey?: string, moduleCode?: string) {
  const key = `${permissionKey || ''}`.trim()
  const module = `${moduleCode || ''}`.trim()
  if (
    key === 'collaboration_workspace.manage' ||
    key.startsWith('collaboration_workspace.member.') ||
    key.startsWith('collaboration_workspace.boundary.') ||
    key.startsWith('collaboration_workspace.message.') ||
    module === 'collaboration_workspace' ||
    module === 'collaboration_workspace_member' ||
    module === 'collaboration_workspace_boundary' ||
    module === 'collaboration_workspace_message'
  ) {
    return 'collaboration'
  }
  if (
    key.startsWith('personal.') ||
    module === 'personal' ||
    module === 'personal_workspace'
  ) {
    return 'personal'
  }
  if (
    key.startsWith('system.') ||
    key.startsWith('feature_package.') ||
    key.startsWith('message.') ||
    module === 'role' ||
    module === 'user' ||
    module === 'menu' ||
    module === 'menu_backup' ||
    module === 'permission_action' ||
    module === 'permission_key' ||
    module === 'api_endpoint' ||
    module === 'feature_package' ||
    module === 'page' ||
    module === 'collaboration_workspace_member_admin'
  ) {
    return 'common'
  }
  return 'common'
}

export function normalizePermissionGroup(value: any): Api.SystemManage.PermissionGroupItem | undefined {
  if (!value) return undefined
  return {
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
}

export function normalizePermissionAction(item: V5PermissionActionLike): Api.SystemManage.PermissionActionItem {
  const permissionKey = normalizePermissionKey(item?.permission_key || item?.action_key)
  const legacy = derivePermissionSegments(permissionKey)
  const consumerTypes = item?.consumer_types || []
  const duplicateKeys = item?.duplicate_keys || []
  const moduleGroup = normalizePermissionGroup(item?.module_group)
  const featureGroup = normalizePermissionGroup(item?.feature_group)
  return {
    id: item?.id || '',
    resourceCode: legacy.resourceCode,
    actionCode: legacy.actionCode,
    moduleCode: moduleGroup?.code || legacy.resourceCode || '',
    moduleGroupId: item?.module_group_id || moduleGroup?.id || '',
    featureGroupId: item?.feature_group_id || featureGroup?.id || '',
    moduleGroup,
    featureGroup,
    permissionKey,
    featureKind: featureGroup?.code || 'system',
    dataPolicy: item?.data_policy || '',
    name: item?.name || '',
    description: item?.description || '',
    dataPermissionCode: item?.data_permission_code || '',
    dataPermissionName: item?.data_permission_name || '',
    apiCount: Number(item?.api_count ?? 0),
    pageCount: Number(item?.page_count ?? 0),
    packageCount: Number(item?.package_count ?? 0),
    consumerTypes: Array.isArray(consumerTypes)
      ? consumerTypes.map((value) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    usagePattern: item?.usage_pattern || 'unused',
    usageNote: item?.usage_note || '',
    duplicatePattern: item?.duplicate_pattern || 'none',
    duplicateGroup: item?.duplicate_group || '',
    duplicateKeys: Array.isArray(duplicateKeys)
      ? duplicateKeys.map((value) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    duplicateNote: item?.duplicate_note || '',
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? 0,
    isBuiltin: Boolean(item?.is_builtin ?? false),
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizePermissionAuditSummary(
  item: V5PermissionAuditSummaryLike | undefined
): Api.SystemManage.PermissionActionAuditSummary {
  return {
    totalCount: Number(item?.total_count ?? 0),
    unusedCount: Number(item?.unused_count ?? 0),
    apiOnlyCount: Number(item?.api_only_count ?? 0),
    pageOnlyCount: Number(item?.page_only_count ?? 0),
    packageOnlyCount: Number(item?.package_only_count ?? 0),
    multiConsumerCount: Number(item?.multi_consumer_count ?? 0),
    crossContextMirrorCount: Number(
      item?.cross_context_mirror_count ?? 0
    ),
    suspectedDuplicateCount: Number(
      item?.suspected_duplicate_count ?? 0
    )
  }
}

export function normalizeApiEndpoint(item: any): Api.SystemManage.APIEndpointItem {
  const permissionKeysRaw = item?.permission_keys || []
  const permissionContextsRaw = item?.permission_contexts || []
  const permissionKeys = Array.isArray(permissionKeysRaw)
    ? permissionKeysRaw.map((v: any) => `${v || ''}`.trim()).filter(Boolean)
    : []
  const permissionKey =
    normalizePermissionKey(item?.permission_key) || permissionKeys[0] || ''
  return {
    id: item?.id || '',
    code: item?.code || '',
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || '',
    handler: item?.handler || '',
    summary: item?.summary || '',
    permissionKey,
    permissionKeys,
    permissionContexts: Array.isArray(permissionContextsRaw)
      ? permissionContextsRaw.map((v: any) => `${v || ''}`.trim()).filter(Boolean)
      : [],
    permissionBindingMode:
      item?.permission_binding_mode ||
      (permissionKeys.length > 1 ? 'shared' : permissionKeys.length === 1 ? 'single' : 'none'),
    sharedAcrossContexts: Boolean(item?.shared_across_contexts),
    permissionNote: item?.permission_note || '',
    authMode:
      item?.auth_mode ||
      item?.access_mode ||
      (permissionKey ? 'permission' : 'jwt'),
    categoryId: item?.category_id || item?.category?.id || '',
    category: item?.category
      ? {
          id: item.category.id || '',
          code: item.category.code || '',
          name: item.category.name || '',
          nameEn: item.category.name_en || item.category.nameEn || '',
          sortOrder: item.category.sort_order ?? item.category.sortOrder ?? 0,
          status: item.category.status || 'normal'
        }
      : undefined,
    dataPermissionCode: item?.data_permission_code || '',
    dataPermissionName: item?.data_permission_name || '',
    runtimeExists: Boolean(item?.runtime_exists),
    stale: Boolean(item?.stale),
    staleReason: item?.stale_reason || '',
    status: item?.status || 'normal',
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeFeaturePackage(item: any): Api.SystemManage.FeaturePackageItem {
  const packageKey = item?.package_key || ''
  const workspaceScope = item?.workspace_scope || 'all'
  const appKeysRaw = item?.app_keys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    appKeys,
    packageKey,
    packageType: item?.package_type || 'base',
    name: item?.name || '',
    description: item?.description || '',
    workspaceScope,
    isBuiltin: Boolean(item?.is_builtin ?? false),
    actionCount: item?.action_count ?? 0,
    menuCount: item?.menu_count ?? 0,
    collaborationWorkspaceCount:
      item?.collaboration_workspace_count ?? 0,
    status: item?.status || 'normal',
    sortOrder: item?.sort_order ?? 0,
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeRole(item: any): Api.SystemManage.RoleListItem {
  const appKeysRaw = item?.app_keys || item?.appKeys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    roleId: item?.role_id || item?.roleId || item?.id || '',
    roleName: item?.role_name || item?.roleName || item?.name || '',
    roleCode: item?.role_code || item?.roleCode || item?.code || '',
    description: item?.description || item?.remark || '',
    appKeys,
    sortOrder: item?.sort_order ?? item?.sortOrder ?? 0,
    status: item?.status || 'normal',
    priority: item?.priority ?? 0,
    customParams: item?.custom_params || item?.customParams || {},
    createTime: item?.create_time || item?.createTime || '',
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || null,
    isGlobal: Boolean(item?.is_global ?? item?.isGlobal ?? appKeys.length === 0),
    canEditPermission: Boolean(item?.can_edit_permission ?? item?.canEditPermission ?? true)
  }
}

export function normalizeCollaborationWorkspace(
  item: any
): Api.SystemManage.CollaborationWorkspaceListItem {
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
    ownerId: item?.owner_id || '',
    adminUsers: item?.admin_users || [],
    adminUserIds: item?.admin_user_ids || [],
    currentRoleCode: item?.current_role_code || '',
    memberStatus: item?.member_status || ''
  }
}

export function normalizePageItem(item: V5PageLike | undefined): Api.SystemManage.PageItem {
  const meta = item?.meta || {}
  const rawRemoteBinding = item?.remote_binding || {}
  const remoteBinding = {
    manifestUrl: `${rawRemoteBinding?.manifest_url || meta?.manifest_url || meta?.manifestUrl || ''}`.trim(),
    remoteAppKey: `${rawRemoteBinding?.remote_app_key || meta?.remote_app_key || meta?.remoteAppKey || ''}`.trim(),
    remotePageKey: `${rawRemoteBinding?.remote_page_key || meta?.remote_page_key || meta?.remotePageKey || ''}`.trim(),
    remoteEntryUrl: `${rawRemoteBinding?.remote_entry_url || meta?.remote_entry_url || meta?.remoteEntryUrl || ''}`.trim(),
    remoteRoutePath: `${rawRemoteBinding?.remote_route_path || meta?.remote_route_path || meta?.remoteRoutePath || ''}`.trim(),
    remoteModule: `${rawRemoteBinding?.remote_module || meta?.remote_module || meta?.remoteModule || ''}`.trim(),
    remoteModuleName: `${rawRemoteBinding?.remote_module_name || meta?.remote_module_name || meta?.remoteModuleName || ''}`.trim(),
    remoteUrl: `${rawRemoteBinding?.remote_url || meta?.remote_url || meta?.remoteUrl || ''}`.trim(),
    runtimeVersion: `${rawRemoteBinding?.runtime_version || meta?.runtime_version || meta?.runtimeVersion || meta?.version || ''}`.trim(),
    healthCheckUrl: `${rawRemoteBinding?.health_check_url || meta?.health_check_url || meta?.healthCheckUrl || ''}`.trim()
  }
  const hasRemoteBinding = Object.values(remoteBinding).some(Boolean)
  const rawVisibilityScope =
    `${item?.visibility_scope || meta?.visibilityScope || ''}`.trim()
  const rawSpaceKeysSource = item?.space_keys || meta?.spaceKeys
  const rawSpaceKeys: string[] = Array.isArray(rawSpaceKeysSource) ? rawSpaceKeysSource : []
  // spaceKeys 由后端把 page_space_bindings、菜单继承和父页继承统一编译后下发，前端不再自行猜测。
  const spaceKeys = rawSpaceKeys
    .map((value) => normalizeMenuSpaceKey(`${value || ''}`))
    .filter(Boolean)
  const pageSpaceBindings = Array.isArray(item?.page_space_bindings)
    ? item.page_space_bindings
        .map((binding) => ({
          spaceKey: normalizeMenuSpaceKey(`${binding?.space_key || ''}`),
          source: `${binding?.source || ''}`.trim() || undefined
        }))
        .filter((binding) => binding.spaceKey)
    : []
  const spaceType = `${item?.space_type || meta?.spaceType || ''}`.trim()
  const hostKey = `${item?.host_key || meta?.hostKey || ''}`.trim()
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    pageKey: item?.page_key || '',
    name: item?.name || '',
    routeName: item?.route_name || '',
    routePath: item?.route_path || '',
    component: item?.component || '',
    pageType: item?.page_type || 'inner',
    visibilityScope: rawVisibilityScope || undefined,
    source: item?.source || 'manual',
    moduleKey: item?.module_key || '',
    sortOrder: item?.sort_order ?? 0,
    parentMenuId: item?.parent_menu_id || '',
    parentMenuName: item?.parent_menu_name || '',
    parentPageKey: item?.parent_page_key || '',
    parentPageName: item?.parent_page_name || '',
    displayGroupKey: item?.display_group_key || '',
    displayGroupName: item?.display_group_name || '',
    activeMenuPath: item?.active_menu_path || '',
    breadcrumbMode: item?.breadcrumb_mode || 'inherit_menu',
    accessMode: item?.access_mode || 'inherit',
    permissionKey: item?.permission_key || '',
    inheritPermission: Boolean(item?.inherit_permission ?? true),
    keepAlive: Boolean(item?.keep_alive ?? false),
    isFullPage: Boolean(item?.is_full_page ?? false),
    isIframe: Boolean(meta?.isIframe ?? item?.is_iframe ?? false),
    isHideTab: Boolean(meta?.isHideTab ?? item?.is_hide_tab ?? false),
    link: `${meta?.link || ''}`.trim(),
    remoteBinding: hasRemoteBinding ? remoteBinding : undefined,
    spaceKeys,
    pageSpaceBindings,
    spaceScope:
      `${item?.space_scope || meta?.spaceScope || ''}`.trim() || undefined,
    spaceType,
    hostKey,
    status: item?.status || 'normal',
    meta: {
      ...meta,
      ...(spaceKeys.length ? { spaceKeys } : {}),
      ...(rawVisibilityScope ? { visibilityScope: rawVisibilityScope } : {}),
      ...(`${item?.space_scope || meta?.spaceScope || ''}`.trim()
        ? {
            spaceScope: `${item?.space_scope || meta?.spaceScope || ''}`.trim()
          }
        : {}),
      ...(spaceType ? { spaceType } : {}),
      ...(hostKey ? { hostKey } : {})
    },
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizePageMenuOption(
  item: V5PageMenuOptionLike | undefined
): Api.SystemManage.PageMenuOptionItem {
  return {
    id: item?.id || '',
    name: item?.name || '',
    title: item?.title || '',
    path: item?.path || '',
    children: Array.isArray(item?.children) ? item.children.map(normalizePageMenuOption) : []
  }
}

export function normalizePageUnregisteredItem(
  item: V5PageUnregisteredLike | undefined
): Api.SystemManage.PageUnregisteredItem {
  return {
    filePath: item?.file_path || '',
    component: item?.component || '',
    pageKey: item?.page_key || '',
    name: item?.name || '',
    routeName: item?.route_name || '',
    routePath: item?.route_path || '',
    pageType: item?.page_type || 'inner',
    moduleKey: item?.module_key || '',
    parentMenuId: item?.parent_menu_id || '',
    parentMenuName: item?.parent_menu_name || '',
    activeMenuPath: item?.active_menu_path || ''
  }
}

export function normalizeMenuSpace(item: any): Api.SystemManage.MenuSpaceItem {
  const allowedRoleCodes = item?.allowed_role_codes ?? []
  const rawAccessMode =
    item?.access_mode ||
    item?.meta?.access_mode ||
    'all'
  const accessMode = `${rawAccessMode}`.trim()
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    spaceKey: normalizeMenuSpaceKey(item?.space_key),
    name: item?.name || '',
    description: item?.description || '',
    defaultHomePath: item?.default_home_path || '',
    isDefault: Boolean(item?.is_default ?? false),
    status: item?.status || 'normal',
    hostCount: item?.host_count ?? 0,
    hosts: Array.isArray(item?.hosts)
      ? item.hosts.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    menuCount: Number(item?.menu_count ?? 0),
    pageCount: Number(item?.page_count ?? 0),
    accessMode,
    allowedRoleCodes: Array.isArray(allowedRoleCodes)
      ? allowedRoleCodes.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    meta: item?.meta || {},
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeMenuSpaceHostBinding(item: any): Api.SystemManage.MenuSpaceHostBindingItem {
  const meta = item?.meta || {}
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    appName: item?.app_name || '',
    host: `${item?.host || ''}`.trim(),
    spaceKey: normalizeMenuSpaceKey(item?.space_key),
    spaceName: item?.space_name || '',
    description: item?.description || '',
    isDefault: Boolean(item?.is_default ?? false),
    status: item?.status || 'normal',
    scheme: `${item?.scheme || item?.meta?.scheme || meta?.scheme || ''}`.trim(),
    routePrefix:
      `${item?.route_prefix || meta?.route_prefix || meta?.routePrefix || ''}`.trim(),
    authMode:
      `${item?.auth_mode || meta?.auth_mode || meta?.authMode || 'inherit_host'}`.trim() ||
      'inherit_host',
    loginHost:
      `${item?.login_host || meta?.login_host || meta?.loginHost || ''}`.trim(),
    callbackHost:
      `${item?.callback_host || meta?.callback_host || meta?.callbackHost || ''}`.trim(),
    cookieScopeMode:
      `${item?.cookie_scope_mode || meta?.cookie_scope_mode || meta?.cookieScopeMode || 'inherit'}`.trim() ||
      'inherit',
    cookieDomain:
      `${item?.cookie_domain || meta?.cookie_domain || meta?.cookieDomain || ''}`.trim(),
    meta,
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeApp(item: any): Api.SystemManage.AppItem {
  const primaryHostsRaw = item?.primary_hosts || []
  const primaryHosts = Array.isArray(primaryHostsRaw)
    ? primaryHostsRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  const capabilities =
    item?.capabilities && typeof item.capabilities === 'object' && !Array.isArray(item.capabilities)
      ? item.capabilities
      : {}
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    name: item?.name || '',
    description: item?.description || '',
    defaultSpaceKey: item?.default_space_key || '',
    spaceMode: item?.space_mode || 'single',
    authMode: item?.auth_mode || 'inherit_host',
    frontendEntryUrl: `${item?.frontend_entry_url || ''}`.trim(),
    backendEntryUrl: `${item?.backend_entry_url || ''}`.trim(),
    healthCheckUrl: `${item?.health_check_url || ''}`.trim(),
    manifestUrl: `${item?.manifest_url || item?.meta?.manifest_url || item?.meta?.manifestUrl || ''}`.trim(),
    runtimeVersion: `${item?.runtime_version || item?.meta?.runtime_version || item?.meta?.runtimeVersion || item?.meta?.version || ''}`.trim(),
    probeStatus: `${item?.probe_status || ''}`.trim() || undefined,
    probeTarget: `${item?.probe_target || ''}`.trim() || undefined,
    probeMessage: `${item?.probe_message || ''}`.trim() || undefined,
    probeCheckedAt: `${item?.probe_checked_at || ''}`.trim() || undefined,
    capabilities,
    isDefault: Boolean(item?.is_default ?? false),
    status: item?.status || 'normal',
    hostCount: Number(item?.host_count ?? 0),
    primaryHost: item?.primary_host || primaryHosts[0] || '',
    menuSpaceCount: Number(
      item?.menu_space_count ?? item?.space_count ?? 0
    ),
    menuCount: Number(item?.menu_count ?? 0),
    pageCount: Number(item?.page_count ?? 0),
    meta: item?.meta || {},
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeAppHostBinding(item: any): Api.SystemManage.AppHostBindingItem {
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    appName: item?.app_name || '',
    matchType: item?.match_type || 'host_exact',
    host: `${item?.host || ''}`.trim(),
    pathPattern: `${item?.path_pattern || ''}`.trim(),
    priority: Number(item?.priority ?? 0),
    defaultSpaceKey: item?.default_space_key || '',
    description: item?.description || '',
    isPrimary: Boolean(item?.is_primary ?? false),
    status: item?.status || 'normal',
    meta: item?.meta || {},
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeMenuSpaceEntryBinding(item: any): Api.SystemManage.MenuSpaceEntryBindingItem {
  return {
    id: item?.id || '',
    appKey: item?.app_key || '',
    appName: item?.app_name || '',
    spaceKey: item?.space_key || '',
    spaceName: item?.space_name || '',
    matchType: item?.match_type || 'host_exact',
    host: `${item?.host || ''}`.trim(),
    pathPattern: `${item?.path_pattern || ''}`.trim(),
    priority: Number(item?.priority ?? 0),
    description: item?.description || '',
    isPrimary: Boolean(item?.is_primary ?? false),
    status: item?.status || 'normal',
    meta: item?.meta || {},
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizePageBreadcrumbPreviewItem(
  item: V5PageBreadcrumbPreviewLike | undefined
): Api.SystemManage.PageBreadcrumbPreviewItem {
  return {
    type: item?.type || 'page',
    title: item?.title || '',
    path: item?.path || '',
    pageKey: item?.page_key || ''
  }
}

export function normalizeRefreshStats(item: V5RefreshStatsLike | undefined): Api.SystemManage.RefreshStats {
  return {
    requestedPackageCount: Number(
      item?.requested_package_count ?? 0
    ),
    impactedPackageCount: Number(item?.impacted_package_count ?? 0),
    roleCount: Number(item?.role_count ?? 0),
    collaborationWorkspaceCount: Number(item?.collaboration_workspace_count ?? 0),
    userCount: Number(item?.user_count ?? 0),
    elapsedMilliseconds: Number(item?.elapsed_milliseconds ?? 0),
    finishedAt: item?.finished_at || ''
  }
}

export function normalizeRiskAudit(item: V5RiskAuditItem): Api.SystemManage.RiskAuditItem {
  return {
    id: item?.id || '',
    operatorId: item?.operator_id || '',
    objectType: item?.object_type || '',
    objectId: item?.object_id || '',
    operationType: item?.operation_type || '',
    beforeSummary: (item?.before_summary || {}) as V5RiskAuditSummary,
    afterSummary: (item?.after_summary || {}) as V5RiskAuditSummary,
    impactSummary: (item?.impact_summary || {}) as V5RiskAuditSummary,
    requestId: item?.request_id || '',
    createdAt: item?.created_at || ''
  }
}

export function normalizePermissionBatchTemplate(
  item: V5PermissionBatchTemplateItem
): Api.SystemManage.PermissionBatchTemplateItem {
  return {
    id: item?.id || '',
    name: item?.name || '',
    description: item?.description || '',
    payload: (item?.payload || {}) as V5PermissionBatchTemplatePayload,
    createdBy: item?.created_by || '',
    createdAt: item?.created_at || '',
    updatedAt: item?.updated_at || ''
  }
}

export function normalizeFeaturePackageRelationNode(
  item: any
): Api.SystemManage.FeaturePackageRelationNode {
  const packageKey = item?.package_key || ''
  const appKeysRaw = item?.app_keys || []
  const appKeys = Array.isArray(appKeysRaw)
    ? appKeysRaw.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
    : []
  return {
    id: item?.id || '',
    packageKey,
    name: item?.name || '',
    packageType: item?.package_type || 'base',
    workspaceScope: item?.workspace_scope || 'all',
    appKeys,
    status: item?.status || 'normal',
    referenceCount: Number(item?.reference_count ?? 0),
    children: Array.isArray(item?.children)
      ? item.children.map((child: any) => normalizeFeaturePackageRelationNode(child))
      : []
  }
}

export function normalizeFeaturePackageRelationTree(
  item: any
): Api.SystemManage.FeaturePackageRelationTree {
  const cycleDependencies = item?.cycle_dependencies || []
  const isolatedBaseKeys = item?.isolated_base_keys || []
  return {
    roots: Array.isArray(item?.roots)
      ? item.roots.map((node: any) => normalizeFeaturePackageRelationNode(node))
      : [],
    cycleDependencies: Array.isArray(cycleDependencies)
      ? cycleDependencies.map((cycle: any) =>
          Array.isArray(cycle)
            ? cycle.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
            : []
        )
      : [],
    isolatedBaseKeys: Array.isArray(isolatedBaseKeys)
      ? isolatedBaseKeys.map((value: any) => `${value || ''}`.trim()).filter(Boolean)
      : []
  }
}

export function normalizePermissionActionConsumers(
  item: V5PermissionActionConsumerDetailsLike | undefined
): Api.SystemManage.PermissionActionConsumerDetails {
  const featurePackages = item?.feature_packages || []
  return {
    permissionKey: item?.permission_key || '',
    apis: Array.isArray(item?.apis)
      ? item.apis.map((api: any) => ({
          code: api?.code || '',
          method: api?.method || '',
          path: api?.path || '',
          summary: api?.summary || ''
        }))
      : [],
    pages: Array.isArray(item?.pages)
      ? item.pages.map((page: any) => ({
          pageKey: page?.page_key || page?.pageKey || '',
          name: page?.name || '',
          routePath: page?.route_path || page?.routePath || '',
          accessMode: page?.access_mode || page?.accessMode || ''
        }))
      : [],
    featurePackages: Array.isArray(featurePackages)
      ? featurePackages.map((pkg: any) => ({
          id: pkg?.id || '',
          packageKey: pkg?.package_key || pkg?.packageKey || '',
          name: pkg?.name || '',
          packageType: pkg?.package_type || pkg?.packageType || '',
          contextType: pkg?.context_type || pkg?.contextType || 'common'
        }))
      : [],
    roles: Array.isArray(item?.roles)
      ? item.roles.map((role: any) => ({
          id: role?.id || '',
          code: role?.code || '',
          name: role?.name || '',
          contextType: role?.context_type || role?.contextType || ''
        }))
      : []
  }
}

export function normalizeUnregisteredApiScanConfig(item: any): Api.SystemManage.APIUnregisteredScanConfig {
  return {
    enabled: Boolean(item?.enabled),
    frequencyMinutes: Number(item?.frequency_minutes ?? 60),
    defaultCategoryId: item?.default_category_id || '',
    defaultPermissionKey: item?.default_permission_key || '',
    markAsNoPermission: Boolean(item?.mark_as_no_permission)
  }
}

export function normalizePageAccessTraceResult(
  item: V5PageAccessTraceResultLike | undefined
): Api.SystemManage.PageAccessTraceResult {
  const visibleMenuIds = item?.visible_menu_ids || []
  return {
    userId: item?.user_id || '',
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || '',
    spaceKey: item?.space_key || '',
    authenticated: Boolean(item?.authenticated),
    superAdmin: Boolean(item?.super_admin),
    actionKeyCount: Number(item?.action_key_count ?? 0),
    visibleMenuIds: Array.isArray(visibleMenuIds)
      ? visibleMenuIds.map((value) => `${value || ''}`.trim()).filter(Boolean)
      : [],
    menus: Array.isArray(item?.menus)
      ? item.menus.map((menu: any) => ({
          id: menu?.id || '',
          parentId: menu?.parent_id || menu?.parentId || '',
          name: menu?.name || '',
          title: menu?.title || '',
          path: menu?.path || '',
          fullPath: menu?.full_path || menu?.fullPath || '',
          kind: menu?.kind || '',
          icon: menu?.icon || '',
          sortOrder: Number(menu?.sort_order ?? menu?.sortOrder ?? 0),
          hidden: Boolean(menu?.hidden),
          visible: Boolean(menu?.visible)
        }))
      : [],
    roles: Array.isArray(item?.roles)
      ? item.roles.map((role: any) => ({
          roleId: role?.role_id || role?.roleId || '',
          roleCode: role?.role_code || role?.roleCode || '',
          roleName: role?.role_name || role?.roleName || '',
          status: role?.status || ''
        }))
      : [],
    pages: Array.isArray(item?.pages)
      ? item.pages.map((page: any) => ({
          pageKey: page?.page_key || page?.pageKey || '',
          pageName: page?.page_name || page?.pageName || '',
          routePath: page?.route_path || page?.routePath || '',
          accessMode: page?.access_mode || page?.accessMode || '',
          permissionKey: page?.permission_key || page?.permissionKey || '',
          parentPageKey: page?.parent_page_key || page?.parentPageKey || '',
          parentMenuId: page?.parent_menu_id || page?.parentMenuId || '',
          activeMenuPath: page?.active_menu_path || page?.activeMenuPath || '',
          visible: Boolean(page?.visible),
          reason: page?.reason || '',
          matchedActionKey: page?.matched_action_key || page?.matchedActionKey || '',
          effectiveChain: Array.isArray(page?.effective_chain || page?.effectiveChain)
            ? page?.effective_chain || page?.effectiveChain
            : []
        }))
      : []
  }
}

export function normalizeRuntimeMenuTree(item: any): AppRouteRecord {
  const meta = item?.meta || {}
  const children = Array.isArray(item?.children)
    ? item.children.map((child: any) => normalizeRuntimeMenuTree(child))
    : []
  const spaceKey = normalizeMenuSpaceKey(item?.space_key || meta?.spaceKey)
  const spaceType = `${item?.space_type || meta?.spaceType || ''}`.trim()
  const hostKey = `${item?.host_key || meta?.hostKey || ''}`.trim()
  return {
    id: item?.id || '',
    kind: item?.kind || '',
    path: item?.path || '',
    name: item?.name || '',
    component: item?.component || '',
    parent_id: item?.parent_id || '',
    sort_order: item?.sort_order ?? 0,
    redirect: item?.redirect || '',
    spaceKey,
    spaceType,
    hostKey,
    meta: {
      title: meta?.title || item?.title || item?.name || '',
      icon: meta?.icon || '',
      // accessMode 由后端运行时菜单显式下发；空值时交给前端/后续链路按默认 permission 处理。
      accessMode: `${meta?.accessMode || ''}`.trim() || undefined,
      // activePath 仅在需要高亮指定菜单时才返回，缺省表示沿用当前路由 path。
      activePath: `${meta?.activePath || ''}`.trim() || undefined,
      // link 仅外链菜单需要，未返回时保持 undefined，避免误判成空链接。
      link: `${meta?.link || ''}`.trim() || undefined,
      spaceKey: spaceKey || undefined,
      spaceType: spaceType || undefined,
      hostKey: hostKey || undefined,
      // 运行时菜单里未返回 isEnable 表示“沿用默认启用”，不能归一成 false，
      // 否则前端兜底过滤会把整棵菜单树误判为禁用。
      isEnable: meta?.isEnable === false ? false : undefined,
      // 后端运行时菜单对这些展示型布尔字段采用“只在为 true 时才返回”的约定，
      // 所以前端归一化时只认显式 true，缺省一律视为未开启该特性。
      isHide: meta?.isHide === true,
      isIframe: meta?.isIframe === true,
      isHideTab: meta?.isHideTab === true,
      keepAlive: meta?.keepAlive === true,
      fixedTab: meta?.fixedTab === true,
      isFullPage: meta?.isFullPage === true
    },
    children
  } as AppRouteRecord
}

export function normalizeRuntimeNavigationManifest(item: any): Api.SystemManage.RuntimeNavigationManifest {
  const currentApp = item?.current_app
  const currentSpace = item?.current_space || {}
  const menuTree = item?.menu_tree || []
  const entryRoutes = item?.entry_routes || []
  const managedPages = item?.managed_pages || []
  const space = currentSpace?.space ? normalizeMenuSpace(currentSpace.space) : undefined
  const binding = currentSpace?.binding
    ? normalizeMenuSpaceHostBinding(currentSpace.binding)
    : undefined

  return {
    currentApp: currentApp
      ? {
          app: normalizeApp(currentApp?.app || {}),
          binding: currentApp?.binding ? normalizeAppHostBinding(currentApp.binding) : undefined,
          resolvedBy: currentApp?.resolved_by || currentApp?.resolvedBy || '',
          requestHost: currentApp?.request_host || currentApp?.requestHost || ''
        }
      : undefined,
    // currentSpace 是后端对 Host / 显式 space_key 解析后的最终上下文，前端菜单树和受管页面都必须跟随它。
    currentSpace: {
      space,
      binding,
      resolvedBy: currentSpace?.resolved_by || currentSpace?.resolvedBy || '',
      requestHost: currentSpace?.request_host || currentSpace?.requestHost || '',
      accessGranted: Boolean(currentSpace?.access_granted ?? currentSpace?.accessGranted ?? true)
    },
    context: {
      ...(item?.context || {}),
      app_key: item?.context?.app_key || '',
      space_key: item?.context?.space_key || '',
      requested_space_key:
        item?.context?.requested_space_key || '',
      request_host: item?.context?.request_host || ''
    },
    // menuTree 已经完成启用态、空间和权限裁剪；前端这里只做归一化与动态注册。
    menuTree: Array.isArray(menuTree)
      ? menuTree.map((entry: any) => normalizeRuntimeMenuTree(entry))
      : [],
    entryRoutes: Array.isArray(entryRoutes)
      ? entryRoutes.map((entry: any) => normalizeRuntimeMenuTree(entry))
      : [],
    managedPages: Array.isArray(managedPages)
      ? managedPages.map((entry: any) => normalizePageItem(entry))
      : [],
    versionStamp: item?.version_stamp || ''
  }
}

export function normalizeApiEndpointCategory(item: any): Api.SystemManage.APIEndpointCategoryItem {
  return {
    id: item?.id || '',
    code: item?.code || '',
    name: item?.name || '',
    nameEn: item?.name_en || '',
    sortOrder: item?.sort_order ?? 0,
    status: item?.status || 'normal'
  }
}

export function normalizeUnregisteredApiRoute(item: any): Api.SystemManage.APIUnregisteredRouteItem {
  return {
    method: item?.method || '',
    path: item?.path || '',
    spec: item?.spec || `${item?.method || ''} ${item?.path || ''}`.trim(),
    handler: item?.handler || '',
    hasMeta: Boolean(item?.has_meta),
    meta: item?.meta
      ? {
          summary: item.meta.summary || '',
          category_code: item.meta.category_code || '',
          context_scope: item.meta.context_scope || 'optional',
          source: item.meta.source || '',
          feature_kind: item.meta.feature_kind || 'system',
          permission_keys: Array.isArray(item.meta.permission_keys) ? item.meta.permission_keys : []
        }
      : undefined
  }
}

export function normalizeFastEnterConfig(item: any): Api.SystemManage.FastEnterConfig {
  const applications = Array.isArray(item?.applications) ? item.applications : []
  const quickLinks = Array.isArray(item?.quickLinks)
    ? item.quickLinks
    : Array.isArray(item?.quick_links)
      ? item.quick_links
      : []

  return {
    applications: applications.map(
      (entry: any): FastEnterApplication => ({
        id: entry?.id || '',
        name: entry?.name || '',
        description: entry?.description || '',
        icon: entry?.icon || 'ri:apps-2-line',
        iconColor: entry?.iconColor || entry?.icon_color || '#377dff',
        enabled: entry?.enabled !== false,
        order: entry?.order ?? 0,
        routeName: entry?.routeName || entry?.route_name || '',
        link: entry?.link || ''
      })
    ),
    quickLinks: quickLinks.map(
      (entry: any): FastEnterQuickLink => ({
        id: entry?.id || '',
        name: entry?.name || '',
        enabled: entry?.enabled !== false,
        order: entry?.order ?? 0,
        routeName: entry?.routeName || entry?.route_name || '',
        link: entry?.link || ''
      })
    ),
    minWidth: item?.min_width ?? 1450
  }
}

export function normalizeUserPermissionMenuTree(
  item: V5UserPermissionMenuTreeItem
): Api.SystemManage.UserPermissionMenuNode {
  return {
    id: item?.id || '',
    name: item?.name || '',
    title: item?.title || '',
    path: item?.path || '',
    component: item?.component || '',
    hidden: Boolean(item?.hidden),
    sort: item?.sort_order ?? 0,
    children: (item?.children || []).map((child) => normalizeUserPermissionMenuTree(child))
  }
}

export function normalizeUserCollaborationWorkspaceItem(
  item: any
): Api.SystemManage.CollaborationWorkspaceListItem {
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
    currentRoleCode: item?.current_role_code || '',
    memberStatus: item?.member_status || ''
  }
}

export function normalizeUserPermissionDiagnosisResponse(
  item: V5UserPermissionDiagnosisResponse
): Api.SystemManage.UserPermissionDiagnosisResponse {
  const normalizePackages = (items: components['schemas']['FeaturePackageRef'][] | undefined) =>
    (items || []).map(normalizeFeaturePackage)
  const normalizeGroup = (
    value: components['schemas']['PermissionGroupItem'] | undefined
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

  const normalizeRoleResult = (
    role: V5UserPermissionRoleResult
  ): Api.SystemManage.UserPermissionRoleResult => ({
    roleId: role?.role_id || '',
    roleCode: role?.role_code || '',
    roleName: role?.role_name || '',
    inherited: Boolean(role?.inherited),
    refreshedAt: role?.refreshed_at || '',
    availableActionCount: role?.available_action_count ?? 0,
    disabledActionCount: role?.disabled_action_count ?? 0,
    effectiveActionCount: role?.effective_action_count ?? 0,
    matched: Boolean(role?.matched),
    disabled: Boolean(role?.disabled),
    available: Boolean(role?.available),
    sourcePackages: normalizePackages(role?.source_packages)
  })

  const normalizeAction = (
    action: V5UserPermissionDiagnosisAction | undefined
  ): Api.SystemManage.UserPermissionDiagnosisAction | null =>
    action
      ? {
          id: action?.id || '',
          permissionKey: action?.permission_key || '',
          name: action?.name || '',
          description: action?.description || '',
          status: action?.status || '',
          selfStatus: action?.self_status || '',
          contextType: action?.context_type || '',
          featureKind: action?.feature_kind || '',
          moduleCode: action?.module_code || '',
          moduleGroupStatus: action?.module_group_status || '',
          featureGroupStatus: action?.feature_group_status || '',
          moduleGroup: normalizeGroup(action?.module_group),
          featureGroup: normalizeGroup(action?.feature_group)
        }
      : null

  const diagnosis = item?.diagnosis
    ? {
        permissionKey: item.diagnosis?.permission_key || '',
        allowed: Boolean(item.diagnosis?.allowed),
        reasonText: item.diagnosis?.reason_text || '',
        reasons: item.diagnosis?.reasons || [],
        matchedInSnapshot: Boolean(item.diagnosis?.matched_in_snapshot),
        bypassedBySuperAdmin: Boolean(item.diagnosis?.bypassed_by_super_admin),
        blockedByCollaborationWorkspace: Boolean(
          item.diagnosis?.blocked_by_collaboration_workspace
        ),
        denialStage: item.diagnosis?.denial_stage || '',
        denialReason: item.diagnosis?.denial_reason || '',
        memberStatus: item.diagnosis?.member_status || '',
        memberMatched: Boolean(item.diagnosis?.member_matched),
        boundaryState: item.diagnosis?.boundary_state || '',
        boundaryConfigured: Boolean(item.diagnosis?.boundary_configured),
        roleChainMatched: Boolean(item.diagnosis?.role_chain_matched),
        roleChainDisabled: Boolean(item.diagnosis?.role_chain_disabled),
        roleChainAvailable: Boolean(item.diagnosis?.role_chain_available),
        action: normalizeAction(item.diagnosis?.action),
        sourcePackages: normalizePackages(item.diagnosis?.source_packages),
        roleResults: (item.diagnosis?.role_results || []).map((role) => normalizeRoleResult(role))
      }
    : null

  const context = {
    type: item?.context?.type || 'personal',
    collaborationWorkspaceId: item?.context?.current_collaboration_workspace_id || '',
    collaborationWorkspaceName: item?.context?.current_collaboration_workspace_name || ''
  } as Api.SystemManage.UserPermissionContext & {
    collaborationWorkspaceName?: string
  }

  return {
    user: {
      id: item?.user?.id || '',
      userName: item?.user?.user_name || '',
      nickName: item?.user?.nick_name || '',
      status: item?.user?.status || 'inactive',
      isSuperAdmin: Boolean(item?.user?.is_super_admin)
    },
    context,
    snapshot: {
      refreshedAt: item?.snapshot?.refreshed_at || '',
      updatedAt: item?.snapshot?.updated_at || '',
      roleCount: item?.snapshot?.role_count ?? 0,
      directPackageCount: item?.snapshot?.direct_package_count ?? 0,
      expandedPackageCount: item?.snapshot?.expanded_package_count ?? 0,
      actionCount: item?.snapshot?.action_count ?? 0,
      disabledActionCount: item?.snapshot?.disabled_action_count ?? 0,
      menuCount: item?.snapshot?.menu_count ?? 0,
      hasPackageConfig: Boolean(item?.snapshot?.has_package_config),
      derivedActionCount: item?.snapshot?.derived_action_count ?? 0,
      blockedActionCount: item?.snapshot?.blocked_action_count ?? 0,
      effectiveActionCount: item?.snapshot?.effective_action_count ?? 0
    },
    roles: (item?.roles || []).map((role) => normalizeRoleResult(role)),
    collaborationWorkspaceMember: item?.collaboration_workspace_member
      ? {
          id: item?.collaboration_workspace_member?.id || '',
          collaborationWorkspaceId:
            item?.collaboration_workspace_member?.collaboration_workspace_id || '',
          userId: item?.collaboration_workspace_member?.user_id || '',
          roleCode: item?.collaboration_workspace_member?.role_code || '',
          status: item?.collaboration_workspace_member?.status || '',
          matched: Boolean(item?.collaboration_workspace_member?.matched)
        }
      : null,
    collaborationWorkspacePackages: normalizePackages(item?.collaboration_workspace_packages),
    diagnosis,
    menus: (item?.menus || []).map((menu) => normalizeUserPermissionMenuTree(menu))
  }
}
