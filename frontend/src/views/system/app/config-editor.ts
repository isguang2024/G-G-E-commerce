export type JsonObject = Record<string, any>

export interface CapabilityFormState {
  security: {
    corsOrigins: string[]
    csp: string
  }
}

const EDITABLE_CAPABILITY_KEYS = new Set([
  'auth',
  'cors_origins',
  'csp'
])

const DEPRECATED_CAPABILITY_FIELDS: Record<string, string[]> = {
  auth: ['session_mode', 'sessionMode'],
  routing: [
    'entry_mode',
    'entryMode',
    'route_prefix',
    'routePrefix',
    'supports_public_runtime',
    'supportsPublicRuntime'
  ],
  runtime: ['kind', 'supports_worktab', 'supportsWorktab'],
  navigation: [
    'supports_multi_space',
    'supportsMultiSpace',
    'default_landing_mode',
    'defaultLandingMode',
    'supports_space_badges',
    'supportsSpaceBadges'
  ],
  integration: ['supports_broadcast_channel', 'supportsBroadcastChannel']
}

function isPlainObject(value: unknown): value is JsonObject {
  return Boolean(value) && !Array.isArray(value) && typeof value === 'object'
}

function cloneObject<T extends JsonObject>(value?: T | null): T {
  if (!isPlainObject(value)) {
    return {} as T
  }
  return JSON.parse(JSON.stringify(value)) as T
}

function cleanObject<T extends JsonObject>(value: T) {
  return Object.fromEntries(
    Object.entries(value).filter(([, current]) => {
      if (current == null) return false
      if (Array.isArray(current)) return current.length > 0
      if (typeof current === 'string') return current.trim().length > 0
      if (isPlainObject(current)) return Object.keys(current).length > 0
      return true
    })
  ) as T
}

function toTrimmedString(value: unknown) {
  return `${value ?? ''}`.trim()
}

function normalizeStringArray(value: unknown) {
  if (Array.isArray(value)) {
    return Array.from(
      new Set(
        value
          .map((item) => toTrimmedString(item))
          .filter(Boolean)
      )
    )
  }
  const single = toTrimmedString(value)
  return single ? [single] : []
}

export function formatJsonObject(value?: JsonObject) {
  try {
    return JSON.stringify(
      value && isPlainObject(value) && Object.keys(value).length ? value : {},
      null,
      2
    )
  } catch {
    return '{}'
  }
}

export function parseJSONObjectText(rawText: string, label: string) {
  const raw = `${rawText || ''}`.trim()
  if (!raw) return {}
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    throw new Error(`${label}不是合法 JSON`)
  }
  if (!isPlainObject(parsed)) {
    throw new Error(`${label}必须是 JSON 对象`)
  }
  return parsed
}

export function omitEditableCapabilitySections(value?: JsonObject) {
  const source = cloneObject(value)
  for (const key of EDITABLE_CAPABILITY_KEYS) {
    delete source[key]
  }
  return source
}

export function pickEditableCapabilitySections(value?: JsonObject) {
  const source = isPlainObject(value) ? value : {}
  const next: JsonObject = {}
  const corsOrigins = normalizeStringArray(source.cors_origins)
  const csp = toTrimmedString(source.csp)
  if (corsOrigins.length) next.cors_origins = corsOrigins
  if (csp) next.csp = csp
  return next
}

export function omitDeprecatedCapabilityFields(value?: JsonObject) {
  const source = cloneObject(value)
  for (const [sectionKey, fieldKeys] of Object.entries(DEPRECATED_CAPABILITY_FIELDS)) {
    const section = source[sectionKey]
    if (!isPlainObject(section)) continue
    for (const fieldKey of fieldKeys) {
      delete section[fieldKey]
    }
    if (!Object.keys(section).length) {
      delete source[sectionKey]
    }
  }
  return source
}

export function createCapabilityFormState(value?: JsonObject): CapabilityFormState {
  return {
    security: {
      corsOrigins: normalizeStringArray(value?.cors_origins),
      csp: toTrimmedString(value?.csp)
    }
  }
}

export function serializeCapabilityFormState(form: CapabilityFormState) {
  return cleanObject({
    ...(form.security.corsOrigins.length
      ? { cors_origins: normalizeStringArray(form.security.corsOrigins) }
      : {}),
    ...(form.security.csp.trim() ? { csp: form.security.csp.trim() } : {})
  })
}

export function summarizeManagedCapabilities(value?: JsonObject) {
  const source = pickEditableCapabilitySections(value)
  const summary: string[] = []
  const corsCount = normalizeStringArray(source.cors_origins).length
  if (corsCount) summary.push(`安全 ${corsCount} 个来源`)
  if (toTrimmedString(source.csp)) summary.push('已声明 CSP')
  return summary
}
