import { requestData } from '@/shared/api/client'
import type {
  AccessTraceFilter,
  AccessTraceResult,
  ApiScanConfig,
  ApiEndpointCategory,
  ApiEndpointOverview,
  ApiEndpointRecord,
  ApiUnregisteredRouteRecord,
  ApiEndpointSavePayload,
  FastEnterConfig,
  FastEnterItem,
  MenuSpaceHostBindingRecord,
  MenuSpaceHostBindingSavePayload,
  MenuSpaceRecord,
  MenuSpaceSavePayload,
  PageBreadcrumbPreviewItem,
  PageDefinition,
  PageGroupOption,
  PageSavePayload,
  PageSyncResult,
  PageUnregisteredRecord,
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

function normalizeTraceEntries(input: unknown, labelPrefix: string) {
  if (!Array.isArray(input)) {
    return [] as Array<{ label: string; value: string }>
  }

  return input.map((item, index) => ({
    label: `${labelPrefix} ${index + 1}`,
    value: stringifyValue(item),
  }))
}

function normalizePage(input: Record<string, unknown>): PageDefinition {
  const meta = ((input.meta || {}) as Record<string, unknown>) || {}
  const spaceKeys = Array.isArray(input.space_keys || input.spaceKeys || meta.spaceKeys)
    ? ((input.space_keys || input.spaceKeys || meta.spaceKeys) as unknown[])
        .map(normalizeText)
        .filter(Boolean)
    : []

  return {
    id: normalizeText(input.id),
    pageKey: normalizeText(input.page_key || input.pageKey),
    name: normalizeText(input.name),
    routeName: normalizeText(input.route_name || input.routeName),
    routePath: normalizeText(input.route_path || input.routePath),
    component: normalizeText(input.component),
    pageType: normalizeText(input.page_type || input.pageType || 'inner'),
    source: normalizeText(input.source || 'manual'),
    moduleKey: normalizeText(input.module_key || input.moduleKey),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    parentMenuId: normalizeText(input.parent_menu_id || input.parentMenuId),
    parentMenuName: normalizeText(input.parent_menu_name || input.parentMenuName),
    parentPageKey: normalizeText(input.parent_page_key || input.parentPageKey),
    parentPageName: normalizeText(input.parent_page_name || input.parentPageName),
    displayGroupKey: normalizeText(input.display_group_key || input.displayGroupKey),
    displayGroupName: normalizeText(input.display_group_name || input.displayGroupName),
    activeMenuPath: normalizeText(input.active_menu_path || input.activeMenuPath),
    breadcrumbMode: normalizeText(input.breadcrumb_mode || input.breadcrumbMode || 'inherit_menu'),
    accessMode: normalizeText(input.access_mode || input.accessMode || 'inherit'),
    permissionKey: normalizeText(input.permission_key || input.permissionKey),
    inheritPermission: Boolean(input.inherit_permission ?? input.inheritPermission ?? true),
    keepAlive: Boolean(input.keep_alive ?? input.keepAlive),
    isFullPage: Boolean(input.is_full_page ?? input.isFullPage),
    isIframe: Boolean(meta.isIframe ?? input.is_iframe ?? input.isIframe),
    isHideTab: Boolean(meta.isHideTab ?? input.is_hide_tab ?? input.isHideTab),
    link: normalizeText(meta.link || input.link),
    spaceKey: normalizeText(input.space_key || input.spaceKey || meta.spaceKey),
    spaceKeys,
    spaceScope: normalizeText(input.space_scope || input.spaceScope || meta.spaceScope) || undefined,
    status: normalizeText(input.status || 'normal'),
    meta,
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizePageMenuOption(input: Record<string, unknown>): PageGroupOption {
  return {
    id: normalizeText(input.id),
    name: normalizeText(input.name),
    title: normalizeText(input.title),
    path: normalizeText(input.path),
    children: Array.isArray(input.children)
      ? input.children.map((item) => normalizePageMenuOption(item as Record<string, unknown>))
      : [],
  }
}

function normalizePageUnregistered(input: Record<string, unknown>): PageUnregisteredRecord {
  return {
    filePath: normalizeText(input.file_path || input.filePath),
    component: normalizeText(input.component),
    pageKey: normalizeText(input.page_key || input.pageKey),
    name: normalizeText(input.name),
    routeName: normalizeText(input.route_name || input.routeName),
    routePath: normalizeText(input.route_path || input.routePath),
    pageType: normalizeText(input.page_type || input.pageType || 'inner'),
    moduleKey: normalizeText(input.module_key || input.moduleKey),
    parentMenuId: normalizeText(input.parent_menu_id || input.parentMenuId),
    parentMenuName: normalizeText(input.parent_menu_name || input.parentMenuName),
    activeMenuPath: normalizeText(input.active_menu_path || input.activeMenuPath),
    spaceKey: normalizeText(input.space_key || input.spaceKey),
    spaceType: normalizeText(input.space_type || input.spaceType),
    hostKey: normalizeText(input.host_key || input.hostKey),
  }
}

function normalizePageBreadcrumbPreview(input: Record<string, unknown>): PageBreadcrumbPreviewItem {
  return {
    type: normalizeText(input.type || 'page'),
    title: normalizeText(input.title),
    path: normalizeText(input.path),
    pageKey: normalizeText(input.page_key || input.pageKey),
  }
}

function normalizeApiCategory(input: Record<string, unknown>): ApiEndpointCategory {
  return {
    id: normalizeText(input.id),
    code: normalizeText(input.code),
    name: normalizeText(input.name),
    nameEn: normalizeText(input.name_en || input.nameEn),
    description: normalizeText(input.description),
    sortOrder: Number(input.sort_order || input.sortOrder || 0),
    status: normalizeText(input.status || 'normal'),
  }
}

function normalizeUnregisteredApiRoute(input: Record<string, unknown>): ApiUnregisteredRouteRecord {
  return {
    method: normalizeText(input.method),
    path: normalizeText(input.path),
    handler: normalizeText(input.handler),
    summary: normalizeText(input.summary),
    featureKind: normalizeText(input.feature_kind || input.featureKind || 'system'),
    suggestedPermissionKey: normalizeText(input.suggested_permission_key || input.suggestedPermissionKey),
    categoryCode: normalizeText(input.category_code || input.categoryCode),
  }
}

function normalizeApiScanConfig(input: Record<string, unknown>): ApiScanConfig {
  return {
    enabled: Boolean(input.enabled),
    frequencyMinutes: Number(input.frequency_minutes || input.frequencyMinutes || 0),
    defaultCategoryId: normalizeText(input.default_category_id || input.defaultCategoryId),
    defaultPermissionKey: normalizeText(input.default_permission_key || input.defaultPermissionKey),
    markAsNoPermission: Boolean(input.mark_as_no_permission ?? input.markAsNoPermission),
  }
}

function normalizeApiEndpoint(input: Record<string, unknown>): ApiEndpointRecord {
  const permissionKeys = Array.isArray(input.permission_keys || input.permissionKeys)
    ? ((input.permission_keys || input.permissionKeys) as unknown[])
        .map(normalizeText)
        .filter(Boolean)
    : []

  return {
    id: normalizeText(input.id),
    code: normalizeText(input.code),
    method: normalizeText(input.method),
    path: normalizeText(input.path),
    spec: normalizeText(input.spec),
    featureKind: normalizeText(input.feature_kind || input.featureKind || 'system'),
    handler: normalizeText(input.handler),
    summary: normalizeText(input.summary),
    permissionKey: normalizeText(input.permission_key || input.permissionKey || permissionKeys[0]),
    permissionKeys,
    permissionContexts: Array.isArray(input.permission_contexts || input.permissionContexts)
      ? ((input.permission_contexts || input.permissionContexts) as unknown[])
          .map(normalizeText)
          .filter(Boolean)
      : [],
    permissionBindingMode: normalizeText(input.permission_binding_mode || input.permissionBindingMode || 'none'),
    sharedAcrossContexts: Boolean(input.shared_across_contexts ?? input.sharedAcrossContexts),
    permissionNote: normalizeText(input.permission_note || input.permissionNote),
    authMode: normalizeText(input.auth_mode || input.authMode || 'jwt'),
    categoryId: normalizeText(input.category_id || input.categoryId || (input.category as Record<string, unknown> | undefined)?.id),
    category:
      input.category && typeof input.category === 'object'
        ? normalizeApiCategory(input.category as Record<string, unknown>)
        : undefined,
    contextScope: normalizeText(input.context_scope || input.contextScope || 'optional'),
    source: normalizeText(input.source || 'sync'),
    dataPermissionCode: normalizeText(input.data_permission_code || input.dataPermissionCode),
    dataPermissionName: normalizeText(input.data_permission_name || input.dataPermissionName),
    runtimeExists: Boolean(input.runtime_exists ?? input.runtimeExists),
    stale: Boolean(input.stale),
    staleReason: normalizeText(input.stale_reason || input.staleReason),
    status: normalizeText(input.status || 'normal'),
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeFastEnterItem(input: Record<string, unknown>): FastEnterItem {
  return {
    id: normalizeText(input.id),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    icon: normalizeText(input.icon),
    iconColor: normalizeText(input.iconColor || input.icon_color),
    enabled: Boolean(input.enabled ?? true),
    order: Number(input.order || 0),
    routeName: normalizeText(input.routeName || input.route_name),
    link: normalizeText(input.link) || undefined,
  }
}

function normalizeMenuSpace(input: Record<string, unknown>): MenuSpaceRecord {
  return {
    id: normalizeText(input.id),
    spaceKey: normalizeText(input.space_key || input.spaceKey),
    name: normalizeText(input.name),
    description: normalizeText(input.description),
    defaultHomePath: normalizeText(input.default_home_path || input.defaultHomePath),
    isDefault: Boolean(input.is_default ?? input.isDefault),
    status: normalizeText(input.status || 'normal'),
    hostCount: Number(input.host_count || input.hostCount || 0),
    hosts: Array.isArray(input.hosts) ? input.hosts.map(normalizeText).filter(Boolean) : [],
    menuCount: Number(input.menu_count || input.menuCount || 0),
    pageCount: Number(input.page_count || input.pageCount || 0),
    accessMode: normalizeText(input.access_mode || input.accessMode || ((input.meta || {}) as Record<string, unknown>).accessMode || 'all'),
    allowedRoleCodes: Array.isArray(input.allowed_role_codes || input.allowedRoleCodes)
      ? ((input.allowed_role_codes || input.allowedRoleCodes) as unknown[])
          .map(normalizeText)
          .filter(Boolean)
      : [],
    meta: ((input.meta || {}) as Record<string, unknown>) || {},
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function normalizeMenuSpaceHostBinding(input: Record<string, unknown>): MenuSpaceHostBindingRecord {
  const meta = ((input.meta || {}) as Record<string, unknown>) || {}
  return {
    id: normalizeText(input.id),
    host: normalizeText(input.host),
    spaceKey: normalizeText(input.space_key || input.spaceKey),
    spaceName: normalizeText(input.space_name || input.spaceName),
    description: normalizeText(input.description),
    isDefault: Boolean(input.is_default ?? input.isDefault),
    status: normalizeText(input.status || 'normal'),
    scheme: normalizeText(input.scheme || meta.scheme || 'https'),
    routePrefix: normalizeText(input.route_prefix || input.routePrefix || meta.routePrefix),
    authMode: normalizeText(input.auth_mode || input.authMode || meta.authMode || 'inherit_host'),
    loginHost: normalizeText(input.login_host || input.loginHost || meta.loginHost),
    callbackHost: normalizeText(input.callback_host || input.callbackHost || meta.callbackHost),
    cookieScopeMode: normalizeText(input.cookie_scope_mode || input.cookieScopeMode || meta.cookieScopeMode || 'inherit'),
    cookieDomain: normalizeText(input.cookie_domain || input.cookieDomain || meta.cookieDomain),
    meta,
    createdAt: normalizeText(input.created_at || input.createdAt),
    updatedAt: normalizeText(input.updated_at || input.updatedAt),
  }
}

function buildPagePayload(payload: PageSavePayload) {
  return {
    page_key: payload.pageKey,
    name: payload.name,
    route_name: payload.routeName,
    route_path: payload.routePath,
    component: payload.component,
    page_type: payload.pageType,
    module_key: payload.moduleKey,
    sort_order: payload.sortOrder,
    parent_menu_id: payload.parentMenuId || undefined,
    parent_page_key: payload.parentPageKey || undefined,
    display_group_key: payload.displayGroupKey || undefined,
    active_menu_path: payload.activeMenuPath || undefined,
    breadcrumb_mode: payload.breadcrumbMode,
    access_mode: payload.accessMode,
    permission_key: payload.permissionKey || undefined,
    inherit_permission: payload.inheritPermission,
    keep_alive: payload.keepAlive,
    is_full_page: payload.isFullPage,
    space_key: payload.spaceKey || undefined,
    meta: {
      ...(payload.meta || {}),
      ...(payload.isHideTab ? { isHideTab: true } : {}),
      ...(payload.link ? { link: payload.link } : {}),
    },
  }
}

export async function fetchPageList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/pages',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizePage(item)),
  }
}

export async function fetchPageDetail(pageId: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: `/api/v1/pages/${pageId}`,
  })

  return normalizePage(result)
}

export async function createPage(payload: PageSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/pages',
    data: buildPagePayload(payload),
  })

  return normalizePage(result)
}

export async function updatePage(pageId: string, payload: PageSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/pages/${pageId}`,
    data: buildPagePayload(payload),
  })

  return normalizePage(result)
}

export async function deletePage(pageId: string) {
  await requestData<unknown>({
    method: 'DELETE',
    url: `/api/v1/pages/${pageId}`,
  })
}

export async function fetchPageMenuOptions(spaceKey?: string) {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: '/api/v1/pages/menu-options',
    params: spaceKey ? { space_key: spaceKey } : undefined,
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records) ? result.records.map((item) => normalizePageMenuOption(item)) : [],
  }
}

export async function fetchPageUnregisteredList() {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: '/api/v1/pages/unregistered',
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records)
      ? result.records.map((item) => normalizePageUnregistered(item))
      : [],
  }
}

export async function syncPages() {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/pages/sync',
  })

  return {
    createdCount: Number(result.createdCount || result.created_count || 0),
    skippedCount: Number(result.skippedCount || result.skipped_count || 0),
    createdKeys: Array.isArray(result.createdKeys || result.created_keys)
      ? ((result.createdKeys || result.created_keys) as unknown[])
          .map(normalizeText)
          .filter(Boolean)
      : [],
  } satisfies PageSyncResult
}

export async function fetchPageBreadcrumbPreview(pageId: string) {
  const result = await requestData<{
    items?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: `/api/v1/pages/${pageId}/breadcrumb-preview`,
  })

  return {
    total: Number(result.total || 0),
    items: Array.isArray(result.items)
      ? result.items.map((item) => normalizePageBreadcrumbPreview(item))
      : [],
  }
}

export async function fetchApiEndpointList(params?: Record<string, unknown>) {
  const result = await requestData<PaginationEnvelope<Record<string, unknown>>>({
    method: 'GET',
    url: '/api/v1/api-endpoints',
    params,
  })

  const normalized = toPaginatedResult(result)
  return {
    ...normalized,
    records: normalized.records.map((item) => normalizeApiEndpoint(item)),
  }
}

export async function fetchApiEndpointOverview() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/api-endpoints/overview',
  })

  return {
    totalCount: Number(result.totalCount || result.total_count || 0),
    uncategorizedCount: Number(result.uncategorizedCount || result.uncategorized_count || 0),
    staleCount: Number(result.staleCount || result.stale_count || 0),
    noPermissionCount: Number(result.noPermissionCount || result.no_permission_count || 0),
    sharedPermissionCount: Number(result.sharedPermissionCount || result.shared_permission_count || 0),
    crossContextSharedCount: Number(result.crossContextSharedCount || result.cross_context_shared_count || 0),
  } satisfies ApiEndpointOverview
}

export async function fetchApiCategories() {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: '/api/v1/api-endpoints/categories',
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records) ? result.records.map((item) => normalizeApiCategory(item)) : [],
  }
}

export async function fetchUnregisteredApiRoutes(params?: Record<string, unknown>) {
  const result = await requestData<{
    current?: number
    total?: number
    size?: number
    records?: Array<Record<string, unknown>>
  }>({
    method: 'GET',
    url: '/api/v1/api-endpoints/unregistered',
    params,
  })

  return {
    current: Number(result.current || 1),
    total: Number(result.total || 0),
    size: Number(result.size || result.records?.length || 0),
    records: Array.isArray(result.records)
      ? result.records.map((item) => normalizeUnregisteredApiRoute(item))
      : [],
  }
}

export async function fetchUnregisteredApiScanConfig() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/api-endpoints/unregistered/scan-config',
  })

  return normalizeApiScanConfig(result)
}

export async function saveUnregisteredApiScanConfig(payload: Partial<ApiScanConfig>) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: '/api/v1/api-endpoints/unregistered/scan-config',
    data: {
      enabled: payload.enabled,
      frequency_minutes: payload.frequencyMinutes,
      default_category_id: payload.defaultCategoryId || undefined,
      default_permission_key: payload.defaultPermissionKey || undefined,
      mark_as_no_permission: payload.markAsNoPermission,
    },
  })

  return normalizeApiScanConfig(result)
}

export async function createApiEndpoint(payload: ApiEndpointSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/api-endpoints',
    data: {
      code: payload.code,
      method: payload.method,
      path: payload.path,
      summary: payload.summary,
      auth_mode: payload.authMode,
      permission_key: payload.permissionKey || undefined,
      category_id: payload.categoryId || undefined,
      context_scope: payload.contextScope,
      source: payload.source,
      status: payload.status,
    },
  })

  return normalizeApiEndpoint(result)
}

export async function updateApiEndpoint(endpointId: string, payload: ApiEndpointSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/api-endpoints/${endpointId}`,
    data: {
      code: payload.code,
      method: payload.method,
      path: payload.path,
      summary: payload.summary,
      auth_mode: payload.authMode,
      permission_key: payload.permissionKey || undefined,
      category_id: payload.categoryId || undefined,
      context_scope: payload.contextScope,
      source: payload.source,
      status: payload.status,
    },
  })

  return normalizeApiEndpoint(result)
}

export async function updateApiEndpointContextScope(endpointId: string, contextScope: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/api-endpoints/${endpointId}/context-scope`,
    data: {
      context_scope: contextScope,
    },
  })

  return normalizeApiEndpoint(result)
}

export async function createApiCategory(payload: Partial<ApiEndpointCategory>) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/api-endpoints/categories',
    data: {
      code: payload.code,
      name: payload.name,
      name_en: payload.nameEn,
      description: payload.description,
      sort_order: payload.sortOrder,
      status: payload.status,
    },
  })

  return normalizeApiCategory(result)
}

export async function updateApiCategory(categoryId: string, payload: Partial<ApiEndpointCategory>) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: `/api/v1/api-endpoints/categories/${categoryId}`,
    data: {
      code: payload.code,
      name: payload.name,
      name_en: payload.nameEn,
      description: payload.description,
      sort_order: payload.sortOrder,
      status: payload.status,
    },
  })

  return normalizeApiCategory(result)
}

export async function syncApiEndpoints() {
  await requestData<unknown>({
    method: 'POST',
    url: '/api/v1/api-endpoints/sync',
  })
}

export async function cleanupStaleApiEndpoints(ids: string[]) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/api-endpoints/cleanup-stale',
    data: { ids },
  })

  return {
    deletedCount: Number(result.deletedCount || result.deleted_count || 0),
  }
}

export async function fetchFastEnterConfig() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/system/fast-enter',
  })

  return {
    applications: Array.isArray(result.applications)
      ? result.applications.map((item) => normalizeFastEnterItem(item as Record<string, unknown>))
      : [],
    quickLinks: Array.isArray(result.quickLinks || result.quick_links)
      ? ((result.quickLinks || result.quick_links) as unknown[]).map((item) =>
          normalizeFastEnterItem(item as Record<string, unknown>),
        )
      : [],
    minWidth: Number(result.minWidth || result.min_width || 0),
  } satisfies FastEnterConfig
}

export async function updateFastEnterConfig(config: FastEnterConfig) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: '/api/v1/system/fast-enter',
    data: {
      applications: config.applications.map((item) => ({
        id: item.id,
        name: item.name,
        description: item.description,
        icon: item.icon,
        iconColor: item.iconColor,
        enabled: item.enabled,
        order: item.order,
        routeName: item.routeName,
        link: item.link,
      })),
      quickLinks: config.quickLinks.map((item) => ({
        id: item.id,
        name: item.name,
        description: item.description,
        icon: item.icon,
        iconColor: item.iconColor,
        enabled: item.enabled,
        order: item.order,
        routeName: item.routeName,
        link: item.link,
      })),
      minWidth: config.minWidth,
    },
  })

  return {
    applications: Array.isArray(result.applications)
      ? result.applications.map((item) => normalizeFastEnterItem(item as Record<string, unknown>))
      : [],
    quickLinks: Array.isArray(result.quickLinks || result.quick_links)
      ? ((result.quickLinks || result.quick_links) as unknown[]).map((item) =>
          normalizeFastEnterItem(item as Record<string, unknown>),
        )
      : [],
    minWidth: Number(result.minWidth || result.min_width || 0),
  } satisfies FastEnterConfig
}

export async function fetchMenuSpaceMode() {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/system/menu-space-mode',
  })

  return normalizeText(result.mode || 'single') || 'single'
}

export async function updateMenuSpaceMode(mode: string) {
  const result = await requestData<Record<string, unknown>>({
    method: 'PUT',
    url: '/api/v1/system/menu-space-mode',
    data: { mode },
  })

  return normalizeText(result.mode || mode) || mode
}

export async function fetchMenuSpaceList() {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: '/api/v1/system/menu-spaces',
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records) ? result.records.map((item) => normalizeMenuSpace(item)) : [],
  }
}

export async function saveMenuSpace(payload: MenuSpaceSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/system/menu-spaces',
    data: {
      space_key: payload.spaceKey,
      name: payload.name,
      description: payload.description,
      default_home_path: payload.defaultHomePath,
      is_default: payload.isDefault,
      status: payload.status,
      access_mode: payload.accessMode,
      allowed_role_codes: payload.allowedRoleCodes,
    },
  })

  return normalizeMenuSpace(result)
}

export async function initializeMenuSpaceFromDefault(spaceKey: string, force = false) {
  return requestData<Record<string, unknown>>({
    method: 'POST',
    url: `/api/v1/system/menu-spaces/${spaceKey}/initialize-default`,
    params: force ? { force: true } : undefined,
  })
}

export async function fetchMenuSpaceHostBindings() {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
    total?: number
  }>({
    method: 'GET',
    url: '/api/v1/system/menu-space-host-bindings',
  })

  return {
    total: Number(result.total || 0),
    records: Array.isArray(result.records)
      ? result.records.map((item) => normalizeMenuSpaceHostBinding(item))
      : [],
  }
}

export async function saveMenuSpaceHostBinding(payload: MenuSpaceHostBindingSavePayload) {
  const result = await requestData<Record<string, unknown>>({
    method: 'POST',
    url: '/api/v1/system/menu-space-host-bindings',
    data: {
      host: payload.host,
      space_key: payload.spaceKey,
      description: payload.description,
      scheme: payload.scheme,
      route_prefix: payload.routePrefix,
      auth_mode: payload.authMode,
      login_host: payload.loginHost || undefined,
      callback_host: payload.callbackHost || undefined,
      cookie_scope_mode: payload.cookieScopeMode,
      cookie_domain: payload.cookieDomain || undefined,
    },
  })

  return normalizeMenuSpaceHostBinding(result)
}

export async function fetchPageAccessTrace(filters: AccessTraceFilter) {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/pages/access-trace',
    params: {
      user_id: filters.userId,
      tenant_id: filters.tenantId,
      page_key: filters.pageKey,
      page_keys: filters.pageKeys,
      route_path: filters.routePath,
      space_key: filters.spaceKey,
    },
  })

  const summaryEntries = Object.entries(result).filter(([key]) => key !== 'trace' && key !== 'records')
  return {
    summary: summaryEntries.map(([key, value]) => ({
      label: key,
      value: stringifyValue(value),
    })),
    traceEntries: normalizeTraceEntries(result.trace, '轨迹'),
    recordEntries: normalizeTraceEntries(result.records, '记录'),
    raw: result,
  } satisfies AccessTraceResult
}
