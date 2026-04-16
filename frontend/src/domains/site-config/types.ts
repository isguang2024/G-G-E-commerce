// site-config 类型别名（从 OpenAPI schema 派生）。
// 这里只做聚合与重导出，避免前端业务代码直接依赖 components['schemas'] 长路径。

import type { components } from '@/api/v5/schema'

export type SiteConfigSummary = components['schemas']['SiteConfigSummary']
export type SiteConfigListResponse = components['schemas']['SiteConfigListResponse']
export type SiteConfigSaveRequest = components['schemas']['SiteConfigSaveRequest']

export type SiteConfigSetSummary = components['schemas']['SiteConfigSetSummary']
export type SiteConfigSetListResponse = components['schemas']['SiteConfigSetListResponse']
export type SiteConfigSetSaveRequest = components['schemas']['SiteConfigSetSaveRequest']
export type SiteConfigSetItemsRequest = components['schemas']['SiteConfigSetItemsRequest']

export type SiteConfigResolvedItem = components['schemas']['SiteConfigResolvedItem']
export type SiteConfigResolveResponse = components['schemas']['SiteConfigResolveResponse']

// value_type 枚举（与后端对齐）。
export const SITE_CONFIG_VALUE_TYPES = [
  'string',
  'number',
  'bool',
  'image',
  'json',
  'svg'
] as const
export type SiteConfigValueType = (typeof SITE_CONFIG_VALUE_TYPES)[number]

// source 枚举。
export type SiteConfigSource = 'app' | 'global' | 'default'

// 常量：空字符串代表全局作用域；'__all__' 代表列出所有作用域。
export const SITE_CONFIG_GLOBAL_APP_KEY = ''
export const SITE_CONFIG_ALL_SCOPES = '__all__'

// 常用的值容器：后端统一用 JSON 对象，string 类型约定 value 字段。
export interface SiteConfigScalarValue {
  value?: string | number | boolean
  [key: string]: unknown
}

export interface SiteConfigImageValue {
  url?: string
  [key: string]: unknown
}

// 工具：从 ResolvedItem 里读取 string。
export function readResolvedString(
  item: SiteConfigResolvedItem | undefined,
  fallback = ''
): string {
  if (!item) return fallback
  const raw = (item.value as SiteConfigScalarValue | undefined)?.value
  if (typeof raw === 'string') return raw
  if (typeof raw === 'number' || typeof raw === 'boolean') return String(raw)
  return fallback
}

export function readResolvedNumber(
  item: SiteConfigResolvedItem | undefined,
  fallback = 0
): number {
  if (!item) return fallback
  const raw = (item.value as SiteConfigScalarValue | undefined)?.value
  if (typeof raw === 'number') return raw
  if (typeof raw === 'string') {
    const n = Number(raw)
    return Number.isFinite(n) ? n : fallback
  }
  return fallback
}

export function readResolvedBool(
  item: SiteConfigResolvedItem | undefined,
  fallback = false
): boolean {
  if (!item) return fallback
  const raw = (item.value as SiteConfigScalarValue | undefined)?.value
  if (typeof raw === 'boolean') return raw
  if (typeof raw === 'string') return raw === 'true'
  if (typeof raw === 'number') return raw !== 0
  return fallback
}

export function readResolvedImage(
  item: SiteConfigResolvedItem | undefined,
  fallback = ''
): string {
  if (!item) return fallback
  const url = (item.value as SiteConfigImageValue | undefined)?.url
  return typeof url === 'string' && url.length > 0 ? url : fallback
}
