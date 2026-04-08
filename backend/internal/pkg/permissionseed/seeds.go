package permissionseed

import (
	"crypto/sha1"
	"strings"

	"github.com/google/uuid"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
)

type PermissionKeySeed struct {
	ID               uuid.UUID
	Key              string
	Name             string
	Description      string
	ContextType      string
	ModuleCode       string
	ModuleGroupCode  string
	FeatureGroupCode string
	FeatureKind      string
	Status           string
	SortOrder        int
	IsBuiltin        bool
}

type PermissionGroupSeed struct {
	ID          uuid.UUID
	GroupType   string
	Code        string
	Name        string
	NameEn      string
	Description string
	Status      string
	SortOrder   int
	IsBuiltin   bool
}

type FeaturePackageSeed struct {
	ID             uuid.UUID
	PackageKey     string
	PackageType    string
	Name           string
	Description    string
	WorkspaceScope string
	ContextType    string
	AppKeys        []string
	IsBuiltin      bool
	Status         string
	SortOrder      int
	MenuNames      []string
	PermissionKeys []string
}

type FeaturePackageBundleSeed struct {
	ParentPackageKey string
	ChildPackageKey  string
}

type MenuSeed struct {
	SpaceKey   string
	Kind       string
	Name       string
	ParentName string
	Path       string
	Component  string
	Title      string
	Icon       string
	SortOrder  int
	Meta       usermodel.MetaJSON
}

type MenuSpaceSeed struct {
	SpaceKey        string
	Name            string
	Description     string
	DefaultHomePath string
	IsDefault       bool
	Status          string
	Meta            usermodel.MetaJSON
}

type RolePackageBindingSeed struct {
	ID         uuid.UUID
	RoleCode   string
	PackageKey string
}

type APIEndpointCategorySeed struct {
	ID        uuid.UUID
	Code      string
	Name      string
	NameEn    string
	SortOrder int
	Status    string
}

type PageSeed struct {
	SpaceKey          string
	SpaceKeys         []string
	PageKey           string
	Name              string
	RouteName         string
	RoutePath         string
	Component         string
	PageType          string
	VisibilityScope   string
	Source            string
	ModuleKey         string
	SortOrder         int
	ParentMenuName    string
	ParentPageKey     string
	DisplayGroupKey   string
	ActiveMenuPath    string
	BreadcrumbMode    string
	AccessMode        string
	PermissionKey     string
	InheritPermission bool
	KeepAlive         bool
	IsFullPage        bool
	Status            string
	Meta              usermodel.MetaJSON
}

func StableID(kind, key string) uuid.UUID {
	target := strings.TrimSpace(kind) + ":" + strings.TrimSpace(key)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("permission-seed:"+target), 5)
}

func DefaultAPIEndpointCategories() []APIEndpointCategorySeed {
	items := []APIEndpointCategorySeed{
		{Code: "auth", Name: "认证", NameEn: "Authentication", SortOrder: 10, Status: "normal"},
		{Code: "system", Name: "系统", NameEn: "System", SortOrder: 20, Status: "normal"},
		{Code: "user", Name: "用户", NameEn: "User", SortOrder: 30, Status: "normal"},
		{Code: "role", Name: "角色", NameEn: "Role", SortOrder: 40, Status: "normal"},
		{Code: "permission_key", Name: "功能键", NameEn: "Permission Key", SortOrder: 50, Status: "normal"},
		{Code: "feature_package", Name: "功能包", NameEn: "Feature Package", SortOrder: 60, Status: "normal"},
		{Code: "api_endpoint", Name: "API 管理", NameEn: "API Management", SortOrder: 70, Status: "normal"},
		{Code: "page", Name: "页面", NameEn: "Page", SortOrder: 80, Status: "normal"},
		{Code: "menu", Name: "菜单", NameEn: "Menu", SortOrder: 90, Status: "normal"},
		{Code: "media", Name: "媒体", NameEn: "Media", SortOrder: 110, Status: "normal"},
		{Code: "collaboration_workspace", Name: "协作空间", NameEn: "Collaboration Workspace", SortOrder: 120, Status: "normal"},
		{Code: "message", Name: "消息", NameEn: "Message", SortOrder: 130, Status: "normal"},
		{Code: "navigation", Name: "导航运行时", NameEn: "Navigation Runtime", SortOrder: 140, Status: "normal"},
		{Code: "workspace", Name: "工作空间", NameEn: "Workspace", SortOrder: 150, Status: "normal"},
		{Code: "uncategorized", Name: "未分类", NameEn: "Uncategorized", SortOrder: 999, Status: "normal"},
	}
	for i := range items {
		items[i].ID = StableID("api-endpoint-category", items[i].Code)
	}
	return items
}

func DefaultPermissionKeys() []PermissionKeySeed {
	items := []PermissionKeySeed{
		newPermissionKeySeed("role", "list", "查看角色列表", "允许查看角色列表"),
		newPermissionKeySeed("role", "get", "查看角色详情", "允许查看角色详情"),
		newPermissionKeySeed("role", "create", "创建角色", "允许创建角色"),
		newPermissionKeySeed("role", "update", "更新角色", "允许更新角色"),
		newPermissionKeySeed("role", "delete", "删除角色", "允许删除角色"),
		newPermissionKeySeed("role", "assign_menu", "配置角色菜单权限", "允许为角色配置菜单权限"),
		newPermissionKeySeed("role", "assign_action", "配置角色功能权限", "允许为角色配置功能权限"),
		newPermissionKeySeed("role", "assign_data", "配置角色数据权限", "允许为角色配置数据权限"),
		newPermissionKeySeed("permission_key", "list", "查看功能权限列表", "允许查看功能权限列表"),
		newPermissionKeySeed("permission_key", "get", "查看功能权限详情", "允许查看功能权限详情"),
		newPermissionKeySeed("permission_key", "create", "创建功能权限", "允许创建功能权限"),
		newPermissionKeySeed("permission_key", "update", "更新功能权限", "允许更新功能权限"),
		newPermissionKeySeed("permission_key", "delete", "删除功能权限", "允许删除功能权限"),
		newPermissionKeySeed("user", "list", "查看用户列表", "允许查看用户列表"),
		newPermissionKeySeed("user", "get", "查看用户详情", "允许查看用户详情"),
		newPermissionKeySeed("user", "create", "创建用户", "允许创建用户"),
		newPermissionKeySeed("user", "update", "更新用户", "允许更新用户"),
		newPermissionKeySeed("user", "delete", "删除用户", "允许删除用户"),
		newPermissionKeySeed("user", "assign_role", "分配用户角色", "允许为用户分配角色"),
		newPermissionKeySeed("menu", "list", "查看菜单管理树", "允许查看全部菜单管理树"),
		newPermissionKeySeed("menu", "create", "创建菜单", "允许创建菜单"),
		newPermissionKeySeed("menu", "update", "更新菜单", "允许更新菜单"),
		newPermissionKeySeed("menu", "delete", "删除菜单", "允许删除菜单"),
		newPermissionKeySeed("system", "view_page_catalog", "查看页面文件映射", "允许查看页面文件映射"),
		newPermissionKeySeed("collaboration_workspace", "list", "查看协作空间列表", "允许查看协作空间列表"),
		newPermissionKeySeed("collaboration_workspace", "get", "查看协作空间详情", "允许查看协作空间详情"),
		newPermissionKeySeed("collaboration_workspace", "create", "创建协作空间", "允许创建协作空间"),
		newPermissionKeySeed("collaboration_workspace", "update", "更新协作空间", "允许更新协作空间"),
		newPermissionKeySeed("collaboration_workspace", "delete", "删除协作空间", "允许删除协作空间"),
		newPermissionKeySeed("collaboration_workspace", "configure_action_boundary", "配置协作空间功能权限边界", "允许配置协作空间功能权限边界"),
		newPermissionKeySeed("collaboration_workspace_member_admin", "list", "查看协作空间成员列表", "允许在系统管理中查看协作空间成员列表"),
		newPermissionKeySeed("collaboration_workspace_member_admin", "create", "添加协作空间成员", "允许在系统管理中添加协作空间成员"),
		newPermissionKeySeed("collaboration_workspace_member_admin", "delete", "移除协作空间成员", "允许在系统管理中移除协作空间成员"),
		newPermissionKeySeed("collaboration_workspace_member_admin", "update_role", "更新协作空间成员身份", "允许在系统管理中更新协作空间成员身份"),
		newPermissionKeySeed("collaboration_workspace_member", "create", "添加当前协作空间成员", "允许在当前协作空间中添加成员"),
		newPermissionKeySeed("collaboration_workspace_member", "delete", "移除当前协作空间成员", "允许在当前协作空间中移除成员"),
		newPermissionKeySeed("collaboration_workspace_member", "update_role", "更新当前协作空间成员身份", "允许在当前协作空间中更新成员身份"),
		newPermissionKeySeed("collaboration_workspace_member", "assign_role", "配置当前协作空间成员角色", "允许在当前协作空间中配置成员角色"),
		newPermissionKeySeed("collaboration_workspace_member", "assign_action", "配置当前协作空间成员功能权限", "允许在当前协作空间中配置成员功能权限"),
		newPermissionKeySeed("collaboration_workspace_boundary", "configure_action_boundary", "查看和配置当前协作空间功能权限边界", "允许查看和配置当前协作空间功能权限边界"),
		newPermissionKeySeed("api_endpoint", "list", "查看 API 注册表", "允许查看 API 注册表"),
		newPermissionKeySeed("api_endpoint", "sync", "同步 API 注册表", "允许同步 API 注册表"),
		newPermissionKeySeed("feature_package", "manage", "个人空间功能包管理", "允许查看和维护个人空间功能包"),
		newPermissionKeySeed("feature_package", "assign_collaboration_workspace", "配置协作空间功能包", "允许给协作空间开通功能包"),
		newPermissionKeySeed("page", "list", "查看页面管理列表", "允许查看页面管理列表"),
		newPermissionKeySeed("page", "get", "查看页面详情", "允许查看页面详情"),
		newPermissionKeySeed("page", "create", "创建页面", "允许创建页面"),
		newPermissionKeySeed("page", "update", "更新页面", "允许更新页面"),
		newPermissionKeySeed("page", "delete", "删除页面", "允许删除页面"),
		newPermissionKeySeed("page", "sync", "同步页面注册表", "允许同步页面注册表"),
		newPermissionKeySeed("fast_enter", "manage", "快捷应用管理", "允许维护顶部快捷应用和快捷链接配置"),
		newPermissionKeySeed("message", "manage", "消息管理", "允许发送站内通知、消息和待办"),
		newPermissionKeySeed("collaboration_workspace_message", "manage", "协作空间消息发送", "允许协作空间管理员向当前协作空间发送站内通知、消息和待办"),
		newPermissionKeySeed("system_permission", "manage_action_registry", "管理功能权限注册表", "允许维护功能权限注册信息"),
		newPermissionKeySeed("system_permission", "assign_role_action", "配置角色功能权限", "允许为角色分配功能权限"),
	}
	for i := range items {
		items[i].ID = StableID("permission-key", items[i].Key)
		items[i].SortOrder = i + 1
	}
	return items
}

func DefaultPermissionGroups() []PermissionGroupSeed {
	items := append(DefaultPermissionFeatureGroups(), DefaultPermissionModuleGroups()...)
	for i := range items {
		items[i].ID = StableID("permission-group", items[i].GroupType+":"+items[i].Code)
	}
	return items
}

func DefaultPermissionFeatureGroups() []PermissionGroupSeed {
	return []PermissionGroupSeed{
		{
			GroupType:   "feature",
			Code:        "system",
			Name:        "系统功能",
			NameEn:      "System Feature",
			Description: "系统初始化、权限治理与后台管理能力",
			Status:      "normal",
			SortOrder:   1,
			IsBuiltin:   true,
		},
		{
			GroupType:   "feature",
			Code:        "business",
			Name:        "业务功能",
			NameEn:      "Business Feature",
			Description: "面向业务域扩展的功能能力",
			Status:      "normal",
			SortOrder:   2,
			IsBuiltin:   true,
		},
	}
}

func DefaultPermissionModuleGroups() []PermissionGroupSeed {
	return []PermissionGroupSeed{
		{
			GroupType:   "module",
			Code:        "role",
			Name:        "角色管理",
			NameEn:      "Role Management",
			Description: "系统角色、角色菜单、角色功能权限与数据权限管理",
			Status:      "normal",
			SortOrder:   100,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "permission_key",
			Name:        "权限键管理",
			NameEn:      "Permission Key Management",
			Description: "功能权限键、功能分组与接口绑定管理",
			Status:      "normal",
			SortOrder:   110,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "user",
			Name:        "用户管理",
			NameEn:      "User Management",
			Description: "个人空间用户、角色与例外权限管理",
			Status:      "normal",
			SortOrder:   120,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "menu",
			Name:        "菜单管理",
			NameEn:      "Menu Management",
			Description: "个人空间菜单树维护与菜单权限治理",
			Status:      "normal",
			SortOrder:   130,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "system",
			Name:        "系统工具",
			NameEn:      "System Utilities",
			Description: "系统页映射与系统工具能力",
			Status:      "normal",
			SortOrder:   150,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "collaboration_workspace",
			Name:        "协作空间管理",
			NameEn:      "Collaboration Workspace Management",
			Description: "协作空间、协作空间边界与协作空间主体治理",
			Status:      "normal",
			SortOrder:   160,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "collaboration_workspace_member_admin",
			Name:        "协作空间成员管理",
			NameEn:      "Collaboration Workspace Member Admin",
			Description: "协作空间成员与身份配置",
			Status:      "normal",
			SortOrder:   170,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "collaboration_workspace_member",
			Name:        "当前协作空间成员",
			NameEn:      "Current Collaboration Workspace Members",
			Description: "当前协作空间上下文内的成员管理与成员权限配置",
			Status:      "normal",
			SortOrder:   180,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "collaboration_workspace_boundary",
			Name:        "协作空间边界",
			NameEn:      "Collaboration Workspace Boundary",
			Description: "当前协作空间功能边界与协作空间角色能力边界管理",
			Status:      "normal",
			SortOrder:   190,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "api_endpoint",
			Name:        "API 管理",
			NameEn:      "API Management",
			Description: "API 单元、未注册路由与接口元数据管理",
			Status:      "normal",
			SortOrder:   200,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "page",
			Name:        "页面管理",
			NameEn:      "Page Management",
			Description: "页面注册表、数据内页与全局页管理",
			Status:      "normal",
			SortOrder:   205,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "fast_enter",
			Name:        "快捷应用管理",
			NameEn:      "Fast Enter Management",
			Description: "顶部快捷应用和快捷链接配置管理",
			Status:      "normal",
			SortOrder:   207,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "message",
			Name:        "消息管理",
			NameEn:      "Message Management",
			Description: "站内消息发送、模板归属与收件对象管理",
			Status:      "normal",
			SortOrder:   208,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "collaboration_workspace_message",
			Name:        "协作空间消息",
			NameEn:      "Collaboration Workspace Message",
			Description: "协作空间管理菜单下的消息发送能力",
			Status:      "normal",
			SortOrder:   209,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "feature_package",
			Name:        "功能包管理",
			NameEn:      "Feature Package Management",
			Description: "功能包、组合包与开通关系管理",
			Status:      "normal",
			SortOrder:   210,
			IsBuiltin:   true,
		},
		{
			GroupType:   "module",
			Code:        "system_permission",
			Name:        "权限系统配置",
			NameEn:      "Permission System Settings",
			Description: "权限注册表与系统级权限配置能力",
			Status:      "normal",
			SortOrder:   220,
			IsBuiltin:   true,
		},
	}
}

func DefaultFeaturePackages() []FeaturePackageSeed {
	items := []FeaturePackageSeed{
		{
			PackageKey:     "platform_admin.system_manage",
			PackageType:    "base",
			Name:           "平台管理员系统管理包",
			Description:    "包含平台管理员系统管理核心能力",
			WorkspaceScope: "all",
			ContextType:    "common",
			AppKeys:        []string{systemmodels.DefaultAppKey},
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      1,
			MenuNames:      []string{"System", "SystemAccess", "Role", "User", "ActionPermission", "FeaturePackage", "CollaborationWorkspaceRoot", "CollaborationWorkspaceManage", "CollaborationWorkspaceRolesAndPermissions", "SystemNavigation", "AppManage", "PageManagement", "FastEnterManage", "MenuSpaceManage", "SystemIntegration", "MessageManage"},
			PermissionKeys: []string{"system.user.manage", "system.role.manage", "system.permission.manage", "system.page.manage", "system.page.sync", "system.fast_enter.manage", "message.manage", "collaboration_workspace.manage"},
		},
		{
			PackageKey:     "platform_admin.menu_manage",
			PackageType:    "base",
			Name:           "平台管理员菜单管理包",
			Description:    "包含平台管理员菜单管理与菜单备份能力",
			WorkspaceScope: "all",
			ContextType:    "common",
			AppKeys:        []string{systemmodels.DefaultAppKey},
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      2,
			MenuNames:      []string{"System", "SystemNavigation", "AppManage", "Menus", "MenuSpaceManage"},
			PermissionKeys: []string{"system.menu.manage", "system.menu.backup"},
		},
		{
			PackageKey:     "platform_admin.api_manage",
			PackageType:    "base",
			Name:           "平台管理员接口管理包",
			Description:    "包含 API 注册表查看、同步与平台功能包治理能力",
			WorkspaceScope: "all",
			ContextType:    "common",
			AppKeys:        []string{systemmodels.DefaultAppKey},
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      3,
			MenuNames:      []string{"System", "SystemAccess", "FeaturePackage", "SystemIntegration", "ApiEndpoint"},
			PermissionKeys: []string{"system.api_registry.view", "system.api_registry.sync", "feature_package.manage", "feature_package.assign_collaboration_workspace"},
		},
		{
			PackageKey:     "collaboration_workspace.member_admin",
			PackageType:    "base",
			Name:           "协作空间成员管理包",
			Description:    "包含协作空间成员、角色和功能权限配置能力",
			WorkspaceScope: "collaboration",
			ContextType:    "collaboration",
			AppKeys:        []string{systemmodels.DefaultAppKey},
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      10,
			MenuNames:      []string{"CollaborationWorkspaceRoot", "CollaborationWorkspaceMembers", "CollaborationWorkspaceRolesAndPermissions", "CollaborationWorkspaceMessageManage"},
			PermissionKeys: []string{"collaboration_workspace.member.manage", "collaboration_workspace.boundary.manage", "collaboration_workspace.message.manage"},
		},
		{
			PackageKey:     "platform_admin.admin_bundle",
			PackageType:    "bundle",
			Name:           "平台管理员组合包",
			Description:    "统一聚合平台系统、菜单与接口管理基础包",
			WorkspaceScope: "all",
			ContextType:    "common",
			AppKeys:        []string{systemmodels.DefaultAppKey},
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      4,
		},
	}
	for i := range items {
		items[i].ID = StableID("feature-package", items[i].PackageKey)
	}
	return items
}

func DefaultFeaturePackageBundles() []FeaturePackageBundleSeed {
	return []FeaturePackageBundleSeed{
		{ParentPackageKey: "platform_admin.admin_bundle", ChildPackageKey: "platform_admin.system_manage"},
		{ParentPackageKey: "platform_admin.admin_bundle", ChildPackageKey: "platform_admin.menu_manage"},
		{ParentPackageKey: "platform_admin.admin_bundle", ChildPackageKey: "platform_admin.api_manage"},
	}
}

func DefaultMenus() []MenuSeed {
	metaSuperAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}
	metaCollaborationWorkspaceAccessOnly := usermodel.MetaJSON{"keepAlive": true}
	metaJWT := usermodel.MetaJSON{"accessMode": "jwt"}
	items := []MenuSeed{
		{Name: "Dashboard", Path: "/dashboard", Component: "/index/index", Title: "menus.dashboard.title", Icon: "ri:pie-chart-line", SortOrder: 1, Meta: metaJWT},
		{Name: "System", Path: "/system", Component: "/index/index", Title: "menus.system.title", Icon: "ri:user-3-line", SortOrder: 2, Meta: metaSuperAdminAndAdmin},
		{Name: "CollaborationWorkspaceRoot", Path: "/collaboration-workspace", Component: "/index/index", Title: "协作空间", Icon: "ri:group-line", SortOrder: 5, Meta: metaCollaborationWorkspaceAccessOnly},
		{Name: "Console", ParentName: "Dashboard", Path: "console", Component: "/dashboard/console", Title: "menus.dashboard.console", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": false, "fixedTab": true}},
		{Name: "WorkspaceInbox", ParentName: "Dashboard", Path: "/workspace/inbox", Component: "/workspace/inbox", Title: "消息中心", SortOrder: 3, Meta: usermodel.MetaJSON{"isHide": true, "keepAlive": true, "accessMode": "jwt"}},
		{Name: "SystemAccess", ParentName: "System", Path: "access", Component: "", Title: "身份与权限", Icon: "ri:shield-user-line", SortOrder: 1, Meta: metaSuperAdminAndAdmin},
		{Name: "SystemNavigation", ParentName: "System", Path: "navigation", Component: "", Title: "导航与界面", Icon: "ri:layout-grid-line", SortOrder: 2, Meta: metaSuperAdmin},
		{Name: "SystemIntegration", ParentName: "System", Path: "integration", Component: "", Title: "接口与注册", Icon: "ri:link-m", SortOrder: 3, Meta: metaSuperAdmin},
		{Name: "Role", ParentName: "SystemAccess", Path: "/system/role", Component: "/system/role", Title: "menus.system.role", SortOrder: 1, Meta: metaSuperAdmin},
		{Name: "User", ParentName: "SystemAccess", Path: "/system/user", Component: "/system/user", Title: "menus.system.user", SortOrder: 2, Meta: metaSuperAdminAndAdmin},
		{Name: "CollaborationWorkspaceManage", ParentName: "CollaborationWorkspaceRoot", Path: "workspaces", Component: "/collaboration-workspace/workspace", Title: "协作空间管理", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": true, "accessMode": "permission", "requiredAction": "collaboration_workspace.manage"}},
		{Name: "CollaborationWorkspaceRolesAndPermissions", ParentName: "CollaborationWorkspaceRoot", Path: "roles", Component: "/system/collaboration-workspace-roles-permissions", Title: "协作空间角色与权限", SortOrder: 3, Meta: usermodel.MetaJSON{"keepAlive": true, "accessMode": "permission", "requiredAction": "collaboration_workspace.manage"}},
		{Name: "AppManage", ParentName: "SystemNavigation", Path: "/system/app", Component: "/system/app", Title: "应用管理", SortOrder: 1, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "Menus", ParentName: "SystemNavigation", Path: "/system/menu", Component: "/system/menu", Title: "menus.system.menu", SortOrder: 2, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "ActionPermission", ParentName: "SystemAccess", Path: "/system/action-permission", Component: "/system/action-permission", Title: "功能权限", SortOrder: 3, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "ApiEndpoint", ParentName: "SystemIntegration", Path: "/system/api-endpoint", Component: "/system/api-endpoint", Title: "API管理", SortOrder: 1, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "MessageManage", ParentName: "SystemIntegration", Path: "/system/message", Component: "/system/message", Title: "消息发送", SortOrder: 2, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "CollaborationWorkspaceMessageManage", ParentName: "CollaborationWorkspaceRoot", Path: "message", Component: "/collaboration-workspace/message", Title: "协作空间消息发送", SortOrder: 4, Meta: metaCollaborationWorkspaceAccessOnly},
		{Name: "FeaturePackage", ParentName: "SystemAccess", Path: "/system/feature-package", Component: "/system/feature-package", Title: "功能包管理", SortOrder: 4, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "PageManagement", ParentName: "SystemNavigation", Path: "/system/page", Component: "/system/page", Title: "页面管理", SortOrder: 3, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "FastEnterManage", ParentName: "SystemNavigation", Path: "/system/fast-enter", Component: "/system/fast-enter", Title: "快捷应用管理", SortOrder: 4, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "MenuSpaceManage", ParentName: "SystemNavigation", Path: "/system/menu-space", Component: "/system/menu-space", Title: "高级空间配置", SortOrder: 90, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "CollaborationWorkspaceMembers", ParentName: "CollaborationWorkspaceRoot", Path: "members", Component: "/collaboration-workspace/members", Title: "协作空间成员", SortOrder: 2, Meta: metaCollaborationWorkspaceAccessOnly},
	}
	for i := range items {
		if strings.TrimSpace(items[i].SpaceKey) == "" {
			items[i].SpaceKey = systemmodels.DefaultMenuSpaceKey
		}
		if strings.TrimSpace(items[i].Kind) == "" {
			items[i].Kind = deriveMenuSeedKind(items[i])
		}
	}
	return items
}

func deriveMenuSeedKind(item MenuSeed) string {
	link := ""
	if item.Meta != nil {
		if value, ok := item.Meta["link"].(string); ok {
			link = strings.TrimSpace(value)
		}
	}
	switch {
	case link != "":
		return systemmodels.MenuKindExternal
	case strings.TrimSpace(item.Component) != "" && strings.TrimSpace(item.Component) != "/index/index":
		return systemmodels.MenuKindEntry
	default:
		return systemmodels.MenuKindDirectory
	}
}

func DefaultMenuSpaces() []MenuSpaceSeed {
	return []MenuSpaceSeed{
		{
			SpaceKey:        systemmodels.DefaultMenuSpaceKey,
			Name:            "默认菜单空间",
			Description:     "兼容当前单域单菜单运行模式",
			DefaultHomePath: "/dashboard/console",
			IsDefault:       true,
			Status:          "normal",
			Meta:            usermodel.MetaJSON{},
		},
		{
			SpaceKey:        "ops",
			Name:            "运营空间",
			Description:     "用于验证多空间导航与空间级页面可见性。",
			DefaultHomePath: "/dashboard/console",
			IsDefault:       false,
			Status:          "normal",
			Meta:            usermodel.MetaJSON{},
		},
	}
}

func DeprecatedDefaultMenuNames() []string {
	return []string{"PageAssociation", "CollaborationWorkspaceManagementRedirect", "Scope", "UserCenter"}
}

func DefaultRolePackageBindings() []RolePackageBindingSeed {
	items := []RolePackageBindingSeed{
		{RoleCode: "admin", PackageKey: "platform_admin.admin_bundle"},
		{RoleCode: "collaboration_workspace_admin", PackageKey: "collaboration_workspace.member_admin"},
	}
	for i := range items {
		items[i].ID = StableID("role-package-binding", items[i].RoleCode+":"+items[i].PackageKey)
	}
	return items
}

func DefaultFeaturePackageKeys() []string {
	items := DefaultFeaturePackages()
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.PackageKey)
	}
	return keys
}

func DefaultRoleCodes() []string {
	bindings := DefaultRolePackageBindings()
	result := make([]string, 0, len(bindings))
	seen := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		roleCode := strings.TrimSpace(binding.RoleCode)
		if roleCode == "" {
			continue
		}
		if _, ok := seen[roleCode]; ok {
			continue
		}
		seen[roleCode] = struct{}{}
		result = append(result, roleCode)
	}
	return result
}

func DefaultPages() []PageSeed {
	// 页面种子只保留非菜单直达页。
	// 常规入口页直接由 DefaultMenus 维护，避免菜单和 ui_pages 双写同一路由。
	items := []PageSeed{
		{
			PageKey:           "display.system_pages",
			Name:              "系统页面",
			RouteName:         "display.system_pages",
			PageType:          "display_group",
			VisibilityScope:   "app",
			Source:            "manual",
			SortOrder:         10,
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "inherit",
			InheritPermission: true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "workspace.user_center",
			Name:              "个人中心",
			RouteName:         "UserCenter",
			RoutePath:         "/user-center",
			Component:         "/system/user-center",
			PageType:          "standalone",
			VisibilityScope:   "app",
			Source:            "manual",
			ModuleKey:         "account",
			SortOrder:         15,
			DisplayGroupKey:   "display.system_pages",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "jwt",
			InheritPermission: true,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{"isHideTab": true},
		},
		{
			PageKey:           "workspace.ops_console",
			Name:              "运营空间控制台",
			RouteName:         "WorkspaceOpsConsole",
			RoutePath:         "/workspace/ops-console",
			Component:         "/system/access-trace",
			PageType:          "standalone",
			VisibilityScope:   "spaces",
			SpaceKeys:         []string{"ops"},
			Source:            "manual",
			ModuleKey:         "system",
			SortOrder:         18,
			DisplayGroupKey:   "display.system_pages",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "page.list",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "system.message.template.manage",
			Name:              "消息模板",
			RouteName:         "MessageTemplateManage",
			RoutePath:         "/system/message-template",
			Component:         "/system/message-template",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         35,
			ParentMenuName:    "MessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "system.message.sender.manage",
			Name:              "发送人管理",
			RouteName:         "MessageSenderManage",
			RoutePath:         "/system/message-sender",
			Component:         "/system/message-sender",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         34,
			ParentMenuName:    "MessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "system.message.recipient_group.manage",
			Name:              "接收组管理",
			RouteName:         "MessageRecipientGroupManage",
			RoutePath:         "/system/message-recipient-group",
			Component:         "/system/message-recipient-group",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         33,
			ParentMenuName:    "MessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "system.message.record.manage",
			Name:              "消息发送记录",
			RouteName:         "MessageRecordManage",
			RoutePath:         "/system/message-record",
			Component:         "/system/message-record",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         36,
			ParentMenuName:    "MessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "system.message.more",
			Name:              "系统更多入口",
			RouteName:         "SystemMore",
			RoutePath:         "/system/more",
			Component:         "/system/more",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         37,
			ParentMenuName:    "MessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "collaboration_workspace.message.template.manage",
			Name:              "协作空间消息模板",
			RouteName:         "CollaborationWorkspaceMessageTemplateManage",
			RoutePath:         "/collaboration-workspace/message-template",
			Component:         "/collaboration-workspace/message-template",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace_message",
			SortOrder:         45,
			ParentMenuName:    "CollaborationWorkspaceMessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "collaboration_workspace.message.sender.manage",
			Name:              "协作空间发送人管理",
			RouteName:         "CollaborationWorkspaceMessageSenderManage",
			RoutePath:         "/collaboration-workspace/message-sender",
			Component:         "/collaboration-workspace/message-sender",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace_message",
			SortOrder:         44,
			ParentMenuName:    "CollaborationWorkspaceMessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "collaboration_workspace.message.recipient_group.manage",
			Name:              "协作空间接收组管理",
			RouteName:         "CollaborationWorkspaceMessageRecipientGroupManage",
			RoutePath:         "/collaboration-workspace/message-recipient-group",
			Component:         "/collaboration-workspace/message-recipient-group",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace_message",
			SortOrder:         43,
			ParentMenuName:    "CollaborationWorkspaceMessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "collaboration_workspace.message.record.manage",
			Name:              "协作空间发送记录",
			RouteName:         "CollaborationWorkspaceMessageRecordManage",
			RoutePath:         "/collaboration-workspace/message-record",
			Component:         "/collaboration-workspace/message-record",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace_message",
			SortOrder:         46,
			ParentMenuName:    "CollaborationWorkspaceMessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			PageKey:           "collaboration_workspace.message.more",
			Name:              "协作空间更多入口",
			RouteName:         "CollaborationWorkspaceMore",
			RoutePath:         "/collaboration-workspace/more",
			Component:         "/collaboration-workspace/more",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace_message",
			SortOrder:         47,
			ParentMenuName:    "CollaborationWorkspaceMessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
	}
	return items
}

// LegacyMenuBackedPages 仅供历史迁移回放使用。
// 这些页面已被菜单 entry 接管，默认初始化与新数据不再写入 ui_pages。
func LegacyMenuBackedPages() []PageSeed {
	items := []PageSeed{
		{
			SpaceKey:          systemmodels.DefaultMenuSpaceKey,
			PageKey:           "workspace.inbox",
			Name:              "消息中心",
			RouteName:         "WorkspaceInbox",
			RoutePath:         "/workspace/inbox",
			Component:         "/workspace/inbox",
			PageType:          "standalone",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         20,
			ParentMenuName:    "Dashboard",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/dashboard/console",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "jwt",
			InheritPermission: true,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{"isHideTab": false},
		},
		{
			SpaceKey:          systemmodels.DefaultMenuSpaceKey,
			PageKey:           "system.menu_space.manage",
			Name:              "菜单空间",
			RouteName:         "MenuSpaceManage",
			RoutePath:         "/system/menu-space",
			Component:         "/system/menu-space",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "menu_space",
			SortOrder:         25,
			ParentMenuName:    "MenuSpaceManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/menu-space",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "system.menu.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			SpaceKey:          systemmodels.DefaultMenuSpaceKey,
			PageKey:           "system.message.manage",
			Name:              "消息发送",
			RouteName:         "MessageManage",
			RoutePath:         "/system/message",
			Component:         "/system/message",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "message",
			SortOrder:         30,
			ParentMenuName:    "MessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/system/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			SpaceKey:          systemmodels.DefaultMenuSpaceKey,
			PageKey:           "collaboration_workspace.index",
			Name:              "CollaborationWorkspace",
			RouteName:         "CollaborationWorkspaceIndex",
			RoutePath:         "/collaboration-workspace/workspaces",
			Component:         "/collaboration-workspace/workspace",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace",
			SortOrder:         35,
			ParentMenuName:    "CollaborationWorkspaceManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/workspaces",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
		{
			SpaceKey:          systemmodels.DefaultMenuSpaceKey,
			PageKey:           "collaboration_workspace.message.manage",
			Name:              "协作空间消息发送",
			RouteName:         "CollaborationWorkspaceMessageManage",
			RoutePath:         "/collaboration-workspace/message",
			Component:         "/collaboration-workspace/message",
			PageType:          "inner",
			Source:            "manual",
			ModuleKey:         "collaboration_workspace_message",
			SortOrder:         40,
			ParentMenuName:    "CollaborationWorkspaceMessageManage",
			DisplayGroupKey:   "display.system_pages",
			ActiveMenuPath:    "/collaboration-workspace/message",
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        "permission",
			PermissionKey:     "collaboration_workspace.message.manage",
			InheritPermission: false,
			KeepAlive:         true,
			Status:            "normal",
			Meta:              usermodel.MetaJSON{},
		},
	}
	return items
}

func newPermissionKeySeed(resourceCode, actionCode, name, description string) PermissionKeySeed {
	mapping := permissionkey.FromLegacy(resourceCode, actionCode)
	moduleCode := strings.TrimSpace(mapping.ResourceCode)
	if moduleCode == "" {
		moduleCode = strings.TrimSpace(resourceCode)
	}
	displayName := strings.TrimSpace(mapping.Name)
	if displayName == "" {
		displayName = name
	}
	displayDescription := strings.TrimSpace(mapping.Description)
	if displayDescription == "" {
		displayDescription = description
	}
	contextType := strings.TrimSpace(mapping.ContextType)
	if contextType == "" {
		contextType = permissionkey.FromKey(mapping.Key).ContextType
	}
	return PermissionKeySeed{
		Key:              mapping.Key,
		Name:             displayName,
		Description:      displayDescription,
		ContextType:      contextType,
		ModuleCode:       moduleCode,
		ModuleGroupCode:  moduleCode,
		FeatureGroupCode: "system",
		FeatureKind:      "system",
		Status:           "normal",
		IsBuiltin:        true,
	}
}
