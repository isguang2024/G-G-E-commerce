package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MetaJSON map[string]interface{}

func (m MetaJSON) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *MetaJSON) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, m)
}

func (m MetaJSON) String() string {
	if m == nil {
		return "{}"
	}
	b, _ := json.Marshal(m)
	return string(b)
}

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email          string         `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Username       string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"username"`
	PasswordHash   string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname       string         `gorm:"type:varchar(100)" json:"nickname"`
	AvatarURL      string         `gorm:"type:varchar(500)" json:"avatar_url"`
	Phone          string         `gorm:"type:varchar(20)" json:"phone"`
	SystemRemark   string         `gorm:"type:text" json:"system_remark"`
	Status         string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	IsSuperAdmin   bool           `gorm:"default:false" json:"is_super_admin"`
	LastLoginAt    *time.Time     `json:"last_login_at"`
	LastLoginIP    string         `gorm:"type:varchar(45)" json:"last_login_ip"`
	RegisterSource string         `gorm:"type:varchar(20);default:'self'" json:"register_source"`
	InvitedBy      *uuid.UUID     `gorm:"type:uuid" json:"invited_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type Role struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id"`
	Code         string         `gorm:"type:varchar(50);not null" json:"code"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Description  string         `gorm:"type:varchar(255)" json:"description"`
	Priority     int            `gorm:"default:0" json:"priority"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CustomParams MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"custom_params"`
	Status       string         `gorm:"type:varchar(20);default:'normal'" json:"status"`
	IsSystem     bool           `gorm:"default:false" json:"is_system"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

type Menu struct {
	ID            uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ParentID      *uuid.UUID       `gorm:"type:uuid" json:"parent_id"`
	ManageGroupID *uuid.UUID       `gorm:"type:uuid;index" json:"manage_group_id"`
	AppKey        string           `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	SpaceKey      string           `gorm:"type:varchar(100);not null;default:'default';index" json:"space_key"`
	Kind          string           `gorm:"type:varchar(20);not null;default:'directory';index" json:"kind"`
	Path          string           `gorm:"type:varchar(255)" json:"path"`
	Name          string           `gorm:"type:varchar(100)" json:"name"`
	Component     string           `gorm:"type:varchar(255)" json:"component"`
	Title         string           `gorm:"type:varchar(100)" json:"title"`
	Icon          string           `gorm:"type:varchar(100)" json:"icon"`
	SortOrder     int              `gorm:"default:0" json:"sort_order"`
	Hidden        bool             `gorm:"default:false" json:"hidden"`
	ManageGroup   *MenuManageGroup `gorm:"foreignKey:ManageGroupID" json:"manage_group,omitempty"`

	Meta      MetaJSON       `gorm:"type:jsonb" json:"meta"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Children []*Menu `gorm:"-" json:"children,omitempty"`
}

func (Menu) TableName() string {
	return "menus"
}

type MenuManageGroup struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	Status    string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuManageGroup) TableName() string {
	return "menu_manage_groups"
}

type UIPage struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey            string         `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	PageKey           string         `gorm:"type:varchar(150);not null" json:"page_key"`
	Name              string         `gorm:"type:varchar(150);not null" json:"name"`
	RouteName         string         `gorm:"type:varchar(150);not null" json:"route_name"`
	RoutePath         string         `gorm:"type:varchar(255);not null" json:"route_path"`
	Component         string         `gorm:"type:varchar(255);not null" json:"component"`
	SpaceKey          string         `gorm:"type:varchar(100);not null;default:'default';index" json:"space_key"`
	PageType          string         `gorm:"type:varchar(20);not null;default:'inner'" json:"page_type"`
	Source            string         `gorm:"type:varchar(20);not null;default:'manual'" json:"source"`
	ModuleKey         string         `gorm:"type:varchar(100);not null;default:''" json:"module_key"`
	SortOrder         int            `gorm:"not null;default:0" json:"sort_order"`
	ParentMenuID      *uuid.UUID     `gorm:"type:uuid" json:"parent_menu_id"`
	ParentPageKey     string         `gorm:"type:varchar(150);not null;default:''" json:"parent_page_key"`
	DisplayGroupKey   string         `gorm:"type:varchar(150);not null;default:''" json:"display_group_key"`
	ActiveMenuPath    string         `gorm:"type:varchar(255);not null;default:''" json:"active_menu_path"`
	BreadcrumbMode    string         `gorm:"type:varchar(20);not null;default:'inherit_menu'" json:"breadcrumb_mode"`
	AccessMode        string         `gorm:"type:varchar(20);not null;default:'inherit'" json:"access_mode"`
	PermissionKey     string         `gorm:"type:varchar(150);not null;default:''" json:"permission_key"`
	InheritPermission bool           `gorm:"not null;default:true" json:"inherit_permission"`
	KeepAlive         bool           `gorm:"not null;default:false" json:"keep_alive"`
	IsFullPage        bool           `gorm:"not null;default:false" json:"is_full_page"`
	Status            string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta              MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UIPage) TableName() string {
	return "ui_pages"
}

type PermissionGroup struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	GroupType   string         `gorm:"type:varchar(20);not null" json:"group_type"`
	Code        string         `gorm:"type:varchar(100);not null" json:"code"`
	Name        string         `gorm:"type:varchar(150);not null" json:"name"`
	NameEn      string         `gorm:"type:varchar(150)" json:"name_en"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	IsBuiltin   bool           `gorm:"not null;default:false" json:"is_builtin"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PermissionGroup) TableName() string {
	return "permission_groups"
}

type UserRole struct {
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"role_id"`
	TenantID *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

type RoleHiddenMenu struct {
	AppKey    string    `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (RoleHiddenMenu) TableName() string {
	return "role_hidden_menus"
}

type PermissionKey struct {
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code            string           `gorm:"type:varchar(36);uniqueIndex" json:"code"`
	APIEndpointCode string           `gorm:"type:varchar(36)" json:"api_endpoint_code"`
	PermissionKey   string           `gorm:"type:varchar(150);index:idx_permission_keys_permission_key" json:"permission_key"`
	ModuleCode      string           `gorm:"type:varchar(100);not null;default:''" json:"module_code"`
	ModuleGroupID   *uuid.UUID       `gorm:"type:uuid;index" json:"module_group_id"`
	FeatureGroupID  *uuid.UUID       `gorm:"type:uuid;index" json:"feature_group_id"`
	ContextType     string           `gorm:"type:varchar(20);not null;default:'team'" json:"context_type"`
	FeatureKind     string           `gorm:"type:varchar(20);not null;default:'system'" json:"feature_kind"`
	Name            string           `gorm:"type:varchar(150);not null" json:"name"`
	Description     string           `gorm:"type:varchar(255)" json:"description"`
	Status          string           `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	SortOrder       int              `gorm:"default:0" json:"sort_order"`
	IsBuiltin       bool             `gorm:"not null;default:false" json:"is_builtin"`
	ModuleGroup     *PermissionGroup `gorm:"foreignKey:ModuleGroupID" json:"module_group,omitempty"`
	FeatureGroup    *PermissionGroup `gorm:"foreignKey:FeatureGroupID" json:"feature_group,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

func (PermissionKey) TableName() string {
	return "permission_keys"
}

type FeaturePackage struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey      string         `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	PackageKey  string         `gorm:"type:varchar(100);not null" json:"package_key"`
	PackageType string         `gorm:"type:varchar(20);not null;default:'base'" json:"package_type"`
	Name        string         `gorm:"type:varchar(150);not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	ContextType string         `gorm:"type:varchar(20);not null;default:'team'" json:"context_type"`
	IsBuiltin   bool           `gorm:"not null;default:false" json:"is_builtin"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (FeaturePackage) TableName() string {
	return "feature_packages"
}

type FeaturePackageBundle struct {
	PackageID      uuid.UUID `gorm:"type:uuid;primaryKey" json:"package_id"`
	ChildPackageID uuid.UUID `gorm:"type:uuid;primaryKey" json:"child_package_id"`
	CreatedAt      time.Time `json:"created_at"`
}

func (FeaturePackageBundle) TableName() string {
	return "feature_package_bundles"
}

type FeaturePackageKey struct {
	PackageID uuid.UUID `gorm:"type:uuid;primaryKey" json:"package_id"`
	ActionID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (FeaturePackageKey) TableName() string {
	return "feature_package_keys"
}

type FeaturePackageMenu struct {
	PackageID uuid.UUID `gorm:"type:uuid;primaryKey" json:"package_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (FeaturePackageMenu) TableName() string {
	return "feature_package_menus"
}

type TeamFeaturePackage struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey    string     `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	TeamID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"team_id"`
	PackageID uuid.UUID  `gorm:"type:uuid;not null;index" json:"package_id"`
	Enabled   bool       `gorm:"not null;default:true" json:"enabled"`
	GrantedBy *uuid.UUID `gorm:"type:uuid" json:"granted_by"`
	GrantedAt *time.Time `json:"granted_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (TeamFeaturePackage) TableName() string {
	return "team_feature_packages"
}

type UserFeaturePackage struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey    string     `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	PackageID uuid.UUID  `gorm:"type:uuid;not null;index" json:"package_id"`
	Enabled   bool       `gorm:"not null;default:true" json:"enabled"`
	GrantedBy *uuid.UUID `gorm:"type:uuid" json:"granted_by"`
	GrantedAt *time.Time `json:"granted_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (UserFeaturePackage) TableName() string {
	return "user_feature_packages"
}

type RoleFeaturePackage struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey    string     `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	RoleID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"role_id"`
	PackageID uuid.UUID  `gorm:"type:uuid;not null;index" json:"package_id"`
	Enabled   bool       `gorm:"not null;default:true" json:"enabled"`
	GrantedBy *uuid.UUID `gorm:"type:uuid" json:"granted_by"`
	GrantedAt *time.Time `json:"granted_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (RoleFeaturePackage) TableName() string {
	return "role_feature_packages"
}

type RoleDisabledAction struct {
	AppKey    string    `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	ActionID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (RoleDisabledAction) TableName() string {
	return "role_disabled_actions"
}

type RoleDataPermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	ResourceCode string    `gorm:"type:varchar(100);primaryKey" json:"resource_code"`
	DataScope    string    `gorm:"type:varchar(30);not null;column:data_scope" json:"data_scope"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RoleDataPermission) TableName() string {
	return "role_data_permissions"
}

type TeamBlockedMenu struct {
	AppKey    string    `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	TeamID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamBlockedMenu) TableName() string {
	return "team_blocked_menus"
}

type TeamBlockedAction struct {
	AppKey    string    `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	TeamID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
	ActionID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamBlockedAction) TableName() string {
	return "team_blocked_actions"
}

type UserActionPermission struct {
	AppKey    string     `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	UserID    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"user_id"`
	ActionID  uuid.UUID  `gorm:"type:uuid;primaryKey" json:"action_id"`
	TenantID  *uuid.UUID `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	Effect    string     `gorm:"type:varchar(20);not null;default:'allow'" json:"effect"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (UserActionPermission) TableName() string {
	return "user_action_permissions"
}

type UserHiddenMenu struct {
	AppKey    string    `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserHiddenMenu) TableName() string {
	return "user_hidden_menus"
}

type PlatformUserAccessSnapshot struct {
	AppKey             string              `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	UserID             uuid.UUID           `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleIDs            []string            `gorm:"type:jsonb;serializer:json" json:"role_ids"`
	RolePackageIDs     []string            `gorm:"type:jsonb;serializer:json" json:"role_package_ids"`
	UserPackageIDs     []string            `gorm:"type:jsonb;serializer:json" json:"user_package_ids"`
	DirectPackageIDs   []string            `gorm:"type:jsonb;serializer:json" json:"direct_package_ids"`
	ExpandedPackageIDs []string            `gorm:"type:jsonb;serializer:json" json:"expanded_package_ids"`
	ActionIDs          []string            `gorm:"type:jsonb;serializer:json" json:"action_ids"`
	ActionSourceMap    map[string][]string `gorm:"type:jsonb;serializer:json" json:"action_source_map"`
	AvailableMenuIDs   []string            `gorm:"type:jsonb;serializer:json" json:"available_menu_ids"`
	AvailableMenuMap   map[string][]string `gorm:"type:jsonb;serializer:json" json:"available_menu_map"`
	MenuIDs            []string            `gorm:"type:jsonb;serializer:json" json:"menu_ids"`
	MenuSourceMap      map[string][]string `gorm:"type:jsonb;serializer:json" json:"menu_source_map"`
	HiddenMenuIDs      []string            `gorm:"type:jsonb;serializer:json" json:"hidden_menu_ids"`
	DisabledActionIDs  []string            `gorm:"type:jsonb;serializer:json" json:"disabled_action_ids"`
	HasPackageConfig   bool                `gorm:"not null;default:false" json:"has_package_config"`
	RefreshedAt        time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP" json:"refreshed_at"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func (PlatformUserAccessSnapshot) TableName() string {
	return "platform_user_access_snapshots"
}

type PlatformRoleAccessSnapshot struct {
	AppKey             string              `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	RoleID             uuid.UUID           `gorm:"type:uuid;primaryKey" json:"role_id"`
	PackageIDs         []string            `gorm:"type:jsonb;serializer:json" json:"package_ids"`
	ExpandedPackageIDs []string            `gorm:"type:jsonb;serializer:json" json:"expanded_package_ids"`
	AvailableActionIDs []string            `gorm:"type:jsonb;serializer:json" json:"available_action_ids"`
	ActionSourceMap    map[string][]string `gorm:"type:jsonb;serializer:json" json:"action_source_map"`
	DisabledActionIDs  []string            `gorm:"type:jsonb;serializer:json" json:"disabled_action_ids"`
	EffectiveActionIDs []string            `gorm:"type:jsonb;serializer:json" json:"effective_action_ids"`
	AvailableMenuIDs   []string            `gorm:"type:jsonb;serializer:json" json:"available_menu_ids"`
	MenuSourceMap      map[string][]string `gorm:"type:jsonb;serializer:json" json:"menu_source_map"`
	HiddenMenuIDs      []string            `gorm:"type:jsonb;serializer:json" json:"hidden_menu_ids"`
	EffectiveMenuIDs   []string            `gorm:"type:jsonb;serializer:json" json:"effective_menu_ids"`
	RefreshedAt        time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP" json:"refreshed_at"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func (PlatformRoleAccessSnapshot) TableName() string {
	return "platform_role_access_snapshots"
}

type TeamAccessSnapshot struct {
	AppKey             string              `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	TeamID             uuid.UUID           `gorm:"type:uuid;primaryKey" json:"team_id"`
	PackageIDs         []string            `gorm:"type:jsonb;serializer:json" json:"package_ids"`
	ExpandedPackageIDs []string            `gorm:"type:jsonb;serializer:json" json:"expanded_package_ids"`
	DerivedActionIDs   []string            `gorm:"type:jsonb;serializer:json" json:"derived_action_ids"`
	DerivedActionMap   map[string][]string `gorm:"type:jsonb;serializer:json" json:"derived_action_map"`
	BlockedActionIDs   []string            `gorm:"type:jsonb;serializer:json" json:"blocked_action_ids"`
	EffectiveActionIDs []string            `gorm:"type:jsonb;serializer:json" json:"effective_action_ids"`
	DerivedMenuIDs     []string            `gorm:"type:jsonb;serializer:json" json:"derived_menu_ids"`
	DerivedMenuMap     map[string][]string `gorm:"type:jsonb;serializer:json" json:"derived_menu_map"`
	BlockedMenuIDs     []string            `gorm:"type:jsonb;serializer:json" json:"blocked_menu_ids"`
	EffectiveMenuIDs   []string            `gorm:"type:jsonb;serializer:json" json:"effective_menu_ids"`
	RefreshedAt        time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP" json:"refreshed_at"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func (TeamAccessSnapshot) TableName() string {
	return "team_access_snapshots"
}

type TeamRoleAccessSnapshot struct {
	AppKey             string              `gorm:"type:varchar(100);not null;default:'platform-admin';primaryKey" json:"app_key"`
	TeamID             uuid.UUID           `gorm:"type:uuid;primaryKey" json:"team_id"`
	RoleID             uuid.UUID           `gorm:"type:uuid;primaryKey" json:"role_id"`
	PackageIDs         []string            `gorm:"type:jsonb;serializer:json" json:"package_ids"`
	ExpandedPackageIDs []string            `gorm:"type:jsonb;serializer:json" json:"expanded_package_ids"`
	AvailableActionIDs []string            `gorm:"type:jsonb;serializer:json" json:"available_action_ids"`
	DisabledActionIDs  []string            `gorm:"type:jsonb;serializer:json" json:"disabled_action_ids"`
	ActionIDs          []string            `gorm:"type:jsonb;serializer:json" json:"action_ids"`
	ActionSourceMap    map[string][]string `gorm:"type:jsonb;serializer:json" json:"action_source_map"`
	AvailableMenuIDs   []string            `gorm:"type:jsonb;serializer:json" json:"available_menu_ids"`
	HiddenMenuIDs      []string            `gorm:"type:jsonb;serializer:json" json:"hidden_menu_ids"`
	MenuIDs            []string            `gorm:"type:jsonb;serializer:json" json:"menu_ids"`
	MenuSourceMap      map[string][]string `gorm:"type:jsonb;serializer:json" json:"menu_source_map"`
	Inherited          bool                `gorm:"not null;default:false" json:"inherited"`
	RefreshedAt        time.Time           `gorm:"not null;default:CURRENT_TIMESTAMP" json:"refreshed_at"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func (TeamRoleAccessSnapshot) TableName() string {
	return "team_role_access_snapshots"
}

type APIEndpoint struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code         string         `gorm:"type:varchar(36);uniqueIndex" json:"code"`
	Method       string         `gorm:"type:varchar(10);not null" json:"method"`
	Path         string         `gorm:"type:varchar(255);not null" json:"path"`
	AppScope     string         `gorm:"type:varchar(20);not null;default:'shared';index" json:"app_scope"`
	AppKey       string         `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	FeatureKind  string         `gorm:"type:varchar(20);not null;default:'system'" json:"feature_kind"`
	Handler      string         `gorm:"type:varchar(255)" json:"handler"`
	Summary      string         `gorm:"type:varchar(255)" json:"summary"`
	CategoryID   *uuid.UUID     `gorm:"type:uuid;index" json:"category_id"`
	ContextScope string         `gorm:"type:varchar(20);not null;default:'optional'" json:"context_scope"`
	Source       string         `gorm:"type:varchar(20);not null;default:'sync'" json:"source"`
	Status       string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (APIEndpoint) TableName() string {
	return "api_endpoints"
}

type APIEndpointCategory struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code      string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"code"`
	Name      string         `gorm:"type:varchar(150);not null" json:"name"`
	NameEn    string         `gorm:"type:varchar(150);not null;default:''" json:"name_en"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	Status    string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (APIEndpointCategory) TableName() string {
	return "api_endpoint_categories"
}

type APIEndpointPermissionBinding struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EndpointCode  string         `gorm:"type:varchar(36);not null;index" json:"endpoint_code"`
	PermissionKey string         `gorm:"type:varchar(150);not null;index" json:"permission_key"`
	MatchMode     string         `gorm:"type:varchar(10);not null;default:'ANY'" json:"match_mode"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (APIEndpointPermissionBinding) TableName() string {
	return "api_endpoint_permission_bindings"
}

type Tenant struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name       string         `gorm:"type:varchar(100);not null" json:"name"`
	Remark     string         `gorm:"type:text" json:"remark"`
	LogoURL    string         `gorm:"type:varchar(500)" json:"logo_url"`
	Plan       string         `gorm:"type:varchar(50);default:'free'" json:"plan"`
	OwnerID    uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	MaxMembers int            `gorm:"default:10" json:"max_members"`
	Status     string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Tenant) TableName() string {
	return "tenants"
}

type TenantMember struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null" json:"tenant_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	RoleCode  string         `gorm:"type:varchar(50);default:'member'" json:"role_code"`
	RoleID    *uuid.UUID     `gorm:"type:uuid" json:"role_id"`
	Status    string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	JoinedAt  time.Time      `json:"joined_at"`
	InvitedBy *uuid.UUID     `gorm:"type:uuid" json:"invited_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (TenantMember) TableName() string {
	return "tenant_members"
}

type MemberSearchParams struct {
	UserID   string
	UserName string
	RoleCode string
}

type APIKey struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CreatedBy  *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	Name       string     `gorm:"type:varchar(200);not null" json:"name"`
	KeyHash    string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

type MediaAsset struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	UploadedBy *uuid.UUID `gorm:"type:uuid" json:"uploaded_by"`
	Filename   string     `gorm:"type:varchar(500);not null" json:"filename"`
	StorageKey string     `gorm:"type:varchar(1000);not null" json:"storage_key"`
	URL        string     `gorm:"type:varchar(1000);not null" json:"url"`
	MimeType   string     `gorm:"type:varchar(100)" json:"mime_type"`
	Size       int64      `json:"size"`
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	AltText    string     `gorm:"type:varchar(500)" json:"alt_text"`
	Hash       string     `gorm:"type:varchar(64);index" json:"hash"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (MediaAsset) TableName() string {
	return "media_assets"
}

type MenuBackup struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	AppKey      string    `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	// SpaceKey 为空表示历史全局备份；非空表示该备份仅针对对应菜单空间。
	SpaceKey  string         `gorm:"type:varchar(100);not null;default:'';index" json:"space_key"`
	MenuData  string         `gorm:"type:text;not null" json:"menu_data"` // JSON 格式的菜单数据
	CreatedBy *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuBackup) TableName() string {
	return "menu_backups"
}

type SystemSetting struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Key       string         `gorm:"type:varchar(150);not null" json:"key"`
	Value     MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"value"`
	Status    string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

type Message struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MessageType          string         `gorm:"type:varchar(20);not null;default:'notice'" json:"message_type"`
	BizType              string         `gorm:"type:varchar(100);not null;default:''" json:"biz_type"`
	ScopeType            string         `gorm:"type:varchar(20);not null;default:'platform'" json:"scope_type"`
	ScopeID              *uuid.UUID     `gorm:"type:uuid;index" json:"scope_id"`
	SenderID             *uuid.UUID     `gorm:"type:uuid;index" json:"sender_id"`
	SenderType           string         `gorm:"type:varchar(30);not null;default:'system'" json:"sender_type"`
	SenderUserID         *uuid.UUID     `gorm:"type:uuid;index" json:"sender_user_id"`
	SenderNameSnapshot   string         `gorm:"type:varchar(150);not null;default:''" json:"sender_name_snapshot"`
	SenderAvatarSnapshot string         `gorm:"type:varchar(500);not null;default:''" json:"sender_avatar_snapshot"`
	SenderServiceKey     string         `gorm:"type:varchar(100);not null;default:''" json:"sender_service_key"`
	AudienceType         string         `gorm:"type:varchar(30);not null;default:'specified_users'" json:"audience_type"`
	AudienceScope        string         `gorm:"type:varchar(20);not null;default:'platform'" json:"audience_scope"`
	TargetTenantID       *uuid.UUID     `gorm:"type:uuid;index" json:"target_tenant_id"`
	TargetRoleCodes      []string       `gorm:"type:jsonb;serializer:json" json:"target_role_codes"`
	TargetUserIDs        []string       `gorm:"type:jsonb;serializer:json" json:"target_user_ids"`
	TargetGroupIDs       []string       `gorm:"type:jsonb;serializer:json" json:"target_group_ids"`
	TemplateID           *uuid.UUID     `gorm:"type:uuid;index" json:"template_id"`
	Title                string         `gorm:"type:varchar(255);not null" json:"title"`
	Summary              string         `gorm:"type:text;not null;default:''" json:"summary"`
	Content              string         `gorm:"type:text;not null;default:''" json:"content"`
	Priority             string         `gorm:"type:varchar(20);not null;default:'normal'" json:"priority"`
	ActionType           string         `gorm:"type:varchar(20);not null;default:'none'" json:"action_type"`
	ActionTarget         string         `gorm:"type:varchar(500);not null;default:''" json:"action_target"`
	Status               string         `gorm:"type:varchar(20);not null;default:'published'" json:"status"`
	PublishedAt          *time.Time     `json:"published_at"`
	ExpiredAt            *time.Time     `json:"expired_at"`
	Meta                 MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Message) TableName() string {
	return "messages"
}

type MessageDelivery struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	MessageID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"message_id"`
	RecipientUserID uuid.UUID      `gorm:"type:uuid;not null;index" json:"recipient_user_id"`
	RecipientTeamID *uuid.UUID     `gorm:"type:uuid;index" json:"recipient_team_id"`
	BoxType         string         `gorm:"type:varchar(20);not null;default:'notice'" json:"box_type"`
	DeliveryStatus  string         `gorm:"type:varchar(20);not null;default:'unread'" json:"delivery_status"`
	TodoStatus      string         `gorm:"type:varchar(20);not null;default:''" json:"todo_status"`
	ReadAt          *time.Time     `json:"read_at"`
	ArchivedAt      *time.Time     `json:"archived_at"`
	DoneAt          *time.Time     `json:"done_at"`
	LastActionAt    *time.Time     `json:"last_action_at"`
	Meta            MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MessageDelivery) TableName() string {
	return "message_deliveries"
}

type MessageTemplate struct {
	ID                   uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TemplateKey          string         `gorm:"type:varchar(120);not null;uniqueIndex" json:"template_key"`
	Name                 string         `gorm:"type:varchar(150);not null" json:"name"`
	Description          string         `gorm:"type:text;not null;default:''" json:"description"`
	MessageType          string         `gorm:"type:varchar(20);not null;default:'notice'" json:"message_type"`
	OwnerScope           string         `gorm:"type:varchar(20);not null;default:'platform'" json:"owner_scope"`
	OwnerTenantID        *uuid.UUID     `gorm:"type:uuid;index" json:"owner_tenant_id"`
	AudienceType         string         `gorm:"type:varchar(30);not null;default:'specified_users'" json:"audience_type"`
	TitleTemplate        string         `gorm:"type:text;not null;default:''" json:"title_template"`
	SummaryTemplate      string         `gorm:"type:text;not null;default:''" json:"summary_template"`
	ContentTemplate      string         `gorm:"type:text;not null;default:''" json:"content_template"`
	ActionType           string         `gorm:"type:varchar(20);not null;default:'none'" json:"action_type"`
	ActionTargetTemplate string         `gorm:"type:text;not null;default:''" json:"action_target_template"`
	Status               string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta                 MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MessageTemplate) TableName() string {
	return "message_templates"
}

type MessageSender struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScopeType   string         `gorm:"type:varchar(20);not null;default:'platform';index" json:"scope_type"`
	ScopeID     *uuid.UUID     `gorm:"type:uuid;index" json:"scope_id"`
	Name        string         `gorm:"type:varchar(120);not null" json:"name"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	AvatarURL   string         `gorm:"type:varchar(500);not null;default:''" json:"avatar_url"`
	IsDefault   bool           `gorm:"not null;default:false" json:"is_default"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta        MetaJSON       `gorm:"type:jsonb;serializer:json" json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MessageSender) TableName() string {
	return "message_senders"
}

type MessageRecipientGroup struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScopeType   string         `gorm:"type:varchar(20);not null;default:'platform';index" json:"scope_type"`
	ScopeID     *uuid.UUID     `gorm:"type:uuid;index" json:"scope_id"`
	Name        string         `gorm:"type:varchar(120);not null" json:"name"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	MatchMode   string         `gorm:"type:varchar(20);not null;default:'manual'" json:"match_mode"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta        MetaJSON       `gorm:"type:jsonb;serializer:json" json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MessageRecipientGroup) TableName() string {
	return "message_recipient_groups"
}

type MessageRecipientGroupTarget struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	GroupID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"group_id"`
	TargetType string         `gorm:"type:varchar(30);not null" json:"target_type"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	TenantID   *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id"`
	RoleCode   string         `gorm:"type:varchar(80);not null;default:''" json:"role_code"`
	PackageKey string         `gorm:"type:varchar(120);not null;default:''" json:"package_key"`
	SortOrder  int            `gorm:"not null;default:0" json:"sort_order"`
	Meta       MetaJSON       `gorm:"type:jsonb;serializer:json" json:"meta"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MessageRecipientGroupTarget) TableName() string {
	return "message_recipient_group_targets"
}

type RiskOperationAudit struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OperatorID     *uuid.UUID     `gorm:"type:uuid;index" json:"operator_id"`
	ObjectType     string         `gorm:"type:varchar(80);not null;index" json:"object_type"`
	ObjectID       string         `gorm:"type:varchar(120);not null;index" json:"object_id"`
	OperationType  string         `gorm:"type:varchar(80);not null;index" json:"operation_type"`
	BeforeSummary  MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"before_summary"`
	AfterSummary   MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"after_summary"`
	ImpactSummary  MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"impact_summary"`
	RequestID      string         `gorm:"type:varchar(120);not null;default:'';index" json:"request_id"`
	CreatedAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (RiskOperationAudit) TableName() string {
	return "risk_operation_audits"
}

type FeaturePackageVersion struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PackageID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"package_id"`
	VersionNo   int            `gorm:"not null;default:1" json:"version_no"`
	ChangeType  string         `gorm:"type:varchar(50);not null;default:'update'" json:"change_type"`
	Snapshot    MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"snapshot"`
	OperatorID  *uuid.UUID     `gorm:"type:uuid;index" json:"operator_id"`
	RequestID   string         `gorm:"type:varchar(120);not null;default:'';index" json:"request_id"`
	CreatedAt   time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP;index" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (FeaturePackageVersion) TableName() string {
	return "feature_package_versions"
}

type PermissionBatchTemplate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(120);not null;uniqueIndex:idx_permission_batch_templates_name_deleted" json:"name"`
	Description string         `gorm:"type:varchar(255);not null;default:''" json:"description"`
	Payload     MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"payload"`
	CreatedBy   *uuid.UUID     `gorm:"type:uuid;index" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;uniqueIndex:idx_permission_batch_templates_name_deleted" json:"deleted_at,omitempty"`
}

func (PermissionBatchTemplate) TableName() string {
	return "permission_batch_templates"
}
