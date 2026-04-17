// 参数管理 API 客户端封装。
//
// 契约来自 backend/api/openapi/domains/siteconfig，经 openapi-typescript 生成到
// @/api/v5/schema。封装目标：
//   - 将 resolve 接口的 keys/set_codes 从数组聚合为逗号分隔字符串（后端要求）。
//   - 将 POST / PUT / DELETE 的错误以 throw 抛出，调用方 await 即可。
//   - 统一返回业务体类型，屏蔽 v5Client 的 { data, error } 形态。

import { v5Client, unwrap } from '@/domains/governance/api/_shared'
import type {
  SiteConfigListResponse,
  SiteConfigLookupResponse,
  SiteConfigResolveResponse,
  SiteConfigSaveRequest,
  SiteConfigSetItemsRequest,
  SiteConfigSetListResponse,
  SiteConfigSetSaveRequest,
  SiteConfigSetSummary,
  SiteConfigSummary,
  SiteConfigManageScopeType,
  SiteConfigRuntimeScopeType
} from './types'

// ── Resolve ─────────────────────────────────────────────────────────────────

export interface ResolveSiteConfigsParams {
  scopeType?: SiteConfigRuntimeScopeType
  scopeKey?: string
  keys?: string[]
  setCodes?: string[]
}

function joinCSV(values?: string[]): string | undefined {
  if (!values || values.length === 0) return undefined
  const trimmed = values.map((s) => s.trim()).filter(Boolean)
  return trimmed.length > 0 ? trimmed.join(',') : undefined
}

export async function fetchResolveSiteConfigs(
  params: ResolveSiteConfigsParams = {}
): Promise<SiteConfigResolveResponse> {
  const query: Record<string, string> = {}
  if (params.scopeType) query.scope_type = params.scopeType
  if (params.scopeKey && params.scopeKey.length > 0) query.scope_key = params.scopeKey
  const keys = joinCSV(params.keys)
  if (keys) query.keys = keys
  const setCodes = joinCSV(params.setCodes)
  if (setCodes) query.set_codes = setCodes
  return unwrap(
    v5Client.GET('/site-configs/resolve', {
      params: { query: Object.keys(query).length ? query : {} }
    })
  )
}

export interface LookupSiteConfigParams {
  configKey: string
  scopeType?: SiteConfigRuntimeScopeType
  scopeKey?: string
}

export async function fetchLookupSiteConfig(
  params: LookupSiteConfigParams
): Promise<SiteConfigLookupResponse> {
  const query: {
    config_key: string
    scope_type?: SiteConfigRuntimeScopeType
    scope_key?: string
  } = { config_key: params.configKey }
  if (params.scopeType) query.scope_type = params.scopeType
  if (params.scopeKey && params.scopeKey.length > 0) query.scope_key = params.scopeKey
  return unwrap(
    v5Client.GET('/site-configs/lookup', {
      params: { query }
    })
  )
}

// ── 参数项 ──────────────────────────────────────────────────────────────────

export interface ListSiteConfigsParams {
  scopeType?: SiteConfigManageScopeType
  scopeKey?: string
}

export async function fetchListSiteConfigs(
  params: ListSiteConfigsParams = {}
): Promise<SiteConfigListResponse> {
  const query: Record<string, string> = {}
  if (params.scopeType) query.scope_type = params.scopeType
  if (params.scopeKey && params.scopeKey.length > 0) query.scope_key = params.scopeKey
  return unwrap(
    v5Client.GET('/site-configs', {
      params: { query }
    })
  )
}

export async function fetchUpsertSiteConfig(
  body: SiteConfigSaveRequest
): Promise<SiteConfigSummary> {
  return unwrap(v5Client.POST('/site-configs', { body }))
}

export async function fetchUpdateSiteConfig(
  id: string,
  body: SiteConfigSaveRequest
): Promise<SiteConfigSummary> {
  return unwrap(
    v5Client.PUT('/site-configs/{id}', { params: { path: { id } }, body })
  )
}

export async function fetchDeleteSiteConfig(id: string): Promise<void> {
  const { error } = await v5Client.DELETE('/site-configs/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

// ── 参数集合 ────────────────────────────────────────────────────────────────

export async function fetchListSiteConfigSets(): Promise<SiteConfigSetListResponse> {
  return unwrap(v5Client.GET('/site-configs/sets'))
}

export async function fetchUpsertSiteConfigSet(
  body: SiteConfigSetSaveRequest
): Promise<SiteConfigSetSummary> {
  return unwrap(v5Client.POST('/site-configs/sets', { body }))
}

export async function fetchUpdateSiteConfigSet(
  id: string,
  body: SiteConfigSetSaveRequest
): Promise<SiteConfigSetSummary> {
  return unwrap(
    v5Client.PUT('/site-configs/sets/{id}', { params: { path: { id } }, body })
  )
}

export async function fetchDeleteSiteConfigSet(id: string): Promise<void> {
  const { error } = await v5Client.DELETE('/site-configs/sets/{id}', {
    params: { path: { id } }
  })
  if (error) throw error
}

export async function fetchUpdateSiteConfigSetItems(
  id: string,
  body: SiteConfigSetItemsRequest
): Promise<SiteConfigSetSummary> {
  return unwrap(
    v5Client.PUT('/site-configs/sets/{id}/items', {
      params: { path: { id } },
      body
    })
  )
}
