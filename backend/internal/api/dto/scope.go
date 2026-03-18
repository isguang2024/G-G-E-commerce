package dto

// ScopeListRequest 作用域列表请求
type ScopeListRequest struct {
	Current int    `form:"current"`
	Size    int    `form:"size"`
	Name    string `form:"name"`
	Code    string `form:"code"`
}

// ScopeCreateRequest 创建作用域请求
type ScopeCreateRequest struct {
	Code               string `json:"code" binding:"required,max=50"`
	Name               string `json:"name" binding:"required,max=100"`
	Description        string `json:"description"`
	DataPermissionCode string `json:"data_permission_code" binding:"max=50"`
	DataPermissionName string `json:"data_permission_name" binding:"max=100"`
	SortOrder          int    `json:"sort_order"`
}

// ScopeUpdateRequest 更新作用域请求
type ScopeUpdateRequest struct {
	Name               string `json:"name" binding:"max=100"`
	Description        string `json:"description"`
	DataPermissionCode string `json:"data_permission_code" binding:"max=50"`
	DataPermissionName string `json:"data_permission_name" binding:"max=100"`
	SortOrder          int    `json:"sort_order"`
}
