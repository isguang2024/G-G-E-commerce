package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SocialAuthProvider 第三方社交登录提供方配置。
// tenant_id 用于多租户隔离，查询必须显式附带 tenant 条件。
type SocialAuthProvider struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID     string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	ProviderKey  string         `gorm:"type:varchar(64);not null" json:"provider_key"`
	ProviderName string         `gorm:"type:varchar(128);not null" json:"provider_name"`
	AuthURL      string         `gorm:"type:text;not null;default:''" json:"auth_url"`
	TokenURL     string         `gorm:"type:text;not null;default:''" json:"token_url"`
	UserInfoURL  string         `gorm:"type:text;not null;default:''" json:"user_info_url"`
	Scope        string         `gorm:"type:text;not null;default:''" json:"scope"`
	ClientID     string         `gorm:"type:text;not null;default:''" json:"client_id"`
	ClientSecret string         `gorm:"type:text;not null;default:''" json:"client_secret"`
	RedirectURI  string         `gorm:"type:text;not null;default:''" json:"redirect_uri"`
	Enabled      bool           `gorm:"not null;default:false" json:"enabled"`
	Config       MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"config"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (SocialAuthProvider) TableName() string { return "social_auth_providers" }

// UserSocialAccount 用户与第三方社交账号绑定关系。
// 通过 (tenant_id, provider_key, provider_uid) 保证跨租户不串绑。
type UserSocialAccount struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID         string         `gorm:"type:varchar(64);not null;default:'default';index" json:"tenant_id"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ProviderKey      string         `gorm:"type:varchar(64);not null;index" json:"provider_key"`
	ProviderUID      string         `gorm:"type:varchar(255);not null;index" json:"provider_uid"`
	ProviderUsername string         `gorm:"type:varchar(255);not null;default:''" json:"provider_username"`
	ProviderEmail    string         `gorm:"type:varchar(255);not null;default:''" json:"provider_email"`
	AvatarURL        string         `gorm:"type:text;not null;default:''" json:"avatar_url"`
	Profile          MetaJSON       `gorm:"type:jsonb;not null;default:'{}'::jsonb" json:"profile"`
	LinkedAt         time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"linked_at"`
	LastLoginAt      *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (UserSocialAccount) TableName() string { return "user_social_accounts" }
