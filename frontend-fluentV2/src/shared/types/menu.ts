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
  kind: string
  path: string
  name: string
  title: string
  component: string
  icon: string
  sortOrder: number
  hidden: boolean
  meta: Record<string, unknown>
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
