type ActionRequirement = string

type UserActionInfo = Partial<Api.Auth.UserInfo> | null | undefined

export function resolveActionKey(action: ActionRequirement): { key: string } {
  return { key: `${action || ''}`.trim() }
}

export function buildScopedActionKey(action: string): string {
  return resolveActionKey(action).key
}

export function hasScopedActionPermission(userInfo: UserActionInfo, action: ActionRequirement): boolean {
  if (userInfo?.is_super_admin) {
    return true
  }

  const actions = new Set(userInfo?.actions || [])
  return actions.has(resolveActionKey(action).key)
}
