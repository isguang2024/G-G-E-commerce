package model

import (
	"time"

	"github.com/google/uuid"
)

// TenantMember 团队成员关联表
type TenantMember struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_user" json:"tenant_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_tenant_user" json:"user_id"`
	RoleID    *uuid.UUID    `gorm:"type:uuid" json:"role_id"`    // 角色ID，关联roles表
	Status    string        `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	InvitedBy *uuid.UUID    `gorm:"type:uuid" json:"invited_by"`
	JoinedAt  *time.Time    `json:"joined_at"`
	CreatedAt time.Time     `json:"created_at"`
}

// TableName 指定表名
func (TenantMember) TableName() string {
	return "tenant_members"
}
