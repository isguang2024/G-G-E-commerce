package dto

// MenuCreateRequest 创建菜单
type MenuCreateRequest struct {
	AppKey    string                 `json:"app_key"`
	ParentID  *string                `json:"parent_id"`
	SpaceKey  string                 `json:"space_key"`
	Kind      string                 `json:"kind"`
	Path      string                 `json:"path" binding:"max=255"`
	Name      string                 `json:"name" binding:"required,max=100"`
	Component string                 `json:"component"`
	Title     string                 `json:"title" binding:"required,max=100"`
	Icon      string                 `json:"icon"`
	SortOrder int                    `json:"sort_order"`
	Meta      map[string]interface{} `json:"meta"`
	Hidden    bool                   `json:"hidden"`
}

// MenuUpdateRequest 更新菜单
type MenuUpdateRequest struct {
	AppKey    string                 `json:"app_key"`
	ParentID  *string                `json:"parent_id"` // 空字符串表示移至顶级；省略字段表示保持不变
	SpaceKey  string                 `json:"space_key"`
	Kind      string                 `json:"kind"`
	Path      string                 `json:"path" binding:"max=255"`
	Name      string                 `json:"name" binding:"max=100"`
	Component string                 `json:"component"`
	Title     string                 `json:"title" binding:"max=100"`
	Icon      string                 `json:"icon"`
	SortOrder int                    `json:"sort_order"`
	Meta      map[string]interface{} `json:"meta"`
	Hidden    bool                   `json:"hidden"`
}
