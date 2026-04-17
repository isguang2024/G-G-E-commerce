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
	"role:manage":                            {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "personal"},
	"role:list":                              {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "personal"},
	"role:get":                               {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "personal"},
	"role:create":                            {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "personal"},
	"role:update":                            {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "personal"},
	"role:delete":                            {Key: "system.role.manage", ResourceCode: "role", ActionCode: "manage", Name: "角色管理", Description: "允许查看和维护角色", ContextType: "personal"},
	"role:assign_menu":                       {Key: "system.role.assign", ResourceCode: "role", ActionCode: "assign", Name: "配置角色权限", Description: "允许为角色配置菜单、功能、数据权限", ContextType: "personal"},
	"role:assign_action":                     {Key: "system.role.assign", ResourceCode: "role", ActionCode: "assign", Name: "配置角色权限", Description: "允许为角色配置菜单、功能、数据权限", ContextType: "personal"},
	"role:assign_data":                       {Key: "system.role.assign", ResourceCode: "role", ActionCode: "assign", Name: "配置角色权限", Description: "允许为角色配置菜单、功能、数据权限", ContextType: "personal"},
	"role:assign":                            {Key: "system.role.assign", ResourceCode: "role", ActionCode: "assign", Name: "配置角色权限", Description: "允许为角色配置菜单、功能、数据权限", ContextType: "personal"},
	"user:manage":                            {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:list":                              {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:get":                               {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:read":                              {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:create":                            {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:update":                            {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:delete":                            {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户", ContextType: "personal"},
	"user:assign_role":                       {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户（含角色分配）", ContextType: "personal"},
	"user:assign_menu":                       {Key: "system.user.manage", ResourceCode: "user", ActionCode: "manage", Name: "用户管理", Description: "允许查看和维护用户（含菜单裁剪）", ContextType: "personal"},
	"menu:list":                              {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "personal"},
	"menu:create":                            {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "personal"},
	"menu:update":                            {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "personal"},
	"menu:delete":                            {Key: "system.menu.manage", ResourceCode: "menu", ActionCode: "manage", Name: "菜单管理", Description: "允许查看和维护菜单", ContextType: "personal"},
	"permission_key:manage":                  {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_action:list":                 {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_action:get":                  {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_action:create":               {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_action:update":               {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_action:delete":               {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_key:list":                    {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_key:get":                     {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_key:create":                  {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_key:update":                  {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"permission_key:delete":                  {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"api_endpoint:list":                      {Key: "system.api_registry.view", ResourceCode: "api_endpoint", ActionCode: "view", Name: "查看 API 注册表", Description: "允许查看 API 注册表", ContextType: "personal"},
	"api_endpoint:sync":                      {Key: "system.api_registry.sync", ResourceCode: "api_endpoint", ActionCode: "sync", Name: "允许同步 API 注册表", Description: "允许同步 API 注册表", ContextType: "personal"},
	"page:list":                              {Key: "system.page.manage", ResourceCode: "page", ActionCode: "manage", Name: "页面管理", Description: "允许查看和维护页面注册表", ContextType: "personal"},
	"page:get":                               {Key: "system.page.manage", ResourceCode: "page", ActionCode: "manage", Name: "页面管理", Description: "允许查看和维护页面注册表", ContextType: "personal"},
	"page:create":                            {Key: "system.page.manage", ResourceCode: "page", ActionCode: "manage", Name: "页面管理", Description: "允许查看和维护页面注册表", ContextType: "personal"},
	"page:update":                            {Key: "system.page.manage", ResourceCode: "page", ActionCode: "manage", Name: "页面管理", Description: "允许查看和维护页面注册表", ContextType: "personal"},
	"page:delete":                            {Key: "system.page.manage", ResourceCode: "page", ActionCode: "manage", Name: "页面管理", Description: "允许查看和维护页面注册表", ContextType: "personal"},
	"page:sync":                              {Key: "system.page.sync", ResourceCode: "page", ActionCode: "sync", Name: "页面注册同步", Description: "允许同步页面注册表", ContextType: "personal"},
	"fast_enter:manage":                      {Key: "system.fast_enter.manage", ResourceCode: "fast_enter", ActionCode: "manage", Name: "快捷应用管理", Description: "允许维护顶部快捷应用和快捷链接配置", ContextType: "personal"},
	"message:manage":                         {Key: "message.manage", ResourceCode: "message", ActionCode: "manage", Name: "消息管理", Description: "允许发送站内通知、消息和待办", ContextType: "personal"},
	"collaboration_workspace_message:manage": {Key: "collaboration.message.manage", ResourceCode: "collaboration_message", ActionCode: "manage", Name: "协作消息发送", Description: "允许协作空间管理员向当前协作上下文发送站内通知、消息和待办", ContextType: "collaboration"},
	"feature_package:list":                   {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包", ContextType: "personal"},
	"feature_package:get":                    {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包", ContextType: "personal"},
	"feature_package:create":                 {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包", ContextType: "personal"},
	"feature_package:update":                 {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包", ContextType: "personal"},
	"feature_package:delete":                 {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包", ContextType: "personal"},
	"feature_package:assign_action":          {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包", ContextType: "personal"},
	"feature_package:assign_bundle":          {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "manage", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包组合关系", ContextType: "personal"},
	"feature_package:assign_collaboration_workspace":             {Key: "feature_package.assign_workspace", ResourceCode: "feature_package", ActionCode: "assign_workspace", Name: "工作空间功能包开通", Description: "允许为工作空间开通和配置功能包", ContextType: "personal"},
	"feature_package:assign_menu":                                {Key: "feature_package.manage", ResourceCode: "feature_package", ActionCode: "assign_menu", Name: "个人空间功能包管理", Description: "允许查看和维护个人空间功能包菜单", ContextType: "personal"},
	"system:view_page_catalog":                                   {Key: "system.page_catalog.view", ResourceCode: "system", ActionCode: "view_page_catalog", Name: "查看页面文件映射", Description: "允许查看页面文件映射", ContextType: "personal"},
	"collaboration_workspace:list":                               {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间", ContextType: "personal"},
	"collaboration_workspace:get":                                {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间", ContextType: "personal"},
	"collaboration_workspace:create":                             {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间", ContextType: "personal"},
	"collaboration_workspace:update":                             {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间", ContextType: "personal"},
	"collaboration_workspace:delete":                             {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间", ContextType: "personal"},
	"collaboration_workspace:configure_action_boundary":          {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间及其边界", ContextType: "personal"},
	"collaboration_workspace:configure_menu_boundary":            {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间及其边界", ContextType: "personal"},
	"collaboration_workspace_member_admin:list":                  {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间及其成员", ContextType: "personal"},
	"collaboration_workspace_member_admin:create":                {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间及其成员", ContextType: "personal"},
	"collaboration_workspace_member_admin:delete":                {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间及其成员", ContextType: "personal"},
	"collaboration_workspace_member_admin:update_role":           {Key: "workspace.manage", ResourceCode: "workspace", ActionCode: "manage", Name: "工作空间管理", Description: "允许查看和维护协作工作空间及其成员", ContextType: "personal"},
	"collaboration_workspace_member:create":                      {Key: "collaboration.member.manage", ResourceCode: "collaboration_member", ActionCode: "manage", Name: "当前协作成员管理", Description: "允许在当前协作上下文中查看和维护本协作空间成员", ContextType: "collaboration"},
	"collaboration_workspace_member:delete":                      {Key: "collaboration.member.manage", ResourceCode: "collaboration_member", ActionCode: "manage", Name: "当前协作成员管理", Description: "允许在当前协作上下文中查看和维护本协作空间成员", ContextType: "collaboration"},
	"collaboration_workspace_member:update_role":                 {Key: "collaboration.member.manage", ResourceCode: "collaboration_member", ActionCode: "manage", Name: "当前协作成员管理", Description: "允许在当前协作上下文中查看和维护本协作空间成员", ContextType: "collaboration"},
	"collaboration_workspace_member:assign_role":                 {Key: "collaboration.member.manage", ResourceCode: "collaboration_member", ActionCode: "manage", Name: "当前协作成员管理", Description: "允许在当前协作上下文中为本协作空间成员分配角色", ContextType: "collaboration"},
	"collaboration_workspace_member:assign_action":               {Key: "collaboration.member.manage", ResourceCode: "collaboration_member", ActionCode: "manage", Name: "当前协作成员管理", Description: "允许在当前协作上下文中为本协作空间成员配置功能权限", ContextType: "collaboration"},
	"collaboration_workspace_boundary:configure_action_boundary": {Key: "collaboration.boundary.manage", ResourceCode: "collaboration_boundary", ActionCode: "manage", Name: "当前协作功能边界管理", Description: "允许在当前协作上下文中查看和配置本协作空间的功能边界", ContextType: "collaboration"},
	"collaboration_workspace:configure_menu_boundary_current":    {Key: "collaboration.boundary.manage", ResourceCode: "workspace", ActionCode: "configure_action_boundary", Name: "当前协作功能边界管理", Description: "允许在当前协作上下文中查看和配置本协作空间的菜单边界", ContextType: "collaboration"},
	"collaboration_workspace:configure_action_boundary_current":  {Key: "collaboration.boundary.manage", ResourceCode: "workspace", ActionCode: "configure_action_boundary", Name: "当前协作功能边界管理", Description: "允许在当前协作上下文中查看和配置本协作空间的功能边界", ContextType: "collaboration"},
	"system_permission:manage_action_registry":                   {Key: "system.permission.manage", ResourceCode: "permission_key", ActionCode: "manage", Name: "功能权限管理", Description: "允许查看和维护功能权限", ContextType: "personal"},
	"system_permission:assign_role_action":                       {Key: "system.role.assign", ResourceCode: "role", ActionCode: "assign", Name: "配置角色权限", Description: "允许为角色配置菜单、功能、数据权限", ContextType: "personal"},
	"system.upload.config:manage":                                {Key: "system.upload.config.manage", ResourceCode: "upload", ActionCode: "manage", Name: "上传配置管理", Description: "允许管理存储 Provider / Bucket / UploadKey / Rule 配置", ContextType: "personal"},
	"system.media:view":                                          {Key: "system.media.view", ResourceCode: "upload", ActionCode: "view", Name: "查看媒体文件", Description: "允许查看媒体文件列表", ContextType: "personal"},
	"system.media:upload":                                        {Key: "system.media.upload", ResourceCode: "upload", ActionCode: "upload", Name: "上传媒体文件", Description: "允许使用直传链路上传媒体（prepare/upload/complete 以及发现可用上传配置）", ContextType: "personal"},
	"system.media:manage":                                        {Key: "system.media.manage", ResourceCode: "upload", ActionCode: "manage", Name: "管理媒体文件", Description: "允许删除媒体文件及后续后台治理动作", ContextType: "personal"},
	"system.site_config:view":                                    {Key: "system.site_config.view", ResourceCode: "site_config", ActionCode: "view", Name: "查询参数项（管理端）", Description: "允许查询参数项列表", ContextType: "personal"},
	"system.site_config:manage":                                  {Key: "system.site_config.manage", ResourceCode: "site_config", ActionCode: "manage", Name: "管理参数项", Description: "允许新增、更新、删除参数项和参数集合", ContextType: "personal"},
	"observability.log:read":                                     {Key: "observability.log.read", ResourceCode: "observability", ActionCode: "read", Name: "查看日志", Description: "允许查看审计日志、前端遥测以及注册记录", ContextType: "common"},
	"observability.audit:read":                                   {Key: "observability.log.read", ResourceCode: "observability", ActionCode: "read", Name: "查看日志", Description: "允许查看审计日志、前端遥测以及注册记录", ContextType: "common"},
	"observability.telemetry:read":                               {Key: "observability.log.read", ResourceCode: "observability", ActionCode: "read", Name: "查看日志", Description: "允许查看审计日志、前端遥测以及注册记录", ContextType: "common"},
	"system.register_log:read":                                   {Key: "observability.log.read", ResourceCode: "observability", ActionCode: "read", Name: "查看日志", Description: "允许查看审计日志、前端遥测以及注册记录", ContextType: "common"},
	"observability.policy:manage":                                {Key: "observability.policy.manage", ResourceCode: "observability", ActionCode: "manage", Name: "日志策略管理", Description: "允许查看和维护日志策略", ContextType: "common"},
	"observability.policy:read":                                  {Key: "observability.policy.manage", ResourceCode: "observability", ActionCode: "manage", Name: "日志策略管理", Description: "允许查看和维护日志策略", ContextType: "common"},
	"observability.policy:write":                                 {Key: "observability.policy.manage", ResourceCode: "observability", ActionCode: "manage", Name: "日志策略管理", Description: "允许查看和维护日志策略", ContextType: "common"},
	"system.register_entry:manage":                               {Key: "system.register_entry.manage", ResourceCode: "system", ActionCode: "manage", Name: "注册入口管理", Description: "允许查看和维护注册入口、登录页模板", ContextType: "common"},
	"system.register_entry:read":                                 {Key: "system.register_entry.manage", ResourceCode: "system", ActionCode: "manage", Name: "注册入口管理", Description: "允许查看和维护注册入口、登录页模板", ContextType: "common"},
	"system.register_entry:write":                                {Key: "system.register_entry.manage", ResourceCode: "system", ActionCode: "manage", Name: "注册入口管理", Description: "允许查看和维护注册入口、登录页模板", ContextType: "common"},
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
	switch {
	case target == "collaboration_workspace.manage":
		return "workspace.manage"
	case target == "feature_package.assign_collaboration_workspace":
		return "feature_package.assign_workspace"
	case strings.HasPrefix(target, "collaboration_workspace.member."):
		return "collaboration." + strings.TrimPrefix(target, "collaboration_workspace.")
	case strings.HasPrefix(target, "collaboration_workspace.boundary."):
		return "collaboration." + strings.TrimPrefix(target, "collaboration_workspace.")
	case strings.HasPrefix(target, "collaboration_workspace.message."):
		return "collaboration." + strings.TrimPrefix(target, "collaboration_workspace.")
	default:
		return target
	}
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
			mapping.ContextType = normalizeContextTypeForKey(mapping.Key, mapping.ContextType, mapping.ResourceCode)
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
		ContextType:  normalizeContextTypeForKey(target, deriveContextType(target), resource),
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
		ContextType:  normalizeContextTypeForKey(key, deriveContextType(key), resource),
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
	case strings.HasPrefix(target, "personal."):
		return "personal"
	case strings.HasPrefix(target, "workspace."):
		return "personal"
	case strings.HasPrefix(target, "collaboration."),
		strings.HasPrefix(target, "collaboration_workspace."):
		return "collaboration"
	default:
		return "common"
	}
}

func normalizeContextTypeForKey(permissionKey, contextType, moduleCode string) string {
	key := strings.TrimSpace(permissionKey)
	normalized := strings.TrimSpace(contextType)
	switch {
	case strings.HasPrefix(key, "personal."):
		return "personal"
	case strings.HasPrefix(key, "workspace."):
		return "personal"
	case strings.HasPrefix(key, "collaboration."),
		strings.HasPrefix(key, "collaboration_workspace."):
		return "collaboration"
	case strings.HasPrefix(key, "system."),
		strings.HasPrefix(key, "feature_package."),
		strings.HasPrefix(key, "api_endpoint."),
		strings.HasPrefix(key, "menu."),
		strings.HasPrefix(key, "page."),
		strings.HasPrefix(key, "role."),
		strings.HasPrefix(key, "user."),
		strings.HasPrefix(key, "observability."),
		key == "message.manage",
		strings.HasPrefix(key, "fast_enter."),
		strings.HasPrefix(key, "system_permission."),
		strings.HasPrefix(key, "collaboration_workspace_member_admin."),
		strings.HasPrefix(key, "collaboration_workspace_member."):
		return "common"
	case normalized != "":
		return normalized
	case strings.TrimSpace(moduleCode) != "":
		return deriveContextType(key)
	default:
		return deriveContextType(key)
	}
}
