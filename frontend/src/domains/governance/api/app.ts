import {
  v5Client,
  unwrap,
  normalizeApp,
  normalizeAppHostBinding,
  normalizeMenuSpace,
  normalizeMenuSpaceHostBinding,
  normalizeMenuSpaceEntryBinding,
  normalizeMenuSpaceKey,
  normalizeFastEnterConfig,
  type V5Query,
  type V5RequestBody
} from './_shared'

export async function fetchGetViewPages(force = false) {
  const query: V5Query<'/system/view-pages', 'get'> = force ? { force: '1' } : {}
  const raw = await unwrap(
    v5Client.GET('/system/view-pages', {
      params: { query }
    })
  )
  return {
    pages: (raw?.pages || []).map((item) => ({
      filePath: item?.file_path || '',
      componentPath: item?.component_path || ''
    })),
    refreshed: Boolean(raw?.refreshed),
    refreshedAt: `${raw?.refreshed_at || ''}`
  }
}

export async function fetchGetFastEnterConfig() {
  const res = await unwrap(v5Client.GET('/system/fast-enter'))
  return normalizeFastEnterConfig(res)
}

export async function fetchUpdateFastEnterConfig(data: Api.SystemManage.FastEnterConfig) {
  const body: V5RequestBody<'/system/fast-enter', 'put'> = {
    applications: (data.applications || []).map((item) => ({
      id: item.id || '',
      name: item.name || '',
      description: item.description || '',
      icon: item.icon || '',
      iconColor: item.iconColor || '',
      enabled: Boolean(item.enabled),
      order: Number(item.order ?? 0),
      routeName: item.routeName,
      link: item.link
    })),
    quickLinks: (data.quickLinks || []).map((item) => ({
      id: item.id || '',
      name: item.name || '',
      enabled: Boolean(item.enabled),
      order: Number(item.order ?? 0),
      routeName: item.routeName,
      link: item.link
    })),
    minWidth: data.minWidth
  }
  const res = await unwrap(v5Client.PUT('/system/fast-enter', { body }))
  return normalizeFastEnterConfig(res)
}

export async function fetchGetCurrentMenuSpace(spaceKey: string | undefined, appKey: string) {
  const query: V5Query<'/system/menu-spaces/current', 'get'> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(v5Client.GET('/system/menu-spaces/current', { params: { query } }))
  return {
    space: normalizeMenuSpace(res?.space || {}),
    binding: res?.binding ? normalizeMenuSpaceHostBinding(res.binding) : undefined,
    resolvedBy: `${res?.resolved_by || ''}`.trim(),
    requestHost: `${res?.request_host || ''}`.trim(),
    accessGranted: Boolean(res?.access_granted ?? true)
  }
}

export async function fetchGetMenuSpaceMode(appKey: string) {
  const query: V5Query<'/system/menu-space-mode', 'get'> = { app_key: appKey }
  const res = await unwrap(v5Client.GET('/system/menu-space-mode', { params: { query } }))
  return { mode: `${res?.mode || 'single'}`.trim() || 'single' }
}

export async function fetchUpdateMenuSpaceMode(appKey: string, mode: string) {
  const body: V5RequestBody<'/system/menu-space-mode', 'put'> = { app_key: appKey, mode }
  const res = await unwrap(v5Client.PUT('/system/menu-space-mode', { body }))
  return { mode: `${res?.mode || mode || 'single'}`.trim() || 'single' }
}

export async function fetchGetApps() {
  const res = await unwrap(v5Client.GET('/system/apps'))
  return {
    records: (res.records || []).map(normalizeApp),
    total: Number(res.total || 0)
  }
}

export async function fetchSaveApp(data: Api.SystemManage.AppSaveParams) {
  const body: V5RequestBody<'/system/apps', 'post'> = {
    app_key: data.app_key,
    name: data.name,
    description: data.description,
    space_mode: data.space_mode,
    default_space_key: data.default_space_key,
    auth_mode: data.auth_mode,
    frontend_entry_url: data.frontend_entry_url,
    backend_entry_url: data.backend_entry_url,
    health_check_url: data.health_check_url,
    manifest_url: data.manifest_url,
    runtime_version: data.runtime_version,
    capabilities: data.capabilities,
    status: data.status,
    is_default: data.is_default,
    meta: data.meta
  }
  const res = await unwrap(v5Client.POST('/system/apps', { body }))
  return normalizeApp(res)
}

export async function fetchGetCurrentApp(appKey?: string) {
  const query: V5Query<'/system/apps/current', 'get'> = appKey ? { app_key: appKey } : {}
  const res = await unwrap(
    v5Client.GET('/system/apps/current', {
      params: { query }
    })
  )
  return {
    app: normalizeApp(res?.app || {}),
    binding: res?.binding ? normalizeAppHostBinding(res.binding) : undefined,
    resolvedBy: `${res?.resolved_by || ''}`.trim(),
    requestHost: `${res?.request_host || ''}`.trim()
  }
}

export async function fetchGetAppPreflight(appKey: string) {
  const query: V5Query<'/system/apps/preflight', 'get'> = { app_key: appKey }
  const res = await unwrap(
    v5Client.GET('/system/apps/preflight', {
      params: { query }
    })
  )
  return {
    appKey: `${res?.app_key || appKey}`.trim(),
    name: `${res?.name || ''}`.trim(),
    requestHost: `${res?.request_host || ''}`.trim(),
    manifestUrl: `${res?.manifest_url || ''}`.trim(),
    runtimeVersion: `${res?.runtime_version || ''}`.trim(),
    probeStatus: `${res?.probe_status || ''}`.trim(),
    probeTarget: `${res?.probe_target || ''}`.trim(),
    probeMessage: `${res?.probe_message || ''}`.trim(),
    probeCheckedAt: `${res?.probe_checked_at || ''}`.trim(),
    summary: {
      level: `${res?.summary?.level || 'info'}`.trim() || 'info',
      blockingCount: Number(res?.summary?.blocking_count || 0),
      warningCount: Number(res?.summary?.warning_count || 0),
      infoCount: Number(res?.summary?.info_count || 0),
      successCount: Number(res?.summary?.success_count || 0)
    },
    checks: (res?.checks || []).map((item) => ({
      key: `${item?.key || ''}`.trim(),
      title: `${item?.title || ''}`.trim(),
      level: `${item?.level || 'info'}`.trim() || 'info',
      passed: Boolean(item?.passed),
      value: `${item?.value || ''}`.trim(),
      hint: `${item?.hint || ''}`.trim()
    })),
    previewItems: (res?.preview_items || []).map((item) => ({
      label: `${item?.label || ''}`.trim(),
      value: `${item?.value || ''}`.trim(),
      hint: `${item?.hint || ''}`.trim()
    }))
  } as Api.SystemManage.AppPreflightResponse
}

export async function fetchGetAppHostBindings(appKey?: string) {
  const query: V5Query<'/system/app-host-bindings', 'get'> = appKey ? { app_key: appKey } : {}
  const res = await unwrap(
    v5Client.GET('/system/app-host-bindings', {
      params: { query }
    })
  )
  return {
    records: (res?.records || []).map(normalizeAppHostBinding),
    total: Number(res?.total || 0)
  }
}

export async function fetchSaveAppHostBinding(data: Api.SystemManage.AppHostBindingSaveParams) {
  const body: V5RequestBody<'/system/app-host-bindings', 'post'> = {
    id: data.id,
    app_key: data.app_key,
    match_type: data.match_type,
    host: data.host || '',
    path_pattern: data.path_pattern,
    priority: data.priority,
    description: data.description,
    is_primary: data.is_primary,
    default_space_key: data.default_space_key,
    status: data.status,
    meta: data.meta
  }
  const res = await unwrap(v5Client.POST('/system/app-host-bindings', { body }))
  return normalizeAppHostBinding(res)
}

export async function fetchDeleteAppHostBinding(id: string, appKey?: string) {
  const query: V5Query<'/system/app-host-bindings/{id}', 'delete'> = appKey
    ? { app_key: appKey }
    : {}
  await unwrap(
    v5Client.DELETE('/system/app-host-bindings/{id}', {
      params: { path: { id }, query }
    })
  )
}

export async function fetchGetMenuSpaceEntryBindings(appKey: string) {
  const query: V5Query<'/system/menu-space-entry-bindings', 'get'> = { app_key: appKey }
  const res = await unwrap(
    v5Client.GET('/system/menu-space-entry-bindings', {
      params: { query }
    })
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpaceEntryBinding),
    total: Number(res?.total || 0)
  }
}

export async function fetchSaveMenuSpaceEntryBinding(
  data: Api.SystemManage.MenuSpaceEntryBindingSaveParams
) {
  const body: V5RequestBody<'/system/menu-space-entry-bindings', 'post'> = {
    id: data.id,
    app_key: data.app_key,
    space_key: data.space_key,
    match_type: data.match_type,
    host: data.host || '',
    path_pattern: data.path_pattern,
    priority: data.priority,
    description: data.description,
    is_primary: data.is_primary,
    status: data.status,
    meta: data.meta
  }
  const res = await unwrap(v5Client.POST('/system/menu-space-entry-bindings', { body }))
  return normalizeMenuSpaceEntryBinding(res)
}

export async function fetchDeleteMenuSpaceEntryBinding(id: string, appKey?: string) {
  const query: V5Query<'/system/menu-space-entry-bindings/{id}', 'delete'> = appKey
    ? { app_key: appKey }
    : {}
  await unwrap(
    v5Client.DELETE('/system/menu-space-entry-bindings/{id}', {
      params: { path: { id }, query }
    })
  )
}

export async function fetchGetMenuSpaces(appKey: string) {
  const query: V5Query<'/system/menu-spaces', 'get'> = { app_key: appKey }
  const res = await unwrap(
    v5Client.GET('/system/menu-spaces', {
      params: { query }
    })
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpace),
    total: res?.total || 0
  }
}

export async function fetchSaveMenuSpace(data: Api.SystemManage.MenuSpaceSaveParams) {
  const body: V5RequestBody<'/system/menu-spaces', 'post'> = {
    app_key: data.app_key,
    space_key: data.space_key,
    name: data.name,
    description: data.description,
    default_home_path: data.default_home_path,
    is_default: data.is_default,
    status: data.status,
    access_mode: data.access_mode,
    allowed_role_codes: data.allowed_role_codes,
    meta: data.meta
  }
  const res = await unwrap(v5Client.POST('/system/menu-spaces', { body }))
  return normalizeMenuSpace(res)
}

export async function fetchInitializeMenuSpaceFromDefault(
  appKey: string,
  spaceKey: string,
  force = false
) {
  void appKey
  void force
  const res = await unwrap(
    v5Client.POST('/system/menu-spaces/{spaceKey}/initialize-default', {
      params: { path: { spaceKey: normalizeMenuSpaceKey(spaceKey) } }
    })
  )
  return {
    sourceSpaceKey: res?.source_space_key || '',
    targetSpaceKey: res?.target_space_key || normalizeMenuSpaceKey(spaceKey),
    forceReinitialized: Boolean(res?.force_reinitialized ?? false),
    clearedMenuCount: Number(res?.cleared_menu_count ?? 0),
    clearedPageCount: Number(res?.cleared_page_count ?? 0),
    clearedPackageMenuLinkCount: Number(res?.cleared_package_menu_link_count ?? 0),
    createdMenuCount: Number(res?.created_menu_count ?? 0),
    createdPageCount: Number(res?.created_page_count ?? 0),
    createdPackageMenuLinkCount: Number(res?.created_package_menu_link_count ?? 0)
  }
}

export async function fetchGetMenuSpaceHostBindings(appKey: string) {
  void appKey
  const res = await unwrap(v5Client.GET('/system/menu-space-host-bindings', {}))
  return {
    records: (res?.records || []).map(normalizeMenuSpaceHostBinding),
    total: res?.total || 0
  }
}

export async function fetchSaveMenuSpaceHostBinding(
  data: Api.SystemManage.MenuSpaceHostBindingSaveParams
) {
  const body: V5RequestBody<'/system/menu-space-host-bindings', 'post'> = {
    app_key: data.app_key,
    host: data.host,
    space_key: data.space_key,
    description: data.description,
    is_default: data.is_default,
    status: data.status,
    meta: data.meta
  }
  const res = await unwrap(v5Client.POST('/system/menu-space-host-bindings', { body }))
  return normalizeMenuSpaceHostBinding(res)
}
