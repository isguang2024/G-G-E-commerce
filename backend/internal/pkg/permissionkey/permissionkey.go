package permissionkey

import "strings"

type Mapping struct {
	Key          string
	ResourceCode string
	ActionCode   string
	Name         string
	Description  string
	ContextType  string
}

var mappings = map[string]Mapping{
	"role:list":                                {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "platform"},
	"role:get":                                 {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "platform"},
	"role:create":                              {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "platform"},
	"role:update":                              {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "platform"},
	"role:delete":                              {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "platform"},
	"role:assign_menu":                         {Key: "system.role.assign_menu", ResourceCode: "role", ActionCode: "assign_menu", Name: "配置角色菜单权限", Description: "允许为角色配置菜单权限", ContextType: "platform"},
	"role:assign_action":                       {Key: "system.role.assign_action", ResourceCode: "role", ActionCode: "assign_action", Name: "配置角色功能权限", Description: "允许为角色配置功能权限", ContextType: "platform"},
	"role:assign_data":                         {Key: "system.role.assign_data", ResourceCode: "role", ActionCode: "assign_data", Name: "配置角色数据权限", Description: "允许为角色配置数据权限", ContextType: "platform"},
	"user:list":                                {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "platform"},
	"user:get":                                 {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "platform"},
	"user:create":                              {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "platform"},
	"user:update":                              {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "platform"},
	"user:delete":                              {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "platform"},
	"user:assign_role":                         {Key: "system.user.assign_role", ResourceCode: "user", ActionCode: "assign_role", Name: "分配用户角色", Description: "允许为用户分配角色", ContextType: "platform"},
	"user:assign_action":                       {Key: "system.user.assign_action", ResourceCode: "user", ActionCode: "assign_action", Name: "配置用户功能权限", Description: "允许为用户配置功能权限", ContextType: "platform"},
	"user:assign_menu":                         {Key: "system.user.manage", ResourceCode: "user", ActionCode: "assign_menu", Name: "配置用户菜单裁剪", Description: "允许为用户配置菜单裁剪", ContextType: "platform"},
	"menu:list":                                {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "platform"},
	"menu:create":                              {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "platform"},
	"menu:update":                              {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "platform"},
	"menu:delete":                              {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "platform"},
	"menu_backup:create":                       {Key: "system.menu.backup", ResourceCode: "menu_backup", ActionCode: "manage", Name: "菜单备份管理", Description: "允许创建、查看、恢复和删除菜单备份", ContextType: "platform"},
	"menu_backup:list":                         {Key: "system.menu.backup", ResourceCode: "menu_backup", ActionCode: "manage", Name: "菜单备份管理", Description: "允许创建、查看、恢复和删除菜单备份", ContextType: "platform"},
	"menu_backup:delete":                       {Key: "system.menu.backup", ResourceCode: "menu_backup", ActionCode: "manage", Name: "菜单备份管理", Description: "允许创建、查看、恢复和删除菜单备份", ContextType: "platform"},
	"menu_backup:restore":                      {Key: "system.menu.backup", ResourceCode: "menu_backup", ActionCode: "manage", Name: "菜单备份管理", Description: "允许创建、查看、恢复和删除菜单备份", ContextType: "platform"},
	"permission_action:list":                   {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_action:get":                    {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_action:create":                 {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_action:update":                 {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_action:delete":                 {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_key:list":                      {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_key:get":                       {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_key:create":                    {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_key:update":                    {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"permission_key:delete":                    {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"api_endpoint:list":                        {Key: "system.api_registry.view", ResourceCode: "api_endpoint", ActionCode: "view", Name: "查看 API 注册表", Description: "允许查看 API 注册表", ContextType: "platform"},
	"api_endpoint:sync":                        {Key: "system.api_registry.sync", ResourceCode: "api_endpoint", ActionCode: "sync", Name: "允许同步 API 注册表", Description: "允许同步 API 注册表", ContextType: "platform"},
	"feature_package:list":                     {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包", ContextType: "platform"},
	"feature_package:get":                      {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包", ContextType: "platform"},
	"feature_package:create":                   {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包", ContextType: "platform"},
	"feature_package:update":                   {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包", ContextType: "platform"},
	"feature_package:delete":                   {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包", ContextType: "platform"},
	"feature_package:assign_action":            {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包", ContextType: "platform"},
	"feature_package:assign_bundle":            {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "平台功能包管理", Description: "允许查看和维护平台功能包组合关系", ContextType: "platform"},
	"feature_package:assign_team":              {Key: "platform.package.assign", ResourceCode: "feature_package", ActionCode: "assign_team", Name: "平台团队功能包开通", Description: "允许为团队开通和配置功能包", ContextType: "platform"},
	"feature_package:assign_menu":              {Key: "platform.package.manage", ResourceCode: "feature_package", ActionCode: "assign_menu", Name: "平台功能包管理", Description: "允许查看和维护平台功能包菜单", ContextType: "platform"},
	"system:view_page_catalog":                 {Key: "system.page_catalog.view", ResourceCode: "system", ActionCode: "view_page_catalog", Name: "查看页面文件映射", Description: "允许查看页面文件映射", ContextType: "platform"},
	"tenant:list":                              {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队", ContextType: "platform"},
	"tenant:get":                               {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队", ContextType: "platform"},
	"tenant:create":                            {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队", ContextType: "platform"},
	"tenant:update":                            {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队", ContextType: "platform"},
	"tenant:delete":                            {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队", ContextType: "platform"},
	"tenant:configure_action_boundary":         {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队及其边界", ContextType: "platform"},
	"tenant:configure_menu_boundary":           {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队及其边界", ContextType: "platform"},
	"tenant_member_admin:list":                 {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队及其成员", ContextType: "platform"},
	"tenant_member_admin:create":               {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队及其成员", ContextType: "platform"},
	"tenant_member_admin:delete":               {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队及其成员", ContextType: "platform"},
	"tenant_member_admin:update_role":          {Key: "tenant.manage", ResourceCode: "tenant", ActionCode: "manage", Name: "团队管理", Description: "允许查看和维护团队及其成员", ContextType: "platform"},
	"team_member:create":                       {Key: "team.member.manage", ResourceCode: "team_member", ActionCode: "manage", Name: "当前团队成员管理", Description: "允许在当前团队上下文中查看和维护本团队成员", ContextType: "team"},
	"team_member:delete":                       {Key: "team.member.manage", ResourceCode: "team_member", ActionCode: "manage", Name: "当前团队成员管理", Description: "允许在当前团队上下文中查看和维护本团队成员", ContextType: "team"},
	"team_member:update_role":                  {Key: "team.member.manage", ResourceCode: "team_member", ActionCode: "manage", Name: "当前团队成员管理", Description: "允许在当前团队上下文中查看和维护本团队成员", ContextType: "team"},
	"team_member:assign_role":                  {Key: "team.member.manage", ResourceCode: "team_member", ActionCode: "manage", Name: "当前团队成员管理", Description: "允许在当前团队上下文中为本团队成员分配角色", ContextType: "team"},
	"team_member:assign_action":                {Key: "team.member.manage", ResourceCode: "team_member", ActionCode: "manage", Name: "当前团队成员管理", Description: "允许在当前团队上下文中为本团队成员配置功能权限", ContextType: "team"},
	"team:configure_menu_boundary":             {Key: "team.boundary.manage", ResourceCode: "team", ActionCode: "configure_action_boundary", Name: "当前团队功能边界管理", Description: "允许在当前团队上下文中查看和配置本团队的菜单边界", ContextType: "team"},
	"team:configure_action_boundary":           {Key: "team.boundary.manage", ResourceCode: "team", ActionCode: "configure_action_boundary", Name: "当前团队功能边界管理", Description: "允许在当前团队上下文中查看和配置本团队的功能边界", ContextType: "team"},
	"system_permission:manage_action_registry": {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "platform"},
	"system_permission:assign_role_action":     {Key: "system.role.assign_action", ResourceCode: "role", ActionCode: "assign_action", Name: "配置角色功能权限", Description: "允许为角色配置功能权限", ContextType: "platform"},
}

func Normalize(key string) string {
	target := strings.TrimSpace(key)
	if target == "" {
		return ""
	}
	if strings.Contains(target, ":") {
		parts := strings.SplitN(target, ":", 2)
		if len(parts) == 2 {
			return FromLegacy(parts[0], parts[1]).Key
		}
		return strings.ReplaceAll(target, ":", ".")
	}
	return target
}

func ListMappings() []Mapping {
	unique := make(map[string]Mapping, len(mappings))
	for _, mapping := range mappings {
		if strings.TrimSpace(mapping.Key) == "" {
			continue
		}
		unique[mapping.Key] = mapping
	}
	result := make([]Mapping, 0, len(unique))
	for _, mapping := range unique {
		result = append(result, mapping)
	}
	return result
}

func FromKey(key string) Mapping {
	target := Normalize(key)
	if target == "" {
		return Mapping{}
	}
	for _, mapping := range mappings {
		if strings.TrimSpace(mapping.Key) == target {
			return mapping
		}
	}
	parts := strings.Split(target, ".")
	if len(parts) == 0 {
		return Mapping{Key: target}
	}
	action := parts[len(parts)-1]
	resource := strings.Join(parts[:len(parts)-1], "_")
	if resource == "" {
		resource = target
	}
	if action == "" {
		action = "manage"
	}
	return Mapping{
		Key:          target,
		ResourceCode: strings.ReplaceAll(resource, ".", "_"),
		ActionCode:   action,
		ContextType:  deriveContextType(target),
	}
}

func FromLegacy(resourceCode, actionCode string) Mapping {
	resource := strings.TrimSpace(resourceCode)
	action := strings.TrimSpace(actionCode)
	legacy := resource + ":" + action
	if mapping, ok := mappings[legacy]; ok {
		return mapping
	}
	key := normalizeLegacyKey(resource, action)
	return Mapping{
		Key:          key,
		ResourceCode: resource,
		ActionCode:   action,
		Name:         key,
		ContextType:  deriveContextType(key),
	}
}

func normalizeLegacyKey(resourceCode, actionCode string) string {
	resource := strings.Trim(strings.ReplaceAll(strings.TrimSpace(resourceCode), ":", "."), ".")
	action := strings.Trim(strings.ReplaceAll(strings.TrimSpace(actionCode), ":", "."), ".")
	switch {
	case resource == "" && action == "":
		return ""
	case resource == "":
		return action
	case action == "":
		return resource
	default:
		return resource + "." + action
	}
}

func deriveContextType(permissionKey string) string {
	target := strings.TrimSpace(permissionKey)
	switch {
	case strings.HasPrefix(target, "system."),
		strings.HasPrefix(target, "tenant."),
		strings.HasPrefix(target, "platform."):
		return "platform"
	default:
		return "team"
	}
}
