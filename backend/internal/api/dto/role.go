package dto

// RoleListRequest 角色列表请求
type RoleListRequest struct {
	Current     int    `form:"current"`
	Size        int    `form:"size"`
	RoleName    string `form:"roleName"`
	RoleCode    string `form:"roleCode"`
	Description string `form:"description"`
	// Enabled 角色状态过滤：true 启用，false 禁用，未提供则不过滤
	Enabled *bool `form:"enabled"`
	// 创建时间范围过滤，格式为 YYYY-MM-DD（含起始日，含结束日）
	StartTime  string `form:"startTime"`
	EndTime    string `form:"endTime"`
	Scope      string   `form:"scope"`    // 单一作用域过滤：global | team，空则不过滤
	Scopes     []string `form:"scopes"`   // 多作用域过滤
	GlobalOnly bool   `form:"globalOnly"` // 兼容：true 时按 scope=team 过滤（团队角色及权限页）
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Code        string `json:"code" binding:"required,max=50"`
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	ScopeIDs    []string `json:"scope_ids" binding:"required,min=1"` // 多作用域ID
	Priority    int      `json:"priority"`                            // 优先级
	Status      string   `json:"status"`                              // normal/suspended
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	Code        string `json:"code" binding:"max=50"` // 角色编码（可选，修改时需要检查唯一性）
	Name        string `json:"name" binding:"max=100"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	ScopeIDs    []string `json:"scope_ids"` // 多作用域ID
	Priority    int      `json:"priority"`  // 优先级
	Status      string   `json:"status"`    // normal/suspended
}

// RoleMenusRequest 角色菜单权限请求（保存时传菜单 ID 列表）
type RoleMenusRequest struct {
	MenuIDs []string `json:"menu_ids"`
}

type RoleDataPermissionItem struct {
	ResourceCode string `json:"resource_code" binding:"required,max=100"`
	ScopeCode    string `json:"scope_code" binding:"required,max=30"`
}

type RoleDataPermissionsRequest struct {
	Permissions []RoleDataPermissionItem `json:"permissions"`
}
