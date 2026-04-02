export type MenuNodeKind = 'directory' | 'entry' | 'external'

export interface MenuNodeMeta extends Record<string, unknown> {
  title?: string
  icon?: string
  link?: string
  activePath?: string
  accessMode?: string
  isHide?: boolean
  isIframe?: boolean
  isHideTab?: boolean
  keepAlive?: boolean
  fixedTab?: boolean
  isFullPage?: boolean
}

export interface MenuManageGroup {
  id: string
  name: string
  sortOrder: number
  status: string
}

export interface MenuNode {
  id: string
  parentId?: string
  manageGroupId?: string
  manageGroup?: MenuManageGroup
  spaceKey: string
  kind: MenuNodeKind
  path: string
  name: string
  title: string
  component: string
  icon: string
  sortOrder: number
  hidden: boolean
  meta: MenuNodeMeta
  children: MenuNode[]
}

export interface MenuPageBinding {
  pageKey: string
  name: string
  routePath: string
  routeName?: string
  component: string
  parentMenuId?: string
  accessMode?: string
  permissionKey?: string
  parentPageKey?: string
  pageType: string
  spaceKey?: string
  spaceKeys: string[]
}

export interface MenuNodeDetail extends MenuNode {
  parent?: Pick<MenuNode, 'id' | 'title' | 'path'>
  childCount: number
  linkedPages: MenuPageBinding[]
  permissionKeys: string[]
}

export interface MenuDeletePreview {
  mode: 'single' | 'cascade' | 'promote_children'
  menuCount: number
  childCount: number
  affectedPageCount: number
  affectedRelationCount: number
}

export interface MenuMutationDraft {
  parentId: string | null
  manageGroupId: string | null
  spaceKey: string
  kind: MenuNodeKind
  path: string
  name: string
  title: string
  component: string
  icon: string
  sortOrder: number
  hidden: boolean
  accessMode: string
  activePath: string
  externalLink: string
  keepAlive: boolean
  fixedTab: boolean
  isFullPage: boolean
}
