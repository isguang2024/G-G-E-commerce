package dto

// CollaborationWorkspaceListRequest 协作空间列表请求
type CollaborationWorkspaceListRequest struct {
	Current int    `form:"current"`
	Size    int    `form:"size"`
	Name    string `form:"name"`
	Status  string `form:"status"`
}

// CollaborationWorkspaceCreateRequest 创建协作空间请求
type CollaborationWorkspaceCreateRequest struct {
	Name         string   `json:"name" binding:"required,max=200"`
	Remark       string   `json:"remark" binding:"max=500"`
	LogoURL      string   `json:"logo_url"`
	Plan         string   `json:"plan"`
	MaxMembers   int      `json:"max_members"`
	Status       string   `json:"status"`
	AdminUserIDs []string `json:"admin_user_ids"`
}

// CollaborationWorkspaceUpdateRequest 更新协作空间请求
type CollaborationWorkspaceUpdateRequest struct {
	Name         string   `json:"name" binding:"max=200"`
	Remark       string   `json:"remark" binding:"max=500"`
	LogoURL      string   `json:"logo_url"`
	Plan         string   `json:"plan"`
	MaxMembers   int      `json:"max_members"`
	Status       string   `json:"status"`
	AdminUserIDs []string `json:"admin_user_ids"`
}

// CollaborationWorkspaceAddMemberRequest 添加协作空间成员请求
type CollaborationWorkspaceAddMemberRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	RoleCode string `json:"role_code"`
}

// CollaborationWorkspaceUpdateMemberRoleRequest 更新成员角色请求
type CollaborationWorkspaceUpdateMemberRoleRequest struct {
	RoleCode string `json:"role_code" binding:"required"`
}

// CollaborationWorkspaceSetMemberRolesRequest 设置成员角色请求
type CollaborationWorkspaceSetMemberRolesRequest struct {
	RoleIDs []string `json:"role_ids" binding:"required"`
}
