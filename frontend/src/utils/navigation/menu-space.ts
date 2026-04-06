import type { MenuSpaceConfig, MenuSpaceDefinition, MenuSpaceHostBinding } from '@/types/config'

export const DEFAULT_MENU_SPACE_KEY = 'default'
export const SHARED_MENU_SPACE_KEY = 'shared'

export function normalizeMenuSpaceKey(value?: unknown): string {
  if (Array.isArray(value)) {
    for (let i = value.length - 1; i >= 0; i -= 1) {
      const item = `${value[i] ?? ''}`.trim()
      if (item) return item
    }
    return ''
  }
  return `${value ?? ''}`.trim()
}

export function normalizeMenuHost(host?: string): string {
  const target = normalizeMenuSpaceKey(host).toLowerCase()
  if (!target) {
    return ''
  }
  return target.split(':')[0]
}

export function normalizeMenuSpaceScheme(value?: unknown): 'http' | 'https' {
  return `${value || ''}`.trim().toLowerCase() === 'http' ? 'http' : 'https'
}

export function normalizeMenuSpaceRoutePrefix(value?: unknown): string {
  const target = `${value || ''}`.trim()
  if (!target) return ''
  const normalized = `/${target.replace(/^\/+/, '')}`.replace(/\/+/g, '/')
  return normalized === '/' ? '' : normalized.replace(/\/$/, '')
}

export function shouldUseFullMenuSpaceNavigation(
  binding: MenuSpaceHostBinding | undefined,
  currentHost?: string,
  currentProtocol?: string,
  currentPathname?: string
): boolean {
  if (!binding?.host) {
    return false
  }
  const normalizedCurrentHost =
    normalizeMenuHost(currentHost) ||
    (typeof window !== 'undefined' ? normalizeMenuHost(window.location.hostname) : '')
  const normalizedCurrentProtocol =
    `${currentProtocol || (typeof window !== 'undefined' ? window.location.protocol : 'https:')}`.replace(
      /:$/,
      ''
    )
  if (normalizeMenuHost(binding.host) !== normalizedCurrentHost) {
    return true
  }
  if (
    normalizeMenuSpaceScheme(binding.scheme) !== normalizeMenuSpaceScheme(normalizedCurrentProtocol)
  ) {
    return true
  }
  const normalizedRoutePrefix = normalizeMenuSpaceRoutePrefix(binding.routePrefix)
  if (!normalizedRoutePrefix) {
    return false
  }
  const normalizedCurrentPathname =
    normalizeMenuSpaceRoutePrefix(currentPathname) ||
    (typeof window !== 'undefined' ? normalizeMenuSpaceRoutePrefix(window.location.pathname) : '')
  if (!normalizedCurrentPathname) {
    return true
  }
  return !(
    normalizedCurrentPathname === normalizedRoutePrefix ||
    normalizedCurrentPathname.startsWith(`${normalizedRoutePrefix}/`)
  )
}

export function createFallbackMenuSpaceConfig(): MenuSpaceConfig {
  return {
    defaultSpaceKey: DEFAULT_MENU_SPACE_KEY,
    spaces: [
      {
        spaceKey: DEFAULT_MENU_SPACE_KEY,
        spaceName: '默认菜单空间',
        spaceType: 'default',
        enabled: true,
        isDefault: true,
        defaultLandingRoute: '/'
      }
    ],
    hostBindings: []
  }
}

export function resolveMenuSpaceKeyByHost(
  host: string,
  config: MenuSpaceConfig,
  fallbackKey = DEFAULT_MENU_SPACE_KEY
): string {
  const normalizedHost = normalizeMenuHost(host)
  if (!normalizedHost) {
    return fallbackKey
  }

  const matched = (config.hostBindings || []).find((binding) => {
    if (!binding?.enabled && binding?.enabled !== undefined) {
      return false
    }
    return normalizeMenuHost(binding.host) === normalizedHost
  })
  if (matched?.spaceKey) {
    return normalizeMenuSpaceKey(matched.spaceKey) || fallbackKey
  }

  return fallbackKey
}

export function resolveMenuSpaceDefinition(
  spaceKey: string,
  config: MenuSpaceConfig
): MenuSpaceDefinition | undefined {
  const normalizedKey = normalizeMenuSpaceKey(spaceKey)
  if (!normalizedKey) {
    return undefined
  }
  return (config.spaces || []).find(
    (item) => normalizeMenuSpaceKey(item.spaceKey) === normalizedKey
  )
}

export function resolveMenuSpaceHostBinding(
  spaceKey: string,
  config: MenuSpaceConfig
): MenuSpaceHostBinding | undefined {
  const normalizedSpaceKey = normalizeMenuSpaceKey(spaceKey)
  if (!normalizedSpaceKey) {
    return undefined
  }
  const bindings = (config.hostBindings || []).filter((item) => {
    if (!item) return false
    if (item.enabled === false) return false
    return normalizeMenuSpaceKey(item.spaceKey) === normalizedSpaceKey
  })
  return bindings.find((item) => Boolean(item.isPrimary)) || bindings[0]
}

export function buildMenuSpaceTargetUrl(
  binding: MenuSpaceHostBinding | undefined,
  targetPath: string
): string {
  const normalizedPath = normalizeMenuSpaceRoutePrefix(targetPath) || '/'
  if (!binding?.host) {
    return normalizedPath
  }
  const routePrefix = normalizeMenuSpaceRoutePrefix(binding.routePrefix)
  const basePath = routePrefix ? `${routePrefix}/` : '/'
  return `${normalizeMenuSpaceScheme(binding.scheme)}://${normalizeMenuHost(binding.host)}${basePath}#${normalizedPath}`
}

export function isMenuSpaceVisible(
  targetSpaceKey: string,
  currentSpaceKey: string,
  defaultSpaceKey = DEFAULT_MENU_SPACE_KEY
): boolean {
  const target = normalizeMenuSpaceKey(targetSpaceKey) || normalizeMenuSpaceKey(defaultSpaceKey)
  const current = normalizeMenuSpaceKey(currentSpaceKey) || normalizeMenuSpaceKey(defaultSpaceKey)

  if (!target) {
    return current === normalizeMenuSpaceKey(defaultSpaceKey)
  }
  if (target === SHARED_MENU_SPACE_KEY) {
    return true
  }
  return target === current
}
