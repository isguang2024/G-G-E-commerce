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
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    *uuid.UUID     `gorm:"type:uuid;index" json:"tenant_id"`
	Code        string         `gorm:"type:varchar(50);not null" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	Priority    int            `gorm:"default:0" json:"priority"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Status      string         `gorm:"type:varchar(20);default:'normal'" json:"status"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Role) TableName() string {
	return "roles"
}

type Menu struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ParentID  *uuid.UUID `gorm:"type:uuid" json:"parent_id"`
	Path      string     `gorm:"type:varchar(255)" json:"path"`
	Name      string     `gorm:"type:varchar(100)" json:"name"`
	Component string     `gorm:"type:varchar(255)" json:"component"`
	Title     string     `gorm:"type:varchar(100)" json:"title"`
	Icon      string     `gorm:"type:varchar(100)" json:"icon"`
	SortOrder int        `gorm:"default:0" json:"sort_order"`
	Hidden    bool       `gorm:"default:false" json:"hidden"`

	Meta      MetaJSON       `gorm:"type:jsonb" json:"meta"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Children []*Menu `gorm:"-" json:"children,omitempty"`
}

func (Menu) TableName() string {
	return "menus"
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
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (RoleHiddenMenu) TableName() string {
	return "role_hidden_menus"
}

type PermissionAction struct {
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

func (PermissionAction) TableName() string {
	return "permission_keys"
}

type FeaturePackage struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
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

type FeaturePackageAction struct {
	PackageID uuid.UUID `gorm:"type:uuid;primaryKey" json:"package_id"`
	ActionID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (FeaturePackageAction) TableName() string {
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
	TeamID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamBlockedMenu) TableName() string {
	return "team_blocked_menus"
}

type TeamBlockedAction struct {
	TeamID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"team_id"`
	ActionID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamBlockedAction) TableName() string {
	return "team_blocked_actions"
}

type UserActionPermission struct {
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
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	MenuID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserHiddenMenu) TableName() string {
	return "user_hidden_menus"
}

type PlatformUserAccessSnapshot struct {
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
	Module       string         `gorm:"type:varchar(100);not null" json:"module"`
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
	EndpointID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"endpoint_id"`
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
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	MenuData    string         `gorm:"type:text;not null" json:"menu_data"` // JSON 格式的菜单数据
	CreatedBy   *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuBackup) TableName() string {
	return "menu_backups"
}
