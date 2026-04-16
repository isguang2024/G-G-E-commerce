// site-config 管理面 API 客户端封装。
//
// 契约来自 backend/api/openapi/domains/siteconfig，经 openapi-typescript 生成到
// @/api/v5/schema。封装目标：
//   - 将 resolve 接口的 keys/set_codes 从数组聚合为逗号分隔字符串（后端要求）。
//   - 将 POST / PUT / DELETE 的错误以 throw 抛出，调用方 await 即可。
//   - 统一返回业务体类型，屏蔽 v5Client 的 { data, error } 形态。

import { v5Client, unwrap } from '@/domains/governance/api/_shared'
import type {
  SiteConfigListResponse,
  SiteConfigResolveResponse,
  SiteConfigSaveRequest,
  SiteConfigSetItemsRequest,
  SiteConfigSetListResponse,
  SiteConfigSetSaveRequest,
  SiteConfigSetSummary,
  SiteConfigSummary
} from './types'

// ── Resolve ─────────────────────────────────────────────────────────────────

export interface ResolveSiteConfigsParams {
  appKey?: string
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
  if (params.appKey && params.appKey.length > 0) query.app_key = params.appKey
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

// ── Configs ─────────────────────────────────────────────────────────────────

/**
 * appKey:
 *   - undefined: 省略查询，后端默认仅返回全局
 *   - '': 显式全局
 *   - '__all__': 所有作用域
 *   - 其它: 指定 app_key
 */
export async function fetchListSiteConfigs(
  appKey?: string
): Promise<SiteConfigListResponse> {
  return unwrap(
    v5Client.GET('/site-configs', {
      params: { query: appKey === undefined ? {} : { app_key: appKey } }
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

// ── Sets ────────────────────────────────────────────────────────────────────

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
