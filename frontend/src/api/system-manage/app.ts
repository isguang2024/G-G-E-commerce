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
  toV5Body,
  toV5ListResponse,
  toV5Record,
  type V5Query,
  type V5RequestBody
} from './_shared'

export async function fetchGetViewPages(force = false) {
  const query: V5Query<'/system/view-pages', 'get'> = force ? { force: '1' } : {}
  const raw = toV5Record(
    await unwrap(
    v5Client.GET('/system/view-pages', {
      params: { query }
    })
  )
  )
  return {
    pages: (raw?.pages || []) as Array<{ filePath: string; componentPath: string }>,
    refreshed: Boolean(raw?.refreshed),
    refreshedAt: `${raw?.refreshed_at || raw?.refreshedAt || ''}`
  }
}

export async function fetchGetFastEnterConfig() {
  const res = await unwrap(v5Client.GET('/system/fast-enter'))
  return normalizeFastEnterConfig(res)
}

export async function fetchUpdateFastEnterConfig(data: Api.SystemManage.FastEnterConfig) {
  const body: V5RequestBody<'/system/fast-enter', 'put'> = toV5Body(data)
  const res = await unwrap(v5Client.PUT('/system/fast-enter', { body }))
  return normalizeFastEnterConfig(res)
}

export async function fetchGetCurrentMenuSpace(spaceKey: string | undefined, appKey: string) {
  const query: V5Query<'/system/menu-spaces/current', 'get'> = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res = await unwrap(
    v5Client.GET('/system/menu-spaces/current', { params: { query } })
  )
  return {
    space: normalizeMenuSpace(res?.space || {}),
    binding: res?.binding ? normalizeMenuSpaceHostBinding(res.binding) : undefined,
    resolvedBy: `${res?.resolved_by || res?.resolvedBy || ''}`.trim(),
    requestHost: `${res?.request_host || res?.requestHost || ''}`.trim(),
    accessGranted: Boolean(res?.access_granted ?? res?.accessGranted ?? true)
  }
}

export async function fetchGetMenuSpaceMode(appKey: string) {
  const query: V5Query<'/system/menu-space-mode', 'get'> = { app_key: appKey }
  const res = await unwrap(v5Client.GET('/system/menu-space-mode', { params: { query } }))
  return { mode: `${res?.mode || 'single'}`.trim() || 'single' }
}

export async function fetchUpdateMenuSpaceMode(appKey: string, mode: string) {
  const body: V5RequestBody<'/system/menu-space-mode', 'put'> = { app_key: appKey, mode }
  const res = toV5Record(await unwrap(v5Client.PUT('/system/menu-space-mode', { body })))
  return { mode: `${res?.mode || mode || 'single'}`.trim() || 'single' }
}

export async function fetchGetApps() {
  const res = toV5ListResponse(await unwrap(v5Client.GET('/system/apps')))
  return {
    records: (res.records || []).map(normalizeApp),
    total: Number(res.total || 0)
  }
}

export async function fetchSaveApp(data: Api.SystemManage.AppSaveParams) {
  const body: V5RequestBody<'/system/apps', 'post'> = toV5Body(data)
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
    resolvedBy: `${res?.resolved_by || res?.resolvedBy || ''}`.trim(),
    requestHost: `${res?.request_host || res?.requestHost || ''}`.trim()
  }
}

export async function fetchGetAppHostBindings(appKey?: string) {
  const query: V5Query<'/system/app-host-bindings', 'get'> = appKey ? { app_key: appKey } : {}
  const res = toV5ListResponse(
    await unwrap(
    v5Client.GET('/system/app-host-bindings', {
      params: { query }
    })
  )
  )
  return {
    records: (res?.records || []).map(normalizeAppHostBinding),
    total: Number(res?.total || 0)
  }
}

export async function fetchSaveAppHostBinding(data: Api.SystemManage.AppHostBindingSaveParams) {
  const body: V5RequestBody<'/system/app-host-bindings', 'post'> = toV5Body(data)
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
  const res = toV5ListResponse(
    await unwrap(
    v5Client.GET('/system/menu-space-entry-bindings', {
      params: { query }
    })
  )
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpaceEntryBinding),
    total: Number(res?.total || 0)
  }
}

export async function fetchSaveMenuSpaceEntryBinding(
  data: Api.SystemManage.MenuSpaceEntryBindingSaveParams
) {
  const body: V5RequestBody<'/system/menu-space-entry-bindings', 'post'> = toV5Body(data)
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
  const res = toV5ListResponse(
    await unwrap(
    v5Client.GET('/system/menu-spaces', {
      params: { query }
    })
  )
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpace),
    total: res?.total || 0
  }
}

export async function fetchSaveMenuSpace(data: Api.SystemManage.MenuSpaceSaveParams) {
  const body: V5RequestBody<'/system/menu-spaces', 'post'> = toV5Body(data)
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
  const res = toV5Record(
    await unwrap(
    v5Client.POST('/system/menu-spaces/{spaceKey}/initialize-default', {
      params: { path: { spaceKey: normalizeMenuSpaceKey(spaceKey) } }
    })
  )
  )
  return {
    sourceSpaceKey: res?.source_space_key || res?.sourceSpaceKey || '',
    targetSpaceKey:
      res?.target_space_key || res?.targetSpaceKey || normalizeMenuSpaceKey(spaceKey),
    forceReinitialized: Boolean(res?.force_reinitialized ?? res?.forceReinitialized ?? false),
    clearedMenuCount: Number(res?.cleared_menu_count ?? res?.clearedMenuCount ?? 0),
    clearedPageCount: Number(res?.cleared_page_count ?? res?.clearedPageCount ?? 0),
    clearedPackageMenuLinkCount: Number(
      res?.cleared_package_menu_link_count ?? res?.clearedPackageMenuLinkCount ?? 0
    ),
    createdMenuCount: Number(res?.created_menu_count ?? res?.createdMenuCount ?? 0),
    createdPageCount: Number(res?.created_page_count ?? res?.createdPageCount ?? 0),
    createdPackageMenuLinkCount: Number(
      res?.created_package_menu_link_count ?? res?.createdPackageMenuLinkCount ?? 0
    )
  }
}

export async function fetchGetMenuSpaceHostBindings(appKey: string) {
  void appKey
  const res = toV5ListResponse(
    await unwrap(
    v5Client.GET('/system/menu-space-host-bindings', {
    })
  )
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpaceHostBinding),
    total: res?.total || 0
  }
}

export async function fetchSaveMenuSpaceHostBinding(
  data: Api.SystemManage.MenuSpaceHostBindingSaveParams
) {
  const body: V5RequestBody<'/system/menu-space-host-bindings', 'post'> = toV5Body(data)
  const res = await unwrap(v5Client.POST('/system/menu-space-host-bindings', { body }))
  return normalizeMenuSpaceHostBinding(res)
}
