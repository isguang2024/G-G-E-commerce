export function normalizeManagedAppKey(value?: string | null) {
  return `${value || ''}`.trim().toLowerCase()
}

export function resolveManagedAppKey(...managedAppKeys: Array<string | null | undefined>) {
  for (const managedAppKey of managedAppKeys) {
    const normalized = normalizeManagedAppKey(managedAppKey)
    if (normalized) {
      return normalized
    }
  }
  return ''
}

export function resolveManagedAppStorageKey(
  explicitKey?: string | null,
  routeName?: string | symbol | null,
  routePath?: string | null
) {
  const normalizedExplicitKey = `${explicitKey || ''}`.trim()
  if (normalizedExplicitKey) {
    return normalizedExplicitKey
  }
  const routeNameKey = typeof routeName === 'string' ? routeName.trim() : ''
  if (routeNameKey) {
    return `managed-app:${routeNameKey}`
  }
  const normalizedPath = `${routePath || ''}`.trim().replace(/[^\w-]+/g, ':')
  if (normalizedPath) {
    return `managed-app:${normalizedPath}`
  }
  return 'managed-app:default'
}
