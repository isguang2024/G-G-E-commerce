package dto

type PermissionActionListRequest struct {
	Current        int    `form:"current"`
	Size           int    `form:"size"`
	Keyword        string `form:"keyword"`
	PermissionKey  string `form:"permission_key"`
	Name           string `form:"name"`
	ModuleCode     string `form:"module_code"`
	ModuleGroupID  string `form:"module_group_id"`
	FeatureGroupID string `form:"feature_group_id"`
	ContextType    string `form:"context_type"`
	FeatureKind    string `form:"feature_kind"`
	Status         string `form:"status"`
	IsBuiltin      string `form:"is_builtin"`
}

type PermissionActionCreateRequest struct {
	PermissionKey  string `json:"permission_key" binding:"required,max=150"`
	ModuleCode     string `json:"module_code" binding:"max=100"`
	ModuleGroupID  string `json:"module_group_id" binding:"max=36"`
	FeatureGroupID string `json:"feature_group_id" binding:"max=36"`
	ContextType    string `json:"context_type" binding:"max=20"`
	FeatureKind    string `json:"feature_kind" binding:"max=20"`
	Name           string `json:"name" binding:"required,max=150"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	SortOrder      int    `json:"sort_order"`
}

type PermissionActionUpdateRequest struct {
	PermissionKey  string `json:"permission_key" binding:"max=150"`
	ModuleCode     string `json:"module_code" binding:"max=100"`
	ModuleGroupID  string `json:"module_group_id" binding:"max=36"`
	FeatureGroupID string `json:"feature_group_id" binding:"max=36"`
	ContextType    string `json:"context_type" binding:"max=20"`
	FeatureKind    string `json:"feature_kind" binding:"max=20"`
	Name           string `json:"name" binding:"max=150"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	SortOrder      int    `json:"sort_order"`
}

type PermissionGroupListRequest struct {
	Current   int    `form:"current"`
	Size      int    `form:"size"`
	GroupType string `form:"group_type"`
	Keyword   string `form:"keyword"`
	Status    string `form:"status"`
}

type PermissionGroupSaveRequest struct {
	Code        string `json:"code" binding:"required,max=100"`
	Name        string `json:"name" binding:"required,max=150"`
	NameEn      string `json:"name_en" binding:"max=150"`
	Description string `json:"description"`
	GroupType   string `json:"group_type" binding:"required,max=20"`
	Status      string `json:"status" binding:"max=20"`
	SortOrder   int    `json:"sort_order"`
}

type RoleActionPermissionsRequest struct {
	ActionIDs []string `json:"action_ids"`
}

type RoleFeaturePackagesRequest struct {
	PackageIDs []string `json:"package_ids"`
}

type TenantActionPermissionsRequest struct {
	ActionIDs []string `json:"action_ids"`
}

type TenantMenuPermissionsRequest struct {
	MenuIDs []string `json:"menu_ids"`
}

type UserActionPermissionItem struct {
	ActionID string `json:"action_id" binding:"required"`
	Effect   string `json:"effect" binding:"required"`
}

type UserActionPermissionsRequest struct {
	Actions []UserActionPermissionItem `json:"actions"`
}

type FeaturePackageListRequest struct {
	Current     int    `form:"current"`
	Size        int    `form:"size"`
	Keyword     string `form:"keyword"`
	PackageKey  string `form:"package_key"`
	PackageType string `form:"package_type"`
	Name        string `form:"name"`
	ContextType string `form:"context_type"`
	Status      string `form:"status"`
}

type FeaturePackageCreateRequest struct {
	PackageKey  string `json:"package_key" binding:"required,max=100"`
	PackageType string `json:"package_type" binding:"max=20"`
	Name        string `json:"name" binding:"required,max=150"`
	Description string `json:"description"`
	ContextType string `json:"context_type" binding:"max=20"`
	Status      string `json:"status" binding:"max=20"`
	SortOrder   int    `json:"sort_order"`
}

type FeaturePackageUpdateRequest struct {
	PackageKey  string `json:"package_key" binding:"max=100"`
	PackageType string `json:"package_type" binding:"max=20"`
	Name        string `json:"name" binding:"max=150"`
	Description string `json:"description"`
	ContextType string `json:"context_type" binding:"max=20"`
	Status      string `json:"status" binding:"max=20"`
	SortOrder   int    `json:"sort_order"`
}

type FeaturePackageActionSetRequest struct {
	ActionIDs []string `json:"action_ids"`
}

type FeaturePackageMenuSetRequest struct {
	MenuIDs []string `json:"menu_ids"`
}

type TeamFeaturePackageSetRequest struct {
	PackageIDs []string `json:"package_ids"`
}

type FeaturePackageTeamSetRequest struct {
	TeamIDs []string `json:"team_ids"`
}

type FeaturePackageChildSetRequest struct {
	ChildPackageIDs []string `json:"child_package_ids"`
}
