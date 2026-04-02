import { requestData } from '@/shared/api/client'
import type {
  MenuSpace,
  NavigationGroupKey,
  RuntimeCurrentSpace,
  RuntimeCurrentSpaceBinding,
  RuntimeManagedPage,
  RuntimeNavItem,
  RuntimeNavigationManifest,
} from '@/shared/types/navigation'

function normalizeSpaceKind(value: string): MenuSpace['kind'] {
  switch (`${value || ''}`.trim()) {
    case 'platform':
      return 'platform'
    case 'team':
      return 'team'
    case 'default':
      return 'default'
    default:
      return 'custom'
  }
}

function inferGroup(path: string): NavigationGroupKey {
  if (path.startsWith('/workspace')) return 'workspace'
  if (path.startsWith('/team')) return 'team'
  if (path.startsWith('/message')) return 'message'
  if (path.startsWith('/system')) return 'system'
  return 'welcome'
}

function inferIcon(rawIcon: string) {
  const icon = rawIcon.trim().toLowerCase()
  if (!icon) {
    return undefined
  }

  if (icon.includes('menu')) return 'menu'
  if (icon.includes('page')) return 'page'
  if (icon.includes('role')) return 'role'
  if (icon.includes('user')) return 'user'
  if (icon.includes('api')) return 'api'
  if (icon.includes('package')) return 'package'
  if (icon.includes('message')) return 'message'
  if (icon.includes('team')) return 'team'
  if (icon.includes('space')) return 'space'
  if (icon.includes('inbox')) return 'inbox'
  if (icon.includes('workspace')) return 'workspace'
  if (icon.includes('home') || icon.includes('dashboard') || icon.includes('console')) return 'home'
  return undefined
}

function normalizeMenuSpace(input: Record<string, unknown>): MenuSpace {
  const meta = (input.meta || {}) as Record<string, unknown>
  return {
    id: `${input.id || ''}`.trim(),
    key: `${input.space_key || input.spaceKey || ''}`.trim(),
    label: `${input.name || ''}`.trim(),
    description: `${input.description || ''}`.trim(),
    defaultLandingRoute: `${input.default_home_path || input.defaultHomePath || '/welcome'}`.trim() || '/welcome',
    status: `${input.status || 'normal'}`.trim(),
    kind: normalizeSpaceKind(`${meta.spaceType || meta.space_type || 'custom'}`),
    isDefault: Boolean(input.is_default ?? input.isDefault),
    accessMode: `${input.access_mode || input.accessMode || meta.access_mode || meta.accessMode || 'all'}`.trim() || 'all',
    allowedRoleCodes: Array.isArray(input.allowed_role_codes ?? input.allowedRoleCodes)
      ? ((input.allowed_role_codes ?? input.allowedRoleCodes) as unknown[])
          .map((item: unknown) => `${item || ''}`.trim())
          .filter(Boolean)
      : [],
    hosts: Array.isArray(input.hosts)
      ? input.hosts.map((item) => `${item || ''}`.trim()).filter(Boolean)
      : [],
  }
}

function normalizeBinding(input?: Record<string, unknown>): RuntimeCurrentSpaceBinding | undefined {
  if (!input) {
    return undefined
  }

  return {
    host: `${input.host || ''}`.trim(),
    spaceKey: `${input.space_key || input.spaceKey || ''}`.trim(),
    spaceName: `${input.space_name || input.spaceName || ''}`.trim(),
    routePrefix: `${input.route_prefix || input.routePrefix || ''}`.trim(),
    authMode: `${input.auth_mode || input.authMode || 'inherit_host'}`.trim(),
    loginHost: `${input.login_host || input.loginHost || ''}`.trim(),
    callbackHost: `${input.callback_host || input.callbackHost || ''}`.trim(),
    cookieScopeMode: `${input.cookie_scope_mode || input.cookieScopeMode || 'inherit'}`.trim(),
    cookieDomain: `${input.cookie_domain || input.cookieDomain || ''}`.trim(),
  }
}

function normalizeCurrentSpace(input?: Record<string, unknown>): RuntimeCurrentSpace | undefined {
  if (!input?.space || typeof input.space !== 'object') {
    return undefined
  }

  return {
    space: normalizeMenuSpace(input.space as Record<string, unknown>),
    binding: normalizeBinding(input.binding as Record<string, unknown> | undefined),
    resolvedBy: `${input.resolved_by || input.resolvedBy || ''}`.trim(),
    requestHost: `${input.request_host || input.requestHost || ''}`.trim(),
    accessGranted: Boolean(input.access_granted ?? input.accessGranted ?? true),
  }
}

function normalizeRuntimeNavItem(input: Record<string, unknown>): RuntimeNavItem {
  const path = `${input.path || ''}`.trim()
  const meta = (input.meta || {}) as Record<string, unknown>
  const group = inferGroup(path)
  const rawIcon = `${meta.icon || ''}`.trim()

  return {
    id: `${input.id || path}`.trim(),
    routeId: `${input.name || path}`.trim(),
    path,
    label: `${meta.title || input.title || input.name || path}`.trim(),
    title: `${meta.title || input.title || input.name || path}`.trim(),
    icon: inferIcon(rawIcon),
    group,
    status: 'placeholder',
    spaceKey: `${input.space_key || input.spaceKey || meta.spaceKey || 'default'}`.trim(),
    hidden: Boolean(meta.isHide ?? input.hidden ?? false),
    component: `${input.component || ''}`.trim(),
    permissionKey: `${meta.permissionKey || input.permission_key || input.permissionKey || ''}`.trim() || undefined,
    accessMode: `${meta.accessMode || input.access_mode || input.accessMode || ''}`.trim() || undefined,
    manageGroupName: `${input.manage_group_name || input.manageGroupName || ''}`.trim() || undefined,
    meta,
    children: Array.isArray(input.children)
      ? input.children.map((item) => normalizeRuntimeNavItem(item as Record<string, unknown>))
      : undefined,
  }
}

function normalizeRuntimeManagedPage(input: Record<string, unknown>): RuntimeManagedPage {
  return {
    pageKey: `${input.page_key || input.pageKey || ''}`.trim(),
    name: `${input.name || ''}`.trim(),
    routePath: `${input.route_path || input.routePath || ''}`.trim(),
    routeName: `${input.route_name || input.routeName || ''}`.trim() || undefined,
    component: `${input.component || ''}`.trim(),
    pageType: `${input.page_type || input.pageType || 'inner'}`.trim(),
    parentMenuId: `${input.parent_menu_id || input.parentMenuId || ''}`.trim() || undefined,
    parentPageKey: `${input.parent_page_key || input.parentPageKey || ''}`.trim() || undefined,
    activeMenuPath: `${input.active_menu_path || input.activeMenuPath || ''}`.trim() || undefined,
    breadcrumbMode: `${input.breadcrumb_mode || input.breadcrumbMode || ''}`.trim() || undefined,
    accessMode: `${input.access_mode || input.accessMode || ''}`.trim() || undefined,
    permissionKey: `${input.permission_key || input.permissionKey || ''}`.trim() || undefined,
    keepAlive: Boolean(input.keep_alive ?? input.keepAlive ?? false),
    isFullPage: Boolean(input.is_full_page ?? input.isFullPage ?? false),
    spaceKey: `${input.space_key || input.spaceKey || ''}`.trim() || undefined,
    spaceKeys: Array.isArray(input.space_keys || input.spaceKeys)
      ? ((input.space_keys || input.spaceKeys) as unknown[])
          .map((item: unknown) => `${item || ''}`.trim())
          .filter(Boolean)
      : [],
    spaceScope: `${input.space_scope || input.spaceScope || ''}`.trim() || undefined,
    status: `${input.status || ''}`.trim() || undefined,
    meta: ((input.meta || {}) as Record<string, unknown>) || {},
  }
}

export async function fetchMenuSpaces() {
  const result = await requestData<{
    records?: Array<Record<string, unknown>>
  }>({
    method: 'GET',
    url: '/api/v1/system/menu-spaces',
  })

  return (result.records || []).map(normalizeMenuSpace)
}

export async function fetchRuntimeNavigationManifest(spaceKey?: string): Promise<RuntimeNavigationManifest> {
  const result = await requestData<Record<string, unknown>>({
    method: 'GET',
    url: '/api/v1/runtime/navigation',
    params: spaceKey ? { space_key: spaceKey } : undefined,
  })

  const context = (result.context || {}) as Record<string, unknown>
  return {
    currentSpace: normalizeCurrentSpace(result.current_space as Record<string, unknown> | undefined),
    context: {
      spaceKey: `${context.space_key || context.spaceKey || ''}`.trim(),
      requestHost: `${context.request_host || context.requestHost || ''}`.trim(),
      requestedSpaceKey: `${context.requested_space_key || context.requestedSpaceKey || ''}`.trim(),
      authenticated: Boolean(context.authenticated),
      superAdmin: Boolean(context.super_admin ?? context.superAdmin),
      visibleMenuCount: Number(context.visible_menu_count ?? context.visibleMenuCount ?? 0),
      managedPageCount: Number(context.managed_page_count ?? context.managedPageCount ?? 0),
      actionKeyCount: Number(context.action_key_count ?? context.actionKeyCount ?? 0),
      userId: `${context.user_id || context.userId || ''}`.trim() || undefined,
      tenantId: `${context.tenant_id || context.tenantId || ''}`.trim() || undefined,
    },
    menuTree: Array.isArray(result.menu_tree)
      ? result.menu_tree.map((item) => normalizeRuntimeNavItem(item as Record<string, unknown>))
      : [],
    entryRoutes: Array.isArray(result.entry_routes)
      ? result.entry_routes.map((item) => normalizeRuntimeNavItem(item as Record<string, unknown>))
      : [],
    managedPages: Array.isArray(result.managed_pages)
      ? result.managed_pages.map((item) => normalizeRuntimeManagedPage(item as Record<string, unknown>))
      : [],
    versionStamp: `${result.version_stamp || result.versionStamp || ''}`.trim(),
  }
}
