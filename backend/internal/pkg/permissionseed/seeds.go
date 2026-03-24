package permissionseed

import (
	"crypto/sha1"
	"strings"

	"github.com/google/uuid"

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
	ContextType    string
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
	Name       string
	ParentName string
	Path       string
	Component  string
	Title      string
	Icon       string
	SortOrder  int
	Meta       usermodel.MetaJSON
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
		{Code: "permission_action", Name: "功能键", NameEn: "Permission Key", SortOrder: 50, Status: "normal"},
		{Code: "feature_package", Name: "功能包", NameEn: "Feature Package", SortOrder: 60, Status: "normal"},
		{Code: "api_endpoint", Name: "API 管理", NameEn: "API Management", SortOrder: 70, Status: "normal"},
		{Code: "menu", Name: "菜单", NameEn: "Menu", SortOrder: 80, Status: "normal"},
		{Code: "menu_backup", Name: "菜单备份", NameEn: "Menu Backup", SortOrder: 90, Status: "normal"},
		{Code: "tenant", Name: "团队", NameEn: "Tenant", SortOrder: 100, Status: "normal"},
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
		newPermissionKeySeed("permission_action", "list", "查看功能权限列表", "允许查看功能权限列表"),
		newPermissionKeySeed("permission_action", "get", "查看功能权限详情", "允许查看功能权限详情"),
		newPermissionKeySeed("permission_action", "create", "创建功能权限", "允许创建功能权限"),
		newPermissionKeySeed("permission_action", "update", "更新功能权限", "允许更新功能权限"),
		newPermissionKeySeed("permission_action", "delete", "删除功能权限", "允许删除功能权限"),
		newPermissionKeySeed("user", "list", "查看用户列表", "允许查看用户列表"),
		newPermissionKeySeed("user", "get", "查看用户详情", "允许查看用户详情"),
		newPermissionKeySeed("user", "create", "创建用户", "允许创建用户"),
		newPermissionKeySeed("user", "update", "更新用户", "允许更新用户"),
		newPermissionKeySeed("user", "delete", "删除用户", "允许删除用户"),
		newPermissionKeySeed("user", "assign_role", "分配用户角色", "允许为用户分配角色"),
		newPermissionKeySeed("user", "assign_action", "配置用户功能权限", "允许为用户配置平台级功能权限"),
		newPermissionKeySeed("menu", "list", "查看菜单管理树", "允许查看全部菜单管理树"),
		newPermissionKeySeed("menu", "create", "创建菜单", "允许创建菜单"),
		newPermissionKeySeed("menu", "update", "更新菜单", "允许更新菜单"),
		newPermissionKeySeed("menu", "delete", "删除菜单", "允许删除菜单"),
		newPermissionKeySeed("menu_backup", "create", "创建菜单备份", "允许创建菜单备份"),
		newPermissionKeySeed("menu_backup", "list", "查看菜单备份列表", "允许查看菜单备份列表"),
		newPermissionKeySeed("menu_backup", "delete", "删除菜单备份", "允许删除菜单备份"),
		newPermissionKeySeed("menu_backup", "restore", "恢复菜单备份", "允许恢复菜单备份"),
		newPermissionKeySeed("system", "view_page_catalog", "查看页面文件映射", "允许查看页面文件映射"),
		newPermissionKeySeed("tenant", "list", "查看团队列表", "允许查看团队列表"),
		newPermissionKeySeed("tenant", "get", "查看团队详情", "允许查看团队详情"),
		newPermissionKeySeed("tenant", "create", "创建团队", "允许创建团队"),
		newPermissionKeySeed("tenant", "update", "更新团队", "允许更新团队"),
		newPermissionKeySeed("tenant", "delete", "删除团队", "允许删除团队"),
		newPermissionKeySeed("tenant", "configure_action_boundary", "配置团队功能权限边界", "允许配置团队功能权限边界"),
		newPermissionKeySeed("tenant_member_admin", "list", "查看团队成员列表", "允许在系统管理中查看团队成员列表"),
		newPermissionKeySeed("tenant_member_admin", "create", "添加团队成员", "允许在系统管理中添加团队成员"),
		newPermissionKeySeed("tenant_member_admin", "delete", "移除团队成员", "允许在系统管理中移除团队成员"),
		newPermissionKeySeed("tenant_member_admin", "update_role", "更新团队成员身份", "允许在系统管理中更新团队成员身份"),
		newPermissionKeySeed("team_member", "create", "添加当前团队成员", "允许在当前团队中添加成员"),
		newPermissionKeySeed("team_member", "delete", "移除当前团队成员", "允许在当前团队中移除成员"),
		newPermissionKeySeed("team_member", "update_role", "更新当前团队成员身份", "允许在当前团队中更新成员身份"),
		newPermissionKeySeed("team_member", "assign_role", "配置当前团队成员角色", "允许在当前团队中配置成员角色"),
		newPermissionKeySeed("team_member", "assign_action", "配置当前团队成员功能权限", "允许在当前团队中配置成员功能权限"),
		newPermissionKeySeed("team", "configure_action_boundary", "查看和配置当前团队功能权限边界", "允许查看和配置当前团队功能权限边界"),
		newPermissionKeySeed("api_endpoint", "list", "查看 API 注册表", "允许查看 API 注册表"),
		newPermissionKeySeed("api_endpoint", "sync", "同步 API 注册表", "允许同步 API 注册表"),
		newPermissionKeySeed("feature_package", "list", "查看功能包列表", "允许查看功能包列表"),
		newPermissionKeySeed("feature_package", "get", "查看功能包详情", "允许查看功能包详情"),
		newPermissionKeySeed("feature_package", "create", "创建功能包", "允许创建功能包"),
		newPermissionKeySeed("feature_package", "update", "更新功能包", "允许更新功能包"),
		newPermissionKeySeed("feature_package", "delete", "删除功能包", "允许删除功能包"),
		newPermissionKeySeed("feature_package", "assign_action", "配置功能包权限", "允许配置功能包包含的功能权限"),
		newPermissionKeySeed("feature_package", "assign_team", "配置团队功能包", "允许给团队开通功能包"),
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
	items := []PermissionGroupSeed{
		{GroupType: "feature", Code: "system", Name: "系统功能", NameEn: "System Feature", Description: "系统初始化和管理能力", Status: "normal", SortOrder: 1, IsBuiltin: true},
		{GroupType: "feature", Code: "business", Name: "业务功能", NameEn: "Business Feature", Description: "业务扩展能力", Status: "normal", SortOrder: 2, IsBuiltin: true},
	}
	moduleSeen := map[string]struct{}{}
	for _, item := range DefaultPermissionKeys() {
		code := strings.TrimSpace(item.ModuleGroupCode)
		if code == "" {
			continue
		}
		if _, ok := moduleSeen[code]; ok {
			continue
		}
		moduleSeen[code] = struct{}{}
		items = append(items, PermissionGroupSeed{
			GroupType:   "module",
			Code:        code,
			Name:        code,
			NameEn:      strings.ToUpper(strings.ReplaceAll(code, "_", " ")),
			Description: "默认模块分组",
			Status:      "normal",
			SortOrder:   len(items) + 1,
			IsBuiltin:   true,
		})
	}
	for i := range items {
		items[i].ID = StableID("permission-group", items[i].GroupType+":"+items[i].Code)
	}
	return items
}

func DefaultFeaturePackages() []FeaturePackageSeed {
	items := []FeaturePackageSeed{
		{
			PackageKey:     "platform.system_admin",
			PackageType:    "base",
			Name:           "平台系统管理包",
			Description:    "包含平台系统管理核心能力",
			ContextType:    "platform",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      1,
			MenuNames:      []string{"System", "Role", "User", "ActionPermission", "TeamRolesAndPermissions"},
			PermissionKeys: []string{"system.user.manage", "system.role.manage", "system.permission.manage"},
		},
		{
			PackageKey:     "platform.menu_admin",
			PackageType:    "base",
			Name:           "平台菜单管理包",
			Description:    "包含菜单管理与菜单备份能力",
			ContextType:    "platform",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      2,
			MenuNames:      []string{"System", "Menus"},
			PermissionKeys: []string{"system.menu.manage", "system.menu.backup"},
		},
		{
			PackageKey:     "platform.api_admin",
			PackageType:    "base",
			Name:           "平台接口管理包",
			Description:    "包含 API 注册表查看与同步能力",
			ContextType:    "platform",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      3,
			MenuNames:      []string{"System", "ApiEndpoint", "FeaturePackage"},
			PermissionKeys: []string{"system.api_registry.view", "system.api_registry.sync", "platform.package.manage", "platform.package.assign"},
		},
		{
			PackageKey:     "team.member_admin",
			PackageType:    "base",
			Name:           "团队成员管理包",
			Description:    "包含团队成员、角色和功能权限配置能力",
			ContextType:    "team",
			IsBuiltin:      true,
			Status:         "normal",
			SortOrder:      10,
			MenuNames:      []string{"TeamRoot", "TeamMembers", "TeamRolesAndPermissions"},
			PermissionKeys: []string{"team.member.manage", "team.boundary.manage"},
		},
		{
			PackageKey:  "platform.admin_bundle",
			PackageType: "bundle",
			Name:        "平台管理员组合包",
			Description: "统一聚合平台系统、菜单与接口管理基础包",
			ContextType: "platform",
			IsBuiltin:   true,
			Status:      "normal",
			SortOrder:   4,
		},
	}
	for i := range items {
		items[i].ID = StableID("feature-package", items[i].PackageKey)
	}
	return items
}

func DefaultFeaturePackageBundles() []FeaturePackageBundleSeed {
	return []FeaturePackageBundleSeed{
		{ParentPackageKey: "platform.admin_bundle", ChildPackageKey: "platform.system_admin"},
		{ParentPackageKey: "platform.admin_bundle", ChildPackageKey: "platform.menu_admin"},
		{ParentPackageKey: "platform.admin_bundle", ChildPackageKey: "platform.api_admin"},
	}
}

func DefaultMenus() []MenuSeed {
	metaSuperAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}}
	metaSuperAdminAndAdmin := usermodel.MetaJSON{"roles": []interface{}{"R_SUPER", "R_ADMIN"}}
	metaTeamAccessOnly := usermodel.MetaJSON{"keepAlive": true}
	return []MenuSeed{
		{Name: "Dashboard", Path: "/dashboard", Component: "/index/index", Title: "menus.dashboard.title", Icon: "ri:pie-chart-line", SortOrder: 1, Meta: metaSuperAdminAndAdmin},
		{Name: "System", Path: "/system", Component: "/index/index", Title: "menus.system.title", Icon: "ri:user-3-line", SortOrder: 2, Meta: metaSuperAdminAndAdmin},
		{Name: "Result", Path: "/result", Component: "/index/index", Title: "menus.result.title", Icon: "ri:checkbox-circle-line", SortOrder: 3},
		{Name: "Exception", Path: "/exception", Component: "/index/index", Title: "menus.exception.title", Icon: "ri:error-warning-line", SortOrder: 4},
		{Name: "TeamRoot", Path: "/team", Component: "/index/index", Title: "menus.system.team", Icon: "ri:team-line", SortOrder: 5, Meta: metaTeamAccessOnly},
		{Name: "Console", ParentName: "Dashboard", Path: "console", Component: "/dashboard/console", Title: "menus.dashboard.console", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": false, "fixedTab": true}},
		{Name: "UserCenter", ParentName: "Dashboard", Path: "user-center", Component: "/system/user-center", Title: "menus.system.userCenter", SortOrder: 2, Meta: usermodel.MetaJSON{"isHide": true, "keepAlive": true, "isHideTab": true}},
		{Name: "Role", ParentName: "System", Path: "role", Component: "/system/role", Title: "menus.system.role", SortOrder: 1, Meta: metaSuperAdmin},
		{Name: "User", ParentName: "System", Path: "user", Component: "/system/user", Title: "menus.system.user", SortOrder: 2, Meta: metaSuperAdminAndAdmin},
		{Name: "TeamRolesAndPermissions", ParentName: "TeamRoot", Path: "roles", Component: "/system/team-roles-permissions", Title: "menus.system.teamRolesAndPermissions", SortOrder: 3, Meta: metaTeamAccessOnly},
		{Name: "Menus", ParentName: "System", Path: "menu", Component: "/system/menu", Title: "menus.system.menu", SortOrder: 4, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "ActionPermission", ParentName: "System", Path: "action-permission", Component: "/system/action-permission", Title: "功能权限", SortOrder: 5, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "ApiEndpoint", ParentName: "System", Path: "api-endpoint", Component: "/system/api-endpoint", Title: "API管理", SortOrder: 6, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "FeaturePackage", ParentName: "System", Path: "feature-package", Component: "/system/feature-package", Title: "功能包管理", SortOrder: 7, Meta: usermodel.MetaJSON{"roles": []interface{}{"R_SUPER"}, "keepAlive": true}},
		{Name: "TeamMembers", ParentName: "TeamRoot", Path: "members", Component: "/team/team-members", Title: "menus.system.teamMembers", SortOrder: 2, Meta: metaTeamAccessOnly},
		{Name: "ResultSuccess", ParentName: "Result", Path: "success", Component: "/result/success", Title: "menus.result.success", Icon: "ri:checkbox-circle-line", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": true}},
		{Name: "ResultFail", ParentName: "Result", Path: "fail", Component: "/result/fail", Title: "menus.result.fail", Icon: "ri:close-circle-line", SortOrder: 2, Meta: usermodel.MetaJSON{"keepAlive": true}},
		{Name: "Exception403", ParentName: "Exception", Path: "403", Component: "/exception/403", Title: "menus.exception.forbidden", SortOrder: 1, Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}},
		{Name: "Exception404", ParentName: "Exception", Path: "404", Component: "/exception/404", Title: "menus.exception.notFound", SortOrder: 2, Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}},
		{Name: "Exception500", ParentName: "Exception", Path: "500", Component: "/exception/500", Title: "menus.exception.serverError", SortOrder: 3, Meta: usermodel.MetaJSON{"keepAlive": true, "isHideTab": true, "isFullPage": true}},
	}
}

func DeprecatedDefaultMenuNames() []string {
	return []string{"PageAssociation", "TeamManagementRedirect", "Scope"}
}

func DefaultRolePackageBindings() []RolePackageBindingSeed {
	items := []RolePackageBindingSeed{
		{RoleCode: "admin", PackageKey: "platform.admin_bundle"},
		{RoleCode: "team_admin", PackageKey: "team.member_admin"},
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
