import {
  request,
  APP_BASE,
  APP_HOST_BINDING_BASE,
  SYSTEM_BASE,
  normalizeApp,
  normalizeAppHostBinding,
  normalizeMenuSpace,
  normalizeMenuSpaceHostBinding,
  normalizeMenuSpaceKey,
  normalizeFastEnterConfig
} from './_shared'

export function fetchGetViewPages(force = false) {
  return request.get<{
    pages: Array<{ filePath: string; componentPath: string }>
    refreshed: boolean
    refreshedAt: string
  }>({
    url: `${SYSTEM_BASE}/view-pages`,
    params: force ? { force: 1 } : undefined
  })
}

export function fetchGetFastEnterConfig() {
  return request
    .get<Api.SystemManage.FastEnterConfig>({
      url: `${SYSTEM_BASE}/fast-enter`
    })
    .then((res) => normalizeFastEnterConfig(res))
}

export function fetchUpdateFastEnterConfig(data: Api.SystemManage.FastEnterConfig) {
  return request
    .put<Api.SystemManage.FastEnterConfig>({
      url: `${SYSTEM_BASE}/fast-enter`,
      data
    })
    .then((res) => normalizeFastEnterConfig(res))
}

export function fetchGetCurrentMenuSpace(spaceKey: string | undefined, appKey: string) {
  return request
    .get<Api.SystemManage.CurrentMenuSpaceResponse>({
      url: `${SYSTEM_BASE}/menu-spaces/current`,
      params:
        spaceKey || appKey
          ? {
              ...(spaceKey ? { space_key: normalizeMenuSpaceKey(spaceKey) } : {}),
              ...(appKey ? { app_key: appKey } : {})
            }
          : undefined
    })
    .then((res: any) => ({
      space: normalizeMenuSpace(res?.space || {}),
      binding: res?.binding ? normalizeMenuSpaceHostBinding(res.binding) : undefined,
      resolvedBy: `${res?.resolved_by || res?.resolvedBy || ''}`.trim(),
      requestHost: `${res?.request_host || res?.requestHost || ''}`.trim(),
      accessGranted: Boolean(res?.access_granted ?? res?.accessGranted ?? true)
    }))
}

export function fetchGetMenuSpaceMode(appKey: string) {
  return request
    .get<Api.SystemManage.MenuSpaceModeResponse>({
      url: `${SYSTEM_BASE}/menu-space-mode`,
      params: { app_key: appKey }
    })
    .then((res: any) => ({
      mode: `${res?.mode || 'single'}`.trim() || 'single'
    }))
}

export function fetchUpdateMenuSpaceMode(appKey: string, mode: string) {
  return request
    .put<Api.SystemManage.MenuSpaceModeResponse>({
      url: `${SYSTEM_BASE}/menu-space-mode`,
      data: { app_key: appKey, mode }
    })
    .then((res: any) => ({
      mode: `${res?.mode || 'single'}`.trim() || 'single'
    }))
}

export function fetchGetApps() {
  return request
    .get<{ records: Api.SystemManage.AppItem[]; total: number }>({
      url: APP_BASE
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeApp),
      total: Number(res?.total || 0)
    }))
}

export function fetchSaveApp(data: Api.SystemManage.AppSaveParams) {
  return request
    .post<Api.SystemManage.AppItem>({
      url: APP_BASE,
      data
    })
    .then((res) => normalizeApp(res))
}

export function fetchGetCurrentApp(appKey?: string) {
  return request
    .get<Api.SystemManage.CurrentAppResponse>({
      url: `${APP_BASE}/current`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res: any) => ({
      app: normalizeApp(res?.app || {}),
      binding: res?.binding ? normalizeAppHostBinding(res.binding) : undefined,
      resolvedBy: `${res?.resolved_by || res?.resolvedBy || ''}`.trim(),
      requestHost: `${res?.request_host || res?.requestHost || ''}`.trim()
    }))
}

export function fetchGetAppHostBindings(appKey?: string) {
  return request
    .get<{ records: Api.SystemManage.AppHostBindingItem[]; total: number }>({
      url: APP_HOST_BINDING_BASE,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeAppHostBinding),
      total: Number(res?.total || 0)
    }))
}

export function fetchSaveAppHostBinding(data: Api.SystemManage.AppHostBindingSaveParams) {
  return request
    .post<Api.SystemManage.AppHostBindingItem>({
      url: APP_HOST_BINDING_BASE,
      data
    })
    .then((res) => normalizeAppHostBinding(res))
}

export function fetchGetMenuSpaces(appKey: string) {
  return request
    .get<{ records: Api.SystemManage.MenuSpaceItem[]; total: number }>({
      url: `${SYSTEM_BASE}/menu-spaces`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeMenuSpace),
      total: res?.total || 0
    }))
}

export function fetchSaveMenuSpace(data: Api.SystemManage.MenuSpaceSaveParams) {
  return request
    .post<Api.SystemManage.MenuSpaceItem>({
      url: `${SYSTEM_BASE}/menu-spaces`,
      data
    })
    .then((res) => normalizeMenuSpace(res))
}

export function fetchInitializeMenuSpaceFromDefault(
  appKey: string,
  spaceKey: string,
  force = false
) {
  return request
    .post<Api.SystemManage.MenuSpaceInitializeResult>({
      url: `${SYSTEM_BASE}/menu-spaces/${normalizeMenuSpaceKey(spaceKey)}/initialize-default`,
      params: force ? { app_key: appKey, force: true } : { app_key: appKey }
    })
    .then((res: any) => ({
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
    }))
}

export function fetchGetMenuSpaceHostBindings(appKey: string) {
  return request
    .get<{ records: Api.SystemManage.MenuSpaceHostBindingItem[]; total: number }>({
      url: `${SYSTEM_BASE}/menu-space-host-bindings`,
      params: appKey ? { app_key: appKey } : undefined
    })
    .then((res) => ({
      records: (res?.records || []).map(normalizeMenuSpaceHostBinding),
      total: res?.total || 0
    }))
}

export function fetchSaveMenuSpaceHostBinding(
  data: Api.SystemManage.MenuSpaceHostBindingSaveParams
) {
  return request
    .post<Api.SystemManage.MenuSpaceHostBindingItem>({
      url: `${SYSTEM_BASE}/menu-space-host-bindings`,
      data
    })
    .then((res) => normalizeMenuSpaceHostBinding(res))
}
