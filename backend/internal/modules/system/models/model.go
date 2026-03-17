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
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	Priority    int            `gorm:"default:0" json:"priority"`
	ScopeID     uuid.UUID      `gorm:"type:uuid" json:"scope_id"`
	Scope       Scope          `gorm:"foreignKey:ScopeID" json:"scope,omitempty"`
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

type Scope struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Scope) TableName() string {
	return "scopes"
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

type UserRole struct {
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	RoleID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"role_id"`
	TenantID *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

type RoleMenu struct {
	RoleID uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	MenuID uuid.UUID `gorm:"type:uuid;primaryKey" json:"menu_id"`
}

func (RoleMenu) TableName() string {
	return "role_menus"
}

type PermissionAction struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ResourceCode          string         `gorm:"type:varchar(100);not null" json:"resource_code"`
	ActionCode            string         `gorm:"type:varchar(100);not null" json:"action_code"`
	ModuleCode            string         `gorm:"type:varchar(100);not null;default:''" json:"module_code"`
	Category              string         `gorm:"type:varchar(100);default:''" json:"category"`
	Source                string         `gorm:"type:varchar(20);not null;default:'business'" json:"source"`
	FeatureKind           string         `gorm:"type:varchar(20);not null;default:'system'" json:"feature_kind"`
	Name                  string         `gorm:"type:varchar(150);not null" json:"name"`
	Description           string         `gorm:"type:varchar(255)" json:"description"`
	ScopeID               uuid.UUID      `gorm:"type:uuid;index" json:"scope_id"`
	Scope                 Scope          `gorm:"foreignKey:ScopeID" json:"scope,omitempty"`
	RequiresTenantContext bool           `gorm:"default:false" json:"requires_tenant_context"`
	Status                string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	SortOrder             int            `gorm:"default:0" json:"sort_order"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PermissionAction) TableName() string {
	return "permission_actions"
}

type RoleActionPermission struct {
	RoleID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	ActionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	Effect   string    `gorm:"type:varchar(20);not null;default:'allow'" json:"effect"`
}

func (RoleActionPermission) TableName() string {
	return "role_action_permissions"
}

type RoleDataPermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"role_id"`
	ResourceCode string    `gorm:"type:varchar(100);primaryKey" json:"resource_code"`
	ScopeCode    string    `gorm:"type:varchar(30);not null" json:"scope_code"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (RoleDataPermission) TableName() string {
	return "role_data_permissions"
}

type TenantActionPermission struct {
	TenantID uuid.UUID `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	ActionID uuid.UUID `gorm:"type:uuid;primaryKey" json:"action_id"`
	Enabled  bool      `gorm:"not null;default:true" json:"enabled"`
}

func (TenantActionPermission) TableName() string {
	return "tenant_action_permissions"
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

type APIEndpoint struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Method                string         `gorm:"type:varchar(10);not null" json:"method"`
	Path                  string         `gorm:"type:varchar(255);not null" json:"path"`
	Module                string         `gorm:"type:varchar(100);not null" json:"module"`
	FeatureKind           string         `gorm:"type:varchar(20);not null;default:'system'" json:"feature_kind"`
	Handler               string         `gorm:"type:varchar(255)" json:"handler"`
	Summary               string         `gorm:"type:varchar(255)" json:"summary"`
	ResourceCode          string         `gorm:"type:varchar(100)" json:"resource_code"`
	ActionCode            string         `gorm:"type:varchar(100)" json:"action_code"`
	ScopeID               *uuid.UUID     `gorm:"type:uuid;index" json:"scope_id"`
	Scope                 Scope          `gorm:"foreignKey:ScopeID" json:"scope,omitempty"`
	RequiresTenantContext bool           `gorm:"default:false" json:"requires_tenant_context"`
	Status                string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (APIEndpoint) TableName() string {
	return "api_endpoints"
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
