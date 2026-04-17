package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MenuDefinition struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey        string         `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	MenuKey       string         `gorm:"type:varchar(150);not null" json:"menu_key"`
	Kind          string         `gorm:"type:varchar(20);not null;default:'directory';index" json:"kind"`
	Path          string         `gorm:"type:varchar(255)" json:"path"`
	Name          string         `gorm:"type:varchar(100)" json:"name"`
	Component     string         `gorm:"type:varchar(255)" json:"component"`
	PageKey       string         `gorm:"type:varchar(150);not null;default:''" json:"page_key"`
	PermissionKey string         `gorm:"type:varchar(150);not null;default:''" json:"permission_key"`
	DefaultTitle  string         `gorm:"type:varchar(100)" json:"default_title"`
	DefaultIcon   string         `gorm:"type:varchar(100)" json:"default_icon"`
	Status        string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta          MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuDefinition) TableName() string {
	return "menu_definitions"
}

type SpaceMenuPlacement struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey        string         `gorm:"type:varchar(100);not null;default:'platform-admin';index" json:"app_key"`
	MenuSpaceKey  string         `gorm:"type:varchar(100);not null;default:'default';index" json:"menu_space_key"`
	MenuKey       string         `gorm:"type:varchar(150);not null;index" json:"menu_key"`
	ParentMenuKey string         `gorm:"type:varchar(150);not null;default:''" json:"parent_menu_key"`
	SortOrder     int            `gorm:"not null;default:0" json:"sort_order"`
	Hidden        bool           `gorm:"not null;default:false" json:"hidden"`
	TitleOverride string         `gorm:"type:varchar(100);not null;default:''" json:"title_override"`
	IconOverride  string         `gorm:"type:varchar(100);not null;default:''" json:"icon_override"`
	MetaOverride  MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta_override"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SpaceMenuPlacement) TableName() string {
	return "space_menu_placements"
}
