package model

import (
	"time"

	"github.com/google/uuid"
)

// Category 分类模型
type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"` // NULL = 系统预置分类
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	Code        string     `gorm:"type:varchar(20);index" json:"code"`
	NameZh      string     `gorm:"type:varchar(200);not null" json:"name_zh"`
	NameEn      string     `gorm:"type:varchar(200)" json:"name_en"`
	Slug        string     `gorm:"type:varchar(200)" json:"slug"`
	Path        string     `gorm:"type:ltree;index:idx_categories_path,type:gist" json:"path"` // PostgreSQL ltree
	Depth       int        `gorm:"not null;default:0" json:"depth"`
	Icon        string     `gorm:"type:varchar(100)" json:"icon"`
	CoverImage  string     `gorm:"type:varchar(500)" json:"cover_image"`
	Description string     `gorm:"type:text" json:"description"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}
