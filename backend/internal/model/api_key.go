package model

import (
	"time"

	"github.com/google/uuid"
)

// APIKey API 密钥表
type APIKey struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	Name        string     `gorm:"type:varchar(200);not null" json:"name"`
	KeyHash     string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	KeyPrefix   string     `gorm:"type:varchar(10)" json:"key_prefix"`
	Permissions string     `gorm:"type:jsonb;default:'[\"products:read\"]'" json:"permissions"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName 指定表名
func (APIKey) TableName() string {
	return "api_keys"
}
