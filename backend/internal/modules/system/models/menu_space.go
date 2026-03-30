package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const DefaultMenuSpaceKey = "default"

// MenuSpace 菜单空间定义
type MenuSpace struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SpaceKey        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"space_key"`
	Name            string         `gorm:"type:varchar(150);not null" json:"name"`
	Description     string         `gorm:"type:text;not null;default:''" json:"description"`
	DefaultHomePath string         `gorm:"type:varchar(255);not null;default:''" json:"default_home_path"`
	IsDefault       bool           `gorm:"not null;default:false" json:"is_default"`
	Status          string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta            MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuSpace) TableName() string {
	return "menu_spaces"
}

// MenuSpaceHostBinding 菜单空间 Host 绑定
type MenuSpaceHostBinding struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SpaceKey    string         `gorm:"type:varchar(100);not null;index" json:"space_key"`
	Host        string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"host"`
	Description string         `gorm:"type:text;not null;default:''" json:"description"`
	IsDefault   bool           `gorm:"not null;default:false" json:"is_default"`
	Status      string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta        MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MenuSpaceHostBinding) TableName() string {
	return "menu_space_host_bindings"
}
