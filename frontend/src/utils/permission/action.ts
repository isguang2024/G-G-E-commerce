type ActionRequirement = string

type UserActionInfo = Partial<Api.Auth.UserInfo> | null | undefined

export function buildScopedActionKey(action: string, scopeCode?: string): string {
  const { key, scope } = resolveActionKey(action)
  const normalizedScope = normalizeScopeCode(scopeCode || scope)
  return normalizedScope ? `${key}@${normalizedScope}` : key
}

export function resolveActionKey(action: ActionRequirement): { key: string; scope?: string } {
  const raw = `${action || ''}`.trim()
  const atIndex = raw.lastIndexOf('@')
  if (atIndex <= 0 || atIndex === raw.length - 1) {
    return { key: raw }
  }
  return {
    key: raw.slice(0, atIndex),
    scope: raw.slice(atIndex + 1)
  }
}

export function hasScopedActionPermission(
  userInfo: UserActionInfo,
  action: ActionRequirement,
  scopeCode?: string
): boolean {
  if (userInfo?.is_super_admin) {
    return true
  }

  const requirement = resolveActionKey(action)
  const normalizedScope = normalizeScopeCode(scopeCode || requirement.scope)
  const scopedActions = new Set(userInfo?.scoped_actions || userInfo?.scopedActions || [])
  const actions = new Set(userInfo?.actions || [])
  const hasScopedActionData = scopedActions.size > 0

  if (normalizedScope) {
    if (hasScopedActionData) {
      return scopedActions.has(`${requirement.key}@${normalizedScope}`)
    }
    return actions.has(requirement.key)
  }
  return actions.has(requirement.key)
}

function normalizeScopeCode(scopeCode?: string): string | undefined {
  const trimmed = `${scopeCode || ''}`.trim()
  if (!trimmed) {
    return undefined
  }
  if (trimmed === 'tenant') {
    return 'team'
  }
  return trimmed
}
