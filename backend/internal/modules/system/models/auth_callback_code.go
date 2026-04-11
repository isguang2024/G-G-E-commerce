package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthCallbackCode struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID           string         `gorm:"type:varchar(100);not null;default:'default';index" json:"tenant_id"`
	Code               string         `gorm:"type:varchar(120);not null;uniqueIndex" json:"code"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	TargetAppKey       string         `gorm:"type:varchar(100);not null;index" json:"target_app_key"`
	RedirectURI        string         `gorm:"type:text;not null" json:"redirect_uri"`
	TargetPath         string         `gorm:"type:varchar(500);not null;default:''" json:"target_path"`
	NavigationSpaceKey string         `gorm:"type:varchar(100);not null;default:''" json:"navigation_space_key"`
	State              string         `gorm:"type:varchar(200);not null" json:"state"`
	Nonce              string         `gorm:"type:varchar(200);not null" json:"nonce"`
	RequestHost        string         `gorm:"type:varchar(255);not null;default:''" json:"request_host"`
	Status             string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	ExpiresAt          time.Time      `gorm:"not null;index" json:"expires_at"`
	UsedAt             *time.Time     `json:"used_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AuthCallbackCode) TableName() string {
	return "auth_callback_codes"
}
