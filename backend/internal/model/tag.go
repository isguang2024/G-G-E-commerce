package model

import (
	"time"

	"github.com/google/uuid"
)

// Tag 标签模型
type Tag struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	TagGroupID  *uuid.UUID `gorm:"type:uuid" json:"tag_group_id"`
	Name        string     `gorm:"type:varchar(100);not null" json:"name"`
	Color       string     `gorm:"type:varchar(7)" json:"color"`
	Description string     `gorm:"type:varchar(300)" json:"description"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName 指定表名
func (Tag) TableName() string {
	return "tags"
}
