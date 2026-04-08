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

// 以 userInfo.actions 引用为 key 缓存 normalized Set，避免每次校验都重建。
const actionSetCache = new WeakMap<object, Set<string>>()

function getNormalizedActionSet(userInfo: UserActionInfo): Set<string> {
  const list = userInfo?.actions
  if (!list || !Array.isArray(list)) return new Set()
  const cached = actionSetCache.get(list as unknown as object)
  if (cached) return cached
  const set = new Set(list.map((item) => normalizeActionKey(item)))
  actionSetCache.set(list as unknown as object, set)
  return set
}

export function hasScopedActionPermission(
  userInfo: UserActionInfo,
  action: ActionRequirement
): boolean {
  if (userInfo?.is_super_admin) {
    return true
  }

  return getNormalizedActionSet(userInfo).has(resolveActionKey(action).key)
}
