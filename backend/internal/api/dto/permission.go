package dto

type PermissionKeyListRequest struct {
	Current          int    `form:"current"`
	Size             int    `form:"size"`
	Keyword          string `form:"keyword"`
	PermissionKey    string `form:"permission_key"`
	Name             string `form:"name"`
	ModuleCode       string `form:"module_code"`
	ModuleGroupID    string `form:"module_group_id"`
	FeatureGroupID   string `form:"feature_group_id"`
	ContextType      string `form:"context_type"`
	FeatureKind      string `form:"feature_kind"`
	Status           string `form:"status"`
	IsBuiltin        string `form:"is_builtin"`
	UsagePattern     string `form:"usage_pattern"`
	DuplicatePattern string `form:"duplicate_pattern"`
}

type PermissionKeyCreateRequest struct {
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

type PermissionKeyUpdateRequest struct {
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

type PermissionKeyEndpointBindRequest struct {
	EndpointCode string `json:"endpoint_code" binding:"required,max=36"`
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

type RoleKeyPermissionsRequest struct {
	AppKey string   `json:"app_key" binding:"max=100"`
	KeyIDs []string `json:"action_ids"`
}

type RoleFeaturePackagesRequest struct {
	AppKey     string   `json:"app_key" binding:"max=100"`
	PackageIDs []string `json:"package_ids"`
}

type CollaborationWorkspaceKeyPermissionsRequest struct {
	AppKey string   `json:"app_key" binding:"max=100"`
	KeyIDs []string `json:"action_ids"`
}

type CollaborationWorkspaceMenuPermissionsRequest struct {
	AppKey  string   `json:"app_key" binding:"max=100"`
	MenuIDs []string `json:"menu_ids"`
}

type UserKeyPermissionItem struct {
	KeyID  string `json:"action_id" binding:"required"`
	Effect string `json:"effect" binding:"required"`
}

type UserKeyPermissionsRequest struct {
	AppKey  string                  `json:"app_key" binding:"max=100"`
	Actions []UserKeyPermissionItem `json:"actions"`
}

type FeaturePackageListRequest struct {
	Current     int    `form:"current"`
	Size        int    `form:"size"`
	AppKey      string `form:"app_key"`
	Keyword     string `form:"keyword"`
	PackageKey  string `form:"package_key"`
	PackageType string `form:"package_type"`
	Name        string `form:"name"`
	WorkspaceScope string `form:"workspace_scope"`
	Status      string `form:"status"`
}

type FeaturePackageCreateRequest struct {
	AppKey      string `json:"app_key" binding:"max=100"`
	AppKeys     []string `json:"app_keys"`
	PackageKey  string `json:"package_key" binding:"required,max=100"`
	PackageType string `json:"package_type" binding:"max=20"`
	Name        string `json:"name" binding:"required,max=150"`
	Description string `json:"description"`
	WorkspaceScope string `json:"workspace_scope" binding:"max=20"`
	Status      string `json:"status" binding:"max=20"`
	SortOrder   int    `json:"sort_order"`
}

type FeaturePackageUpdateRequest struct {
	AppKey      string `json:"app_key" binding:"max=100"`
	AppKeys     []string `json:"app_keys"`
	PackageKey  string `json:"package_key" binding:"max=100"`
	PackageType string `json:"package_type" binding:"max=20"`
	Name        string `json:"name" binding:"max=150"`
	Description string `json:"description"`
	WorkspaceScope string `json:"workspace_scope" binding:"max=20"`
	Status      string `json:"status" binding:"max=20"`
	SortOrder   int    `json:"sort_order"`
}

type FeaturePackageKeySetRequest struct {
	AppKey    string   `json:"app_key" binding:"max=100"`
	ActionIDs []string `json:"action_ids"`
}

type FeaturePackageMenuSetRequest struct {
	AppKey  string   `json:"app_key" binding:"max=100"`
	MenuIDs []string `json:"menu_ids"`
}

type CollaborationWorkspaceFeaturePackageSetRequest struct {
	AppKey     string   `json:"app_key" binding:"max=100"`
	PackageIDs []string `json:"package_ids"`
}

type FeaturePackageCollaborationWorkspaceSetRequest struct {
	AppKey                    string   `json:"app_key" binding:"max=100"`
	CollaborationWorkspaceIDs []string `json:"collaboration_workspace_ids"`
}

type FeaturePackageChildSetRequest struct {
	AppKey          string   `json:"app_key" binding:"max=100"`
	ChildPackageIDs []string `json:"child_package_ids"`
}
