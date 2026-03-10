package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role 全局角色模型（仅全局，供各应用使用）
// ScopeID: 关联到 scopes 表的作用域ID
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code        string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	ScopeID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"scope_id"`
	Scope       Scope          `gorm:"foreignKey:ScopeID" json:"scope,omitempty"`
	Enabled     bool           `gorm:"default:true;not null" json:"enabled"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// UserRole 用户-全局角色关联（多对多）
// TenantID 为空表示全局拥有；非空表示用户在该团队内拥有该全局角色（用于 scope=team 的角色）
type UserRole struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_role_tenant;uniqueIndex:idx_user_role_tenant" json:"user_id"`
	RoleID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_user_role_tenant" json:"role_id"`
	TenantID  *uuid.UUID     `gorm:"type:uuid;index:idx_user_role_tenant;uniqueIndex:idx_user_role_tenant" json:"tenant_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}

// RoleMenu 全局角色-菜单关联（用于非团队场景或 scope=global 的权限）
type RoleMenu struct {
	RoleID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_role_menu" json:"role_id"`
	MenuID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_role_menu" json:"menu_id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (RoleMenu) TableName() string {
	return "role_menus"
}
