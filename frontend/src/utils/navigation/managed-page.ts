type ManagedPageRouteLike = {
  pageKey?: string
  routePath?: string
  activeMenuPath?: string
  parentMenuId?: string
  parentPageKey?: string
}

interface ResolveManagedPageRouteOptions<T extends ManagedPageRouteLike> {
  getPageByKey: (pageKey: string) => T | undefined
  getMenuPathById?: (menuId: string) => string | undefined
}

function normalizeValue(value?: unknown): string {
  if (Array.isArray(value)) {
    for (let i = value.length - 1; i >= 0; i -= 1) {
      const item = `${value[i] ?? ''}`.trim()
      if (item) return item
    }
    return ''
  }
  return `${value ?? ''}`.trim()
}

function isHttpPath(path: string): boolean {
  return /^https?:\/\//i.test(path)
}

export function normalizeManagedPagePath(path?: unknown): string {
  const target = normalizeValue(path)
  if (!target) {
    return ''
  }
  if (isHttpPath(target)) {
    return target
  }
  const normalized = `/${target.replace(/^\/+/, '')}`.replace(/\/+/g, '/')
  return normalized !== '/' ? normalized.replace(/\/$/, '') : normalized
}

export function isSingleSegmentManagedPagePath(path?: unknown): boolean {
  const normalized = normalizeValue(path).replace(/^\/+/, '').replace(/\/+$/, '')
  return Boolean(normalized) && !normalized.includes('/')
}

export function joinManagedPagePath(basePath?: unknown, segment?: unknown): string {
  const left = normalizeManagedPagePath(basePath)
  const right = normalizeValue(segment).replace(/^\/+/, '')
  if (!right) {
    return left
  }
  if (!left) {
    return normalizeManagedPagePath(right)
  }
  return normalizeManagedPagePath(`${left}/${right}`)
}

function resolveManagedPageBasePath<T extends ManagedPageRouteLike>(
  page: T,
  options: ResolveManagedPageRouteOptions<T>,
  seen: Set<string>
): string {
  const explicitActivePath = normalizeManagedPagePath(page.activeMenuPath)
  if (explicitActivePath) {
    return explicitActivePath
  }

  const parentPageKey = normalizeValue(page.parentPageKey)
  if (parentPageKey) {
    const parentPage = options.getPageByKey(parentPageKey)
    if (parentPage) {
      const parentFullPath = resolveManagedPageRoutePath(parentPage, options, seen)
      if (parentFullPath) {
        return parentFullPath
      }
    }
  }

  const parentMenuId = normalizeValue(page.parentMenuId)
  if (!parentMenuId) {
    return ''
  }
  return normalizeManagedPagePath(options.getMenuPathById?.(parentMenuId))
}

export function resolveManagedPageRoutePath<T extends ManagedPageRouteLike>(
  page: T,
  options: ResolveManagedPageRouteOptions<T>,
  seen = new Set<string>()
): string {
  const pageKey = normalizeValue(page.pageKey)
  if (pageKey) {
    if (seen.has(pageKey)) {
      return ''
    }
    seen.add(pageKey)
  }

  const rawRoutePath = normalizeValue(page.routePath)
  if (!rawRoutePath) {
    return resolveManagedPageBasePath(page, options, seen)
  }
  if (isHttpPath(rawRoutePath)) {
    return rawRoutePath
  }

  const treatAsAbsolute =
    rawRoutePath.startsWith('/') && !isSingleSegmentManagedPagePath(rawRoutePath)
  if (treatAsAbsolute) {
    return normalizeManagedPagePath(rawRoutePath)
  }

  const segment = normalizeValue(rawRoutePath.replace(/^\/+/, ''))
  const basePath = resolveManagedPageBasePath(page, options, seen)
  if (basePath && !isHttpPath(basePath)) {
    return joinManagedPagePath(basePath, segment)
  }
  return normalizeManagedPagePath(segment)
}
