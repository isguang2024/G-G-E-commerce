package dto

// UserListRequest 用户列表请求（分页与筛选）
type UserListRequest struct {
	Current        int    `form:"current"` // 页码，与前端一致
	Size           int    `form:"size"`    // 每页条数
	ID             string `form:"id"`      // 用户ID
	UserName       string `form:"userName"`
	UserPhone      string `form:"userPhone"` // 手机号
	UserEmail      string `form:"userEmail"` // 邮箱
	Status         string `form:"status"`
	RoleID         string `form:"roleId"`         // 角色ID（UUID字符串，查询拥有该角色的用户）
	RegisterSource string `form:"registerSource"` // 注册来源
	InvitedBy      string `form:"invitedBy"`      // 邀请人ID
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username     string   `json:"username" binding:"required,min=2,max=100"`
	Password     string   `json:"password" binding:"required,min=6"`
	Email        string   `json:"email"`
	Nickname     string   `json:"nickname"`
	Phone        string   `json:"phone"`
	SystemRemark string   `json:"systemRemark"`
	Status       string   `json:"status"`
	RoleIDs      []string `json:"roleIds"` // 角色 ID 列表（UUID 字符串）
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	Email        string   `json:"email"`
	Nickname     string   `json:"nickname"`
	Phone        string   `json:"phone"`
	SystemRemark string   `json:"systemRemark"`
	Status       string   `json:"status"`
	RoleIDs      []string `json:"roleIds"`
}

// UserAssignRolesRequest 分配角色请求
type UserAssignRolesRequest struct {
	RoleIDs []string `json:"roleIds" binding:"required"`
}

type UserPermissionDiagnosisRequest struct {
	CollaborationWorkspaceID string `form:"collaboration_workspace_id"`
	PermissionKey            string `form:"permission_key"`
}

type UserPermissionRefreshRequest struct {
	CollaborationWorkspaceID string `json:"collaboration_workspace_id"`
}
