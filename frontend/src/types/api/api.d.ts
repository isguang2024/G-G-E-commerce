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
      // 后端已编译导航显隐，这些字段主要保留给页面内按钮提示或兼容链路使用。
      requiredAction?: string
      requiredActions?: string[]
      actionMatchMode?: 'any' | 'all'
      actionVisibilityMode?: 'hide' | 'show'
      // spaceKey / spaceType / hostKey 都属于后端显式下发的运行时上下文，前端不自行推导。
      spaceKey?: string
      spaceType?: string
      hostKey?: string
      [k: string]: unknown
    }

    interface MenuManageGroupItem {
      id: string
      name: string
      sortOrder: number
      status: 'normal' | 'disabled' | string
      createdAt?: string
      updatedAt?: string
    }

    interface MenuManageGroupSaveParams {
      name: string
      sort_order?: number
      status?: 'normal' | 'disabled' | string
    }

    interface MenuSpaceItem {
      id?: string
      spaceKey: string
      name: string
      description?: string
      defaultHomePath?: string
      isDefault?: boolean
      status?: 'normal' | 'disabled' | string
      hostCount?: number
      hosts?: string[]
      menuCount?: number
      pageCount?: number
      accessMode?: 'all' | 'platform_admin' | 'team_admin' | 'role_codes' | string
      allowedRoleCodes?: string[]
      meta?: Record<string, any>
      createdAt?: string
      updatedAt?: string
    }

    interface MenuSpaceHostBindingItem {
      id?: string
      host: string
      spaceKey: string
      spaceName?: string
      description?: string
      isDefault?: boolean
      status?: 'normal' | 'disabled' | string
      scheme?: 'http' | 'https' | string
      routePrefix?: string
      authMode?: 'inherit_host' | 'centralized_login' | 'shared_cookie' | string
      loginHost?: string
      callbackHost?: string
      cookieScopeMode?: 'inherit' | 'host_only' | 'parent_domain' | string
      cookieDomain?: string
      meta?: Record<string, any>
      createdAt?: string
      updatedAt?: string
    }

    interface CurrentMenuSpaceResponse {
      space: MenuSpaceItem
      binding?: MenuSpaceHostBindingItem
      resolvedBy?: string
      requestHost?: string
      accessGranted?: boolean
    }

    interface MenuSpaceModeResponse {
      mode: 'single' | 'multi' | string
    }

    interface MenuSpaceInitializeResult {
      sourceSpaceKey: string
      targetSpaceKey: string
      forceReinitialized?: boolean
      clearedMenuCount?: number
      clearedPageCount?: number
      clearedPackageMenuLinkCount?: number
      createdMenuCount: number
      createdPageCount: number
      createdPackageMenuLinkCount: number
    }

    interface MenuSpaceSaveParams {
      space_key: string
      name: string
      description?: string
      default_home_path?: string
      is_default?: boolean
      status?: 'normal' | 'disabled' | string
      access_mode?: 'all' | 'platform_admin' | 'team_admin' | 'role_codes' | string
      allowed_role_codes?: string[]
      meta?: Record<string, any>
    }

    interface MenuSpaceHostBindingSaveParams {
      host: string
      space_key: string
      description?: string
      is_default?: boolean
      status?: 'normal' | 'disabled' | string
      meta?: {
        scheme?: 'http' | 'https' | string
        route_prefix?: string
        auth_mode?: 'inherit_host' | 'centralized_login' | 'shared_cookie' | string
        login_host?: string
        callback_host?: string
        cookie_scope_mode?: 'inherit' | 'host_only' | 'parent_domain' | string
        cookie_domain?: string
        [key: string]: any
      }
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

    interface UserPermissionContext {
      type: 'platform' | 'team' | string
      tenantId?: string
      tenantName?: string
    }

    interface UserPermissionSnapshotSummary {
      refreshedAt?: string
      updatedAt?: string
      roleCount?: number
      directPackageCount?: number
      expandedPackageCount?: number
      actionCount?: number
      disabledActionCount?: number
      menuCount?: number
      hasPackageConfig?: boolean
      derivedActionCount?: number
      blockedActionCount?: number
      effectiveActionCount?: number
    }

    interface UserPermissionRoleResult {
      roleId: string
      roleCode: string
      roleName: string
      inherited?: boolean
      refreshedAt?: string
      availableActionCount?: number
      disabledActionCount?: number
      effectiveActionCount?: number
      matched?: boolean
      disabled?: boolean
      available?: boolean
      sourcePackages?: FeaturePackageItem[]
    }

    interface UserPermissionDiagnosisAction {
      id: string
      permissionKey: string
      name?: string
      description?: string
      status?: string
      selfStatus?: string
      contextType?: string
      featureKind?: string
      moduleCode?: string
      moduleGroupStatus?: string
      featureGroupStatus?: string
      moduleGroup?: PermissionGroupItem
      featureGroup?: PermissionGroupItem
    }

    interface UserPermissionDiagnosisResult {
      permissionKey: string
      allowed: boolean
      reasonText?: string
      reasons: string[]
      matchedInSnapshot?: boolean
      bypassedBySuperAdmin?: boolean
      blockedByTeam?: boolean
      denialStage?: string
      denialReason?: string
      memberStatus?: string
      memberMatched?: boolean
      boundaryState?: string
      boundaryConfigured?: boolean
      roleChainMatched?: boolean
      roleChainDisabled?: boolean
      roleChainAvailable?: boolean
      action?: UserPermissionDiagnosisAction | null
      sourcePackages?: FeaturePackageItem[]
      roleResults?: UserPermissionRoleResult[]
    }

    interface UserPermissionMenuNode {
      id: string
      name?: string
      title?: string
      path?: string
      component?: string
      hidden?: boolean
      sort?: number
      children?: UserPermissionMenuNode[]
    }

    interface UserPermissionDiagnosisResponse {
      user: {
        id: string
        userName?: string
        nickName?: string
        status: string
        isSuperAdmin?: boolean
      }
      context: UserPermissionContext
      snapshot: UserPermissionSnapshotSummary
      roles: UserPermissionRoleResult[]
      teamMember?: {
        id?: string
        tenantId?: string
        userId?: string
        roleCode?: string
        status?: string
        matched?: boolean
      } | null
      teamPackages?: FeaturePackageItem[]
      diagnosis?: UserPermissionDiagnosisResult | null
      menus?: UserPermissionMenuNode[]
    }

    interface UserPermissionDiagnosisParams {
      tenantId?: string
      permissionKey?: string
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
      customParams?: Record<string, any>
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
      custom_params?: Record<string, any>
      status?: string
    }

    /** 更新角色参数 */
    interface RoleUpdateParams {
      code?: string
      name?: string
      description?: string
      sort_order?: number
      priority?: number
      custom_params?: Record<string, any>
      status?: string
    }

    interface PermissionGroupItem {
      id: string
      groupType: 'module' | 'feature' | string
      code: string
      name: string
      nameEn?: string
      description?: string
      status: string
      sortOrder?: number
      isBuiltin?: boolean
    }

    interface PermissionActionItem {
      id: string
      resourceCode?: string
      actionCode?: string
      moduleCode?: string
      moduleGroupId?: string
      featureGroupId?: string
      moduleGroup?: PermissionGroupItem
      featureGroup?: PermissionGroupItem
      contextType?: 'platform' | 'team' | 'common' | string
      permissionKey?: string
      featureKind?: 'system' | 'business' | string
      name: string
      description?: string
      dataPermissionCode?: string
      dataPermissionName?: string
      apiCount?: number
      pageCount?: number
      packageCount?: number
      consumerTypes?: string[]
      usagePattern?: 'unused' | 'api_only' | 'page_only' | 'package_only' | 'multi_consumer' | string
      usageNote?: string
      duplicatePattern?: 'none' | 'cross_context_mirror' | 'suspected_duplicate' | string
      duplicateGroup?: string
      duplicateKeys?: string[]
      duplicateNote?: string
      status: string
      sortOrder?: number
      isBuiltin?: boolean
      createdAt?: string
      updatedAt?: string
    }

    interface PermissionActionAuditSummary {
      totalCount: number
      unusedCount: number
      apiOnlyCount: number
      pageOnlyCount: number
      packageOnlyCount: number
      multiConsumerCount: number
      crossContextMirrorCount: number
      suspectedDuplicateCount: number
    }

    interface PermissionActionCleanupResult {
      deletedCount: number
      deletedKeys: string[]
    }

    interface PermissionActionConsumerApiItem {
      code: string
      method: string
      path: string
      summary?: string
    }

    interface PermissionActionConsumerPageItem {
      pageKey: string
      name: string
      routePath: string
      accessMode: string
    }

    interface PermissionActionConsumerPackageItem {
      id: string
      packageKey: string
      name: string
      packageType?: string
      contextType?: string
    }

    interface PermissionActionConsumerRoleItem {
      id: string
      code: string
      name: string
      contextType?: string
    }

    interface PermissionActionConsumerDetails {
      permissionKey: string
      apis: PermissionActionConsumerApiItem[]
      pages: PermissionActionConsumerPageItem[]
      featurePackages: PermissionActionConsumerPackageItem[]
      roles: PermissionActionConsumerRoleItem[]
    }

    type PermissionActionList = Api.Common.PaginatedResponse<PermissionActionItem> & {
      auditSummary?: PermissionActionAuditSummary
    }
    type PermissionGroupList = Api.Common.PaginatedResponse<PermissionGroupItem>

    interface FeaturePackageItem {
      id: string
      packageKey: string
      packageType?: 'base' | 'bundle' | string
      name: string
      description?: string
      contextType?: 'platform' | 'team' | 'common' | string
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

    interface RefreshStats {
      requestedPackageCount?: number
      impactedPackageCount?: number
      roleCount?: number
      teamCount?: number
      userCount?: number
      elapsedMilliseconds?: number
      finishedAt?: string
    }

    interface FeaturePackageRelationNode {
      id: string
      packageKey: string
      name: string
      packageType: 'base' | 'bundle' | string
      contextType: 'platform' | 'team' | 'common' | string
      status: string
      referenceCount: number
      children?: FeaturePackageRelationNode[]
    }

    interface FeaturePackageRelationTree {
      roots: FeaturePackageRelationNode[]
      cycleDependencies: string[][]
      isolatedBaseKeys: string[]
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
      context_type?: 'platform' | 'team' | 'common' | string
      status?: string
      sort_order?: number
    }

    interface FeaturePackageUpdateParams {
      package_key?: string
      package_type?: 'base' | 'bundle' | string
      name?: string
      description?: string
      context_type?: 'platform' | 'team' | 'common' | string
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

    interface FeaturePackageImpactPreview {
      packageId: string
      roleCount: number
      teamCount: number
      userCount: number
      menuCount: number
      actionCount: number
    }

    interface FeaturePackageVersionItem {
      id: string
      packageId: string
      versionNo: number
      changeType: string
      snapshot: Record<string, any>
      operatorId?: string
      requestId?: string
      createdAt?: string
    }

    interface RiskAuditItem {
      id: string
      operatorId?: string
      objectType: string
      objectId: string
      operationType: string
      beforeSummary?: Record<string, any>
      afterSummary?: Record<string, any>
      impactSummary?: Record<string, any>
      requestId?: string
      createdAt?: string
    }

    interface PermissionImpactPreview {
      permissionKey: string
      apiCount: number
      pageCount: number
      packageCount: number
      roleCount: number
      teamCount: number
      userCount: number
    }

    interface PermissionBatchUpdateParams {
      ids: string[]
      status?: string
      moduleGroupId?: string
      featureGroupId?: string
      templateName?: string
    }

    interface PermissionBatchUpdateResult {
      updatedCount: number
      skippedIds: string[]
    }

    interface PermissionBatchTemplateItem {
      id: string
      name: string
      description?: string
      payload?: Record<string, any>
      createdBy?: string
      createdAt?: string
      updatedAt?: string
    }

    interface PageMenuOptionItem {
      id: string
      name: string
      title?: string
      path?: string
      children?: PageMenuOptionItem[]
    }

    interface PageItem {
      id: string
      pageKey: string
      name: string
      routeName: string
      routePath: string
      component: string
      pageType: 'group' | 'display_group' | 'inner' | 'global' | string
      source: 'seed' | 'sync' | 'manual' | string
      moduleKey?: string
      sortOrder?: number
      parentMenuId?: string
      parentMenuName?: string
      parentPageKey?: string
      parentPageName?: string
      displayGroupKey?: string
      displayGroupName?: string
      activeMenuPath?: string
      breadcrumbMode?: 'inherit_menu' | 'inherit_page' | 'custom' | string
      accessMode?: 'inherit' | 'public' | 'jwt' | 'permission' | string
      permissionKey?: string
      inheritPermission?: boolean
      keepAlive?: boolean
      isFullPage?: boolean
      isIframe?: boolean
      isHideTab?: boolean
      link?: string
      // spaceKey 仅作为兼容展示字段；真正的空间归属优先看 spaceKeys / spaceScope。
      spaceKey?: string
      // 后端会把独立页的空间暴露编译成显式列表；为空表示默认全局共享或从父链继承。
      spaceKeys?: string[]
      // global 表示全局共享；bound 表示通过 page_space_bindings 或兼容字段显式约束。
      spaceScope?: 'global' | 'bound' | string
      spaceType?: string
      hostKey?: string
      status: 'normal' | 'suspended' | string
      meta?: Record<string, any>
      createdAt?: string
      updatedAt?: string
    }

    interface PageUnregisteredItem {
      filePath: string
      component: string
      pageKey: string
      name: string
      routeName: string
      routePath: string
      pageType: 'group' | 'display_group' | 'inner' | 'global' | string
      moduleKey?: string
      parentMenuId?: string
      parentMenuName?: string
      activeMenuPath?: string
      spaceKey?: string
      spaceKeys?: string[]
      spaceScope?: 'global' | 'bound' | string
      spaceType?: string
      hostKey?: string
    }

    interface PageSyncResult {
      createdCount: number
      skippedCount: number
      createdKeys: string[]
    }

    interface PageBreadcrumbPreviewItem {
      type: 'menu' | 'page' | string
      title: string
      path?: string
      pageKey?: string
    }

    interface PageAccessTraceParams {
      userId: string
      tenantId?: string
      pageKey?: string
      pageKeys?: string
      routePath?: string
      spaceKey?: string
    }

    interface PageAccessTraceRoleItem {
      roleId: string
      roleCode: string
      roleName: string
      status: string
    }

    interface PageAccessTracePageItem {
      pageKey: string
      pageName: string
      routePath: string
      accessMode: string
      permissionKey?: string
      parentPageKey?: string
      parentMenuId?: string
      activeMenuPath?: string
      visible: boolean
      reason: string
      matchedActionKey?: string
      effectiveChain?: string[]
    }

    interface PageAccessTraceResult {
      userId: string
      tenantId?: string
      spaceKey: string
      authenticated: boolean
      superAdmin: boolean
      actionKeyCount: number
      visibleMenuIds: string[]
      roles: PageAccessTraceRoleItem[]
      pages: PageAccessTracePageItem[]
    }

    type PageList = Api.Common.PaginatedResponse<PageItem>

    type PageSearchParams = Partial<
      Pick<PageItem, 'pageType' | 'moduleKey' | 'accessMode' | 'source' | 'status'> &
        Api.Common.CommonSearchParams & {
          keyword?: string
          parentMenuId?: string
          spaceKey?: string
        }
    >

    interface PageSaveParams {
      page_key: string
      name: string
      route_name: string
      route_path: string
      component: string
      page_type?: 'group' | 'display_group' | 'inner' | 'global' | string
      source?: 'seed' | 'sync' | 'manual' | string
      module_key?: string
      sort_order?: number
      parent_menu_id?: string
      parent_page_key?: string
      display_group_key?: string
      active_menu_path?: string
      breadcrumb_mode?: 'inherit_menu' | 'inherit_page' | 'custom' | string
      access_mode?: 'inherit' | 'public' | 'jwt' | 'permission' | string
      permission_key?: string
      inherit_permission?: boolean
      keep_alive?: boolean
      is_full_page?: boolean
      // 仅少量独立页会继续显式写入 space_key；挂到菜单/父页的页面由后端自动忽略该值。
      space_key?: string
      space_keys?: string[]
      space_type?: string
      host_key?: string
      status?: 'normal' | 'suspended' | string
      meta?: Record<string, any>
    }

    interface RuntimeNavigationManifest {
      currentSpace?: {
        space?: MenuSpaceItem
        binding?: MenuSpaceHostBindingItem
        resolvedBy?: string
        requestHost?: string
        accessGranted?: boolean
      }
      context?: {
        space_key?: string
        requested_space_key?: string
        request_host?: string
        authenticated?: boolean
        super_admin?: boolean
        user_id?: string
        tenant_id?: string
        visible_menu_count?: number
        managed_page_count?: number
        action_key_count?: number
      }
      menuTree: AppRouteRecord[]
      // entryRoutes 是给审计、兜底和后续增量注册预留的扁平入口视图；
      // 当前前端仍以 menuTree 注册树形导航，避免再拼第二套路由结构。
      entryRoutes: AppRouteRecord[]
      // managedPages 只包含非菜单直达页；菜单 entry 自己就是页面入口，不会重复出现在这里。
      managedPages: PageItem[]
      versionStamp?: string
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
      code: string
      method: string
      path: string
      spec?: string
      featureKind?: 'system' | 'business' | string
      handler?: string
      summary?: string
      permissionKey?: string
      permissionKeys?: string[]
      permissionContexts?: string[]
      permissionBindingMode?:
        | 'none'
        | 'public'
        | 'global_jwt'
        | 'self_jwt'
        | 'api_key'
        | 'single'
        | 'shared'
        | 'cross_context_shared'
        | string
      sharedAcrossContexts?: boolean
      permissionNote?: string
      authMode?: 'public' | 'jwt' | 'permission' | 'api_key' | string
      categoryId?: string
      category?: APIEndpointCategoryItem
      contextScope?: 'required' | 'forbidden' | 'optional' | string
      source?: 'sync' | 'seed' | 'manual' | string
      dataPermissionCode?: string
      dataPermissionName?: string
      runtimeExists?: boolean
      stale?: boolean
      staleReason?: string
      status: string
      createdAt?: string
      updatedAt?: string
    }

    interface APIEndpointCategoryItem {
      id: string
      code: string
      name: string
      nameEn: string
      sortOrder?: number
      status?: string
    }

    interface APIEndpointCategoryCountItem {
      categoryId: string
      count: number
    }

    interface APIEndpointOverview {
      totalCount: number
      uncategorizedCount: number
      staleCount: number
      noPermissionCount: number
      sharedPermissionCount: number
      crossContextSharedCount: number
      categoryCounts: APIEndpointCategoryCountItem[]
    }

    interface APIUnregisteredRouteItem {
      method: string
      path: string
      spec: string
      handler?: string
      hasMeta?: boolean
      meta?: {
        summary?: string
        category_code?: string
        context_scope?: string
        source?: string
        feature_kind?: string
        permission_keys?: string[]
      }
    }

    type APIEndpointList = Api.Common.PaginatedResponse<APIEndpointItem>
    type APIUnregisteredRouteList = Api.Common.PaginatedResponse<APIUnregisteredRouteItem>

    interface PermissionActionEndpointResponse {
      records: APIEndpointItem[]
      total: number
    }

    type APIEndpointSearchParams = Partial<
      Pick<APIEndpointItem, 'method' | 'path' | 'status'> &
        Api.Common.CommonSearchParams & {
          keyword?: string
          permissionKey?: string
          featureKind?: string
          categoryId?: string
          contextScope?: string
          source?: string
          permissionPattern?: string
          hasPermissionKey?: boolean
          hasCategory?: boolean
        }
    >

    type PermissionActionSearchParams = Partial<
      Pick<PermissionActionItem, 'name' | 'status'> &
        Api.Common.CommonSearchParams & {
          keyword?: string
          permissionKey?: string
          moduleCode?: string
          moduleGroupId?: string
          featureGroupId?: string
          contextType?: string
          featureKind?: string
          isBuiltin?: boolean
          usagePattern?: string
          duplicatePattern?: string
        }
    >
    interface PermissionActionCreateParams {
      permission_key: string
      module_code?: string
      module_group_id?: string
      feature_group_id?: string
      context_type?: 'platform' | 'team' | 'common' | string
      feature_kind?: 'system' | 'business' | string
      name: string
      description?: string
      status?: string
      sort_order?: number
    }

    interface PermissionActionUpdateParams {
      permission_key?: string
      module_code?: string
      module_group_id?: string
      feature_group_id?: string
      context_type?: 'platform' | 'team' | 'common' | string
      feature_kind?: 'system' | 'business' | string
      name?: string
      description?: string
      status?: string
      sort_order?: number
    }

    type PermissionGroupSearchParams = Partial<
      Pick<PermissionGroupItem, 'status'> &
        Api.Common.CommonSearchParams & {
          groupType?: string
          keyword?: string
        }
    >

    interface PermissionGroupSaveParams {
      code: string
      name: string
      name_en?: string
      description?: string
      group_type: 'module' | 'feature' | string
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

    interface FeaturePackageActionSetParams {
      action_ids: string[]
    }

    interface FeaturePackageChildSetParams {
      child_package_ids: string[]
    }

    interface TeamFeaturePackageSetParams {
      package_ids: string[]
    }

    interface FastEnterBaseItem {
      id?: string
      name: string
      enabled?: boolean
      order?: number
      routeName?: string
      link?: string
    }

    interface FastEnterApplicationItem extends FastEnterBaseItem {
      description: string
      icon: string
      iconColor: string
    }

    type FastEnterQuickLinkItem = FastEnterBaseItem

    interface FastEnterConfig {
      applications: FastEnterApplicationItem[]
      quickLinks: FastEnterQuickLinkItem[]
      minWidth?: number
    }

    /** 创建菜单参数（与后端 MenuCreateRequest 一致） */
    interface MenuCreateParams {
      parent_id: string | null
      manage_group_id?: string | null
      // directory 只做分组；entry 直接作为页面入口；external 只维护外链展示。
      kind?: 'directory' | 'entry' | 'external' | string
      path: string
      name: string
      component?: string
      title: string
      icon?: string
      sort_order?: number
      meta?: MenuMetaConfig
      space_key?: string
      space_type?: string
      host_key?: string
      hidden?: boolean
    }

    /** 更新菜单参数（与后端 MenuUpdateRequest 一致） */
    interface MenuUpdateParams {
      parent_id: string | null
      manage_group_id?: string | null
      kind?: 'directory' | 'entry' | 'external' | string
      path?: string
      name?: string
      component?: string
      title?: string
      icon?: string
      sort_order?: number
      meta?: MenuMetaConfig
      space_key?: string
      space_type?: string
      host_key?: string
      hidden?: boolean
    }

    interface APIUnregisteredScanConfig {
      enabled: boolean
      frequencyMinutes: number
      defaultCategoryId?: string
      defaultPermissionKey?: string
      markAsNoPermission: boolean
    }

    interface MenuDeleteParams {
      mode?: 'single' | 'cascade' | 'promote_children' | string
      targetParentId?: string | null
      target_parent_id?: string | null
    }

    interface MenuDeletePreviewItem {
      mode?: 'single' | 'cascade' | 'promote_children' | string
      menuCount?: number
      childCount?: number
      affectedPageCount?: number
      affectedRelationCount?: number
      menu_count?: number
      child_count?: number
      affected_page_count?: number
      affected_relation_count?: number
    }

    interface MenuBackupItem {
      id: string
      name: string
      description?: string
      // scope_type 只保留“空间 / 全局”主语义；真正的来源区分看 scope_origin。
      space_key?: string
      space_name?: string
      scope_type?: 'space' | 'global' | string
      // scope_origin 用来区分正式全空间备份和历史兼容全局备份，便于列表提示和恢复确认文案更清晰。
      scope_origin?: 'space' | 'global' | 'legacy_global' | string
      created_at?: string
      created_by?: string
    }

    interface MenuBackupCreateParams {
      name: string
      description?: string
      // scope_type 显式声明要备份当前空间还是全部空间，避免再依赖省略 space_key 推断语义。
      scope_type?: 'space' | 'global' | string
      // 仅当 scope_type=space 时需要传入当前菜单空间；global 备份会忽略该字段。
      space_key?: string
    }
  }

  namespace Message {
    type BoxType = 'notice' | 'message' | 'todo'
    type DeliveryStatus = 'unread' | 'read' | string
    type TodoStatus = 'pending' | 'done' | 'ignored' | string
    type SenderType = 'system' | 'platform_user' | 'team_user' | 'service' | 'automation' | string
    type ActionType = 'route' | 'external_link' | 'api' | 'none' | string
    type AudienceType =
      | 'all_users'
      | 'tenant_admins'
      | 'tenant_users'
      | 'specified_users'
      | 'recipient_group'
      | 'role'
      | 'feature_package'
      | string

    interface InboxSummary {
      unread_total: number
      notice_count: number
      message_count: number
      todo_count: number
    }

    interface InboxQuery extends Partial<Api.Common.CommonSearchParams> {
      box_type?: BoxType | ''
      unread_only?: boolean
    }

    interface InboxItem {
      id: string
      message_id: string
      box_type: BoxType
      delivery_status: DeliveryStatus
      todo_status?: TodoStatus
      read_at?: string
      done_at?: string
      last_action_at?: string
      recipient_team_id?: string
      title: string
      summary?: string
      content?: string
      priority?: string
      action_type?: ActionType
      action_target?: string
      message_type?: BoxType
      biz_type?: string
      scope_type?: string
      scope_id?: string
      sender_type?: SenderType
      sender_name_snapshot?: string
      sender_avatar_snapshot?: string
      sender_service_key?: string
      audience_type?: string
      audience_scope?: string
      target_tenant_id?: string
      published_at?: string
      expired_at?: string
      created_at: string
      meta?: Record<string, any>
    }

    type InboxListResponse = Api.Common.PaginatedResponse<InboxItem>
    type InboxDetail = InboxItem

    interface TodoActionParams {
      action: 'done' | 'ignored'
    }

    interface DispatchAudienceOption {
      value: AudienceType
      label: string
      description: string
    }

    interface DispatchTemplateOption {
      id: string
      template_key: string
      name: string
      description: string
      message_type: BoxType
      owner_scope: 'platform' | 'team' | string
      audience_type: AudienceType
      title_template?: string
      summary_template?: string
      content_template?: string
    }

    interface DispatchTeamOption {
      id: string
      name: string
    }

    interface DispatchUserOption {
      id: string
      name: string
      display_name: string
      description?: string
      team_id?: string
      team_name?: string
    }

    interface DispatchRecipientGroupOption {
      id: string
      name: string
      description?: string
      match_mode?: string
      estimated_count?: number
    }

    interface DispatchRoleOption {
      id: string
      code: string
      name: string
      description?: string
    }

    interface DispatchFeaturePackageOption {
      id: string
      package_key: string
      name: string
      description?: string
    }

    interface DispatchOptions {
      sender_scope: 'platform' | 'team' | string
      current_tenant_id?: string
      current_tenant_name?: string
      sender_options: DispatchSenderOption[]
      default_sender_id?: string
      audience_options: DispatchAudienceOption[]
      template_options: DispatchTemplateOption[]
      teams: DispatchTeamOption[]
      users: DispatchUserOption[]
      recipient_groups: DispatchRecipientGroupOption[]
      roles: DispatchRoleOption[]
      feature_packages: DispatchFeaturePackageOption[]
      default_message_type: BoxType
      default_audience_type: AudienceType
      default_priority: string
      supports_external_link: boolean
    }

    interface DispatchSenderOption {
      id: string
      name: string
      description?: string
      avatar_url?: string
      is_default?: boolean
    }

    interface DispatchParams {
      sender_id?: string
      template_id?: string
      template_key?: string
      message_type: BoxType
      audience_type: AudienceType
      target_tenant_ids?: string[]
      target_user_ids?: string[]
      target_group_ids?: string[]
      title: string
      summary?: string
      content?: string
      priority?: string
      action_type?: ActionType
      action_target?: string
      biz_type?: string
      expired_at?: string
    }

    interface DispatchResult {
      message_id: string
      delivery_count: number
      dispatch_status: 'queued' | 'processing' | 'published' | 'failed' | string
    }

    interface MessageTemplateQuery extends Partial<Api.Common.CommonSearchParams> {
      keyword?: string
    }

    interface MessageTemplateItem {
      id: string
      template_key: string
      name: string
      description?: string
      message_type: BoxType
      owner_scope: 'platform' | 'team' | string
      owner_tenant_id?: string
      owner_tenant_name?: string
      audience_type: AudienceType
      title_template?: string
      summary_template?: string
      content_template?: string
      status: 'normal' | 'disabled' | string
      editable: boolean
      meta?: Record<string, any>
      created_at?: string
      updated_at?: string
    }

    type MessageTemplateListResponse = Api.Common.PaginatedResponse<MessageTemplateItem>

    interface MessageTemplateSaveParams {
      template_key?: string
      name: string
      description?: string
      message_type: BoxType
      audience_type: AudienceType
      title_template?: string
      summary_template?: string
      content_template?: string
      status?: 'normal' | 'disabled' | string
    }

    interface MessageSenderItem {
      id: string
      scope_type: 'platform' | 'team' | string
      scope_id?: string
      name: string
      description?: string
      avatar_url?: string
      is_default?: boolean
      status: 'normal' | 'disabled' | string
      editable: boolean
      meta?: Record<string, any>
      created_at?: string
      updated_at?: string
    }

    interface MessageSenderListResponse {
      records: MessageSenderItem[]
    }

    interface MessageSenderSaveParams {
      name: string
      description?: string
      avatar_url?: string
      is_default?: boolean
      status?: 'normal' | 'disabled' | string
      meta?: Record<string, any>
    }

    interface MessageRecipientGroupTargetItem {
      id: string
      target_type: 'user' | 'tenant_users' | 'tenant_admins' | string
      user_id?: string
      user_name?: string
      tenant_id?: string
      tenant_name?: string
      role_code?: string
      role_name?: string
      package_key?: string
      package_name?: string
      sort_order?: number
      meta?: Record<string, any>
    }

    interface MessageRecipientGroupItem {
      id: string
      scope_type: 'platform' | 'team' | string
      scope_id?: string
      name: string
      description?: string
      match_mode?: 'manual' | string
      status: 'normal' | 'disabled' | string
      editable: boolean
      estimated_count?: number
      meta?: Record<string, any>
      targets: MessageRecipientGroupTargetItem[]
      created_at?: string
      updated_at?: string
    }

    interface MessageRecipientGroupListResponse {
      records: MessageRecipientGroupItem[]
    }

    interface MessageRecipientGroupTargetSaveParams {
      target_type: 'user' | 'tenant_users' | 'tenant_admins' | 'role' | 'feature_package' | string
      user_id?: string
      tenant_id?: string
      role_code?: string
      package_key?: string
      sort_order?: number
      meta?: Record<string, any>
    }

    interface MessageRecipientGroupSaveParams {
      name: string
      description?: string
      match_mode?: 'manual' | string
      status?: 'normal' | 'disabled' | string
      meta?: Record<string, any>
      targets: MessageRecipientGroupTargetSaveParams[]
    }

    interface DispatchRecordQuery extends Partial<Api.Common.CommonSearchParams> {
      keyword?: string
      message_type?: BoxType | ''
      audience_type?: AudienceType | ''
    }

    interface DispatchRecordSummary {
      total_messages: number
      total_deliveries: number
      read_deliveries: number
      todo_messages: number
    }

    interface DispatchRecordItem {
      id: string
      title: string
      summary?: string
      content?: string
      message_type: BoxType
      audience_type: AudienceType
      scope_type: 'platform' | 'team' | string
      scope_id?: string
      target_tenant_id?: string
      target_tenant_name?: string
      sender_name?: string
      template_name?: string
      priority?: string
      status?: string
      published_at?: string
      created_at: string
      delivery_count: number
      read_count: number
      unread_count: number
      pending_todo_count: number
    }

    interface DispatchRecordDeliveryItem {
      id: string
      recipient_user_id: string
      recipient_name: string
      recipient_team_id?: string
      recipient_team_name?: string
      delivery_status: DeliveryStatus
      todo_status?: TodoStatus
      read_at?: string
      done_at?: string
      last_action_at?: string
      source_group_id?: string
      source_group_name?: string
      source_rule_type?: string
      source_rule_label?: string
      source_target_id?: string
      source_target_type?: string
      source_target_value?: string
    }

    interface DispatchRecordDetail extends DispatchRecordItem {
      deliveries: DispatchRecordDeliveryItem[]
    }

    interface DispatchRecordListResponse extends Api.Common.PaginatedResponse<DispatchRecordItem> {
      summary: DispatchRecordSummary
    }
  }
}
