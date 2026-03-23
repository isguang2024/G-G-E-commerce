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
      created_at: string
      updated_at?: string
      // 兼容字段
      userId?: string | number
      userName?: string
      avatar?: string
      roles?: string[]
      buttons?: string[]
      [k: string]: unknown
    }
  }

  /** 系统管理类型 */
  namespace SystemManage {
    interface MenuMetaConfig {
      roles?: string[]
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
      priority?: number
      status?: string
    }

    /** 更新角色参数 */
    interface RoleUpdateParams {
      code?: string
      name?: string
      description?: string
      sort_order?: number
      priority?: number
      status?: string
    }

    interface PermissionActionItem {
      id: string
      resourceCode: string
      actionCode: string
      moduleCode?: string
      contextType?: 'platform' | 'team' | string
      permissionKey?: string
      source?: 'system' | 'api' | 'business' | string
      featureKind?: 'system' | 'business' | string
      name: string
      description?: string
      dataPermissionCode?: string
      dataPermissionName?: string
      status: string
      sortOrder?: number
      createdAt?: string
      updatedAt?: string
    }

    type PermissionActionList = Api.Common.PaginatedResponse<PermissionActionItem>

    interface FeaturePackageItem {
      id: string
      packageKey: string
      packageType?: 'base' | 'bundle' | string
      name: string
      description?: string
      contextType?: 'platform' | 'team' | string
      isBuiltin?: boolean
      actionCount?: number
      menuCount?: number
      teamCount?: number
      status: string
      sortOrder?: number
      createdAt?: string
      updatedAt?: string
    }

    interface FeaturePackageTeamBinding {
      team_ids: string[]
    }

    type FeaturePackageList = Api.Common.PaginatedResponse<FeaturePackageItem>

    type FeaturePackageSearchParams = Partial<
      Pick<FeaturePackageItem, 'name' | 'status'> &
        Api.Common.CommonSearchParams & {
          keyword?: string
          packageKey?: string
          packageType?: string
          contextType?: string
        }
    >

    interface FeaturePackageCreateParams {
      package_key: string
      package_type?: 'base' | 'bundle' | string
      name: string
      description?: string
      context_type?: 'platform' | 'team' | string
      status?: string
      sort_order?: number
    }

    interface FeaturePackageUpdateParams {
      package_key?: string
      package_type?: 'base' | 'bundle' | string
      name?: string
      description?: string
      context_type?: 'platform' | 'team' | string
      status?: string
      sort_order?: number
    }

    interface FeaturePackageActionResponse {
      action_ids: string[]
      actions: PermissionActionItem[]
    }

    interface FeaturePackageBundleResponse {
      child_package_ids: string[]
      packages: FeaturePackageItem[]
    }

    interface FeaturePackageMenuResponse {
      menu_ids: string[]
      menus: AppRouteRecord[]
    }

    interface FeaturePackageTeamSetParams {
      team_ids: string[]
    }

    interface TeamFeaturePackageResponse {
      package_ids: string[]
      packages: FeaturePackageItem[]
    }

    interface RoleFeaturePackageResponse {
      package_ids: string[]
      packages: FeaturePackageItem[]
      inherited?: boolean
    }

    interface UserFeaturePackageResponse {
      package_ids: string[]
      packages: FeaturePackageItem[]
    }

    interface RoleActionBoundaryResponse {
      action_ids: string[]
      available_action_ids: string[]
      actions?: PermissionActionItem[]
      expanded_package_ids?: string[]
      disabled_action_ids?: string[]
      derived_sources?: Array<{
        action_id: string
        package_ids: string[]
      }>
    }

    interface RoleMenuBoundaryResponse {
      menu_ids: string[]
      available_menu_ids: string[]
      expanded_package_ids?: string[]
      hidden_menu_ids?: string[]
      derived_sources?: Array<{
        menu_id: string
        package_ids: string[]
      }>
    }

    interface TeamActionOriginsResponse {
      derived_action_ids: string[]
      derived_sources?: Array<{
        action_id: string
        package_ids: string[]
      }>
      blocked_action_ids?: string[]
    }

    interface TeamMenuOriginsResponse {
      derived_menu_ids: string[]
      derived_sources?: Array<{
        menu_id: string
        package_ids: string[]
      }>
      blocked_menu_ids: string[]
    }

    interface APIEndpointItem {
      id: string
      method: string
      path: string
      spec?: string
      module: string
      featureKind?: 'system' | 'business' | string
      handler?: string
      summary?: string
      permissionKey?: string
      authMode?: 'public' | 'jwt' | 'permission' | 'api_key' | string
      resourceCode?: string
      actionCode?: string
      dataPermissionCode?: string
      dataPermissionName?: string
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
        }
    >

    type PermissionActionSearchParams = Partial<
      Pick<PermissionActionItem, 'name' | 'status' | 'source'> &
          Api.Common.CommonSearchParams & {
            keyword?: string
            permissionKey?: string
            resourceCode?: string
            actionCode?: string
            moduleCode?: string
            contextType?: string
            featureKind?: string
        }
    >
    interface PermissionActionCreateParams {
      permission_key: string
      resource_code?: string
      action_code?: string
      module_code?: string
      context_type?: 'platform' | 'team' | string
      feature_kind?: 'system' | 'business' | string
      name: string
      description?: string
      status?: string
      sort_order?: number
    }

    interface PermissionActionUpdateParams {
      permission_key?: string
      resource_code?: string
      action_code?: string
      module_code?: string
      context_type?: 'platform' | 'team' | string
      feature_kind?: 'system' | 'business' | string
      name?: string
      description?: string
      status?: string
      sort_order?: number
    }

    type RoleActionPermissionItem = string

    interface RoleDataPermissionItem {
      resourceCode: string
      dataScope: string
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
      action_id: string
      effect: 'allow' | 'deny'
      action?: PermissionActionItem
    }

    interface UserActionPermissionResponse {
      actions: UserActionPermissionItem[] // 历史例外审计字段，主链候选集优先使用 available_action_ids / available_actions
      available_action_ids?: string[]
      available_actions?: PermissionActionItem[]
      expanded_package_ids?: string[]
      derived_sources?: Array<{
        action_id: string
        package_ids: string[]
      }>
      has_package_config?: boolean
    }

    interface UserMenuBoundaryResponse {
      menu_ids: string[]
      available_menu_ids?: string[]
      hidden_menu_ids?: string[]
      expanded_package_ids?: string[]
      derived_sources?: Array<{
        menu_id: string
        package_ids: string[]
      }>
      has_package_config?: boolean
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
      action_id: string
      effect: 'allow' | 'deny'
      action?: PermissionActionItem
    }

    interface TeamMemberActionPermissionResponse {
      actions: TeamMemberActionPermissionItem[] // 历史成员例外审计字段，主链候选集优先使用 available_action_ids / available_actions
      available_action_ids?: string[]
      available_actions?: PermissionActionItem[]
      derived_sources?: Array<{
        action_id: string
        package_ids: string[]
      }>
    }

    interface FeaturePackageActionSetParams {
      action_ids: string[]
    }

    interface FeaturePackageChildSetParams {
      child_package_ids: string[]
    }

    interface TeamFeaturePackageSetParams {
      package_ids: string[]
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
