package dto

// TenantListRequest 团队列表请求
type TenantListRequest struct {
	Current int    `form:"current"`
	Size    int    `form:"size"`
	Name    string `form:"name"`
	Status  string `form:"status"`
}

// TenantCreateRequest 创建团队请求
type TenantCreateRequest struct {
	Name         string   `json:"name" binding:"required,max=200"`
	Remark       string   `json:"remark" binding:"max=500"`
	LogoURL      string   `json:"logo_url"`
	Plan         string   `json:"plan"`
	MaxMembers   int      `json:"max_members"`
	Status       string   `json:"status"`
	AdminUserIDs []string `json:"admin_user_ids"`
}

// TenantUpdateRequest 更新团队请求
type TenantUpdateRequest struct {
	Name         string   `json:"name" binding:"max=200"`
	Remark       string   `json:"remark" binding:"max=500"`
	LogoURL      string   `json:"logo_url"`
	Plan         string   `json:"plan"`
	MaxMembers   int      `json:"max_members"`
	Status       string   `json:"status"`
	AdminUserIDs []string `json:"admin_user_ids"`
}

// TenantAddMemberRequest 添加团队成员请求
type TenantAddMemberRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	RoleCode string `json:"role_code"`
}

// TenantUpdateMemberRoleRequest 更新成员角色请求
type TenantUpdateMemberRoleRequest struct {
	RoleCode string `json:"role_code" binding:"required"`
}

// TenantSetMemberRolesRequest 设置成员角色请求
type TenantSetMemberRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}
