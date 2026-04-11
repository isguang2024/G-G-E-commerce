import { normalizeManagedAppKey } from '@/hooks/business/managed-app-scope'

export const ACTIVE_APP_SCOPE_STORAGE_KEY = 'gg:active-app-key'
export const APP_SCOPE_GLOBAL = 'global'

function isBrowser(): boolean {
  return typeof window !== 'undefined'
}

export function normalizeAppScopeKey(value?: string | null): string {
  const normalized = normalizeManagedAppKey(value)
  return normalized || APP_SCOPE_GLOBAL
}

function readAppContextStoreFallback(): string {
  if (!isBrowser()) return APP_SCOPE_GLOBAL

  const keys = Object.keys(window.localStorage)
  const appContextStorageKey = keys.find((key) => key.endsWith('-appContextStore'))
  if (!appContextStorageKey) return APP_SCOPE_GLOBAL

  try {
    const payload = window.localStorage.getItem(appContextStorageKey)
    if (!payload) return APP_SCOPE_GLOBAL
    const parsed = JSON.parse(payload) as { runtimeAppKey?: string; managedAppKey?: string }
    return normalizeAppScopeKey(parsed?.managedAppKey || parsed?.runtimeAppKey)
  } catch {
    return APP_SCOPE_GLOBAL
  }
}

export function readActiveAppScopeKey(): string {
  if (!isBrowser()) return APP_SCOPE_GLOBAL

  const sessionValue = normalizeAppScopeKey(
    window.sessionStorage.getItem(ACTIVE_APP_SCOPE_STORAGE_KEY)
  )
  if (sessionValue !== APP_SCOPE_GLOBAL) return sessionValue

  const localValue = normalizeAppScopeKey(window.localStorage.getItem(ACTIVE_APP_SCOPE_STORAGE_KEY))
  if (localValue !== APP_SCOPE_GLOBAL) return localValue

  return readAppContextStoreFallback()
}

export function writeActiveAppScopeKey(value?: string | null): string {
  const next = normalizeAppScopeKey(value)
  if (!isBrowser()) return next
  window.sessionStorage.setItem(ACTIVE_APP_SCOPE_STORAGE_KEY, next)
  window.localStorage.setItem(ACTIVE_APP_SCOPE_STORAGE_KEY, next)
  return next
}
