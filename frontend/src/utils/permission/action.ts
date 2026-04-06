type ActionRequirement = string

type UserActionInfo = Partial<Api.Auth.UserInfo> | null | undefined

function normalizeActionKey(action: ActionRequirement): string {
  const target = `${action || ''}`.trim()
  if (!target) return ''
  if (target.includes(':')) {
    const [resource, actionCode] = target.split(':', 2)
    return [resource, actionCode].filter(Boolean).join('.')
  }
  return target
}

export function resolveActionKey(action: ActionRequirement): { key: string } {
  return { key: normalizeActionKey(action) }
}

export function buildScopedActionKey(action: string): string {
  return resolveActionKey(action).key
}

export function hasScopedActionPermission(
  userInfo: UserActionInfo,
  action: ActionRequirement
): boolean {
  if (userInfo?.is_super_admin) {
    return true
  }

  const actions = new Set((userInfo?.actions || []).map((item) => normalizeActionKey(item)))
  return actions.has(resolveActionKey(action).key)
}
