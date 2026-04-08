import {
  v5Client,
  unwrap,
  normalizeApp,
  normalizeAppHostBinding,
  normalizeMenuSpace,
  normalizeMenuSpaceHostBinding,
  normalizeMenuSpaceKey,
  normalizeFastEnterConfig
} from './_shared'

export async function fetchGetViewPages(force = false) {
  const raw = (await unwrap(
    v5Client.GET('/system/view-pages', {
      params: { query: (force ? { force: 1 } : {}) as any }
    })
  )) as any
  return {
    pages: (raw?.pages || []) as Array<{ filePath: string; componentPath: string }>,
    refreshed: Boolean(raw?.refreshed),
    refreshedAt: raw?.refreshedAt || raw?.refreshed_at || ''
  }
}

export async function fetchGetFastEnterConfig() {
  const res: any = await unwrap(
    v5Client.GET('/system/fast-enter', { params: { query: {} as any } })
  )
  return normalizeFastEnterConfig(res)
}

export async function fetchUpdateFastEnterConfig(data: Api.SystemManage.FastEnterConfig) {
  const res: any = await unwrap(
    v5Client.PUT('/system/fast-enter', { body: data as any })
  )
  return normalizeFastEnterConfig(res)
}

export async function fetchGetCurrentMenuSpace(spaceKey: string | undefined, appKey: string) {
  const query: any = {}
  if (spaceKey) query.space_key = normalizeMenuSpaceKey(spaceKey)
  if (appKey) query.app_key = appKey
  const res: any = await unwrap(
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
  const res: any = await unwrap(
    v5Client.GET('/system/menu-space-mode', { params: { query: { app_key: appKey } as any } })
  )
  return { mode: `${res?.mode || 'single'}`.trim() || 'single' }
}

export async function fetchUpdateMenuSpaceMode(appKey: string, mode: string) {
  const res: any = await unwrap(
    v5Client.PUT('/system/menu-space-mode', { body: { app_key: appKey, mode } as any })
  )
  return { mode: `${res?.mode || 'single'}`.trim() || 'single' }
}

export async function fetchGetApps() {
  const res: any = await unwrap(v5Client.GET('/system/apps', { params: { query: {} as any } }))
  return {
    records: (res?.records || []).map(normalizeApp),
    total: Number(res?.total || 0)
  }
}

export async function fetchSaveApp(data: Api.SystemManage.AppSaveParams) {
  const res: any = await unwrap(v5Client.POST('/system/apps', { body: data as any }))
  return normalizeApp(res)
}

export async function fetchGetCurrentApp(appKey?: string) {
  const res: any = await unwrap(
    v5Client.GET('/system/apps/current', {
      params: { query: (appKey ? { app_key: appKey } : {}) as any }
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
  const res: any = await unwrap(
    v5Client.GET('/system/app-host-bindings', {
      params: { query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    records: (res?.records || []).map(normalizeAppHostBinding),
    total: Number(res?.total || 0)
  }
}

export async function fetchSaveAppHostBinding(data: Api.SystemManage.AppHostBindingSaveParams) {
  const res: any = await unwrap(
    v5Client.POST('/system/app-host-bindings', { body: data as any })
  )
  return normalizeAppHostBinding(res)
}

export async function fetchGetMenuSpaces(appKey: string) {
  const res: any = await unwrap(
    v5Client.GET('/system/menu-spaces', {
      params: { query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpace),
    total: res?.total || 0
  }
}

export async function fetchSaveMenuSpace(data: Api.SystemManage.MenuSpaceSaveParams) {
  const res: any = await unwrap(
    v5Client.POST('/system/menu-spaces', { body: data as any })
  )
  return normalizeMenuSpace(res)
}

export async function fetchInitializeMenuSpaceFromDefault(
  appKey: string,
  spaceKey: string,
  force = false
) {
  const query: any = { app_key: appKey }
  if (force) query.force = true
  const res: any = await unwrap(
    v5Client.POST('/system/menu-spaces/{spaceKey}/initialize-default', {
      params: { path: { spaceKey: normalizeMenuSpaceKey(spaceKey) }, query }
    })
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
  const res: any = await unwrap(
    v5Client.GET('/system/menu-space-host-bindings', {
      params: { query: (appKey ? { app_key: appKey } : {}) as any }
    })
  )
  return {
    records: (res?.records || []).map(normalizeMenuSpaceHostBinding),
    total: res?.total || 0
  }
}

export async function fetchSaveMenuSpaceHostBinding(
  data: Api.SystemManage.MenuSpaceHostBindingSaveParams
) {
  const res: any = await unwrap(
    v5Client.POST('/system/menu-space-host-bindings', { body: data as any })
  )
  return normalizeMenuSpaceHostBinding(res)
}
