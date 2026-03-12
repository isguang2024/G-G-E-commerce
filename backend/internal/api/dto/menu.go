package dto

// MenuCreateRequest 创建菜单
type MenuCreateRequest struct {
	ParentID  *string                `json:"parent_id"`
	Path      string                 `json:"path" binding:"required,max=255"`
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
	ParentID  *string                `json:"parent_id"` // 空字符串或 null 表示移至顶级
	Path      string                 `json:"path" binding:"max=255"`
	Name      string                 `json:"name" binding:"max=100"`
	Component string                 `json:"component"`
	Title     string                 `json:"title" binding:"max=100"`
	Icon      string                 `json:"icon"`
	SortOrder int                    `json:"sort_order"`
	Meta      map[string]interface{} `json:"meta"`
	Hidden    bool                   `json:"hidden"`
}
