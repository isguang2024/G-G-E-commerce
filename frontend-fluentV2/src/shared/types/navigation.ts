export type RouteStatus = 'implemented' | 'placeholder'
export type NavigationGroupKey = 'welcome' | 'workspace' | 'team' | 'message' | 'system'
export type NavIconKey =
  | 'home'
  | 'workspace'
  | 'team'
  | 'message'
  | 'system'
  | 'menu'
  | 'page'
  | 'role'
  | 'user'
  | 'api'
  | 'package'
  | 'inbox'
  | 'space'
  | string

export interface BreadcrumbSegment {
  label: string
  path?: string
}

export interface MenuSpace {
  id: string
  key: string
  label: string
  description: string
  defaultLandingRoute: string
  status: string
  kind: 'default' | 'platform' | 'team' | 'custom'
  isDefault: boolean
  accessMode: string
  allowedRoleCodes: string[]
  hosts: string[]
}

export type NavigationSpace = MenuSpace
export type SpaceKey = string

export interface RuntimeNavItem {
  id: string
  routeId: string
  path: string
  label: string
  title: string
  icon: NavIconKey
  group: NavigationGroupKey
  status: RouteStatus
  spaceKey: string
  hidden: boolean
  component: string
  permissionKey?: string
  accessMode?: string
  manageGroupName?: string
  meta: Record<string, unknown>
  children?: RuntimeNavItem[]
}

export interface RuntimeManagedPage {
  pageKey: string
  name: string
  routePath: string
  routeName?: string
  component: string
  pageType: string
  parentMenuId?: string
  parentPageKey?: string
  activeMenuPath?: string
  breadcrumbMode?: string
  accessMode?: string
  permissionKey?: string
  keepAlive?: boolean
  isFullPage?: boolean
  spaceKey?: string
  spaceKeys: string[]
  spaceScope?: string
  status?: string
  meta: Record<string, unknown>
}

export interface RuntimeCurrentSpaceBinding {
  host: string
  spaceKey: string
  spaceName: string
  routePrefix: string
  authMode: string
  loginHost: string
  callbackHost: string
  cookieScopeMode: string
  cookieDomain: string
}

export interface RuntimeCurrentSpace {
  space: MenuSpace
  binding?: RuntimeCurrentSpaceBinding
  resolvedBy: string
  requestHost: string
  accessGranted: boolean
}

export interface RuntimeNavigationManifest {
  currentSpace?: RuntimeCurrentSpace
  context: {
    spaceKey: string
    requestHost: string
    requestedSpaceKey: string
    authenticated: boolean
    superAdmin: boolean
    visibleMenuCount: number
    managedPageCount: number
    actionKeyCount: number
    userId?: string
    tenantId?: string
  }
  menuTree: RuntimeNavItem[]
  entryRoutes: RuntimeNavItem[]
  managedPages: RuntimeManagedPage[]
  versionStamp: string
}

export interface NavigationItem {
  id: string
  routeId: string
  path: string
  label: string
  icon: NavIconKey
  group: NavigationGroupKey
  status: RouteStatus
  spaceKey: string
  children?: NavigationItem[]
}

export interface RouteDefinition {
  id: string
  path: string
  group: NavigationGroupKey
  status: RouteStatus
  shellTitle: string
  subtitle: string
}

export interface ShellTab {
  routeId: string
  path: string
  label: string
  group: NavigationGroupKey
  groupLabel: string
  pinned?: boolean
}

export interface RouteContext {
  routeId: string
  path: string
  title: string
  subtitle: string
  group: NavigationGroupKey
  groupLabel: string
  status: RouteStatus
  breadcrumbs: BreadcrumbSegment[]
  spaceKey?: string
  source: 'local' | 'runtime-menu' | 'runtime-page'
  pageKey?: string
  permissionKey?: string
  accessMode?: string
  manageGroupName?: string
}
