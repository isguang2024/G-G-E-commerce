package model

import (
	"time"

	"github.com/google/uuid"
)

// TagGroup 标签分组表
type TagGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	Color     string    `gorm:"type:varchar(7)" json:"color"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (TagGroup) TableName() string {
	return "tag_groups"
}
