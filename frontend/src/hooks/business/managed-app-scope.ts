export function normalizeManagedAppKey(value?: string | null) {
  return `${value || ''}`.trim().toLowerCase()
}

export function resolveManagedAppKey(routeAppKey?: string | null, managedAppKey?: string | null) {
  const normalizedRouteAppKey = normalizeManagedAppKey(routeAppKey)
  if (normalizedRouteAppKey) {
    return normalizedRouteAppKey
  }
  return normalizeManagedAppKey(managedAppKey)
}
