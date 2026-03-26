package dto

// RoleListRequest 角色列表请求
type RoleListRequest struct {
	Current     int    `form:"current"`
	Size        int    `form:"size"`
	RoleName    string `form:"roleName"`
	RoleCode    string `form:"roleCode"`
	Description string `form:"description"`
	Enabled     *bool  `form:"enabled"`
	StartTime   string `form:"startTime"`
	EndTime     string `form:"endTime"`
}

// RoleCreateRequest 创建角色请求
type RoleCreateRequest struct {
	Code         string         `json:"code" binding:"required,max=50"`
	Name         string         `json:"name" binding:"required,max=100"`
	Description  string         `json:"description"`
	SortOrder    int            `json:"sort_order"`
	Priority     int            `json:"priority"`
	CustomParams map[string]any `json:"custom_params"`
	Status       string         `json:"status"`
}

// RoleUpdateRequest 更新角色请求
type RoleUpdateRequest struct {
	Code         string         `json:"code" binding:"max=50"`
	Name         string         `json:"name" binding:"max=100"`
	Description  string         `json:"description"`
	SortOrder    int            `json:"sort_order"`
	Priority     int            `json:"priority"`
	CustomParams map[string]any `json:"custom_params"`
	Status       string         `json:"status"`
}

// RoleMenusRequest 角色菜单权限请求
type RoleMenusRequest struct {
	MenuIDs []string `json:"menu_ids"`
}

// RoleDataPermissionItem 角色数据权限项
type RoleDataPermissionItem struct {
	ResourceCode string `json:"resource_code" binding:"required,max=100"`
	DataScope    string `json:"data_scope" binding:"required,max=30"`
}

// RoleDataPermissionsRequest 角色数据权限请求
type RoleDataPermissionsRequest struct {
	Permissions []RoleDataPermissionItem `json:"permissions"`
}
