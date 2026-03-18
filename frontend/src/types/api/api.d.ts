/**
 * API 接口类型定义模块
 *
 * 提供所有后端接口的类型定义
 *
 * ## 主要功能
 *
 * - 通用类型（分页参数、响应结构等）
 * - 认证类型（登录、用户信息等）
 * - 系统管理类型（用户、角色等）
 * - 全局命名空间声明
 *
 * ## 使用场景
 *
 * - API 请求参数类型约束
 * - API 响应数据类型定义
 * - 接口文档类型同步
 *
 * ## 注意事项
 *
 * - 在 .vue 文件使用需要在 eslint.config.mjs 中配置 globals: { Api: 'readonly' }
 * - 使用全局命名空间，无需导入即可使用
 *
 * ## 使用方式
 *
 * ```typescript
 * const params: Api.Auth.LoginParams = { userName: 'admin', password: '123456' }
 * const response: Api.Auth.UserInfo = await fetchUserInfo()
 * ```
 *
 * @module types/api/api
 * @author Art Design Pro Team
 */

declare namespace Api {
  /** 通用类型 */
  namespace Common {
    /** 分页参数 */
    interface PaginationParams {
      /** 当前页码 */
      current: number
      /** 每页条数 */
      size: number
      /** 总条数 */
      total: number
    }

    /** 通用搜索参数 */
    type CommonSearchParams = Pick<PaginationParams, 'current' | 'size'>

    /** 分页响应基础结构 */
    interface PaginatedResponse<T = any> {
      records: T[]
      current: number
      size: number
      total: number
    }

    /** 启用状态 */
    type EnableStatus = '1' | '2'
  }

  /** 认证类型 */
  namespace Auth {
    /** 登录参数 */
    interface LoginParams {
      username: string
      password: string
    }

    /** 登录响应数据 */
    interface LoginResponseData {
      access_token: string
      refresh_token: string
      expires_in: number
      user: UserInfo
    }

    /** 登录响应 */
    interface LoginResponse {
      access_token: string
      refresh_token: string
      expires_in: number
      user: UserInfo
    }

    /** 用户信息 */
    interface UserInfo {
      id: string
      email: string
      username: string
      nickname: string
      avatar_url?: string
      phone?: string
      status: string
      is_super_admin: boolean
      current_tenant_id?: string
      actions?: string[]
      scoped_actions?: string[]
      created_at: string
      updated_at?: string
      // 兼容字段
      userId?: string | number
      userName?: string
      avatar?: string
      roles?: string[]
      buttons?: string[]
      scopedActions?: string[]
    }
  }

  /** 系统管理类型 */
  namespace SystemManage {
    interface MenuMetaConfig {
      roles?: string[]
      requiresTenantContext?: boolean
      requiredAction?: string
      requiredActions?: string[]
      actionMatchMode?: 'any' | 'all'
      actionVisibilityMode?: 'hide' | 'show'
      [k: string]: unknown
    }

    /** 用户列表 */
    type UserList = Api.Common.PaginatedResponse<UserListItem>

    /** 用户列表项 */
    interface UserListItem {
      id: string
      avatar?: string
      status: string
      userName: string
      nickName: string
      userPhone: string
      userEmail: string
      systemRemark?: string
      lastLoginTime?: string
      lastLoginIP?: string
      userRoles: string[]
      roleDetails?: Array<{ code: string; name: string }>
      registerSource?: string
      invitedBy?: string
      invitedByName?: string
      createBy?: string
      createTime: string
      updateBy?: string
      updateTime: string
    }

    /** 用户搜索参数 */
    type UserSearchParams = Partial<
      Pick<
        UserListItem,
        'userName' | 'userPhone' | 'userEmail' | 'status' | 'registerSource' | 'invitedBy'
      > &
        Api.Common.CommonSearchParams & {
          id?: string // 用户ID
          roleId?: string // 角色ID
        }
    >

    /** 创建用户参数 */
    interface UserCreateParams {
      username: string
      password: string
      email?: string
      nickname?: string
      phone?: string
      systemRemark?: string
      status?: string
      roleIds?: string[]
    }

    /** 更新用户参数 */
    interface UserUpdateParams {
      email?: string
      nickname?: string
      phone?: string
      systemRemark?: string
      status?: string
      roleIds?: string[]
    }

    /** 角色列表 */
    type RoleList = Api.Common.PaginatedResponse<RoleListItem>

    /** 角色列表项 */
    interface RoleListItem {
      roleId: string
      roleName: string
      roleCode: string
      description: string
      sortOrder?: number
      status?: string // normal/suspended
      priority?: number // 优先级
      createTime: string
      tenantId?: string | null
      isGlobal?: boolean
      scopeId?: string
      scopeCode?: string
      scopeName?: string
      scope?: string // 兼容旧字段
      canEditPermission?: boolean
    }

    /** 角色搜索参数 */
    type RoleSearchParams = Partial<
      Pick<RoleListItem, 'roleName' | 'roleCode' | 'description' | 'enabled'> &
        Api.Common.CommonSearchParams & { globalOnly?: boolean }
    >

    /** 创建角色参数（与后端 code/name 对应） */
    interface RoleCreateParams {
      code: string
      name: string
      description?: string
      sort_order?: number
      scope_ids: string[]
      priority?: number
      status?: string
    }

    /** 更新角色参数 */
    interface RoleUpdateParams {
      code?: string
      name?: string
      description?: string
      scope_ids?: string[]
      sort_order?: number
      priority?: number
      status?: string
    }

    interface RoleScopeItem {
      scopeId: string
      scopeCode: string
      scopeName: string
      dataPermissionCode?: string
      dataPermissionName?: string
    }

    interface RoleListItem {
      scopeIds?: string[]
      scopes?: RoleScopeItem[]
    }

      interface PermissionActionItem {
        id: string
        resourceCode: string
        actionCode: string
        moduleCode?: string
        permissionKey?: string
        category?: string
      source?: 'system' | 'api' | 'business' | string
      featureKind?: 'system' | 'business' | string
      name: string
      description?: string
      scopeId?: string
      scopeCode?: string
      scopeName?: string
      dataPermissionCode?: string
      dataPermissionName?: string
      scope?: string
      requiresTenantContext: boolean
      status: string
      sortOrder?: number
      createdAt?: string
      updatedAt?: string
    }

    type PermissionActionList = Api.Common.PaginatedResponse<PermissionActionItem>

    interface APIEndpointItem {
      id: string
      method: string
      path: string
      module: string
      featureKind?: 'system' | 'business' | string
      handler?: string
      summary?: string
      resourceCode?: string
      actionCode?: string
      scopeId?: string
      scopeCode?: string
      scopeName?: string
      dataPermissionCode?: string
      dataPermissionName?: string
      requiresTenantContext: boolean
      status: string
      createdAt?: string
      updatedAt?: string
    }

    type APIEndpointList = Api.Common.PaginatedResponse<APIEndpointItem>

    type APIEndpointSearchParams = Partial<
      Pick<APIEndpointItem, 'method' | 'path' | 'module' | 'status'> &
        Api.Common.CommonSearchParams & {
          featureKind?: string
          resourceCode?: string
          actionCode?: string
          scopeCode?: string
          requiresTenantContext?: boolean
        }
    >

    type PermissionActionSearchParams = Partial<
      Pick<PermissionActionItem, 'name' | 'status' | 'category' | 'source'> &
          Api.Common.CommonSearchParams & {
            keyword?: string
            resourceCode?: string
            actionCode?: string
            moduleCode?: string
            featureKind?: string
            scopeId?: string
            scopeCode?: string
          requiresTenantContext?: boolean
        }
    >

      interface PermissionActionCreateParams {
        resource_code: string
        action_code: string
        module_code?: string
        category?: string
        feature_kind?: 'system' | 'business' | string
      name: string
      description?: string
      scope_id: string
      requires_tenant_context?: boolean
      status?: string
      sort_order?: number
    }

      interface PermissionActionUpdateParams {
        resource_code?: string
        action_code?: string
        module_code?: string
        category?: string
      feature_kind?: 'system' | 'business' | string
      name?: string
      description?: string
      scope_id?: string
      requires_tenant_context?: boolean
      status?: string
      sort_order?: number
    }

    interface RoleActionPermissionItem {
      actionId: string
      effect: 'allow' | 'deny'
    }

    interface RoleDataPermissionItem {
      resourceCode: string
      scopeCode: string
    }

    interface RoleDataPermissionResourceItem {
      resourceCode: string
      resourceName: string
    }

    interface RoleDataPermissionScopeOption {
      scopeCode: string
      scopeName: string
    }

    interface UserActionPermissionItem {
      actionId: string
      effect: 'allow' | 'deny'
      action?: PermissionActionItem
    }

    /** 作用域列表 */
    type ScopeList = Api.Common.PaginatedResponse<ScopeListItem>

    /** 作用域列表项 */
    interface ScopeListItem {
      scopeId: string
      scopeCode: string
      scopeName: string
      description?: string
      isSystem?: boolean
      dataPermissionCode?: string
      dataPermissionName?: string
      sortOrder?: number
      createTime?: string
    }

    /** 作用域搜索参数 */
    type ScopeSearchParams = Partial<
      Pick<ScopeListItem, 'scopeCode' | 'scopeName'> & Api.Common.CommonSearchParams
    >

    /** 创建作用域参数 */
    interface ScopeCreateParams {
      code: string
      name: string
      description?: string
      data_permission_code?: string
      data_permission_name?: string
      sort_order?: number
    }

    /** 更新作用域参数 */
    interface ScopeUpdateParams {
      code?: string
      name?: string
      description?: string
      data_permission_code?: string
      data_permission_name?: string
      sort_order?: number
    }

    /** 团队列表 */
    type TeamList = Api.Common.PaginatedResponse<TeamListItem>

    /** 团队列表项（与后端 tenantToMap 一致） */
    interface TeamListItem {
      id: string
      name: string
      remark: string
      logoUrl?: string
      plan?: string
      maxMembers: number
      maxProducts?: number
      status: string
      createTime: string
      updateTime: string
      ownerId?: string
      adminUserIds?: string[]
      adminUsers?: Array<{
        user_id: string
        user_name?: string
        nick_name?: string
      }>
      currentRoleCode?: string
      memberStatus?: string
    }

    /** 团队搜索参数（与后端 TenantListRequest 一致） */
    interface TeamSearchParams extends Api.Common.CommonSearchParams {
      name?: string
      status?: string
    }

    /** 创建团队参数（与后端 TenantCreateRequest 一致） */
    interface TeamCreateParams {
      name: string
      remark?: string
      logo_url?: string
      plan?: string
      max_members?: number
      status?: string
      admin_user_ids?: string[]
    }

    /** 更新团队参数（与后端 TenantUpdateRequest 一致） */
    interface TeamUpdateParams {
      name?: string
      remark?: string
      logo_url?: string
      plan?: string
      max_members?: number
      status?: string
      admin_user_ids?: string[]
    }

    /** 团队成员项（与后端 ListMembers 返回一致） */
    interface TeamMemberItem {
      id: string
      tenantId?: string
      userId: string
      role: string // 角色名称（如"团队管理员"、"团队成员"）
      roleCode?: string // 角色编码（如"team_admin"、"team_member"）
      status: string
      joinedAt: string | null
      userName: string
      nickName: string
      userEmail: string
      avatar?: string
    }

    interface TeamMemberActionPermissionItem {
      actionId: string
      effect: 'allow' | 'deny'
      action?: PermissionActionItem
    }

    /** 创建菜单参数（与后端 MenuCreateRequest 一致） */
    interface MenuCreateParams {
      parent_id: string | null
      path: string
      name: string
      component?: string
      title: string
      icon?: string
      sort_order?: number
      meta?: MenuMetaConfig
      hidden?: boolean
    }

    /** 更新菜单参数（与后端 MenuUpdateRequest 一致） */
    interface MenuUpdateParams {
      parent_id: string | null
      path?: string
      name?: string
      component?: string
      title?: string
      icon?: string
      sort_order?: number
      meta?: MenuMetaConfig
      hidden?: boolean
    }
  }
}
