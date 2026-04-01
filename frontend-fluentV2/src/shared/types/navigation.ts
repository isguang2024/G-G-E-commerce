export type RouteStatus = 'implemented' | 'placeholder'
export type NavigationGroupKey = 'welcome' | 'workspace' | 'team' | 'message' | 'system'
export type SpaceKey = 'default' | 'platform-governance' | 'team-collaboration'

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

export interface BreadcrumbSegment {
  label: string
  path?: string
}

export interface NavigationSpace {
  key: SpaceKey
  label: string
  description: string
  kind: 'default' | 'platform' | 'team'
  defaultLandingRoute: string
}

export interface NavigationItem {
  id: string
  routeId: string
  path: string
  label: string
  icon: NavIconKey
  group: NavigationGroupKey
  status: RouteStatus
  spaceVisibility: SpaceKey[] | 'all'
  children?: NavigationItem[]
}

export interface PageMeta {
  routeId: string
  title: string
  subtitle: string
  groupLabel: string
  spaceKey?: SpaceKey
  breadcrumbs: BreadcrumbSegment[]
}

export interface RouteDefinition {
  id: string
  path: string
  group: NavigationGroupKey
  status: RouteStatus
  shellTitle: string
}

export interface ShellTab {
  routeId: string
  path: string
  label: string
  group: NavigationGroupKey
  groupLabel: string
  pinned?: boolean
}
