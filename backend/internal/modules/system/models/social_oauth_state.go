package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialOAuthState OAuth state 持久化，防止回放与 CSRF。
type SocialOAuthState struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	ProviderKey  string         `gorm:"type:varchar(64);not null;index" json:"provider_key"`
	State        string         `gorm:"type:varchar(128);not null;index" json:"state"`
	LoginPageKey string         `gorm:"type:varchar(128);not null;default:''" json:"login_page_key"`
	PageScene    string         `gorm:"type:varchar(32);not null;default:'login'" json:"page_scene"`
	TargetAppKey string         `gorm:"type:varchar(128);not null;default:''" json:"target_app_key"`
	RequestPath  string         `gorm:"type:text;not null;default:''" json:"request_path"`
	RedirectURI  string         `gorm:"type:text;not null;default:''" json:"redirect_uri"`
	Nonce        string         `gorm:"type:varchar(128);not null;default:''" json:"nonce"`
	Meta         MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"meta"`
	ExpiresAt    time.Time      `gorm:"not null;index" json:"expires_at"`
	UsedAt       *time.Time     `json:"used_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SocialOAuthState) TableName() string { return "social_oauth_states" }
