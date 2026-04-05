package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DefaultAppKey  = "platform-admin"
	DefaultAppName = "平台管理后台"

	AppScopeShared = "shared"
	AppScopeApp    = "app"
)

type App struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey          string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"app_key"`
	Name            string         `gorm:"type:varchar(150);not null" json:"name"`
	Description     string         `gorm:"type:text;not null;default:''" json:"description"`
	SpaceMode       string         `gorm:"type:varchar(20);not null;default:'single'" json:"space_mode"`
	DefaultSpaceKey string         `gorm:"type:varchar(100);not null;default:'default'" json:"default_space_key"`
	AuthMode        string         `gorm:"type:varchar(30);not null;default:'inherit_host'" json:"auth_mode"`
	Status          string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	IsDefault       bool           `gorm:"not null;default:false" json:"is_default"`
	Meta            MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (App) TableName() string {
	return "apps"
}

type AppHostBinding struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AppKey          string         `gorm:"type:varchar(100);not null;index" json:"app_key"`
	Host            string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"host"`
	Description     string         `gorm:"type:text;not null;default:''" json:"description"`
	IsPrimary       bool           `gorm:"not null;default:false" json:"is_primary"`
	DefaultSpaceKey string         `gorm:"type:varchar(100);not null;default:'default'" json:"default_space_key"`
	Status          string         `gorm:"type:varchar(20);not null;default:'normal'" json:"status"`
	Meta            MetaJSON       `gorm:"type:jsonb;default:'{}'::jsonb" json:"meta"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AppHostBinding) TableName() string {
	return "app_host_bindings"
}
