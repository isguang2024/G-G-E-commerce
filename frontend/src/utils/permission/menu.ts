import type { RouteMeta } from '@/types/router'
import { hasScopedActionPermission } from './action'

type MenuPermissionMeta = Partial<RouteMeta> | Record<string, unknown> | null | undefined
type ActionMatchMode = 'any' | 'all'
type ActionVisibilityMode = 'hide' | 'show'

interface MenuActionRequirement {
  actions: string[]
  matchMode: ActionMatchMode
  visibilityMode: ActionVisibilityMode
}

function normalizeAccessMode(meta: MenuPermissionMeta): 'public' | 'jwt' | 'permission' {
  const value = `${meta?.accessMode || ''}`.trim().toLowerCase()
  if (value === 'public' || value === 'jwt' || value === 'permission') {
    return value
  }
  return 'permission'
}

function normalizeActionList(meta: MenuPermissionMeta): string[] {
  const actions: string[] = []
  const requiredAction = `${meta?.requiredAction || ''}`.trim()
  const requiredActions = Array.isArray(meta?.requiredActions) ? meta?.requiredActions : []

  if (requiredAction) {
    actions.push(requiredAction)
  }

  requiredActions.forEach((item) => {
    const value = `${item || ''}`.trim()
    if (value) {
      actions.push(value)
    }
  })

  return Array.from(new Set(actions))
}

export function getMenuActionRequirement(meta: MenuPermissionMeta): MenuActionRequirement {
  return {
    actions: normalizeActionList(meta),
    matchMode: meta?.actionMatchMode === 'all' ? 'all' : 'any',
    visibilityMode: meta?.actionVisibilityMode === 'show' ? 'show' : 'hide'
  }
}

export function hasMenuActionAccess(
  userInfo: Partial<Api.Auth.UserInfo> | null | undefined,
  meta: MenuPermissionMeta
): boolean {
  const accessMode = normalizeAccessMode(meta)
  if (accessMode === 'public' || accessMode === 'jwt') {
    return true
  }
  const requirement = getMenuActionRequirement(meta)
  if (!requirement.actions.length) {
    return true
  }
  if (userInfo?.is_super_admin) {
    return true
  }
  if (requirement.matchMode === 'all') {
    return requirement.actions.every((action) => hasScopedActionPermission(userInfo, action))
  }
  return requirement.actions.some((action) => hasScopedActionPermission(userInfo, action))
}

export function shouldHideMenuWhenActionDenied(meta: MenuPermissionMeta): boolean {
  const accessMode = normalizeAccessMode(meta)
  if (accessMode === 'public' || accessMode === 'jwt') {
    return false
  }
  return getMenuActionRequirement(meta).visibilityMode !== 'show'
}
