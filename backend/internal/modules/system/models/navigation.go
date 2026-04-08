package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	MenuKindDirectory = "directory"
	MenuKindEntry     = "entry"
	MenuKindExternal  = "external"

	PageTypeInner        = "inner"
	PageTypeStandalone   = "standalone"
	PageTypeGroup        = "group"
	PageTypeDisplayGroup = "display_group"
)

// PageSpaceBinding 仅用于无菜单父级、无页面父级的少量独立页暴露控制。
// 常规菜单页与其派生内页都应从菜单或父页面继承空间，不再复制页面定义。
type PageSpaceBinding struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey    string         `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	PageID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"page_id"`
	SpaceKey  string         `gorm:"type:varchar(100);not null;index" json:"space_key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PageSpaceBinding) TableName() string {
	return "page_space_bindings"
}
