package dto

type PermissionActionListRequest struct {
	Current               int    `form:"current"`
	Size                  int    `form:"size"`
	Keyword               string `form:"keyword"`
	Name                  string `form:"name"`
	ResourceCode          string `form:"resource_code"`
	ActionCode            string `form:"action_code"`
	ModuleCode            string `form:"module_code"`
	Category              string `form:"category"`
	Source                string `form:"source"`
	FeatureKind           string `form:"feature_kind"`
	ScopeID               string `form:"scope_id"`
	ScopeCode             string `form:"scope_code"`
	Status                string `form:"status"`
	RequiresTenantContext *bool  `form:"requires_tenant_context"`
}

type PermissionActionCreateRequest struct {
	ResourceCode          string `json:"resource_code" binding:"required,max=100"`
	ActionCode            string `json:"action_code" binding:"required,max=100"`
	ModuleCode            string `json:"module_code" binding:"max=100"`
	Category              string `json:"category" binding:"max=100"`
	FeatureKind           string `json:"feature_kind" binding:"max=20"`
	Name                  string `json:"name" binding:"required,max=150"`
	Description           string `json:"description"`
	ScopeID               string `json:"scope_id" binding:"required"`
	RequiresTenantContext bool   `json:"requires_tenant_context"`
	Status                string `json:"status"`
	SortOrder             int    `json:"sort_order"`
}

type PermissionActionUpdateRequest struct {
	ResourceCode          string `json:"resource_code" binding:"max=100"`
	ActionCode            string `json:"action_code" binding:"max=100"`
	ModuleCode            string `json:"module_code" binding:"max=100"`
	Category              string `json:"category" binding:"max=100"`
	FeatureKind           string `json:"feature_kind" binding:"max=20"`
	Name                  string `json:"name" binding:"max=150"`
	Description           string `json:"description"`
	ScopeID               string `json:"scope_id"`
	RequiresTenantContext *bool  `json:"requires_tenant_context"`
	Status                string `json:"status"`
	SortOrder             int    `json:"sort_order"`
}

type RoleActionPermissionItem struct {
	ActionID string `json:"action_id" binding:"required"`
	Effect   string `json:"effect" binding:"required"`
}

type RoleActionPermissionsRequest struct {
	Actions []RoleActionPermissionItem `json:"actions"`
}

type TenantActionPermissionsRequest struct {
	ActionIDs []string `json:"action_ids"`
}

type UserActionPermissionItem struct {
	ActionID string `json:"action_id" binding:"required"`
	Effect   string `json:"effect" binding:"required"`
}

type UserActionPermissionsRequest struct {
	Actions []UserActionPermissionItem `json:"actions"`
}
