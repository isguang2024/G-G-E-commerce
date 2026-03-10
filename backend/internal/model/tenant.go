package model

import (
	"time"

	"github.com/google/uuid"
)

// Tenant 租户模型
type Tenant struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Remark      string    `gorm:"type:varchar(500);column:remark" json:"remark"`
	LogoURL     string    `gorm:"type:varchar(500)" json:"logo_url"`
	Plan        string    `gorm:"type:varchar(20);not null;default:'free'" json:"plan"`
	OwnerID     *uuid.UUID `gorm:"type:uuid" json:"owner_id"`
	MaxProducts int       `gorm:"default:1000" json:"max_products"`
	MaxMembers  int       `gorm:"default:5" json:"max_members"`
	Settings    string    `gorm:"type:jsonb;default:'{}'" json:"settings"`
	Status      string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}
