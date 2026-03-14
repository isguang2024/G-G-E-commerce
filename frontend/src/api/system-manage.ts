import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

const USER_BASE = '/api/v1/users'
const ROLE_BASE = '/api/v1/roles'
const SCOPE_BASE = '/api/v1/scopes'
const TENANT_BASE = '/api/v1/tenants'
const SYSTEM_BASE = '/api/v1/system'

// 获取用户列表
export function fetchGetUserList(params: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: USER_BASE,
    params
  })
}

/** 获取用户权限（计算后的菜单权限） */
export function fetchGetUserPermissions(userId: string) {
  return request.get<any[]>({
    url: `${USER_BASE}/${userId}/permissions`
  })
}

// 获取用户详情
export function fetchGetUser(id: string) {
  return request.get<Api.SystemManage.UserListItem>({
    url: `${USER_BASE}/${id}`
  })
}

// 创建用户
export function fetchCreateUser(data: Api.SystemManage.UserCreateParams) {
  return request.post<{ id: string }>({
    url: USER_BASE,
    data
  })
}

// 更新用户
export function fetchUpdateUser(id: string, data: Api.SystemManage.UserUpdateParams) {
  return request.put<void>({
    url: `${USER_BASE}/${id}`,
    data
  })
}

// 删除用户
export function fetchDeleteUser(id: string) {
  return request.del<void>({
    url: `${USER_BASE}/${id}`
  })
}

// 分配用户角色
export function fetchAssignUserRoles(id: string, roleIds: string[]) {
  return request.put<void>({
    url: `${USER_BASE}/${id}/roles`,
    data: { roleIds }
  })
}

// 获取角色列表
export function fetchGetRoleList(params: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: ROLE_BASE,
    params
  })
}

// 获取角色列表（简单列表，用于下拉等）
export function fetchGetRoleListSimple() {
  return request.get<Api.SystemManage.RoleList>({
    url: ROLE_BASE,
    params: { current: 1, size: 100 }
  })
}

// 获取角色详情
export function fetchGetRole(id: string) {
  return request.get<Api.SystemManage.RoleListItem>({
    url: `${ROLE_BASE}/${id}`
  })
}

// 创建角色
export function fetchCreateRole(data: Api.SystemManage.RoleCreateParams) {
  return request.post<{ roleId: string }>({
    url: ROLE_BASE,
    data
  })
}

// 更新角色
export function fetchUpdateRole(id: string, data: Api.SystemManage.RoleUpdateParams) {
  return request.put<void>({
    url: `${ROLE_BASE}/${id}`,
    data
  })
}

// 删除角色
export function fetchDeleteRole(id: string) {
  return request.del<void>({
    url: `${ROLE_BASE}/${id}`
  })
}

/** 获取角色已分配的菜单 ID 列表（用于菜单权限配置） */
export function fetchGetRoleMenus(roleId: string) {
  return request.get<{ menu_ids: string[] }>({
    url: `${ROLE_BASE}/${roleId}/menus`
  })
}

/** 设置角色菜单权限 */
export function fetchSetRoleMenus(roleId: string, menuIds: string[]) {
  return request.put<void>({
    url: `${ROLE_BASE}/${roleId}/menus`,
    data: { menu_ids: menuIds }
  })
}

// ========== 作用域管理 ==========
/** 获取作用域列表 */
export function fetchGetScopeList(params: Api.SystemManage.ScopeSearchParams) {
  return request.get<Api.SystemManage.ScopeList>({
    url: SCOPE_BASE,
    params
  })
}

/** 获取所有作用域（用于下拉选择） */
export function fetchGetAllScopes() {
  return request.get<Api.SystemManage.ScopeListItem[]>({
    url: `${SCOPE_BASE}/all`
  })
}

/** 获取作用域详情 */
export function fetchGetScope(id: string) {
  return request.get<Api.SystemManage.ScopeListItem>({
    url: `${SCOPE_BASE}/${id}`
  })
}

/** 创建作用域 */
export function fetchCreateScope(data: Api.SystemManage.ScopeCreateParams) {
  return request.post<{ scopeId: string }>({
    url: SCOPE_BASE,
    data
  })
}

/** 更新作用域 */
export function fetchUpdateScope(id: string, data: Api.SystemManage.ScopeUpdateParams) {
  return request.put<void>({
    url: `${SCOPE_BASE}/${id}`,
    data
  })
}

/** 删除作用域 */
export function fetchDeleteScope(id: string) {
  return request.del<void>({
    url: `${SCOPE_BASE}/${id}`,
    showErrorMessage: false // 禁用自动错误消息显示，由业务代码处理
  })
}

const MENU_BASE = '/api/v1/menus'

/** 获取菜单树（按当前用户角色过滤，用于侧栏；后端菜单模式时使用） */
export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: `${MENU_BASE}/tree`
  })
}

/** 获取完整菜单树（不限角色，用于菜单管理页；需管理员） */
export function fetchGetMenuTreeAll() {
  return request.get<AppRouteRecord[]>({
    url: `${MENU_BASE}/tree`,
    params: { all: 1 }
  })
}

/** 枚举 views 页面文件（后端 Redis 缓存，支持强制刷新） */
export function fetchGetViewPages(force = false) {
  return request.get<{
    pages: Array<{ filePath: string; componentPath: string }>
    refreshed: boolean
    refreshedAt: string
  }>({
    url: `${SYSTEM_BASE}/view-pages`,
    params: force ? { force: 1 } : undefined
  })
}

/** 创建菜单 */
export function fetchCreateMenu(data: Api.SystemManage.MenuCreateParams) {
  return request.post<{ id: string }>({
    url: MENU_BASE,
    data
  })
}

/** 更新菜单 */
export function fetchUpdateMenu(id: string, data: Api.SystemManage.MenuUpdateParams) {
  return request.put<void>({
    url: `${MENU_BASE}/${id}`,
    data
  })
}

/** 删除菜单 */
export function fetchDeleteMenu(id: string) {
  return request.del<void>({
    url: `${MENU_BASE}/${id}`
  })
}

/** 更新菜单排序（全量重排） */
export function fetchUpdateMenuSort(data: { id: string; sort_order: number }[]) {
  return request.put<void>({
    url: `${MENU_BASE}/sort`,
    data
  })
}

/** 根据父级ID更新子节点排序（全量重排） */
export function fetchUpdateMenuSortByParent(parentId: string | null, menuIds: string[]) {
  return request.put<void>({
    url: `${MENU_BASE}/sort-by-parent`,
    data: {
      parent_id: parentId,
      menu_ids: menuIds
    }
  })
}

// 菜单备份相关API
const MENU_BACKUP_BASE = '/api/v1/menus/backups'

/** 创建菜单备份 */
export function fetchCreateMenuBackup(data: { name: string; description?: string }) {
  return request.post<void>({
    url: MENU_BACKUP_BASE,
    data
  })
}

/** 获取菜单备份列表 */
export function fetchGetMenuBackupList() {
  return request.get<
    {
      id: string
      name: string
      description: string
      created_at: string
      created_by: string
    }[]
  >({
    url: MENU_BACKUP_BASE
  })
}

/** 删除菜单备份 */
export function fetchDeleteMenuBackup(id: string) {
  return request.del<void>({
    url: `${MENU_BACKUP_BASE}/${id}`
  })
}

/** 恢复菜单备份 */
export function fetchRestoreMenuBackup(id: string) {
  return request.post<void>({
    url: `${MENU_BACKUP_BASE}/${id}/restore`
  })
}
