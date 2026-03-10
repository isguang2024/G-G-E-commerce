package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Group 分组模型
type Group struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ParentID    *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id"`
	Path        string         `gorm:"type:ltree;index:idx_groups_path,type:gist" json:"path"` // PostgreSQL ltree
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Color       string         `gorm:"type:varchar(7)" json:"color"`
	Icon        string         `gorm:"type:varchar(50)" json:"icon"`
	Depth       int            `gorm:"not null;default:0" json:"depth"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedBy   *uuid.UUID     `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName 指定表名
func (Group) TableName() string {
	return "groups"
}
