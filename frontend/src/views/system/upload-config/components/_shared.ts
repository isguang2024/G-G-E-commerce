import { ElMessage } from 'element-plus'
import {
  getDriverExtraDefaults,
  getDriverExtraSections,
  type DriverExtraField,
  type DriverExtraScope,
  type DriverExtraSection,
  type StorageDriver
} from '@/domains/upload-config/driver-extra-registry'

// ─── Types ───────────────────────────────────────────────────────────────────

export type DriverExtraValueMap = Record<string, any>

export interface DriverExtraDraft {
  values: DriverExtraValueMap
  objectText: Record<string, string>
  customText: string
  activePanels: string[]
}

// ─── Text helpers ────────────────────────────────────────────────────────────

export function splitCommaValues(value: string): string[] {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

export function normalizeObjectValue<T extends Record<string, unknown>>(
  value: unknown
): T | undefined {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return undefined
  return { ...(value as T) }
}

export function stringifyJsonEditor(value: unknown): string {
  const normalized = normalizeObjectValue<Record<string, unknown>>(value)
  return normalized ? JSON.stringify(normalized, null, 2) : ''
}

export function parseJsonEditor(
  value: string,
  label: string
): Record<string, unknown> | undefined | null {
  const text = value.trim()
  if (!text) return undefined
  try {
    const parsed = JSON.parse(text)
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
      ElMessage.warning(`${label}必须是 JSON 对象`)
      return null
    }
    return parsed as Record<string, unknown>
  } catch {
    ElMessage.warning(`${label}不是合法的 JSON`)
    return null
  }
}

export function formatSchemaConfigured(value: unknown): string {
  return normalizeObjectValue(value) ? '已配置' : '-'
}

// ─── Driver extra helpers ────────────────────────────────────────────────────

export function flattenDriverExtraFields(sections: DriverExtraSection[]): DriverExtraField[] {
  return sections.flatMap((section) => section.fields)
}

export function defaultOpenExtraPanels(
  sections: DriverExtraSection[],
  customText = ''
): string[] {
  const panels = sections.filter((section) => section.defaultOpen).map((section) => section.key)
  if (customText.trim()) panels.push('custom')
  return panels
}

export function buildDriverExtraDraft(
  driver: StorageDriver | '' | undefined,
  scope: DriverExtraScope,
  value?: unknown
): DriverExtraDraft {
  const sections = getDriverExtraSections(driver, scope)
  const fields = flattenDriverExtraFields(sections)
  const defaults = getDriverExtraDefaults(driver, scope)
  const source = normalizeObjectValue<DriverExtraValueMap>(value) || {}
  const knownKeys = new Set(fields.map((field) => field.key))
  const values: DriverExtraValueMap = { ...defaults }
  const custom: Record<string, unknown> = {}
  for (const [key, item] of Object.entries(source)) {
    if (knownKeys.has(key)) {
      values[key] = item
    } else {
      custom[key] = item
    }
  }
  const objectText: Record<string, string> = {}
  for (const field of fields) {
    if (field.type !== 'object') continue
    const current = values[field.key]
    if (current === undefined || current === null || current === '') {
      objectText[field.key] = ''
      continue
    }
    objectText[field.key] =
      typeof current === 'string' ? current : JSON.stringify(current, null, 2)
  }
  const customText = stringifyJsonEditor(custom)
  return {
    values,
    objectText,
    customText,
    activePanels: defaultOpenExtraPanels(sections, customText)
  }
}

export function buildDriverExtraBody(
  label: string,
  driver: StorageDriver | '' | undefined,
  scope: DriverExtraScope,
  values: DriverExtraValueMap,
  objectText: Record<string, string>,
  customText: string
): Record<string, unknown> | undefined | null {
  const sections = getDriverExtraSections(driver, scope)
  const fields = flattenDriverExtraFields(sections)
  const knownKeys = new Set(fields.map((field) => field.key))
  const payload: Record<string, unknown> = {}

  for (const field of fields) {
    if (field.type === 'object') {
      const parsed = parseJsonEditor(objectText[field.key] || '', `${label}${field.label}`)
      if (parsed === null) return null
      if (parsed) payload[field.key] = parsed
      continue
    }
    const current = values[field.key]
    if (field.type === 'boolean') {
      if (typeof current === 'boolean') payload[field.key] = current
      continue
    }
    if (field.type === 'number') {
      if (
        current !== undefined &&
        current !== null &&
        current !== '' &&
        Number.isFinite(Number(current))
      ) {
        payload[field.key] = Number(current)
      }
      continue
    }
    const text = typeof current === 'string' ? current.trim() : String(current ?? '').trim()
    if (text) payload[field.key] = text
  }

  const custom = parseJsonEditor(customText, `${label}自定义扩展参数`)
  if (custom === null) return null
  if (custom) {
    for (const key of Object.keys(custom)) {
      if (knownKeys.has(key)) {
        ElMessage.warning(
          `${label}自定义扩展参数中的 ${key} 已由结构化字段托管，请直接使用上方表单`
        )
        return null
      }
    }
    Object.assign(payload, custom)
  }
  return Object.keys(payload).length ? payload : undefined
}

export function formatDriverExtraValue(value: unknown): string {
  if (value === undefined || value === null || value === '') return ''
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

export function formatDriverExtraFieldTip(field: DriverExtraField): string {
  const parts: string[] = []
  if (field.description) parts.push(field.description)
  if (field.defaultValue !== undefined) {
    parts.push(`推荐默认：${formatDriverExtraValue(field.defaultValue)}`)
  }
  return parts.join(' ')
}

export function readExtraStringValue(values: DriverExtraValueMap, key: string): string {
  const value = values[key]
  return typeof value === 'string'
    ? value
    : value === undefined || value === null
      ? ''
      : String(value)
}

export function readExtraNumberValue(
  values: DriverExtraValueMap,
  key: string
): number | undefined {
  const value = values[key]
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() && Number.isFinite(Number(value))) {
    return Number(value)
  }
  return undefined
}

export function readExtraBooleanValue(values: DriverExtraValueMap, key: string): boolean {
  return values[key] === true
}

export function setExtraValue(values: DriverExtraValueMap, key: string, value: unknown) {
  values[key] = value
}

// ─── Size formatter ──────────────────────────────────────────────────────────

export function formatBytes(value: number): string {
  if (!value || value <= 0) return '不限'
  const units = ['B', 'KB', 'MB', 'GB']
  let size = value
  let unit = 0
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024
    unit += 1
  }
  return `${size.toFixed(unit === 0 ? 0 : 2)} ${units[unit]}`
}

// ─── Labels (fallback only — 表格/筛选页在拿不到字典时的兜底) ───────────────

export const driverLabel: Record<string, string> = {
  local: '本地存储',
  aliyun_oss: '阿里云 OSS'
}

export const statusLabel: Record<string, string> = {
  ready: '启用',
  disabled: '停用',
  error: '异常'
}

export const statusType: Record<string, 'success' | 'info' | 'danger'> = {
  ready: 'success',
  disabled: 'info',
  error: 'danger'
}

export const visibilityLabel: Record<string, string> = {
  public: '公开',
  private: '私有'
}

export const uploadModeLabel: Record<string, string> = {
  auto: '自动选择',
  direct: '前端直传',
  relay: '后端中转',
  inherit: '继承上传配置'
}

export const visibilityOverrideLabel: Record<string, string> = {
  inherit: '继承上传配置',
  public: '公开',
  private: '私有'
}

export const filenameStrategyLabel: Record<string, string> = {
  uuid: '随机（UUID）',
  original: '保留原名',
  timestamp: '时间戳',
  hashed: '哈希'
}
